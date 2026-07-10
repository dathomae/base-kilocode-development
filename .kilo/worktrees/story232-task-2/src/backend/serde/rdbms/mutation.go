package rdbms

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	c "chronocone/backend/constants"
	"chronocone/backend/serde/rdbms/gen/postgres/public/enum"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/lib/pq"
)

// ResolveStatus is the result of a ResolveEntityID lookup. Callers must switch on it;
// do not degrade it to a bool before reactivation logic is implemented.
type ResolveStatus int8

const (
	ResolveNotFound   ResolveStatus = 0
	ResolveActive     ResolveStatus = 1
	ResolveTombstoned ResolveStatus = 2
)

// LockableEntity describes the minimal contract a value must satisfy to be
// processed by ExecuteBitemporalMutation. EBM owns time: the staleness check
// is driven by the caller-supplied maxTransactionStart parameter, not by
// interrogating the entity. As a result LockableEntity exposes only an
// identity (EntityID) and an immutable natural key (NaturalKey) used for
// immutability verification on tables that carry a 'key' column.
//
// Implementers are passed directly to Go-Jet's qrm and INSERT MODELS: their
// Go type name must normalize to their table name, and their field names
// must normalize to column names (lowercase, underscores stripped). No
// wrapping/Unwrap indirection — the entity is the persistence record.
type LockableEntity interface {
	EntityID() int64
	NaturalKey() string
}

type BitemporalMutationBuilder[T LockableEntity] func(freshState []T, effectiveAt postgres.Expression) []postgres.InsertStatement

// validateRegistry checks that every registered entry has non-empty IDColumns and
// NonMetaColumns. Called on first use to catch misconfiguration early.
func validateRegistry() error {
	for typ, entry := range BitemporalTableRegistry {
		if len(entry.IDColumns) == 0 {
			return fmt.Errorf("BitemporalTableRegistry: entry for %v has empty IDColumns", typ)
		}
		// NonMetaColumns may legitimately be empty for tables with only ID + meta columns,
		// but warn rather than error — an empty list just means the FLUSH copies no extra cols.
	}
	return nil
}

// findFieldForCol locates a struct field on entity whose name normalizes
// (lowercase + strip underscores) to the same value as the column's name.
// Traverses promoted fields from embedded structs and pointers via
// FieldByNameFunc, so it works for entities composed of persistence mixins
// or test wrappers that embed a generated model.
func findFieldForCol(entity any, colName string) reflect.Value {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	norm := normalizeColName(colName)
	return v.FieldByNameFunc(func(n string) bool {
		return strings.ToLower(n) == norm
	})
}

// buildEntityIDCondition builds a per-entity AND condition from all
// IDColumns, extracting values from the entity by normalized field-name
// match.
func buildEntityIDCondition(entry BitemporalRegistryEntry, entity any) postgres.BoolExpression {
	var conditions []postgres.BoolExpression
	for _, col := range entry.IDColumns {
		field := findFieldForCol(entity, col.Name())
		if !field.IsValid() {
			panic(fmt.Sprintf("buildEntityIDCondition: field for column %q not found in %T", col.Name(), entity))
		}
		switch c := col.(type) {
		case postgres.ColumnInteger:
			conditions = append(conditions, c.EQ(postgres.Int(field.Int())))
		case postgres.ColumnString:
			conditions = append(conditions, c.EQ(postgres.String(field.String())))
		default:
			panic(fmt.Sprintf("buildEntityIDCondition: unsupported column type %T for column %q", col, col.Name()))
		}
	}
	return postgres.AND(conditions...)
}

// buildBatchIDCondition builds the WHERE condition for a set of entities against a table's
// IDColumns. Single-column tables use the efficient ANY($1::int8[]) form; composite-key
// tables produce an OR of per-entity AND conditions.
func buildBatchIDCondition(entry BitemporalRegistryEntry, entities []LockableEntity) postgres.BoolExpression {
	if len(entry.IDColumns) == 1 {
		intCol := entry.IDColumns[0].(postgres.ColumnInteger)
		ids := make([]int64, len(entities))
		for i, e := range entities {
			ids[i] = e.EntityID()
		}
		return intCol.EQ(postgres.ANY(postgres.Int64Array(ids...)))
	}
	var orConds []postgres.BoolExpression
	for _, entity := range entities {
		orConds = append(orConds, buildEntityIDCondition(entry, entity))
	}
	return postgres.OR(orConds...)
}

// buildCompositeIDWhereSQL builds a raw-SQL WHERE fragment for composite-key tables
// (e.g. "(metric_id = 42 AND dimension_id = 5) OR (metric_id = 43 AND dimension_id = 6)").
// Only integer and string column types are supported; integer values are formatted directly
// and string values are single-quote escaped. Returns "" if entities is empty.
func buildCompositeIDWhereSQL(entry BitemporalRegistryEntry, entities []LockableEntity) string {
	if len(entities) == 0 {
		return ""
	}
	var orParts []string
	for _, entity := range entities {
		var andParts []string
		for _, col := range entry.IDColumns {
			field := findFieldForCol(entity, col.Name())
			if !field.IsValid() {
				panic(fmt.Sprintf("buildCompositeIDWhereSQL: field for column %q not found in %T", col.Name(), entity))
			}
			switch col.(type) {
			case postgres.ColumnInteger:
				andParts = append(andParts, fmt.Sprintf("%s = %d", col.Name(), field.Int()))
			case postgres.ColumnString:
				escaped := strings.ReplaceAll(field.String(), "'", "''")
				andParts = append(andParts, fmt.Sprintf("%s = '%s'", col.Name(), escaped))
			}
		}
		orParts = append(orParts, "("+strings.Join(andParts, " AND ")+")")
	}
	return strings.Join(orParts, " OR ")
}

// tableHasKeyCol returns true when the registry entry's NonMetaColumns include a column
// named "key". Junction and localization tables do not have a key column; key immutability
// is enforced by the composite primary key itself.
func tableHasKeyCol(entry BitemporalRegistryEntry) bool {
	for _, c := range entry.NonMetaColumns {
		if c.Name() == "key" {
			return true
		}
	}
	return false
}

// ExecuteBitemporalMutation orchestrates a multi-entity atomic mutation following the
// lock-read-compute-write lifecycle. It supports both new entities (zero ID) and existing ones (populated ID).
// New entities are created; existing entities are updated. Returns the resulting entities with IDs populated.
// If builder is nil, a DefaultBitemporalInserter is used which re-asserts the models in knownState.
func ExecuteBitemporalMutation[T LockableEntity](
	ctx context.Context,
	db Executor,
	maxTransactionStart time.Time,
	effectiveAt *time.Time,
	knownState []T,
	builder BitemporalMutationBuilder[T],
) ([]T, error) {
	if len(knownState) == 0 {
		return nil, nil
	}

	if err := validateRegistry(); err != nil {
		return nil, fmt.Errorf("ExecuteBitemporalMutation: %w", err)
	}

	// Query the canonical current time from PostgreSQL once. Using the DB clock
	// (not time.Now()) means every server in a fleet — K8s pods, cloud function
	// instances — agrees on the same epoch. PostgreSQL is the single writer, so
	// its CURRENT_TIMESTAMP is the authoritative wall clock for all bitemporal
	// boundaries.
	var pgNow time.Time
	if err := db.QueryRowContext(ctx, "SELECT CURRENT_TIMESTAMP").Scan(&pgNow); err != nil {
		return nil, fmt.Errorf("ExecuteBitemporalMutation: failed to query current timestamp: %w", err)
	}

	// Guard: Effective date must be at or after the current DB-clock time.
	if effectiveAt != nil && effectiveAt.Before(pgNow) {
		return nil, fmt.Errorf("ExecuteBitemporalMutation: effectiveAt (%v) must be at or after current time (%v)", *effectiveAt, pgNow)
	}

	effExpr := postgres.CURRENT_TIMESTAMP()
	if effectiveAt != nil {
		effExpr = postgres.TimestampzT(*effectiveAt)
	}

	// Phase 1: Grouping & Lock Key Generation
	// groups is keyed by model reflect.Type to allow per-table registry lookups downstream.
	// Entities are partitioned by ID: newEntities (zero ID) and existingEntities (non-zero ID).
	groups := make(map[reflect.Type][]LockableEntity)
	groupsNew := make(map[reflect.Type][]LockableEntity)
	groupsExisting := make(map[reflect.Type][]LockableEntity)
	allIDs := make([]int64, 0, len(knownState))

	for _, entity := range knownState {
		id := entity.EntityID()
		if id < 0 {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: system entities cannot be mutated")
		}

		typ := reflect.TypeOf(entity)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		entry, ok := BitemporalTableRegistry[typ]
		if !ok {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: type %v (%T) not found in BitemporalTableRegistry", typ, entity)
		}

		groups[typ] = append(groups[typ], entity)

		if id == 0 {
			groupsNew[typ] = append(groupsNew[typ], entity)
		} else {
			groupsExisting[typ] = append(groupsExisting[typ], entity)
			lockKey := advisoryLockKey(entry.LockType, id)
			allIDs = append(allIDs, lockKey)
		}
	}

	// Phase 2: Batch Lock (Deterministic to prevent deadlocks)
	// Only lock existing entities; new entities don't have a pre-existing lock context.
	if len(allIDs) > 0 {
		if err := acquireLocks(ctx, db, allIDs); err != nil {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: failed to acquire locks: %w", err)
		}
	}

	// Phase 3: Staleness & Immutability Check (only for existing entities)
	// A zero maxTransactionStart means the caller has no prior read to compare
	// against and opts out of staleness detection. This is correct for domain
	// services that do not yet implement optimistic concurrency control.
	if !maxTransactionStart.IsZero() {
	for typ, entities := range groupsExisting {
		entry := BitemporalTableRegistry[typ]
		tbl := entry.Table
		hasKey := tableHasKeyCol(entry)

		if len(entry.IDColumns) == 1 {
			// Single-ID tables: batch query using ANY($1::int8[]).
			intCol := entry.IDColumns[0].(postgres.ColumnInteger)
			ids := make([]int64, len(entities))
			idToEntity := make(map[int64]LockableEntity, len(entities))
			for i, e := range entities {
				ids[i] = e.EntityID()
				idToEntity[e.EntityID()] = e
			}

			txStartCol := postgres.TimestampzColumn("transaction_start")
			projections := []postgres.Projection{intCol.AS("id"), txStartCol.AS("tx_start")}
			if hasKey {
				projections = append(projections, postgres.StringColumn("key").AS("key"))
			}

			stmt := postgres.SELECT(projections[0], projections[1:]...).
				FROM(tbl).
				WHERE(intCol.EQ(postgres.ANY(postgres.Int64Array(ids...))))

			var dbStates []struct {
				ID      int64     `sql:"id"`
				TxStart time.Time `sql:"tx_start"`
				Key     string    `sql:"key"`
			}
			if err := stmt.QueryContext(ctx, db, &dbStates); err != nil {
				return nil, fmt.Errorf("ExecuteBitemporalMutation: state fetch failed for %s: %w", tbl.Alias(), err)
			}

			maxTxStartMap := make(map[int64]time.Time, len(ids))
			keyMap := make(map[int64]string, len(ids))
			for _, s := range dbStates {
				if s.TxStart.After(maxTxStartMap[s.ID]) {
					maxTxStartMap[s.ID] = s.TxStart
				}
				keyMap[s.ID] = s.Key
			}

			for id, entity := range idToEntity {
				dbMaxTxStart, exists := maxTxStartMap[id]
				if !exists {
					return nil, fmt.Errorf("ExecuteBitemporalMutation: entity %d not found in table %s", id, tbl.Alias())
				}
				if dbMaxTxStart.After(maxTransactionStart) {
					return nil, fmt.Errorf("ExecuteBitemporalMutation: staleness detected in table %s for entity %d (db: %v, known: %v)",
						tbl.Alias(), id, dbMaxTxStart, maxTransactionStart)
				}
				if hasKey && keyMap[id] != entity.NaturalKey() {
					return nil, fmt.Errorf("ExecuteBitemporalMutation: key immutability violation for entity %d (db: %q, known: %q)",
						id, keyMap[id], entity.NaturalKey())
				}
			}
		} else {
			// Composite-ID tables: per-entity queries to avoid cross-entity
			// row confusion when multiple entities share a partial key.
			txStartCol := postgres.TimestampzColumn("transaction_start")
			for _, entity := range entities {
				idCond := buildEntityIDCondition(entry, entity)
				stmt := postgres.SELECT(txStartCol.AS("tx_start")).
					FROM(tbl).WHERE(idCond)

				var dbStates []struct {
					TxStart time.Time `sql:"tx_start"`
				}
				if err := stmt.QueryContext(ctx, db, &dbStates); err != nil {
					return nil, fmt.Errorf("ExecuteBitemporalMutation: state fetch failed for %s: %w", tbl.Alias(), err)
				}
				if len(dbStates) == 0 {
					return nil, fmt.Errorf("ExecuteBitemporalMutation: entity %s not found in table %s", entity.NaturalKey(), tbl.Alias())
				}

				var maxTxStart time.Time
				for _, s := range dbStates {
					if s.TxStart.After(maxTxStart) {
						maxTxStart = s.TxStart
					}
				}
				if maxTxStart.After(maxTransactionStart) {
					return nil, fmt.Errorf("ExecuteBitemporalMutation: staleness detected in table %s for entity %s (db: %v, known: %v)",
						tbl.Alias(), entity.NaturalKey(), maxTxStart, maxTransactionStart)
				}
				// Key immutability: for composite-key tables the PK itself is immutable;
				// there is no separate 'key' column to verify.
			}
		}
	}
	} // end staleness check (skipped when maxTransactionStart is zero)

	// Phase 4: Compute
	if builder == nil {
		builder = DefaultBitemporalInserter[T]
	}
	insertStmts := builder(knownState, effExpr)

	// Phase 5: Write (Sterile Bitemporal Lifecycle)
	//
	// EBM owns time. The pipeline preserves two invariants:
	//   (1) Every active row that this mutation modifies, deletes, or shortens
	//       is first copied into the historical partition with transaction_end
	//       capped at CURRENT_TIMESTAMP, so no superseded belief is destroyed.
	//   (2) No active row carries a transaction_start older than the moment
	//       the row's *current* belief was formed. Consequently TRUNCATE — the
	//       only in-place active mutation EBM still performs — refreshes
	//       transaction_start alongside validity_end.
	//
	// Task 5 (Story043H) evaluation: A Go-Jet INSERT…SELECT for FLUSH was
	// prototyped. The Jet form requires programmatic column-list construction,
	// type assertions for every projection, and separate handling for the two
	// overridden columns (transaction_end→CURRENT_TIMESTAMP, mode→'historical').
	// It is materially less readable than the raw SQL below and offers no
	// additional type-safety benefit once IDColumns and NonMetaColumns are
	// pre-computed in the registry. Raw SQL is retained.
	for typ, entities := range groupsExisting {
		entry, ok := BitemporalTableRegistry[typ]
		if !ok {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: type %v not found in registry during flush phase", typ)
		}
		tbl := entry.Table

		// Build the column name strings from registry metadata.
		idColNames := make([]string, len(entry.IDColumns))
		for i, c := range entry.IDColumns {
			idColNames[i] = c.Name()
		}
		idColsStr := strings.Join(idColNames, ", ")

		nonMetaNames := make([]string, len(entry.NonMetaColumns))
		for i, c := range entry.NonMetaColumns {
			nonMetaNames[i] = c.Name()
		}
		nonMetaColsSuffix := ""
		if len(nonMetaNames) > 0 {
			nonMetaColsSuffix = ", " + strings.Join(nonMetaNames, ", ")
		}

		// 1. FLUSH: Copy every active row whose validity extends past the new
		// effective time into the historical partition. The filter is
		// "validity_end > eff", which captures both rows that will be truncated
		// (validity_start < eff) and any rows that will be purged (validity_start >= eff).
		var flushSQL string
		var flushArgs []any

		if len(entry.IDColumns) == 1 {
			// Single-ID tables: use parameterized ANY($1) for efficiency.
			ids := make([]int64, len(entities))
			for i, e := range entities {
				ids[i] = e.EntityID()
			}
			flushSQL = fmt.Sprintf(`
				INSERT INTO %s (
					%s, validity_start, validity_end, transaction_start, transaction_end, mode, insertion_metadata%s
				)
				OVERRIDING SYSTEM VALUE
				SELECT
					%s, validity_start, validity_end, transaction_start, CURRENT_TIMESTAMP, 'historical', insertion_metadata%s
				FROM %s
				WHERE mode = 'active'
				  AND %s = ANY($1)
				  AND validity_end > $2`,
				tbl.TableName(), idColsStr, nonMetaColsSuffix,
				idColsStr, nonMetaColsSuffix, tbl.TableName(), idColNames[0])
			flushEffective := pgNow
			if effectiveAt != nil {
				flushEffective = *effectiveAt
			}
			flushArgs = []any{pq.Array(ids), flushEffective}
		} else {
			// Composite-ID tables: embed per-entity conditions directly.
			whereID := buildCompositeIDWhereSQL(entry, entities)
			flushSQL = fmt.Sprintf(`
				INSERT INTO %s (
					%s, validity_start, validity_end, transaction_start, transaction_end, mode, insertion_metadata%s
				)
				OVERRIDING SYSTEM VALUE
				SELECT
					%s, validity_start, validity_end, transaction_start, CURRENT_TIMESTAMP, 'historical', insertion_metadata%s
				FROM %s
				WHERE mode = 'active'
				  AND (%s)
				  AND validity_end > $1`,
				tbl.TableName(), idColsStr, nonMetaColsSuffix,
				idColsStr, nonMetaColsSuffix, tbl.TableName(), whereID)
			flushEffective := pgNow
			if effectiveAt != nil {
				flushEffective = *effectiveAt
			}
			flushArgs = []any{flushEffective}
		}

		if _, err := db.ExecContext(ctx, flushSQL, flushArgs...); err != nil {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: flush failed for %s: %w", tbl.Alias(), err)
		}

		// Build the batch ID condition for Go-Jet PURGE and TRUNCATE statements.
		idBatchCond := buildBatchIDCondition(entry, entities)

		startCol := postgres.TimestampzColumn("validity_start")
		endCol := postgres.TimestampzColumn("validity_end")
		txStartCol := postgres.TimestampzColumn("transaction_start")
		modeCol := postgres.StringColumn("mode")

		// 2. PURGE: Delete only active rows that are wholly superseded (validity_start >= eff).
		purgeStmt := tbl.DELETE().
			WHERE(postgres.AND(
				modeCol.EQ(enum.BitemporalMode.Active),
				idBatchCond,
				startCol.GT_EQ(effExpr),
			))

		if _, err := purgeStmt.ExecContext(ctx, db); err != nil {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: purge failed for %s: %w", tbl.Alias(), err)
		}

		// 3. TRUNCATE ACTIVE: Shorten the active row that straddles eff to end at eff,
		// and refresh its transaction_start so the row honestly reports when its current
		// belief was formed. Without the transaction_start refresh, the truncated row and
		// the historical row it superseded would both satisfy the §3 predicate at any U
		// in [old_tx_start, CURRENT_TIMESTAMP), violating bitemporal unicity.
		truncateStmt := tbl.UPDATE(endCol, txStartCol).
			SET(effExpr, postgres.CURRENT_TIMESTAMP()).
			WHERE(postgres.AND(
				modeCol.EQ(enum.BitemporalMode.Active),
				idBatchCond,
				startCol.LT(effExpr),
				endCol.GT(effExpr),
			))

		if _, err := truncateStmt.ExecContext(ctx, db); err != nil {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: truncate failed for %s: %w", tbl.Alias(), err)
		}
	}

	// Phase 5b: Write (Assert) with RETURNING to capture inserted rows
	// Execute insert statements and collect the newly inserted entities by re-querying.
	// The inserted rows now have their database-assigned IDs.
	var result []T
	keyToEntity := make(map[string]T) // Map from key to the latest version of each entity

	for _, stmt := range insertStmts {
		wrapped := overridingInsert{stmt}
		if _, err := wrapped.ExecContext(ctx, db); err != nil {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: assert failed: %w", err)
		}
	}

	// Phase 6: Fetch inserted/mutated entities
	// Query the active partition for rows whose validity_end is past the
	// effective time. T must be qrm-compatible (type/field naming aligned to
	// the table), so we scan directly into []T with no wrapper trickery.
	for typ, entities := range groups {
		entry, ok := BitemporalTableRegistry[typ]
		if !ok {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: type %v not found in registry during fetch phase", typ)
		}
		tbl := entry.Table

		if len(entities) == 0 {
			continue
		}

		stmt, err := selectAll(tbl)
		if err != nil {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: unable to build query for %s: %w", tbl.Alias(), err)
		}

		modeCol := postgres.StringColumn("mode")
		txStartCol := postgres.TimestampzColumn("transaction_start")
		endCol := postgres.TimestampzColumn("validity_end")

		// Exclude rows truncated by this mutation (their validity_end == eff);
		// keep newly inserted rows (validity_end > eff, typically InfinityTime).
		stmt = stmt.WHERE(postgres.AND(
			modeCol.EQ(enum.BitemporalMode.Active),
			endCol.GT(effExpr),
		)).ORDER_BY(txStartCol.DESC())

		var fetched []T
		if err := stmt.QueryContext(ctx, db, &fetched); err != nil {
			return nil, fmt.Errorf("ExecuteBitemporalMutation: result fetch failed for %s: %w", tbl.Alias(), err)
		}

		inputKeys := make(map[string]bool, len(entities))
		for _, e := range entities {
			inputKeys[e.NaturalKey()] = true
		}
		seen := make(map[string]bool, len(fetched))
		for _, f := range fetched {
			key := f.NaturalKey()
			if inputKeys[key] && !seen[key] {
				keyToEntity[key] = f
				seen[key] = true
			}
		}
	}

	// Collect results, preserving the original order from knownState.
	for _, entity := range knownState {
		if f, ok := keyToEntity[entity.NaturalKey()]; ok {
			result = append(result, f)
		}
	}

	return result, nil
}

// DefaultBitemporalInserter is a generic builder that re-asserts the models currently held
// in the Go application as the new 'Active' truth.
//
// New entities (EntityID==0) omit identity columns from the INSERT column list so the
// database sequence generates the ID. Existing entities (EntityID!=0) include identity
// columns and rely on the OVERRIDING SYSTEM VALUE wrapper in Phase 5b to preserve them.
// Mode and TransactionEnd are always excluded; they are controlled by database defaults.
func DefaultBitemporalInserter[T LockableEntity](freshState []T, effectiveAt postgres.Expression) []postgres.InsertStatement {
	if len(freshState) == 0 {
		return nil
	}

	typ := reflect.TypeOf(freshState[0])
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	entry, found := BitemporalTableRegistry[typ]
	if !found {
		panic(fmt.Sprintf("DefaultBitemporalInserter: type %v not found in BitemporalTableRegistry", typ))
	}

	tbl := entry.Table

	// Extract AllColumns from the table struct via reflection.
	var allCols postgres.ColumnList
	vTbl := reflect.ValueOf(tbl)
	if vTbl.Kind() == reflect.Ptr {
		vTbl = vTbl.Elem()
	}
	if f := vTbl.FieldByName("AllColumns"); f.IsValid() {
		if cl, ok := f.Interface().(postgres.ColumnList); ok {
			allCols = cl
		}
	}

	// Identify the identity/primary-key column names so they can be excluded for new entities.
	idColNames := make(map[string]bool, len(entry.IDColumns))
	for _, c := range entry.IDColumns {
		idColNames[c.Name()] = true
	}

	// Build two column lists:
	//   insertColsWithID    — used for existing entities; includes ID cols (preserved via OVERRIDING SYSTEM VALUE)
	//   insertColsWithoutID — used for new entities; omits ID cols so the DB sequence generates them
	// Both lists exclude mode and transaction_end (controlled by DB defaults).
	excludedAlways := map[string]bool{
		"mode":            true,
		"transaction_end": true,
	}
	var insertColsWithID, insertColsWithoutID postgres.ColumnList
	for _, col := range allCols {
		name := col.Name()
		if excludedAlways[name] {
			continue
		}
		insertColsWithID = append(insertColsWithID, col)
		if !idColNames[name] {
			insertColsWithoutID = append(insertColsWithoutID, col)
		}
	}

	// Partition freshState into new (ID=0) and existing (ID!=0) entities.
	// Entities are passed directly to Jet's INSERT … MODELS; promoted fields
	// from embedded mixins surface via reflect.FieldByName traversal.
	var newModels, existingModels []any
	for _, entity := range freshState {
		if entity.EntityID() == 0 {
			newModels = append(newModels, entity)
		} else {
			existingModels = append(existingModels, entity)
		}
	}

	var stmts []postgres.InsertStatement

	// New entities: omit identity columns; the DB sequence assigns the ID.
	for i := 0; i < len(newModels); i += 1000 {
		end := i + 1000
		if end > len(newModels) {
			end = len(newModels)
		}
		stmts = append(stmts, tbl.INSERT(insertColsWithoutID).MODELS(newModels[i:end]))
	}

	// Existing entities: include identity columns; OVERRIDING SYSTEM VALUE in Phase 5b preserves them.
	for i := 0; i < len(existingModels); i += 1000 {
		end := i + 1000
		if end > len(existingModels) {
			end = len(existingModels)
		}
		stmts = append(stmts, tbl.INSERT(insertColsWithID).MODELS(existingModels[i:end]))
	}

	return stmts
}

// valuesClauseRE matches the VALUES keyword that introduces a row list in an INSERT
// statement. It requires whitespace before and an opening paren after so that an
// identifier or constant coincidentally containing the substring "values" cannot
// false-match (e.g., a column named "values_count" or a string literal 'values').
//
// Story043H Task 6: This regex replaces the previous strings.LastIndex(upper,"VALUES")
// pivot, which would misfire if any column name or constant value contained "VALUES".
var valuesClauseRE = regexp.MustCompile(`(?i)\sVALUES\s*\(`)

type overridingInsert struct {
	postgres.InsertStatement
}

func (s overridingInsert) Sql() (string, []interface{}) {
	query, args := s.InsertStatement.Sql()
	loc := valuesClauseRE.FindStringIndex(query)
	if loc == nil {
		return query, args
	}
	// loc[0] is the position of the leading whitespace before VALUES.
	// Insert "OVERRIDING SYSTEM VALUE " before the VALUES keyword (i.e., after the space).
	pivot := loc[0] + 1 // skip the leading whitespace character
	return query[:pivot] + "OVERRIDING SYSTEM VALUE " + query[pivot:], args
}

func (s overridingInsert) ExecContext(ctx context.Context, db Executor) (sql.Result, error) {
	query, args := s.Sql()
	return db.ExecContext(ctx, query, args...)
}

// resolveEntityID looks up the internal integer ID for a single-column-key entity by its
// natural key string. It performs two passes to distinguish active, tombstoned, and absent
// entities, enabling callers to make correct decisions during reactivation workflows.
//
// The table must have a single IDColumn (integer) and a 'key' column.
// Pass 1: active partition (mode='active', transaction_end = 3000-01-01 00:00:00 UTC, validity_end > now),
//
//	ORDER BY validity_start DESC, LIMIT 1.
//
// Pass 2: historical partition (mode='historical'), ORDER BY transaction_end DESC,
//
//	validity_end DESC, LIMIT 1.
func resolveEntityID(ctx context.Context, db Executor, tbl postgres.Table, key string) (int64, ResolveStatus, error) {
	// Locate the registry entry so we use the correct ID column name.
	var entry BitemporalRegistryEntry
	var found bool
	for _, e := range BitemporalTableRegistry {
		if e.Table.TableName() == tbl.TableName() && e.Table.Alias() == tbl.Alias() {
			entry = e
			found = true
			break
		}
	}
	if !found {
		return 0, ResolveNotFound, fmt.Errorf("ResolveEntityID: table %s.%s not found in BitemporalTableRegistry", tbl.TableName(), tbl.Alias())
	}
	if len(entry.IDColumns) != 1 {
		return 0, ResolveNotFound, fmt.Errorf("ResolveEntityID: table %s has %d IDColumns; only single-column-key tables are supported", tbl.TableName(), len(entry.IDColumns))
	}

	idColName := entry.IDColumns[0].Name()
	keyCol := postgres.StringColumn("key")
	txEndCol := postgres.TimestampzColumn("transaction_end")
	modeCol := postgres.StringColumn("mode")
	validityEndCol := postgres.TimestampzColumn("validity_end")
	validityStartCol := postgres.TimestampzColumn("validity_start")

	var dest struct {
		ID int64 `sql:"id"`
	}

	// Pass 1: active partition only.
	activeStmt := postgres.SELECT(postgres.IntegerColumn(idColName).AS("id")).
		FROM(tbl).
		WHERE(postgres.AND(
			keyCol.EQ(postgres.String(key)),
			modeCol.EQ(enum.BitemporalMode.Active),
			txEndCol.EQ(postgres.TimestampzT(c.InfinityTime)),
			validityEndCol.GT(postgres.NOW()),
		)).
		ORDER_BY(validityStartCol.DESC()).
		LIMIT(1)

	err := activeStmt.QueryContext(ctx, db, &dest)
	if err == nil {
		return dest.ID, ResolveActive, nil
	}
	if err != sql.ErrNoRows && !strings.Contains(err.Error(), "no rows in result set") {
		return 0, ResolveNotFound, err
	}

	// Pass 2: historical partition.
	txEndCol2 := postgres.TimestampzColumn("transaction_end")
	validityEndCol2 := postgres.TimestampzColumn("validity_end")

	historicalStmt := postgres.SELECT(postgres.IntegerColumn(idColName).AS("id")).
		FROM(tbl).
		WHERE(postgres.AND(
			keyCol.EQ(postgres.String(key)),
			modeCol.EQ(enum.BitemporalMode.Historical),
		)).
		ORDER_BY(txEndCol2.DESC(), validityEndCol2.DESC()).
		LIMIT(1)

	dest.ID = 0
	err = historicalStmt.QueryContext(ctx, db, &dest)
	if err == nil {
		return dest.ID, ResolveTombstoned, nil
	}
	if err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
		return 0, ResolveNotFound, nil
	}
	return 0, ResolveNotFound, err
}

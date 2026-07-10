// Package rdbms provides the relational-database serde layer for ChronoCone.
//
// # Registry design
//
// BitemporalTableRegistry maps entity Go types to the metadata EBM needs to
// orchestrate bitemporal mutations: the Go-Jet table, the advisory lock
// type, the ID columns, and the precomputed non-meta column list used by
// FLUSH.
//
// Entries for tables that have a corresponding domain entity are registered
// by that entity's package via RegisterEntity (called from an init()). The
// rdbms package itself pre-registers generated-model types (model.*) for
// tables that don't yet have domain entities — keeping EBM's unit tests
// usable without forcing a domain entity to exist for every table.
//
// Entities (whether domain or generated) MUST satisfy Go-Jet's qrm naming
// discipline: the Go type name normalizes to the table name and field
// names normalize to column names. Embedded persistence mixins carry an
// `alias:"<table_name>"` tag so qrm rebinds column-prefix lookup to the
// embedding entity's table.
package rdbms

import (
	"reflect"
	"strings"

	"chronocone/backend/serde/rdbms/gen/postgres/public/model"
	"chronocone/backend/serde/rdbms/gen/postgres/public/table"

	"github.com/go-jet/jet/v2/postgres"
)

// AdvisoryLockTableType is a strongly typed enumeration for PostgreSQL advisory lock table types.
// It uses int16 to strictly bound it to the 10-bit packing convention (0-1023).
type AdvisoryLockTableType int16

const (
	TableTypeLanguage                    AdvisoryLockTableType = 1
	TableTypeStorageConfig               AdvisoryLockTableType = 2
	TableTypeStorageConfigLocalization   AdvisoryLockTableType = 3
	TableTypeNamespaceFamily             AdvisoryLockTableType = 4
	TableTypeNamespaceFamilyLocalization AdvisoryLockTableType = 5
	TableTypeNamespace                   AdvisoryLockTableType = 6
	TableTypeNamespaceLocalization       AdvisoryLockTableType = 7
	TableTypeCategory                    AdvisoryLockTableType = 8
	TableTypeCategoryLocalization        AdvisoryLockTableType = 9
	TableTypeCategoryShape               AdvisoryLockTableType = 10
	TableTypeCategoryShapeLocalization   AdvisoryLockTableType = 11
	TableTypeCategoryAttr                AdvisoryLockTableType = 12
	TableTypeCategoryAttrLocalization    AdvisoryLockTableType = 13
	TableTypeElement                     AdvisoryLockTableType = 14
	TableTypeElementLocalization         AdvisoryLockTableType = 15
	TableTypeElementAttr                 AdvisoryLockTableType = 16
	TableTypeElementAttrLocalization     AdvisoryLockTableType = 17
	TableTypeMetric                      AdvisoryLockTableType = 18
	TableTypeMetricLocalization          AdvisoryLockTableType = 19
	TableTypeMetricDim                   AdvisoryLockTableType = 20
	TableTypeTransformer                 AdvisoryLockTableType = 21
	TableTypeTransformerLocalization     AdvisoryLockTableType = 22
	TableTypeTransformerInput            AdvisoryLockTableType = 23
	TableTypeTransformerOutput           AdvisoryLockTableType = 24
	TableTypeShifter                     AdvisoryLockTableType = 25
	TableTypeShifterLocalization         AdvisoryLockTableType = 26
	TableTypeShifterInput                AdvisoryLockTableType = 27
	TableTypeShifterOutput               AdvisoryLockTableType = 28
	TableTypeTsDocType                   AdvisoryLockTableType = 29
	TableTypeTsDocTypeLocalization       AdvisoryLockTableType = 30
	TableTypeTsDocRowType                AdvisoryLockTableType = 31
	TableTypeTsDocRowTypeLocalization    AdvisoryLockTableType = 32
	TableTypeTsDocLocalization           AdvisoryLockTableType = 33
	TableTypeNamespaceFamilyTypeLoc      AdvisoryLockTableType = 34
	TableTypeNamespaceTypeLoc            AdvisoryLockTableType = 35
)

// BitemporalRegistryEntry holds the metadata required for bitemporal orchestration.
//
//   - Table is the Go-Jet table value used to emit SQL.
//   - LockType identifies the table for advisory-lock key packing.
//   - IDColumns is the ordered list of columns that uniquely identify an entity
//     instance. Single column for simple entity tables; two or more for junction
//     and localization tables that carry a composite primary key.
//   - NonMetaColumns is the precomputed projection of columns that are neither
//     identity columns nor bitemporal meta-columns (validity_*, transaction_*,
//     mode, insertion_metadata). Used by the FLUSH step to copy business payload.
type BitemporalRegistryEntry struct {
	Table          postgres.Table
	LockType       AdvisoryLockTableType
	IDColumns      []postgres.Column
	NonMetaColumns postgres.ColumnList
}

// bitemporalMetaColNames is the set of column names that are always excluded from NonMetaColumns.
var bitemporalMetaColNames = map[string]bool{
	"validity_start":     true,
	"validity_end":       true,
	"transaction_start":  true,
	"transaction_end":    true,
	"mode":               true,
	"insertion_metadata": true,
}

// computeNonMetaCols derives the NonMetaColumns list for a table entry by reading AllColumns
// from the generated table struct via reflection and excluding the ID columns and the six
// standard bitemporal meta-columns.
func computeNonMetaCols(tbl postgres.Table, idCols []postgres.Column) postgres.ColumnList {
	idNames := make(map[string]bool, len(idCols))
	for _, c := range idCols {
		idNames[c.Name()] = true
	}

	v := reflect.ValueOf(tbl)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	f := v.FieldByName("AllColumns")
	if !f.IsValid() {
		return nil
	}
	all, ok := f.Interface().(postgres.ColumnList)
	if !ok {
		return nil
	}

	var out postgres.ColumnList
	for _, c := range all {
		n := c.Name()
		if !idNames[n] && !bitemporalMetaColNames[n] {
			out = append(out, c)
		}
	}
	return out
}

func registryEntry(tbl postgres.Table, lockType AdvisoryLockTableType, idCols ...postgres.Column) BitemporalRegistryEntry {
	return BitemporalRegistryEntry{
		Table:          tbl,
		LockType:       lockType,
		IDColumns:      idCols,
		NonMetaColumns: computeNonMetaCols(tbl, idCols),
	}
}

// BitemporalTableRegistry maps entity Go types to their orchestration metadata.
//
// Pre-populated with generated-model types so EBM's own unit tests can mutate
// any bitemporal table directly. Domain entities register additional entries
// via RegisterEntity from their package init().
var BitemporalTableRegistry = buildRegistry()

// RegisterEntity registers a Go entity type with the bitemporal registry.
// Call from a domain package's init() to make EBM aware of a domain entity
// type. Returns silently if the type is already registered (idempotent).
func RegisterEntity(entityType reflect.Type, tbl postgres.Table, lockType AdvisoryLockTableType, idCols ...postgres.Column) {
	BitemporalTableRegistry[entityType] = registryEntry(tbl, lockType, idCols...)
}

func buildRegistry() map[reflect.Type]BitemporalRegistryEntry {
	return map[reflect.Type]BitemporalRegistryEntry{
		// Section 1: Core
		//
		// Language uses a composite natural key (language, country) — no integer surrogate.
		reflect.TypeOf(model.Language{}): registryEntry(
			table.Language, TableTypeLanguage,
			table.Language.Language, table.Language.Country,
		),
		reflect.TypeOf(model.StorageConfig{}): registryEntry(
			table.StorageConfig, TableTypeStorageConfig,
			table.StorageConfig.StorageConfigID,
		),
		reflect.TypeOf(model.StorageConfigLocalization{}): registryEntry(
			table.StorageConfigLocalization, TableTypeStorageConfigLocalization,
			table.StorageConfigLocalization.StorageConfigID,
			table.StorageConfigLocalization.Language,
			table.StorageConfigLocalization.Country,
		),

		// Section 2: Policy & Identity
		reflect.TypeOf(model.NamespaceFamilyTypeLocalization{}): registryEntry(
			table.NamespaceFamilyTypeLocalization, TableTypeNamespaceFamilyTypeLoc,
			table.NamespaceFamilyTypeLocalization.NsFamTypeID,
			table.NamespaceFamilyTypeLocalization.Language,
			table.NamespaceFamilyTypeLocalization.Country,
		),
		reflect.TypeOf(model.NamespaceTypeLocalization{}): registryEntry(
			table.NamespaceTypeLocalization, TableTypeNamespaceTypeLoc,
			table.NamespaceTypeLocalization.NsTypeID,
			table.NamespaceTypeLocalization.Language,
			table.NamespaceTypeLocalization.Country,
		),
		reflect.TypeOf(model.NamespaceFamily{}): registryEntry(
			table.NamespaceFamily, TableTypeNamespaceFamily,
			table.NamespaceFamily.NsFamID,
		),
		reflect.TypeOf(model.NamespaceFamilyLocalization{}): registryEntry(
			table.NamespaceFamilyLocalization, TableTypeNamespaceFamilyLocalization,
			table.NamespaceFamilyLocalization.NsFamID,
			table.NamespaceFamilyLocalization.Language,
			table.NamespaceFamilyLocalization.Country,
		),
		reflect.TypeOf(model.Namespace{}): registryEntry(
			table.Namespace, TableTypeNamespace,
			table.Namespace.NsID,
		),
		reflect.TypeOf(model.NamespaceLocalization{}): registryEntry(
			table.NamespaceLocalization, TableTypeNamespaceLocalization,
			table.NamespaceLocalization.NsID,
			table.NamespaceLocalization.Language,
			table.NamespaceLocalization.Country,
		),

		// Section 3: Taxonomy
		reflect.TypeOf(model.Category{}): registryEntry(
			table.Category, TableTypeCategory,
			table.Category.CategoryID,
		),
		reflect.TypeOf(model.CategoryLocalization{}): registryEntry(
			table.CategoryLocalization, TableTypeCategoryLocalization,
			table.CategoryLocalization.CategoryID,
			table.CategoryLocalization.Language,
			table.CategoryLocalization.Country,
		),
		reflect.TypeOf(model.CategoryShape{}): registryEntry(
			table.CategoryShape, TableTypeCategoryShape,
			table.CategoryShape.CategoryShapeID,
		),
		reflect.TypeOf(model.CategoryShapeLocalization{}): registryEntry(
			table.CategoryShapeLocalization, TableTypeCategoryShapeLocalization,
			table.CategoryShapeLocalization.CategoryShapeID,
			table.CategoryShapeLocalization.Language,
			table.CategoryShapeLocalization.Country,
		),
		reflect.TypeOf(model.CategoryAttr{}): registryEntry(
			table.CategoryAttr, TableTypeCategoryAttr,
			table.CategoryAttr.CategoryAttrID,
		),
		reflect.TypeOf(model.CategoryAttrLocalization{}): registryEntry(
			table.CategoryAttrLocalization, TableTypeCategoryAttrLocalization,
			table.CategoryAttrLocalization.CategoryAttrID,
			table.CategoryAttrLocalization.Language,
			table.CategoryAttrLocalization.Country,
		),
		reflect.TypeOf(model.Element{}): registryEntry(
			table.Element, TableTypeElement,
			table.Element.ElementID,
		),
		reflect.TypeOf(model.ElementLocalization{}): registryEntry(
			table.ElementLocalization, TableTypeElementLocalization,
			table.ElementLocalization.ElementID,
			table.ElementLocalization.Language,
			table.ElementLocalization.Country,
		),
		reflect.TypeOf(model.ElementAttr{}): registryEntry(
			table.ElementAttr, TableTypeElementAttr,
			table.ElementAttr.ElementAttrID,
		),
		reflect.TypeOf(model.ElementAttrLocalization{}): registryEntry(
			table.ElementAttrLocalization, TableTypeElementAttrLocalization,
			table.ElementAttrLocalization.ElementAttrID,
			table.ElementAttrLocalization.Language,
			table.ElementAttrLocalization.Country,
		),

		// Section 4: Morphisms
		reflect.TypeOf(model.Metric{}): registryEntry(
			table.Metric, TableTypeMetric,
			table.Metric.MetricID,
		),
		reflect.TypeOf(model.MetricLocalization{}): registryEntry(
			table.MetricLocalization, TableTypeMetricLocalization,
			table.MetricLocalization.MetricID,
			table.MetricLocalization.Language,
			table.MetricLocalization.Country,
		),
		// MetricDim is a junction table with composite key (metric_id, dimension_id).
		reflect.TypeOf(model.MetricDim{}): registryEntry(
			table.MetricDim, TableTypeMetricDim,
			table.MetricDim.MetricID, table.MetricDim.DimensionID,
		),
		reflect.TypeOf(model.Transformer{}): registryEntry(
			table.Transformer, TableTypeTransformer,
			table.Transformer.TransformerID,
		),
		reflect.TypeOf(model.TransformerLocalization{}): registryEntry(
			table.TransformerLocalization, TableTypeTransformerLocalization,
			table.TransformerLocalization.TransformerID,
			table.TransformerLocalization.Language,
			table.TransformerLocalization.Country,
		),
		// TransformerInput is a junction table with composite key (transformer_id, metric_id).
		reflect.TypeOf(model.TransformerInput{}): registryEntry(
			table.TransformerInput, TableTypeTransformerInput,
			table.TransformerInput.TransformerID, table.TransformerInput.MetricID,
		),
		// TransformerOutput is a junction table with composite key (transformer_id, metric_id).
		reflect.TypeOf(model.TransformerOutput{}): registryEntry(
			table.TransformerOutput, TableTypeTransformerOutput,
			table.TransformerOutput.TransformerID, table.TransformerOutput.MetricID,
		),
		reflect.TypeOf(model.Shifter{}): registryEntry(
			table.Shifter, TableTypeShifter,
			table.Shifter.ShifterID,
		),
		reflect.TypeOf(model.ShifterLocalization{}): registryEntry(
			table.ShifterLocalization, TableTypeShifterLocalization,
			table.ShifterLocalization.ShifterID,
			table.ShifterLocalization.Language,
			table.ShifterLocalization.Country,
		),
		// ShifterInput is a junction table with composite key (shifter_id, metric_id).
		reflect.TypeOf(model.ShifterInput{}): registryEntry(
			table.ShifterInput, TableTypeShifterInput,
			table.ShifterInput.ShifterID, table.ShifterInput.MetricID,
		),
		// ShifterOutput is a junction table with composite key (shifter_id, metric_id).
		reflect.TypeOf(model.ShifterOutput{}): registryEntry(
			table.ShifterOutput, TableTypeShifterOutput,
			table.ShifterOutput.ShifterID, table.ShifterOutput.MetricID,
		),

		// Section 5: Timeseries
		reflect.TypeOf(model.TsDocType{}): registryEntry(
			table.TsDocType, TableTypeTsDocType,
			table.TsDocType.TsDocTypeID,
		),
		reflect.TypeOf(model.TsDocTypeLocalization{}): registryEntry(
			table.TsDocTypeLocalization, TableTypeTsDocTypeLocalization,
			table.TsDocTypeLocalization.TsDocTypeID,
			table.TsDocTypeLocalization.Language,
			table.TsDocTypeLocalization.Country,
		),
		reflect.TypeOf(model.TsDocRowType{}): registryEntry(
			table.TsDocRowType, TableTypeTsDocRowType,
			table.TsDocRowType.TsDocRowTypeID,
		),
		reflect.TypeOf(model.TsDocRowTypeLocalization{}): registryEntry(
			table.TsDocRowTypeLocalization, TableTypeTsDocRowTypeLocalization,
			table.TsDocRowTypeLocalization.TsDocRowTypeID,
			table.TsDocRowTypeLocalization.Language,
			table.TsDocRowTypeLocalization.Country,
		),
		reflect.TypeOf(model.TsDocLocalization{}): registryEntry(
			table.TsDocLocalization, TableTypeTsDocLocalization,
			table.TsDocLocalization.TsDocID,
			table.TsDocLocalization.Language,
			table.TsDocLocalization.Country,
		),
	}
}

// normalizeColName converts column names to lowercase without underscores for matching
// struct field names: "metric_id" → "metricid" matches struct field MetricID → "metricid".
// Used by buildEntityIDCondition to extract composite-key values from model structs.
func normalizeColName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "_", ""))
}

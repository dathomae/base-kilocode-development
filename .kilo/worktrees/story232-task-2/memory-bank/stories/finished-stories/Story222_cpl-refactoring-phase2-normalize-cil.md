# Story222: CPL Refactoring — Phase 2 Normalize CIL

## Goal

Unify the two competing CIL representations and replace `interface{}` bridges between parse and analysis with typed structures. This eliminates the half-migrated state and restores compile-time type safety across phase boundaries.

This story uses the pre-completed analysis from [revampstory222_plan.md](../plans/revampstory222_plan.md) as its authoritative reference. That plan catalogued every old CIL reference and every type-switch site, so this story contains only concrete implementation tasks — no investigative or design subtasks.

## References

- [CPL Refactoring Plan](../plans/cpl_refactoring_plan.md)
- [Revamp Story222 Plan](../plans/revampstory222_plan.md) — authoritative reference for exact file/line changes
- [Story221: Phase 1 Cleanup](../stories/finished-stories/Story221_cpl-refactoring-phase1-cleanup.md)

## Dependencies

- **[Story221_cpl-refactoring-phase1-cleanup.md](../stories/finished-stories/Story221_cpl-refactoring-phase1-cleanup.md)** — Done. Phase 1 removed dead visitor methods from `CILAbstractTransformer` and simplified the CIL tree walk.

## Dependent Stories

- **[Story223_cpl-refactoring-phase3-split-bloat.md](Story223_cpl-refactoring-phase3-split-bloat.md)** — Not Started

## Tasks

### Task Sizing

Tasks 1 and 3 are single-file, additive or deletion-only changes — annotated `[Small]`. Tasks 2, 4, and 5 introduce coordinated multi-file structural changes across packages and introduce new interfaces or type bridges — routed to `code-for-story-implementor`.

| Task | Size | Rationale |
|------|------|-----------|
| 1 — Define Connector CIL Node | `[Small]` | Creates 1 file (`cil/cil_connector.go`), no existing file modifications, purely additive |
| 2 — Migrate Connectors and CILTrees | Medium | Modifies 7+ files across `analysis/` and `cil/`, 4 test files, coordinated type changes |
| 3 — Remove Dead Node Types | `[Small]` | Deletes 1 file (`src/main/go/analysis/cil_nodes.go`), single mechanical operation after precondition verification |
| 4 — Define Typed Postfix Expression | Medium | Creates 1 file, modifies 3 existing files across `parsing/`, `analysis/`, `parser/`, introduces new interface + 3 structs |
| 5 — Typed SymbolDefinition.Data | Medium | Creates 1 file, modifies ~8 files across `symboltable/`, `parsing/`, `analysis/`, `parser/`, coordinated type migration |

### Parallel Execution

- **Group A (Tasks 1, 4):** Task 1 creates `cil/cil_connector.go` (purely additive in `cil/` package). Task 4 creates `src/main/go/parsing/postfix_element.go` and modifies `src/main/go/parsing/parser_metric_definition.go`, `src/main/go/analysis/metrics_analysis.go`, and `src/main/go/parser/cpl_metrics_listener.go`. Completely disjoint packages with no shared files or infrastructure.

### Execution Order

```
[Task 1 ‖ Task 4] → Task 2 → Task 3 → Task 5
```

- Tasks 1 and 4 run in parallel — independent work in disjoint packages
- Task 2 depends on Task 1 (uses `cil.CILConnectorNode` created in Task 1)
- Task 3 depends on Task 2 (deletes `src/main/go/analysis/cil_nodes.go` only after all consumers are migrated)
- Task 5 depends on Task 4 (modifies `symbolDefToPostfixElem` which Task 4 also changes)

### Worktree Suitability

- **Group A (Tasks 1, 4):** — fully disjoint packages (`cil/` vs `parsing/`/`analysis/`/`parser/`), no shared infrastructure, safe for in-process subagents

### Worktree Reintegration

None — no worktree groups.

---

### Task 1: Define Connector CIL Node in the `cil` Package — Completed `[Small]`

**Goal**: Create `cil.CILConnectorNode` to replace `analysis.CILConnectorReferenceNode`. This struct implements `cil.CILNode` (the `cilNode()` marker method already defined in `src/main/go/cil/cil_nodes.go`) and serves as a leaf node in CIL expression trees representing an input data source.

**Test-First Development**:

a. Create `src/main/go/cil/cil_connector.go` with the following content:

```go
package cil

// CILConnectorNode represents a reference to a connector by name.
// It implements the CILNode interface and serves as a leaf node
// in CIL expression trees, representing an input data source.
type CILConnectorNode struct {
    ConnectorName string
}

func (c *CILConnectorNode) cilNode() {}
```

b. Write a test in `src/main/go/cil/cil_connector_test.go` (or extend an existing `_test.go` file in `cil/`) verifying:
   - Construction: `node := &cil.CILConnectorNode{ConnectorName: "TestConnector"}`
   - Interface satisfaction: assign to `var _ cil.CILNode = node`
   - ConnectorName field access: `node.ConnectorName == "TestConnector"`
   - Zero value: `(cil.CILConnectorNode{}).ConnectorName == ""`

c. Run `go test ./src/main/go/cil/...` — must pass with zero failures.

**No other files are modified in this task.**

**Acceptance Criteria**:
- `src/main/go/cil/cil_connector.go` exists with `CILConnectorNode` struct and `cilNode()` method
- `CILConnectorNode` satisfies `cil.CILNode` interface
- `go test ./src/main/go/cil/...` passes

---

### Task 2: Migrate Connectors and CILTrees to the Unified Representation — Completed

**Goal**: Update all production and test code to use `cil.CILConnectorNode` and `cil.CILNode` instead of the old `analysis.CILAbstractNode`-based types. The `CILTrees` map in `AnalysisDefinitions` transitions from `map[string]CILAbstractNode` to `map[string]cil.CILNode`.

The `CILTrees` map is populated by connectors but never read by production synthesis code (synthesis reads `CILRoot` from individual metric definitions). It is consumed only in tests. This makes the migration low-risk.

**NOT modified in this task**: `src/main/go/analysis/cil_nodes.go` (deleted in Task 3).

**Test-First Development**:

a. Update **test files** to expect new types:

   1. **`src/main/go/analysis/cpl_analyzer_connectors_test.go`** — replace all `map[string]analysis.CILAbstractNode` with `map[string]cil.CILNode`, replace `analysis.NewCILConnectorReferenceNode("...")` with `&cil.CILConnectorNode{ConnectorName: "..."}`, replace `analysis.CILNodeTypeConnectorReference` assertions with `!= nil` check and type-assert to `*cil.CILConnectorNode`. Affected lines: 35, 55, 88, 137, 163, 183, 187, 190–199, 205, 215, 219–222, 227, 231–234, 249–253, 293–296, 318–326, 370–374, 380–382, 406. Ensure the `cil` import is present.

   2. **`src/main/go/analysis/connectors_scope_test.go`** — replace `cilNode.(*analysis.CILConnectorReferenceNode)` type assertion with `cilNode.(*cil.CILConnectorNode)` and `require.True(t, ok, "Should be CILConnectorNode")`. Remove `assert.Equal(t, analysis.CILNodeTypeConnectorReference, ...)` assertions. Ensure `CILTrees` map access compiles (value type changes from `CILAbstractNode` to `cil.CILNode`).

   3. **`src/main/go/analysis/analyze_connectors_integration_test.go`** — same type assertion updates as `connectors_scope_test.go` (lines 160–167). Ensure `CILTrees` map assertions compile.

   4. **`src/main/go/analysis/metrics_analysis_test.go`** — update any `CILTrees` map access assertions (lines 604–605, 676–681). The value type changes from `CILAbstractNode` to `cil.CILNode`; type assertions must match.

b. Update **production code** to produce new types:

   1. **`src/main/go/analysis/connectors_analysis.go`** (line 256):
      - Change `analyzer.cilTrees[connectorName] = NewCILConnectorReferenceNode(connectorName)` to `analyzer.cilTrees[connectorName] = &cil.CILConnectorNode{ConnectorName: connectorName}`
      - Add `"github.com/chronocone/chronocone_planning_language/src/main/go/cil"` to the import block
      - Update comments referencing `CILConnectorReferenceNode` to `CILConnectorNode`

   2. **`src/main/go/analysis/cpl_analyzer.go`**:
      - Line 150: `CILTrees map[string]CILAbstractNode` → `CILTrees map[string]cil.CILNode`
      - Line 224: `cilTrees map[string]CILAbstractNode` → `cilTrees map[string]cil.CILNode`
      - Line 255: `make(map[string]CILAbstractNode)` → `make(map[string]cil.CILNode)`
      - Line 285: No change needed (just copies the map)
      - Add `cil` import (check existing imports; may already be present)
      - Update doc comment references from `CILConnectorReferenceNode` to `CILConnectorNode`

   3. **`src/main/go/cil/cil_transformer.go`** (lines 35, 87): Update comments referencing `CILConnectorReferenceNode` to `CILConnectorNode`.

c. Run `go build ./src/main/go/...` — must compile without errors.

d. Run `go test ./src/main/go/analysis/... ./cil/...` — must pass with zero failures.

**Acceptance Criteria**:
- `src/main/go/analysis/connectors_analysis.go` uses `&cil.CILConnectorNode{...}` instead of `NewCILConnectorReferenceNode(...)`
- `src/main/go/analysis/cpl_analyzer.go` uses `map[string]cil.CILNode` for both `CILTrees` and `cilTrees`
- All test files reference `cil.CILNode` and `*cil.CILConnectorNode`, never `analysis.CILAbstractNode` or `*analysis.CILConnectorReferenceNode`
- `src/main/go/analysis/cil_nodes.go` still exists (deleted in Task 3)
- `go build ./src/main/go/...` and `go test ./src/main/go/analysis/... ./cil/...` pass

---

### Task 3: Remove Dead Node Types from `src/main/go/analysis/cil_nodes.go` — Completed `[Small]`

**Goal**: Delete `src/main/go/analysis/cil_nodes.go` after confirming zero remaining references to `CILAbstractNode`, `CILNodeType`, and `CILConnectorReferenceNode`.

**Test-First Development**:

a. Run the following precondition check:

```bash
grep -rn 'CILAbstractNode\|CILNodeType\|CILConnectorReferenceNode' \
  src/chronocone_planning_language/src/main/go/src/main/go/analysis/cil_nodes.go
```

Then verify zero matches EXCEPT `src/main/go/analysis/cil_nodes.go` itself and documentation/specification files under `memory-bank/`:

```bash
grep -rn 'CILAbstractNode\|CILNodeType\|CILConnectorReferenceNode' \
  src/chronocone_planning_language/src/main/go/ --include='*.go'
```

Expected: only `src/main/go/analysis/cil_nodes.go` should match.

b. Delete the file `src/main/go/analysis/cil_nodes.go`.

c. Verify the `cil` import in `src/main/go/analysis/connectors_analysis.go` is still needed (it is — for `cil.CILConnectorNode`).

d. Run `go build ./src/main/go/...` — must compile.

e. Run `go test ./src/main/go/...` — must pass with zero failures.

**Acceptance Criteria**:
- `src/main/go/analysis/cil_nodes.go` is deleted
- `go build ./src/main/go/...` compiles without errors
- `go test ./src/main/go/...` passes

---

### Task 4: Define Typed Postfix Expression — Completed

**Goal**: Replace `[]interface{}` postfix elements with typed `parsing.PostfixElement` structures, eliminating the type-switch on `interface{}` in `buildCILTreeFromPostfix`.

The postfix flow has two entry points:
- **Path A**: `ParserMetricDefinition.CalculatePostfix` (populated by `collectCalculatePostfix` in `src/main/go/parser/cpl_metrics_listener.go`) → `buildCILTree` → `buildCILTreeFromPostfix`
- **Path B**: `ClauseData.ConditionPostfix` (already `[]SymbolDefinition`) → `symbolDefToPostfixElem` → `buildCILTreeFromSymbolDefs` → `buildCILTreeFromPostfix`

Both paths converge on `buildCILTreeFromPostfix`. Task 2a makes that function operate on `[]parsing.PostfixElement`.

**Test-First Development**:

a. Create `src/main/go/src/main/go/parsing/postfix_element.go`:

```go
package parsing

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// PostfixElement is the interface for typed postfix expression elements.
type PostfixElement interface {
    isPostfixElement()
}

// PostfixLiteral represents a numeric literal (int, int64, or float64).
type PostfixLiteral struct {
    Value float64
}
func (PostfixLiteral) isPostfixElement() {}

// PostfixIdentifier represents a named reference (metric, dimension, connector).
type PostfixIdentifier struct {
    Name string
}
func (PostfixIdentifier) isPostfixElement() {}

// PostfixOperator represents an operator or boolean/nil literal carried as a DefinitionType.
type PostfixOperator struct {
    DefType symboltable.DefinitionType
}
func (PostfixOperator) isPostfixElement() {}
```

b. Write tests in `src/main/go/parsing/postfix_element_test.go`:
   - Interface satisfaction for each concrete type: `var _ PostfixElement = PostfixLiteral{}`
   - Construction and field access for each type
   - A sample postfix expression construction (mix of literals, identifiers, and operators)

c. Run `go test ./src/main/go/parsing/...` — must pass.

d. Update **`src/main/go/parsing/parser_metric_definition.go`** (line 22):
   - Change `CalculatePostfix []interface{}` to `CalculatePostfix []PostfixElement`
   - Update the doc comment referencing `interface{}` to `PostfixElement`

e. Update **`src/main/go/analysis/metrics_analysis.go`** — `buildCILTreeFromPostfix` (line 392):
   - Change signature from `postfix []interface{}` to `postfix []parsing.PostfixElement`
   - Replace the `switch v := elem.(type)` block with:
     ```go
     switch elem := elem.(type) {
     case parsing.PostfixLiteral:
         stack = append(stack, &cil.CILLiteral{Value: elem.Value})
     case parsing.PostfixIdentifier:
         kind, err := resolveInputRefKind(elem.Name, metricName, analyzer)
         if err != nil { return nil, err }
         stack = append(stack, &cil.CILInputRef{Kind: kind, Name: elem.Name})
     case parsing.PostfixOperator:
         opStr := operatorSymbol(elem.DefType)
         switch elem.DefType {
         // same operator dispatch as before, keyed on elem.DefType
         // (copy the existing case bodies from the old symboltable.DefinitionType switch)
         }
     default:
         return nil, fmt.Errorf("unexpected postfix element type %T", elem)
     }
     ```
   - Update `buildCILTree` (line 373) to convert `[]PostfixElement` to pass through (no conversion needed if `CalculatePostfix` is already `[]PostfixElement`)

   - Update `symbolDefToPostfixElem` (line 981):
     - Change return type from `interface{}` to `parsing.PostfixElement`
     - LiteralDefinition case: convert `def.Data.(int)` → `parsing.PostfixLiteral{Value: float64(v)}`
     - Identifier case: convert `string` → `parsing.PostfixIdentifier{Name: name}`
     - Default case: convert `DefinitionType` → `parsing.PostfixOperator{DefType: def.DefinitionType}`

   - Update `buildCILTreeFromSymbolDefs` (line 1026):
     - Change `[]interface{}` to `[]parsing.PostfixElement` for the local postfix variable

f. Update **`src/main/go/parser/cpl_metrics_listener.go`** — `collectCalculatePostfix` (line 310):
   - Change return type from `[]interface{}` to `[]parsing.PostfixElement`
   - Replace `postfix = append(postfix, literalValue)` with `postfix = append(postfix, parsing.PostfixLiteral{Value: float64(literalValue)})`
   - Replace all `postfix = append(postfix, symboltable.XxxDefinition)` with `postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.XxxDefinition})`
   - Replace `postfix = append(postfix, identifierName)` with `postfix = append(postfix, parsing.PostfixIdentifier{Name: identifierName})`

g. Run `go build ./src/main/go/...` — must compile.

h. Run `go test ./src/main/go/parsing/... ./analysis/... ./parser/...` — must pass.

Note: No changes to `src/main/go/parser/cpl_conditional_listener.go` are needed. The conditional path (`ClauseData.ConditionPostfix`) is already `[]SymbolDefinition` and flows through `symbolDefToPostfixElem`, which is updated above.

**Acceptance Criteria**:
- `src/main/go/parsing/postfix_element.go` exists with `PostfixElement` interface and three concrete types
- `ParserMetricDefinition.CalculatePostfix` is `[]PostfixElement`
- `buildCILTreeFromPostfix` operates on `[]PostfixElement`, not `[]interface{}`
- `buildCILTreeFromSymbolDefs` produces `[]PostfixElement`
- `go build ./src/main/go/...` compiles without errors
- `go test ./src/main/go/parsing/... ./analysis/... ./parser/...` passes with zero failures

---

### Task 5: Migrate SymbolDefinition.Data to Typed Carriers — Completed

**Goal**: Replace `interface{}` in `SymbolDefinition.Data` with a `symboltable.DefinitionData` sum-type interface and named concrete carriers. This provides compile-time safety at every site that reads `def.Data`.

**Scope boundary**: Only `SymbolDefinition.Data` is being typed. `SymbolTable.Data` (scope-level) remains `interface{}`. Changing `SymbolTable.Data` requires implementing `IsDefinitionData()` on additional types (`*parsing.ParserMetricDefinition`, `*symboltable.ConnectorDeclaration`, `*ast.ConceptDefinitionData`) and is deferred to a future increment.

**Test-First Development**:

a. Create `src/main/go/src/main/go/symboltable/definition_data.go`:

```go
package symboltable

// DefinitionData is the interface for type-specific data attached to SymbolDefinition.Data.
type DefinitionData interface {
    IsDefinitionData()
}

// StringData holds a single string value (identifier text, constraint values).
type StringData struct{ Value string }
func (StringData) IsDefinitionData() {}

// IntData holds an integer literal value.
type IntData struct{ Value int }
func (IntData) IsDefinitionData() {}

// FloatData holds a float literal value.
type FloatData struct{ Value float64 }
func (FloatData) IsDefinitionData() {}

// BoolData holds a boolean literal value (true/false).
type BoolData struct{ Value bool }
func (BoolData) IsDefinitionData() {}
```

b. Write tests in `src/main/go/symboltable/definition_data_test.go`:
   - Interface satisfaction for each concrete type
   - Construction and field access for each type
   - Verify `nil` satisfies `DefinitionData` (nil interface value)

c. Run `go test ./src/main/go/symboltable/...` — must pass.

d. Update **`src/main/go/symboltable/symbol_table.go`**:
   - `SymbolDefinition.Data` (line 395): `Data interface{}` → `Data DefinitionData`
   - `DefineWithData` (line 490): `data interface{}` → `data DefinitionData`
   - Add `func (ClauseData) IsDefinitionData() {}` after the `ClauseData` struct (line 406)

e. Add `IsDefinitionData()` to **`src/main/go/parsing/parser_input_definition.go`**:
   ```go
   // IsDefinitionData satisfies the symboltable.DefinitionData interface.
   func (ParserInputDefinition) IsDefinitionData() {}
   ```
   The `parsing` package already imports `symboltable`, so no new import is needed.

f. Update **`src/main/go/analysis/metrics_analysis.go`** — `symbolDefToPostfixElem` (line 981):
   - The `def.Data` field is now `symboltable.DefinitionData` instead of `interface{}`
   - Change the Identifier case: previously `def.Data.(string)`; now use `def.Data.(symboltable.StringData).Value`:
     ```go
     case symboltable.Identifier:
         if def.Name != nil {
             return parsing.PostfixIdentifier{Name: *def.Name}, nil
         }
         if sd, ok := def.Data.(symboltable.StringData); ok {
             return parsing.PostfixIdentifier{Name: sd.Value}, nil
         }
         return parsing.PostfixIdentifier{}, fmt.Errorf("identifier definition in metric %q has neither name nor string data", metricName)
     ```
   - Change the LiteralDefinition case: previously `def.Data` was `interface{}` (could be `int`, `string`, etc.); now use typed carriers:
     ```go
     case symboltable.LiteralDefinition:
         if def.Data == nil {
             return parsing.PostfixLiteral{}, fmt.Errorf("literal definition in metric %q has nil data", metricName)
         }
         switch d := def.Data.(type) {
         case symboltable.IntData:
             return parsing.PostfixLiteral{Value: float64(d.Value)}, nil
         case symboltable.FloatData:
             return parsing.PostfixLiteral{Value: d.Value}, nil
         case symboltable.StringData:
             return parsing.PostfixIdentifier{Name: d.Value}, nil
         default:
             return parsing.PostfixLiteral{}, fmt.Errorf("literal definition in metric %q has unexpected data type %T", metricName, def.Data)
         }
     ```

g. Update **`src/main/go/analysis/metrics_analysis.go`** — `extractClauseData` (line 1274):
   - `def.Data` is now `symboltable.DefinitionData`; type-assert to `*symboltable.ClauseData` or `symboltable.ClauseData`:
     ```go
     clauseData, ok := def.Data.(*symboltable.ClauseData)
     if !ok {
         cd, ok2 := def.Data.(symboltable.ClauseData)
         if !ok2 {
             return nil, fmt.Errorf("clause definition in metric %q has unexpected Data type %T", metricName, def.Data)
         }
         clauseData = &cd
     }
     ```

h. Update **`src/main/go/analysis/types_analysis.go`** (around line 108):
   - Find all `def.Data.(string)` assertions and replace with `def.Data.(symboltable.StringData).Value`
   - Handle the case where the assertion fails (return an appropriate error)

i. Update **`src/main/go/analysis/cpl_types.go`** (around line 1065):
   - Same `def.Data.(string)` → `def.Data.(symboltable.StringData).Value` conversion

j. Update **`src/main/go/parser/cpl_metrics_listener.go`** — `ExitLiteral` method:
   - Boolean literals (`True`/`False`): change `Data: true` / `Data: false` to `Data: symboltable.BoolData{Value: true}` / `Data: symboltable.BoolData{Value: false}`
   - Nil literal: `Data: nil` — no change needed (nil satisfies `DefinitionData`)
   - Integer literal: change `Data: intValue` to `Data: symboltable.IntData{Value: intValue}`
   - String literal: change `Data: literalText` to `Data: symboltable.StringData{Value: literalText}`

k. Update **`src/main/go/parser/parser_support.go`** — `ExitPrimaryExpr` (line 306):
   - Change `identifierText` (string) data to `symboltable.StringData{Value: identifierText}`

l. Update any tests that construct `SymbolDefinition` with bare `Data` values:
   - Replace `Data: "some_string"` with `Data: symboltable.StringData{Value: "some_string"}`
   - Replace `Data: 123` with `Data: symboltable.IntData{Value: 123}`
   - Replace `Data: true` with `Data: symboltable.BoolData{Value: true}`
   - (Nil data requires no change)

m. Run `go build ./src/main/go/...` — must compile.

n. Run `go test ./src/main/go/parser/... ./analysis/... ./symboltable/...` — must pass with zero failures.

**Acceptance Criteria**:
- `src/main/go/symboltable/definition_data.go` exists with `DefinitionData` interface and `StringData`, `IntData`, `FloatData`, `BoolData` carriers
- `SymbolDefinition.Data` is `DefinitionData`, not `interface{}`
- `DefineWithData` accepts `DefinitionData`, not `interface{}`
- `ClauseData` implements `DefinitionData`
- `parsing.ParserInputDefinition` implements `DefinitionData`
- All production code reads `def.Data` through typed access (no `interface{}` type assertions on `SymbolDefinition.Data`)
- `SymbolTable.Data` remains `interface{}` (scope-level — not changed in this story)
- `go build ./src/main/go/...` compiles without errors
- `go test ./src/main/go/parser/... ./analysis/... ./symboltable/...` passes with zero failures

---

## Test-First Development

This story follows the Logical TDD Lifecycle. Each coding task integrates the test-and-implement cycle:

- **Task 1**: Write interface satisfaction test → implement struct → run tests
- **Task 2**: Update test expectations to use `cil.CILNode`/`cil.CILConnectorNode` → update production code → run tests
- **Task 3**: Verify zero references (precondition check) → delete file → run tests
- **Task 4**: Write `PostfixElement` tests → create types → update production code → run tests
- **Task 5**: Write `DefinitionData` tests → create types → update production code and call sites → run tests

Tests are not broken out as separate tasks. Each task includes all test updates needed for that task's changes.

## Constraints

1. Compile-time type safety must be improved — no new `interface{}` fields introduced.
2. All existing tests must pass after each task.
3. Changes span `analysis`, `cil`, `parsing`, `parser`, `symboltable`, and `synthesis` packages.
4. `SymbolTable.Data` (scope-level) remains `interface{}` until a future increment. Only `SymbolDefinition.Data` is typed in this story.
5. The `synthesis/synthesize.go` file does not reference old CIL types or `CILTrees` — it reads `CILRoot` from individual metric definitions. No synthesis changes are needed.

## Intent

Eliminate the half-migrated CIL state and the `interface{}` anti-pattern so that later refactoring phases (split-bloat, reduce-duplication, address-fragility) have a stable, type-safe foundation.

The goal is NOT to change behavior — it is to make the existing behavior explicit and type-safe at compile time. Every `switch` on `interface{}`, every `.(type)` assertion, and every `map[string]CILAbstractNode` should be replaced with a typed construct.

## Acceptance Criteria

1. Only one CIL representation exists: `cil.CILNode`. `analysis.CILAbstractNode` and its enum-based node types are deleted.
2. `AnalysisDefinitions.CILTrees` uses `map[string]cil.CILNode`.
3. `ParserMetricDefinition.CalculatePostfix` uses `[]PostfixElement` typed elements.
4. `SymbolDefinition.Data` uses `symboltable.DefinitionData` typed carriers instead of `interface{}`.
5. `go test ./src/main/go/...` passes with zero failures.

## Notes

- The `CILTrees` map in `AnalysisDefinitions` is populated by connectors but never read by production synthesis code — it is consumed only in tests. This makes the migration in Task 2 low-risk.
- `SymbolTable.Data` at the scope level is not being changed in this story. The plan identifies files that would need updating for scope-level data typing (`src/main/go/analysis/concept_analysis.go`, `src/main/go/analysis/connectors_analysis.go` line 180, `parser/cpl_concepts_listener.go`), but these are deferred.
- The `src/main/go/parser/cpl_conditional_listener.go` file requires no changes in this story. Clause condition postfix data is already `[]SymbolDefinition` and flows through `symbolDefToPostfixElem`, which is updated in Tasks 4 and 5.
- `BoolData` is added to the typed data carriers beyond what the plan listed, because `ExitLiteral` in `src/main/go/parser/cpl_metrics_listener.go` passes `true`/`false` boolean values as `Data` for `TrueLiteralDefinition`/`FalseLiteralDefinition`.
- Task 5 requires updating all `DefineWithData` call sites that pass bare `string`, `int`, or `bool` values — these are listed explicitly in subtasks j and k.

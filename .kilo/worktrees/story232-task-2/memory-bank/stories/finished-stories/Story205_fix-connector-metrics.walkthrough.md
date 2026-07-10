# Task 1 Walkthrough: Fix resolveInputRefKind to check analyzer.connectors

## Purpose

This walkthrough documents the fix to [`resolveInputRefKind()`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:255) in the CPL metrics analyzer (Story205, Task 1). The function classifies identifier references in metric `calculate` clauses into their appropriate input kinds (`"connector"` or `"metric"`) for CIL tree construction. Prior to this fix, connector names were misidentified as metric references because the function only checked `analyzer.measures` and `analyzer.metrics`, omitting `analyzer.connectors`.

## Data Flow

### 1. Metric calculate clause parsing

The CPL parser produces a postfix expression sequence (`CalculatePostfix`) for each metric's `calculate` clause. String identifiers in this sequence represent references to other metrics, measures, or connectors.

### 2. buildCILTree calls resolveInputRefKind

During CIL tree construction, [`buildCILTree()`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:268) processes each postfix element. When it encounters a string identifier, it calls [`resolveInputRefKind()`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:255) to determine whether the identifier represents a `"connector"` (measure or connector reference) or a `"metric"` (metric reference).

### 3. resolveInputRefKind classification

The function checks the analyzer's lookup maps in priority order:

| Map checked           | Kind returned | Description                                 |
| --------------------- | ------------- | ------------------------------------------- |
| `analyzer.measures`   | `"connector"` | Direct measure reference                    |
| `analyzer.connectors` | `"connector"` | Connector reference (**added in this fix**) |
| `analyzer.metrics`    | `"metric"`    | Metric-to-metric reference                  |
| _(none found)_        | `"metric"`    | Default fallback for forward references     |

## The Fix

### Before

```go
func resolveInputRefKind(name string, metricName string, analyzer *CPLAnalyzer) string {
    if _, exists := analyzer.measures[name]; exists {
        return "connector"
    }
    if _, exists := analyzer.metrics[name]; exists {
        return "metric"
    }
    return "metric" // default for forward references
}
```

### After

```go
func resolveInputRefKind(name string, metricName string, analyzer *CPLAnalyzer) string {
    if _, exists := analyzer.measures[name]; exists {
        return "connector"
    }
    if _, exists := analyzer.connectors[name]; exists {
        return "connector"
    }
    if _, exists := analyzer.metrics[name]; exists {
        return "metric"
    }
    return "metric" // default for forward references
}
```

The change is a 3-line insertion (lines 259-261 in [`metrics_analysis.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:259)):

```go
if _, exists := analyzer.connectors[name]; exists {
    return "connector"
}
```

This check is placed **after** the `measures` check and **before** the `metrics` check, ensuring that connector names are correctly classified as `"connector"` rather than falling through to the `"metric"` default.

## Test Coverage

4 new test functions in [`metrics_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:2477) (Story205 Task 1 Phase B):

| Test                                                                                                                                 | Coverage                                                        |
| ------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------- |
| [`TestResolveInputRefKind_MeasureLookup`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:2479)       | Verify that names in `analyzer.measures` return `"connector"`   |
| [`TestResolveInputRefKind_ConnectorLookup`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:2518)     | Verify that names in `analyzer.connectors` return `"connector"` |
| [`TestResolveInputRefKind_MetricLookup`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:2553)        | Verify that names in `analyzer.metrics` return `"metric"`       |
| [`TestResolveInputRefKind_UnknownNameFallback`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:2589) | Verify that unknown names fall back to `"metric"`               |

All tests exercise `resolveInputRefKind` indirectly through `AnalyzeMetricScope` / `buildCILTree` since the function is unexported.

## Verification

- All 4 new tests pass (`go test ./analysis/... -run TestResolveInputRefKind`)
- Full `go test ./analysis/...` passes with no regressions
- Existing connector CIL tree tests continue to pass, confirming the fix integrates correctly with the broader metrics analysis pipeline

## Post-Mortem

### What went well

- **Surgical fix**: The root cause was immediately apparent upon inspection — `resolveInputRefKind` checked two of the three relevant analyzer maps but not `connectors`. The fix required only 3 lines of code.
- **Clear root cause**: The bug manifested as connector names being misclassified as `"metric"` in CILInputRef nodes, which was traced directly to the missing map lookup.
- **Good test coverage**: The 4 tests cover all classification paths (measures, connectors, metrics, unknown) and prevent regression.

### Lessons learned

- **Check all relevant maps in lookup functions**: When a classifier function inspects multiple analyzer maps, each map that could contain a valid reference must be explicitly checked. The `connectors` map was added after `resolveInputRefKind` was originally written, and the original function was not updated to account for the new map.
- **Map addition checklist**: When adding a new entity map to the analyzer, audit all existing lookup functions that classify or resolve names to ensure they check the new map.

### Files changed

| File                                                                                                         | Change                                                                 |
| ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------- |
| [`metrics_analysis.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go)           | Added `analyzer.connectors` check to `resolveInputRefKind()` (3 lines) |
| [`metrics_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go) | Added 4 test functions for resolveInputRefKind classification          |

---

# Task 2 Walkthrough: Pre-populate connector MetricComputations in generated compute functions

## Purpose

This walkthrough documents the fix to [`generatePostfixComputeBody()`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:695) in the CIL transformer (Story205, Task 2). When connector inputs are loaded from Parquet files into the `inputs` map, they are raw `*calculation.ChronoconeTensor[any]` values — not registered `MetricComputation` objects. During postfix evaluation, [`EvaluatePostfixFloat64`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:716) calls `controller.GetMetricComputation(name)` for each identifier in the postfix table. If the identifier is a connector name, `GetMetricComputation` returns `nil` because the connector tensor was never registered as a `MetricComputation`, causing evaluation to fail. This fix pre-registers every entry in the `inputs` map as a `MetricComputation` before `EvaluatePostfixFloat64` runs.

## Data Flow

### 1. Connector loading (outside generated code)

At runtime, connector Parquet files are read and their tensors are placed into the `inputs` map keyed by connector name:

```go
inputs["revenue_amt"] = loadedTensor
inputs["cost_amt"] = loadedTensor
```

These are raw tensors, not wrapped in `MetricComputation` objects.

### 2. Metric compute function generation

[`TransformMetric()`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:742) generates a `computeX(controller, inputs)` function. For arithmetic expressions (binary/unary ops), it calls [`generatePostfixComputeBody()`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:695) to emit the function body.

### 3. generatePostfixComputeBody now emits connector registration

Before calling `EvaluatePostfixFloat64`, the generated body includes a registration block (from [`generateConnectorRegistrationCode()`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:657)):

```go
for name, tensor := range inputs {
    mc := calculation.NewMetricComputation(0)
    if err := controller.RegisterMetricComputation(name, mc); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to register metric computation for %s: %v\n", name, err)
        os.Exit(1)
    }
    if err := mc.SetTensor(tensor); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to set tensor for %s: %v\n", name, err)
        os.Exit(1)
    }
}
```

This loop iterates over every entry in `inputs`, creates a `MetricComputation`, registers it on the controller under the same name, and sets the tensor. After this block, `controller.GetMetricComputation("revenue_amt")` returns a valid `MetricComputation` with the loaded tensor.

### 4. Postfix evaluation succeeds

With all connector inputs now registered as `MetricComputation` objects, [`EvaluatePostfixFloat64`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:716) can resolve connector names in the postfix table and retrieve their tensors for arithmetic operations.

## The Fix

### New method: generateConnectorRegistrationCode

Added at line 657 in [`cil_transformer.go`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:657):

```go
func (t *CILAbstractTransformer) generateConnectorRegistrationCode() string {
    return "\tfor name, tensor := range inputs {\n" +
        "\t\tmc := calculation.NewMetricComputation(0)\n" +
        "\t\tif err := controller.RegisterMetricComputation(name, mc); err != nil {\n" +
        "\t\t\tfmt.Fprintf(os.Stderr, \"Failed to register metric computation for %s: %v\\n\", name, err)\n" +
        "\t\t\tos.Exit(1)\n" +
        "\t\t}\n" +
        "\t\tif err := mc.SetTensor(tensor); err != nil {\n" +
        "\t\t\tfmt.Fprintf(os.Stderr, \"Failed to set tensor for %s: %v\\n\", name, err)\n" +
        "\t\t\tos.Exit(1)\n" +
        "\t\t}\n" +
        "\t}\n"
}
```

### Modified method: generatePostfixComputeBody

At line 715 in [`cil_transformer.go`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:715), the connector registration block is emitted immediately after the postfix table literal and before `EvaluatePostfixFloat64`:

```go
opsLiteral := t.GeneratePostfixTableLiteral(ops)

var sb strings.Builder
sb.WriteString(fmt.Sprintf("\tops := %s\n", opsLiteral))
sb.WriteString(t.generateConnectorRegistrationCode())          // ← added
sb.WriteString("\tresult, err := calculation.EvaluatePostfixFloat64(controller, ops)\n")
```

This ordering is critical: registration must happen **before** evaluation so that `GetMetricComputation` calls inside the postfix evaluator resolve successfully.

## Test Coverage

2 new test functions in [`cil_transformer_test.go`](src/chronocone_planning_language/src/main/go/cil/cil_transformer_test.go:3041) (Story205 Task 2 Phase B):

| Test                                                                                                                                            | Coverage                                                                                                                                                                          |
| ----------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [`TestGenerateConnectorRegistrationCode`](src/chronocone_planning_language/src/main/go/cil/cil_transformer_test.go:3044)                        | Verifies `generateConnectorRegistrationCode` returns non-empty string containing the `for` loop, `RegisterMetricComputation`, `NewMetricComputation(0)`, and `SetTensor` patterns |
| [`TestGeneratePostfixComputeBody_ContainsConnectorRegistration`](src/chronocone_planning_language/src/main/go/cil/cil_transformer_test.go:3069) | Verifies `generatePostfixComputeBody` output contains the connector registration block and that it appears **before** `EvaluatePostfixFloat64` in the generated code              |

## Verification

- Both new tests pass (`go test ./cil/... -run TestGenerateConnectorRegistrationCode` and `-run TestGeneratePostfixComputeBody_ContainsConnectorRegistration`)
- Full `go test ./cil/...` passes with no regressions
- Existing postfix table tests (`TestTransformMetric_ArithmeticBinaryOp`, `TestTransformMetric_PostfixTableOutput`, etc.) continue to pass, confirming the fix integrates correctly with the broader CIL transformer pipeline

## Post-Mortem

### What went well

- **Surgical fix**: The root cause was clear — connector tensors in `inputs` were not registered as `MetricComputation` objects, so `GetMetricComputation` returned `nil` during postfix evaluation. The fix required only a single new method and one additional call site.
- **Clear data flow**: The fix follows a straightforward pattern: register everything in `inputs` before evaluating. This is robust because it handles any connector that ends up in the `inputs` map, regardless of how it was loaded.
- **Good test coverage**: The two tests cover both the unit behavior (registration code generation) and the integration behavior (ordering within the compute body), preventing regression on both dimensions.

### Lessons learned

- **When adding new input types, audit all code generation paths**: Connector inputs were handled correctly in the analyzer (Task 1) and in the visitor (`VisitCILInputRef` emits `inputs["name"]` for connectors), but the code generation path that produces runtime registration code was missed. Each new input type needs a checklist: analyzer classification, visitor emission, runtime registration, and end-to-end loading.
- **Registration-before-evaluation is a lifecycle invariant**: Any generated compute function that relies on `GetMetricComputation` during evaluation must ensure all referenced inputs are pre-registered. This is now explicitly documented in the `generatePostfixComputeBody` contract.

### Files changed

| File                                                                                                  | Change                                                                                                                                                            |
| ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [`cil_transformer.go`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go)           | Added `generateConnectorRegistrationCode()` method (lines 657–669); modified `generatePostfixComputeBody()` to call it before `EvaluatePostfixFloat64` (line 715) |
| [`cil_transformer_test.go`](src/chronocone_planning_language/src/main/go/cil/cil_transformer_test.go) | Added `TestGenerateConnectorRegistrationCode` and `TestGeneratePostfixComputeBody_ContainsConnectorRegistration` test functions (lines 3041–3117)                 |

---

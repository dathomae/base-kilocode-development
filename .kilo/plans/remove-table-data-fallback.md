# Plan: Remove Legacy `table.Data` Fallback from `AnalyzeMetricScope`

## Objective

Remove the backward-compatibility `table.Data` cast fallback in `AnalyzeMetricScope` (lines 141–151 of `src/main/go/analysis/metrics_analysis.go`). All metric scope definitions are now registered via `DefineWithData` with `MetricScopeDefinition` type (Story273), making the legacy `table.Data` fallback dead code. Removing it ensures a single canonical retrieval path through `Definitions` and eliminates maintenance overhead.

## Context

- Story273 added `DefineWithData` calls for metrics and aggregators
- `AnalyzeMetricScope` was updated to iterate `Definitions` first, with a `table.Data` fallback kept for backward compatibility
- All parser-produced metric scopes now push `MetricScopeDefinition` entries via `DefineWithData`
- Future story (Story256, aggregators-increment2-analyze) will follow the same `Definitions`-only pattern
- Removing the fallback forces any remaining `Data`-only test constructions to be updated, ensuring full coverage of the canonical path

## Files to Modify

1. `src/main/go/analysis/metrics_analysis.go` — remove the `table.Data` fallback block
2. `src/main/go/analysis/metrics_analysis_test.go` — update `TestAnalyzeMetricsScope_DuplicateMetricNames_ReturnsError` to call `DefineWithData`

## Steps

### Step 1: Remove the `table.Data` fallback from `AnalyzeMetricScope`

In `metrics_analysis.go`, replace lines 130–151 (the definition retrieval block):

```go
	// Retrieve structured metric definition from Definitions (primary path)
	var metricDef *parsing.ParserMetricDefinition
	for _, def := range table.Definitions {
		if def.DefinitionType == symboltable.MetricScopeDefinition {
			if md, ok := def.Data.(*parsing.ParserMetricDefinition); ok {
				metricDef = md
				break
			}
		}
	}

	// Fallback to table.Data for backward compatibility
	if metricDef == nil {
		if table.Data == nil {
			return fmt.Errorf("metric scope has no structured definition data")
		}
		var ok bool
		metricDef, ok = table.Data.(*parsing.ParserMetricDefinition)
		if !ok {
			return fmt.Errorf("metric scope data has unexpected type %T", table.Data)
		}
	}
```

with:

```go
	// Retrieve structured metric definition from Definitions
	var metricDef *parsing.ParserMetricDefinition
	for _, def := range table.Definitions {
		if def.DefinitionType == symboltable.MetricScopeDefinition {
			if md, ok := def.Data.(*parsing.ParserMetricDefinition); ok {
				metricDef = md
				break
			}
		}
	}
	if metricDef == nil {
		return fmt.Errorf("metric scope has no MetricScopeDefinition in its Definitions")
	}
```

Update the function comment on lines 52–128 to remove fallback-related error cases:
- Remove: `//   - "metric scope has no structured definition data" - table.Data is nil`
- Remove: `//   - "metric scope data has unexpected type %T" - table.Data is not a ParserMetricDefinition`
- Add: `//   - "metric scope has no MetricScopeDefinition in its Definitions" - No MetricScopeDefinition entry found in Definitions`

### Step 2: Update `TestAnalyzeMetricsScope_DuplicateMetricNames_ReturnsError`

This test currently sets `metricScope.Data` directly without calling `DefineWithData`. Add `DefineWithData` calls to both metric scopes:

```go
	firstMetricScope.Data = &parsing.ParserMetricDefinition{...}
	nameCopy1 := "duplicate_metric"
	_ = firstMetricScope.DefineWithData(&nameCopy1, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, ...)

	secondMetricScope.Data = &parsing.ParserMetricDefinition{...}
	nameCopy2 := "duplicate_metric"
	_ = secondMetricScope.DefineWithData(&nameCopy2, symboltable.MetricScopeDefinition, symboltable.SourceLocation{}, ...)
```

Note: `DefineWithData` won't detect the duplicate because the two `MetricScope` entries are siblings (different scopes). The duplicate is correctly caught by `analyzer.metricAnalyzer.HasMetric(metricName)` at line 161.

### Step 3: Run full regression verification

```bash
go test ./src/main/go/... -count=1 -timeout 120s
go vet ./src/main/go/analysis/...
go build ./src/main/go/...
```

## Acceptance Criteria

- `AnalyzeMetricScope` retrieves metric definitions exclusively from `table.Definitions`
- `table.Data` is no longer accessed in `AnalyzeMetricScope`
- Error message for missing definition references "MetricScopeDefinition in its Definitions"
- All tests pass with no regressions
- `go vet` shows no new warnings
- `go build` succeeds

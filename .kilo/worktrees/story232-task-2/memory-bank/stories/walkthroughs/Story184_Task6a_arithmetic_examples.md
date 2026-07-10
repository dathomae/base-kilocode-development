# Story184 Task 6a — Arithmetic Example CPL Files Walkthrough

## 1. Overview

Task 6a created six example CPL files that demonstrate arithmetic expression parsing in metric [`calculate`](src/chronocone_planning_language/src/main/antlr/CPLMetrics.g4:45) clauses. Each file isolates a specific arithmetic pattern — from simple binary addition to deeply nested expressions with parentheses — providing concrete, human-readable sources that exercise the Increment 2 expression grammar introduced in Story184.

**Purpose:**

- Provide standalone CPL examples that can be parsed and inspected independently of the test suite.
- Serve as reference material for Task 6c (integration verification) by supplying real files that the parser and analyzer must handle correctly.
- Support Task 6b (IO spec examples) by giving declared metrics that can be referenced in output specifications.
- Document the expected CIL tree shape for each arithmetic pattern, bridging CPL source syntax to the internal [`CILMetricNode`](src/chronocone_planning_language/src/main/go/cil/cil_metric_node.go:7) representation.

**Input:** A `.cpl` file containing a `metrics` section with one metric and an arithmetic expression in its `calculate` clause.

**Output:** After parsing and analysis, a `CILMetricNode` whose `CILNode` root is either a [`CILLiteral`](src/chronocone_planning_language/src/main/go/cil/cil_literal.go:7), [`CILInputRef`](src/chronocone_planning_language/src/main/go/cil/cil_metric_node.go:13), [`CILBinaryOp`](src/chronocone_planning_language/src/main/go/cil/cil_binary_op.go:8), or [`CILUnaryOp`](src/chronocone_planning_language/src/main/go/cil/cil_unary_op.go:8).

**Design Reference:** [ERO005 Metrics Parsing Phase — Increment 2](memory-bank/specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md) and [ERO005 Metrics Analysis Phase — Increment 2](memory-bank/specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md).

---

## 2. File Inventory

All files live in [`src/chronocone_planning_language/examples/metrics_arithmetic/`](src/chronocone_planning_language/examples/metrics_arithmetic/).

| #   | File                                                                                                                          | Metric Name         | Calculate Expression | Expected CIL Tree Structure                                                                                                                                                                                              |
| --- | ----------------------------------------------------------------------------------------------------------------------------- | ------------------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1   | [`simple_addition.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/simple_addition.cpl)                     | `sum_metric`        | `input_a + input_b`  | `CILBinaryOp{Op:"+", Left: CILInputRef{input_a}, Right: CILInputRef{input_b}}`                                                                                                                                           |
| 2   | [`multiplication_precedence.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/multiplication_precedence.cpl) | `precedence_metric` | `a * b + c`          | `CILBinaryOp{Op:"+", Left: CILBinaryOp{Op:"*", Left: CILInputRef{a}, Right: CILInputRef{b}}, Right: CILInputRef{c}}`                                                                                                     |
| 3   | [`subtraction_division.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/subtraction_division.cpl)           | `sub_div_metric`    | `a - b / c`          | `CILBinaryOp{Op:"-", Left: CILInputRef{a}, Right: CILBinaryOp{Op:"/", Left: CILInputRef{b}, Right: CILInputRef{c}}}`                                                                                                     |
| 4   | [`unary_negation.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/unary_negation.cpl)                       | `negate_metric`     | `-a + b`             | `CILBinaryOp{Op:"+", Left: CILUnaryOp{Op:"negate", Operand: CILInputRef{a}}, Right: CILInputRef{b}}`                                                                                                                     |
| 5   | [`parenthesized.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/parenthesized.cpl)                         | `paren_metric`      | `(a + b) * (c - d)`  | `CILBinaryOp{Op:"*", Left: CILBinaryOp{Op:"+", Left: CILInputRef{a}, Right: CILInputRef{b}}, Right: CILBinaryOp{Op:"-", Left: CILInputRef{c}, Right: CILInputRef{d}}}`                                                   |
| 6   | [`complex_expression.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/complex_expression.cpl)               | `complex_metric`    | `a * b + c / d - e`  | `CILBinaryOp{Op:"-", Left: CILBinaryOp{Op:"+", Left: CILBinaryOp{Op:"*", Left: CILInputRef{a}, Right: CILInputRef{b}}, Right: CILBinaryOp{Op:"/", Left: CILInputRef{c}, Right: CILInputRef{d}}}, Right: CILInputRef{e}}` |

### ASCII Tree Visualizations

#### File 1: `input_a + input_b`

```
     (+)
    /   \
  (a)   (b)
```

#### File 2: `(a * b) + c`

```
       (+)
      /   \
    (*)   (c)
   /   \
 (a)   (b)
```

#### File 3: `a - (b / c)`

```
       (-)
      /   \
    (a)   (/)
         /   \
       (b)   (c)
```

#### File 4: `(-a) + b`

```
       (+)
      /   \
   (neg)  (b)
    |
   (a)
```

#### File 5: `(a + b) * (c - d)`

```
            (*)
          /     \
        (+)     (-)
       /   \   /   \
     (a)  (b) (c)  (d)
```

#### File 6: `((a * b) + (c / d)) - e`

```
                  (-)
               /       \
             (+)       (e)
            /   \
          (*)   (/)
         / \   / \
       (a)(b)(c)(d)
```

---

## 3. Design Decisions

### Metrics-Only CPL (No concepts/types/units/measures/connectors)

All six example files contain **only** a `metrics` section. There are no `concepts`, `types`, `units`, `measures`, or `connectors` sections. This is an intentional design choice with three motivations:

1. **Follows Existing Integration Test Pattern** — The parse→analyze integration tests in [`metrics_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:836) (starting at line 836) use metrics-only CPL strings to verify the full pipeline. For example, [`TestParseAndAnalyze_SimpleAddition`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:838) parses:

   ```cpl
   metrics
       arith_metric
           inputs
               metric input_a
               metric input_b
           calculate
               input_a + input_b
   ```

   The example files replicate this exact structural pattern, keeping the examples aligned with the canonical integration tests.

2. **Keeps Examples Focused on Arithmetic Expressions** — The goal of Task 6a is to demonstrate arithmetic expression parsing, not to exercise concepts, types, units, or connector parsing. A minimal `metrics` section with `inputs` and `calculate` isolates the variable under test: the `calculate` expression.

3. **Avoids Pre-Populated State Requirements** — Full analysis of a CPL document with concepts, types, units, measures, and connectors requires those entities to be declared and validated. Metrics-only CPL files can be parsed directly without needing a pre-populated analyzer state, making them useful for quick parser verification and standalone documentation.

**Trade-off:** The `example_definitions.go` file (see Key Files Reference) documents the _complete_ structural definitions that would accompany each example in a full CPL program — including concepts, types, units, measures, and connectors. This allows readers to understand how the metrics-only snippet would fit into a larger document, while the `.cpl` files themselves remain minimal.

---

## 4. How to Use

### Parse an Example File

From the repository root:

```go
package main

import (
    "fmt"
    "os"

    "github.com/chronocone/chronocone_planning_language/src/main/go/parser"
)

func main() {
    data, err := os.ReadFile("src/chronocone_planning_language/examples/metrics_arithmetic/simple_addition.cpl")
    if err != nil {
        panic(err)
    }

    doc, err := parser.Parse(string(data))
    if err != nil {
        panic(err)
    }

    fmt.Printf("Parsed %d top-level scopes\n", len(doc.Children))
}
```

### Run the Companion Test

The companion test [`metrics_arithmetic_test.go`](src/chronocone_planning_language/examples/metrics_arithmetic/metrics_arithmetic_test.go) verifies that all six example files parse without errors:

```bash
cd src/chronocone_planning_language/examples/metrics_arithmetic
go test -v ./...
```

Expected output:

```
=== RUN   TestExampleFilesParse
=== RUN   TestExampleFilesParse/simple_addition.cpl
=== RUN   TestExampleFilesParse/multiplication_precedence.cpl
=== RUN   TestExampleFilesParse/subtraction_division.cpl
=== RUN   TestExampleFilesParse/unary_negation.cpl
=== RUN   TestExampleFilesParse/parenthesized.cpl
=== RUN   TestExampleFilesParse/complex_expression.cpl
--- PASS: TestExampleFilesParse (0.00s)
```

### Full Parse→Analyze Pipeline

To verify that an example file produces the correct CIL tree, run the same pattern used by the integration tests in [`metrics_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:836):

```go
import (
    "github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
    "github.com/chronocone/chronocone_planning_language/src/main/go/parser"
    "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

rootTable, err := parser.Parse(string(data))
// ... create analyzer with required measures pre-populated ...
entryHandlers := map[symboltable.ScopeType]analysis.AnalysisFunc{
    symboltable.MetricsScope: analysis.AnalyzeMetricsScope,
    symboltable.MetricScope:  analysis.AnalyzeMetricScope,
}
err = analysis.WalkSymbolTableWithHandlers(rootTable, entryHandlers, nil, analyzer)
// ... assert on metric.CILRoot structure ...
```

---

## 5. Relationship to Other Tasks

### Task 6c — Integration Verification

These example files are the **primary input** for Task 6c. The integration tests in [`metrics_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:836) verify the parse→analyze pipeline using inline CPL strings. The example files provide the same content as persistent, version-controlled artifacts that can be:

- Inspected by humans without reading Go test code.
- Used in external tooling (e.g., CLI parse commands, IDE plugins).
- Referenced in documentation and walkthroughs.

The six expressions cover every operator and structural pattern needed for Task 6c:

| Pattern              | Example File                    | Integration Test Equivalent                                                                                                                                                                                                                                         |
| -------------------- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Binary `+`           | `simple_addition.cpl`           | [`TestParseAndAnalyze_SimpleAddition`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:838)                                                                                                                                          |
| Precedence `*` / `+` | `multiplication_precedence.cpl` | [`TestParseAndAnalyze_MultiplicationWithPrecedence`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:904)                                                                                                                            |
| Binary `-` and `/`   | `subtraction_division.cpl`      | Covered by [`TestBuildCILTree_BinarySubtraction`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:1136) and [`TestBuildCILTree_BinaryDivision`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:1234) |
| Unary negation       | `unary_negation.cpl`            | [`TestBuildCILTree_NegationInExpression`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:1368)                                                                                                                                      |
| Parentheses          | `parenthesized.cpl`             | [`TestBuildCILTree_ComplexExpression`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:1489)                                                                                                                                         |
| Deep nesting         | `complex_expression.cpl`        | [`TestBuildCILTree_ComplexExpression`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:1489)                                                                                                                                         |

### Task 6b — IO Spec Examples

Task 6b creates IO specification examples that reference output metrics. The six metrics declared in these example files (`sum_metric`, `precedence_metric`, `sub_div_metric`, `negate_metric`, `paren_metric`, `complex_metric`) serve as **declared outputs** that can be selected in an IO spec. Because the example files are metrics-only, they can be loaded alongside an IO spec to verify output metric name validation without requiring a full CPL document.

---

## 6. Key Files Reference

| File                                                                                                                             | Purpose                                                                                        |
| -------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| [`simple_addition.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/simple_addition.cpl)                        | Binary addition: `input_a + input_b`                                                           |
| [`multiplication_precedence.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/multiplication_precedence.cpl)    | Operator precedence: `a * b + c` → `(a * b) + c`                                               |
| [`subtraction_division.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/subtraction_division.cpl)              | Subtraction and division: `a - b / c` → `a - (b / c)`                                          |
| [`unary_negation.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/unary_negation.cpl)                          | Unary negation: `-a + b` → `(-a) + b`                                                          |
| [`parenthesized.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/parenthesized.cpl)                            | Parentheses overriding precedence: `(a + b) * (c - d)`                                         |
| [`complex_expression.cpl`](src/chronocone_planning_language/examples/metrics_arithmetic/complex_expression.cpl)                  | Deeply nested expression: `a * b + c / d - e`                                                  |
| [`example_definitions.go`](src/chronocone_planning_language/examples/metrics_arithmetic/example_definitions.go)                  | Structural definitions (Phase A) documenting the complete CPL structure each file represents   |
| [`metrics_arithmetic_test.go`](src/chronocone_planning_language/examples/metrics_arithmetic/metrics_arithmetic_test.go)          | Companion test verifying all six files parse successfully                                      |
| [`metrics_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:836)                 | Integration tests (lines 836–974) verifying parse→analyze pipeline for arithmetic expressions  |
| [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4)                 | Root grammar with shared `expression`, `additionExpr`, `multiplicationExpr`, `unaryExpr` rules |
| [`CPLMetrics.g4`](src/chronocone_planning_language/src/main/antlr/CPLMetrics.g4)                                                 | Metrics grammar with `calculateClause` referencing shared `expression` rule                    |
| [`metrics_analysis.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go)                               | Analysis phase containing `buildCILTree` (postfix evaluator) and `collectInputRefs`            |
| [`cil_binary_op.go`](src/chronocone_planning_language/src/main/go/cil/cil_binary_op.go)                                          | `CILBinaryOp` struct definition                                                                |
| [`cil_unary_op.go`](src/chronocone_planning_language/src/main/go/cil/cil_unary_op.go)                                            | `CILUnaryOp` struct definition                                                                 |
| [`cil_literal.go`](src/chronocone_planning_language/src/main/go/cil/cil_literal.go)                                              | `CILLiteral` struct definition                                                                 |
| [`cil_metric_node.go`](src/chronocone_planning_language/src/main/go/cil/cil_metric_node.go)                                      | `CILMetricNode` and `CILInputRef` struct definitions                                           |
| [`Story184_cpl-metrics-arithmetic-parse-analyze.md`](memory-bank/stories/Story184_cpl-metrics-arithmetic-parse-analyze.md)       | Parent story file defining Tasks 1–6                                                           |
| [`ERO005_metrics-parsing-phase.md`](memory-bank/specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md) | Increment 2 parsing design document                                                            |
| [`ERO005_metrics-analysis-phase.md`](memory-bank/specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)    | Increment 2 analysis design document                                                           |

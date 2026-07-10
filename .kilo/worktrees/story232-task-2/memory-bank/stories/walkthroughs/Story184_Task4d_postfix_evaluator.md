# Story184 Task 4d — Postfix Evaluator Walkthrough

## 1. Overview

The postfix evaluator, implemented as [`buildCILTree`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:249), converts a `ParserMetricDefinition.CalculatePostfix` (`[]interface{}`) into a [`CILMetricNode`](src/chronocone_planning_language/src/main/go/cil/cil_metric_node.go:7) containing a structured CIL expression tree. This is the bridge between the parser's flat postfix representation and the analysis phase's hierarchical CIL tree, enabling downstream synthesis to traverse and evaluate arithmetic expressions.

**Input:** A `ParserMetricDefinition` whose `CalculatePostfix` field is a slice of `interface{}` elements — operands (numeric literals, string identifiers) and operators (symbol table definition type constants).

**Output:** A `*CILMetricNode` with:

- `CILNode`: The root of the CIL expression tree (a [`CILLiteral`](src/chronocone_planning_language/src/main/go/cil/cil_literal.go:7), [`CILBinaryOp`](src/chronocone_planning_language/src/main/go/cil/cil_binary_op.go:8), [`CILUnaryOp`](src/chronocone_planning_language/src/main/go/cil/cil_unary_op.go:8), or [`CILInputRef`](src/chronocone_planning_language/src/main/go/cil/cil_metric_node.go:13))
- `InputRefs`: A deduplicated slice of `CILInputRef` entries collected from all identifier references in the expression tree

**Design Reference:** [ERO005 Increment 2](memory-bank/specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md:355) — Simple Arithmetic in Calculate.

---

## 2. Algorithm Description

The evaluator uses a classic **stack-based postfix evaluation** algorithm. Postfix (Reverse Polish Notation) eliminates the need for parentheses and operator precedence handling — the order of elements in the postfix slice already encodes precedence.

### Step-by-Step Algorithm

1. **Initialize** an empty stack (`[]cil.CILNode`).

2. **Iterate** over each element in `metricDef.CalculatePostfix` in order.

3. **Type-switch** on each element:
   - **Operand** (int, int64, float64, string): Push the corresponding CIL node onto the stack.
   - **Binary operator** (`PlusOperatorDefinition`, `MinusOperatorDefinition`, `MultiplyOperatorDefinition`, `DivideOperatorDefinition`): Pop two nodes from the stack (first popped = right operand, second popped = left operand), construct a `CILBinaryOp`, and push it back.
   - **Unary operator** (`NegateOperatorDefinition`): Pop one node from the stack, construct a `CILUnaryOp`, and push it back.
   - **Unknown type**: Return an error.

4. **Verify** the stack has exactly one element after processing all postfix elements. If not, return an error.

5. **Walk** the resulting CIL tree via [`collectInputRefs`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:342) to collect all `CILInputRef` nodes.

6. **Return** a `CILMetricNode` with the root `CILNode` and the collected `InputRefs`.

### Stack Convention

For binary operators, the **first popped** value is the **right** operand and the **second popped** is the **left** operand. This follows standard postfix convention:

```
Postfix: [a, b, +]
Step 1: Push a        → stack: [a]
Step 2: Push b        → stack: [a, b]
Step 3: Pop b (right), Pop a (left) → Push CILBinaryOp{Op: "+", Left: a, Right: b}
```

---

## 3. Operand/Operator Mapping Table

| Postfix Element Type    | Go Type / Constant                       | CIL Node Produced                                                                                                | Stack Effect  |
| ----------------------- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ------------- |
| Integer literal         | `int`                                    | [`CILLiteral{Value: float64(v)}`](src/chronocone_planning_language/src/main/go/cil/cil_literal.go:7)             | Push 1        |
| 64-bit integer literal  | `int64`                                  | [`CILLiteral{Value: float64(v)}`](src/chronocone_planning_language/src/main/go/cil/cil_literal.go:7)             | Push 1        |
| Float literal           | `float64`                                | [`CILLiteral{Value: v}`](src/chronocone_planning_language/src/main/go/cil/cil_literal.go:7)                      | Push 1        |
| Identifier reference    | `string`                                 | [`CILInputRef{Kind: "metric", Name: v}`](src/chronocone_planning_language/src/main/go/cil/cil_metric_node.go:13) | Push 1        |
| Addition operator       | `symboltable.PlusOperatorDefinition`     | [`CILBinaryOp{Op: "+", Left, Right}`](src/chronocone_planning_language/src/main/go/cil/cil_binary_op.go:8)       | Pop 2, Push 1 |
| Subtraction operator    | `symboltable.MinusOperatorDefinition`    | [`CILBinaryOp{Op: "-", Left, Right}`](src/chronocone_planning_language/src/main/go/cil/cil_binary_op.go:8)       | Pop 2, Push 1 |
| Multiplication operator | `symboltable.MultiplyOperatorDefinition` | [`CILBinaryOp{Op: "*", Left, Right}`](src/chronocone_planning_language/src/main/go/cil/cil_binary_op.go:8)       | Pop 2, Push 1 |
| Division operator       | `symboltable.DivideOperatorDefinition`   | [`CILBinaryOp{Op: "/", Left, Right}`](src/chronocone_planning_language/src/main/go/cil/cil_binary_op.go:8)       | Pop 2, Push 1 |
| Negation operator       | `symboltable.NegateOperatorDefinition`   | [`CILUnaryOp{Op: "negate", Operand}`](src/chronocone_planning_language/src/main/go/cil/cil_unary_op.go:8)        | Pop 1, Push 1 |

**Type note:** Operators are matched via `symboltable.DefinitionType` (a named `int` type). The type switch uses `case symboltable.DefinitionType:` to catch all operator constants, then a nested `switch v` to dispatch to the specific operator.

---

## 4. Example Walkthrough

### Expression: `a * b + c`

**CPL Source:**

```cpl
calculate
    a * b + c
```

**Postfix (from parser):**

```go
[]interface{}{"a", "b", MultiplyOperatorDefinition, "c", PlusOperatorDefinition}
```

This postfix encodes `a b * c +` which, due to operator precedence applied during parsing, evaluates as `(a * b) + c`.

### Step-by-Step Evaluation

| Step | Element                      | Action                                                                                                                 | Stack After                  |
| ---- | ---------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| 1    | `"a"` (string)               | Push `CILInputRef{Kind:"metric", Name:"a"}`                                                                            | `[InputRef(a)]`              |
| 2    | `"b"` (string)               | Push `CILInputRef{Kind:"metric", Name:"b"}`                                                                            | `[InputRef(a), InputRef(b)]` |
| 3    | `MultiplyOperatorDefinition` | Pop `InputRef(b)` as right, Pop `InputRef(a)` as left. Push `CILBinaryOp{Op:"*", Left:InputRef(a), Right:InputRef(b)}` | `[BinaryOp(*)]`              |
| 4    | `"c"` (string)               | Push `CILInputRef{Kind:"metric", Name:"c"}`                                                                            | `[BinaryOp(*), InputRef(c)]` |
| 5    | `PlusOperatorDefinition`     | Pop `InputRef(c)` as right, Pop `BinaryOp(*)` as left. Push `CILBinaryOp{Op:"+", Left:BinaryOp(*), Right:InputRef(c)}` | `[BinaryOp(+)]`              |

### Resulting CIL Tree

```
CILMetricNode
├── CILNode: CILBinaryOp{Op: "+"}
│   ├── Left: CILBinaryOp{Op: "*"}
│   │   ├── Left: CILInputRef{Kind: "metric", Name: "a"}
│   │   └── Right: CILInputRef{Kind: "metric", Name: "b"}
│   └── Right: CILInputRef{Kind: "metric", Name: "c"}
└── InputRefs: [
      {Kind: "metric", Name: "a"},
      {Kind: "metric", Name: "b"},
      {Kind: "metric", Name: "c"}
    ]
```

### ASCII Tree Visualization

```
         (+)
        /   \
      (*)   (c)
     /   \
   (a)   (b)
```

---

## 5. Error Cases

The evaluator returns descriptive errors for all invalid postfix sequences:

| Error Condition              | Error Message Pattern                                                       | Trigger                                                                                                     |
| ---------------------------- | --------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| **Empty postfix**            | `"postfix stack has 0 elements after evaluation, expected exactly 1"`       | `CalculatePostfix` is `[]interface{}{}`                                                                     |
| **Binary stack underflow**   | `"postfix stack underflow: need 2 operands for binary operator '+'"`        | Fewer than 2 elements on stack when a binary operator is encountered (e.g., `[42, PlusOperatorDefinition]`) |
| **Unary stack underflow**    | `"postfix stack underflow: need 1 operand for unary operator 'negate'"`     | Empty stack when `NegateOperatorDefinition` is encountered (e.g., `[NegateOperatorDefinition]`)             |
| **Excess elements on stack** | `"postfix stack has N elements after evaluation, expected exactly 1"`       | Multiple operands with no operators to consume them (e.g., `[1, 2, 3]`)                                     |
| **Unexpected type**          | `"unexpected postfix element type bool"`                                    | An element in `CalculatePostfix` has an unrecognized Go type (e.g., `true` of type `bool`)                  |
| **Unknown operator value**   | `"unexpected postfix element type symboltable.DefinitionType with value N"` | A `symboltable.DefinitionType` value that doesn't match any known operator constant                         |

All errors are wrapped by [`AnalyzeMetricScope`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:85) with the metric name context: `"failed to build CIL tree for metric <name>: <underlying error>"`.

---

## 6. InputRefs Collection

The [`collectInputRefs`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:342) function performs a recursive tree walk to discover all `CILInputRef` nodes in the expression tree.

### Algorithm

1. Initialize a `seen` map (`map[string]bool`) for deduplication and an empty `refs` slice.
2. Call [`collectInputRefsRecursive`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:350) on the root node.
3. The recursive function type-switches on each node:
   - **`*CILInputRef`**: If the name hasn't been seen, add it to `seen` and append to `refs`.
   - **`*CILBinaryOp`**: Recurse into `Left` and `Right` children.
   - **`*CILUnaryOp`**: Recurse into `Operand`.
   - **`*CILLiteral`**: Leaf node — no children to recurse into.
   - **`nil`**: Guard against nil nodes.

### Deduplication

InputRefs are deduplicated by **name** (not by pointer identity). If the same identifier appears multiple times in an expression (e.g., `a + a * b`), it appears only once in the `InputRefs` slice. The order of first occurrence is preserved.

### Example

For the expression `(a + b) * a` (postfix: `["a", "b", PlusOperatorDefinition, "a", MultiplyOperatorDefinition]`):

```
InputRefs: [{Kind: "metric", Name: "a"}, {Kind: "metric", Name: "b"}]
```

`"a"` appears twice in the tree but only once in `InputRefs`.

---

## 7. Integration Points

### Call Chain

```
Analyze(rootTable)
  └── WalkSymbolTableWithHandlers(rootTable, handlers, ...)
        └── AnalyzeMetricScope(table, context, analyzer)     ← entry point
              ├── validateInput(...)                          ← validates inputs against concepts/measures
              └── buildCILTree(metricDef)                     ← postfix evaluator
                    ├── [stack-based postfix evaluation]
                    └── collectInputRefs(stack[0])            ← tree walk for InputRefs
```

### Within [`AnalyzeMetricScope`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:85)

1. The function retrieves the `ParserMetricDefinition` from `table.Data`.
2. Validates the metric name and checks for duplicates.
3. Validates each input reference against the analyzer's `concepts` and `measures` maps.
4. Calls `buildCILTree(metricDef)` to construct the CIL expression tree.
5. Assembles the `AnalysisMetricDefinition` with the CIL root and stores it in `analyzer.metrics`.

### Downstream Consumers

- **Synthesis phase**: The `CILMetricNode.CILNode` tree is traversed to generate computation graphs for tensor operations.
- **Dependency DAG** (future Increment 8): `InputRefs` enables building the metric dependency graph for topological ordering and cycle detection.
- **CIL generation**: The expression tree is serialized into the Chronocone Intermediate Language for execution.

### Increment 1 Backwards Compatibility

When `CalculatePostfix` contains a single literal (e.g., `[]interface{}{0}`), the evaluator produces a `CILMetricNode` with a `CILLiteral` root and empty `InputRefs`. This is identical to the original Increment 1 behavior, ensuring no breaking changes.

---

## 8. Test Coverage Summary

The test file [`metrics_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:822) contains **17 top-level test functions (22 total test cases including sub-cases)** for the postfix evaluator, organized into three groups:

### Literal and Identifier Tests (7 tests)

| #   | Test Name                             | What It Verifies                                                                                                                                                                             |
| --- | ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `TestBuildCILTree_LiteralValue`       | Table-driven test with 7 sub-cases: zero, positive int, negative int, float64, int64, negative float64. Each produces a `CILLiteral` with the correct `float64` value and empty `InputRefs`. |
| 2   | `TestBuildCILTree_IdentifierInputRef` | A single string `"my_input"` produces a `CILInputRef` with `Kind: "metric"` and the correct `Name`. Also verifies `InputRefs` collection on `CILMetricNode`.                                 |

### Binary Operator Tests (4 tests)

| #   | Test Name                               | What It Verifies                                                                                                                          |
| --- | --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| 3   | `TestBuildCILTree_BinaryAddition`       | Postfix `[42, "input_a", PlusOperatorDefinition]` → `CILBinaryOp{Op:"+"}` with `Left: CILLiteral{42}`, `Right: CILInputRef{input_a}`.     |
| 4   | `TestBuildCILTree_BinarySubtraction`    | Postfix `[10, 3, MinusOperatorDefinition]` → `CILBinaryOp{Op:"-"}` with both operands as `CILLiteral`.                                    |
| 5   | `TestBuildCILTree_BinaryMultiplication` | Postfix `["id_a", "id_b", MultiplyOperatorDefinition]` → `CILBinaryOp{Op:"*"}` with both operands as `CILInputRef`.                       |
| 6   | `TestBuildCILTree_BinaryDivision`       | Postfix `[100, "divisor", DivideOperatorDefinition]` → `CILBinaryOp{Op:"/"}` with `Left: CILLiteral{100}`, `Right: CILInputRef{divisor}`. |

### Unary Operator Tests (3 tests)

| #   | Test Name                                    | What It Verifies                                                                                                                                                                                      |
| --- | -------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 7   | `TestBuildCILTree_UnaryNegation`             | Postfix `[5, NegateOperatorDefinition]` → `CILUnaryOp{Op:"negate"}` with `Operand: CILLiteral{5}`.                                                                                                    |
| 8   | `TestBuildCILTree_UnaryNegationOfIdentifier` | Postfix `["my_input", NegateOperatorDefinition]` → `CILUnaryOp{Op:"negate"}` with `Operand: CILInputRef{my_input}`.                                                                                   |
| 9   | `TestBuildCILTree_NegationInExpression`      | Postfix `["id_a", "id_b", NegateOperatorDefinition, PlusOperatorDefinition]` (representing `a + (-b)`) → 3-level tree: `BinaryOp(+)` with `Left: InputRef(a)`, `Right: UnaryOp(negate, InputRef(b))`. |

### Complex Expression Tests (3 tests)

| #   | Test Name                                   | What It Verifies                                                                                                                                                                                                                                                                          |
| --- | ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 10  | `TestBuildCILTree_ArithmeticWithPrecedence` | Postfix `["id_a", "id_b", MultiplyOperatorDefinition, 0, PlusOperatorDefinition]` (representing `a * b + 0`) → 3-level tree with `BinaryOp(+)` at root, `BinaryOp(*)` as left child, `CILLiteral(0)` as right child.                                                                      |
| 11  | `TestBuildCILTree_ComplexExpression`        | Postfix `["id_a", "id_b", PlusOperatorDefinition, "id_c", "id_d", DivideOperatorDefinition, MinusOperatorDefinition]` (representing `(a + b) - (c / d)`) → 3-level tree with `BinaryOp(-)` at root, `BinaryOp(+)` as left child, `BinaryOp(/)` as right child. All 4 leaf nodes verified. |
| 12  | `TestBuildCILTree_InputRefsCollection`      | Postfix `["id_a", "id_b", PlusOperatorDefinition, "id_a", MultiplyOperatorDefinition]` (representing `(a + b) * a`) → verifies `InputRefs` contains exactly 2 unique entries (`id_a`, `id_b`) despite `id_a` appearing twice.                                                             |

### Error Case Tests (5 tests)

| #   | Test Name                                   | What It Verifies                                                                                   |
| --- | ------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| 13  | `TestBuildCILTree_EmptyPostfix`             | Empty `CalculatePostfix` returns an error containing `"postfix"`.                                  |
| 14  | `TestBuildCILTree_UnbalancedStack_BinaryOp` | Postfix `[42, PlusOperatorDefinition]` (only 1 operand for binary op) returns `"underflow"` error. |
| 15  | `TestBuildCILTree_UnbalancedStack_UnaryOp`  | Postfix `[NegateOperatorDefinition]` (no operand for unary op) returns `"underflow"` error.        |
| 16  | `TestBuildCILTree_ExcessElementsOnStack`    | Postfix `[1, 2, 3]` (3 operands, no operators) returns `"stack"` error.                            |
| 17  | `TestBuildCILTree_UnexpectedType`           | Postfix `[true]` (bool type) returns `"unexpected"` error.                                         |

### Test Design Pattern

All postfix evaluator tests use **direct symbol table construction** rather than full parsing. This isolates the evaluator from parser changes and allows precise control over the `CalculatePostfix` slice content. Each test:

1. Creates a `SymbolTable` with a `MetricScope` child
2. Sets `table.Data` to a `ParserMetricDefinition` with the desired `CalculatePostfix`
3. Calls `analysis.AnalyzeMetricScope(metricScope, nil, analyzer)` directly
4. Asserts on the resulting `CILMetricNode` structure

---

## Key Files Reference

| File                                                                                                                          | Purpose                                                              |
| ----------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| [`metrics_analysis.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go)                            | Contains `buildCILTree` (line 249) and `collectInputRefs` (line 342) |
| [`metrics_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go)                  | 17 postfix evaluator tests starting at line 822                      |
| [`cil_binary_op.go`](src/chronocone_planning_language/src/main/go/cil/cil_binary_op.go)                                       | `CILBinaryOp` struct definition                                      |
| [`cil_unary_op.go`](src/chronocone_planning_language/src/main/go/cil/cil_unary_op.go)                                         | `CILUnaryOp` struct definition                                       |
| [`cil_literal.go`](src/chronocone_planning_language/src/main/go/cil/cil_literal.go)                                           | `CILLiteral` struct definition                                       |
| [`cil_metric_node.go`](src/chronocone_planning_language/src/main/go/cil/cil_metric_node.go)                                   | `CILMetricNode` and `CILInputRef` struct definitions                 |
| [`ERO005_metrics-analysis-phase.md`](memory-bank/specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md) | Increment 2 design (lines 355–447)                                   |

# ERO005: Metrics Synthesis Phase

## Title

Metrics Synthesis Phase — CIL Tree Walking and Go Code Generation

## Purpose/Goal

Define the synthesis phase logic for generating Go code from metric CIL trees, including CIL tree walking for each CIL node type, tensor computation code generation, IO specification integration, and multi-metric executable generation across all 8 increments of the metrics incremental implementation as defined in [ERO005 — Metrics Incremental Implementation](../../eros/ERO005_metrics-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- Epic005_cpl-metrics-implementation.md

## Context

### Situation

The synthesis phase is the final stage of the CPL compilation pipeline. It receives CIL trees from the analysis phase — one tree per metric and one single-node tree per connector — and generates executable Go code. The synthesis phase walks each CIL tree using the [`CILAbstractTransformer`](../synthesis.md#cilabstracttransformer), which organizes code generation into a hierarchy of [`ParallelOperationGroup`](../synthesis.md#paralleloperationgroup), [`ComputationGroup`](../synthesis.md#computationgroup), and [`Computation`](../synthesis.md#computation) objects.

The analysis phase ERO ([ERO005_metrics-analysis-phase.md](../../go-analysis/eros/ERO005_metrics-analysis-phase.md)) defines how CIL trees are constructed from postfix expression sequences. This ERO defines how those CIL trees are walked to generate executable Go code that performs tensor computations, handles cross-metric dependencies, integrates IO specifications, and produces output.

**Key Principle:** The synthesis phase walks CIL trees using the standard [`CILAbstractTransformer`](../synthesis.md#cilabstracttransformer) infrastructure. Each CIL node type maps to a specific Go code generation pattern. The same walking mechanism is used for all increments — even Increment 1 where the CIL tree is a single [`CILLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#cil-node-types-increment-1) node.

### Constraints

- Must use the existing [`CILAbstractTransformer`](../synthesis.md#cilabstracttransformer) infrastructure for CIL tree walking
- Must follow the [Three-Level Function Hierarchy](../tensor-representation.md#three-level-function-hierarchy) for tensor operations
- Must generate code compatible with the [`ChronoconeComputationController`](../synthesis.md#chronoconecomputationcontroller) and [`MetricComputation`](../synthesis.md#metriccomputation) synchronization model
- Must support the Setup/Compute/Cleanup lifecycle defined in [synthesis.md](../synthesis.md#setup-compute-and-cleanup-lifecycle)
- Must integrate IO specifications for output metric selection, dimensional configuration, and compute architecture
- Must generate code that respects the topological evaluation order from the analysis phase's dependency DAG
- Must support the incremental feature set of each increment
- Types are initially represented as `float64` (detailed type selection is deferred to [type-selection.md](../../type-selection.md))

### Requirements

- Walk CIL trees to generate Go code for each CIL node type encountered
- Generate `computeX()` functions for each metric that create tensor operands, submit computation chunks, and set results via [`MetricComputation.SetTensor()`](../synthesis.md#metriccomputation)
- Generate connector loading code that reads parquet data into tensors (leveraging [ERO004_connectors-synthesis-phase.md](ERO004_connectors-synthesis-phase.md))
- Generate `Setup()`, `Compute()`, and `Cleanup()` lifecycle functions
- Generate `main()` function that loads IO specification, configures the controller, and orchestrates the lifecycle
- Integrate IO specification for output metric selection, dimensional/temporal configuration, and compute architecture
- Support tensor calculation parallelism using compute architecture from IO spec (Increment 3)
- Generate multi-metric executables with topological evaluation order (Increment 8)
- Implement dead-code elimination in generated output (only synthesize reachable metrics)

### Assumptions

- CIL trees are complete and valid (all semantic validation performed during analysis)
- The analysis phase has produced a valid dependency DAG with topological ordering
- Connector synthesis (parquet reading) is already implemented per [ERO004_connectors-synthesis-phase.md](ERO004_connectors-synthesis-phase.md)
- The [`CILAbstractTransformer`](../synthesis.md#cilabstracttransformer) infrastructure is available and can be extended with metric-specific node handlers
- The [`ChronoconeComputationController`](../synthesis.md#chronoconecomputationcontroller) and tensor operation primitives are available in the synthesis support library
- Each increment produces a working executable that reads CPL, IO spec, and data, then outputs computed metric values

## Evaluation Criteria

1. **End-to-End Completeness** (High importance): Each increment must produce a working executable that reads CPL, IO spec, and data, then outputs computed metric values.
2. **Correctness** (High importance): Generated code must correctly implement the semantics of each CIL node type (literals, arithmetic, conditionals, time shifting).
3. **Pipeline Consistency** (High importance): The same CIL tree walking mechanism must be used for all increments, validating the pipeline architecture from the start.
4. **IO Specification Integration** (Medium importance): Generated code must correctly consume IO specifications for output selection, dimensional configuration, and compute architecture.
5. **Performance** (Medium importance): Generated code should leverage tensor chunking and parallelism where appropriate.

## Options Considered

Three approaches were evaluated for synthesizing Go code from metric CIL trees. **Full CIL tree walker up front** was rejected because it violates the incremental approach — no working end-to-end pipeline would exist until all handlers are complete, and there is high risk of building handlers for node types whose semantics aren't fully validated. **Template-based code generation** was rejected because templates cannot easily handle the hierarchical computation grouping (ParallelOperationGroup → ComputationGroup → Computation) or cross-metric dependency synchronization via MetricComputation channels. **Incremental CIL tree walking** was selected because it mirrors the successful pattern used for connectors (ERO004): Increment 1 handles only `CILLiteral` nodes, validating the entire pipeline immediately; each subsequent increment adds handlers for new CIL node types as they are introduced by the analysis phase.

## Decision

**Chosen Option:** Option 1 — Incremental CIL Tree Walking

**Rationale:**
The incremental approach delivers a working end-to-end pipeline in the very first increment: a CIL tree with a single [`CILLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#cil-node-types-increment-1) node is walked to generate a Go function that returns `0`. This validates the entire synthesis pipeline immediately — CIL tree walking, code generation, Setup/Compute/Cleanup lifecycle, and executable compilation.

Each subsequent increment adds handlers for new CIL node types as they are introduced by the analysis phase. The transformer grows organically: Increment 1 requires only a literal handler, Increment 2 adds binary op and input ref handlers, and so on. This mirrors the successful pattern used for connectors (ERO004) and is consistent with all existing synthesis phase EROs (ERO001–ERO004).

## Implications

### Positive Implications

- Working end-to-end pipeline validated in Increment 1, reducing architectural risk
- Each increment produces a demonstrable, testable result
- CIL tree walking infrastructure is exercised and validated early
- Consistent with the existing incremental implementation pattern across all CPL components
- Generated code structure (compute functions, lifecycle, synchronization) is established early and extended incrementally

### Negative Implications / Trade-offs

- Early increments produce trivial generated code (e.g., `return 0`), providing limited demonstration value until Increment 2
- The transformer must be extended across multiple increments, requiring careful coordination with analysis phase changes
- Generated code patterns evolve: early increments use simple scalar returns; later increments introduce tensor operations, chunking, and parallelism
- Increment 8 (full IO spec integration) is needed before multi-metric executables with topological ordering can be generated

### Future Considerations

- Detailed type selection (fixed-point, overflow detection, `math/big`) is deferred — see [type-selection.md](../../type-selection.md)
- Aggregator and spreader synthesis will require additional EROs
- Lookup table synthesis is deferred to a future design
- Horizontal scaling (multi-machine) is deferred; current design assumes single-machine execution

## Synthesis by Increment

### Increment 1: Basic Metric Declarations with Identity Calculation

**Overview**

Increment 1 introduces the minimal metric synthesis infrastructure. The analysis phase produces one CIL tree per metric containing a single [`CILLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#cil-node-types-increment-1) node with value `0`, and one single-node CIL tree per connector ([`CILConnectorReferenceNode`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#cil-node-types-increment-1)). The synthesis phase walks these trees to generate a Go executable that loads connector data and returns `0` for each metric cell.

**CIL Tree Walking — Literal Nodes**

The [`CILAbstractTransformer`](../synthesis.md#cilabstracttransformer) is extended with a handler for [`CILLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#cil-node-types-increment-1) nodes:

```go
// VisitCILLiteral handles a literal value node in the CIL tree.
// Generates a Go constant expression.
func (t *CILAbstractTransformer) VisitCILLiteral(node *CILLiteral) error {
    // Push a code snippet that returns the literal value
    snippet := fmt.Sprintf("return %s", node.Value)
    t.currentComputation.PushSnippet(snippet)
    return nil
}
```

**Generated compute function (Increment 1):**

```go
// computeRevenue computes the revenue metric.
// In Increment 1, this simply returns 0 for all cells.
func computeRevenue(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    // Allocate output tensor with dimensions from IO spec
    result, err := controller.AllocateTensor[float64]("Product", "Region", "Time")
    if err != nil {
        // Handle error
        return
    }

    // Fill tensor with literal value 0
    for i := range result.Data {
        result.Data[i] = 0.0
    }

    // Set the result to unblock dependent computations
    controller.GetMetricComputation("revenue").SetTensor(result)
}
```

**Connector Loading**

Connector CIL trees ([`CILConnectorReferenceNode`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#cil-node-types-increment-1)) are walked to generate parquet-reading code, reusing the infrastructure from [ERO004_connectors-synthesis-phase.md](ERO004_connectors-synthesis-phase.md). The generated code includes the `connectorMetadata` map and `loadParquetConnector` function.

**Generated main function (Increment 1):**

```go
func main() {
    if len(os.Args) < 2 {
        log.Fatal("Usage: <program> <io_specification.json>")
    }

    ioSpec, err := LoadIOSpecification(os.Args[1])
    if err != nil {
        log.Fatalf("Failed to load IO specification: %v", err)
    }

    controller := NewChronoconeComputationController()
    controller.SetArchitectureConfig(&ioSpec.Architecture)

    if err := Setup(controller); err != nil {
        log.Fatalf("Setup failed: %v", err)
    }

    // Load input tensors from connectors
    inputs := make(map[string]*ChronoconeTensor[any])
    for _, inputSpec := range ioSpec.Inputs {
        meta, ok := connectorMetadata[inputSpec.ConnectorName]
        if !ok {
            log.Fatalf("Unknown connector: %s", inputSpec.ConnectorName)
        }
        tensor, err := loadParquetConnector(controller, &inputSpec, meta)
        if err != nil {
            log.Fatalf("Failed to load input %s: %v", inputSpec.ConnectorName, err)
        }
        inputs[inputSpec.ConnectorName] = tensor
    }

    controller.StartWorkers()

    if err := Compute(controller, inputs); err != nil {
        log.Fatalf("Compute failed: %v", err)
    }

    controller.StopWorkers()

    // Write output
    for _, outputSpec := range ioSpec.Outputs {
        if err := writeMetricOutput(controller, &outputSpec); err != nil {
            log.Fatalf("Failed to write output %s: %v", outputSpec.MetricName, err)
        }
    }

    if err := Cleanup(controller); err != nil {
        log.Fatalf("Cleanup failed: %v", err)
    }
}
```

**Setup/Compute/Cleanup (Increment 1):**

- **Setup()**: Registers [`MetricComputation`](../synthesis.md#metriccomputation) objects for all metrics. The consumer count for each metric is 1 (the output writer).
- **Compute()**: Launches a goroutine for each metric's `computeX()` function. Since metrics in Increment 1 have no dependencies on other metrics, all goroutines can run independently.
- **Cleanup()**: Releases resources.

**Limitations**

- No arithmetic expressions (added in Increment 2)
- No IO specification output metric selection (all metrics are computed; added in Increment 2)
- No tensor computation parallelism (added in Increment 3)
- No conditional expressions (added in Increment 4)
- No input aliases or where clauses (added in Increment 5)
- No time shifting (added in Increment 6)
- No short form metrics (added in Increment 7)
- No multi-metric topological ordering or dead-code elimination (added in Increment 8)

### Increment 2: Simple Arithmetic in Calculate with IO Specification Acceptance

**Overview**

Increment 2 extends the CIL tree walker to handle [`CILBinaryOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types), [`CILUnaryOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types), and [`CILInputRef`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types) nodes. The walker generates Go code that performs element-wise arithmetic using **package-level arithmetic functions** (e.g., `calculation.TensorAddFloat64`).

Arithmetic is implemented as sibling functions keyed by numeric type rather than as methods on `ChronoconeTensor[T any]`. `ChronoconeTensor` uses the constraint `T any` — the `any` constraint does not support arithmetic operators (`+`, `-`, `*`, `/`), so arithmetic cannot be defined as methods on the generic type. Package-level functions accept `*ChronoconeTensor[any]` and internally type-assert elements to the appropriate concrete type.

**Type-Dispatch via Sibling Functions**: The synthesis phase tracks Go types per measure in the `measureTypes` map. For each arithmetic operation, it selects the correct function variant based on operand types:

| CPL Numeric Type                 | Go Element Type       | Arithmetic Functions                                                                                                                    |
| :------------------------------- | :-------------------- | :-------------------------------------------------------------------------------------------------------------------------------------- |
| `Number` (float)                 | `float64`             | `TensorAddFloat64`, `TensorSubtractFloat64`, `TensorMultiplyFloat64`, `TensorDivideFloat64`, `TensorNegateFloat64`, and scalar variants |
| `Number(precision=N)`            | `int64` (fixed-point) | `TensorAddInt`, `TensorSubtractInt`, `TensorMultiplyInt`, `TensorDivideInt`, `TensorNegateInt`                                          |
| `Number(precision=N)` (overflow) | `*big.Rat`            | `TensorAddRational`, etc. (via `math/big`)                                                                                              |
| `BigFloat`                       | `*big.Float`          | `TensorAddBigFloat`, etc.                                                                                                               |
| `BigInt`                         | `*big.Int`            | `TensorAddBigInt`, etc.                                                                                                                 |

All arithmetic functions reside in the `calculation` package, separate from the tensor type definition. Each returns `(*ChronoconeTensor[any], error)` and validates dimension compatibility, nil inputs, and element-type consistency before performing the operation. See [Type Selection for Numeric Calculations](../../type-selection.md) for the full type-dispatch design.

**CIL Tree Walking — Binary Operations**

```go
// VisitCILBinaryOp handles a binary operation node in the CIL tree.
// Generates Go code using high-level tensor operation methods.
func (t *CILAbstractTransformer) VisitCILBinaryOp(node *CILBinaryOp) error {
    // Walk children first (post-order)
    if err := t.VisitNode(node.Left); err != nil {
        return err
    }
    if err := t.VisitNode(node.Right); err != nil {
        return err
    }

    // Determine operand types to choose scalar vs tensor operation
    leftIsScalar := isScalarNode(node.Left)
    rightIsScalar := isScalarNode(node.Right)

    var snippet string
    switch {
    case leftIsScalar && rightIsScalar:
        // Both scalars: compute at compile time
        snippet = fmt.Sprintf("return %s %s %s", leftValue, node.Operator, rightValue)
    case leftIsScalar:
        // Scalar op Tensor: use high-level scalar-tensor method
        snippet = fmt.Sprintf("result, err := rightOperand.%sScalar(controller, %s)", opMethodName, leftValue)
    case rightIsScalar:
        // Tensor op Scalar: use high-level tensor-scalar method
        snippet = fmt.Sprintf("result, err := leftOperand.%sScalar(controller, %s)", opMethodName, rightValue)
    default:
        // Tensor op Tensor: use high-level tensor-tensor method
        snippet = fmt.Sprintf("result, err := leftOperand.%s(controller, rightOperand)", opMethodName)
    }

    t.currentComputation.PushSnippet(snippet)
    return nil
}
```

**CIL Tree Walking — Input References**

```go
// VisitCILInputRef handles an input reference node in the CIL tree.
// Generates Go code to access the input tensor for the current cell.
func (t *CILAbstractTransformer) VisitCILInputRef(node *CILInputRef) error {
    // Determine if this is a metric reference or dimension reference
    if t.isMetricReference(node.Alias) {
        // Reference to another metric: use MetricComputation.GetTensor()
        snippet := fmt.Sprintf("%sTensor := controller.GetMetricComputation(\"%s\").GetTensor()",
            node.Alias, node.Alias)
        t.currentComputation.PushSnippet(snippet)
    } else {
        // Reference to a dimension: access from inputs map
        snippet := fmt.Sprintf("%sValue := inputs[\"%s\"]", node.Alias, node.Alias)
        t.currentComputation.PushSnippet(snippet)
    }
    return nil
}
```

**Generated compute function (Increment 2):**

For the CPL expression `demand_forecast * Price + 10`:

```go
func computeRevenue(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[any],
) {
    // Get input tensors
    demandForecastTensor := controller.GetMetricComputation("demand_forecast").GetTensor()
    priceTensor := inputs["Price"]

    // Step 1: demand_forecast * Price (tensor-tensor multiply)
    temp1, err := calculation.TensorMultiplyFloat64(controller, demandForecastTensor, priceTensor)
    if err != nil {
        // Handle error
        return
    }

    // Step 2: temp1 + 10 (tensor-scalar add)
    result, err := calculation.TensorAddScalarFloat64(controller, temp1, 10.0)
    if err != nil {
        // Handle error
        return
    }

    controller.GetMetricComputation("revenue").SetTensor(result)

    // Release input tensors
    controller.GetMetricComputation("demand_forecast").ReleaseTensor()
}
```

**IO Specification Integration**

Starting in Increment 2, the synthesis phase reads the IO specification to determine:

- Which metrics are output targets (only these have their results written)
- The dimensions, celltype, and grain for each output metric
- The compute architecture configuration (used in Increment 3)

The generated `main()` function iterates over `ioSpec.Outputs` to write only the specified output metrics.

**Limitations**

- No conditional expressions (added in Increment 4)
- No comparison or logical operators (added in Increment 4)
- No `Nil` literal handling (added in Increment 4)

### Increment 3: Tensor Calculation Parallelism

**Overview**

Increment 3 is **synthesis-only** — there are no changes to the parser or analysis phases. The synthesis phase uses the compute architecture from the IO specification (CPU cores, cache size, memory) to configure tensor chunking and worker goroutine parallelism.

**Synthesis Changes**

The generated code already uses the high-level tensor operation methods ([`Add`](../tensor-representation.md#three-level-function-hierarchy), [`Multiply`](../tensor-representation.md#three-level-function-hierarchy), etc.) which internally handle chunking and parallel execution. In Increment 3, the synthesis phase:

1. Reads `ioSpec.Architecture.CPUCoreCount` to configure the number of worker goroutines
2. Reads `ioSpec.Architecture.CacheSizeBytes` to determine optimal chunk sizes
3. Generates code that calls `controller.SetArchitectureConfig()` with these values
4. The tensor library's [`CacheOptimizedChunking`](../synthesis.md#cacheoptimizedchunking) strategy uses cache size to determine chunk dimensions

No new CIL node types are introduced. The existing generated code from Increment 2 automatically benefits from parallelism because the high-level tensor operations use the controller's work channel and worker pool internally.

**Generated code pattern (Increment 3):**

```go
func main() {
    // ... load IO spec ...

    controller := NewChronoconeComputationController()

    // Configure architecture from IO spec
    // This sets up the work channel with buffer size = core count
    controller.SetArchitectureConfig(&ioSpec.Architecture)

    // ... Setup, load inputs ...

    // Start worker goroutines (count = CPUCoreCount)
    controller.StartWorkers()

    // Compute phase: chunks are submitted to work channel
    // Workers execute chunks in parallel across cores
    if err := Compute(controller, inputs); err != nil {
        log.Fatalf("Compute failed: %v", err)
    }

    controller.StopWorkers()
    // ... write outputs, cleanup ...
}
```

### Increment 4: Conditional Expressions

**Overview**

Increment 4 extends the CIL tree walker to handle [`CILConditional`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1), [`CILWhenClause`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1), [`CILComparisonOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1), [`CILLogicalOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1), [`CILNotOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1), [`CILNilLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1), and [`CILBoolLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1) nodes. The walker generates Go `if`/`else` blocks and boolean expressions.

**CIL Tree Walking — Conditionals**

```go
// VisitCILConditional handles an if/else conditional node.
// Generates Go if/else control flow.
func (t *CILAbstractTransformer) VisitCILConditional(node *CILConditional) error {
    var snippet strings.Builder

    // Walk condition
    if err := t.VisitNode(node.IfBranch.Condition); err != nil {
        return err
    }
    conditionCode := t.currentComputation.PopSnippet()

    // Walk if body
    if err := t.VisitNode(node.IfBranch.Body); err != nil {
        return err
    }
    ifBodyCode := t.currentComputation.PopSnippet()

    snippet.WriteString(fmt.Sprintf("if %s {\n    %s\n}", conditionCode, ifBodyCode))

    // Walk else-if branches
    for _, elif := range node.ElseIfs {
        if err := t.VisitNode(elif.Condition); err != nil {
            return err
        }
        elifCond := t.currentComputation.PopSnippet()

        if err := t.VisitNode(elif.Body); err != nil {
            return err
        }
        elifBody := t.currentComputation.PopSnippet()

        snippet.WriteString(fmt.Sprintf(" else if %s {\n    %s\n}", elifCond, elifBody))
    }

    // Walk else branch
    if node.ElseBranch != nil {
        if err := t.VisitNode(node.ElseBranch); err != nil {
            return err
        }
        elseBody := t.currentComputation.PopSnippet()
        snippet.WriteString(fmt.Sprintf(" else {\n    %s\n}", elseBody))
    }

    t.currentComputation.PushSnippet(snippet.String())
    return nil
}
```

**CIL Tree Walking — Comparison and Logical Operators**

Comparison operators ([`CILComparisonOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)) and logical operators ([`CILLogicalOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1), [`CILNotOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)) are handled similarly to arithmetic binary/unary operators, emitting the corresponding Go operators (`==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`, `!`).

**Nil Literal Handling**

[`CILNilLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1) nodes emit a typed `nil` in the generated Go code. The truthiness rules (Nil is falsy) are enforced by the analysis phase; the synthesis phase simply emits the nil literal.

**Generated compute function (Increment 4):**

For the CPL expression:

```
if demand_forecast > 0
    demand_forecast * Price
else
    0
end if
```

```go
func computeRevenue(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    demandForecastTensor := controller.GetMetricComputation("demand_forecast").GetTensor()
    priceTensor := inputs["Price"]

    // Condition: demand_forecast > 0
    // The comparison produces a mask tensor
    zeroTensor, _ := controller.AllocateTensor[float64](demandForecastTensor.DimensionNames()...)
    for i := range zeroTensor.Data {
        zeroTensor.Data[i] = 0.0
    }
    conditionMask, err := demandForecastTensor.GreaterThan(controller, zeroTensor)
    if err != nil {
        return
    }

    // If body: demand_forecast * Price
    ifBodyResult, err := demandForecastTensor.Multiply(controller, priceTensor)
    if err != nil {
        return
    }

    // Else body: 0
    elseBodyResult, _ := controller.AllocateTensor[float64](demandForecastTensor.DimensionNames()...)

    // Combine using mask: result = conditionMask ? ifBody : elseBody
    result, err := ifBodyResult.Where(controller, conditionMask, elseBodyResult)
    if err != nil {
        return
    }

    controller.GetMetricComputation("revenue").SetTensor(result)
    controller.GetMetricComputation("demand_forecast").ReleaseTensor()
}
```

**Note on Conditional Implementation:** The generated code uses tensor-level mask operations rather than per-element `if` statements. This enables vectorized execution on the full tensor. The [`Where`](../tensor-representation.md#three-level-function-hierarchy) method selects elements from two tensors based on a boolean mask.

### Increment 5: Input Aliases and Where Clauses

**Overview**

Increment 5 extends the synthesis phase to handle input aliases and where clauses. The analysis phase has already resolved aliases to their underlying metric/dimension references and recorded where-clause aspect filters. The synthesis phase uses this information to generate code that:

- Uses alias names as Go variable names in generated compute functions
- Applies aspect filtering (ACTUALS/FORECAST) via slice selection

**Synthesis Changes**

The CIL tree walker already uses alias names from [`CILInputRef`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types) nodes when generating variable names. No new CIL node types are introduced for aliases — the alias is simply the name stored in the [`CILInputRef.Alias`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types) field.

For where clauses, the synthesis phase generates aspect-filtering code that selects the appropriate time slice from the input tensor:

```go
// Generated code for an input with 'where aspect=ACTUALS'
actualsSlice := inputTensor.SliceByAspect("ACTUALS")
```

**No new CIL node types are introduced in this increment.** The synthesis changes are limited to how input references are resolved in generated variable names and how aspect filtering is applied.

### Increment 6: Time Shifting (Backward/Forward)

**Overview**

Increment 6 extends the CIL tree walker to handle [`CILTimeShift`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#enhanced-analysisinputdefinition-increment-6) nodes. The walker generates Go code for array index offset calculations: `currentCellIndex - shiftAmount` for backward, `currentCellIndex + shiftAmount` for forward.

**CIL Tree Walking — Time Shift**

```go
// VisitCILTimeShift handles a time-shift node in the CIL tree.
// Generates Go code for array index offset calculation.
func (t *CILAbstractTransformer) VisitCILTimeShift(node *CILTimeShift) error {
    // Walk the shift amount expression
    if err := t.VisitNode(node.Expression); err != nil {
        return err
    }
    shiftAmountCode := t.currentComputation.PopSnippet()

    direction := 1
    if node.Direction == "backward" {
        direction = -1
    }

    snippet := fmt.Sprintf(
        "shiftedIndex := currentIndex + (%d * (%s))",
        direction, shiftAmountCode,
    )
    t.currentComputation.PushSnippet(snippet)
    return nil
}
```

**Generated compute function (Increment 6):**

For a metric with `backward(weeks=1)` self-reference:

```go
func computeInventory(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    orderArrivalsTensor := controller.GetMetricComputation("order_arrivals").GetTensor()
    inventoryTensor := controller.GetMetricComputation("inventory").GetTensor()

    result, _ := controller.AllocateTensor[float64](inventoryTensor.DimensionNames()...)

    // For each cell, compute: order_arrivals[current] + inventory[current - 1]
    for i := range result.Data {
        priorWeekIndex := i - 1  // backward(weeks=1)
        var priorInventory float64
        if priorWeekIndex >= 0 {
            priorInventory = inventoryTensor.Data[priorWeekIndex]
        } else {
            priorInventory = 0.0  // Boundary: before start of time
        }
        result.Data[i] = orderArrivalsTensor.Data[i] + priorInventory
    }

    controller.GetMetricComputation("inventory").SetTensor(result)
    controller.GetMetricComputation("order_arrivals").ReleaseTensor()
}
```

**Self-Reference Handling**

Self-referencing metrics (metric referencing itself via backward/forward alias) require special handling in the generated code. The metric's own tensor must be allocated before the computation begins, and the time-shifted access reads from the same tensor being written. The topological order ensures that earlier time cells are computed before later ones.

### Increment 7: Short Form Metrics

**Overview**

Increment 7 requires **no synthesis changes**. Short form metrics (`revenue = demand_forecast * Price`) are converted to equivalent long form by the parser and analysis phases. The CIL trees produced for short form metrics are identical to those produced for equivalent long form metrics. The CIL tree walker processes them identically.

**Synthesis Changes**

None. The synthesis phase is unaffected by whether a metric was originally declared in short form or long form.

### Increment 8: Full Synthesis with IO Specification Integration

**Overview**

Increment 8 involves no new CIL node types. All syntax is already supported by Increments 1–7. Increment 8 adds:

- Multi-metric executable generation with topological evaluation order
- Dead-code elimination (only synthesize reachable metrics)
- Full IO specification integration for dimensional/temporal configuration

**Topological Evaluation Order**

The analysis phase provides a topological ordering of metrics via the dependency DAG. The synthesis phase uses this ordering to:

1. Generate `computeX()` functions in dependency order (inputs before dependents)
2. Register [`MetricComputation`](../synthesis.md#metriccomputation) objects with correct consumer counts
3. Generate the `Compute()` function that launches goroutines respecting the DAG

```go
// Generated Compute() function with topological ordering
func Compute(controller *ChronoconeComputationController, inputs map[string]*ChronoconeTensor[float64]) error {
    // Launch goroutines for root metrics (no dependencies)
    go computeDemandForecast(controller, inputs)
    go computePrice(controller, inputs)

    // Launch goroutines for dependent metrics
    // These will block on MetricComputation.GetTensor() until inputs are ready
    go computeRevenue(controller, inputs)
    go computeProfit(controller, inputs)

    // Wait for output metrics to complete
    for _, outputName := range outputMetricNames {
        controller.GetMetricComputation(outputName).GetTensor()
    }

    return nil
}
```

**Dead-Code Elimination**

The synthesis phase only generates code for metrics that are reachable from the IO specification's output metrics. The analysis phase has already removed unreachable metrics from the definition set. The synthesis phase simply iterates over the remaining metrics.

**IO Specification Integration**

The generated `main()` function uses the IO specification to:

- Determine output metrics and their dimensions, celltype, and grain
- Configure tensor allocation with correct dimension sizes
- Write output in the appropriate format

```go
// Generated output writing code
for _, outputSpec := range ioSpec.Outputs {
    metricComp := controller.GetMetricComputation(outputSpec.MetricName)
    tensor := metricComp.GetTensor()

    // Write tensor to parquet with specified dimensions and grain
    err := writeTensorToParquet(tensor, outputSpec.Dimensions, outputSpec.Grain)
    if err != nil {
        log.Fatalf("Failed to write output %s: %v", outputSpec.MetricName, err)
    }
}
```

## Synthesis Data Structures

### CIL Node Type to Go Code Mapping

| CIL Node Type                                                                                                          | Generated Go Code Pattern                                                | Introduced  |
| ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ | ----------- | --------------------------- | ----------- |
| [`CILLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#cil-node-types-increment-1)                     | Constant value or `controller.AllocateTensor` + fill loop                | Increment 1 |
| [`CILConnectorReferenceNode`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#cil-node-types-increment-1)      | `loadParquetConnector()` call (from ERO004)                              | Increment 1 |
| [`CILInputRef`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types)                            | `controller.GetMetricComputation(name).GetTensor()` or `inputs[name]`    | Increment 2 |
| [`CILBinaryOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types)                            | High-level tensor op methods (`Add`, `SubtractScalar`, `Multiply`, etc.) | Increment 2 |
| [`CILUnaryOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types)                             | High-level tensor op method (`Negate`)                                   | Increment 2 |
| [`CILComparisonOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)                      | High-level tensor comparison methods (`GreaterThan`, `EqualTo`, etc.)    | Increment 4 |
| [`CILLogicalOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)                         | Go `&&`, `                                                               |             | ` operators on mask tensors | Increment 4 |
| [`CILNotOp`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)                             | Go `!` operator on mask tensor                                           | Increment 4 |
| [`CILConditional`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)                       | `tensor.Where(conditionMask, trueBranch, falseBranch)`                   | Increment 4 |
| [`CILWhenClause`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)                        | Equivalent to chained `Where` calls                                      | Increment 4 |
| [`CILNilLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)                        | Typed `nil` or zero-value tensor                                         | Increment 4 |
| [`CILBoolLiteral`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#new-cil-node-types-1)                       | `true`/`false` constant or boolean tensor                                | Increment 4 |
| [`CILTimeShift`](../../go-analysis/eros/ERO005_metrics-analysis-phase.md#enhanced-analysisinputdefinition-increment-6) | Array index offset: `currentIndex ± shiftAmount`                         | Increment 6 |

### Incremental Structure and Function Evolution

| Increment | Synthesis Changes                                                                                                                                                                                 |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1         | `VisitCILLiteral`, `VisitCILConnectorReferenceNode` handlers; basic `computeX()`, `Setup()`, `Compute()`, `Cleanup()`, `main()` generation                                                        |
| 2         | `VisitCILBinaryOp`, `VisitCILUnaryOp`, `VisitCILInputRef` handlers; IO spec output selection; high-level tensor operation method calls                                                            |
| 3         | Architecture configuration from IO spec (`SetArchitectureConfig`); worker pool sizing; chunk size optimization                                                                                    |
| 4         | `VisitCILConditional`, `VisitCILWhenClause`, `VisitCILComparisonOp`, `VisitCILLogicalOp`, `VisitCILNotOp`, `VisitCILNilLiteral`, `VisitCILBoolLiteral` handlers; mask-based conditional execution |
| 5         | Alias-based variable naming in generated code; aspect-filtering slice selection                                                                                                                   |
| 6         | `VisitCILTimeShift` handler; array index offset calculation; self-reference boundary handling                                                                                                     |
| 7         | None (short form produces identical CIL trees)                                                                                                                                                    |
| 8         | Topological ordering in `Compute()`; dead-code elimination; full IO spec dimensional/temporal configuration; multi-metric output writing                                                          |

## Notes

**Related Documents:**

- [Metrics Design](../../metrics-design.md) — Full metrics design with core rules, expression handling, dependency DAG, and 8 end-to-end increments
- [ERO005 — Metrics Incremental Implementation](../../eros/ERO005_metrics-incremental-implementation.md) — Top-level ERO defining the 8-increment approach
- [ERO005 — Metrics Parsing Phase](../../go-parser-target/eros/ERO005_metrics-parsing-phase.md) — Parsing phase data structures (produces parser definitions)
- [ERO005 — Metrics Analysis Phase](../../go-analysis/eros/ERO005_metrics-analysis-phase.md) — Analysis phase data structures (produces CIL trees consumed by this synthesis phase)
- [ERO004 — Connectors Synthesis Phase](ERO004_connectors-synthesis-phase.md) — Connector synthesis (parquet reading reused by metrics)
- [Synthesis Design](../synthesis.md) — Full synthesis infrastructure design (CILAbstractTransformer, ComputationController, MetricComputation)
- [Tensor Representation](../tensor-representation.md) — Tensor data structures and three-level function hierarchy
- [IO Specifications Requirements](../../../../requirements/cpl/io_specs.md) — IO specification requirements
- [Metrics Syntax Requirements](../../../../requirements/cpl/metrics-syntax.md) — Syntax requirements for metrics
- [Type Selection for Numeric Calculations](../../type-selection.md) — Deferred design for fixed-point and overflow handling
- [Fixed-Point Arithmetic Pattern](../../../../patterns/fixed-point-arithmetic.md) — Pattern for fixed-point representation
- [CPL Architecture — Intermediate Language](../../../architecture/cpl/analysis/intermediate_language.md) — CIL specification

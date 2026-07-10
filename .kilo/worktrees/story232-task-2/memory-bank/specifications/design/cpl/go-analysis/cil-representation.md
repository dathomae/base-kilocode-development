## Chronocone Intermediate Language Representation

This document provides detailed design for the CIL representation. A higher level description of CIL is given in [Calculation/Operations Framework In Chronocone](../../architecture/cpl/analysis/intermediate_language.md).

The basic form of the CIL representation is a tree where each CIL node has:

- A CIL operation node that defines an operation and a list of one or more CIL nodes
- A constant value
- The name of a metric. For the purposes of this representation a connector defines a metric with fixed dimensions.
- A unit group name
- A measure group name
- A tensor

CIL operations may be used for calculations related to metrics (which all involve tensors) and calculations related to unit and measure transformations, which will always use unit and measure group names.

Representation of tensors is defined in [Tensor Representation In The Synthesizer Support Library](../go-synthesis-target/tensor-representation.md).

The upper level control for synthesized code is defined in [Synthesizing Go Code From Chronocone Intermediate Representation](../go-synthesis-target/synthesis.md).

---

## Conceptual Overview

### CIL Node Hierarchy

The Chronocone Intermediate Language (CIL) is represented as a tree of operations. The abstract node hierarchy provides a type-safe, language-agnostic representation of calculations before they are resolved to concrete implementations.

#### Abstract Node Types

A [`CILAbstractNode`](#cilabstractnode) represents a node in the CIL tree that has not yet been assigned an implementation type. The abstract node can be one of five variants:

1. **Data Constant** ([`CILDataConstantNode`](#cildataconstantnode)): A constant value with an associated CLP measure name. The value is stored as a string representation that will be converted to an actual typed value in the concrete node representation.

2. **Unit Name Constant** ([`CILUnitNameConstantNode`](#cilunitnameconstantnode)): A reference to a CLP unit name. Used in unit and measure transformation calculations.

3. **Measure Name Constant** ([`CILMeasureNameConstantNode`](#cilmeasurenameconstantnode)): A reference to a CLP measure name. Used in unit and measure transformation calculations.

4. **Operation Node** ([`CILAbstractOperationNode`](#cilabstractoperationnode)): An operation with child nodes. Operations are categorized as arithmetic or boolean operations that perform calculations on tensor data.

5. **Metric Reference** ([`CILMetricReferenceNode`](#cilmetricreferencenode)): A reference to a metric by name. Used to include metric data as inputs to calculations.

6. **Connector Reference** ([`CILConnectorReferenceNode`](#cilconnectorreferencenode)): A reference to a connector by name. Used to include external data as inputs to calculations.

#### Operation Node Types

An [`CILAbstractOperationNode`](#cilabstractoperationnode) is a specialized abstract node that represents an operation with child operands. Operations are defined as constants and include:

**Arithmetic Operations:**

- `Negate`: Negates a numeric value (unary)
- `Add`: Adds two values (binary)
- `Subtract`: Subtracts the second value from the first (binary)
- `Multiply`: Multiplies two values (binary)
- `Divide`: Divides the first value by the second (binary)

**Boolean Operations:**

- `Not`: Logical negation (unary)
- `And`: Logical conjunction (binary)
- `Or`: Logical disjunction (binary)
- `EqualTo`: Equality comparison (binary)
- `GreaterThan`: Greater than comparison (binary)
- `LessThan`: Less than comparison (binary)
- `GreaterThanOrEqual`: Greater than or equal comparison (binary)
- `LessThanOrEqual`: Less than or equal comparison (binary)

#### Node Relationships

The CIL abstract node hierarchy forms a tree structure where:

- Leaf nodes are constants (data or unit name) or metric/connector references
- Internal nodes are operations that reference child nodes
- Each operation node maintains an ordered list of children that are also [`CILAbstractNode`](#cilabstractnode) instances

#### Synthesis

The abstract CIL representation is transformed into executable Go code during the synthesis phase. See [Synthesizing Go Code From Chronocone Intermediate Representation](../go-synthesis-target/synthesis.md) for details on the code generation process.

---

## Go Implementation

### Node Type Constants

The `CILNodeType` enum identifies the type of an abstract node.

```go
// CILNodeType identifies the type of a CILAbstractNode
type CILNodeType int

const (
    CILNodeTypeDataConstant CILNodeType = iota
    CILNodeTypeUnitNameConstant
    CILNodeTypeMeasureNameConstant
    CILNodeTypeOperation
    CILNodeTypeMetricReference
    CILNodeTypeConnectorReference
)

// String returns the string representation of the node type
func (t CILNodeType) String() string
```

### Operation Type Constants

The `CILOperationType` enum identifies the specific operation performed by an operation node.

```go
// CILOperationType identifies the type of operation in a CILAbstractOperationNode
type CILOperationType int

const (
    // Arithmetic operations
    CILOperationTypeNegate CILOperationType = iota
    CILOperationTypeAdd
    CILOperationTypeSubtract
    CILOperationTypeMultiply
    CILOperationTypeDivide

    // Boolean operations
    CILOperationTypeNot
    CILOperationTypeAnd
    CILOperationTypeOr
    CILOperationTypeEqualTo
    CILOperationTypeGreaterThan
    CILOperationTypeLessThan
    CILOperationTypeGreaterThanOrEqual
    CILOperationTypeLessThanOrEqual

    // Mask formation operation
    CILOperationTypeMaskFromData
)

// String returns the string representation of the operation type
func (t CILOperationType) String() string

// IsArithmetic returns true if the operation is an arithmetic operation
func (t CILOperationType) IsArithmetic() bool

// IsBoolean returns true if the operation is a boolean operation
func (t CILOperationType) IsBoolean() bool

```

### CILAbstractNode Interface

`CILAbstractNode` is the root interface for all nodes in the CIL abstract syntax tree.

```go
// CILAbstractNode is the interface for all abstract nodes in the CIL tree.
// It provides the NodeType method to identify the specific type of node.
type CILAbstractNode interface {
    // NodeType returns the type of this node (data constant, unit name constant,
    // operation, or metric reference)
    NodeType() CILNodeType
}
```

### CILDataConstantNode

`CILDataConstantNode` represents a constant data value with an associated measure name. The value is stored as a string that will be parsed and converted to a typed value during code generation.

```go
// CILDataConstantNode represents a data constant with a measure name and string value.
// The value will be converted to an actual typed value during code generation.
type CILDataConstantNode struct {
    measureName string // CLP measure name for type information
    value       string // String representation of the constant value
}

// NodeType returns CILNodeTypeDataConstant
func (n *CILDataConstantNode) NodeType() CILNodeType

// MeasureName returns the CLP measure name associated with this constant
func (n *CILDataConstantNode) MeasureName() string

// Value returns the string representation of the constant value
func (n *CILDataConstantNode) Value() string

// NewCILDataConstantNode creates a new data constant node
// Parameters:
//   - measureName: The CLP measure name for type lookup
//   - value: The string representation of the constant value
// Returns:
//   - A pointer to the created CILDataConstantNode
func NewCILDataConstantNode(measureName string, value string) *CILDataConstantNode
```

### CILUnitNameConstantNode

`CILUnitNameConstantNode` represents a reference to a CLP unit name. Used in unit and measure transformation calculations.

```go
// CILUnitNameConstantNode represents a unit name constant.
// Used in unit and measure transformation calculations.
type CILUnitNameConstantNode struct {
    unitName string // CLP unit name
}

// NodeType returns CILNodeTypeUnitNameConstant
func (n *CILUnitNameConstantNode) NodeType() CILNodeType

// UnitName returns the CLP unit name
func (n *CILUnitNameConstantNode) UnitName() string

// NewCILUnitNameConstantNode creates a new unit name constant node
// Parameters:
//   - unitName: The CLP unit name
// Returns:
//   - A pointer to the created CILUnitNameConstantNode
func NewCILUnitNameConstantNode(unitName string) *CILUnitNameConstantNode
```

### CILMeasureNameConstantNode

`CILMeasureNameConstantNode` represents a reference to a CLP measure name. Used in unit and measure transformation calculations.

```go
// CILMeasureNameConstantNode represents a measure name constant.
// Used in unit and measure transformation calculations.
type CILMeasureNameConstantNode struct {
    measureName string // CLP measure name
}

// NodeType returns CILNodeTypeMeasureNameConstant
func (n *CILMeasureNameConstantNode) NodeType() CILNodeType

// MeasureName returns the CLP measure name
func (n *CILMeasureNameConstantNode) MeasureName() string

// NewCILMeasureNameConstantNode creates a new measure name constant node
// Parameters:
//   - measureName: The CLP measure name
// Returns:
//   - A pointer to the created CILMeasureNameConstantNode
func NewCILMeasureNameConstantNode(measureName string) *CILMeasureNameConstantNode
```

### CILMetricReferenceNode

`CILMetricReferenceNode` represents a reference to a metric by name. The metric name is used to look up the metric definition when needed.

```go
// CILMetricReferenceNode represents a reference to a metric.
// The metric name is used to look up the metric definition.
type CILMetricReferenceNode struct {
    metricName string // Name of the referenced metric
}

// NodeType returns CILNodeTypeMetricReference
func (n *CILMetricReferenceNode) NodeType() CILNodeType

// MetricName returns the name of the referenced metric
func (n *CILMetricReferenceNode) MetricName() string

// NewCILMetricReferenceNode creates a new metric reference node
// Parameters:
//   - metricName: The name of the metric to reference
// Returns:
//   - A pointer to the created CILMetricReferenceNode
func NewCILMetricReferenceNode(metricName string) *CILMetricReferenceNode
```

### CILConnectorReferenceNode

`CILConnectorReferenceNode` represents a reference to a connector by name. The connector name is used to look up the connector definition and its associated IO specification when needed.

```go
// CILConnectorReferenceNode represents a reference to a connector.
// The connector name is used to look up the connector definition.
type CILConnectorReferenceNode struct {
    connectorName string // Name of the referenced connector
}

// NodeType returns CILNodeTypeConnectorReference
func (n *CILConnectorReferenceNode) NodeType() CILNodeType

// ConnectorName returns the name of the referenced connector
func (n *CILConnectorReferenceNode) ConnectorName() string

// NewCILConnectorReferenceNode creates a new connector reference node
// Parameters:
//   - connectorName: The name of the connector to reference
// Returns:
//   - A pointer to the created CILConnectorReferenceNode
func NewCILConnectorReferenceNode(connectorName string) *CILConnectorReferenceNode
```

### CILAbstractOperationNode

`CILAbstractOperationNode` represents an operation with child nodes. The operation type determines the operation to be performed, and the children provide the operands.

```go
// CILAbstractOperationNode represents an operation in the CIL tree.
// All operations use this single struct type; the operation type determines
// the behavior via logic/switch statements during synthesis and evaluation.
type CILAbstractOperationNode struct {
    operationType CILOperationType  // The specific operation type
    children      []CILAbstractNode // Child nodes (operands)
}

// NodeType returns CILNodeTypeOperation
func (n *CILAbstractOperationNode) NodeType() CILNodeType

// OperationType returns the specific operation type
func (n *CILAbstractOperationNode) OperationType() CILOperationType

// Children returns the slice of child nodes
func (n *CILAbstractOperationNode) Children() []CILAbstractNode

// ChildCount returns the number of children
func (n *CILAbstractOperationNode) ChildCount() int

// ChildAt returns the child at the specified index.
// Returns an error if the index is out of bounds.
func (n *CILAbstractOperationNode) ChildAt(index int) (CILAbstractNode, error)

// NewCILAbstractOperationNode creates a new operation node with the given operation type and operands.
// The number of operands must match the operation's arity:
//   - Unary operations: Negate, Not (1 operand)
//   - Binary operations: Add, Subtract, Multiply, Divide, And, Or,
//     EqualTo, GreaterThan, LessThan, GreaterThanOrEqual, LessThanOrEqual (2 operands)
// Parameters:
//   - operationType: The type of operation to perform
//   - operands: The child nodes (operands) for the operation
// Returns:
//   - A pointer to the created CILAbstractOperationNode
//   - An error if the operand count is incorrect or any operand is nil
func NewCILAbstractOperationNode(operationType CILOperationType, operands ...CILAbstractNode) (*CILAbstractOperationNode, error)
```

### Operation Arity

Each operation type expects a specific number of operands:

| Operation          | Type       | Arity  | Description                                       |
| ------------------ | ---------- | ------ | ------------------------------------------------- |
| Negate             | Arithmetic | Unary  | Negates a numeric value (-x)                      |
| Add                | Arithmetic | Binary | Adds two values (x + y)                           |
| Subtract           | Arithmetic | Binary | Subtracts the second value from the first (x - y) |
| Multiply           | Arithmetic | Binary | Multiplies two values (x \* y)                    |
| Divide             | Arithmetic | Binary | Divides the first value by the second (x / y)     |
| Not                | Boolean    | Unary  | Logical negation (NOT x)                          |
| And                | Boolean    | Binary | Logical conjunction (x AND y)                     |
| Or                 | Boolean    | Binary | Logical disjunction (x OR y)                      |
| EqualTo            | Boolean    | Binary | Equality comparison (x = y)                       |
| GreaterThan        | Boolean    | Binary | Greater than comparison (x > y)                   |
| LessThan           | Boolean    | Binary | Less than comparison (x < y)                      |
| GreaterThanOrEqual | Boolean    | Binary | Greater than or equal comparison (x >= y)         |
| LessThanOrEqual    | Boolean    | Binary | Less than or equal comparison (x <= y)            |
| MaskFromData       | Boolean    | Unary  | Creates mask tensor from data tensor              |

### Binary Operation Operand Types

Binary arithmetic operations (`Add`, `Subtract`, `Multiply`, `Divide`) in CIL are defined abstractly and do not distinguish between operand types at the representation level. The operands can be:

1. **Constant-Tensor**: One operand is a [`CILDataConstantNode`](memory-bank/specifications/design/cpl/go-analysis/cil-representation.md:163) and the other is a tensor-producing node (either a [`CILMetricReferenceNode`](memory-bank/specifications/design/cpl/go-analysis/cil-representation.md:243) or another operation)

2. **Tensor-Tensor**: Both operands are tensor-producing nodes

During synthesis, the [`CILAbstractTransformer`](memory-bank/specifications/design/cpl/go-synthesis-target/synthesis.md:383) examines the operand types and selects the appropriate implementation:

- For constant-tensor operations: Uses scalar-tensor operation functions (e.g., `addOp`)
- For tensor-tensor operations: Uses tensor-tensor operation functions (e.g., `addTensorsOp`)

See [Synthesizing Go Code From Chronocone Intermediate Representation](../go-synthesis-target/synthesis.md#operation-code-generation) for details on the generated code patterns.

### CIL to Chunk Execution Mapping

During synthesis, each [`CILAbstractOperationNode`](#cilabstractoperationnode) is transformed into Go code that performs the operation using the three-level function hierarchy defined in the synthesis target. The transformation follows this pattern:

```
CILAbstractOperationNode
    ├── operationType: CILOperationType
    └── children: []CILAbstractNode
         ↓ Synthesis Transform
Generated Go Code (Level 1 - High-Level API)
    └── tensor.Operation(controller, operand)
         ↓ Internally calls (Level 2 - Orchestration)
    apply_binary_to_tensor(opFunc, chunkingStrategy, ...)
         ↓ Creates and submits
ChronoconeTensorComputationChunk (Level 3 - Worker Execution)
    ├── operands: []*ChronoconeTensorComputationChunkOperand
    ├── matrixOp OR pairwiseOp OR unaryOp (operation function)
    └── output: *ChronoconeTensorComputationChunkOperand
         ↓ Submit to Work Channel
Worker Goroutine Execution
    └── chunk.matrixOp(chunk.operands, chunk.output)
```

**Three-Level Hierarchy:**

1. **Level 1 (Generated Code)**: The synthesized code calls high-level tensor methods like `tensor.SubtractScalar()` or `tensor.Add()`. These methods provide a clean API that matches CIL operations.

2. **Level 2 (Orchestration)**: The high-level methods internally call `apply_binary_to_tensor()` (or similar primitives) with the appropriate operation function (e.g., `subtractOp`), chunking strategy, and operands. This level handles dividing tensors into chunks and submitting work.

3. **Level 3 (Worker Execution)**: Worker goroutines receive `ChronoconeTensorComputationChunk` objects and execute the operation functions (e.g., `subtractOp`) on `ChronoconeTensorComputationChunkOperand` objects.

See [Three-Level Function Hierarchy](../go-synthesis-target/tensor-representation.md#three-level-function-hierarchy) for complete details on this design, [CIL to Synthesis Relationship](../go-synthesis-target/synthesis.md#cil-to-synthesis-relationship) for the transformation specification, and [ChronoconeComputationController](../go-synthesis-target/synthesis.md#chronoconecomputationcontroller) for the work dispatch mechanism.

---

## Type Relationships

The following diagram shows the relationships between the types defined in this document:

### CILAbstractNode Hierarchy

```
CILAbstractNode (interface)
    ├── CILDataConstantNode
    ├── CILUnitNameConstantNode
    ├── CILMeasureNameConstantNode
    ├── CILMetricReferenceNode
    ├── CILConnectorReferenceNode
    └── CILAbstractOperationNode
```

The [`CILAbstractOperationNode`](#cilabstractoperationnode) struct implements all operations using a logic/switch-driven approach. The [`CILOperationType`](#operation-type-constants) enum determines which operation is performed, and the operands are stored as children in the node.

During synthesis, the abstract CIL tree is transformed into executable Go code. The synthesis process:

1. Walks the CIL tree using the [`CILAbstractTransformer`](../go-synthesis-target/synthesis.md#cilabstracttransformer)
2. Generates code that creates [`ChronoconeTensorComputationChunk`](../go-synthesis-target/tensor-representation.md#chronoconetensorcomputationchunk) objects
3. Each chunk encapsulates the operation function and its operands
4. Generated code submits chunks to the [`ChronoconeComputationController.workChannel`](../go-synthesis-target/synthesis.md#chronoconecomputationcontroller)
5. Worker goroutines execute chunks in parallel, with the number of workers equal to the IO specification's core count

See [CIL to Synthesis Relationship](../go-synthesis-target/synthesis.md#cil-to-synthesis-relationship) for the detailed transformation mapping, and [Synthesizing Go Code From Chronocone Intermediate Representation](../go-synthesis-target/synthesis.md) for the complete code generation specification.

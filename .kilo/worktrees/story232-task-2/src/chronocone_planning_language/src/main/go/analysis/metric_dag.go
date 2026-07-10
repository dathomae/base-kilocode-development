package analysis

import (
	"container/heap"
	"container/list"
	"fmt"
	"sort"
	"strings"
)

// stringHeap implements heap.Interface for deterministic string ordering.
type stringHeap []string

func (h stringHeap) Len() int            { return len(h) }
func (h stringHeap) Less(i, j int) bool  { return h[i] < h[j] }
func (h stringHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *stringHeap) Push(x interface{}) { *h = append(*h, x.(string)) }
func (h *stringHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// MetricDAG represents the dependency graph between metrics.
//
// The DAG is built as a second pass after all AnalysisMetricDefinition entries
// are assembled. It captures which metrics depend on which other metrics via
// metric-type input references. RootNodes are metrics with no dependencies on
// other metrics. Order is populated by topological sort (Task 2).
type MetricDAG struct {
	// Nodes is a lookup map from metric name to its DAG node.
	Nodes map[string]*DAGNode

	// RootNodes is the list of metric names that have no dependencies on other metrics.
	RootNodes []string

	// Order is the topological evaluation order of metrics.
	// Populated by topological sort in Task 2.
	Order []string
}

// DAGNode represents a single metric in the dependency DAG.
type DAGNode struct {
	// MetricName is the name of the metric this node represents.
	MetricName string

	// Dependencies is the list of metric names this metric depends on.
	Dependencies []string

	// Dependents is the list of metric names that depend on this metric.
	Dependents []string

	// Visited is used for DFS traversal.
	Visited bool

	// InProgress is used for cycle detection (DFS gray-set).
	InProgress bool
}

// TopologicalSort computes a topological ordering of the DAG using Kahn's algorithm.
//
// The algorithm processes nodes with zero in-degree first, then decrements the
// in-degree of each dependent. If the result length does not equal the number of
// nodes, a cycle is detected and an error is returned. The result is stored in
// dag.Order.
//
// Returns:
//   - A slice of metric names in topological order (dependencies before dependents).
//   - An error if the DAG contains a cycle.
func (dag *MetricDAG) TopologicalSort() ([]string, error) {
	if len(dag.Nodes) == 0 {
		dag.Order = []string{}
		return dag.Order, nil
	}

	// Step 1: Compute in-degree counts for all nodes.
	// In-degree is the number of dependencies a node has.
	inDegree := make(map[string]int, len(dag.Nodes))
	for name, node := range dag.Nodes {
		inDegree[name] = len(node.Dependencies)
	}

	// Step 2: Initialize a queue with all nodes that have in-degree 0.
	// Use a min-heap for deterministic ordering.
	var queue stringHeap
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	heap.Init(&queue)

	// Step 3: Process the queue.
	var result []string
	for queue.Len() > 0 {
		current := heap.Pop(&queue).(string)
		result = append(result, current)

		// Decrement in-degree of all dependents.
		dependents := make([]string, len(dag.Nodes[current].Dependents))
		copy(dependents, dag.Nodes[current].Dependents)
		sort.Strings(dependents)

		for _, dep := range dependents {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				heap.Push(&queue, dep)
			}
		}
	}

	// Step 4: If result length != number of nodes, a cycle exists.
	if len(result) != len(dag.Nodes) {
		dag.Order = nil
		return nil, fmt.Errorf("cycle detected in metric dependency DAG")
	}

	dag.Order = result
	return result, nil
}

// DetectCycles performs DFS-based cycle detection with three-color marking.
//
// The algorithm resets all Visited and InProgress flags, then performs a DFS
// from each unvisited node. If a dependency is already InProgress, a cycle is
// detected. The cycle path is collected and checked against isPermittedSelfReference
// to allow time-shifted self-references.
//
// Parameters:
//   - metrics: The metrics map used to look up time-shift information for self-reference checking.
//
// Returns:
//   - nil if no cycles (or only permitted time-shifted self-references) are found.
//   - An error describing the cycle path if a non-permitted cycle is detected.
func (dag *MetricDAG) DetectCycles(metrics map[string]AnalysisMetricDefinition) error {
	// Step 1: Reset all Visited and InProgress flags.
	for _, node := range dag.Nodes {
		node.Visited = false
		node.InProgress = false
	}

	// Step 2: For each unvisited node, do DFS.
	var path []string

	var dfs func(nodeName string) error
	dfs = func(nodeName string) error {
		node := dag.Nodes[nodeName]
		node.InProgress = true
		path = append(path, nodeName)

		for _, dep := range node.Dependencies {
			depNode := dag.Nodes[dep]
			if depNode.InProgress {
				// Found a cycle — collect the cycle path.
				cycleStart := 0
				for i, name := range path {
					if name == dep {
						cycleStart = i
						break
					}
				}
				cyclePath := append(path[cycleStart:], dep)
				if !isPermittedSelfReference(cyclePath, metrics) {
					cycleStr := strings.Join(cyclePath, " → ")
					return fmt.Errorf("circular dependency: %s", cycleStr)
				}
				// Permitted self-reference — continue checking other dependencies.
				continue
			}
			if !depNode.Visited {
				if err := dfs(dep); err != nil {
					return err
				}
			}
		}

		path = path[:len(path)-1]
		node.InProgress = false
		node.Visited = true
		return nil
	}

	// Sort node names for deterministic DFS traversal.
	nodeNames := make([]string, 0, len(dag.Nodes))
	for name := range dag.Nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	for _, name := range nodeNames {
		if !dag.Nodes[name].Visited {
			if err := dfs(name); err != nil {
				return err
			}
		}
	}

	return nil
}

// MarkReachable performs a BFS traversal starting from each output metric name,
// following Dependencies edges to identify all metrics reachable from the outputs.
//
// The traversal follows the reverse direction of evaluation: starting from output
// metrics and traversing downward through their dependency chains to find all
// ancestor metrics needed for computation. Metrics not in the reachable set are
// dead code and can be eliminated.
//
// Parameters:
//   - outputMetricNames: The names of output metrics to start traversal from.
//     Unknown metric names are silently skipped.
//
// Returns:
//   - A map from metric name to true for all reachable metrics.
func (dag *MetricDAG) MarkReachable(outputMetricNames []string) map[string]bool {
	reachable := make(map[string]bool)

	// BFS queue — start from each output metric name that exists in the DAG.
	// Use container/list for O(1) dequeue when processing large DAGs.
	queue := list.New()
	for _, name := range outputMetricNames {
		if _, exists := dag.Nodes[name]; exists {
			reachable[name] = true
			queue.PushBack(name)
		}
	}

	// BFS traversal following Dependencies edges.
	// Dependencies point from a metric to the metrics it depends on.
	// Starting from an output, we follow Dependencies to find all required ancestors.
	for queue.Len() > 0 {
		front := queue.Front()
		queue.Remove(front)
		current := front.Value.(string)

		node := dag.Nodes[current]
		for _, dep := range node.Dependencies {
			if !reachable[dep] {
				reachable[dep] = true
				queue.PushBack(dep)
			}
		}
	}

	return reachable
}

// isPermittedSelfReference checks whether a detected cycle is a permitted
// time-shifted self-reference.
//
// A self-reference is permitted when the cycle consists of exactly one unique
// node (a self-loop), that metric has an input with TimeShift != nil whose
// Name matches the metric name, and there is no non-time-shifted self-reference
// for the same metric (which would indicate a real cycle).
//
// Parameters:
//   - cyclicDependencies: The cycle path detected by DFS (e.g., ["A", "A"]).
//   - metrics: The metrics map for looking up input definitions.
//
// Returns:
//   - true if the cycle is a permitted time-shifted self-reference with no
//     conflicting non-time-shifted self-reference.
//   - false otherwise.
func isPermittedSelfReference(cyclicDependencies []string, metrics map[string]AnalysisMetricDefinition) bool {
	// A self-loop has exactly 2 elements: ["A", "A"].
	if len(cyclicDependencies) != 2 {
		return false
	}
	if cyclicDependencies[0] != cyclicDependencies[1] {
		return false
	}

	metricName := cyclicDependencies[0]
	metric, exists := metrics[metricName]
	if !exists {
		return false
	}

	hasTimeShiftedSelfRef := false
	hasNonTimeShiftedSelfRef := false

	for _, input := range metric.Inputs {
		if input.Kind == "metric" && input.Name == metricName {
			if input.TimeShift != nil {
				hasTimeShiftedSelfRef = true
			} else {
				hasNonTimeShiftedSelfRef = true
			}
		}
	}

	// Only permit if there IS a time-shifted self-reference AND there is NO
	// non-time-shifted self-reference (which would indicate a real cycle).
	return hasTimeShiftedSelfRef && !hasNonTimeShiftedSelfRef
}

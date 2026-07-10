// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"
	"strconv"
	"strings"
)

// GetCollapsedDimensionName returns the name of the dimension that is
// removed during aggregation: the first dimension in fromDimensions that
// is NOT equal to toDimension. If all dimensions match toDimension (or
// fromDimensions is empty), it returns an empty string.
func GetCollapsedDimensionName(fromDimensions []string, toDimension string) string {
	for _, dim := range fromDimensions {
		if dim != toDimension {
			return dim
		}
	}
	return ""
}

// TransformMetric walks the CIL tree for a metric and generates a Go
// computeX() function. It converts the CIL expression tree rooted at
// the metric's CILNode into a postfix operation table via the postfix
// walker, then generates the corresponding compute function body that
// calls calculation.EvaluatePostfixFloat64.
//
// All CIL node types (CILLiteral, CILBinaryOp, CILUnaryOp, CILConditional,
// CILWhenClause, CILInputRef) are routed through the unified postfix
// walker path, except for bare CILLiterals with dimension references
// which use AllocateTensor plus a fill-loop to preserve shape.
// The legacy visitor-snippet-stack approach has been removed.
//
// Special case: When the CIL node is a bare CILLiteral and InputRefs
// contain dimension references, the generated code allocates a tensor
// with the correct dimension shape and fills it with the literal value.
// This is necessary because EvaluatePostfixFloat64 produces a scalar
// tensor for a single literal push, which would lose the dimension
// information needed by downstream output writing.
//
// Parameters:
//   - metricName: The name of the metric.
//   - root: The root CIL node for the metric (expected to be *CILMetricNode).
//
// Returns:
//   - string: The generated computeX() function code.
//   - error: An error if root is not a *CILMetricNode, if the CILNode is
//     nil, or if postfix table generation fails.
func (t *CILAbstractTransformer) TransformMetric(metricName string, root interface{}) (string, error) {
	if t == nil {
		return "", fmt.Errorf("transformer is nil")
	}

	// Type-assert root to *CILMetricNode
	metricNode, ok := root.(*CILMetricNode)
	if !ok {
		return "", fmt.Errorf("root is not a *CILMetricNode, got %T", root)
	}
	if metricNode == nil {
		return "", fmt.Errorf("metric node is nil")
	}

	if metricNode.CILNode == nil {
		return "", fmt.Errorf("metric node CILNode is nil")
	}

	// Scan InputRefs to determine tensor dimensions and whether the inputs
	// map is referenced. InputRefs of kind "metric" or "connector" mean
	// the generated code will reference the inputs parameter; otherwise the
	// parameter is unused and needs a suppression assignment.
	var dimNames []string
	usesInputs := false
	for _, ref := range metricNode.InputRefs {
		switch ref.Kind {
		case "dimension":
			dimNames = append(dimNames, ref.Name)
		case "metric", "connector":
			usesInputs = true
		}
	}

	// Generate the computeX() function wrapper
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("func compute%s(controller *calculation.ChronoconeComputationController, inputs map[string]*calculation.ChronoconeTensor[any]) {\n", metricName))
	if !usesInputs {
		sb.WriteString("\t_ = inputs\n")
	}

	// Special case: bare literal with dimension refs — allocate a properly
	// shaped tensor and fill it with the literal value. EvaluatePostfixFloat64
	// produces a scalar for single-literals, which would lose dimension
	// information needed by downstream output writing.
	if lit, ok := metricNode.CILNode.(*CILLiteral); ok {
		if lit == nil {
			return "", fmt.Errorf("literal node is nil")
		}
		if len(dimNames) > 0 {
			sb.WriteString(fmt.Sprintf("\tresult, err := calculation.AllocateTensor(controller"))
			for _, dim := range dimNames {
				sb.WriteString(fmt.Sprintf(", %q", dim))
			}
			sb.WriteString(")\n")
			sb.WriteString(fmt.Sprintf("\tcheck(err, \"Failed to allocate tensor for %s\")\n", metricName))
			sb.WriteString("\tdata := result.Data()\n")
			sb.WriteString("\tfor i := range data {\n")
			sb.WriteString(fmt.Sprintf("\t\tdata[i] = float64(%s)\n", strconv.FormatFloat(lit.Value, 'f', -1, 64)))
			sb.WriteString("\t}\n")
			sb.WriteString(fmt.Sprintf("\tif err := controller.GetMetricComputation(%q).SetTensor(result); err != nil {\n", metricName))
			sb.WriteString(fmt.Sprintf("\t\tcheck(err, \"Failed to set tensor for %s\")\n", metricName))
			sb.WriteString("\t}\n")
		} else {
			// Literal without dimension refs — use postfix path.
			postfixBody, err := t.generatePostfixComputeBody(metricName, metricNode)
			if err != nil {
				return "", fmt.Errorf("failed to generate postfix compute body: %w", err)
			}
			sb.WriteString(postfixBody)
		}
	} else if agg, ok := metricNode.CILNode.(*CILAggregateNode); ok {
		if agg == nil {
			return "", fmt.Errorf("aggregate node is nil")
		}

		// Compute collapsedDim: the dimension in FromDimensions that is NOT in ToDimension.
		collapsedDimName := GetCollapsedDimensionName(agg.FromDimensions, agg.ToDimension)
		collapsedDim := -1
		for i, dim := range agg.FromDimensions {
			if dim != agg.ToDimension {
				collapsedDim = i
				break
			}
		}
		if collapsedDim == -1 {
			return "", fmt.Errorf("could not determine collapsed dimension from FromDimensions %v and ToDimension %q", agg.FromDimensions, agg.ToDimension)
		}

		// Compute targetDim: index of ToDimension in FromDimensions, or -1 if empty (scalar aggregation).
		targetDim := -1
		if agg.ToDimension != "" {
			for i, dim := range agg.FromDimensions {
				if dim == agg.ToDimension {
					targetDim = i
					break
				}
			}
		}

		// Walk the SourceCIL subtree to postfix.
		if agg.SourceCIL == nil {
			return "", fmt.Errorf("aggregate node SourceCIL is nil")
		}
		ops, err := WalkCILTreeToPostfix(agg.SourceCIL)
		if err != nil {
			return "", fmt.Errorf("failed to walk aggregate source CIL to postfix: %w", err)
		}

		opsLiteral := t.GeneratePostfixTableLiteral(ops)
		matrixVar := fmt.Sprintf("aggregationMatrix_%s", collapsedDimName)

		var aggSb strings.Builder
		aggSb.WriteString(fmt.Sprintf("\tops := %s\n", opsLiteral))
		aggSb.WriteString("\tsourceResult, err := calculation.EvaluatePostfixFloat64(controller, ops)\n")
		aggSb.WriteString("\tcheck(err, \"Failed to evaluate source for aggregation\")\n")
		// Convert *ChronoconeTensor[any] to *ChronoconeTensor[float64] for tensorops.ReduceDimension.
		aggSb.WriteString("\tsourceData := sourceResult.Data()\n")
		aggSb.WriteString("\tsourceFloat64 := make([]float64, len(sourceData))\n")
		aggSb.WriteString("\tfor i, v := range sourceData {\n")
		aggSb.WriteString("\t\tsourceFloat64[i] = v.(float64)\n")
		aggSb.WriteString("\t}\n")
		aggSb.WriteString("\tsourceTensor := calculation.NewChronoconeTensor(sourceFloat64, sourceResult.DimensionNames(), sourceResult.Shape())\n")
		aggSb.WriteString(fmt.Sprintf("\tresult, err := tensorops.ReduceDimension(sourceTensor, %s, %d, %d)\n", matrixVar, collapsedDim, targetDim))
		aggSb.WriteString("\tif err != nil {\n")
		aggSb.WriteString("\t\tcheck(fmt.Errorf(\"aggregation failed: %w\", err), \"ReduceDimension failed\")\n")
		aggSb.WriteString("\t}\n")
		// Convert result from float64 back to any for SetTensor
		aggSb.WriteString("\tresultData := result.Data()\n")
		aggSb.WriteString("\tresultAny := make([]any, len(resultData))\n")
		aggSb.WriteString("\tfor i, v := range resultData {\n")
		aggSb.WriteString("\t\tresultAny[i] = v\n")
		aggSb.WriteString("\t}\n")
		aggSb.WriteString("\tanyResult := calculation.NewChronoconeTensor(resultAny, result.DimensionNames(), result.Shape())\n")
		aggSb.WriteString(fmt.Sprintf("\tif err := controller.GetMetricComputation(%q).SetTensor(anyResult); err != nil {\n", metricName))
		aggSb.WriteString(fmt.Sprintf("\t\tcheck(err, \"Failed to set tensor for %s\")\n", metricName))
		aggSb.WriteString("\t}\n")

		sb.WriteString(aggSb.String())
	} else {
		// All other node types go through the unified postfix compute body path.
		postfixBody, err := t.generatePostfixComputeBody(metricName, metricNode)
		if err != nil {
			return "", fmt.Errorf("failed to generate postfix compute body: %w", err)
		}
		sb.WriteString(postfixBody)
	}

	sb.WriteString("}\n")

	return sb.String(), nil
}

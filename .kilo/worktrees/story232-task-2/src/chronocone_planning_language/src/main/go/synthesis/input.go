// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// SynthesisInput is the stable contract between the analysis and synthesis phases.
// It carries all data extracted from AnalysisDefinitions that synthesis needs to
// generate code, decoupling the two phases so that internal changes in analysis
// do not propagate into synthesis.
type SynthesisInput struct {
	// Connectors contains all analyzed connector definitions.
	Connectors map[string]analysis.AnalysisConnectorDefinition

	// Measures contains all analyzed measure definitions.
	Measures map[string]analysis.AnalysisMeasureDefinition

	// Metrics contains all analyzed metric definitions.
	Metrics map[string]analysis.AnalysisMetricDefinition

	// Concepts contains all analyzed concept definitions.
	Concepts map[string]analysis.AnalysisConceptDefinition

	// Types contains all analyzed type definitions.
	Types map[string]analysis.AnalysisTypeDefinition

	// IOSpec is the IO specification defining inputs, outputs, and architecture.
	IOSpec *cio.IOSpecification

	// CILTrees maps connector/metric names to their CIL tree root nodes.
	CILTrees map[string]cil.CILNode

	// Lookups contains all analyzed lookup definitions from the analysis phase.
	// This map is populated so that synthesis can reference lookup metadata
	// when generating code for metric expressions that use lookups.
	Lookups map[string]*analysis.LookupInfo

	// MeasureUnitExpressions maps measure names to their synthesized Go unit
	// expression strings. Populated from AnalysisMeasureDefinition.UnitExpression
	// via ExpressionToCILNode and SynthesizeExpression.
	MeasureUnitExpressions map[string]string

	// UnitConversionCode contains the generated Go code for unit conversions.
	// Populated from AnalysisUnitDefinition.Conversions via CILConversionPair
	// and SynthesizeConversions.
	UnitConversionCode string

	// UnitConversionPairs is the retained list of unit conversion pairs
	// (before rendering to Go code strings). Needed by the unit checker
	// for validation lookups during synthesis.
	UnitConversionPairs []cil.CILConversionPair

	// MeasureConversionCode contains the generated Go code for measure conversions.
	// Populated from AnalysisMeasureDefinition.Conversions via
	// CILConversionPair and SynthesizeConversions.
	MeasureConversionCode string

	// MeasureConversionPairs is the retained list of measure conversion pairs.
	MeasureConversionPairs []cil.CILConversionPair
	// Transformer is the CIL abstract transformer populated from connector
	// and measure registrations.
	Transformer *cil.CILAbstractTransformer

	// DAG is the metric dependency DAG built during the analysis phase.
	// It provides topological evaluation order and dependency/dependent
	// relationships for correct code generation.
	DAG *analysis.MetricDAG
}

package analysis

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// AnalyzeConversionsForTest exports the unexported analyzeConversions function
// so that external tests in package analysis_test can verify its behavior.
func AnalyzeConversionsForTest[Def any](
	table *symboltable.SymbolTable,
	targetScopeType symboltable.ScopeType,
	lookup map[string]*Def,
	defName string,
	def *Def,
	units map[string]AnalysisUnitDefinition,
	entityLabel string,
	validatePair func(targetDef, sourceDef *Def) error,
) error {
	return analyzeConversions(table, targetScopeType, lookup, defName, def, units, entityLabel, validatePair)
}

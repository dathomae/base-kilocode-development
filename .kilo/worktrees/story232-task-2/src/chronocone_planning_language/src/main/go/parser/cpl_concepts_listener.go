// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLConcepts.g4 grammar file,
// which defines the concepts section and concept definitions.
//
// Grammar rules covered:
//   - conceptsSection, conceptDefinition, valuesPatternGenerator
//
// Note: The identifier rule is defined in ChronoconePlanningLanguage.g4 and its
// listener methods are in chronocone_planning_language_listener.go
package parser

import (
	"fmt"
	"strconv"

	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterConceptsSection is called when entering the concepts section.
// It pushes a new ConceptsScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterConceptsSection(ctx *antlr.ConceptsSectionContext) {
	l.logEnter("ConceptsSection", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the concepts scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering concepts section: %w", err))
		return
	}

	conceptsTable := current.AddChild(symboltable.ConceptsScope)
	l.symbolStack.Push(conceptsTable)
}

// ExitConceptsSection is called when exiting the concepts section.
// It pops the concepts scope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitConceptsSection(ctx *antlr.ConceptsSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting concepts section: %w", err))
	}
	l.logExit("ConceptsSection", &ctx.BaseParserRuleContext)
}

// EnterConceptDefinition is called when entering a concept definition.
// It pushes a new ConceptScope onto the symbol stack.
//
// Example concept definition: Country=[Values[USA, Canada, Mexico]]
// This corresponds to the 'conceptDefinition' rule in CPLConcepts.g4.
func (l *CPLListener) EnterConceptDefinition(ctx *antlr.ConceptDefinitionContext) {
	l.logEnter("ConceptDefinition", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the concept scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering concept definition: %w", err))
		return
	}

	// Create a child scope for this concept definition and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.ConceptScope))
}

// ExitConceptDefinition is called when exiting a concept definition.
// It pops the concept scope from the symbol stack.
func (l *CPLListener) ExitConceptDefinition(ctx *antlr.ConceptDefinitionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting concept definition: %w", err))
	}
	l.logExit("ConceptDefinition", &ctx.BaseParserRuleContext)
}

// EnterConditionalSpec is called when entering a conditional concept specification.
// It pushes a new ConditionalSpecScope onto the symbol stack for collecting conditional data.
//
// Example: Country[USA] => State=[Values[NY, CA, TX]]
// This corresponds to the 'conditionalSpec' rule in CPLConcepts.g4.
func (l *CPLListener) EnterConditionalSpec(ctx *antlr.ConditionalSpecContext) {
	l.logEnter("ConditionalSpec", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the conditional spec scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering conditional spec: %w", err))
		return
	}

	// Create a child scope for this conditional spec and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.ConditionalSpecScope))
}

// ExitConditionalSpec is called when exiting a conditional concept specification.
// It extracts the parent concept name and parent value, finds the child ConceptLine
// scope to read the generated values, determines the generator type and range bounds
// from the child generator scopes, builds a ConditionalSpec, and appends it to the
// child ConceptScope's ConceptDefinitionData.ConditionalSpecs slice so the analysis
// walker (which visits ConceptScope children of ConditionalSpecScope) can locate it.
func (l *CPLListener) ExitConditionalSpec(ctx *antlr.ConditionalSpecContext) {
	scope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting conditional spec: %w", err))
		l.logExit("ConditionalSpec", &ctx.BaseParserRuleContext)
		return
	}

	// Extract parent concept name from identifier
	identifierCtx := ctx.Identifier()
	if identifierCtx == nil {
		l.AddError(&SemanticError{
			Message: "missing parent concept name in conditional spec",
			Rule:    "conditionalSpec",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConditionalSpec", &ctx.BaseParserRuleContext)
		return
	}
	// Validation: parent concept name must be non-empty
	if parentConcept := identifierCtx.GetText(); parentConcept == "" {
		l.AddError(&SemanticError{
			Message: "conditional spec requires a non-empty parent concept name",
			Rule:    "conditionalSpec",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConditionalSpec", &ctx.BaseParserRuleContext)
		return
	}

	// Extract parent value from Value() context
	valueCtx := ctx.Value()
	if valueCtx == nil {
		l.AddError(&SemanticError{
			Message: "missing parent value in conditional spec",
			Rule:    "conditionalSpec",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConditionalSpec", &ctx.BaseParserRuleContext)
		return
	}
	parentValue := valueCtx.GetText()

	// Validation: parent value must be non-empty
	if parentValue == "" {
		l.AddError(&SemanticError{
			Message: "conditional spec requires a non-empty parent value",
			Rule:    "conditionalSpec",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConditionalSpec", &ctx.BaseParserRuleContext)
		return
	}

	// Find the child ConceptLine scope and read its ConceptDefinitionData.
	// The grammar (CPLConcepts.g4) guarantees exactly one conceptLine per
	// conditionalSpec, so taking the first matching ConceptScope child is
	// sufficient — there will never be multiple ConceptScopes here.
	var childConceptData *ast.ConceptDefinitionData
	var conceptLineScope *symboltable.SymbolTable
	for _, child := range scope.Children {
		if child.ScopeType == symboltable.ConceptScope && child.Data != nil {
			if data, ok := child.Data.(*ast.ConceptDefinitionData); ok {
				childConceptData = data
				conceptLineScope = child
				break
			}
		}
	}

	if childConceptData == nil {
		l.AddError(&SemanticError{
			Message: "conditional spec missing child concept line data",
			Rule:    "conditionalSpec",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConditionalSpec", &ctx.BaseParserRuleContext)
		return
	}

	// Determine generator type and range bounds from concept line's child generators
	var generatorType ast.GeneratorType
	var rangeBounds *ast.RangeBounds
	var sourceConceptName string

	for _, child := range conceptLineScope.Children {
		switch child.ScopeType {
		case symboltable.ValueGeneratorScope:
			if child.Data != nil {
				if _, ok := child.Data.(*ast.ValuesGeneratorDef); ok {
					generatorType = ast.ValuesGenerator
				}
			}
		case symboltable.IntGeneratorScope:
			if child.Data != nil {
				if genDef, ok := child.Data.(*ast.IntRangeGeneratorDef); ok {
					generatorType = ast.IntRangeGenerator
					rangeBounds = &ast.RangeBounds{
						Start: strconv.Itoa(genDef.Start),
						End:   strconv.Itoa(genDef.End),
						Step:  genDef.Step,
					}
				}
			}
		case symboltable.ConceptGeneratorScope:
			if child.Data != nil {
				if genDef, ok := child.Data.(*ast.ConceptRangeGeneratorDef); ok {
					generatorType = ast.ConceptRangeGenerator
					rangeBounds = &ast.RangeBounds{
						Start: genDef.StartValue,
						End:   genDef.EndValue,
					}
					sourceConceptName = genDef.ConceptName
				}
			}
		}
	}

	// Build ConditionalSpec
	conditionalSpec := &ast.ConditionalSpec{
		ParentValue:       parentValue,
		GeneratorType:     generatorType,
		Values:            childConceptData.DefaultValues,
		RangeBounds:       rangeBounds,
		SourceConceptName: sourceConceptName,
	}

	// Append the ConditionalSpec to the child concept data's ConditionalSpecs slice.
	// childConceptData is a pointer already stored in conceptLineScope.Data, so
	// the append modifies the in-place struct directly without needing a reassign.
	childConceptData.ConditionalSpecs = append(childConceptData.ConditionalSpecs, *conditionalSpec)

	l.logExit("ConditionalSpec", &ctx.BaseParserRuleContext)
}

// EnterConceptLine is called when entering a concept line.
// When inside a conditional spec, it pushes a new ConceptScope for collecting
// the child concept's generator data. For regular concept definitions, it
// does not push a scope to preserve backward compatibility with the existing
// symbol table structure where generators are direct children of the
// ConceptDefinition scope.
//
// Example: State=[Values[NY, CA, TX]]
// This corresponds to the 'conceptLine' rule in CPLConcepts.g4.
func (l *CPLListener) EnterConceptLine(ctx *antlr.ConceptLineContext) {
	l.logEnter("ConceptLine", &ctx.BaseParserRuleContext)

	// Only push a scope when inside a conditional spec.
	// For regular concept definitions, generators remain direct children
	// of the ConceptDefinition scope.
	parent := ctx.GetParent()
	if _, ok := parent.(*antlr.ConditionalSpecContext); ok {
		current, err := l.symbolStack.Peek()
		if err != nil {
			l.AddError(fmt.Errorf("symbol stack error entering concept line: %w", err))
			return
		}
		l.symbolStack.Push(current.AddChild(symboltable.ConceptScope))
	}
}

// ExitConceptLine is called when exiting a concept line.
// When inside a conditional spec, it extracts the concept name, walks through
// child generator scopes to collect values, builds a ConceptDefinitionData,
// stores it in the scope's Data field, and pops the concept line scope from
// the symbol stack. For regular concept definitions, it is a no-op.
func (l *CPLListener) ExitConceptLine(ctx *antlr.ConceptLineContext) {
	// Only process when inside a conditional spec
	parent := ctx.GetParent()
	if _, ok := parent.(*antlr.ConditionalSpecContext); !ok {
		l.logExit("ConceptLine", &ctx.BaseParserRuleContext)
		return
	}

	scope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting concept line: %w", err))
		l.logExit("ConceptLine", &ctx.BaseParserRuleContext)
		return
	}

	// Extract concept name from identifier
	identifierCtx := ctx.Identifier()
	if identifierCtx == nil {
		l.AddError(&SemanticError{
			Message: "missing concept name in concept line",
			Rule:    "conceptLine",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptLine", &ctx.BaseParserRuleContext)
		return
	}
	conceptName := identifierCtx.GetText()

	// Validation: concept name must be non-empty
	if conceptName == "" {
		l.AddError(&SemanticError{
			Message: "concept line requires a non-empty concept name",
			Rule:    "conceptLine",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptLine", &ctx.BaseParserRuleContext)
		return
	}

	// Walk through child generator scopes to collect values.
	// Reads from table.Definitions (identifier symbols), the same data source
	// used by AnalyzeValueGeneratorScope for regular concepts. This unifies
	// both paths to a single data source and avoids dual-path maintenance risk.
	var values []string
	for _, child := range scope.Children {
		if child.ScopeType == symboltable.ValueGeneratorScope {
			for _, def := range child.Definitions {
				if def.DefinitionType == symboltable.Identifier && def.Name != nil {
					values = append(values, *def.Name)
				}
			}
		}
	}

	// Build ConceptDefinitionData
	conceptData := &ast.ConceptDefinitionData{
		Name:          conceptName,
		DefaultValues: values,
		SourceLocation: ast.SourceLocation{
			Line:   ctx.GetStart().GetLine(),
			Column: ctx.GetStart().GetColumn(),
		},
	}

	// Store in scope Data for analysis phase
	scope.Data = conceptData

	l.logExit("ConceptLine", &ctx.BaseParserRuleContext)
}

// EnterValuesPatternGenerator is called when entering a Values[...] generator.
// It initializes a new ValuesGeneratorDef to collect the values.
//
// Example: Values[USA, Canada, Mexico]
// This corresponds to the 'valuesPatternGenerator' rule in CPLConcepts.g4.
func (l *CPLListener) EnterValuesPatternGenerator(ctx *antlr.ValuesPatternGeneratorContext) {
	l.logEnter("ValuesPatternGenerator", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the concept scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering concept definition: %w", err))
		return
	}

	// Create a child scope for this concept definition and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.ValueGeneratorScope))
}

// ExitValuesPatternGenerator is called when exiting a Values[...] generator.
// It pops the value generator scope from the symbol stack.
//
// Both the regular concept analysis path (AnalyzeValueGeneratorScope) and the
// conditional spec path (ExitConceptLine) read values from table.Definitions
// (identifier symbols). This keeps value extraction logic in a single code path
// via EnterValue/ExitValue in the main listener.
func (l *CPLListener) ExitValuesPatternGenerator(ctx *antlr.ValuesPatternGeneratorContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting values pattern generator: %w", err))
	}
	l.logExit("ValuesPatternGenerator", &ctx.BaseParserRuleContext)
}

// EnterConceptPatternGenerator is called when entering a Concept[StartValue..EndValue] generator.
// It pushes a new ConceptGeneratorScope onto the symbol stack.
//
// Example: Quarter[Q1..Q4]
// This corresponds to the 'conceptPatternGenerator' rule in CPLConcepts.g4.
func (l *CPLListener) EnterConceptPatternGenerator(ctx *antlr.ConceptPatternGeneratorContext) {
	l.logEnter("ConceptPatternGenerator", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the concept generator scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering concept pattern generator: %w", err))
		return
	}

	// Create a child scope for this concept range generator and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.ConceptGeneratorScope))
}

// ExitConceptPatternGenerator is called when exiting a Concept[StartValue..EndValue] generator.
// It extracts the concept name, start value, and end value, validates them,
// stores a ConceptRangeGeneratorDef in the scope's Data field, and pops the
// concept generator scope from the symbol stack.
func (l *CPLListener) ExitConceptPatternGenerator(ctx *antlr.ConceptPatternGeneratorContext) {
	scope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting concept pattern generator: %w", err))
		l.logExit("ConceptPatternGenerator", &ctx.BaseParserRuleContext)
		return
	}

	// Extract concept name from the context
	identifierCtx := ctx.Identifier()
	if identifierCtx == nil {
		l.AddError(&SemanticError{
			Message: "missing concept name in Concept range generator",
			Rule:    "conceptPatternGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptPatternGenerator", &ctx.BaseParserRuleContext)
		return
	}
	conceptName := identifierCtx.GetText()

	// Extract concept range from the context
	conceptRangeCtx := ctx.ConceptRange()
	if conceptRangeCtx == nil {
		l.AddError(&SemanticError{
			Message: "missing range in Concept range generator",
			Rule:    "conceptPatternGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptPatternGenerator", &ctx.BaseParserRuleContext)
		return
	}

	// Extract start and end values from concept range
	identifiers := conceptRangeCtx.AllIdentifier()
	if len(identifiers) < 2 {
		l.AddError(&SemanticError{
			Message: "Concept range requires start and end values",
			Rule:    "conceptPatternGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptPatternGenerator", &ctx.BaseParserRuleContext)
		return
	}

	startValue := identifiers[0].GetText()
	endValue := identifiers[1].GetText()

	// Validation: concept name must be a valid identifier
	if conceptName == "" {
		l.AddError(&SemanticError{
			Message: "concept name cannot be empty in Concept range generator",
			Rule:    "conceptPatternGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptPatternGenerator", &ctx.BaseParserRuleContext)
		return
	}

	// Validation: start and end values must be non-empty
	if startValue == "" {
		l.AddError(&SemanticError{
			Message: "start value cannot be empty in Concept range generator",
			Rule:    "conceptPatternGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptPatternGenerator", &ctx.BaseParserRuleContext)
		return
	}

	if endValue == "" {
		l.AddError(&SemanticError{
			Message: "end value cannot be empty in Concept range generator",
			Rule:    "conceptPatternGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptPatternGenerator", &ctx.BaseParserRuleContext)
		return
	}

	// Store the ConceptRangeGeneratorDef in the scope's Data field
	scope.Data = &ast.ConceptRangeGeneratorDef{
		ConceptName: conceptName,
		StartValue:  startValue,
		EndValue:    endValue,
	}

	l.logExit("ConceptPatternGenerator", &ctx.BaseParserRuleContext)
}

// EnterConceptRange is called when entering a concept range (StartValue..EndValue).
func (l *CPLListener) EnterConceptRange(ctx *antlr.ConceptRangeContext) {
	l.logEnter("ConceptRange", &ctx.BaseParserRuleContext)
}

// ExitConceptRange is called when exiting a concept range.
func (l *CPLListener) ExitConceptRange(ctx *antlr.ConceptRangeContext) {
	l.logExit("ConceptRange", &ctx.BaseParserRuleContext)
}

// EnterIntRangeGenerator is called when entering an Int[Start, End, Step] generator.
// It pushes a new IntGeneratorScope onto the symbol stack.
//
// Example: Int[1, 10, 2]
// This corresponds to the 'intRangeGenerator' rule in CPLConcepts.g4.
func (l *CPLListener) EnterIntRangeGenerator(ctx *antlr.IntRangeGeneratorContext) {
	l.logEnter("IntRangeGenerator", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the int generator scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering int range generator: %w", err))
		return
	}

	// Create a child scope for this int range generator and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.IntGeneratorScope))
}

// ExitIntRangeGenerator is called when exiting an Int[Start, End, Step] generator.
// It extracts the integer parameters, validates them, stores an IntRangeGeneratorDef
// in the scope's Data field, and pops the int generator scope from the symbol stack.
func (l *CPLListener) ExitIntRangeGenerator(ctx *antlr.IntRangeGeneratorContext) {
	scope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting int range generator: %w", err))
		l.logExit("IntRangeGenerator", &ctx.BaseParserRuleContext)
		return
	}

	// Extract integer parameters from the context
	paramsCtx := ctx.IntRangeParams()
	if paramsCtx == nil {
		l.AddError(&SemanticError{
			Message: "missing range parameters in Int generator",
			Rule:    "intRangeGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("IntRangeGenerator", &ctx.BaseParserRuleContext)
		return
	}

	// Extract integer parameters from the context.
	// Since NUMBER token (added for floating-point arithmetic expression support)
	// appears before INTEGER in the lexer and matches all integer literals, we must
	// collect from both AllNUMBER() and AllINTEGER() to handle both token types.
	integerNodes := paramsCtx.AllNUMBER()
	integerNodes = append(integerNodes, paramsCtx.AllINTEGER()...)
	if len(integerNodes) < 2 {
		l.AddError(&SemanticError{
			Message: "Int range generator requires at least start and end parameters",
			Rule:    "intRangeGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("IntRangeGenerator", &ctx.BaseParserRuleContext)
		return
	}

	start, err := strconv.Atoi(integerNodes[0].GetText())
	if err != nil {
		l.AddError(&SemanticError{
			Message: fmt.Sprintf("invalid start value in Int range generator: %s", integerNodes[0].GetText()),
			Rule:    "intRangeGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("IntRangeGenerator", &ctx.BaseParserRuleContext)
		return
	}

	end, err := strconv.Atoi(integerNodes[1].GetText())
	if err != nil {
		l.AddError(&SemanticError{
			Message: fmt.Sprintf("invalid end value in Int range generator: %s", integerNodes[1].GetText()),
			Rule:    "intRangeGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("IntRangeGenerator", &ctx.BaseParserRuleContext)
		return
	}

	step := 1
	if len(integerNodes) >= 3 {
		step, err = strconv.Atoi(integerNodes[2].GetText())
		if err != nil {
			l.AddError(&SemanticError{
				Message: fmt.Sprintf("invalid step value in Int range generator: %s", integerNodes[2].GetText()),
				Rule:    "intRangeGenerator",
				Line:    ctx.GetStart().GetLine(),
				Column:  ctx.GetStart().GetColumn(),
			})
			l.logExit("IntRangeGenerator", &ctx.BaseParserRuleContext)
			return
		}
	}

	// Track validation errors to avoid storing invalid data
	hasValidationError := false

	// Validation: step cannot be zero
	if step == 0 {
		l.AddError(&SemanticError{
			Message: "step cannot be zero in Int range generator",
			Rule:    "intRangeGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		hasValidationError = true
	}

	// Validation: start must be less than or equal to end when step is positive
	if step > 0 && start > end {
		l.AddError(&SemanticError{
			Message: fmt.Sprintf("start (%d) must be less than or equal to end (%d) when step is positive", start, end),
			Rule:    "intRangeGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		hasValidationError = true
	}

	// Validation: start must be greater than or equal to end when step is negative
	if step < 0 && start < end {
		l.AddError(&SemanticError{
			Message: fmt.Sprintf("start (%d) must be greater than or equal to end (%d) when step is negative", start, end),
			Rule:    "intRangeGenerator",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		hasValidationError = true
	}

	// If any validation error was recorded, do not store invalid data
	if hasValidationError {
		l.logExit("IntRangeGenerator", &ctx.BaseParserRuleContext)
		return
	}

	// Store the IntRangeGeneratorDef in the scope's Data field
	scope.Data = &ast.IntRangeGeneratorDef{
		Start: start,
		End:   end,
		Step:  step,
	}

	l.logExit("IntRangeGenerator", &ctx.BaseParserRuleContext)
}

// EnterConceptValueList is called when entering a concept value list.
// This method extracts all identifier values from the context.
//
// Example: USA, Canada, Mexico in Values[USA, Canada, Mexico]
func (l *CPLListener) EnterConceptValueList(ctx *antlr.ConceptValueListContext) {
	l.logEnter("ConceptValueList", &ctx.BaseParserRuleContext)
}

// ExitConceptValueList is called when exiting a concept value list.
// In the symbol table-driven pattern, value extraction happens during
// the analysis phase by walking the symbol table.
func (l *CPLListener) ExitConceptValueList(ctx *antlr.ConceptValueListContext) {
	// Symbol table-driven parsing: value lists are processed during analysis phase
	l.logExit("ConceptValueList", &ctx.BaseParserRuleContext)
}

// EnterConceptExtension is called when entering a concept extension.
// It pushes a new ConceptScope onto the symbol stack.
//
// Example concept extension: Priority=[Values[Critical]] before Low
// This corresponds to the 'conceptExtension' rule in CPLConcepts.g4.
func (l *CPLListener) EnterConceptExtension(ctx *antlr.ConceptExtensionContext) {
	l.logEnter("ConceptExtension", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the concept scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering concept extension: %w", err))
		return
	}

	// Create a child scope for this concept extension and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.ConceptScope))
}

// ExitConceptExtension is called when exiting a concept extension.
// It extracts the concept name and insertion point, validates them, builds a
// ConceptDefinitionData, stores it in the scope's Data field, and pops the
// concept scope from the symbol stack.
//
// Semantic validation ensures:
//   - The concept name is non-empty (belt-and-suspenders; grammar should
//     guarantee an identifier is present at the start of the rule)
func (l *CPLListener) ExitConceptExtension(ctx *antlr.ConceptExtensionContext) {
	scope, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting concept extension: %w", err))
		l.logExit("ConceptExtension", &ctx.BaseParserRuleContext)
		return
	}

	// Extract concept name from identifier
	identifierCtx := ctx.Identifier()
	if identifierCtx == nil {
		l.AddError(&SemanticError{
			Message: "missing concept name in concept extension",
			Rule:    "conceptExtension",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptExtension", &ctx.BaseParserRuleContext)
		return
	}
	conceptName := identifierCtx.GetText()

	// Validation: concept name must be non-empty
	// Belt-and-suspenders: the grammar (conceptExtension: identifier EQUALS ...)
	// guarantees an identifier is present, but we defensively check for empty text.
	if conceptName == "" {
		l.AddError(&SemanticError{
			Message: "concept extension requires a non-empty concept name",
			Rule:    "conceptExtension",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("ConceptExtension", &ctx.BaseParserRuleContext)
		return
	}

	// Look for InsertionPoint data stored by ExitInsertionPoint
	var insertionPoint ast.InsertionPoint
	if scope.Data != nil {
		if ip, ok := scope.Data.(*ast.InsertionPoint); ok {
			insertionPoint = *ip
		}
	}

	// Build ConceptDefinitionData
	conceptData := &ast.ConceptDefinitionData{
		Name:           conceptName,
		InsertionPoint: insertionPoint,
		SourceLocation: ast.SourceLocation{
			Line:   ctx.GetStart().GetLine(),
			Column: ctx.GetStart().GetColumn(),
		},
	}

	// Store in scope Data for analysis phase
	scope.Data = conceptData

	l.logExit("ConceptExtension", &ctx.BaseParserRuleContext)
}

// EnterInsertionPoint is called when entering an insertion point.
//
// Example: before X or after Y
// This corresponds to the 'insertionPoint' rule in CPLConcepts.g4.
func (l *CPLListener) EnterInsertionPoint(ctx *antlr.InsertionPointContext) {
	l.logEnter("InsertionPoint", &ctx.BaseParserRuleContext)
}

// ExitInsertionPoint is called when exiting an insertion point.
// It extracts the BEFORE/AFTER keyword and anchor identifier, validates them,
// builds an InsertionPoint, and stores it in the current scope's Data field.
//
// Semantic validation ensures:
//   - The anchor identifier is non-empty (belt-and-suspenders; grammar should
//     guarantee an identifier follows BEFORE/AFTER)
//   - Exactly one of BEFORE or AFTER is present (belt-and-suspenders; grammar
//     should guarantee one direction token)
func (l *CPLListener) ExitInsertionPoint(ctx *antlr.InsertionPointContext) {
	beforeToken := ctx.BEFORE()
	afterToken := ctx.AFTER()
	identifierCtx := ctx.Identifier()

	// Validation: anchor identifier must be present and non-empty
	// Belt-and-suspenders: the grammar (insertionPoint: BEFORE identifier | AFTER identifier)
	// guarantees an identifier is present, but we defensively check for nil/empty.
	identifierValue := ""
	if identifierCtx != nil {
		identifierValue = identifierCtx.GetText()
	}
	if identifierValue == "" {
		l.AddError(&SemanticError{
			Message: "before/after requires a value identifier",
			Rule:    "insertionPoint",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("InsertionPoint", &ctx.BaseParserRuleContext)
		return
	}

	// Determine the insertion direction from the BEFORE/AFTER token
	var insertionPoint ast.InsertionPoint
	if beforeToken != nil {
		insertionPoint.BeforeValue = identifierValue
	} else if afterToken != nil {
		insertionPoint.AfterValue = identifierValue
	} else {
		// Belt-and-suspenders: grammar should guarantee one direction token is present
		l.AddError(&SemanticError{
			Message: "insertion point must specify 'before' or 'after'",
			Rule:    "insertionPoint",
			Line:    ctx.GetStart().GetLine(),
			Column:  ctx.GetStart().GetColumn(),
		})
		l.logExit("InsertionPoint", &ctx.BaseParserRuleContext)
		return
	}

	// Store in current scope's Data — will be read by ExitConceptExtension
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitInsertionPoint: %w", err))
		l.logExit("InsertionPoint", &ctx.BaseParserRuleContext)
		return
	}
	current.Data = &insertionPoint

	l.logExit("InsertionPoint", &ctx.BaseParserRuleContext)
}

// EnterGeneratorPatternList is called when entering a generator pattern list.
// This is a placeholder for potential future validation.
func (l *CPLListener) EnterGeneratorPatternList(ctx *antlr.GeneratorPatternListContext) {
	l.logEnter("GeneratorPatternList", &ctx.BaseParserRuleContext)
}

// ExitGeneratorPatternList is called when exiting a generator pattern list.
// This is a placeholder for potential future validation.
func (l *CPLListener) ExitGeneratorPatternList(ctx *antlr.GeneratorPatternListContext) {
	l.logExit("GeneratorPatternList", &ctx.BaseParserRuleContext)
}

// EnterGeneratorPattern is called when entering a generator pattern.
// This is a placeholder for potential future validation.
func (l *CPLListener) EnterGeneratorPattern(ctx *antlr.GeneratorPatternContext) {
	l.logEnter("GeneratorPattern", &ctx.BaseParserRuleContext)
}

// ExitGeneratorPattern is called when exiting a generator pattern.
// This is a placeholder for potential future validation.
func (l *CPLListener) ExitGeneratorPattern(ctx *antlr.GeneratorPatternContext) {
	l.logExit("GeneratorPattern", &ctx.BaseParserRuleContext)
}

# Task 1a Walkthrough: TestParseTwoTypesSections

## What Was Implemented

The [`TestParseTwoTypesSections`](src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go:713) test function was added to [`cpl_parser_types_test.go`](src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go).

## What It Verifies

Parsing a CPL file that contains two separate `types` sections produces two independent [`TypesScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go) children under the root document ([`GlobalScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go)). Each `TypesScope` contains its own [`TypeScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go) children with the type definitions declared in that section.

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    Percentage = Number(precision=2, min=0, max=100)
```

## Assertions Documented

1. **Parse succeeds** — `parser.Parse(input)` returns no error and a non-nil document.
2. **Two TypesScope children** — `doc.Children` has exactly two entries, both with `ScopeType == TypesScope`.
3. **Parent linkage** — Each `TypesScope`'s `Parent` pointer is the root `doc` (`GlobalScope`).
4. **TypeScope children present** — Each `TypesScope` has at least one `TypeScope` child.
5. **Correct type definitions** —
   - The first `TypesScope` contains a `TypeScope` with a definition named `Currency`.
   - The second `TypesScope` contains a `TypeScope` with a definition named `Percentage`.

## Key Finding

Blank lines between `types` sections create extra `INDENT` tokens in the lexer output. Because the `typeDefinition` grammar rule uses the `*` (zero-or-more) quantifier, these extra indent tokens can produce empty or superfluous `TypeScope` children.

To keep the test robust against this known artifact, the implementation uses the existing helper [`filterScopesByType`](src/chronocone_planning_language/src/main/go/parser/cpl_concepts_multi_section_test.go:266) to extract only the `TypeScope` entries, then applies `GreaterOrEqual` assertions rather than insisting on exact child counts. This pattern is consistent with other multi-section tests in the same package.

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "parser_test.TestParseTwoTypesSections"
```

---

# Task 1b Walkthrough: TestParseThreeTypesSections

## What Was Implemented

The [`TestParseThreeTypesSections`](src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go:782) test function was added to [`cpl_parser_types_test.go`](src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go).

## What It Verifies

Parsing a CPL file that contains three separate `types` sections produces three independent [`TypesScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go) children under the root document ([`GlobalScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go)). Each `TypesScope` contains its own [`TypeScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go) children with the type definitions declared in that section.

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    Percentage = Number(precision=2, min=0, max=100)

types
    Amount = Number()
```

## Assertions Documented

1. **Parse succeeds** — `parser.Parse(input)` returns no error and a non-nil document.
2. **Three TypesScope children** — `doc.Children` has exactly three entries, all with `ScopeType == TypesScope`.
3. **Parent linkage** — Each `TypesScope`'s `Parent` pointer is the root `doc` (`GlobalScope`).
4. **TypeScope children present** — Each `TypesScope` has at least one `TypeScope` child.
5. **Correct type definitions** —
   - The first `TypesScope` contains a `TypeScope` with a definition named `Currency`.
   - The second `TypesScope` contains a `TypeScope` with a definition named `Percentage`.
   - The third `TypesScope` contains a `TypeScope` with a definition named `Amount`.

## Key Finding

The same blank-line / extra-`INDENT` artifact noted in Task 1a applies here. To keep the test robust, the implementation reuses the [`filterScopesByType`](src/chronocone_planning_language/src/main/go/parser/cpl_concepts_multi_section_test.go:266) helper to extract only the `TypeScope` entries, then applies `GreaterOrEqual` assertions rather than insisting on exact child counts. This pattern is consistent with Task 1a and other multi-section tests in the same package.

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "parser_test.TestParseThreeTypesSections"
```

---

# Task 1c Walkthrough: TestParseTypesSectionsIsolation

## What Was Implemented

The [`TestParseTypesSectionsIsolation`](src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go:870) test function was added to [`cpl_parser_types_test.go`](src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go).

## What It Verifies

Cross-section type isolation: a type defined in one `types` section is **not** visible in the `TypeScope` children of another `types` section. Each [`TypesScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go) contains its own [`TypeScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go) children, and lookups for a type name fail when performed against `TypeScope` children belonging to a different section.

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    Percentage = Number(precision=2, min=0, max=100)
```

## Assertions Documented

1. **Parse succeeds** — `parser.Parse(input)` returns no error and a non-nil document.
2. **Two TypesScope children** — `doc.Children` has exactly two entries, both with `ScopeType == TypesScope`.
3. **TypeScope children present in each** — Each `TypesScope` has at least one `TypeScope` child (via [`filterScopesByType`](src/chronocone_planning_language/src/main/go/parser/cpl_concepts_multi_section_test.go:266)).
4. **Positive isolation checks** —
   - `Currency` is found in the first `TypesScope`'s `TypeScope` children.
   - `Percentage` is found in the second `TypesScope`'s `TypeScope` children.
5. **Negative isolation checks (cross-Lookup)** —
   - `Currency` **is not** found in any `TypeScope` child of the second `TypesScope`.
   - `Percentage` **is not** found in any `TypeScope` child of the first `TypesScope`.

## Key Finding

This test moves beyond structural verification (1a, 1b) to semantic isolation. The negative assertions use the same `Lookup` API as the positive ones, but expect an error return, confirming that type names are scoped to their own `TypesScope` subtree and do not leak across sibling sections.

The same blank-line / extra-`INDENT` artifact noted in Tasks 1a and 1b applies here. The implementation reuses the [`filterScopesByType`](src/chronocone_planning_language/src/main/go/parser/cpl_concepts_multi_section_test.go:266) helper to extract only the `TypeScope` entries, then iterates over them for both positive and negative lookups, keeping the test robust against empty or superfluous children.

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "parser_test.TestParseTypesSectionsIsolation"
```

---

# Task 1d Walkthrough: TestParseMixedTypesAcrossSections

## What Was Implemented

The [`TestParseMixedTypesAcrossSections`](src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go) test function was added to [`cpl_parser_types_test.go`](src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go).

## What It Verifies

Parsing a CPL file with three separate `types` sections containing different base types (Number, String, Boolean) produces three independent [`TypesScope`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go) children, each containing the correct type definition.

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    NamePattern = String(maximum=50)

types
    ActiveFlag = Boolean()
```

## Assertions Documented

1. **Parse succeeds** — `parser.Parse(input)` returns no error and a non-nil document.
2. **Three TypesScope children** — `doc.Children` has exactly three entries, all with `ScopeType == TypesScope`.
3. **Correct type definitions in each section** —
   - First `TypesScope` contains a `TypeScope` with `Currency`.
   - Second `TypesScope` contains a `TypeScope` with `NamePattern`.
   - Third `TypesScope` contains a `TypeScope` with `ActiveFlag`.

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "parser_test.TestParseMixedTypesAcrossSections"
```

---

# Task 2a Walkthrough: TestAnalyzeDuplicateTypeAcrossSections

## What Was Implemented

The [`TestAnalyzeDuplicateTypeAcrossSections`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go) test function was added to [`types_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go).

## What It Verifies

When two `types` sections both define a type with the same name ("Currency"), the analyzer detects the duplicate across sections and produces an error containing "duplicate".

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    Currency = Number(precision=4)
```

## Assertions Documented

1. `parseAndAnalyze(input)` returns an error.
2. The error message contains "duplicate".

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "analysis_test.TestAnalyzeDuplicateTypeAcrossSections"
```

---

# Task 2b Walkthrough: TestAnalyzeUniqueTypesAcrossSections

## What Was Implemented

The [`TestAnalyzeUniqueTypesAcrossSections`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go) test function was added to [`types_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go).

## What It Verifies

When two `types` sections define different type names ("Currency" and "Percentage"), analysis succeeds and both types appear in `defs.Types` with correct derivation paths.

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    Percentage = Number(precision=2, minimum=0, maximum=100)
```

## Assertions Documented

1. `parseAndAnalyze(input)` succeeds without error.
2. `defs.Types` contains both "Currency" and "Percentage".
3. Each type has the correct `DerivationPath` ending in `"Number"`.

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "analysis_test.TestAnalyzeUniqueTypesAcrossSections"
```

---

# Task 2c Walkthrough: TestAnalyzeThreeSectionsWithDuplicate

## What Was Implemented

The [`TestAnalyzeThreeSectionsWithDuplicate`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go) test function was added to [`types_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go).

## What It Verifies

When three `types` sections are present and the third section repeats a type name from the first ("Currency"), the analyzer detects the duplicate and produces an error mentioning both "duplicate" and the type name "Currency".

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    Percentage = Number(precision=2, minimum=0, maximum=100)

types
    Currency = Number(precision=4)
```

## Assertions Documented

1. `parseAndAnalyze(input)` returns an error.
2. The error message contains "duplicate".
3. The error message contains "Currency".

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "analysis_test.TestAnalyzeThreeSectionsWithDuplicate"
```

---

# Task 2d Walkthrough: TestSymbolTableWalkVisitsAllTypeScopes

## What Was Implemented

The [`TestSymbolTableWalkVisitsAllTypeScopes`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go) test function was added to [`types_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go).

## What It Verifies

The analyzer's symbol table walk visits all `TypeScope` children under all `TypesScope` tables, collecting every type definition into the analyzer's `Types` map. The total count equals the sum across all sections.

## Test Input

```cpl
types
    Currency = Number(precision=2)
    Amount = Number()

types
    Percentage = Number(precision=2, minimum=0, maximum=100)
```

## Assertions Documented

1. `parseAndAnalyze(input)` succeeds without error.
2. All three types ("Currency", "Amount", "Percentage") are present in `defs.Types`.
3. `len(defs.Types)` equals 3 (sum of unique types across sections).

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "analysis_test.TestSymbolTableWalkVisitsAllTypeScopes"
```

---

# Task 3a Walkthrough: TestParseAndAnalyzeNumberAndStringAcrossSections

## What Was Implemented

The [`TestParseAndAnalyzeNumberAndStringAcrossSections`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go) test function was added to [`types_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go).

## What It Verifies

End-to-end parse+analyze with a Number type in the first section and a String type in the second. Both types are correctly analyzed with their respective derivation paths and constraints.

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    NamePattern = String(maximum=50)
```

## Assertions Documented

1. `parseAndAnalyze(input)` succeeds without error.
2. "Currency" has `DerivationPath` `["Currency", "Number"]`.
3. "NamePattern" has `DerivationPath` `["NamePattern", "String"]` with `Maximum` constraint set to 50.

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "analysis_test.TestParseAndAnalyzeNumberAndStringAcrossSections"
```

---

# Task 3b Walkthrough: TestParseAndAnalyzeBooleanAcrossSections

## What Was Implemented

The [`TestParseAndAnalyzeBooleanAcrossSections`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go) test function was added to [`types_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go).

## What It Verifies

End-to-end parse+analyze with a Boolean type in a separate `types` section. The Boolean type is correctly analyzed with its derivation path and nil `ParsedConstraints`.

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    ActiveFlag = Boolean()
```

## Assertions Documented

1. `parseAndAnalyze(input)` succeeds without error.
2. "ActiveFlag" has `DerivationPath` `["ActiveFlag", "Boolean"]`.
3. `ParsedConstraints` is nil (Boolean has no constraints).

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "analysis_test.TestParseAndAnalyzeBooleanAcrossSections"
```

---

# Task 3c Walkthrough: TestParseAndAnalyzeInstantIntervalAcrossSections

## What Was Implemented

The [`TestParseAndAnalyzeInstantIntervalAcrossSections`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go) test function was added to [`types_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go).

## What It Verifies

End-to-end parse+analyze with Instant and Interval types declared in separate `types` sections, requiring concepts and units sections for time-based type support.

## Test Input

```cpl
concepts
    DurationUnit=[Values[month, day]]
    Year = [Int[1900, 3000, 1]]
    Month=[Values[January, ...]]
    DayOfMonth = [Int[1, 31, 1]]
units
    month = DurationUnit
types
    FiscalYearStart = Instant(month=January, day_of_month=1)
types
    FiscalMonth = Interval(starts_at=FiscalYearStart, unit=month)
```

## Assertions Documented

1. `parseAndAnalyze(input)` succeeds without error.
2. "FiscalYearStart" has `DerivationPath` `["FiscalYearStart", "Instant"]`.
3. "FiscalMonth" has `DerivationPath` `["FiscalMonth", "Interval"]`.
4. Interval constraints correctly reference `StartTypeName="FiscalYearStart"` and `UnitName="month"`.

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "analysis_test.TestParseAndAnalyzeInstantIntervalAcrossSections"
```

---

# Task 3d Walkthrough: TestAllTypesInSingleLookupMap

## What Was Implemented

The [`TestAllTypesInSingleLookupMap`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go) test function was added to [`types_analysis_test.go`](src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go).

## What It Verifies

When multiple sections define Number, String, and Boolean types, all type names are present in the single `defs.Types` lookup map. The total count equals the sum of unique types across all sections.

## Test Input

```cpl
types
    Currency = Number(precision=2)

types
    NamePattern = String(maximum=50)

types
    ActiveFlag = Boolean()
```

## Assertions Documented

1. `parseAndAnalyze(input)` succeeds without error.
2. All three type names appear in `defs.Types`.
3. `len(defs.Types)` equals 3.
4. Each type has the correct `DerivationPath`.

## How to Run

```bash
cd src/chronocone_planning_language && ./gradlew test --tests "analysis_test.TestAllTypesInSingleLookupMap"
```

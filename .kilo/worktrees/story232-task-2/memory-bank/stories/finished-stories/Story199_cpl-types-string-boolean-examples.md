# Story199: CPL Types String and Boolean Examples

## Goal

Create `.cpl` example files in `src/chronocone_planning_language/examples/` demonstrating Story196 functionality (String and Boolean built-in types) with both valid declarations and intentional errors. Update the examples README to document the new files.

This story depends on [Story196: CPL Types - String and Boolean Support](finished-stories/Story196_cpl-types-string-boolean.md).

## References

- [Story196: CPL Types - String and Boolean Support](finished-stories/Story196_cpl-types-string-boolean.md) - Parent story implementing String and Boolean types
- [Epic002: CPL Types Implementation](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Incremental implementation ERO
- [types-definitions.md](../specifications/design/cpl/types-definitions.md) - Type design document
- [examples/README.md](../../src/chronocone_planning_language/examples/README.md) - Examples documentation

## Dependencies

- [Story196_cpl-types-string-boolean.md](finished-stories/Story196_cpl-types-string-boolean.md) - Done (String and Boolean type analysis implementation)

## Dependent Stories

None - This is an examples and documentation story that does not block other work.

## Tasks

1. Create valid String example files in `without_errors/` - Completed
   a. Create `string_declarations.cpl` with String base type declarations including minimum, maximum, and pattern constraints - Completed
   b. Create `string_derivations.cpl` with String type derivation chains demonstrating constraint narrowing - Completed
   c. Verify each file parses and analyzes successfully by running through the parse+analyze pipeline - Completed

2. Create valid Boolean example files in `without_errors/` - Completed
   a. Create `boolean_declarations.cpl` with Boolean base type declarations - Completed
   b. Create `boolean_derivations.cpl` with Boolean type derivation chains demonstrating plain inheritance - Completed
   c. Verify each file parses and analyzes successfully by running through the parse+analyze pipeline - Completed

3. Create error example files in `with_errors/` - Completed
   a. Create `invalid_string_unknown_constraint.cpl` - unknown constraint name on String - Completed
   b. Create `invalid_string_bad_regex.cpl` - malformed regex pattern - Completed
   c. Create `invalid_string_empty_pattern.cpl` - empty pattern string - Completed
   d. Create `invalid_string_non_integer.cpl` - non-integer minimum value - Completed
   e. Create `invalid_string_negative_length.cpl` - negative minimum length - Completed
   f. Create `invalid_string_min_greater_than_max.cpl` - minimum greater than maximum - Completed
   g. Create `invalid_string_expanding_maximum.cpl` - expanding inherited maximum constraint - Completed
   h. Create `invalid_string_expanding_minimum.cpl` - expanding inherited minimum constraint - Completed
   i. Create `invalid_string_different_pattern.cpl` - different pattern in derivation chain - Completed
   j. Create `invalid_boolean_constraint.cpl` - constraint passed to Boolean type - Completed
   k. Verify each error example produces the expected parse or analysis error - Completed

4. Update examples README with documentation - Completed
   a. Add sections under "Files (without_errors)" for the four new valid example files - Completed
   b. Add sections under "Files (with_errors)" for the ten new error example files - Completed
   c. Follow existing format: filename, description, CPL snippet, "What it demonstrates" bullet list - Completed
   d. Verify README formatting is correct and all links work - Completed

## Test-First Development

This story follows the Logical TDD Lifecycle. For each task:

1. **Phase A (Structural Definition)**: Define the expected content and behavior of each example file before creating it
2. **Phase B (Logical Tests)**: For valid examples, verify they parse and analyze without errors; for error examples, verify they produce the expected specific error
3. **Phase C (Implementation)**: Create the example files and README documentation

The verification for valid examples is that they compile through the full parse+analyze pipeline. The verification for error examples is that they fail with the specific error message expected from Story196 implementation.

## Constraints

- All `.cpl` example files must follow the established comment style (first line describes purpose/expected behavior)
- Valid examples must parse and analyze successfully using the current parser and analyzer
- Error examples must produce the specific errors documented in the plan
- README additions must follow the existing format and structure
- Example files must align with Story196 acceptance criteria for String and Boolean types

## Intent

The intent of this story is to provide developer reference examples and documentation for the String and Boolean built-in types introduced in Story196, including:

- Clear examples of valid String and Boolean type declarations and derivations
- Reference error examples showing common mistakes and their failure modes
- Updated README documentation enabling developers to understand the examples without reading source code

## Acceptance Criteria

1. Four valid `.cpl` example files are created in `examples/without_errors/`
2. Ten error `.cpl` example files are created in `examples/with_errors/`
3. All valid examples parse and analyze successfully (no errors)
4. All error examples produce the expected parse or analysis errors
5. README is updated with documentation for all 14 new example files
6. README formatting follows existing conventions
7. All example files include descriptive comments on the first line

## Notes

### String Type Constraints (from Story196)

- **minimum**: Integer >= 0, minimum string length (higher value = more restrictive)
- **maximum**: Integer >= 0, maximum string length (lower value = more restrictive)
- **pattern**: Non-empty regex pattern compiled at analysis time

### Boolean Type Constraints (from Story196)

- No constraints allowed. `Boolean()` is the required syntax.

### Derivation Rules (from Story196)

- **String**: Derived types can only narrow constraints (increase minimum, decrease maximum, or add/set pattern). Same pattern is allowed; different pattern is rejected. Expanding maximum or minimum is rejected.
- **Boolean**: Plain inheritance only. Any constraint in derivation is rejected.

### Example File Reference

**Valid examples:**

- `string_declarations.cpl`: AnyString, ShortString, NamePattern, BoundedName, StrictName
- `string_derivations.cpl`: BaseString -> ShortString -> VeryShortString, BasePattern -> SamePattern
- `boolean_declarations.cpl`: ActiveFlag, EnabledFlag
- `boolean_derivations.cpl`: BaseBool -> DerivedBool

**Error examples:**

- `invalid_string_unknown_constraint.cpl`: String(foo=bar)
- `invalid_string_bad_regex.cpl`: String(pattern="[")
- `invalid_string_empty_pattern.cpl`: String(pattern="")
- `invalid_string_non_integer.cpl`: String(minimum=abc)
- `invalid_string_negative_length.cpl`: String(minimum=-1)
- `invalid_string_min_greater_than_max.cpl`: String(minimum=50, maximum=10)
- `invalid_string_expanding_maximum.cpl`: BaseString(maximum=50) -> WideString(maximum=100)
- `invalid_string_expanding_minimum.cpl`: BaseString(minimum=10) -> LooseString(minimum=5)
- `invalid_string_different_pattern.cpl`: BasePattern(pattern="^[a-z]+$") -> NumberPattern(pattern="^[0-9]+$")
- `invalid_boolean_constraint.cpl`: Boolean(active=true)

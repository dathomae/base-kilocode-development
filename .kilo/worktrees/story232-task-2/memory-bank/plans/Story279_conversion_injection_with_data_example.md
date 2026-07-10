# Plan: Story279 — Conversion Injection `with_data` Example

## Purpose

Create an end-to-end `with_data/without_errors/units_conversions/conversion_injection` example that exercises Story232's conversion injection pipeline. The example demonstrates that when a metric computes in one unit and the IO spec requires another, and a valid conversion pair exists, the injection pass wraps the CIL tree with `CILConversionApplication` nodes, the synthesizer emits conversion function calls, and the generated code produces numerically correct output in the target unit.

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `with_data-conversion-injection-example`                                                     |
| story_number | 279                                                                                          |
| description  | Create a with_data/without_errors example verifying Story232's conversion injection: a metric computing EUR with IO spec requiring USD and an EUR→USD conversion pair, demonstrating that injection produces working runtime code with correct numerical results |

## Scope

This story creates a single `with_data` example directory (with companion CPL, IO spec, parquet data, and expected output) that validates the full conversion injection pipeline end-to-end. No code changes to the compiler or test infrastructure are needed — Story277 already established the `with_errors/` and `without_errors/` directory layout, the `units_conversions/core.cpl` shared infrastructure, and the test discovery mechanism.

## Dependencies

- Story232 (conversion injection) — the injection pipeline must be complete and wired into synthesis before this example can run
- Story277 (directory restructure) — `with_data/without_errors/units_conversions/` directory structure and `core.cpl` must exist
- Story230 (conversions) — conversion pair definition + synthesis
- Story231 (unit type checking) — checker must pass the example (bridgeable gap detected, no error)
- The existing `cpl_with_data_test.go` test infrastructure (no modifications needed — already discovers examples under `without_errors/`)

## Files to Create or Modify

| File                                                                                       | Action | Purpose                                                                                           |
| ------------------------------------------------------------------------------------------ | ------ | ------------------------------------------------------------------------------------------------- |
| `examples/with_data/without_errors/units_conversions/conversion_injection/`                | Create | Leaf example directory                                                                            |
| `examples/with_data/without_errors/units_conversions/conversion_injection/iospec.json`     | Create | IO spec with `"unit": "USD"` on output; connector pointing at EUR parquet data                    |
| `examples/with_data/without_errors/units_conversions/conversion_injection/conversion_injection.cpl` | Create | CPL: connector in EUR, metric computes in EUR, IO spec requires USD, EUR↔USD conversion exists   |
| `examples/with_data/without_errors/units_conversions/conversion_injection/data_a.parquet`  | Create | Parquet data with EUR values (e.g., 100.0, 200.0 EUR)                                             |
| `examples/with_data/without_errors/units_conversions/conversion_injection/expected_output.xlsx` | Create | Expected output with USD values (applying the conversion rate)                                    |
| `examples/with_data/without_errors/units_conversions/core.cpl`                             | Modify | May need to add Currency concept infrastructure if not already present from Story277              |

## Example Design

### Scenario

A connector provides revenue data denominated in EUR. A metric computes revenue as an identity passthrough. The IO spec declares the output unit as USD. A conversion pair `EUR→USD` exists (e.g., `EUR(var) = USD(var) * 0.85(EUR)` — representing an exchange rate of 1 EUR = 0.85 USD, or equivalent algebraic form). The synthesis pipeline should:

1. **Story231 unit checker**: Identify that the computed unit (EUR) mismatches the IO spec unit (USD), but find a conversion pair bridging the gap — report no error
2. **Story232 injection**: Walk the CIL tree, find the root expression producing EUR, wrap it in `CILConversionApplication(from=EUR, to=USD)`
3. **Story230 synthesis**: Emit the generated conversion function call (e.g., `convert_EUR_to_USD(expr)`)
4. **Runtime**: Generated code loads EUR values from parquet, applies the conversion, and outputs USD values

### CPL Structure

The `conversion_injection.cpl` file must be self-contained with the connector and metric declarations, using the shared `core.cpl` from the `units_conversions/` category for common definitions.

If the `units_conversions/core.cpl` from Story277 already declares `Currency`-related concepts/units/types (since the error examples in Story277 use Currency), modify it to also include the conversion pair. If Story277's `core.cpl` only has Length infrastructure, extend core.cpl with:

```cpl
concepts
    Currency = [Values[EUR, USD, GBP]]

units
    Currency = Currency

types
    Amount = Number(precision=2)

measures
    revenue = Amount(unit=Currency)

conversions
    EUR(var) = USD(var) * factor(EUR)
where
    factor = ExternalInput::source [EUR_USD_Rate] celltype factor
```

The conversion declaration must use the actual CPL syntax from Story230's implementation for defining a conversion pair between EUR and USD.

### Input Data

Parquet file `data_a.parquet` with columns matching the connector dimensions (e.g., `Region` × `SingleDay` × `Value`):
- Region=North, Monday, Value=100.0 (EUR)
- Region=South, Monday, Value=200.0 (EUR)

### Expected Output

If the conversion rate is (for example) `EUR(var) = USD(var) * 1.10(USD)` meaning 1 EUR = 1.10 USD:
- Region=North, Monday, Value=110.0 (USD)
- Region=South, Monday, Value=220.0 (USD)

The exact conversion factor depends on what CPL syntax is used for the conversion formula. The expected output must match the arithmetic of the chosen conversion pair.

### How the Example Differs from Story277's `conversion_bridge`

- Story277's `conversion_bridge`: Verifies that the unit checker **permits** compilation when a conversion pair exists (the checker sees the gap is bridgeable). Does not require injection to actually produce the conversion at runtime.
- Story279's `conversion_injection`: Verifies that the injection pipeline **actually produces** a conversion at runtime — the generated code calls the conversion function, the output values are transformed from EUR to USD, and the result matches ground truth.

### Verification

The example is discovered automatically by `TestWithDataExamples` (which walks all `without_errors/` leaf directories). The test:
1. Compiles the CPL → verifies no error (checker allows it, injection succeeds)
2. Runs the generated program → verifies no runtime error
3. Verifies output parquet against `expected_output.xlsx` → verifies numerical correctness

### Additional: Generated Code Assertion

Optionally (as a sub-test or separate verification), inspect the generated Go source to confirm that a conversion function call (e.g., `convert_EUR_to_USD`) appears in the output. This validates that injection actually modified the CIL tree and the synthesizer emitted the call. This is distinct from the runtime verification and provides stronger evidence of injection.

## Tasks

### Task Sizing

- **Task 1** (Core infrastructure): `[Small]` — extends `units_conversions/core.cpl` with Currency concepts/units/types/measures if not already present from Story277. Single file modification.
- **Task 2** (Injection example): Medium — creates the complete leaf example with CPL, IO spec, parquet data, expected output, and verification. Not `[Small]` due to multiple coordinated file types requiring internal consistency and the need for correct conversion arithmetic.

### Execution Order

```
Task 1 → Task 2
```

- Task 1 establishes the shared infrastructure that Task 2's example depends on.

### Task 1: Extend `units_conversions/core.cpl` with Currency infrastructure — Not Started `[Small]`

- Add Currency concept, unit, type, and measure declarations to `core.cpl` if not already present
- Add a conversion pair for EUR↔USD using Story230's CPL conversion syntax
- The conversion must be algebraically reversible (Story230 derives the inverse) so both directions work
- Ensure `core.cpl` remains backward-compatible with Story277's `conversion_bridge` example (no conflicts with existing Length infrastructure)

### Task 2: Create `conversion_injection` example — Not Started

#### 2a. Create directory and CPL file

- Create `examples/with_data/without_errors/units_conversions/conversion_injection/`
- Write `conversion_injection.cpl`:
  - Connector: provides EUR data (e.g., `dimensions [Region] grain SingleDay celltype revenue`)
  - Metric: compute identity passthrough of the connector input
  - Declare any additional measures/types/concepts needed (if not in `core.cpl`)

#### 2b. Create IO specification

- Write `iospec.json`:
  - Input: connector_name → parquet file with EUR data
  - Output: metric name, dimensions, celltype, grain, and `"unit": "USD"`

#### 2c. Create parquet input data

- Generate `data_a.parquet` with known EUR values for each dimension combination
- Values should be simple enough to verify manually (e.g., 100.0, 200.0)

#### 2d. Create expected output

- Compute expected USD values by applying the conversion factor to each EUR input
- Create `expected_output.xlsx` with the computed values
- Use the `xls_verify` tool or manually construct the XLSX

#### 2e. Verify end-to-end

- The example is discovered by `TestWithDataExamples`
- Compilation succeeds (checker permits, injection modifies CIL tree, synthesizer emits code)
- Generated program runs without error
- Output values match expected (EUR values converted to USD)
- All existing `with_data` examples continue to pass (including Story277's `conversion_bridge`)

#### 2f. (Optional) Verify generated code contains conversion call

- After compilation, read the generated `main.go` and assert it contains the expected conversion function name (e.g., `convert_EUR_to_USD`)
- This is a stronger assertion than runtime verification alone — it proves injection actually happened

## Constraints

1. Follow the [Logical TDD Lifecycle](../concepts.md#logical-tdd-lifecycle)
2. All CPL must use valid syntax compatible with the current compiler pipeline
3. The `core.cpl` modification must not break Story277's `conversion_bridge` example (which uses Length infrastructure)
4. The conversion pair must be declared using Story230's actual CPL conversion syntax
5. The example must be discoverable by `TestWithDataExamples` (standard leaf directory structure: iospec.json + .cpl + .parquet + expected_output.xlsx)
6. Numerical output must be correct — the conversion arithmetic must match the CPL conversion formula
7. This example does NOT need to test error cases (those are covered by Story277's `with_errors/` examples)
8. The conversion rate used should produce clean numerical results to simplify verification (avoid floating-point edge cases)

## Edge Cases

- **Conversion factor equals 1.0**: If the EUR↔USD conversion rate is 1:1, the output values are numerically equal to input values — this is correct behavior but would not prove injection occurred. Use a non-trivial conversion factor (e.g., 1 EUR = 1.10 USD).
- **Negative values**: Revenue could be negative (e.g., refunds). The conversion should handle negative values correctly. Test data could include a negative row if the domain makes sense.
- **Zero values**: If any input value is 0 EUR, the output should be 0 USD after conversion.
- **Existing `core.cpl` content**: If Story277 already put Currency infrastructure in `core.cpl`, Task 1 becomes a no-op or minor extension. If Story277 did not, Task 1 adds it.
- **Conversion CPL syntax**: The exact syntax for declaring conversions in CPL may differ from the pseudocode shown above. Reference Story230's tests for the actual syntax.

## Verification Commands

After the example is created:

```bash
# Verify the new example passes end-to-end
go test ./src/main/go/cli/compiler/... -run TestWithDataExamples -count=1 -timeout 120s

# Verify existing examples still pass (regression check)
go test ./src/main/go/cli/compiler/... -run TestWithDataExamples_Discovery -count=1

# Full CPL test suite
go test ./src/chronocone_planning_language/src/main/go/... -count=1 -timeout 180s
```

## Notes

- This story is intentionally narrow in scope — a single example that proves the conversion injection pipeline works end-to-end. This follows the project pattern where each `with_data` example validates one specific feature or scenario.
- The `units_conversions/` category will have at least two examples after this story: `conversion_bridge` (Story277 — checker permits compilation) and `conversion_injection` (Story279 — injection produces working runtime code).
- The optional generated-code inspection adds confidence that injection actually modified the CIL tree, but the runtime verification (output matches expected) is the primary acceptance criterion.
- This example intentionally tests the simple case: identity metric with a single conversion at the output boundary. More complex scenarios (conversions injected into operand positions, multiple conversions in one expression) are tested by Story232's own Go integration tests.

# Plan: Adjust Lookups Example and Rename iOSpec `inputs` → `connectors`

## Objective

1. **Rename iospec.json section**: Change `"inputs"` → `"connectors"` and add a new `"lookups"` section to iospec.json (absorbing the separate `_lookups.io.json` file)
2. **Fix the lookups_basic example** data and alignment to produce a coherent demonstration

## Motivation

- "inputs" is confusing because it names only connector data, not all input data
- Lookup data sources should be specified in the same io spec (alongside connectors), using the same `data_format`/`driver`/`driver_parameters` structure, since both connectors and lookups may eventually come from databases or services
- The separate `_lookups.io.json` file (with its own distinct format) is replaced by a `"lookups"` section in iospec.json with the same schema as connectors
- The lookups_basic example has mismatches between the CPL, the data files, and the io spec

## Design Principle: iospec is a Spec, Not Data Storage

The iospec.json file specifies *where* data comes from and what gets produced — it does not contain the data itself. Just as connector input data lives in separate `*.json`/`*.parquet` files (named by convention after the example), lookup data lives in its own `<lookup_name>.json` and `<lookup_name>.parquet` files. The iospec.json `"lookups"` section uses the same structure as `"connectors"` to specify the data source, since in the future both may point to databases, services, or files.

## Proposed iospec.json Format (v2)

```json
{
    "connectors": [
        {
            "connector": "raw_strength",
            "data_format": "parquet",
            "driver": "file",
            "driver_parameters": [{"path": "raw_strength.parquet"}]
        }
    ],
    "lookups": [
        {
            "lookup": "strength_adjusted_for_material",
            "data_format": "parquet",
            "driver": "file",
            "driver_parameters": [{"path": "strength_adjusted_for_material.parquet"}]
        }
    ],
    "outputs": [
        {
            "metric": "product_strength",
            "dimensions": ["Product", "Material"],
            "celltype": "Strength",
            "grain": ["Daily"]
        }
    ],
    "architecture": {
        "cpu_cores": 4,
        "cache_size": 8388608,
        "main_memory_size": 8589934592
    }
}
```

Key changes:
- `"inputs"` → `"connectors"` (JSON key renamed)
- New `"lookups"` array with **same structure as connectors**: `lookup` (name), `data_format`, `driver`, `driver_parameters` — not a separate simpler format
- Lookup data files named `<lookup_name>.json`/`.parquet` (e.g., `strength_adjusted_for_material.json`), living alongside the example
- `"outputs"` and `"architecture"` unchanged

## Understanding the Current CPL and Data

### lookups_basic.cpl (current state)
```
lookups
    strength_adjusted_for_material = 
        dimensions[Material]
        celltype Strength

connectors
    raw_strength = dimensions [Product, Material] grain Daily celltype Strength

metrics
    product_strength = raw_strength * strength_adjusted_for_material
```
Note: Metric renamed to `product_strength` — no collision with the lookup name.
Note: Connector uses `dimensions [Product, Material]` — two dimensions.

### core.cpl (current state)
```
concepts
    Product = [Values[WidgetA, WidgetB, WidgetC]]
    Material = [Values[Steel, Aluminum, Copper]]
    Strength = [Values[PSI]]

units
    Strength=Strength

types
    Strength = Number()

measures
    Strength = Strength(unit=Strength)
```

### Data alignment issues address:

1. **Connector `raw_strength`** has `dimensions [Product, Material]` — so connector data maps (Product × Material × Daily) → Strength values. Products: WidgetA/WidgetB/WidgetC, Materials: Steel/Aluminum/Copper, Days: Sunday-Saturday. All connector values set to 1.0 (as specified).

2. **Lookup `strength_adjusted_for_material`** uses `dimensions[Material]` — the lookup data maps Material values (Steel, Aluminum, Copper) to Strength adjustment factors. Values: Steel=1.0, Aluminum=0.8, Copper=0.9.

3. **Metric naming**: Metric is now `product_strength`, distinct from the lookup name `strength_adjusted_for_material` — no collision.

4. **Output dimensions**: The output metric `product_strength` uses connector `raw_strength` (dims: Product, Material) and lookup `strength_adjusted_for_material` (dims: Material), so output dimensions should be `["Product", "Material"]` and grain `["Daily"]`.

5. **Stale comment**: Line 1-2 of lookups_basic.cpl says "lookup mapping Product to CategoryLabel" — update to reflect Material → Strength lookup.

### Data files needed:

**Connector data** (for `raw_strength` connector):
- Name: `raw_strength.json` (and `raw_strength.parquet`) — named after the connector
- Dimensions: Product (WidgetA, WidgetB, WidgetC) × Material (Steel, Aluminum, Copper) × Daily (Sunday-Saturday)
- That's 3 × 3 × 7 = 63 rows
- All values: 1.0

**Lookup data** (for `strength_adjusted_for_material` lookup):
- Name: `strength_adjusted_for_material.json` (and `.parquet`) — named after the lookup
- Dimensions: Material (Steel, Aluminum, Copper)
- Values: e.g., Steel=1.0, Aluminum=0.8, Copper=0.9 (adjustment factors)
- The old `lookups_basic_product_category.json` file can be removed or repurposed

## Changes Required — Phased

### Phase 1: Go Code Changes (Core iOSpec)

1. **`src/main/go/io/io_specification.go`**
   - Rename `Inputs []InputSpecification` → `Connectors []ConnectorSpecification` (both Go field name and JSON tag `"connectors"` for consistency)
   - Rename `InputSpecification` → `ConnectorSpecification` throughout
   - Add `Lookups []LookupSpecification` field with JSON tag `"lookups"`
   - Define `LookupSpecification` struct mirroring `ConnectorSpecification`:
     ```go
     type LookupSpecification struct {
         LookupName       string                   `json:"lookup"`
         DataFormat       string                   `json:"data_format"`
         Driver           string                   `json:"driver"`
         DriverParameters []map[string]interface{} `json:"driver_parameters"`
     }
     ```
   - Update `LoadIOSpecification` docs

2. **`src/main/go/synthesis/lookup_io_spec.go`** — Delete this file (functionality absorbed into iOSpec)

3. **`src/main/go/synthesis/input.go`**
   - Remove `LookupIOSpec *LookupIOSpec` field
   - Update `IOSpec *cio.IOSpecification` — lookups now come from IOSpec directly
   - Existing `Lookups map[string]*analysis.LookupInfo` stays (analysis phase data, different thing)

4. **`src/main/go/synthesis/synthesize.go`** (Synthesize function and MainFunction generation)
   - Remove `lookupIOSpec *LookupIOSpec` parameter from `Synthesize()` signature — lookups come from `ioSpec.Lookups`
   - Remove `input.LookupIOSpec = lookupIOSpec` assignment
   - All references to `ioSpec.Inputs` → `ioSpec.Connectors`
   - `needsInputs` (renamed to `needsConnectors`) check: also consider `len(ioSpec.Lookups) > 0`
   - `generateInputLoadingCode` → rename to `generateConnectorLoadingCode`, update parameter type
   - All `InputSpecification` references → `ConnectorSpecification`

5. **`src/main/go/synthesis/lookup_synthesis.go`**
   - Update `generateLookupIntegration` to use `ioSpec.Lookups` instead of separate `LookupIOSpec`
   - Update `generateLookupLoading` to iterate `ioSpec.Lookups` and resolve paths via `DriverParameters` (mirroring the `generateInputLoadingCode` approach for connectors)
   - The lookup's `path` is extracted from `DriverParameters[0]["path"]` (same pattern as connectors)

6. **`src/main/go/cli/compiler/cpl_compiler.go`**
   - Remove the `_lookups.io.json` discovery code (lines 218-236)
   - Remove `lookupIOSpec` parameter from `Synthesize()` call (signature changed in item 4)

7. **`src/main/go/connector/connector_loader.go`**
   - Update `ioSpec.Inputs` → `ioSpec.Connectors`
   - Update `InputSpecification` → `ConnectorSpecification`

8. **`src/main/go/cli/expected_xls_generator/generator.go`**
   - All `iospec.Inputs` → `iospec.Connectors`
   - All `InputSpecification` → `ConnectorSpecification`

9. **`src/main/go/cli/expected_xls_generator/main/main.go`**
   - All `iospec.Inputs` → `iospec.Connectors`

### Phase 3: Test Code Changes

10. **`src/main/go/io/io_specification_test.go`** — Update all JSON fixture strings from `"inputs"` to `"connectors"`

11. **`src/main/go/synthesis/synthesize_test.go`**
    - All `ioSpec.Inputs` → `ioSpec.Connectors`
    - Update `TestEndToEndSynthesizeLookups_Basic` to use new iospec structure with embedded lookups
    - Remove separate `LookupIOSpec` construction — pass lookups in iospec.Lookups

12. **`src/main/go/cli/compiler/cpl_compiler_test.go`** — Update JSON fixtures `"inputs"` → `"connectors"`

13. **`src/main/go/cli/compiler/cpl_integration_test.go`** — Update JSON fixtures

14. **`src/main/go/cli/compiler/cpl_example_test.go`** — Update JSON fixtures

### Phase 4: Example iospec.json Files (all 18+6=24 files)

15. **All `examples/with_data/**/iospec.json` files** (18 files) — Rename `"inputs"` → `"connectors"`

16. **All `examples/parse_and_analyze_only/**/*.json` files with iospec sections** (6 files) — Rename `"inputs"` → `"connectors"`

### Phase 5: Lookups Basic Example Fixes

17. **`examples/with_data/lookups/lookups_basic/lookups_basic.cpl`**
    - Update stale comment on line 2 ("Product to CategoryLabel" → e.g., "Material-based strength adjustment")

18. **`examples/with_data/lookups/lookups_basic/iospec.json`**
    - Rename `"inputs"` → `"connectors"` 
    - Add `"lookups"` section with `strength_adjusted_for_material` source
    - Fix output: metric `"product_strength"`, dimensions `["Product", "Material"]`, grain `["Daily"]`, celltype `"Strength"`

19. **`examples/with_data/lookups/lookups_basic/data_dummy.json`** → Replace with `raw_strength.json`
    - Columns: Product, Material, Daily, Value
    - Product: WidgetA/WidgetB/WidgetC, Material: Steel/Aluminum/Copper, Daily: Sunday-Saturday
    - All values: 1.0
    - 63 rows (3 × 3 × 7)

20. **`examples/with_data/lookups/lookups_basic/lookups_basic_product_category.json`** — Replace with `strength_adjusted_for_material.json`
    - Columns: Material, Value
    - Material: Steel→1.0, Aluminum→0.8, Copper→0.9
    - 3 rows

21. **Regenerate parquet files** from the updated JSON files

22. **`examples/with_data/lookups/lookups_basic/expected_output.xlsx`** — Regenerate after data changes (use the xls_generator tool)

23. **Delete `examples/with_data/lookups/lookups_basic/lookups_basic_lookups.io.json`** — No longer needed

### Phase 6: Expected Output Generator Extension (for lookups)

24. **`src/main/go/cli/expected_xls_generator/`**
   - Extend the evaluator to handle `CILLookupNode` in the CIL tree. When evaluating `product_strength = raw_strength * strength_adjusted_for_material`, the evaluator needs to resolve the lookup reference.
   - Read lookup data files alongside connector data files (lookup data files follow the same naming convention: `<lookup_name>.json` in the example directory)
   - Generate correct expected output values: for each (Product, Material, Daily) combination, the expected value is `connector_value * lookup_value_for_material`. Since all connector values are 1.0, the result for (WidgetX, Steel, day) = 1.0 * 1.0 = 1.0; (WidgetX, Aluminum, day) = 1.0 * 0.8 = 0.8; etc.

Note: This is specifically about the evaluator used by the expected_xls_generator (Go-based), not the general CIL evaluation in the CPL compiler/synthesis pipeline.

### Phase 7: Validation

25. Run all tests: `go test ./...`
26. Run the end-to-end lookup test
27. Run example tests for all changed examples

## Risks and Considerations

- **High blast radius**: Changing 24+ JSON files and >10 Go files. Need mechanical precision
- **Test fixture updates**: All inline JSON strings in tests need updating. A single missed one breaks tests
- **Go rename `Inputs` → `Connectors` on struct**: **Decided — rename both field and type for consistency** (`Inputs []InputSpecification` → `Connectors []ConnectorSpecification`). This is a larger diff but keeps code and JSON in sync.
- **Data alignment**: The lookups_basic example changes are independent of the io spec rename. The CPL uses connectors named `raw_strength` not `dummy`, dimensions `[Product, Material]`, and the lookup uses `[Material]`
- **Lookup adjustment factors**: **Decided — Steel=1.0, Aluminum=0.8, Copper=0.9**
- **Expected output generator extension**: **Required** — the Go evaluator in `expected_xls_generator` needs to handle `CILLookupNode` to compute expected values for metrics that reference lookups. Without this, the test infrastructure can't validate the example's output
- **Lookup Go struct in `io_specification.go`**: The `LookupSpecification` struct in `io_specification.go` replaces the `LookupEntry` in the deleted `lookup_io_spec.go` — field naming mirrors `ConnectorSpecification` (using `LookupName` + `json:"lookup"` instead of `ConnectorName` + `json:"connector"`)

## Decisions Recorded

1. **Go struct rename**: `Inputs []InputSpecification` → `Connectors []ConnectorSpecification` — consistent with JSON field names.
2. **Lookup adjustment factors**: Steel=1.0, Aluminum=0.8, Copper=0.9.
3. **Expected output generator**: Extend to handle `CILLookupNode` in CIL expression evaluation.

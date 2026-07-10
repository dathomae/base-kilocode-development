# ERO001: Concepts Synthesis Phase

## Title

Concepts Synthesis Phase - Minimal Synthesizer Impact and Version Checking

## Purpose/Goal

Document how concepts are handled in the synthesis phase. Concept definitions do not generate CIL elements. The synthesizer uses CIL produced by the analyzer to generate code. The synthesized code knows concept names and reads concept values at runtime, but has no knowledge of CIL structures. Version checking in Increment 8 ensures concept definitions haven't changed since synthesis, which could invalidate the generated code.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic001_cpl-concepts-implementation.md](../../../../../epics/Epic001_cpl-concepts-implementation.md) - Epic for CPL concepts implementation

## Context

### Situation

The synthesis phase generates executable code from CIL (Chronocone Intermediate Language) elements produced by the analyzer. The analyzer uses concept definitions (relationships and values) to:

- Validate metrics, aggregators, spreaders, and lookups
- Determine tensor shapes and array dimensions
- Compute loop bounds for generated code

Concept definitions do not generate CIL elements. The synthesized code knows concept names and reads concept values at runtime, but has no knowledge of CIL structures. The computation chain in the synthesized code depends on concept values and relationships known at analysis time.

In Increment 8 (API Integration), version information is embedded in the generated CLI to detect if concept definitions have changed since synthesis. A change in concept definitions (e.g., number of values) could invalidate the generated code.

### Constraints

- Concept definitions do not generate CIL elements
- Synthesized code knows concept names and reads concept values
- Synthesized code has no knowledge of CIL structures
- Synthesized code structure depends on concept values determined during analysis
- Increment 8 adds version checking to detect concept definition changes
- Generated CLI code must include version checking when API concepts are used

### Requirements

- Generate placeholder Go CLI code in Increment 1
- Add concept version checking to generated CLI in Increment 8
- Keep synthesizer focused on generating executable elements from CIL

### Assumptions

- Analysis phase has validated all concept definitions
- CIL captures the computation chain based on concept values known at analysis time
- Synthesized code knows concept names and reads concept values at runtime
- Synthesized code has no knowledge of CIL structures
- Concept definitions are assumed not to change between synthesis and execution
- Generated CLI validates concept versions at runtime to detect changes (Increment 8)

## Data Structures by Increment

### Increment 1: Basic Synthesis Structure

**Overview**

In Increment 1, concept definitions do not generate CIL elements. The analyzer uses concept values to validate the CPL program, but concepts themselves do not appear in the synthesized code. The synthesizer's role for Increment 1 is to generate a placeholder Go CLI file. Future increments will generate actual code from CIL elements, where the code structure reflects concept values determined during analysis.

**Synthesize Function**

The top-level `Synthesize` function takes a file name and writes out the generated Go CLI code:

```go
// Synthesize generates the Go CLI code from the analyzed CPL program.
// In Increment 1, this generates a placeholder CLI that does nothing.
// As CIL elements are added in future increments, they will be passed
// to this function for code generation.
func Synthesize(outputFile string) error {
    file, err := os.Create(outputFile)
    if err != nil {
        return fmt.Errorf("failed to create output file: %w", err)
    }
    defer file.Close()

    // Generate the Go CLI boilerplate
    if err := generateGoCLI(file); err != nil {
        return fmt.Errorf("failed to generate CLI code: %w", err)
    }

    return nil
}

// generateGoCLI writes the placeholder Go CLI code to the provided writer.
// This will be enhanced in future increments to include actual functionality.
func generateGoCLI(w io.Writer) error {
    tmpl := `package main

import (
    "fmt"
    "os"
)

func main() {
    // Placeholder CLI for Increment 1
    // Concepts have been analyzed but do not generate executable code.
    // Future increments will add actual CLI functionality.

    if len(os.Args) < 2 {
        fmt.Println("Usage: cpl-cli <command>")
        fmt.Println("  Commands: (none implemented yet)")
        os.Exit(1)
    }

    command := os.Args[1]
    _ = command // Currently no commands implemented

    fmt.Println("CPL CLI - Increment 1 (Placeholder)")
    fmt.Println("Concepts analyzed successfully, no executable code generated yet.")
}
`
    _, err := w.Write([]byte(tmpl))
    return err
}
```

**Integration with cpl_compiler.go**

```go
func (c *CPLCompiler) Synthesize(analysisDefs *AnalysisDefinitions, outputFile string) error {
    // In Increment 1, there are no CIL elements to generate code from.
    // Future increments will convert analysis definitions to CIL elements
    // and pass them to Synthesize.

    // For now, just generate the placeholder CLI
    if err := Synthesize(outputFile); err != nil {
        return fmt.Errorf("synthesis failed: %w", err)
    }

    return nil
}
```

**Limitations**

- Generates only a placeholder Go CLI that does nothing
- No CIL elements are generated in Increment 1
- No concept version checking (to be added in Increment 8)

### Increment 2: Concept Hierarchy

**Synthesizer Changes**

No changes to the synthesizer. The hierarchy information is used during analysis for validation. Future metrics/aggregators will generate CIL that depends on concept hierarchy, and the synthesizer will generate code with structure reflecting those concept relationships.

**Limitations**

- Synthesizer still generates only the placeholder CLI

### Increment 3: Multiple Concept Declaration Sections in the Same File

**Synthesizer Changes**

No changes to the synthesizer. Multi-section concept declarations are handled entirely within the analysis phase for validation and merging.

**Limitations**

- Synthesizer still generates only the placeholder CLI

### Increment 4: Conditional Specifications

**Synthesizer Changes**

Conditional specifications (child values based on parent values) are resolved in the analysis phase, but the resolved conditional value mappings MUST flow into the synthesis phase so the generated code can respect parent→child value constraints. The synthesis phase requires:

1. **Concept values embedded with conditional awareness** — `BuildConceptValuesMap` builds the concept values map from analysis definitions, using the full unified value set for each concept (shape computation). A parallel `conditionalConceptValues` map preserves parent→child value mappings from `AnalysisConceptDefinition.ConditionalValues` for mask computation.

2. **Dense tensor strategy** — Tensors for dimensions with conditional child concepts use the full cartesian product of independent value sets. For example, a connector with `grain Month, DayInMonth` where `DayInMonth` values vary by `Month` produces a tensor of shape `[len(all_months), max_len(day_values)]`.

3. **Bit-packed validity mask** — Each tensor with conditional concept dimensions carries a `[]uint64` validity mask using the same bit convention as metric conditional masks (bit 0 = valid/selected, zero-init = all-valid). The mask is computed at synthesis time from `conditionalConceptValues` and embedded as a literal in the generated Go code. Invalid cell combinations (e.g., Feb 30) are masked out; downstream mask-aware tensor operations automatically skip them.

4. **Parquet generation respects conditional constraints** — The parquet data generator produces rows only for valid (parent, child) value combinations, not the full cartesian product. For the Month/DayInMonth example, this means 365-ish rows rather than 372 (12 × 31).

5. **Parquet reader tolerates sparse row sets** — The parquet reader accepts fewer rows than the full cartesian product. Cells that exist in the tensor but are not populated from the parquet file remain zero-initialized and are masked as invalid.

**No new CIL elements are introduced.** Concept conditionals do not generate CIL nodes. The synthesis changes are limited to mask computation/embedding and parquet data generation/reading. The mask integrates with the existing `ChronoconeTensor.mask` field and mask-aware arithmetic functions from Story188 (Increment 4 of metrics synthesis).

**Limitations**

- Synthesizer still generates only the placeholder CLI for concepts (concept metadata is embedded in the generated code but concepts themselves remain handled through connectors and metrics)
- Deeply nested conditional hierarchies (grandchild values conditional on child values which are conditional on parent values) are initially supported via full-dimensional mask computation; per-parent-child resolution may need enhancement for complex nested cases

### Increment 5: Int Range Generators

**Synthesizer Changes**

No changes to the synthesizer. Int Range generators are handled entirely within the analysis phase.

**Limitations**

- Synthesizer still generates only the placeholder CLI
- No Concept Range generators (added in Increment 6)
- No before/after clauses for value ordering (added in Increment 7)

### Increment 6: Concept Range Generators

**Synthesizer Changes**

No changes to the synthesizer. Concept Range generators are handled entirely within the analysis phase.

**Limitations**

- Synthesizer still generates only the placeholder CLI
- No before/after clauses for value ordering (added in Increment 7)

### Increment 7: Before/After Clauses for Value Ordering

**Synthesizer Changes**

No changes to the synthesizer. Before/after clauses for value ordering are handled entirely within the analysis phase.

**Limitations**

- Synthesizer still generates only the placeholder CLI

## Options Considered

### Option 1: Generate Concept Definition Code

**Description:**
Generate Go code for concept definitions (metadata, hierarchy structures) that would be available at runtime.

**Pros:**

- Concept metadata available at runtime
- Could support dynamic concept inspection

**Cons:**

- Concept definitions (metadata) are used by the analyzer, not the synthesized code at runtime
- The synthesized code only needs concept names and values, not full definition metadata
- Unnecessary complexity

**Decision:** Rejected. Concept definitions (metadata) are used by the analyzer; the synthesized code only needs concept names and values.

### Option 2: Concepts Known to Synthesized Code

**Description:**
Concept definitions do not generate CIL elements. The analyzer uses concept values to generate appropriate CIL, and the synthesizer generates code from that CIL. The synthesized code knows concept names and reads concept values at runtime, but has no knowledge of CIL structures. The synthesized code structure depends on concept values (e.g., loop bounds, array sizes) known at analysis time.

**Pros:**

- Clean separation: analyzer deals with concept definitions, synthesizer deals with CIL
- Synthesized code can reference concepts by name and read their values
- Version checking detects concept definition changes that would invalidate the generated code

**Cons:**

- API concept integration requires version checking to detect definition changes

**Decision:** Selected. The analyzer uses concept values to produce CIL. The synthesizer generates code from CIL. The synthesized code knows concept names and reads concept values, but has no knowledge of CIL structures. Version checking (Increment 8) ensures concept definitions haven't changed since synthesis.

## Decision

**Chosen Option:** Concepts Known to Synthesized Code (with version checking)

**Rationale:**
Concepts define relationships and values that are used by the analyzer to validate CPL programs and generate appropriate CIL. The CIL captures the computation chain with specific array dimensions, loop bounds, etc. based on concept values known at analysis time. The synthesizer generates executable code from CIL—the synthesized code knows concept names and reads concept values at runtime, but has no knowledge of CIL structures. In Increment 8, version checking detects if concept definitions have changed since synthesis, which could invalidate the generated code.

## Implications

### Positive Implications

- Clean separation between analysis (concept definitions) and synthesis (CIL)
- Synthesized code knows concept names and can read concept values
- Version checking ensures concept consistency when using API concepts
- Generated code structure reflects analysis-time concept values

### Negative Implications

- API concept version checking adds complexity to Increment 8
- If concept definitions change after synthesis, the executable may fail version checks
- Cannot dynamically adapt to concept definition changes at runtime

### Code Generation Strategy

Concept definitions do not generate code in Increments 1-7. The synthesizer generates code from CIL elements produced by the analyzer:

1. **Increment 1:** Placeholder Go CLI that does nothing (no CIL yet)
2. **Increment 2:** No synthesizer changes (hierarchy used in analysis only)
3. **Increment 3:** No synthesizer changes (merging handled in analysis)
4. **Increment 4:** No synthesizer changes (conditional specs handled in analysis)
5. **Increment 5:** No synthesizer changes (Int Range generators handled in analysis)
6. **Increment 6:** No synthesizer changes (Concept Range generators handled in analysis)
7. **Increment 7:** No synthesizer changes (value ordering handled in analysis)
8. **Increment 8:** Add version checking to detect concept definition changes

### Increment 8: Concepts API Integration

**Synthesizer Changes**

Increment 8 adds version checking to the generated CLI. The analyzer loads API concepts and uses them to generate CIL. The synthesizer embeds concept version information in the generated code. At runtime, this validates that concept definitions haven't changed since synthesis—changes could invalidate the generated code structure (e.g., array sizes, loop bounds that were based on concept values).

**Version Checking in Synthesis**

```go
// SynthesizeWithVersionChecking generates the CLI code with embedded
// concept version information for runtime validation.
// This is the only concept-related addition to the synthesizer.
func SynthesizeWithVersionChecking(
    analysisDefs *AnalysisDefinitions,
    outputFile string) error {

    // Build version manifest for API concepts
    versionManifest := buildVersionManifest(analysisDefs)

    // Generate the CLI code with version checking
    if err := generateGoCLIWithVersionCheck(outputFile, versionManifest); err != nil {
        return fmt.Errorf("failed to generate CLI with version check: %w", err)
    }

    return nil
}

// buildVersionManifest extracts version information for all concepts.
// The version hash captures the concept state at synthesis time.
func buildVersionManifest(analysisDefs *AnalysisDefinitions) *VersionManifest {
    manifest := &VersionManifest{
        GeneratedAt:     time.Now(),
        ConceptVersions: make(map[string]ConceptVersionInfo),
    }

    for name, info := range analysisDefs.Concepts {
        manifest.ConceptVersions[name] = ConceptVersionInfo{
            Name:      name,
            ValueHash: computeValueHash(info.Values),
        }
    }

    return manifest
}
```

**Generated Version Checking Code**

The synthesizer generates CLI code that includes version checking for API concepts:

```go
package main

import (
    "fmt"
    "os"
    "time"
)

// Embedded version manifest for runtime validation
var ConceptVersionManifest = &VersionManifest{
    GeneratedAt: time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC),
    ConceptVersions: map[string]ConceptVersionInfo{
        "Country": {
            Name:      "Country",
            ValueHash: "sha256:a1b2c3d4...",
        },
        // ... other concepts
    },
}

// VersionChecker validates concept consistency at runtime
type VersionChecker struct {
    apiClient ConceptsAPIClient
}

func (vc *VersionChecker) ValidateConceptVersions() error {
    for name, expectedVersion := range ConceptVersionManifest.ConceptVersions {
        // Query API for current concept definition
        currentConcept, err := vc.apiClient.GetConcept(name)
        if err != nil {
            return fmt.Errorf("failed to fetch concept %s: %w", name, err)
        }

        // Compute hash of current values
        currentHash := computeValueHash(currentConcept.Values)

        // Compare with expected hash from synthesis time
        if currentHash != expectedVersion.ValueHash {
            return fmt.Errorf(
                "concept %s definition has changed since synthesis: hash mismatch",
                name,
            )
        }
    }
    return nil
}

func main() {
    // Validate concept definitions haven't changed since synthesis
    checker := &VersionChecker{apiClient: getAPIClient()}
    if err := checker.ValidateConceptVersions(); err != nil {
        fmt.Fprintf(os.Stderr, "Concept validation failed: %v\n", err)
        fmt.Fprintln(os.Stderr, "Concept definitions have changed since this executable was generated.")
        os.Exit(1)
    }

    // Continue with normal CLI execution
    // ...
}
```

**Note:** Version checking is the only synthesizer addition specifically for concepts. The analyzer uses concept values to produce CIL with appropriate structure (array sizes, loop bounds). The synthesizer generates executable code from CIL—the synthesized code knows concept names and reads concept values at runtime, but has no knowledge of CIL structures. All concept definition processing (hierarchy, conditional values, range generators, before/after clauses, etc.) is handled within the analysis phase.

## Notes

**Related Documents:**

- [Concepts Incremental Implementation](../../eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
- [Concept Definitions Design](../../concepts-definitions.md) - Full concept design
- [Synthesis Design](../synthesis.md) - General synthesis phase design

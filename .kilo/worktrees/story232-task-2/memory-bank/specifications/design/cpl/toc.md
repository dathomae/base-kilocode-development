# CPL Design - Table of Contents

This directory contains design documents specific to the Chronocone Planning Language (CPL) implementation.

The general organization of CPL includes three phases:

1. **Parse** - Convert CPL text into an Abstract Syntax Tree (AST)
2. **Analyze** - Validate the AST and perform semantic analysis
3. **Synthesize** - Generate the final output (e.g., plan data, dimensions)

Each phase has a sub-directory that captures design documentation and EROs for that phase. In most cases the design documents in the same directory as this toc.md file will have implications for two or more phases.

## Files

**[concepts-definitions.md](concepts-definitions.md)**
: Design for concept definitions across CPL phases.

**[types-definitions.md](types-definitions.md)**
: Design for type definitions across CPL phases.

**[units-and-measures.md](units-and-measures.md)**
: Design for unit and measure definitions across CPL phases.

**[connectors.md](connectors.md)**
: Design for connector definitions across CPL phases.

**[lookups.md](lookups.md)**
: Design for lookup definitions across CPL phases.

**[aggregators.md](aggregators.md)**
: Design for aggregator definitions across CPL phases.

**[spreaders.md](spreaders.md)**
: Design for spreader definitions across CPL phases.

**[metrics-design.md](metrics-design.md)**
: Design for metric definitions across CPL phases.

**[eros/toc.md](eros/toc.md)**
: Epic Restraining Orders (EROs) for CPL design.

**[go-parser-target/toc.md](go-parser-target/toc.md)**
: Go parser target design documents (ANTLR grammar actions).

**[go-analysis/toc.md](go-analysis/toc.md)**
: Go analysis target design documents (post-parsing analysis).

**[go-synthesis-target/toc.md](go-synthesis-target/toc.md)**
: Go synthesis target design documents (plan data/dimensions).

**[example_generation_and_verification.md](example_generation_and_verification.md)**
: ODS-based example generation and verification system, including ODS generator
  and verifier designs.

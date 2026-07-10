# CPL Architecture - Table of Contents

This directory contains architectural documentation for the Chronocone Planning Language (CPL) processing pipeline.

**[introduction.md](introduction.md)**
: Overview of the four-phase CPL lifecycle: parsing, analysis, synthesis, and evaluation.

**[parsing/](parsing/)**
: Syntax analysis phase - parsing `.cpl` files and building the AST.

**[analysis/](analysis/)**
: Semantic analysis phase - validation, AnalysisDefinitions assembly, and building the Data/Control Flow Graph.

**[synthesis/](synthesis/)**
: Code generation phase - transforming CILDefinition trees to executable form.

**[evaluation/](evaluation/)**
: Execution phase - running the synthesized code to produce results.

# Analysis Features

This section covers code analysis features including syntax errors, annotators, inspections, and structural search.

## Topics

**[syntax-errors.md](syntax-errors.md)**
: Displaying syntax errors in the editor

**[annotators.md](annotators.md)**
: Creating custom annotators for code analysis

**[inspections.md](inspections.md)**
: Implementing code inspections

**[global-inspections.md](global-inspections.md)**
: Project-wide analysis inspections

**[inspection-profiles.md](inspection-profiles.md)**
: Managing inspection profiles

**[structural-search.md](structural-search.md)**
: Structural search and replace

## Key Concepts

### Code Analysis Levels

| Level             | Scope        | Purpose               |
| ----------------- | ------------ | --------------------- |
| Syntax            | Single file  | Parse errors          |
| Annotator         | Single file  | Semantic highlighting |
| Local Inspection  | Single file  | Code quality checks   |
| Global Inspection | Project-wide | Cross-file analysis   |

### Analysis Timing

- **Syntax errors**: Immediate, during parsing
- **Annotators**: During highlighting pass
- **Local inspections**: During code analysis pass
- **Global inspections**: On-demand or during full project analysis

## Recommended Reading Order

1. Start with [syntax-errors.md](syntax-errors.md) for basic error display
2. Read [annotators.md](annotators.md) for semantic analysis
3. Study [inspections.md](inspections.md) for code quality checks
4. Explore [global-inspections.md](global-inspections.md) for project-wide analysis
5. Learn [inspection-profiles.md](inspection-profiles.md) for configuration
6. Reference [structural-search.md](structural-search.md) for pattern matching

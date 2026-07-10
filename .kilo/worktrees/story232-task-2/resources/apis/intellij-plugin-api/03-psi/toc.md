# Program Structure Interface (PSI)

The Program Structure Interface (PSI) is the IntelliJ Platform's abstract syntax tree representation of source code. It provides a language-agnostic API for parsing, navigating, and modifying code.

## Topics

**[what-is-psi.md](what-is-psi.md)**
: Introduction to PSI and its role in the platform

**[psi-files.md](psi-files.md)**
: Working with PSI file representations

**[file-view-providers.md](file-view-providers.md)**
: Handling files with multiple languages

**[psi-elements.md](psi-elements.md)**
: Understanding PSI element types

**[navigating-psi.md](navigating-psi.md)**
: Traversing and searching PSI trees

**[psi-references.md](psi-references.md)**
: Working with symbol references

**[modifying-psi.md](modifying-psi.md)**
: Modifying PSI trees safely

**[psi-cookbook.md](psi-cookbook.md)**
: Common PSI patterns and recipes

**[psi-performance.md](psi-performance.md)**
: Performance optimization for PSI operations

## Key Concepts

### PSI Tree

The PSI tree represents the syntactic structure of source code. Each node in the tree is a PSI element representing a language construct like a class, method, statement, or expression.

### Language Independence

PSI provides a unified API across languages. While each language has specific element types, common operations work consistently:

- Finding parent or child elements
- Searching for specific element types
- Resolving references
- Modifying code structure

### View Providers

Files can contain multiple languages (e.g., HTML with embedded JavaScript). View providers manage multiple PSI trees for a single file.

## Recommended Reading Order

1. Start with [what-is-psi.md](what-is-psi.md) for a conceptual overview
2. Read [psi-files.md](psi-files.md) to understand file-level PSI
3. Study [psi-elements.md](psi-elements.md) for element types
4. Learn [navigating-psi.md](navigating-psi.md) for traversal patterns
5. Explore [psi-references.md](psi-references.md) for reference resolution
6. Read [modifying-psi.md](modifying-psi.md) for code modifications
7. Reference [psi-cookbook.md](psi-cookbook.md) for common patterns
8. Study [psi-performance.md](psi-performance.md) for optimization

## When to Use PSI

Use PSI when you need to:

- Parse or analyze source code
- Navigate code structure
- Find symbol definitions or references
- Modify source code programmatically
- Implement refactoring operations
- Create code inspections or intentions

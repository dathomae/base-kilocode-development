# Language Support

This section covers creating custom language support in IntelliJ Platform plugins, from basic file type registration to advanced features like refactoring.

## Topics

**[language-registration.md](language-registration.md)**
: Registering a new language and file type

**[lexer-and-parser.md](lexer-and-parser.md)**
: Implementing lexers and parsers

**[syntax-highlighting.md](syntax-highlighting.md)**
: Syntax highlighting and color schemes

**[code-completion-lang.md](code-completion-lang.md)**
: Language-specific code completion

**[reference-resolution.md](reference-resolution.md)**
: Resolving symbol references

**[find-usages.md](find-usages.md)**
: Find usages functionality

**[rename-refactoring.md](rename-refactoring.md)**
: Rename refactoring support

**[safe-delete.md](safe-delete.md)**
: Safe delete refactoring

**[code-formatting-lang.md](code-formatting-lang.md)**
: Language-specific formatting

**[language-server-protocol.md](language-server-protocol.md)**
: LSP integration

## Key Concepts

### Language Definition

A language is defined by:

- **Language ID**: Unique identifier
- **File Type**: Associated file extensions and icons
- **Parser**: Converts text to PSI tree
- **Syntax Highlighter**: Colors tokens and elements

### PSI for Custom Languages

Custom languages need:

- **AST Node Types**: Token and element types
- **PSI Elements**: Java/Kotlin classes for elements
- **Parser Definition**: Builds PSI from tokens

### Feature Implementation

Language features are implemented through extension points:

- Completion contributors
- Reference contributors
- Inspection tools
- Refactoring handlers

## Recommended Reading Order

1. Start with [language-registration.md](language-registration.md) to define your language
2. Implement [lexer-and-parser.md](lexer-and-parser.md) for parsing
3. Add [syntax-highlighting.md](syntax-highlighting.md) for visual feedback
4. Implement [reference-resolution.md](reference-resolution.md) for navigation
5. Add [code-completion-lang.md](code-completion-lang.md) for completion
6. Implement [find-usages.md](find-usages.md) and [rename-refactoring.md](rename-refactoring.md) for refactoring

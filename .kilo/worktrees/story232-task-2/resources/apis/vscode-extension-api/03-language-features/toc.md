# Language Features

This section covers the programmatic language features that extensions can provide to enhance the editing experience for programming languages.

## Table of Contents

**[overview.md](overview.md)**
: Introduction to language features and provider patterns

**[completion.md](completion.md)**
: Providing code completion suggestions

**[hover.md](hover.md)**
: Showing hover information for symbols

**[diagnostics.md](diagnostics.md)**
: Providing error and warning diagnostics

**[definition.md](definition.md)**
: Implementing go-to-definition functionality

**[references.md](references.md)**
: Finding all references to a symbol

**[symbols.md](symbols.md)**
: Providing document and workspace symbols

## Quick Reference

### Registering a Completion Provider

```typescript
vscode.languages.registerCompletionItemProvider(
  { language: "javascript" },
  new MyCompletionProvider(),
  ".",
);
```

### Registering a Hover Provider

```typescript
vscode.languages.registerHoverProvider(
  { language: "python" },
  new MyHoverProvider(),
);
```

### Creating Diagnostics

```typescript
const diagnostic = new vscode.Diagnostic(
  range,
  "Error message",
  vscode.DiagnosticSeverity.Error,
);
diagnosticCollection.set(document.uri, [diagnostic]);
```

### Provider Pattern

Most language features use the provider pattern:

```typescript
class MyProvider implements vscode.SomeProvider {
  provideSomething(document, position, token): ProviderResult<Result> {
    // Return results or undefined
  }
}
```

## Feature Categories

### Editor Intelligence

Features that help users understand and navigate code:

- **Completion**: Suggest code as users type
- **Hover**: Show documentation on hover
- **Signature Help**: Display function signatures
- **Definition**: Navigate to definitions
- **References**: Find all usages
- **Document Highlights**: Highlight matching symbols

### Code Analysis

Features that analyze code for issues:

- **Diagnostics**: Show errors, warnings, and hints
- **Code Actions**: Provide quick fixes
- **Code Lens**: Show actionable context information

### Code Structure

Features that help understand code organization:

- **Document Symbols**: Outline of file structure
- **Workspace Symbols**: Search symbols across workspace
- **Folding Ranges**: Define collapsible regions
- **Document Links**: Make URLs clickable

### Code Editing

Features that assist with code modification:

- **Formatting**: Auto-format code
- **Rename**: Rename symbols across files
- **On Type Formatting**: Format as you type

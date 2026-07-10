# Language Features Overview

VSCode provides a rich set of language features that extensions can implement to enhance the editing experience. These features include code completion, hover information, diagnostics, navigation, and more.

## The Provider Pattern

Most language features use the provider pattern. A provider is an object that VSCode calls when it needs information about a document or position. Providers implement a specific interface and are registered for particular languages.

### Basic Provider Structure

```typescript
class MyProvider implements vscode.SomeProvider {
  provideSomething(
    document: vscode.TextDocument,
    position: vscode.Position,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<SomeResult> {
    // Return the requested information
    return result;
  }
}
```

### Provider Registration

Providers are registered using `vscode.languages.register*` functions:

```typescript
const disposable = vscode.languages.registerCompletionItemProvider(
  { language: "javascript" },
  new MyCompletionProvider(),
  ".",
);

context.subscriptions.push(disposable);
```

### Language Selectors

Providers are registered for specific languages using selectors:

```typescript
// Single language
{ language: 'javascript' }

// Multiple languages
{ language: ['javascript', 'typescript'] }

// Specific file pattern
{ pattern: '**/*.js' }

// Specific scheme
{ scheme: 'file', language: 'javascript' }

// All files
'*'
```

### Cancellation Tokens

Providers receive a `CancellationToken` that indicates if the request was cancelled:

```typescript
provideSomething(document, position, token): vscode.ProviderResult<Result> {
    if (token.isCancellationRequested) {
        return undefined;
    }

    // Do work, checking periodically
    for (const item of items) {
        if (token.isCancellationRequested) {
            return undefined;
        }
        process(item);
    }

    return results;
}
```

### Provider Results

Providers return `ProviderResult<T>`, which can be:

- The result type `T`
- `undefined` (no results)
- `Thenable<T>` (async result)

```typescript
// Synchronous result
return items;

// No results
return undefined;

// Asynchronous result
return asyncOperation().then((result) => result);
```

## Feature Categories

### Completion

Code completion suggests items as users type. Completion providers return `CompletionItem` objects that appear in the IntelliSense dropdown.

**When it triggers**: User types or explicitly invokes completion (Ctrl+Space)

**What to provide**: Variable names, function names, keywords, snippets, method suggestions

See [completion.md](completion.md) for details.

### Hover

Hover providers show documentation when the user hovers over a symbol.

**When it triggers**: User hovers over a position in the editor

**What to provide**: Type information, documentation, related code snippets

See [hover.md](hover.md) for details.

### Diagnostics

Diagnostics indicate problems in code, such as syntax errors, type errors, or linting issues.

**When it triggers**: Continuously as documents change

**What to provide**: Errors, warnings, information, and hints with optional fixes

See [diagnostics.md](diagnostics.md) for details.

### Definition

Definition providers enable the "Go to Definition" feature.

**When it triggers**: User invokes "Go to Definition" (F12) or clicks a definition link

**What to provide**: The location where a symbol is defined

See [definition.md](definition.md) for details.

### References

Reference providers find all references to a symbol across the workspace.

**When it triggers**: User invokes "Find All References" (Shift+F12)

**What to provide**: Locations where the symbol is referenced

See [references.md](references.md) for details.

### Document Symbols

Document symbol providers return the structure of a document, shown in the outline view and breadcrumb navigation.

**When it triggers**: User opens the outline view or breadcrumb navigation

**What to provide**: Classes, functions, variables, and other structural elements

See [symbols.md](symbols.md) for details.

### Workspace Symbols

Workspace symbol providers enable searching for symbols across the entire workspace.

**When it triggers**: User invokes "Go to Symbol in Workspace" (Ctrl+T)

**What to provide**: Symbols matching the search query

See [symbols.md](symbols.md) for details.

### Signature Help

Signature help providers show function signatures and parameter information.

**When it triggers**: User types a function call (typically after `(` or `,`)

**What to provide**: Function signatures with parameter documentation

### Document Highlights

Document highlight providers highlight all occurrences of the symbol under the cursor.

**When it triggers**: Cursor moves over a symbol

**What to provide**: Ranges where the same symbol appears

### Code Actions

Code action providers offer quick fixes and refactorings for diagnostics.

**When it triggers**: User opens the lightbulb menu or quick fix panel

**What to provide**: Actions to fix problems or refactor code

### Code Lens

Code lens providers show actionable, contextual information inline in the editor.

**When it triggers**: Document opens or changes

**What to provide**: Commands displayed above lines of code

### Formatting

Formatting providers format documents or selections according to style rules.

**When it triggers**: User invokes "Format Document" or "Format Selection"

**What to provide**: Text edits that format the code

### Rename

Rename providers enable renaming symbols across all references.

**When it triggers**: User invokes "Rename Symbol" (F2)

**What to provide**: All locations where the symbol should be renamed

### Folding Ranges

Folding range providers define collapsible regions in documents.

**When it triggers**: Document opens or changes

**What to provide**: Ranges that can be folded

### Document Links

Document link providers make URLs and file paths clickable.

**When it triggers**: Document opens or changes

**What to provide**: Links that open URLs or files

### Document Color

Document color providers find color references and enable color pickers.

**When it triggers**: Document opens or changes

**What to provide**: Color ranges and their values

### Selection Ranges

Selection range providers enable smart selection expansion.

**When it triggers**: User invokes "Expand Selection" or "Shrink Selection"

**What to provide**: Hierarchical selection ranges

### Linked Editing Ranges

Linked editing range providers enable simultaneous editing of related ranges.

**When it triggers**: User types in a linked region (like HTML tags)

**What to provide**: Ranges that should be edited together

### Inlay Hints

Inlay hint providers show additional information inline in the editor.

**When it triggers**: Document opens or changes

**What to provide**: Type annotations, parameter names, etc.

### Type Hierarchy

Type hierarchy providers show type inheritance relationships.

**When it triggers**: User invokes "Show Type Hierarchy"

**What to provide**: Supertypes and subtypes of a symbol

## Implementation Considerations

### Performance

Language features must be fast to maintain editor responsiveness:

**Return quickly**: If computation is expensive, return partial results or use caching.

**Use cancellation**: Check the cancellation token frequently and abort if requested.

**Cache results**: Cache parsed documents or analysis results between calls.

**Debounce expensive operations**: For diagnostics, consider debouncing to avoid recomputing on every keystroke.

### Language Server Protocol

For complex language support, consider implementing a Language Server instead of direct providers:

- **Language Server Protocol (LSP)**: A standard protocol for language features
- **Separate process**: Runs in a separate process, isolating the extension host
- **Rich features**: Supports all language features plus advanced capabilities
- **Multi-editor support**: Same server works in VSCode, Vim, Emacs, and other editors

See the [Language Server Protocol documentation](https://microsoft.github.io/language-server-protocol/) for more information.

### Multiple Providers

Multiple extensions can provide the same feature for the same language. VSCode merges results from all providers. Ensure your provider:

- Returns only relevant results
- Does not duplicate items from other providers
- Handles cases where another provider's results are preferred

## Best Practices

**Implement only needed features**: Not every language needs all features. Start with the most valuable ones.

**Provide accurate results**: Incorrect suggestions or definitions frustrate users. Prefer fewer, accurate results over many, inaccurate ones.

**Handle edge cases**: Consider empty documents, invalid syntax, and unusual file contents.

**Test thoroughly**: Test with real-world code, not just simple examples.

**Document limitations**: If your provider has known limitations, document them clearly.

## Next Steps

- [Completion](completion.md): Implement code completion
- [Hover](hover.md): Show hover information
- [Diagnostics](diagnostics.md): Provide error and warning diagnostics
- [Definition](definition.md): Enable go-to-definition

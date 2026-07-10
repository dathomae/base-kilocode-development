# Document and Workspace Symbols

Symbol providers enable navigation and search within documents and across the workspace. Document symbols appear in the outline view and breadcrumb navigation, while workspace symbols enable searching across all files.

## Document Symbols

Document symbols provide a hierarchical view of the structure within a single file.

### Registering a Document Symbol Provider

Use `vscode.languages.registerDocumentSymbolProvider`:

```typescript
import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  const provider = vscode.languages.registerDocumentSymbolProvider(
    { language: "javascript" },
    new MyDocumentSymbolProvider(),
  );

  context.subscriptions.push(provider);
}
```

### Implementing a Document Symbol Provider

Implement the `DocumentSymbolProvider` interface:

```typescript
class MyDocumentSymbolProvider implements vscode.DocumentSymbolProvider {
  provideDocumentSymbols(
    document: vscode.TextDocument,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<
    vscode.SymbolInformation[] | vscode.DocumentSymbol[]
  > {
    const symbols: vscode.DocumentSymbol[] = [];
    const text = document.getText();

    // Find all function declarations
    const funcPattern = /function\s+(\w+)\s*\(/g;
    let match;

    while ((match = funcPattern.exec(text))) {
      const name = match[1];
      const startPos = document.positionAt(match.index);
      const endPos = findFunctionEnd(text, match.index);

      const symbol = new vscode.DocumentSymbol(
        name,
        "function",
        vscode.SymbolKind.Function,
        new vscode.Range(startPos, endPos),
        new vscode.Range(startPos, startPos.translate(0, name.length)),
      );

      symbols.push(symbol);
    }

    return symbols;
  }
}
```

### DocumentSymbol vs SymbolInformation

**DocumentSymbol**: Provides hierarchical symbols with children:

```typescript
const symbol = new vscode.DocumentSymbol(
  name, // Symbol name
  detail, // Additional detail
  kind, // SymbolKind
  range, // Full range of the symbol
  selectionRange, // Range to select when navigating
);

// Add child symbols
symbol.children = [childSymbol1, childSymbol2];
```

**SymbolInformation**: Flat list of symbols (legacy):

```typescript
const symbol = new vscode.SymbolInformation(
  name, // Symbol name
  kind, // SymbolKind
  containerName, // Parent symbol name
  location, // Location of the symbol
);
```

Use `DocumentSymbol` for new implementations as it supports hierarchy.

### Symbol Kinds

The `SymbolKind` enum determines the icon shown in the outline:

| Kind            | Use Case            |
| --------------- | ------------------- |
| `File`          | Files               |
| `Module`        | Modules, namespaces |
| `Namespace`     | Namespaces          |
| `Package`       | Packages            |
| `Class`         | Classes             |
| `Method`        | Methods             |
| `Property`      | Properties          |
| `Field`         | Fields              |
| `Constructor`   | Constructors        |
| `Enum`          | Enumerations        |
| `Interface`     | Interfaces          |
| `Function`      | Functions           |
| `Variable`      | Variables           |
| `Constant`      | Constants           |
| `String`        | Strings             |
| `Number`        | Numbers             |
| `Boolean`       | Booleans            |
| `Array`         | Arrays              |
| `Object`        | Objects             |
| `Key`           | Keys                |
| `Null`          | Null values         |
| `EnumMember`    | Enum members        |
| `Struct`        | Structs             |
| `Event`         | Events              |
| `Operator`      | Operators           |
| `TypeParameter` | Type parameters     |

### Hierarchical Symbols

Create nested symbol structures:

```typescript
provideDocumentSymbols(document, token) {
    const symbols: vscode.DocumentSymbol[] = [];

    // Find classes
    const classPattern = /class\s+(\w+)/g;
    let match;

    while ((match = classPattern.exec(text))) {
        const classSymbol = createClassSymbol(document, match);

        // Find methods within the class
        const classBody = getClassBody(text, match.index);
        const methods = findMethods(classBody);

        for (const method of methods) {
            const methodSymbol = createMethodSymbol(document, method);
            classSymbol.children.push(methodSymbol);
        }

        symbols.push(classSymbol);
    }

    return symbols;
}
```

### Symbol Tags

Add tags to symbols for special rendering:

```typescript
symbol.tags = [vscode.SymbolTag.Deprecated];
```

Available tags:

- `Deprecated`: Rendered with strikethrough

## Workspace Symbols

Workspace symbols enable searching across all files in the workspace.

### Registering a Workspace Symbol Provider

Use `vscode.languages.registerWorkspaceSymbolProvider`:

```typescript
export function activate(context: vscode.ExtensionContext) {
  const provider = vscode.languages.registerWorkspaceSymbolProvider(
    new MyWorkspaceSymbolProvider(),
  );

  context.subscriptions.push(provider);
}
```

### Implementing a Workspace Symbol Provider

Implement the `WorkspaceSymbolProvider` interface:

```typescript
class MyWorkspaceSymbolProvider implements vscode.WorkspaceSymbolProvider {
  provideWorkspaceSymbols(
    query: string,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<vscode.SymbolInformation[]> {
    const symbols: vscode.SymbolInformation[] = [];

    // Search all files for matching symbols
    const pattern = new RegExp(`\\b${escapeRegex(query)}\\w*\\b`, "gi");

    // Implementation depends on your symbol index
    for (const symbol of this.symbolIndex) {
      if (pattern.test(symbol.name)) {
        symbols.push(
          new vscode.SymbolInformation(
            symbol.name,
            symbol.kind,
            symbol.containerName,
            symbol.location,
          ),
        );
      }
    }

    return symbols;
  }
}
```

### Symbol Information for Workspace

Workspace symbols use `SymbolInformation`:

```typescript
const symbol = new vscode.SymbolInformation(
  name, // Symbol name
  kind, // SymbolKind
  containerName, // Parent symbol name (for hierarchy display)
  new vscode.Location(uri, range), // Location
);
```

### Building a Symbol Index

For performance, build and maintain a symbol index:

```typescript
class WorkspaceSymbolProvider implements vscode.WorkspaceSymbolProvider {
  private symbols: Map<string, SymbolInfo[]> = new Map();

  constructor() {
    this.buildIndex();

    // Update index on file changes
    vscode.workspace.onDidChangeTextDocument((event) => {
      this.updateIndex(event.document);
    });

    vscode.workspace.onDidCreateFiles((event) => {
      event.files.forEach((uri) => this.indexFile(uri));
    });

    vscode.workspace.onDidDeleteFiles((event) => {
      event.files.forEach((uri) => this.removeFromFile(uri));
    });
  }

  private async buildIndex() {
    const files = await vscode.workspace.findFiles("**/*.js");
    for (const file of files) {
      await this.indexFile(file);
    }
  }

  private async indexFile(uri: vscode.Uri) {
    const document = await vscode.workspace.openTextDocument(uri);
    const symbols = this.extractSymbols(document);
    this.symbols.set(uri.toString(), symbols);
  }

  provideWorkspaceSymbols(query: string, token: vscode.CancellationToken) {
    const results: vscode.SymbolInformation[] = [];
    const lowerQuery = query.toLowerCase();

    for (const [uri, symbols] of this.symbols) {
      for (const symbol of symbols) {
        if (symbol.name.toLowerCase().includes(lowerQuery)) {
          results.push(
            new vscode.SymbolInformation(
              symbol.name,
              symbol.kind,
              symbol.containerName,
              new vscode.Location(vscode.Uri.parse(uri), symbol.range),
            ),
          );
        }
      }
    }

    return results;
  }
}
```

### Resolving Workspace Symbols

For symbols that require additional information, implement `resolveWorkspaceSymbol`:

```typescript
class MyWorkspaceSymbolProvider implements vscode.WorkspaceSymbolProvider {
  provideWorkspaceSymbols(query: string, token: vscode.CancellationToken) {
    // Return symbols with minimal information
    return symbols;
  }

  resolveWorkspaceSymbol(
    symbol: vscode.SymbolInformation,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<vscode.SymbolInformation> {
    // Add additional information (like full range)
    const fullSymbol = this.loadFullSymbol(symbol);
    return fullSymbol;
  }
}
```

## Symbol Search Patterns

### Fuzzy Matching

Implement fuzzy matching for better search results:

```typescript
provideWorkspaceSymbols(query: string, token: vscode.CancellationToken) {
    const results: vscode.SymbolInformation[] = [];

    for (const symbol of this.allSymbols) {
        const score = fuzzyMatch(query, symbol.name);
        if (score > 0) {
            const info = new vscode.SymbolInformation(
                symbol.name,
                symbol.kind,
                symbol.containerName,
                symbol.location
            );
            results.push(info);
        }
    }

    // Sort by relevance
    results.sort((a, b) => {
        const scoreA = fuzzyMatch(query, a.name);
        const scoreB = fuzzyMatch(query, b.name);
        return scoreB - scoreA;
    });

    return results;
}
```

### Scoped Search

Limit search to specific contexts:

```typescript
provideWorkspaceSymbols(query: string, token: vscode.CancellationToken) {
    // Parse query for scope hints
    const match = query.match(/^(\w+):(.*)$/);

    if (match) {
        const [, scope, searchTerm] = match;

        switch (scope) {
            case 'class':
                return this.searchClasses(searchTerm);
            case 'func':
                return this.searchFunctions(searchTerm);
            case 'var':
                return this.searchVariables(searchTerm);
        }
    }

    return this.searchAll(query);
}
```

## Best Practices

**Return quickly**: Symbol providers should be fast. Use indexing for large workspaces.

**Handle cancellation**: Check the cancellation token during long operations.

**Use appropriate kinds**: Set the correct `SymbolKind` for proper icons.

**Build hierarchy**: Use `DocumentSymbol` with children for nested structures.

**Provide container names**: Set `containerName` for workspace symbols to show context.

**Update incrementally**: Update your index when files change rather than rebuilding.

**Limit results**: Return a reasonable number of results (e.g., 100) for workspace searches.

## Common Patterns

### JavaScript/TypeScript Symbols

```typescript
provideDocumentSymbols(document, token) {
    const symbols: vscode.DocumentSymbol[] = [];
    const text = document.getText();

    // Classes
    const classPattern = /class\s+(\w+)(?:\s+extends\s+\w+)?\s*\{/g;
    let match;

    while ((match = classPattern.exec(text))) {
        const classSymbol = new vscode.DocumentSymbol(
            match[1],
            '',
            vscode.SymbolKind.Class,
            getRange(document, match.index, getClassEnd(text, match.index)),
            getRange(document, match.index + 6, match[1].length)
        );

        // Add methods and properties
        const classBody = getClassBody(text, match.index);
        classSymbol.children = extractClassMembers(document, classBody);

        symbols.push(classSymbol);
    }

    // Functions
    const funcPattern = /(?:function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[a-zA-Z])\s*=>)/g;
    while ((match = funcPattern.exec(text))) {
        const name = match[1] || match[2];
        symbols.push(new vscode.DocumentSymbol(
            name,
            '',
            vscode.SymbolKind.Function,
            getRange(document, match.index, getFunctionEnd(text, match.index)),
            getRange(document, match.index, name.length)
        ));
    }

    return symbols;
}
```

### Python Symbols

```typescript
provideDocumentSymbols(document, token) {
    const symbols: vscode.DocumentSymbol[] = [];
    const lines = document.getText().split('\n');

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];

        // Classes
        const classMatch = line.match(/^class\s+(\w+)/);
        if (classMatch) {
            const classSymbol = new vscode.DocumentSymbol(
                classMatch[1],
                '',
                vscode.SymbolKind.Class,
                new vscode.Range(i, 0, getClassEndLine(lines, i), lines[getClassEndLine(lines, i)].length),
                new vscode.Range(i, 6, i, 6 + classMatch[1].length)
            );

            // Find methods
            classSymbol.children = findMethodsInClass(lines, i);
            symbols.push(classSymbol);
        }

        // Top-level functions
        const funcMatch = line.match(/^def\s+(\w+)/);
        if (funcMatch && !line.startsWith(' ')) {
            symbols.push(new vscode.DocumentSymbol(
                funcMatch[1],
                '',
                vscode.SymbolKind.Function,
                new vscode.Range(i, 0, getFunctionEndLine(lines, i), lines[getFunctionEndLine(lines, i)].length),
                new vscode.Range(i, 4, i, 4 + funcMatch[1].length)
            ));
        }
    }

    return symbols;
}
```

## Next Steps

- [Definition](definition.md): Implement go-to-definition
- [References](references.md): Find all references to symbols
- [Completion](completion.md): Provide code completion

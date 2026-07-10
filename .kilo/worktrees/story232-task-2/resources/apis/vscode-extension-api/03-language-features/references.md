# References Provider

References providers implement the "Find All References" feature, allowing users to find all locations where a symbol is used throughout the workspace. This helps users understand the impact of changes and navigate between related code.

## Registering a References Provider

Use `vscode.languages.registerReferenceProvider` to register a references provider:

```typescript
import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  const provider = vscode.languages.registerReferenceProvider(
    { language: "javascript" },
    new MyReferenceProvider(),
  );

  context.subscriptions.push(provider);
}
```

## Implementing a References Provider

Implement the `ReferenceProvider` interface:

```typescript
class MyReferenceProvider implements vscode.ReferenceProvider {
  provideReferences(
    document: vscode.TextDocument,
    position: vscode.Position,
    context: vscode.ReferenceContext,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<vscode.Location[]> {
    // Get the word at the position
    const range = document.getWordRangeAtPosition(position);
    if (!range) {
      return undefined;
    }

    const word = document.getText(range);

    // Find all references
    const references = this.findReferences(document, word, context);

    return references;
  }

  private findReferences(
    document: vscode.TextDocument,
    word: string,
    context: vscode.ReferenceContext,
  ): vscode.Location[] {
    // Implementation
    return [];
  }
}
```

## Reference Context

The `ReferenceContext` provides options for the search:

```typescript
interface ReferenceContext {
  // Include the declaration in the results
  includeDeclaration: boolean;
}
```

When `includeDeclaration` is true, include the symbol's definition in the results. When false, only include usages.

## Finding References

### Within a Single Document

Search for references in the current document:

```typescript
provideReferences(document, position, context, token) {
    const word = getWordAtPosition(document, position);
    const locations: vscode.Location[] = [];
    const text = document.getText();

    // Create a pattern to match the word
    const pattern = new RegExp(`\\b${escapeRegex(word)}\\b`, 'g');

    let match;
    while ((match = pattern.exec(text))) {
        if (token.isCancellationRequested) {
            return undefined;
        }

        const startPos = document.positionAt(match.index);
        const endPos = document.positionAt(match.index + match[0].length);
        const range = new vscode.Range(startPos, endPos);

        locations.push(new vscode.Location(document.uri, range));
    }

    return locations;
}
```

### Across the Workspace

Search for references across all files in the workspace:

```typescript
async provideReferences(document, position, context, token) {
    const word = getWordAtPosition(document, position);
    const locations: vscode.Location[] = [];

    // Find all relevant files
    const files = await vscode.workspace.findFiles('**/*.js', '**/node_modules/**');

    for (const file of files) {
        if (token.isCancellationRequested) {
            return undefined;
        }

        const fileLocations = await this.findReferencesInFile(file, word);
        locations.push(...fileLocations);
    }

    // Filter out declaration if needed
    if (!context.includeDeclaration) {
        return this.filterDeclaration(locations, document, position);
    }

    return locations;
}

private async findReferencesInFile(
    uri: vscode.Uri,
    word: string
): Promise<vscode.Location[]> {
    const document = await vscode.workspace.openTextDocument(uri);
    const text = document.getText();
    const locations: vscode.Location[] = [];

    const pattern = new RegExp(`\\b${escapeRegex(word)}\\b`, 'g');

    let match;
    while ((match = pattern.exec(text))) {
        const startPos = document.positionAt(match.index);
        const endPos = document.positionAt(match.index + match[0].length);
        locations.push(new vscode.Location(uri, new vscode.Range(startPos, endPos)));
    }

    return locations;
}
```

### Using Workspace Symbols

For better performance, use a pre-built symbol index:

```typescript
class ReferenceProvider implements vscode.ReferenceProvider {
  private symbolIndex: Map<string, vscode.Location[]>;

  constructor() {
    this.symbolIndex = new Map();
    this.buildIndex();
  }

  private async buildIndex() {
    const files = await vscode.workspace.findFiles("**/*.js");
    for (const file of files) {
      await this.indexFile(file);
    }
  }

  provideReferences(document, position, context, token) {
    const word = getWordAtPosition(document, position);
    let references = this.symbolIndex.get(word) || [];

    if (!context.includeDeclaration) {
      references = references.filter((ref) => !this.isDeclaration(ref));
    }

    return references;
  }
}
```

## Distinguishing References from Definitions

To properly implement `includeDeclaration`, distinguish between references and definitions:

```typescript
private isDefinition(document: vscode.TextDocument, position: vscode.Position): boolean {
    const line = document.lineAt(position.line).text;
    const textBefore = line.substring(0, position.character);

    // Check for definition patterns
    const definitionPatterns = [
        /(?:const|let|var)\s+\w*$/,           // Variable declaration
        /function\s+\w*$/,                     // Function declaration
        /class\s+\w*$/,                        // Class declaration
        /(?:public|private|protected)\s+\w*$/, // Class member
    ];

    return definitionPatterns.some(p => p.test(textBefore));
}
```

## Reference Location

Each reference is a `Location` object:

```typescript
const location = new vscode.Location(
  uri, // File URI
  new vscode.Range(startPos, endPos), // Range of the reference
);
```

The range should encompass the entire symbol name, not just part of it.

## Performance Considerations

### Limit Search Scope

Limit the search to relevant files:

```typescript
// Only search files with matching language
const files = await vscode.workspace.findFiles("**/*.js", "**/node_modules/**");

// Use file watchers to maintain an index
vscode.workspace.onDidChangeTextDocument((event) => {
  this.updateIndex(event.document);
});
```

### Use Cancellation

Check the cancellation token frequently:

```typescript
async provideReferences(document, position, context, token) {
    const results: vscode.Location[] = [];

    for (const file of files) {
        if (token.isCancellationRequested) {
            return undefined;  // Abort the search
        }

        const refs = await this.searchFile(file);
        results.push(...refs);
    }

    return results;
}
```

### Incremental Results

For long searches, consider returning partial results early:

```typescript
provideReferences(document, position, context, token) {
    // Return quick results first
    const quickResults = this.findLocalReferences(document, position);

    // Continue searching in background
    this.searchWorkspace(document, position).then(moreResults => {
        // Update results (requires custom UI)
    });

    return quickResults;
}
```

## Handling Imports and Exports

For modules, track imports and exports:

```typescript
provideReferences(document, position, context, token) {
    const word = getWordAtPosition(document, position);

    // Check if this is an exported symbol
    const exportInfo = this.findExportInfo(document, word);

    if (exportInfo) {
        // Find all files that import this symbol
        const importers = this.findImporters(exportInfo);
        const references: vscode.Location[] = [];

        for (const importer of importers) {
            const usages = this.findUsagesInFile(importer, word);
            references.push(...usages);
        }

        return references;
    }

    // Local symbol - search only in this file
    return this.findLocalReferences(document, word);
}
```

## Best Practices

**Return quickly**: Reference searches should feel responsive. Use indexing for large codebases.

**Handle cancellation**: Long searches can be cancelled. Check the token frequently.

**Respect includeDeclaration**: Filter out definitions when `includeDeclaration` is false.

**Return accurate ranges**: The range should cover the exact symbol text.

**Handle renamed symbols**: If your language supports renaming, track symbol identity across renames.

**Consider scope**: Only return references that are in scope for the given position.

## Common Patterns

### Function References

```typescript
provideReferences(document, position, context, token) {
    const word = getWordAtPosition(document, position);
    const locations: vscode.Location[] = [];

    // Find function calls
    const callPattern = new RegExp(`${escapeRegex(word)}\\s*\\(`, 'g');

    // Search in all files
    for (const file of files) {
        const text = getFileContent(file);
        let match;

        while ((match = callPattern.exec(text))) {
            const startPos = getPosition(text, match.index);
            locations.push(new vscode.Location(file, startPos));
        }
    }

    return locations;
}
```

### Variable References

```typescript
provideReferences(document, position, context, token) {
    const word = getWordAtPosition(document, position);

    // Get the scope of the variable
    const definition = findDefinition(document, position);
    const scope = getScope(definition);

    // Search only within the scope
    const scopeText = document.getText(scope);
    const references: vscode.Location[] = [];

    const pattern = new RegExp(`\\b${escapeRegex(word)}\\b`, 'g');
    let match;

    while ((match = pattern.exec(scopeText))) {
        const absolutePos = scope.start.translate(
            0,
            match.index - document.offsetAt(scope.start)
        );
        references.push(new vscode.Location(document.uri, absolutePos));
    }

    return references;
}
```

## Next Steps

- [Definition](definition.md): Implement go-to-definition
- [Symbols](symbols.md): Provide document and workspace symbols
- [Rename](#): Implement rename refactoring

# Definition Provider

Definition providers implement the "Go to Definition" feature, allowing users to navigate from a symbol usage to its definition. This feature helps users understand code by quickly jumping to where symbols are defined.

## Registering a Definition Provider

Use `vscode.languages.registerDefinitionProvider` to register a definition provider:

```typescript
import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  const provider = vscode.languages.registerDefinitionProvider(
    { language: "javascript" },
    new MyDefinitionProvider(),
  );

  context.subscriptions.push(provider);
}
```

## Implementing a Definition Provider

Implement the `DefinitionProvider` interface:

```typescript
class MyDefinitionProvider implements vscode.DefinitionProvider {
  provideDefinition(
    document: vscode.TextDocument,
    position: vscode.Position,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<vscode.Definition | vscode.LocationLink[]> {
    // Get the word at the position
    const range = document.getWordRangeAtPosition(position);
    if (!range) {
      return undefined;
    }

    const word = document.getText(range);

    // Find the definition
    const definition = findDefinition(document, word, position);
    if (!definition) {
      return undefined;
    }

    return new vscode.Location(definition.uri, definition.range);
  }
}
```

## Definition Types

### Location

A simple `Location` points to a single definition:

```typescript
return new vscode.Location(
  vscode.Uri.file("/path/to/file.js"),
  new vscode.Range(10, 0, 10, 20),
);
```

### Multiple Locations

Return an array of `Location` objects for multiple definitions:

```typescript
return [new vscode.Location(uri1, range1), new vscode.Location(uri2, range2)];
```

### LocationLink

`LocationLink` provides more control over the navigation experience:

```typescript
const link: vscode.LocationLink = {
  // The target location
  targetUri: vscode.Uri.file("/path/to/file.js"),
  targetRange: new vscode.Range(10, 0, 15, 0),

  // Optional: the specific selection within the range
  targetSelectionRange: new vscode.Range(10, 5, 10, 15),

  // Optional: the origin of the link (shown in preview)
  originSelectionRange: range,
};

return [link];
```

**targetUri**: The URI of the document containing the definition

**targetRange**: The full range of the definition

**targetSelectionRange**: The range to select when navigating (e.g., just the function name)

**originSelectionRange**: The range in the source document that triggered the navigation

## Finding Definitions

### Within the Same Document

Find definitions in the current document:

```typescript
provideDefinition(document, position, token) {
    const word = getWordAtPosition(document, position);

    // Search for definition patterns
    const text = document.getText();
    const pattern = new RegExp(`(function|const|let|var)\\s+${word}\\b`, 'g');

    let match;
    while ((match = pattern.exec(text))) {
        const startPos = document.positionAt(match.index);
        const endPos = document.positionAt(match.index + match[0].length);

        return new vscode.Location(document.uri, new vscode.Range(startPos, endPos));
    }

    return undefined;
}
```

### Across Documents

Find definitions in other files:

```typescript
async provideDefinition(document, position, token) {
    const word = getWordAtPosition(document, position);

    // Find all relevant files
    const files = await vscode.workspace.findFiles('**/*.js', '**/node_modules/**');

    for (const file of files) {
        if (token.isCancellationRequested) {
            return undefined;
        }

        const fileContent = await vscode.workspace.fs.readFile(file);
        const text = Buffer.from(fileContent).toString('utf8');

        // Search for definition
        const definition = findDefinitionInText(text, word);
        if (definition) {
            return new vscode.Location(file, definition);
        }
    }

    return undefined;
}
```

### Using a Symbol Index

For better performance, maintain a symbol index:

```typescript
class DefinitionProvider implements vscode.DefinitionProvider {
  private symbolIndex: Map<string, vscode.Location>;

  constructor() {
    this.symbolIndex = new Map();
    this.buildIndex();
  }

  private async buildIndex() {
    const files = await vscode.workspace.findFiles("**/*.js");
    for (const file of files) {
      const symbols = await this.indexFile(file);
      symbols.forEach((s) => this.symbolIndex.set(s.name, s.location));
    }
  }

  provideDefinition(document, position, token) {
    const word = getWordAtPosition(document, position);
    const location = this.symbolIndex.get(word);
    return location;
  }
}
```

## Definition Link Types

### Type Definition

Implement `TypeDefinitionProvider` to navigate to type definitions:

```typescript
vscode.languages.registerTypeDefinitionProvider(
  { language: "typescript" },
  {
    provideTypeDefinition(document, position, token) {
      // Return the location of the type definition
      return findTypeDefinition(document, position);
    },
  },
);
```

### Implementation

Implement `ImplementationProvider` to navigate to implementations:

```typescript
vscode.languages.registerImplementationProvider(
  { language: "typescript" },
  {
    provideImplementation(document, position, token) {
      // Return locations of implementations (e.g., interface implementations)
      return findImplementations(document, position);
    },
  },
);
```

### Declaration

Implement `DeclarationProvider` to navigate to declarations:

```typescript
vscode.languages.registerDeclarationProvider(
  { language: "c" },
  {
    provideDeclaration(document, position, token) {
      // Return the location of the declaration (for header files)
      return findDeclaration(document, position);
    },
  },
);
```

## Peek Definition

The "Peek Definition" feature uses the same provider. VSCode shows the definition in an inline peek window instead of navigating to it. No additional implementation is needed.

## Handling Imports

For imported symbols, resolve the import first:

```typescript
provideDefinition(document, position, token) {
    const word = getWordAtPosition(document, position);

    // Check if this is an imported symbol
    const importInfo = findImportInfo(document, word);
    if (importInfo) {
        // Navigate to the imported file
        const importedDoc = await resolveImport(document.uri, importInfo.source);
        return findDefinitionInDocument(importedDoc, importInfo.importedName);
    }

    // Look for local definition
    return findLocalDefinition(document, word);
}
```

## Best Practices

**Return quickly**: Definition navigation should feel instant. Use caching or indexing for large codebases.

**Handle cancellation**: Definition requests can be cancelled if the user moves the cursor. Check the token frequently.

**Return precise ranges**: The target range should encompass the full definition. The selection range should highlight the symbol name.

**Support multiple definitions**: When a symbol has multiple definitions (e.g., overloaded functions), return all of them.

**Handle edge cases**: Consider partial identifiers, keywords, and built-in symbols.

**Consider context**: The same word might have different definitions in different contexts (inside a function vs. at module level).

## Common Patterns

### Function Definition

```typescript
provideDefinition(document, position, token) {
    const word = getWordAtPosition(document, position);
    const text = document.getText();

    // Match function declarations
    const funcPattern = new RegExp(`function\\s+${word}\\s*\\(`, 'g');
    let match;
    const definitions: vscode.Location[] = [];

    while ((match = funcPattern.exec(text))) {
        const startPos = document.positionAt(match.index);
        definitions.push(new vscode.Location(document.uri, startPos));
    }

    // Match arrow functions and function expressions
    const arrowPattern = new RegExp(`(?:const|let|var)\\s+${word}\\s*=\\s*(?:async\\s+)?(?:\\([^)]*\\)|[a-zA-Z])\\s*=>`, 'g');
    while ((match = arrowPattern.exec(text))) {
        const startPos = document.positionAt(match.index);
        definitions.push(new vscode.Location(document.uri, startPos));
    }

    return definitions.length > 0 ? definitions : undefined;
}
```

### Variable Definition

```typescript
provideDefinition(document, position, token) {
    const word = getWordAtPosition(document, position);
    const text = document.getText();

    // Search backwards from position for the variable declaration
    const offset = document.offsetAt(position);
    const beforeCursor = text.substring(0, offset);

    const pattern = new RegExp(`(const|let|var)\\s+${word}\\b`, 'g');
    let lastMatch: RegExpExecArray | null = null;
    let match;

    while ((match = pattern.exec(beforeCursor))) {
        lastMatch = match;
    }

    if (lastMatch) {
        const startPos = document.positionAt(lastMatch.index);
        const endPos = document.positionAt(lastMatch.index + lastMatch[0].length);
        return new vscode.Location(document.uri, new vscode.Range(startPos, endPos));
    }

    return undefined;
}
```

## Next Steps

- [References](references.md): Find all references to a symbol
- [Symbols](symbols.md): Provide document and workspace symbols
- [Hover](hover.md): Show documentation on hover

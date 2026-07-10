# Code Completion

Code completion provides suggestions as users type, helping them write code faster and with fewer errors. Extensions implement completion providers to offer language-specific suggestions.

## Registering a Completion Provider

Use `vscode.languages.registerCompletionItemProvider` to register a completion provider:

```typescript
import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  const provider = vscode.languages.registerCompletionItemProvider(
    { language: "javascript" },
    new MyCompletionProvider(),
    ".", // Trigger character
  );

  context.subscriptions.push(provider);
}
```

### Trigger Characters

Trigger characters cause completion to activate automatically when typed:

```typescript
vscode.languages.registerCompletionItemProvider(
  { language: "javascript" },
  new MyCompletionProvider(),
  ".",
  '"',
  "'", // Multiple trigger characters
);
```

Common trigger characters include:

- `.`: Member access
- `"` or `'`: String literals
- `<`: Tag opening
- `@`: Decorators or annotations
- ` `: Space (use sparingly)

## Implementing a Completion Provider

Implement the `CompletionItemProvider` interface:

```typescript
class MyCompletionProvider implements vscode.CompletionItemProvider {
  provideCompletionItems(
    document: vscode.TextDocument,
    position: vscode.Position,
    token: vscode.CancellationToken,
    context: vscode.CompletionContext,
  ): vscode.ProviderResult<vscode.CompletionItem[] | vscode.CompletionList> {
    // Return completion items
    return [
      new vscode.CompletionItem("console", vscode.CompletionItemKind.Method),
      new vscode.CompletionItem("log", vscode.CompletionItemKind.Method),
    ];
  }
}
```

### The provideCompletionItems Method

The method receives these parameters:

**document**: The document where completion is requested

**position**: The cursor position where completion is triggered

**token**: Cancellation token to abort the operation

**context**: Additional context about the completion request

### Completion Context

The `CompletionContext` provides information about how completion was triggered:

```typescript
interface CompletionContext {
  // How completion was triggered
  triggerKind: vscode.CompletionTriggerKind;

  // Character that triggered completion (if triggerKind is TriggerCharacter)
  triggerCharacter?: string;
}
```

Trigger kinds:

- `Invoke`: User explicitly invoked completion (Ctrl+Space)
- `TriggerCharacter`: User typed a trigger character
- `TriggerForIncompleteCompletions`: Completion was re-triggered after incomplete results

## Completion Items

`CompletionItem` represents a single suggestion:

```typescript
const item = new vscode.CompletionItem(
  "myFunction",
  vscode.CompletionItemKind.Function,
);
item.detail = "void myFunction()";
item.documentation = "This function does something useful";
item.insertText = "myFunction($1)";
item.insertTextFormat = vscode.InsertTextFormat.Snippet;
```

### Completion Item Properties

**label**: The text displayed in the completion list. Can be a string or an object for rich display:

```typescript
item.label = {
  label: "myFunction",
  description: "MyModule",
  detail: "void myFunction()",
};
```

**kind**: The icon displayed next to the item. Common kinds include:

| Kind        | Use Case              |
| ----------- | --------------------- |
| `Class`     | Class definitions     |
| `Constant`  | Constants             |
| `Function`  | Functions             |
| `Interface` | Interfaces            |
| `Method`    | Methods               |
| `Module`    | Modules or namespaces |
| `Property`  | Properties            |
| `Variable`  | Variables             |
| `Keyword`   | Language keywords     |
| `Snippet`   | Code snippets         |
| `File`      | File names            |
| `Folder`    | Folder names          |

**detail**: Short description shown after the label:

```typescript
item.detail = "string -> number";
```

**documentation**: Longer documentation shown in the details pane:

````typescript
item.documentation = new vscode.MarkdownString(
  "Converts a string to a number.\n\n" +
    '**Example:**\n```javascript\nparseInt("42"); // 42\n```',
);
````

**sortText**: Text used for sorting. Items are sorted alphabetically by this value:

```typescript
item.sortText = "a"; // Appears first
item.sortText = "z"; // Appears last
```

**filterText**: Text used for filtering. If not set, the label is used:

```typescript
item.label = "myFunction";
item.filterText = "mf"; // User can type 'mf' to find 'myFunction'
```

**insertText**: Text to insert when the item is selected. Can be a string or `SnippetString`:

```typescript
// Simple text
item.insertText = "myFunction()";

// Snippet with placeholders
item.insertText = new vscode.SnippetString("myFunction(${1:arg1}, ${2:arg2})");
```

**insertTextFormat**: Format of the insert text:

- `PlainText`: Insert as-is
- `Snippet`: Interpret as snippet syntax

**textEdit**: Replace a range instead of inserting at the cursor:

```typescript
// Replace the current word
const range = document.getWordRangeAtPosition(position);
item.textEdit = vscode.TextEdit.replace(range, "replacementText");
```

**additionalTextEdits**: Additional edits to apply:

```typescript
item.additionalTextEdits = [
  vscode.TextEdit.insert(
    new vscode.Position(0, 0),
    'import { myFunction } from "module";\n',
  ),
];
```

**command**: Command to execute after insertion:

```typescript
item.command = {
  command: "editor.action.formatDocument",
  title: "Format",
};
```

**commitCharacters**: Characters that commit the completion when typed:

```typescript
item.commitCharacters = [".", ",", "(", ")"];
```

**preselect**: Whether to preselect this item:

```typescript
item.preselect = true;
```

**tags**: Additional tags for the item:

```typescript
item.tags = [vscode.CompletionItemTag.Deprecated];
```

## Completion List

Return a `CompletionList` for more control:

```typescript
return new vscode.CompletionList([item1, item2], false); // incomplete: false means no more results available
```

### Incomplete Results

When `isIncomplete` is true, VSCode will re-request completions as the user types:

```typescript
provideCompletionItems(document, position, token, context) {
    // Quick initial results
    const quickResults = getQuickResults();

    return new vscode.CompletionList(quickResults, true);  // incomplete
}
```

This is useful for:

- Large result sets that should be filtered server-side
- Results that require expensive computation
- Network-based completion

## Resolving Completion Items

For expensive operations like fetching documentation, implement `resolveCompletionItem`:

```typescript
class MyCompletionProvider implements vscode.CompletionItemProvider {
  provideCompletionItems(document, position, token, context) {
    // Return items without expensive details
    return items;
  }

  resolveCompletionItem(
    item: vscode.CompletionItem,
    token: vscode.CancellationToken,
  ) {
    // Called when user selects the item
    item.documentation = fetchDocumentation(item.label);
    return item;
  }
}
```

## Context-Aware Completion

Use the document and position to provide relevant suggestions:

```typescript
provideCompletionItems(document, position, token, context) {
    // Get the current line
    const line = document.lineAt(position.line).text;

    // Get the text before the cursor
    const textBeforeCursor = line.substring(0, position.character);

    // Check if we're after a dot
    if (textBeforeCursor.endsWith('.')) {
        return getMemberCompletions(textBeforeCursor);
    }

    // Get the current word
    const wordRange = document.getWordRangeAtPosition(position);
    const currentWord = wordRange ? document.getText(wordRange) : '';

    // Filter completions based on current word
    return allCompletions.filter(item =>
        item.label.startsWith(currentWord)
    );
}
```

## Snippet Completions

Provide snippet-style completions:

```typescript
const snippetItem = new vscode.CompletionItem(
  "for",
  vscode.CompletionItemKind.Snippet,
);
snippetItem.insertText = new vscode.SnippetString(
  [
    "for (let \${1:i} = 0; \${1:i} < \${2:array}.length; \${1:i}++) {",
    "\t\${3:// code}",
    "}",
  ].join("\n"),
);
snippetItem.documentation = new vscode.MarkdownString("Creates a for loop");
```

### Snippet Syntax

- `${1:placeholder}`: Tab stop with placeholder text
- `${1}`: Tab stop without placeholder
- `${1|one,two,three|}`: Choice placeholder
- `$0`: Final cursor position
- `${1:default}`: Placeholder with default value
- `\${`: Literal `${` (escaped)

## File and Folder Completions

Provide file path completions:

```typescript
provideCompletionItems(document, position, token, context) {
    const line = document.lineAt(position.line).text;
    const textBefore = line.substring(0, position.character);

    // Match import statements
    const importMatch = textBefore.match(/from\s+['"]([^'"]*)$/);
    if (importMatch) {
        const partialPath = importMatch[1];
        return getFileCompletions(partialPath);
    }

    return [];
}

async function getFileCompletions(partialPath: string): Promise<vscode.CompletionItem[]> {
    const files = await vscode.workspace.findFiles(`**/${partialPath}*`);
    return files.map(uri => {
        const item = new vscode.CompletionItem(
            path.basename(uri.fsPath),
            vscode.CompletionItemKind.File
        );
        item.insertText = path.basename(uri.fsPath, path.extname(uri.fsPath));
        return item;
    });
}
```

## Best Practices

**Prioritize relevance**: Show the most relevant items first using `sortText` and `preselect`.

**Use appropriate kinds**: Set the correct `CompletionItemKind` for proper icons.

**Provide documentation**: Include `detail` and `documentation` to help users understand suggestions.

**Use snippets wisely**: Snippets help users write common patterns quickly, but avoid overuse.

**Handle cancellation**: Check the cancellation token and abort expensive operations.

**Consider context**: Provide suggestions relevant to the current position and document.

**Avoid duplicates**: Don't return items that VSCode already provides.

**Test performance**: Ensure completion is fast enough for real-time use.

## Common Patterns

### Member Completion

```typescript
provideCompletionItems(document, position) {
    const line = document.lineAt(position.line).text;
    const prefix = line.substring(0, position.character);

    // Check for object.member pattern
    const match = prefix.match(/(\w+)\.\w*$/);
    if (match) {
        const objectName = match[1];
        return getMembers(objectName);
    }

    return [];
}
```

### Function Parameter Completion

```typescript
provideCompletionItems(document, position) {
    const line = document.lineAt(position.line).text;
    const prefix = line.substring(0, position.character);

    // Check for function call
    const match = prefix.match(/(\w+)\([^)]*$/);
    if (match) {
        const functionName = match[1];
        return getParameterCompletions(functionName);
    }

    return [];
}
```

### Import Completion

```typescript
provideCompletionItems(document, position) {
    const line = document.lineAt(position.line).text;
    const prefix = line.substring(0, position.character);

    if (prefix.match(/import\s+.*\s+from\s+['"]?$/)) {
        return getModuleCompletions();
    }

    return [];
}
```

## Next Steps

- [Hover](hover.md): Show documentation on hover
- [Signature Help](#): Show function signatures
- [Diagnostics](diagnostics.md): Provide error and warning diagnostics

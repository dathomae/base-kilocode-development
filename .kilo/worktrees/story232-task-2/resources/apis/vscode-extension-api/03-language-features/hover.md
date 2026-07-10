# Hover Provider

Hover providers display documentation and type information when users hover over symbols in the editor. This feature helps users understand code without navigating away from their current position.

## Registering a Hover Provider

Use `vscode.languages.registerHoverProvider` to register a hover provider:

```typescript
import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  const provider = vscode.languages.registerHoverProvider(
    { language: "javascript" },
    new MyHoverProvider(),
  );

  context.subscriptions.push(provider);
}
```

## Implementing a Hover Provider

Implement the `HoverProvider` interface:

```typescript
class MyHoverProvider implements vscode.HoverProvider {
  provideHover(
    document: vscode.TextDocument,
    position: vscode.Position,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<vscode.Hover> {
    // Get the word at the position
    const range = document.getWordRangeAtPosition(position);
    if (!range) {
      return undefined;
    }

    const word = document.getText(range);

    // Look up documentation for the word
    const docs = getDocumentation(word);
    if (!docs) {
      return undefined;
    }

    return new vscode.Hover(docs, range);
  }
}
```

## The Hover Object

Create a `Hover` object with the content to display:

```typescript
// Simple string
return new vscode.Hover("This is a variable of type string");

// Markdown string
const markdown = new vscode.MarkdownString();
markdown.appendMarkdown("**Type:** `string`\n\n");
markdown.appendMarkdown("A variable that holds text data.");
return new vscode.Hover(markdown);

// Array of markdown parts (displayed vertically)
return new vscode.Hover([
  new vscode.MarkdownString("**Type:** `string`"),
  new vscode.MarkdownString("A variable that holds text data."),
]);
```

### Hover Range

Specify the range that triggers the hover:

```typescript
// Hover applies to the word under cursor
const range = document.getWordRangeAtPosition(position);
return new vscode.Hover(content, range);

// Hover applies to a specific range
const customRange = new vscode.Range(
  new vscode.Position(line, startChar),
  new vscode.Position(line, endChar),
);
return new vscode.Hover(content, customRange);
```

If no range is provided, the hover appears at the cursor position.

## Markdown Content

Use `MarkdownString` for rich formatting:

```typescript
const content = new vscode.MarkdownString();

// Append plain text
content.appendText("Plain text\n");

// Append markdown
content.appendMarkdown("**Bold** and *italic*\n");

// Append code
content.appendCodeblock("const x = 42;", "javascript");

// Append markdown with links
content.appendMarkdown("[Documentation](https://example.com)");
```

### Supported Markdown

VSCode supports a subset of CommonMark:

- **Bold**: `**text**`
- _Italic_: `*text*`
- `Code`: `` `code` ``
- Links: `[text](url)`
- Images: `![alt](url)`
- Lists: `- item` or `1. item`
- Headers: `# Header`
- Code blocks: ` ```language\ncode\n``` `

### Trusted Markdown

For security, markdown is sanitized by default. If your content includes commands or scripts, mark it as trusted:

```typescript
const content = new vscode.MarkdownString(
  "Click [here](command:myExtension.openDocs)",
);
content.isTrusted = true;
content.supportHtml = true;
```

### Command Links

Create clickable links that execute commands:

```typescript
const content = new vscode.MarkdownString();
content.appendMarkdown(
  "Click to [Go to Definition](command:editor.action.goToDeclaration)",
);
content.isTrusted = true;
```

Pass arguments to commands using URI encoding:

```typescript
const args = [{ line: 10, file: "/path/to/file.js" }];
const encodedArgs = encodeURIComponent(JSON.stringify(args));
content.appendMarkdown(`[Jump](command:myExtension.jumpTo?${encodedArgs})`);
```

## Context-Aware Hovers

Use the document and position to provide relevant information:

```typescript
provideHover(document, position, token) {
    // Get the word at the position
    const range = document.getWordRangeAtPosition(position);
    if (!range) {
        return undefined;
    }

    const word = document.getText(range);

    // Check the context around the word
    const line = document.lineAt(position.line).text;
    const textBefore = line.substring(0, position.character);

    // Different hover for different contexts
    if (textBefore.includes('import')) {
        return getImportHover(word);
    }

    if (textBefore.match(/\.\w*$/)) {
        return getMemberHover(word);
    }

    return getSymbolHover(word);
}
```

## Multiple Hover Providers

Multiple extensions can provide hovers for the same language. VSCode merges the results. Each provider should:

- Return `undefined` when it has nothing to contribute
- Return only relevant content for the position
- Keep content concise to avoid overwhelming the user

## Asynchronous Hovers

For hovers that require network requests or expensive computation:

```typescript
async provideHover(document, position, token) {
    // Check for cancellation
    if (token.isCancellationRequested) {
        return undefined;
    }

    const word = getWordAtPosition(document, position);

    // Fetch documentation asynchronously
    const docs = await fetchDocumentation(word);

    // Check cancellation again
    if (token.isCancellationRequested) {
        return undefined;
    }

    return new vscode.Hover(docs);
}
```

## Best Practices

**Keep hovers concise**: Show essential information. Long hovers obscure the editor.

**Show type information**: Display the type of variables, functions, and expressions.

**Include documentation**: Show docstrings or comments for functions and classes.

**Use markdown formatting**: Bold for labels, code formatting for identifiers and types.

**Handle cancellation**: Abort expensive operations when the user moves the cursor.

**Return undefined for irrelevant positions**: Don't return empty hovers; return undefined instead.

**Consider hover timing**: Hovers appear after a brief delay. Ensure your provider is fast enough.

## Common Patterns

### Type Information Hover

```typescript
provideHover(document, position) {
    const info = getTypeAtPosition(document, position);
    if (!info) {
        return undefined;
    }

    const content = new vscode.MarkdownString();
    content.appendMarkdown(`**Type:** \`${info.type}\`\n\n`);
    if (info.documentation) {
        content.appendMarkdown(info.documentation);
    }

    return new vscode.Hover(content);
}
```

### Function Signature Hover

```typescript
provideHover(document, position) {
    const signature = getFunctionSignature(document, position);
    if (!signature) {
        return undefined;
    }

    const content = new vscode.MarkdownString();
    content.appendCodeblock(signature.declaration, document.languageId);
    content.appendMarkdown('\n\n');
    content.appendMarkdown(signature.documentation);

    return new vscode.Hover(content);
}
```

### Error Information Hover

```typescript
provideHover(document, position) {
    const errors = getErrorsAtPosition(document, position);
    if (errors.length === 0) {
        return undefined;
    }

    const content = new vscode.MarkdownString();
    content.appendMarkdown('**Errors:**\n\n');
    errors.forEach(error => {
        content.appendMarkdown(`- ${error.message}\n`);
    });

    return new vscode.Hover(content);
}
```

## Next Steps

- [Completion](completion.md): Provide code completion suggestions
- [Diagnostics](diagnostics.md): Show errors and warnings
- [Definition](definition.md): Implement go-to-definition

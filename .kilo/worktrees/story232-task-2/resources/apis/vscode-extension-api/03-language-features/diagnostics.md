# Diagnostics

Diagnostics indicate problems in source code, such as syntax errors, type errors, or linting issues. Extensions provide diagnostics through diagnostic collections that VSCode displays in the editor.

## Creating Diagnostics

A diagnostic represents a problem in a document:

```typescript
const diagnostic = new vscode.Diagnostic(
  new vscode.Range(new vscode.Position(0, 0), new vscode.Position(0, 10)),
  "Unexpected token",
  vscode.DiagnosticSeverity.Error,
);
```

### Diagnostic Properties

**range**: The document range where the problem occurs

**message**: The human-readable problem description

**severity**: The diagnostic severity level

| Severity      | Use Case                                         |
| ------------- | ------------------------------------------------ |
| `Error`       | Blocking problems that prevent code from working |
| `Warning`     | Potential issues that should be addressed        |
| `Information` | Suggestions or informational notes               |
| `Hint`        | Subtle hints for code improvement                |

**source**: The source of the diagnostic (e.g., linter name):

```typescript
diagnostic.source = "eslint";
```

**code**: An identifier for the diagnostic:

```typescript
diagnostic.code = "no-unused-vars";
// Or with a target URL
diagnostic.code = {
  value: "no-unused-vars",
  target: vscode.Uri.parse("https://eslint.org/docs/rules/no-unused-vars"),
};
```

**relatedInformation**: Additional locations related to the problem:

```typescript
diagnostic.relatedInformation = [
  new vscode.DiagnosticRelatedInformation(
    new vscode.Location(otherUri, otherRange),
    "The variable was declared here",
  ),
];
```

**tags**: Diagnostic tags for special rendering:

```typescript
diagnostic.tags = [
  vscode.DiagnosticTag.Unnecessary,
  vscode.DiagnosticTag.Deprecated,
];
```

- `Unnecessary`: Rendered faded/strikethrough
- `Deprecated`: Rendered with a strike through

## Diagnostic Collections

Diagnostics are managed through `DiagnosticCollection` objects:

```typescript
// Create a collection
const diagnosticCollection =
  vscode.languages.createDiagnosticCollection("myExtension");

// Set diagnostics for a document
diagnosticCollection.set(document.uri, [diagnostic1, diagnostic2]);

// Clear diagnostics for a document
diagnosticCollection.set(document.uri, undefined);

// Clear all diagnostics
diagnosticCollection.clear();

// Dispose the collection
diagnosticCollection.dispose();
```

### Registering a Collection

Register the collection during activation:

```typescript
export function activate(context: vscode.ExtensionContext) {
  const diagnosticCollection =
    vscode.languages.createDiagnosticCollection("myExtension");
  context.subscriptions.push(diagnosticCollection);

  // Subscribe to document changes
  context.subscriptions.push(
    vscode.workspace.onDidChangeTextDocument((event) => {
      updateDiagnostics(event.document, diagnosticCollection);
    }),
  );
}
```

## Updating Diagnostics

Diagnostics should be updated when documents change:

```typescript
function updateDiagnostics(
  document: vscode.TextDocument,
  collection: vscode.DiagnosticCollection,
): void {
  // Only process relevant documents
  if (document.languageId !== "javascript") {
    return;
  }

  const diagnostics: vscode.Diagnostic[] = [];

  // Analyze the document
  const text = document.getText();
  const problems = analyzeCode(text);

  for (const problem of problems) {
    const diagnostic = new vscode.Diagnostic(
      new vscode.Range(
        document.positionAt(problem.start),
        document.positionAt(problem.end),
      ),
      problem.message,
      problem.severity,
    );
    diagnostic.source = "myExtension";
    diagnostic.code = problem.code;
    diagnostics.push(diagnostic);
  }

  // Update the collection
  collection.set(document.uri, diagnostics);
}
```

## Diagnostic Events

### Document Change Events

Update diagnostics when documents change:

```typescript
vscode.workspace.onDidChangeTextDocument((event) => {
  // The document that changed
  const document = event.document;

  // The changes made
  const changes = event.contentChanges;

  // Update diagnostics
  updateDiagnostics(document, diagnosticCollection);
});
```

### Document Open/Save Events

Update diagnostics when documents open or save:

```typescript
// On open
vscode.workspace.onDidOpenTextDocument((document) => {
  updateDiagnostics(document, diagnosticCollection);
});

// On save
vscode.workspace.onDidSaveTextDocument((document) => {
  updateDiagnostics(document, diagnosticCollection);
});
```

### Initial Diagnostics

Compute diagnostics for already-open documents on activation:

```typescript
export function activate(context: vscode.ExtensionContext) {
  const diagnosticCollection =
    vscode.languages.createDiagnosticCollection("myExtension");

  // Process all open text documents
  for (const document of vscode.workspace.textDocuments) {
    updateDiagnostics(document, diagnosticCollection);
  }

  // Subscribe to future changes
  context.subscriptions.push(
    vscode.workspace.onDidChangeTextDocument((event) => {
      updateDiagnostics(event.document, diagnosticCollection);
    }),
  );
}
```

## Performance Considerations

### Debouncing Updates

For expensive analysis, debounce updates to avoid recomputing on every keystroke:

```typescript
let timeout: NodeJS.Timeout | undefined;

vscode.workspace.onDidChangeTextDocument((event) => {
  if (timeout) {
    clearTimeout(timeout);
  }

  timeout = setTimeout(() => {
    updateDiagnostics(event.document, diagnosticCollection);
  }, 500); // Wait 500ms after last change
});
```

### Incremental Updates

For large documents, update only affected diagnostics:

```typescript
vscode.workspace.onDidChangeTextDocument((event) => {
  // Get the changed ranges
  const changes = event.contentChanges;

  // Only recompute diagnostics for changed regions
  for (const change of changes) {
    const affectedRange = getAffectedRange(change);
    recomputeDiagnosticsInRange(event.document, affectedRange);
  }
});
```

### Background Processing

Run expensive analysis in the background:

```typescript
async function updateDiagnosticsAsync(
  document: vscode.TextDocument,
  collection: vscode.DiagnosticCollection,
): Promise<void> {
  // Run analysis in a worker or separate process
  const problems = await analyzeInBackground(document);

  // Check if document is still valid
  if (document.isClosed) {
    return;
  }

  const diagnostics = problems.map((p) => createDiagnostic(p));
  collection.set(document.uri, diagnostics);
}
```

## Code Actions for Diagnostics

Provide quick fixes for diagnostics through code actions:

```typescript
class MyCodeActionProvider implements vscode.CodeActionProvider {
  provideCodeActions(
    document: vscode.TextDocument,
    range: vscode.Range,
    context: vscode.CodeActionContext,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<(vscode.CodeAction | vscode.Command)[]> {
    const actions: vscode.CodeAction[] = [];

    // For each diagnostic in the range
    for (const diagnostic of context.diagnostics) {
      if (diagnostic.code === "no-unused-vars") {
        const action = new vscode.CodeAction(
          "Remove unused variable",
          vscode.CodeActionKind.QuickFix,
        );
        action.diagnostics = [diagnostic];
        action.edit = createRemoveVariableEdit(document, diagnostic.range);
        actions.push(action);
      }
    }

    return actions;
  }
}
```

## Multiple Diagnostic Sources

Multiple extensions can provide diagnostics for the same document. Each diagnostic collection is independent:

```typescript
// Extension A provides linting diagnostics
const lintingCollection = vscode.languages.createDiagnosticCollection("linter");

// Extension B provides type checking diagnostics
const typeCollection =
  vscode.languages.createDiagnosticCollection("typeChecker");
```

VSCode displays diagnostics from all collections, merging them in the Problems panel.

## Best Practices

**Use appropriate severities**: Reserve `Error` for actual problems. Use `Warning` for style issues and `Hint` for suggestions.

**Provide clear messages**: Write messages that explain the problem and suggest a fix when possible.

**Include source information**: Set the `source` property so users know where the diagnostic comes from.

**Use diagnostic codes**: Include codes for documentation links and code actions.

**Keep diagnostics fresh**: Update diagnostics when documents change, but debounce expensive operations.

**Clean up properly**: Clear diagnostics when documents close or when your extension deactivates.

**Consider performance**: Large files with many diagnostics can slow down the editor. Limit the number of diagnostics when practical.

## Common Patterns

### Linter Integration

```typescript
export function activate(context: vscode.ExtensionContext) {
  const collection = vscode.languages.createDiagnosticCollection("eslint");

  const update = (document: vscode.TextDocument) => {
    if (document.languageId !== "javascript") {
      return;
    }

    const diagnostics = runLinter(document);
    collection.set(document.uri, diagnostics);
  };

  // Initial diagnostics
  vscode.workspace.textDocuments.forEach(update);

  // Update on changes
  context.subscriptions.push(
    vscode.workspace.onDidChangeTextDocument((e) => update(e.document)),
    vscode.workspace.onDidOpenTextDocument(update),
    collection,
  );
}
```

### Syntax Error Detection

```typescript
function detectSyntaxErrors(
  document: vscode.TextDocument,
): vscode.Diagnostic[] {
  const diagnostics: vscode.Diagnostic[] = [];
  const text = document.getText();

  try {
    // Parse the document
    parse(text);
  } catch (error) {
    // Create diagnostic for syntax error
    const diagnostic = new vscode.Diagnostic(
      getRangeFromError(error),
      error.message,
      vscode.DiagnosticSeverity.Error,
    );
    diagnostic.source = "parser";
    diagnostics.push(diagnostic);
  }

  return diagnostics;
}
```

## Next Steps

- [Completion](completion.md): Provide code completion suggestions
- [Hover](hover.md): Show documentation on hover
- [Definition](definition.md): Implement go-to-definition

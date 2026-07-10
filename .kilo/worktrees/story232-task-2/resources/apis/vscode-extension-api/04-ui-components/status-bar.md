# Status Bar

The status bar at the bottom of VSCode displays information and provides quick access to commands. Extensions can add items to the status bar to show relevant information or provide clickable actions.

## Creating a Status Bar Item

Use `vscode.window.createStatusBarItem` to create a status bar item:

```typescript
import * as vscode from "vscode";

const statusBarItem = vscode.window.createStatusBarItem(
  vscode.StatusBarAlignment.Right,
  100,
);

statusBarItem.text = "$(check) Ready";
statusBarItem.tooltip = "Extension is ready";
statusBarItem.command = "myExtension.showOptions";

statusBarItem.show();
```

## Status Bar Item Properties

### Alignment

Position items on the left or right side of the status bar:

```typescript
// Left side (near the beginning)
vscode.StatusBarAlignment.Left;

// Right side (near the end)
vscode.StatusBarAlignment.Right;
```

### Priority

Higher priority items appear closer to the edge:

```typescript
// Lower priority (further from edge)
const item1 = vscode.window.createStatusBarItem(
  vscode.StatusBarAlignment.Right,
  0,
);

// Higher priority (closer to edge)
const item2 = vscode.window.createStatusBarItem(
  vscode.StatusBarAlignment.Right,
  100,
);
```

### Text

The text displayed in the status bar. Use `$(icon-name)` syntax for icons:

```typescript
statusBarItem.text = "$(file-text) 3 files";
statusBarItem.text = "$(sync~spin) Syncing..."; // Spinning animation
statusBarItem.text = "$(check) Done";
```

Common icons include: `check`, `x`, `sync`, `file`, `folder`, `gear`, `search`, `play`, `debug`, `alert`, `info`, `warning`, `error`.

### Tooltip

Text shown when hovering over the item:

```typescript
statusBarItem.tooltip = "Click to view details";
statusBarItem.tooltip = new vscode.MarkdownString(
  "**Status:** Ready\n\nClick for options",
);
```

### Command

Command to execute when the item is clicked:

```typescript
statusBarItem.command = "myExtension.showOptions";

// With arguments
statusBarItem.command = {
  command: "myExtension.openFile",
  title: "Open File",
  arguments: ["/path/to/file"],
};
```

### Accessibility Information

Provide accessibility information for screen readers:

```typescript
statusBarItem.accessibilityInformation = {
  label: "Extension status: Ready",
  role: "button",
};
```

### Background and Foreground Colors

Customize the item's appearance:

```typescript
statusBarItem.backgroundColor = new vscode.ThemeColor(
  "statusBarItem.errorBackground",
);
statusBarItem.color = new vscode.ThemeColor("statusBarItem.errorForeground");
```

Standard theme colors:

- `statusBarItem.errorBackground` / `statusBarItem.errorForeground`
- `statusBarItem.warningBackground` / `statusBarItem.warningForeground`
- `statusBarItem.prominentBackground` / `statusBarItem.prominentForeground`

## Showing and Hiding

Status bar items are hidden by default. Show them when ready:

```typescript
// Show the item
statusBarItem.show();

// Hide the item
statusBarItem.hide();

// Check visibility
if (statusBarItem.visible) {
  // Item is visible
}
```

## Updating Status Bar Items

Update properties dynamically:

```typescript
function updateStatus(count: number, isProcessing: boolean) {
  if (isProcessing) {
    statusBarItem.text = "$(sync~spin) Processing...";
    statusBarItem.backgroundColor = undefined;
  } else if (count > 0) {
    statusBarItem.text = `$(check) ${count} items`;
    statusBarItem.backgroundColor = undefined;
  } else {
    statusBarItem.text = "$(warning) No items";
    statusBarItem.backgroundColor = new vscode.ThemeColor(
      "statusBarItem.warningBackground",
    );
  }
}
```

## Disposing Status Bar Items

Clean up when the extension deactivates:

```typescript
export function activate(context: vscode.ExtensionContext) {
  const statusBarItem = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100,
  );

  // Add to subscriptions for automatic disposal
  context.subscriptions.push(statusBarItem);

  statusBarItem.show();
}
```

Or dispose manually:

```typescript
statusBarItem.dispose();
```

## Conditional Visibility

Show status bar items only when relevant:

```typescript
// Show only for specific file types
vscode.window.onDidChangeActiveTextEditor((editor) => {
  if (editor && editor.document.languageId === "javascript") {
    statusBarItem.show();
  } else {
    statusBarItem.hide();
  }
});

// Show based on configuration
const config = vscode.workspace.getConfiguration("myExtension");
if (config.get<boolean>("showStatusBar")) {
  statusBarItem.show();
}
```

## Multiple Status Bar Items

Create multiple items for different purposes:

```typescript
// Main status
const mainStatus = vscode.window.createStatusBarItem(
  vscode.StatusBarAlignment.Left,
  100,
);
mainStatus.text = "$(check) Ready";
mainStatus.show();

// Error count
const errorStatus = vscode.window.createStatusBarItem(
  vscode.StatusBarAlignment.Right,
  50,
);
errorStatus.text = "$(error) 0";
errorStatus.command = "workbench.actions.view.problems";
errorStatus.show();

// Line count
const lineStatus = vscode.window.createStatusBarItem(
  vscode.StatusBarAlignment.Right,
  0,
);
lineStatus.text = "Lines: 0";
lineStatus.show();
```

## Best Practices

**Use appropriate alignment**: Put important information on the left, actions on the right.

**Keep text concise**: Status bar space is limited. Use icons and short text.

**Provide tooltips**: Explain what the item shows and what happens when clicked.

**Use icons**: Icons make items more recognizable and save space.

**Update appropriately**: Keep the status bar current, but avoid excessive updates.

**Hide when not relevant**: Don't show items that don't apply to the current context.

**Use theme colors**: Use `ThemeColor` for colors to support light and dark themes.

## Common Patterns

### Error Counter

```typescript
class ErrorStatus {
  private item: vscode.StatusBarItem;
  private errorCount: number = 0;

  constructor() {
    this.item = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Left,
      100,
    );
    this.item.command = "workbench.actions.view.problems";
    this.update();
    this.item.show();
  }

  setErrors(count: number) {
    this.errorCount = count;
    this.update();
  }

  private update() {
    if (this.errorCount === 0) {
      this.item.text = "$(check) No errors";
      this.item.backgroundColor = undefined;
    } else {
      this.item.text = `$(error) ${this.errorCount}`;
      this.item.backgroundColor = new vscode.ThemeColor(
        "statusBarItem.errorBackground",
      );
    }
    this.item.tooltip = `${this.errorCount} error(s) found`;
  }

  dispose() {
    this.item.dispose();
  }
}
```

### Processing Indicator

```typescript
class ProcessingStatus {
  private item: vscode.StatusBarItem;

  constructor() {
    this.item = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Left,
      100,
    );
  }

  start(message: string = "Processing") {
    this.item.text = `$(sync~spin) ${message}`;
    this.item.show();
  }

  stop() {
    this.item.hide();
  }

  done(message: string = "Done") {
    this.item.text = `$(check) ${message}`;
    setTimeout(() => this.item.hide(), 2000);
  }
}
```

### Git Branch Display

```typescript
class GitBranchStatus {
  private item: vscode.StatusBarItem;

  constructor() {
    this.item = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Left,
      100,
    );
    this.item.command = "git.checkout";
    this.updateBranch();
    this.item.show();
  }

  async updateBranch() {
    const branch = await this.getCurrentBranch();
    this.item.text = `$(git-branch) ${branch}`;
    this.item.tooltip = `Current branch: ${branch}\nClick to checkout`;
  }

  private async getCurrentBranch(): Promise<string> {
    // Implementation to get current git branch
    return "main";
  }
}
```

### Word Count

```typescript
class WordCountStatus {
  private item: vscode.StatusBarItem;

  constructor() {
    this.item = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Right,
      0,
    );

    vscode.window.onDidChangeActiveTextEditor(() => this.update());
    vscode.window.onDidChangeTextEditorSelection(() => this.update());

    this.update();
  }

  private update() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
      this.item.hide();
      return;
    }

    const text = editor.document.getText();
    const wordCount = this.countWords(text);

    this.item.text = `$(pencil) ${wordCount} words`;
    this.item.tooltip = "Word count";
    this.item.show();
  }

  private countWords(text: string): number {
    return text.split(/\s+/).filter((w) => w.length > 0).length;
  }
}
```

## Next Steps

- [Notifications](notifications.md): Show messages and progress indicators
- [Input](input.md): Collect user input
- [Tree Views](tree-views.md): Create hierarchical views

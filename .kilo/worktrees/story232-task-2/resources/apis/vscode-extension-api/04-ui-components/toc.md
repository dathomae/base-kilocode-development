# UI Components

This section covers the user interface components that extensions can create and customize in VSCode.

## Table of Contents

**[webviews.md](webviews.md)**
: Creating custom web-based views with HTML/CSS/JavaScript

**[tree-views.md](tree-views.md)**
: Implementing tree views for the sidebar

**[status-bar.md](status-bar.md)**
: Adding items to the status bar

**[notifications.md](notifications.md)**
: Showing messages, notifications, and progress indicators

**[input.md](input.md)**
: Collecting user input through quick picks and input boxes

## Quick Reference

### Creating a Webview

```typescript
const panel = vscode.window.createWebviewPanel(
  "myWebview",
  "My Webview",
  vscode.ViewColumn.One,
  { enableScripts: true },
);

panel.webview.html = `<!DOCTYPE html>
<html>
<head><title>My Webview</title></head>
<body><h1>Hello from Webview</h1></body>
</html>`;
```

### Creating a Tree View

```typescript
vscode.window.createTreeView("myExtension.view", {
  treeDataProvider: new MyTreeDataProvider(),
});
```

### Adding Status Bar Item

```typescript
const item = vscode.window.createStatusBarItem(
  vscode.StatusBarAlignment.Right,
  100,
);
item.text = "$(check) Ready";
item.show();
```

### Showing a Message

```typescript
vscode.window.showInformationMessage("Operation completed");
vscode.window.showWarningMessage("This may cause issues");
vscode.window.showErrorMessage("Operation failed");
```

### Getting User Input

```typescript
const input = await vscode.window.showInputBox({
  prompt: "Enter a name",
  placeHolder: "Name",
});

const selection = await vscode.window.showQuickPick(
  ["Option 1", "Option 2", "Option 3"],
  { placeHolder: "Select an option" },
);
```

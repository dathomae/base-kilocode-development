# Webviews

Webviews allow extensions to create fully customizable user interfaces using HTML, CSS, and JavaScript. They appear in editor tabs and can display any web content, making them ideal for previews, dashboards, and custom tools.

## Creating a Webview

Use `vscode.window.createWebviewPanel` to create a webview:

```typescript
import * as vscode from "vscode";

const panel = vscode.window.createWebviewPanel(
  "myWebview", // View type identifier
  "My Webview", // Title shown in the tab
  vscode.ViewColumn.One, // Column to show in
  {
    enableScripts: true, // Enable JavaScript
    retainContextWhenHidden: true, // Keep state when hidden
  },
);

panel.webview.html = getWebviewContent();
```

## Webview Options

```typescript
{
    // Enable JavaScript in the webview
    enableScripts?: boolean;

    // Keep webview alive when hidden (uses more memory)
    retainContextWhenHidden?: boolean;

    // Control which resources can be loaded
    localResourceRoots?: vscode.Uri[];

    // Enable find widget in webview
    enableFindWidget?: boolean;

    // Enable command URIs
    enableCommandUris?: boolean;

    // Port mapping for localhost
    portMapping?: { webviewPort: number; extensionHostPort: number }[];
}
```

### Local Resource Roots

For security, webviews can only load resources from specific directories:

```typescript
const panel = vscode.window.createWebviewPanel(
  "myWebview",
  "My Webview",
  vscode.ViewColumn.One,
  {
    localResourceRoots: [
      vscode.Uri.joinPath(context.extensionUri, "media"),
      vscode.Uri.joinPath(context.extensionUri, "dist"),
    ],
  },
);
```

## Setting Webview Content

Set the `html` property to define the webview content:

```typescript
panel.webview.html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My Webview</title>
</head>
<body>
    <h1>Hello from Webview</h1>
    <p>This is a custom webview.</p>
</body>
</html>`;
```

### Loading Local Resources

Use `asWebviewUri` to load local files:

```typescript
const styleUri = panel.webview.asWebviewUri(
  vscode.Uri.joinPath(context.extensionUri, "media", "styles.css"),
);

const scriptUri = panel.webview.asWebviewUri(
  vscode.Uri.joinPath(context.extensionUri, "media", "main.js"),
);

panel.webview.html = `<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="${styleUri}">
</head>
<body>
    <h1>My Webview</h1>
    <script src="${scriptUri}"></script>
</body>
</html>`;
```

### Content Security Policy

Add a CSP for security:

```typescript
panel.webview.html = `<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${panel.webview.cspSource} 'unsafe-inline'; script-src ${panel.webview.cspSource};">
    <link rel="stylesheet" href="${styleUri}">
</head>
<body>
    <script src="${scriptUri}"></script>
</body>
</html>`;
```

## Communication Between Extension and Webview

### Extension to Webview

Send messages from the extension to the webview:

```typescript
// In the extension
panel.webview.postMessage({
  command: "update",
  data: { message: "Hello from extension" },
});

// In the webview JavaScript
window.addEventListener("message", (event) => {
  const message = event.data;
  if (message.command === "update") {
    console.log(message.data.message);
  }
});
```

### Webview to Extension

Send messages from the webview to the extension:

```typescript
// In the webview JavaScript
const vscode = acquireVsCodeApi();

document.getElementById("myButton").addEventListener("click", () => {
  vscode.postMessage({
    command: "buttonClicked",
    data: { value: 42 },
  });
});

// In the extension
panel.webview.onDidReceiveMessage(
  (message) => {
    switch (message.command) {
      case "buttonClicked":
        console.log("Button clicked:", message.data.value);
        break;
    }
  },
  undefined,
  context.subscriptions,
);
```

## The VSCode API

The `acquireVsCodeApi()` function returns an object with methods for webview-to-extension communication:

```typescript
const vscode = acquireVsCodeApi();

// Post a message to the extension
vscode.postMessage({ command: "myCommand" });

// Get and set state (persists across webview reloads)
vscode.setState({ count: 42 });
const state = vscode.getState(); // { count: 42 }
```

## Webview Lifecycle

### Handling Disposal

Clean up when the webview is closed:

```typescript
panel.onDidDispose(
  () => {
    // Clean up resources
    console.log("Webview disposed");
  },
  undefined,
  context.subscriptions,
);
```

### Visibility Changes

React to visibility changes:

```typescript
panel.onDidChangeViewState(
  (e) => {
    if (panel.visible) {
      // Webview became visible
      updateWebviewContent();
    }
  },
  undefined,
  context.subscriptions,
);
```

### Restoring Webviews

To restore webviews when VSCode restarts, use `serializeWebviewPanel`:

```typescript
vscode.window.registerWebviewPanelSerializer("myWebview", {
  async deserializeWebviewPanel(panel: vscode.WebviewPanel, state: any) {
    // Restore the webview with saved state
    panel.webview.html = getWebviewContent(state);
  },
});
```

Save state before disposal:

```typescript
// In the webview
vscode.setState({ lastContent: content });

// The state is passed to deserializeWebviewPanel on restore
```

## Styling Webviews

### Using VSCode's Theme Colors

Access VSCode's theme colors through CSS variables:

```css
body {
  background-color: var(--vscode-editor-background);
  color: var(--vscode-editor-foreground);
}

.button {
  background-color: var(--vscode-button-background);
  color: var(--vscode-button-foreground);
}

.button:hover {
  background-color: var(--vscode-button-hoverBackground);
}

.input {
  background-color: var(--vscode-input-background);
  border: 1px solid var(--vscode-input-border);
}
```

### Common CSS Variables

| Variable                      | Description             |
| ----------------------------- | ----------------------- |
| `--vscode-editor-background`  | Editor background color |
| `--vscode-editor-foreground`  | Editor text color       |
| `--vscode-button-background`  | Button background       |
| `--vscode-button-foreground`  | Button text color       |
| `--vscode-input-background`   | Input field background  |
| `--vscode-input-border`       | Input field border      |
| `--vscode-sideBar-background` | Sidebar background      |
| `--vscode-panel-background`   | Panel background        |

### Font Styling

Use VSCode's font settings:

```css
body {
  font-family: var(--vscode-font-family);
  font-size: var(--vscode-font-size);
  font-weight: var(--vscode-font-weight);
}

code {
  font-family: var(--vscode-editor-font-family);
}
```

## Webview Views

Webview views appear in the sidebar or panel instead of editor tabs.

### Registering a Webview View

```typescript
vscode.window.registerWebviewViewProvider(
  "myExtension.sidebarView",
  new MyWebviewViewProvider(context.extensionUri),
);

class MyWebviewViewProvider implements vscode.WebviewViewProvider {
  constructor(private readonly extensionUri: vscode.Uri) {}

  resolveWebviewView(
    webviewView: vscode.WebviewView,
    context: vscode.WebviewViewResolveContext,
    token: vscode.CancellationToken,
  ) {
    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this.extensionUri],
    };

    webviewView.webview.html = this.getHtmlContent(webviewView.webview);

    webviewView.webview.onDidReceiveMessage((message) => {
      // Handle messages
    });
  }
}
```

### Declaring in Manifest

```json
{
  "contributes": {
    "views": {
      "explorer": [
        {
          "type": "webview",
          "id": "myExtension.sidebarView",
          "name": "My View"
        }
      ]
    }
  }
}
```

## Custom Editors

Custom editors allow webviews to act as editors for specific file types.

### Registering a Custom Editor

```typescript
vscode.window.registerCustomEditorProvider(
  "myExtension.customEditor",
  new MyCustomEditorProvider(context.extensionUri),
  {
    supportsMultipleEditorsPerDocument: false,
  },
);

class MyCustomEditorProvider implements vscode.CustomEditorProvider {
  async openCustomDocument(
    uri: vscode.Uri,
    openContext: vscode.CustomDocumentOpenContext,
    token: vscode.CancellationToken,
  ): Promise<vscode.CustomDocument> {
    // Create the document
    return {
      uri,
      dispose: () => {},
    };
  }

  async resolveCustomEditor(
    document: vscode.CustomDocument,
    webviewPanel: vscode.WebviewPanel,
    token: vscode.CancellationToken,
  ): Promise<void> {
    // Set up the webview editor
    webviewPanel.webview.html = this.getEditorHtml(document.uri);
  }
}
```

### Declaring in Manifest

```json
{
  "contributes": {
    "customEditors": [
      {
        "viewType": "myExtension.customEditor",
        "displayName": "My Custom Editor",
        "selector": [
          {
            "filenamePattern": "*.myext"
          }
        ]
      }
    ]
  }
}
```

## Best Practices

**Use a CSP**: Always include a Content Security Policy to prevent security issues.

**Minimize resource loading**: Only load necessary CSS and JavaScript files.

**Handle disposal**: Clean up event listeners and resources when the webview is disposed.

**Persist state**: Use `setState` and `getState` to preserve user data across reloads.

**Test themes**: Verify your webview looks correct in both light and dark themes.

**Use async operations**: Webview communication is asynchronous; handle this appropriately.

**Validate messages**: Validate all messages received from the webview.

## Common Patterns

### React Webview

```typescript
// Build with webpack to output to dist/webview.js
panel.webview.html = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body>
    <div id="root"></div>
    <script src="${scriptUri}"></script>
</body>
</html>`;
```

### Data Loading

```typescript
// Extension sends data
async function loadData() {
  const data = await fetchData();
  panel.webview.postMessage({ command: "data", data });
}

// Webview receives and displays
window.addEventListener("message", (event) => {
  if (event.data.command === "data") {
    renderData(event.data.data);
  }
});
```

## Next Steps

- [Tree Views](tree-views.md): Create hierarchical views in the sidebar
- [Status Bar](status-bar.md): Add items to the status bar
- [Notifications](notifications.md): Show messages and progress

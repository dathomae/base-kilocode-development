# Tree Views

Tree views display hierarchical data in VSCode's sidebar. They are ideal for showing file structures, dependencies, or any data that benefits from a tree organization.

## Creating a Tree View

Tree views require two parts: declaring the view in the manifest and implementing a data provider.

### Declaring in the Manifest

Add the view to `package.json`:

```json
{
  "contributes": {
    "views": {
      "explorer": [
        {
          "id": "myExtension.filesView",
          "name": "My Files"
        }
      ]
    },
    "viewsContainers": {
      "activitybar": [
        {
          "id": "myExtension",
          "title": "My Extension",
          "icon": "resources/icon.svg"
        }
      ]
    }
  }
}
```

### View Containers

Views can be placed in different containers:

**explorer**: The Explorer sidebar
**debug**: The Debug sidebar
**scm**: The Source Control sidebar
**test**: The Test sidebar

**activitybar**: Custom container with its own icon in the activity bar

```json
{
  "viewsContainers": {
    "activitybar": [
      {
        "id": "myExtension",
        "title": "My Extension",
        "icon": "resources/icon.svg"
      }
    ]
  },
  "views": {
    "myExtension": [
      {
        "id": "myExtension.mainView",
        "name": "Main View"
      }
    ]
  }
}
```

## Implementing a Tree Data Provider

Implement the `TreeDataProvider` interface:

```typescript
import * as vscode from "vscode";

class MyTreeDataProvider implements vscode.TreeDataProvider<TreeItem> {
  private _onDidChangeTreeData: vscode.EventEmitter<
    TreeItem | undefined | null | void
  > = new vscode.EventEmitter<TreeItem | undefined | null | void>();

  readonly onDidChangeTreeData: vscode.Event<
    TreeItem | undefined | null | void
  > = this._onDidChangeTreeData.event;

  refresh(): void {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: TreeItem): vscode.TreeItem {
    return element;
  }

  getChildren(element?: TreeItem): Thenable<TreeItem[]> {
    if (element) {
      return Promise.resolve(element.children || []);
    }
    return Promise.resolve(this.getRootItems());
  }

  private getRootItems(): TreeItem[] {
    return [
      new TreeItem("Item 1", vscode.TreeItemCollapsibleState.Collapsed),
      new TreeItem("Item 2", vscode.TreeItemCollapsibleState.None),
    ];
  }
}
```

## Tree Items

`TreeItem` represents a node in the tree:

```typescript
class TreeItem extends vscode.TreeItem {
  children?: TreeItem[];

  constructor(
    public readonly label: string,
    public readonly collapsibleState: vscode.TreeItemCollapsibleState,
  ) {
    super(label, collapsibleState);
    this.tooltip = `${this.label}`;
    this.description = "Description";
  }
}
```

### TreeItem Properties

**label**: The display text (string or `TreeItemLabel` for highlights)

**collapsibleState**: Whether the item can be expanded

- `None`: No children, no expand arrow
- `Collapsed`: Has children, starts collapsed
- `Expanded`: Has children, starts expanded

**description**: Secondary text shown after the label

**tooltip**: Tooltip shown on hover

**iconPath**: Icon to display (file path, URI, or ThemeIcon)

**contextValue**: Context string for `when` clauses in menus

**command**: Command to execute when clicked

**checkboxState**: Optional checkbox state

### Icons

Use built-in theme icons:

```typescript
item.iconPath = new vscode.ThemeIcon("file");
item.iconPath = new vscode.ThemeIcon("folder");
item.iconPath = new vscode.ThemeIcon("symbol-method");
```

Or custom icons:

```typescript
item.iconPath = vscode.Uri.joinPath(
  context.extensionUri,
  "resources",
  "icon.svg",
);

// Light and dark variants
item.iconPath = {
  light: vscode.Uri.joinPath(
    context.extensionUri,
    "resources",
    "light",
    "icon.svg",
  ),
  dark: vscode.Uri.joinPath(
    context.extensionUri,
    "resources",
    "dark",
    "icon.svg",
  ),
};
```

### Commands on Click

Execute a command when the item is clicked:

```typescript
item.command = {
  command: "myExtension.openFile",
  title: "Open File",
  arguments: [item.uri],
};
```

## Registering the Tree View

Register the tree data provider:

```typescript
export function activate(context: vscode.ExtensionContext) {
  const treeDataProvider = new MyTreeDataProvider();

  vscode.window.createTreeView("myExtension.filesView", {
    treeDataProvider: treeDataProvider,
    showCollapseAll: true,
  });
}
```

## Refreshing the Tree

Trigger a refresh when data changes:

```typescript
class MyTreeDataProvider implements vscode.TreeDataProvider<TreeItem> {
  // ... other methods

  refresh(item?: TreeItem): void {
    this._onDidChangeTreeData.fire(item);
  }
}

// Usage
treeDataProvider.refresh(); // Refresh entire tree
treeDataProvider.refresh(item); // Refresh specific item
```

### Auto-refresh on Changes

```typescript
vscode.workspace.onDidChangeConfiguration((e) => {
  if (e.affectsConfiguration("myExtension")) {
    treeDataProvider.refresh();
  }
});

vscode.workspace.onDidSaveTextDocument(() => {
  treeDataProvider.refresh();
});
```

## Parent-Child Navigation

Implement `getParent` for keyboard navigation:

```typescript
class MyTreeDataProvider implements vscode.TreeDataProvider<TreeItem> {
  getParent(element: TreeItem): vscode.ProviderResult<TreeItem> {
    return element.parent;
  }
}
```

## Drag and Drop

Implement drag and drop support:

```typescript
class MyTreeDataProvider implements vscode.TreeDataProvider<TreeItem> {
  dragMimeTypes: string[] = ["text/uri-list"];
  dropMimeTypes: string[] = ["text/uri-list"];

  handleDrag?(
    source: TreeItem[],
    dataTransfer: vscode.DataTransfer,
    token: vscode.CancellationToken,
  ): void | Thenable<void> {
    dataTransfer.set(
      "text/uri-list",
      source.map((item) => item.uri.toString()).join("\r\n"),
    );
  }

  handleDrop?(
    target: TreeItem | undefined,
    dataTransfer: vscode.DataTransfer,
    token: vscode.CancellationToken,
  ): void | Thenable<void> {
    const uris = dataTransfer.get("text/uri-list");
    // Handle the drop
  }
}
```

Register with drag and drop:

```typescript
vscode.window.createTreeView("myExtension.view", {
  treeDataProvider: provider,
  dragAndDropController: provider,
});
```

## Tree Item Checkbox

Enable checkboxes for multi-selection:

```typescript
class MyTreeDataProvider implements vscode.TreeDataProvider<TreeItem> {
  onDidChangeTreeData: vscode.Event<TreeItem | undefined>;

  getTreeItem(element: TreeItem): vscode.TreeItem {
    element.checkboxState = element.checked
      ? vscode.TreeItemCheckboxState.Checked
      : vscode.TreeItemCheckboxState.Unchecked;
    return element;
  }
}

// Handle checkbox changes
treeView.onDidChangeCheckboxState((event) => {
  for (const [item, state] of event.items) {
    item.checked = state === vscode.TreeItemCheckboxState.Checked;
  }
});
```

## Context Menus

Add context menu items for tree items:

```json
{
  "contributes": {
    "menus": {
      "view/item/context": [
        {
          "command": "myExtension.delete",
          "when": "view == myExtension.filesView && viewItem == file",
          "group": "modification"
        }
      ]
    }
  }
}
```

Use `contextValue` to differentiate items:

```typescript
item.contextValue = "file"; // Matches "viewItem == file"
item.contextValue = "folder"; // Matches "viewItem == folder"
```

## View Title Actions

Add buttons to the view title:

```json
{
  "contributes": {
    "menus": {
      "view/title": [
        {
          "command": "myExtension.refresh",
          "when": "view == myExtension.filesView",
          "group": "navigation"
        }
      ]
    }
  }
}
```

## Tree View Visibility

Check and control visibility:

```typescript
const treeView = vscode.window.createTreeView("myExtension.view", {
  treeDataProvider: provider,
});

// Check if visible
if (treeView.visible) {
  // View is visible
}

// React to visibility changes
treeView.onDidChangeVisibility((e) => {
  if (e.visible) {
    // View became visible
  }
});

// Reveal an item
await treeView.reveal(item, {
  select: true,
  focus: true,
  expand: true,
});
```

## Best Practices

**Lazy load children**: Only load children when the item is expanded, not when the tree is created.

**Use collapsibleState appropriately**: Use `None` for leaf items, `Collapsed` for items with children.

**Provide meaningful descriptions**: Show additional context in the description field.

**Use contextValue**: Enable different context menu actions for different item types.

**Handle empty states**: Show a helpful message when the tree is empty.

**Refresh efficiently**: Only refresh the parts of the tree that changed.

## Common Patterns

### File System Tree

```typescript
class FileTreeDataProvider implements vscode.TreeDataProvider<FileItem> {
  constructor(private rootPath: string) {}

  getTreeItem(element: FileItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: FileItem): Promise<FileItem[]> {
    const folder = element ? element.uri.fsPath : this.rootPath;
    const entries = await fs.promises.readdir(folder, { withFileTypes: true });

    return entries.map((entry) => {
      const uri = vscode.Uri.file(path.join(folder, entry.name));
      const item = new FileItem(entry.name, uri);

      if (entry.isDirectory()) {
        item.collapsibleState = vscode.TreeItemCollapsibleState.Collapsed;
        item.iconPath = vscode.ThemeIcon.Folder;
      } else {
        item.collapsibleState = vscode.TreeItemCollapsibleState.None;
        item.iconPath = vscode.ThemeIcon.File;
      }

      return item;
    });
  }
}
```

### Dependency Tree

```typescript
class DependencyTreeDataProvider implements vscode.TreeDataProvider<DependencyItem> {
  async getChildren(element?: DependencyItem): Promise<DependencyItem[]> {
    if (!element) {
      // Root: get direct dependencies
      return this.getDirectDependencies();
    }

    // Get transitive dependencies
    return this.getTransitiveDependencies(element);
  }

  private async getDirectDependencies(): Promise<DependencyItem[]> {
    const packageJson = await this.readPackageJson();
    return Object.entries(packageJson.dependencies).map(([name, version]) => {
      const item = new DependencyItem(name, version);
      item.contextValue = "dependency";
      return item;
    });
  }
}
```

## Next Steps

- [Webviews](webviews.md): Create custom HTML-based views
- [Status Bar](status-bar.md): Add items to the status bar
- [Menus](../02-core-capabilities/menus.md): Add context menus to tree items

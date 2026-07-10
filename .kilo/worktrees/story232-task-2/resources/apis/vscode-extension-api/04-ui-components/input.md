# User Input

Extensions often need to collect input from users. VSCode provides several APIs for different input scenarios, from simple text input to file selection and custom quick picks.

## Input Box

The input box collects text input from users:

```typescript
const name = await vscode.window.showInputBox({
  prompt: "Enter a name",
  placeHolder: "Name",
});

if (name) {
  console.log("Entered:", name);
}
```

### Input Box Options

```typescript
interface InputBoxOptions {
  // Title shown above the input
  title?: string;

  // Text shown below the input
  prompt?: string;

  // Placeholder text in the input field
  placeHolder?: string;

  // Initial value
  value?: string;

  // Selection range for the initial value [start, end]
  valueSelection?: [number, number];

  // Mask input for passwords
  password?: boolean;

  // Keep open when focus moves away
  ignoreFocusOut?: boolean;

  // Validate input and return error message or undefined
  validateInput?: (
    value: string,
  ) => string | undefined | Thenable<string | undefined>;

  // Buttons to show below the input
  buttons?: vscode.QuickInputButton[];
}
```

### Validation

Validate input in real-time:

```typescript
const filename = await vscode.window.showInputBox({
  prompt: "Enter a filename",
  validateInput: (value) => {
    if (!value) {
      return "Filename is required";
    }
    if (value.includes("/")) {
      return "Filename cannot contain /";
    }
    if (value.includes("\\")) {
      return "Filename cannot contain \\";
    }
    if (/^\./.test(value)) {
      return "Filename cannot start with .";
    }
    return undefined; // Valid
  },
});
```

### Async Validation

Perform asynchronous validation:

```typescript
const username = await vscode.window.showInputBox({
  prompt: "Enter a username",
  validateInput: async (value) => {
    if (!value) {
      return "Username is required";
    }

    // Check if username exists
    const exists = await checkUsernameExists(value);
    if (exists) {
      return "Username already exists";
    }

    return undefined;
  },
});
```

### Password Input

Mask input for sensitive data:

```typescript
const password = await vscode.window.showInputBox({
  prompt: "Enter your password",
  password: true,
});
```

### Pre-filled Input

Provide an initial value with selection:

```typescript
const result = await vscode.window.showInputBox({
  value: "default-name",
  valueSelection: [0, 12], // Select "default-name"
});
```

## Quick Pick

Quick pick shows a list of options for selection:

```typescript
const option = await vscode.window.showQuickPick(
  ["Option 1", "Option 2", "Option 3"],
  {
    placeHolder: "Select an option",
  },
);
```

### Quick Pick Options

```typescript
interface QuickPickOptions {
  // Title shown above the list
  title?: string;

  // Placeholder text
  placeHolder?: string;

  // Match on description when filtering
  matchOnDescription?: boolean;

  // Match on detail when filtering
  matchOnDetail?: boolean;

  // Allow multiple selection
  canPickMany?: boolean;

  // Keep open when focus moves away
  ignoreFocusOut?: boolean;

  // Called when selection changes
  onDidSelectItem?: (item: QuickPickItem) => void;
}
```

### Quick Pick Items

Use `QuickPickItem` for rich items:

```typescript
interface QuickPickItem {
  // Main label
  label: string;

  // Description shown after label
  description?: string;

  // Detail shown below label
  detail?: string;

  // Whether the item is picked (for multi-select)
  picked?: boolean;

  // Whether the item is always visible
  alwaysShow?: boolean;

  // Buttons shown on the item
  buttons?: QuickInputButton[];
}
```

### Custom Quick Pick Items

```typescript
interface MyItem extends vscode.QuickPickItem {
  id: string;
  data: any;
}

const items: MyItem[] = [
  {
    label: "$(file) File 1",
    description: "/path/to/file1",
    detail: "Last modified: today",
    id: "file1",
    data: { size: 1024 },
  },
  {
    label: "$(file) File 2",
    description: "/path/to/file2",
    detail: "Last modified: yesterday",
    id: "file2",
    data: { size: 2048 },
  },
];

const selection = await vscode.window.showQuickPick(items, {
  placeHolder: "Select a file",
  matchOnDescription: true,
  matchOnDetail: true,
});

if (selection) {
  console.log("Selected:", selection.id, selection.data);
}
```

### Multi-Select

Allow selecting multiple items:

```typescript
const selections = await vscode.window.showQuickPick(
  [
    { label: "Option 1", picked: true },
    { label: "Option 2", picked: false },
    { label: "Option 3", picked: true },
  ],
  {
    canPickMany: true,
    placeHolder: "Select options",
  },
);

if (selections) {
  selections.forEach((s) => console.log(s.label));
}
```

### Separators

Group items with separators:

```typescript
const items = [
  { label: "Recent Files", kind: vscode.QuickPickItemKind.Separator },
  { label: "file1.txt" },
  { label: "file2.txt" },
  { label: "Other Files", kind: vscode.QuickPickItemKind.Separator },
  { label: "file3.txt" },
  { label: "file4.txt" },
];

const selection = await vscode.window.showQuickPick(items);
```

## Custom Quick Input

For more complex input scenarios, create a custom quick input:

```typescript
const quickPick = vscode.window.createQuickPick();
quickPick.placeholder = "Select items";
quickPick.canSelectMany = true;
quickPick.items = [
  { label: "Item 1" },
  { label: "Item 2" },
  { label: "Item 3" },
];

quickPick.onDidAccept(() => {
  const selections = quickPick.selectedItems;
  console.log("Selected:", selections);
  quickPick.hide();
});

quickPick.onDidChangeSelection((selection) => {
  console.log("Selection changed:", selection);
});

quickPick.onDidChangeValue((value) => {
  console.log("Filter value:", value);
  // Update items based on filter
});

quickPick.show();
```

### Custom Quick Input Features

```typescript
const quickPick = vscode.window.createQuickPick();

// Set properties
quickPick.title = "Select Files";
quickPick.placeholder = "Type to search";
quickPick.matchOnDescription = true;
quickPick.matchOnDetail = true;
quickPick.ignoreFocusOut = true;

// Add buttons
quickPick.buttons = [
  vscode.QuickInputButtons.Back,
  { iconPath: vscode.ThemeIcon.File, tooltip: "Add File" },
];

// Handle events
quickPick.onDidTriggerButton((button) => {
  if (button === vscode.QuickInputButtons.Back) {
    // Handle back
  }
});

quickPick.onDidHide(() => {
  quickPick.dispose();
});

quickPick.show();
```

## File and Folder Selection

### Open File Dialog

```typescript
const uris = await vscode.window.showOpenDialog({
  canSelectMany: true,
  canSelectFiles: true,
  canSelectFolders: false,
  openLabel: "Select",
  filters: {
    JavaScript: ["js", "jsx"],
    TypeScript: ["ts", "tsx"],
    "All Files": ["*"],
  },
});

if (uris) {
  uris.forEach((uri) => console.log(uri.fsPath));
}
```

### Save File Dialog

```typescript
const uri = await vscode.window.showSaveDialog({
  defaultUri: vscode.Uri.file("untitled.txt"),
  saveLabel: "Export",
  filters: {
    "Text Files": ["txt"],
    "JSON Files": ["json"],
  },
});

if (uri) {
  await saveToFile(uri);
}
```

### Folder Selection

```typescript
const uris = await vscode.window.showOpenDialog({
  canSelectFiles: false,
  canSelectFolders: true,
  canSelectMany: false,
  openLabel: "Select Folder",
});

if (uris && uris.length > 0) {
  console.log("Selected folder:", uris[0].fsPath);
}
```

### Workspace Folder Selection

```typescript
const folder = await vscode.window.showWorkspaceFolderPick({
  placeHolder: "Select a workspace folder",
});

if (folder) {
  console.log("Selected:", folder.uri.fsPath);
}
```

## Best Practices

**Provide clear prompts**: Users should understand what input is expected.

**Use validation**: Validate input immediately and provide helpful error messages.

**Offer sensible defaults**: Pre-fill values when appropriate.

**Use appropriate input types**: Quick pick for selections, input box for free text, dialogs for files.

**Handle cancellation**: Check for undefined results when users cancel.

**Keep it simple**: Don't ask for more information than necessary.

## Common Patterns

### Yes/No Confirmation

```typescript
const result = await vscode.window.showQuickPick(["Yes", "No"], {
  placeHolder: "Are you sure?",
});

return result === "Yes";
```

### File Type Selection

```typescript
const fileTypes = [
  { label: "JavaScript", description: ".js files", extension: "js" },
  { label: "TypeScript", description: ".ts files", extension: "ts" },
  { label: "Python", description: ".py files", extension: "py" },
];

const selection = await vscode.window.showQuickPick(fileTypes, {
  placeHolder: "Select file type",
});

if (selection) {
  console.log("Extension:", selection.extension);
}
```

### Search and Select

```typescript
const quickPick = vscode.window.createQuickPick();
quickPick.placeholder = "Search files...";

quickPick.onDidChangeValue(async (value) => {
  if (value) {
    const files = await searchFiles(value);
    quickPick.items = files.map((f) => ({
      label: f.name,
      description: f.path,
    }));
  } else {
    quickPick.items = [];
  }
});

quickPick.onDidAccept(() => {
  const selection = quickPick.selectedItems[0];
  if (selection) {
    openFile(selection.description);
  }
  quickPick.hide();
});

quickPick.show();
```

### Step-by-Step Input

```typescript
async function collectInput(): Promise<void> {
  // Step 1: Get name
  const name = await vscode.window.showInputBox({
    prompt: "Enter a name",
  });
  if (!name) return;

  // Step 2: Get type
  const type = await vscode.window.showQuickPick(
    ["Type A", "Type B", "Type C"],
    { placeHolder: "Select a type" },
  );
  if (!type) return;

  // Step 3: Get location
  const uri = await vscode.window.showSaveDialog({
    saveLabel: "Create",
    filters: { Files: ["txt"] },
  });
  if (!uri) return;

  // Create with collected input
  await createFile(name, type, uri);
}
```

## Next Steps

- [Notifications](notifications.md): Show messages and progress
- [Commands](../02-core-capabilities/commands.md): Create commands that collect input
- [Configuration](../02-core-capabilities/configuration.md): Store user preferences

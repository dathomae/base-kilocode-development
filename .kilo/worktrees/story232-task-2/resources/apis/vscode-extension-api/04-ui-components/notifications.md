# Notifications and Progress

VSCode provides several ways to communicate with users through messages, notifications, and progress indicators. These help users understand what the extension is doing and make informed decisions.

## Message Types

### Information Messages

Show informational messages for non-critical notifications:

```typescript
vscode.window.showInformationMessage("Operation completed successfully");
```

### Warning Messages

Show warnings for potential issues:

```typescript
vscode.window.showWarningMessage("This action cannot be undone");
```

### Error Messages

Show errors for failures and problems:

```typescript
vscode.window.showErrorMessage("Failed to connect to server");
```

## Message with Actions

Add buttons to messages for user interaction:

```typescript
const selection = await vscode.window.showInformationMessage(
  "Do you want to save changes?",
  "Save",
  "Discard",
  "Cancel",
);

if (selection === "Save") {
  await saveChanges();
} else if (selection === "Discard") {
  discardChanges();
}
```

### Action Items

Use objects for more control over actions:

```typescript
const selection = await vscode.window.showInformationMessage(
  "File has been modified externally. Reload?",
  {
    modal: true,
    detail:
      "The file has been changed on disk. Your changes will be lost if you reload.",
  },
  "Reload",
  "Keep Current",
);
```

### Modal Messages

Modal messages require user interaction before continuing:

```typescript
const result = await vscode.window.showWarningMessage(
  "Are you sure you want to delete this file?",
  {
    modal: true,
    detail: "This action cannot be undone.",
  },
  "Delete",
  "Cancel",
);
```

## Progress Indicators

Show progress for long-running operations:

### Window Progress

Progress shown in the status bar:

```typescript
await vscode.window.withProgress(
  {
    location: vscode.ProgressLocation.Window,
    title: "Processing files",
    cancellable: true,
  },
  async (progress, token) => {
    progress.report({ message: "Starting..." });

    for (let i = 0; i < files.length; i++) {
      if (token.isCancellationRequested) {
        break;
      }

      await processFile(files[i]);
      progress.report({
        message: `Processing ${i + 1}/${files.length}`,
        increment: 100 / files.length,
      });
    }
  },
);
```

### Notification Progress

Progress shown as a notification:

```typescript
await vscode.window.withProgress(
  {
    location: vscode.ProgressLocation.Notification,
    title: "Downloading files",
    cancellable: true,
  },
  async (progress, token) => {
    progress.report({ increment: 0 });

    for (const file of files) {
      if (token.isCancellationRequested) {
        throw new vscode.CancellationError();
      }

      await downloadFile(file);
      progress.report({ increment: 100 / files.length });
    }
  },
);
```

### View Progress

Progress shown in a specific view:

```typescript
await vscode.window.withProgress(
  {
    location: { viewId: "myExtension.view" },
  },
  async (progress) => {
    // Progress shown in the view
  },
);
```

### Progress Options

```typescript
interface ProgressOptions {
  // Where to show the progress
  location: ProgressLocation | { viewId: string };

  // Title shown with the progress
  title?: string;

  // Whether the operation can be cancelled
  cancellable?: boolean;
}
```

### Progress Report

```typescript
progress.report({
  message: "Current step description", // Optional message
  increment: 10, // Optional percentage increment (0-100)
});
```

## Background Tasks

For operations that should continue in the background:

```typescript
vscode.window.withProgress(
  {
    location: vscode.ProgressLocation.Window,
    title: "Background sync",
  },
  async () => {
    // This runs in the background
    await performSync();
  },
);

// Code continues immediately
console.log("Sync started");
```

## Status Bar Messages

Show temporary messages in the status bar:

```typescript
// Show for 5 seconds
vscode.window.setStatusBarMessage("$(check) Operation completed", 5000);

// Show until the returned disposable is disposed
const disposable = vscode.window.setStatusBarMessage(
  "$(sync~spin) Processing...",
);
await doWork();
disposable.dispose();
```

## Quick Pick

Show a list of options for the user to select:

```typescript
const selection = await vscode.window.showQuickPick(
  ["Option 1", "Option 2", "Option 3"],
  {
    placeHolder: "Select an option",
    title: "Choose wisely",
  },
);

if (selection) {
  console.log("Selected:", selection);
}
```

### Quick Pick with Objects

Use objects for more complex items:

```typescript
interface MyItem extends vscode.QuickPickItem {
  id: string;
  value: number;
}

const items: MyItem[] = [
  { label: "Option 1", description: "First option", id: "opt1", value: 1 },
  { label: "Option 2", description: "Second option", id: "opt2", value: 2 },
  { label: "Option 3", description: "Third option", id: "opt3", value: 3 },
];

const selection = await vscode.window.showQuickPick(items, {
  placeHolder: "Select an option",
});

if (selection) {
  console.log("Selected:", selection.id, selection.value);
}
```

### Multi-Select

Allow selecting multiple items:

```typescript
const selections = await vscode.window.showQuickPick(
  ["Option 1", "Option 2", "Option 3"],
  {
    canPickMany: true,
    placeHolder: "Select multiple options",
  },
);

if (selections) {
  console.log("Selected:", selections);
}
```

### Quick Pick Options

```typescript
interface QuickPickOptions {
  title?: string;
  placeHolder?: string;
  matchOnDescription?: boolean; // Match on description text
  matchOnDetail?: boolean; // Match on detail text
  canPickMany?: boolean; // Allow multiple selection
  ignoreFocusOut?: boolean; // Keep open when focus is lost
  onDidSelectItem?: (item: any) => void; // Called on selection change
}
```

## Input Box

Get text input from the user:

```typescript
const input = await vscode.window.showInputBox({
  prompt: "Enter a name for the file",
  placeHolder: "filename.txt",
  value: "default.txt",
  validateInput: (text) => {
    if (text.includes("/")) {
      return "Filename cannot contain /";
    }
    if (text.length === 0) {
      return "Filename is required";
    }
    return null; // Valid input
  },
});

if (input) {
  console.log("Entered:", input);
}
```

### Input Box Options

```typescript
interface InputBoxOptions {
  title?: string;
  prompt?: string; // Text below the input
  placeHolder?: string; // Placeholder text
  value?: string; // Initial value
  valueSelection?: [number, number]; // Selection range
  password?: boolean; // Mask input
  ignoreFocusOut?: boolean; // Keep open when focus is lost
  validateInput?: (value: string) => string | undefined;
  prompt?: string;
}
```

## Open File Dialog

Let users select files:

```typescript
const uris = await vscode.window.showOpenDialog({
  canSelectMany: true,
  openLabel: "Select Files",
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

### Open Dialog Options

```typescript
interface OpenDialogOptions {
  canSelectMany?: boolean;
  canSelectFiles?: boolean;
  canSelectFolders?: boolean;
  defaultUri?: vscode.Uri;
  openLabel?: string;
  title?: string;
  filters?: { [name: string]: string[] };
}
```

## Save File Dialog

Let users choose where to save:

```typescript
const uri = await vscode.window.showSaveDialog({
  defaultUri: vscode.Uri.file("untitled.txt"),
  filters: {
    "Text Files": ["txt"],
    "All Files": ["*"],
  },
});

if (uri) {
  await saveFile(uri);
}
```

## Workspace Folder Picker

Let users select a workspace folder:

```typescript
const folder = await vscode.window.showWorkspaceFolderPick({
  placeHolder: "Select a folder",
});

if (folder) {
  console.log("Selected folder:", folder.uri.fsPath);
}
```

## Best Practices

**Use appropriate message types**: Information for success, warning for potential issues, error for failures.

**Provide actionable options**: When showing messages, include buttons for common actions.

**Show progress for long operations**: Keep users informed about what's happening.

**Allow cancellation**: Make long operations cancellable when possible.

**Validate input**: Use `validateInput` to provide immediate feedback.

**Keep messages concise**: Users often dismiss messages quickly. Make the important information obvious.

**Don't overuse notifications**: Reserve notifications for important information that requires attention.

## Common Patterns

### Confirmation Dialog

```typescript
async function confirmAction(message: string): Promise<boolean> {
  const result = await vscode.window.showWarningMessage(
    message,
    { modal: true },
    "Yes",
    "No",
  );
  return result === "Yes";
}

if (await confirmAction("Delete this file?")) {
  await deleteFile();
}
```

### Progress with Cancellation

```typescript
async function processFiles(files: string[]): Promise<void> {
  await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: "Processing files",
      cancellable: true,
    },
    async (progress, token) => {
      for (let i = 0; i < files.length; i++) {
        if (token.isCancellationRequested) {
          vscode.window.showInformationMessage("Operation cancelled");
          return;
        }

        progress.report({
          message: `File ${i + 1} of ${files.length}`,
          increment: 100 / files.length,
        });

        await processFile(files[i]);
      }
    },
  );
}
```

### Validated Input

```typescript
const email = await vscode.window.showInputBox({
  prompt: "Enter your email address",
  placeHolder: "user@example.com",
  validateInput: (value) => {
    if (!value) {
      return "Email is required";
    }
    if (!value.includes("@")) {
      return "Please enter a valid email address";
    }
    return undefined;
  },
});
```

## Next Steps

- [Input](input.md): More details on collecting user input
- [Status Bar](status-bar.md): Add persistent status information
- [Commands](../02-core-capabilities/commands.md): Create commands for user actions

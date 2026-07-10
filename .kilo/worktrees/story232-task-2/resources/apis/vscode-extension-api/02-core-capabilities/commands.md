# Commands

Commands are the primary way users interact with extensions. A command is an action that users can invoke through the Command Palette, keybindings, menus, or programmatically.

## Registering Commands

Commands are registered using the `vscode.commands.registerCommand` function:

```typescript
import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  const disposable = vscode.commands.registerCommand(
    "myExtension.hello",
    () => {
      vscode.window.showInformationMessage("Hello World!");
    },
  );

  context.subscriptions.push(disposable);
}
```

### Command Identifiers

Command identifiers must be unique across all extensions. Use a naming convention that includes your extension name:

```typescript
// Good: includes extension prefix
"myExtension.formatDocument";
"myExtension.openSettings";

// Avoid: generic names that might conflict
"format";
"openSettings";
```

### Command Handler

The command handler is a function that executes when the command is invoked:

```typescript
const disposable = vscode.commands.registerCommand(
  "myExtension.greet",
  (name: string) => {
    vscode.window.showInformationMessage(`Hello, ${name}!`);
  },
);
```

Handlers can accept arguments passed when the command is invoked.

## Declaring Commands in the Manifest

Commands must be declared in `package.json` to appear in the Command Palette:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "myExtension.hello",
        "title": "Say Hello",
        "category": "My Extension"
      }
    ]
  }
}
```

### Command Properties

**command**: The unique identifier used in `registerCommand`

**title**: The display name shown in the Command Palette

**category**: Optional grouping for related commands. Commands with the same category appear grouped together.

**icon**: Optional icon for the command when displayed in menus:

```json
{
  "command": "myExtension.refresh",
  "title": "Refresh",
  "icon": {
    "light": "resources/light/refresh.svg",
    "dark": "resources/dark/refresh.svg"
  }
}
```

**enablement**: Optional condition for when the command is enabled:

```json
{
  "command": "myExtension.format",
  "title": "Format Document",
  "enablement": "editorLangId == javascript"
}
```

## Command Arguments

Commands can receive arguments when invoked programmatically:

```typescript
// Register a command that accepts arguments
vscode.commands.registerCommand("myExtension.openFile", (uri: vscode.Uri) => {
  vscode.window.showTextDocument(uri);
});

// Execute the command with an argument
vscode.commands.executeCommand(
  "myExtension.openFile",
  vscode.Uri.file("/path/to/file.txt"),
);
```

### Multiple Arguments

Commands can accept multiple arguments:

```typescript
vscode.commands.registerCommand(
  "myExtension.replace",
  (range: vscode.Range, text: string) => {
    const editor = vscode.window.activeTextEditor;
    if (editor) {
      editor.edit((editBuilder) => {
        editBuilder.replace(range, text);
      });
    }
  },
);

// Execute with multiple arguments
vscode.commands.executeCommand(
  "myExtension.replace",
  new vscode.Range(0, 0, 0, 10),
  "replacement text",
);
```

## Executing Commands

### From Code

Execute commands programmatically using `executeCommand`:

```typescript
// Execute a built-in command
await vscode.commands.executeCommand("workbench.action.files.save");

// Execute with arguments
await vscode.commands.executeCommand("editor.action.addComment", {
  lines: [0, 1],
});

// Execute and get a result
const uris = await vscode.commands.executeCommand<vscode.Uri[]>(
  "workbench.action.files.openFile",
);
```

### From the Command Palette

Users can invoke commands through the Command Palette (`Ctrl+Shift+P` or `Cmd+Shift+P`). Commands declared in the manifest appear automatically.

### From Keybindings

Commands can be bound to keyboard shortcuts. See [keybindings.md](keybindings.md) for details.

### From Menus

Commands can be added to various menus. See [menus.md](menus.md) for details.

## Built-in Commands

VSCode provides many built-in commands that extensions can execute:

### Editor Commands

| Command                             | Description                 |
| ----------------------------------- | --------------------------- |
| `editor.action.formatDocument`      | Format the current document |
| `editor.action.organizeImports`     | Organize imports            |
| `editor.action.addComment`          | Add line comment            |
| `editor.action.removeComment`       | Remove line comment         |
| `editor.action.copyLinesDownAction` | Copy line down              |
| `editor.action.moveLinesDownAction` | Move line down              |

### File Commands

| Command                           | Description           |
| --------------------------------- | --------------------- |
| `workbench.action.files.save`     | Save the current file |
| `workbench.action.files.saveAll`  | Save all files        |
| `workbench.action.files.openFile` | Open a file           |
| `workbench.action.files.newFile`  | Create a new file     |

### Navigation Commands

| Command                           | Description               |
| --------------------------------- | ------------------------- |
| `workbench.action.quickOpen`      | Quick open                |
| `workbench.action.gotoSymbol`     | Go to symbol in file      |
| `workbench.action.showAllSymbols` | Go to symbol in workspace |

### View Commands

| Command                          | Description               |
| -------------------------------- | ------------------------- |
| `workbench.action.toggleSidebar` | Toggle sidebar visibility |
| `workbench.action.focusSideBar`  | Focus the sidebar         |
| `workbench.action.terminal.new`  | Create a new terminal     |

## Command Enablement

Commands can be conditionally enabled or disabled using the `enablement` property:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "myExtension.format",
        "title": "Format",
        "enablement": "editorLangId == javascript && editorHasSelection"
      }
    ]
  }
}
```

### When Clause Contexts

The `enablement` clause uses the same syntax as keybinding `when` clauses:

**Editor contexts**:

- `editorLangId`: The language ID of the current file
- `editorHasSelection`: Whether text is selected
- `editorTextFocus`: Whether a text editor has focus
- `editorReadonly`: Whether the editor is read-only

**Resource contexts**:

- `resourceFilename`: The file name of the current resource
- `resourceExtname`: The file extension
- `resourceDirname`: The directory path

**Workspace contexts**:

- `workspaceFolderCount`: Number of workspace folders
- `workspaceIsTrusted`: Whether the workspace is trusted

**Custom contexts**:

- `myExtension:isEnabled`: Custom context set by your extension

### Setting Custom Contexts

Extensions can set custom contexts for use in enablement clauses:

```typescript
// Set a custom context
vscode.commands.executeCommand("setContext", "myExtension.hasProject", true);

// Use in enablement
// "enablement": "myExtension.hasProject"
```

## Text Editor Commands

For commands that operate on text editors, use `registerTextEditorCommand`:

```typescript
const disposable = vscode.commands.registerTextEditorCommand(
  "myExtension.toUpperCase",
  (editor, edit) => {
    const selection = editor.selection;
    const text = editor.document.getText(selection);
    edit.replace(selection, text.toUpperCase());
  },
);
```

The handler receives:

- `editor`: The active text editor
- `edit`: A text edit builder for making changes

This command is automatically disabled when no text editor is active.

## Progress and Feedback

### Showing Progress

For long-running commands, show a progress indicator:

```typescript
vscode.commands.registerCommand("myExtension.analyze", async () => {
  await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: "Analyzing...",
      cancellable: true,
    },
    async (progress, token) => {
      progress.report({ message: "Parsing files..." });
      await parseFiles();

      progress.report({ message: "Running analysis...", increment: 50 });
      await runAnalysis(token);

      progress.report({ message: "Complete!", increment: 50 });
    },
  );
});
```

### Showing Messages

Display feedback to users:

```typescript
// Information message
vscode.window.showInformationMessage("Operation completed");

// Warning message
vscode.window.showWarningMessage("This action cannot be undone");

// Error message
vscode.window.showErrorMessage("Failed to connect to server");
```

### Requesting User Input

Get input from users:

```typescript
// Quick pick selection
const option = await vscode.window.showQuickPick(
  ["Option 1", "Option 2", "Option 3"],
  { placeHolder: "Select an option" },
);

// Text input
const name = await vscode.window.showInputBox({
  placeHolder: "Enter your name",
  prompt: "This will be used for the greeting",
});

// Confirmation
const confirm = await vscode.window.showWarningMessage(
  "Are you sure?",
  "Yes",
  "No",
);
```

## Command Palette Behavior

### Hiding Commands from Palette

To prevent a command from appearing in the Command Palette, set `title` to an empty string:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "myExtension.internal",
        "title": ""
      }
    ]
  }
}
```

This is useful for commands that should only be invoked through keybindings or menus.

### Command Groups

Commands with the same category are grouped in the Command Palette:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "myExtension.format",
        "title": "Format",
        "category": "Formatting"
      },
      {
        "command": "myExtension.sort",
        "title": "Sort Imports",
        "category": "Formatting"
      }
    ]
  }
}
```

## Best Practices

**Use descriptive titles**: Command titles should clearly describe what the command does. Use verbs like "Format", "Analyze", "Open".

**Group related commands**: Use categories to organize related commands in the Command Palette.

**Handle errors gracefully**: Wrap command logic in try-catch blocks and show meaningful error messages.

**Provide feedback**: Show progress for long operations and confirm completion.

**Enable commands conditionally**: Use enablement clauses to disable commands when they cannot run.

**Keep commands focused**: Each command should do one thing well. Create multiple commands for different actions rather than one command with complex behavior.

## Next Steps

- [Menus](menus.md): Add commands to context menus and other UI surfaces
- [Keybindings](keybindings.md): Define keyboard shortcuts for commands
- [Configuration](configuration.md): Add settings for command behavior

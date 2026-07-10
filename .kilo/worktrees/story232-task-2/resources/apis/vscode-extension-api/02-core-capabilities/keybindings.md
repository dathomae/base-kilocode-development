# Keybindings

Keybindings allow users to invoke commands through keyboard shortcuts. Extensions can define default keybindings and users can customize them through VSCode's keyboard shortcut settings.

## Declaring Keybindings

Keybindings are declared in the `package.json` manifest:

```json
{
  "contributes": {
    "keybindings": [
      {
        "command": "myExtension.format",
        "key": "ctrl+shift+f",
        "mac": "cmd+shift+f",
        "when": "editorTextFocus"
      }
    ]
  }
}
```

## Keybinding Properties

### command

The command to execute when the keybinding is triggered. The command must be registered either by your extension or by VSCode.

### key

The key combination for Windows and Linux. Use `ctrl`, `shift`, `alt`, and `meta` (Windows key) as modifiers:

```json
{ "key": "ctrl+shift+f" }
{ "key": "alt+enter" }
{ "key": "ctrl+k ctrl+f" }  // Chord: press Ctrl+K, then Ctrl+F
```

### mac

The key combination for macOS. Use `cmd`, `shift`, `alt`, and `ctrl` as modifiers:

```json
{ "mac": "cmd+shift+f" }
{ "mac": "alt+enter" }
{ "mac": "cmd+k cmd+f" }
```

### macLinux

A keybinding that applies to both macOS and Linux:

```json
{
  "command": "myExtension.toggle",
  "key": "ctrl+alt+t",
  "macLinux": "cmd+alt+t"
}
```

### win

A keybinding specific to Windows:

```json
{
  "command": "myExtension.toggle",
  "win": "ctrl+alt+t",
  "mac": "cmd+alt+t",
  "linux": "ctrl+alt+t"
}
```

### linux

A keybinding specific to Linux:

```json
{
  "command": "myExtension.toggle",
  "linux": "ctrl+alt+t"
}
```

### when

A condition that determines when the keybinding is active. Keybindings are only triggered when the condition evaluates to true:

```json
{
  "when": "editorTextFocus && editorLangId == javascript"
}
```

## Key Chords

Key chords are sequences of two key combinations. They allow more keybindings without conflicting with existing shortcuts:

```json
{
  "command": "myExtension.action1",
  "key": "ctrl+k ctrl+1"
}
```

To trigger this keybinding, the user presses `Ctrl+K`, releases, then presses `Ctrl+1`.

Chords are useful for creating namespaces of related commands:

```json
{
  "keybindings": [
    { "command": "myExtension.format", "key": "ctrl+k ctrl+f" },
    { "command": "myExtension.sort", "key": "ctrl+k ctrl+s" },
    { "command": "myExtension.fix", "key": "ctrl+k ctrl+x" }
  ]
}
```

## When Clause Contexts

The `when` clause controls when keybindings are active. This prevents keybindings from interfering with each other and ensures they only work in appropriate contexts.

### Editor Contexts

| Context                       | Description                                |
| ----------------------------- | ------------------------------------------ |
| `editorTextFocus`             | A text editor has focus                    |
| `editorFocus`                 | Any editor has focus (text, diff, etc.)    |
| `editorHasSelection`          | Text is selected in the editor             |
| `editorReadonly`              | The editor is read-only                    |
| `editorLangId`                | The language ID of the current file        |
| `editorHasMultipleSelections` | Multiple cursors are active                |
| `isInDiffEditor`              | The active editor is a diff view           |
| `isInEmbeddedEditor`          | The editor is embedded (like in peek view) |

### Resource Contexts

| Context            | Description                           |
| ------------------ | ------------------------------------- |
| `resourceFilename` | The file name of the current resource |
| `resourceExtname`  | The file extension                    |
| `resourceDirname`  | The directory path                    |
| `resourceLangId`   | The language ID of the resource       |
| `resourceScheme`   | The URI scheme (file, untitled, etc.) |

### Workspace Contexts

| Context                | Description                          |
| ---------------------- | ------------------------------------ |
| `workspaceFolderCount` | Number of workspace folders          |
| `workspaceIsTrusted`   | Whether the workspace is trusted     |
| `workspaceSchemes`     | List of workspace folder URI schemes |

### View Contexts

| Context        | Description                                |
| -------------- | ------------------------------------------ |
| `sideBarFocus` | The sidebar has focus                      |
| `panelFocus`   | The panel has focus                        |
| `editorIsOpen` | At least one editor is open                |
| `view`         | The ID of the currently focused view       |
| `viewItem`     | The context value of the focused tree item |

### Debug Contexts

| Context                  | Description                                                           |
| ------------------------ | --------------------------------------------------------------------- |
| `inDebugMode`            | A debug session is active                                             |
| `debugState`             | Current debug state: "inactive", "initializing", "running", "stopped" |
| `debugType`              | The type of the active debug session                                  |
| `debugConfigurationType` | The type of the selected debug configuration                          |

### Terminal Contexts

| Context             | Description                                          |
| ------------------- | ---------------------------------------------------- |
| `terminalFocus`     | A terminal has focus                                 |
| `terminalIsOpen`    | A terminal is open                                   |
| `terminalShellType` | The type of shell: "bash", "zsh", "powershell", etc. |

### Operating System Contexts

| Context     | Description        |
| ----------- | ------------------ |
| `isLinux`   | Running on Linux   |
| `isMac`     | Running on macOS   |
| `isWindows` | Running on Windows |

### Custom Contexts

Extensions can set custom contexts:

```typescript
vscode.commands.executeCommand("setContext", "myExtension.isReady", true);
```

Use in keybindings:

```json
{
  "command": "myExtension.run",
  "key": "ctrl+r",
  "when": "myExtension.isReady"
}
```

## When Clause Operators

### Equality

```json
{ "when": "editorLangId == javascript" }
{ "when": "editorLangId != javascript" }
```

### Comparison

```json
{ "when": "workspaceFolderCount >= 1" }
```

### Boolean

```json
{ "when": "editorHasSelection" }
{ "when": "!editorHasSelection" }
```

### Logical Operators

```json
{ "when": "editorTextFocus && editorLangId == javascript" }
{ "when": "editorTextFocus || editorFocus" }
```

### Pattern Matching

```json
{ "when": "resourceFilename =~ /\\.test\\.js$/" }
```

### Contains

```json
{ "when": "editorLangId in javascriptTypescript" }
```

## Keybinding Precedence

When multiple keybindings match, VSCode uses these rules:

1. **Extension keybindings override built-in keybindings**: If an extension defines a keybinding that conflicts with a built-in one, the extension's keybinding takes precedence.

2. **More specific when clauses win**: A keybinding with a more specific `when` clause takes precedence over one with a less specific clause.

3. **User keybindings override all**: Users can customize keybindings through the Keyboard Shortcuts editor, and these customizations override all defaults.

## Overriding Built-in Keybindings

Extensions can override built-in keybindings by defining the same key combination:

```json
{
  "keybindings": [
    {
      "command": "myExtension.format",
      "key": "shift+alt+f",
      "when": "editorTextFocus"
    }
  ]
}
```

This replaces the built-in "Format Document" command with your extension's format command.

To remove a built-in keybinding without replacing it:

```json
{
  "keybindings": [
    {
      "command": "-workbench.action.files.save",
      "key": "ctrl+s"
    }
  ]
}
```

The `-` prefix removes the keybinding.

## Keybinding Guidelines

### Choose Intuitive Shortcuts

Select key combinations that are:

- Easy to remember
- Easy to press
- Consistent with similar commands

```json
// Good: intuitive and consistent
{ "command": "myExtension.format", "key": "ctrl+shift+f" }

// Avoid: awkward to press
{ "command": "myExtension.format", "key": "ctrl+alt+shift+f" }
```

### Use Appropriate When Clauses

Limit keybindings to relevant contexts:

```json
// Good: only active for JavaScript files
{ "when": "editorLangId == javascript" }

// Avoid: always active, may conflict
{ "when": "true" }
```

### Consider Platform Differences

Account for platform conventions:

```json
{
  "key": "ctrl+shift+f",
  "mac": "cmd+shift+f",
  "linux": "ctrl+shift+f"
}
```

On macOS, `cmd` is typically used instead of `ctrl` for editor commands.

### Avoid Common Conflicts

Be aware of frequently used keybindings:

- `Ctrl+S`: Save
- `Ctrl+C/V/X`: Copy/Paste/Cut
- `Ctrl+Z/Y`: Undo/Redo
- `Ctrl+F`: Find
- `Ctrl+P`: Quick Open

Consider using chords (`Ctrl+K ...`) to avoid conflicts.

## Testing Keybindings

### In Extension Development Host

1. Press `F5` to launch the Extension Development Host
2. Test your keybindings in various contexts
3. Verify that `when` clauses work correctly

### Using the Keyboard Shortcuts Editor

Open the Keyboard Shortcuts editor (`Ctrl+K Ctrl+S`) to see all keybindings, including those from your extension. Search for your command to verify it appears correctly.

### Debugging When Clauses

Use the "Developer: Inspect Context Keys" command to see all active context keys and their values. This helps debug `when` clause issues.

## Best Practices

**Document keybindings**: List your extension's keybindings in the README so users know they exist.

**Allow customization**: Users may want to change keybindings. Ensure your commands work when invoked through any method.

**Use when clauses**: Prevent keybindings from firing in inappropriate contexts.

**Test on all platforms**: Verify keybindings work correctly on Windows, macOS, and Linux.

**Consider accessibility**: Avoid key combinations that are difficult to press or require multiple hands.

## Next Steps

- [Commands](commands.md): Register commands for your keybindings
- [Menus](menus.md): Add commands to menus as an alternative to keybindings
- [Configuration](configuration.md): Allow users to configure keybinding behavior

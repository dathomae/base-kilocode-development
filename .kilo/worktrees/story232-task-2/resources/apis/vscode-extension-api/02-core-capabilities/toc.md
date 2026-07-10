# Core Capabilities

This section covers the fundamental capabilities that most extensions use: commands, menus, keybindings, and configuration.

## Table of Contents

**[commands.md](commands.md)**
: Registering and implementing commands

**[menus.md](menus.md)**
: Adding items to VSCode menus

**[keybindings.md](keybindings.md)**
: Defining keyboard shortcuts for commands

**[configuration.md](configuration.md)**
: Adding settings that users can configure

## Quick Reference

### Registering a Command

```typescript
const disposable = vscode.commands.registerCommand("myExtension.hello", () => {
  vscode.window.showInformationMessage("Hello!");
});
context.subscriptions.push(disposable);
```

### Adding a Menu Item

```json
{
  "contributes": {
    "menus": {
      "editor/context": [
        {
          "command": "myExtension.hello",
          "group": "navigation"
        }
      ]
    }
  }
}
```

### Defining a Keybinding

```json
{
  "contributes": {
    "keybindings": [
      {
        "command": "myExtension.hello",
        "key": "ctrl+f1",
        "mac": "cmd+f1"
      }
    ]
  }
}
```

### Adding Configuration

```json
{
  "contributes": {
    "configuration": {
      "title": "My Extension",
      "properties": {
        "myExtension.enableFeature": {
          "type": "boolean",
          "default": true,
          "description": "Enable the feature"
        }
      }
    }
  }
}
```

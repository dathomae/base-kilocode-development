# Menus

VSCode provides multiple menu locations where extensions can add commands. Menu items give users quick access to extension functionality through right-click context menus, editor title bars, and other UI surfaces.

## Menu Locations

Extensions can contribute to these menu locations:

### Editor Context Menu

The right-click menu in text editors:

```json
{
  "contributes": {
    "menus": {
      "editor/context": [
        {
          "command": "myExtension.format",
          "group": "1_modification"
        }
      ]
    }
  }
}
```

### Editor Title Menu

The menu in the editor title bar (the bar containing the file name and tabs):

```json
{
  "contributes": {
    "menus": {
      "editor/title": [
        {
          "command": "myExtension.preview",
          "group": "navigation"
        }
      ]
    }
  }
}
```

### Editor Title Context Menu

The right-click menu on editor tabs:

```json
{
  "contributes": {
    "menus": {
      "editor/title/context": [
        {
          "command": "myExtension.closeOthers",
          "group": "1_close"
        }
      ]
    }
  }
}
```

### Explorer Context Menu

The right-click menu in the file explorer:

```json
{
  "contributes": {
    "menus": {
      "explorer/context": [
        {
          "command": "myExtension.analyze",
          "group": "2_workspace"
        }
      ]
    }
  }
}
```

### View Title Menu

The menu in the title bar of custom views:

```json
{
  "contributes": {
    "menus": {
      "view/title": [
        {
          "command": "myExtension.refresh",
          "when": "view == myExtension.view",
          "group": "navigation"
        }
      ]
    }
  }
}
```

### View Item Context Menu

The right-click menu for items in custom tree views:

```json
{
  "contributes": {
    "menus": {
      "view/item/context": [
        {
          "command": "myExtension.delete",
          "when": "view == myExtension.view && viewItem == file",
          "group": "2_modification"
        }
      ]
    }
  }
}
```

### SCM Title Menu

The menu in the Source Control view title:

```json
{
  "contributes": {
    "menus": {
      "scm/title": [
        {
          "command": "myExtension.commitAll",
          "group": "navigation"
        }
      ]
    }
  }
}
```

### SCM Resource State Context Menu

The context menu for SCM resource states:

```json
{
  "contributes": {
    "menus": {
      "scm/resourceState/context": [
        {
          "command": "myExtension.stage",
          "group": "1_modification"
        }
      ]
    }
  }
}
```

### Debug Toolbar

The debug toolbar that appears during debugging:

```json
{
  "contributes": {
    "menus": {
      "debug/toolbar": [
        {
          "command": "myExtension.stepBack",
          "group": "navigation"
        }
      ]
    }
  }
}
```

### Terminal Context Menu

The context menu in the integrated terminal:

```json
{
  "contributes": {
    "menus": {
      "terminal/context": [
        {
          "command": "myExtension.runInTerminal",
          "group": "2_run"
        }
      ]
    }
  }
}
```

### Command Palette

Control whether commands appear in the Command Palette:

```json
{
  "contributes": {
    "menus": {
      "commandPalette": [
        {
          "command": "myExtension.internal",
          "when": "false"
        }
      ]
    }
  }
}
```

## Menu Item Structure

Each menu item has these properties:

```json
{
  "command": "myExtension.doSomething",
  "when": "editorLangId == javascript",
  "group": "1_modification",
  "submenu": "myExtension.submenu"
}
```

### command

The command to execute when the menu item is clicked. The command must be declared in `contributes.commands`.

### when

A condition that determines when the menu item appears. Uses the same syntax as keybinding when clauses:

```json
{
  "when": "editorLangId == javascript && editorHasSelection"
}
```

Common contexts include:

- `editorLangId`: Current file language
- `editorHasSelection`: Text is selected
- `resourceFilename`: File name pattern
- `resourceDirname`: Directory path
- `isInDiffEditor`: Editor is a diff view

### group

Groups organize menu items into sections. Items in the same group appear together with separators between groups:

```json
{
  "menus": {
    "editor/context": [
      { "command": "myExtension.copy", "group": "1_copy" },
      { "command": "myExtension.cut", "group": "1_copy" },
      { "command": "myExtension.paste", "group": "2_clipboard" }
    ]
  }
}
```

Groups are sorted alphabetically. Use numeric prefixes to control order: `1_first`, `2_second`, etc.

### Standard Groups

VSCode defines standard groups for consistency:

**editor/context**:

- `1_modification`: Cut, copy, paste
- `navigation`: Go to definition, peek
- `z_commands`: Other commands

**editor/title**:

- `navigation`: Primary actions (icons)
- `1_run`: Run, debug
- `2_test`: Test-related actions

**explorer/context**:

- `navigation`: Open, reveal
- `2_workspace`: Workspace actions
- `3_compare`: Compare actions

### submenu

Instead of a command, a menu item can open a submenu:

```json
{
  "contributes": {
    "submenus": {
      "myExtension.submenu": {
        "label": "My Extension"
      }
    },
    "menus": {
      "editor/context": [
        {
          "submenu": "myExtension.submenu",
          "group": "myExtension"
        }
      ],
      "myExtension.submenu": [
        {
          "command": "myExtension.action1"
        },
        {
          "command": "myExtension.action2"
        }
      ]
    }
  }
}
```

### alt

Define an alternative command that runs when holding Alt while clicking:

```json
{
  "menus": {
    "editor/title": [
      {
        "command": "myExtension.run",
        "alt": "myExtension.debug",
        "group": "navigation"
      }
    ]
  }
}
```

This shows the "run" command normally, but executes "debug" when Alt is held.

## Icons in Menus

Menu items can display icons:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "myExtension.refresh",
        "title": "Refresh",
        "icon": {
          "light": "resources/light/refresh.svg",
          "dark": "resources/dark/refresh.svg"
        }
      }
    ]
  }
}
```

Icons are most visible in the editor title bar and view title menus.

### Using Codicons

VSCode provides built-in icons called Codicons:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "myExtension.refresh",
        "title": "Refresh",
        "icon": "$(refresh)"
      }
    ]
  }
}
```

Common Codicons include: `refresh`, `add`, `remove`, `edit`, `save`, `search`, `settings-gear`, `play`, `debug`, `check`, `close`, `more`.

## Dynamic Menu Items

Menu items can use `when` clauses to appear conditionally:

```json
{
  "menus": {
    "editor/context": [
      {
        "command": "myExtension.formatJson",
        "when": "editorLangId == json"
      },
      {
        "command": "myExtension.formatPython",
        "when": "editorLangId == python"
      }
    ]
  }
}
```

### Custom Contexts

Extensions can set custom contexts for dynamic menus:

```typescript
// Set context when a file is selected
vscode.commands.executeCommand("setContext", "myExtension.hasFile", true);

// Use in menu when clause
// "when": "myExtension.hasFile"
```

## Menu Item Arguments

Menu items can pass arguments to commands using the `args` property:

```json
{
  "menus": {
    "editor/context": [
      {
        "command": "myExtension.insertText",
        "args": ["Hello, World!"]
      }
    ]
  }
}
```

The command receives these arguments:

```typescript
vscode.commands.registerCommand("myExtension.insertText", (text: string) => {
  const editor = vscode.window.activeTextEditor;
  if (editor) {
    editor.edit((editBuilder) => {
      editBuilder.insert(editor.selection.active, text);
    });
  }
});
```

## Resource Context in Menus

For menus that operate on resources (files, folders), access the selected resource:

```typescript
vscode.commands.registerCommand("myExtension.analyze", (uri: vscode.Uri) => {
  // uri is the selected file or folder
  console.log("Analyzing:", uri.fsPath);
});
```

When multiple items are selected, the command receives an array:

```typescript
vscode.commands.registerCommand("myExtension.analyze", (uris: vscode.Uri[]) => {
  uris.forEach((uri) => console.log("Analyzing:", uri.fsPath));
});
```

## Best Practices

**Use appropriate groups**: Place menu items in standard groups for consistency with VSCode's built-in menus.

**Limit menu items**: Too many menu items clutter the interface. Only add items for frequently used commands.

**Use when clauses**: Show menu items only when they are relevant to the current context.

**Provide icons**: Icons make menu items more recognizable, especially in the editor title bar.

**Consider submenus**: Group related commands under a submenu to reduce clutter.

**Test in different contexts**: Verify that menu items appear and function correctly in all relevant contexts.

## Common Patterns

### Format Command in Editor Context

```json
{
  "contributes": {
    "commands": [
      {
        "command": "myExtension.format",
        "title": "Format with My Extension"
      }
    ],
    "menus": {
      "editor/context": [
        {
          "command": "myExtension.format",
          "when": "editorLangId == javascript",
          "group": "1_modification"
        }
      ]
    }
  }
}
```

### File Operations in Explorer

```json
{
  "contributes": {
    "menus": {
      "explorer/context": [
        {
          "command": "myExtension.newFile",
          "group": "2_workspace"
        },
        {
          "command": "myExtension.deleteFile",
          "when": "!explorerResourceIsFolder",
          "group": "2_workspace"
        }
      ]
    }
  }
}
```

### View Actions

```json
{
  "contributes": {
    "views": {
      "explorer": [{ "id": "myExtension.view", "name": "My View" }]
    },
    "menus": {
      "view/title": [
        {
          "command": "myExtension.refresh",
          "when": "view == myExtension.view",
          "group": "navigation"
        }
      ],
      "view/item/context": [
        {
          "command": "myExtension.openItem",
          "when": "view == myExtension.view",
          "group": "navigation"
        }
      ]
    }
  }
}
```

## Next Steps

- [Commands](commands.md): Learn about registering and implementing commands
- [Keybindings](keybindings.md): Define keyboard shortcuts for commands
- [Tree Views](../04-ui-components/tree-views.md): Create custom views with context menus

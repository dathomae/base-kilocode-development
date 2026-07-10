# Extension Manifest

The extension manifest is a `package.json` file that declares an extension's metadata, dependencies, and contributions. VSCode reads this file to understand what the extension provides and when to activate it.

## Required Fields

Every extension manifest must include these fields:

```json
{
  "name": "my-extension",
  "displayName": "My Extension",
  "description": "A brief description of what the extension does",
  "version": "1.0.0",
  "engines": {
    "vscode": "^1.74.0"
  },
  "main": "./out/extension.js"
}
```

### name

The extension's identifier used in the marketplace and for programmatic references. Use lowercase letters, numbers, and hyphens. Avoid spaces and special characters. The name must be unique within your publisher namespace.

### displayName

The human-readable name that appears in the VSCode Extensions view and marketplace. This can include spaces and special characters.

### description

A brief summary of the extension's functionality. Keep it concise but informative, as this appears in search results and extension listings.

### version

The extension's version number following semantic versioning (MAJOR.MINOR.PATCH). Increment this when releasing updates.

### engines.vscode

Specifies the minimum VSCode version the extension supports. The extension will not activate in older versions. Use the `^` prefix to indicate compatibility with newer minor versions.

### main

The path to the compiled JavaScript entry point. For TypeScript extensions, this typically points to `./out/extension.js` or `./dist/extension.js` if using webpack.

## Publisher Information

Extensions intended for the marketplace require publisher information:

```json
{
  "publisher": "my-publisher-id",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/username/my-extension.git"
  },
  "bugs": {
    "url": "https://github.com/username/my-extension/issues"
  },
  "homepage": "https://github.com/username/my-extension#readme"
}
```

### publisher

Your publisher ID, which you create when setting up a marketplace account. All extensions you publish will be associated with this ID.

### license

The SPDX license identifier (e.g., "MIT", "Apache-2.0", "GPL-3.0"). This clarifies how others can use and distribute your extension.

### repository

Source code repository information. Including this helps users report issues and contribute to the extension.

## Categories and Keywords

Categories and keywords help users discover your extension:

```json
{
  "categories": ["Programming Languages", "Linters", "Formatters"],
  "keywords": ["python", "linting", "formatting"]
}
```

### categories

An array of categories from a predefined list. Valid categories include:

- Programming Languages
- Snippets
- Linters
- Themes
- Debuggers
- Formatters
- Keymaps
- SCM Providers
- Other

### keywords

An array of search keywords that help users find your extension. Choose terms users might search for.

## Extension Icon

A custom icon makes your extension recognizable in the marketplace:

```json
{
  "icon": "images/icon.png"
}
```

The icon should be a PNG, JPG, or SVG file at least 128x128 pixels. Place it in your extension directory and reference the relative path.

## Activation Events

Activation events determine when VSCode loads your extension. Extensions activate lazily to keep VSCode fast:

```json
{
  "activationEvents": [
    "onCommand:myExtension.sayHello",
    "onLanguage:python",
    "workspaceContains:**/.eslintrc.json"
  ]
}
```

See [activation-events.md](activation-events.md) for a complete reference.

## Contribution Points

The `contributes` section declares what the extension adds to VSCode:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "myExtension.sayHello",
        "title": "Say Hello",
        "category": "My Extension"
      }
    ],
    "menus": {
      "editor/context": [
        {
          "command": "myExtension.sayHello",
          "group": "navigation"
        }
      ]
    },
    "keybindings": [
      {
        "command": "myExtension.sayHello",
        "key": "ctrl+f1",
        "mac": "cmd+f1"
      }
    ],
    "configuration": {
      "title": "My Extension",
      "properties": {
        "myExtension.enableFeature": {
          "type": "boolean",
          "default": true,
          "description": "Enable the special feature"
        }
      }
    }
  }
}
```

### commands

Registers commands that users can invoke through the Command Palette or keybindings. Each command has:

- `command`: A unique identifier (typically `extensionId.commandName`)
- `title`: The display name shown in the Command Palette
- `category`: Optional grouping for related commands

### menus

Adds commands to various menus including:

- `editor/context`: Editor right-click context menu
- `editor/title`: Editor title bar
- `editor/title/context`: Editor title context menu
- `explorer/context`: Explorer context menu
- `view/title`: View title menu
- `view/item/context`: View item context menu

### keybindings

Defines keyboard shortcuts for commands. Specify separate key combinations for different platforms:

- `key`: Default key combination (Windows/Linux)
- `mac`: macOS-specific key combination
- `when`: Optional condition for when the keybinding is active

### configuration

Declares settings that users can configure. Each property has:

- `type`: The value type (string, boolean, number, array, object)
- `default`: The default value
- `description`: Help text shown in the settings UI
- `enum`: Optional list of allowed values

### Other Contribution Points

Additional contribution points include:

- `languages`: Register new languages with file associations
- `grammars`: Provide TextMate grammars for syntax highlighting
- `snippets`: Define code snippets
- `themes`: Contribute color themes
- `iconThemes`: Contribute file icon themes
- `views`: Register custom views
- `viewsContainers`: Register view containers (side bar icons)
- `problemMatchers`: Define problem matchers for tasks
- `taskDefinitions`: Define custom task types
- `debuggers`: Contribute debugger extensions

## Dependencies

Extensions can declare runtime and development dependencies:

```json
{
  "dependencies": {
    "axios": "^1.0.0"
  },
  "devDependencies": {
    "@types/node": "^18.0.0",
    "@types/vscode": "^1.74.0",
    "typescript": "^4.9.0"
  }
}
```

### dependencies

Runtime dependencies bundled with the extension. These packages are included when users install the extension. Keep dependencies minimal to reduce extension size.

### devDependencies

Development-time dependencies not bundled with the extension. These include type definitions, build tools, and testing frameworks.

## Extension Kind

The `extensionKind` property specifies where the extension can run:

```json
{
  "extensionKind": ["workspace", "ui"]
}
```

- `workspace`: Extension runs on the remote machine (for remote development)
- `ui`: Extension runs on the local machine

The array order indicates preference. `["workspace", "ui"]` means the extension prefers to run on the remote machine but can fall back to local.

## Extension Pack

Extension packs bundle multiple extensions together:

```json
{
  "extensionPack": ["publisher.extension1", "publisher.extension2"]
}
```

Extension packs are useful for distributing a set of related extensions as a single installable package.

## Scripts

The `scripts` section defines npm scripts for building and testing:

```json
{
  "scripts": {
    "vscode:prepublish": "npm run compile",
    "compile": "tsc -p ./",
    "watch": "tsc -watch -p ./",
    "test": "node ./out/test/runTest.js",
    "lint": "eslint src --ext ts"
  }
}
```

### vscode:prepublish

This special script runs before publishing the extension. Use it to compile TypeScript, run webpack, or perform other build steps.

## Complete Example

Here is a complete manifest example:

```json
{
  "name": "python-tools",
  "displayName": "Python Tools",
  "description": "Essential tools for Python development",
  "version": "1.0.0",
  "publisher": "my-publisher",
  "license": "MIT",
  "icon": "images/icon.png",
  "engines": {
    "vscode": "^1.74.0"
  },
  "categories": ["Programming Languages", "Linters", "Formatters"],
  "keywords": ["python", "linting", "formatting"],
  "activationEvents": ["onLanguage:python", "onCommand:pythonTools.format"],
  "main": "./out/extension.js",
  "contributes": {
    "commands": [
      {
        "command": "pythonTools.format",
        "title": "Format Python File",
        "category": "Python Tools"
      }
    ],
    "keybindings": [
      {
        "command": "pythonTools.format",
        "key": "ctrl+shift+f",
        "mac": "cmd+shift+f",
        "when": "editorLangId == python"
      }
    ],
    "configuration": {
      "title": "Python Tools",
      "properties": {
        "pythonTools.formatOnSave": {
          "type": "boolean",
          "default": false,
          "description": "Automatically format Python files on save"
        },
        "pythonTools.linter": {
          "type": "string",
          "default": "pylint",
          "enum": ["pylint", "flake8", "mypy"],
          "description": "The linter to use for Python files"
        }
      }
    }
  },
  "scripts": {
    "vscode:prepublish": "npm run compile",
    "compile": "tsc -p ./",
    "watch": "tsc -watch -p ./",
    "test": "node ./out/test/runTest.js"
  },
  "devDependencies": {
    "@types/node": "^18.0.0",
    "@types/vscode": "^1.74.0",
    "typescript": "^4.9.0"
  }
}
```

## Best Practices

**Keep the manifest clean**: Remove unused fields and dependencies. A minimal manifest loads faster and is easier to maintain.

**Use semantic versioning**: Follow MAJOR.MINOR.PATCH conventions. Breaking changes require a major version bump.

**Specify realistic engine requirements**: Set the minimum VSCode version to the oldest version you test against. Avoid requiring the absolute latest unless necessary.

**Document configuration**: Provide clear descriptions for all settings. Include examples for complex settings.

**Organize contributions**: Group related commands, menus, and settings logically. Use categories to organize commands in the Command Palette.

## Next Steps

- [Activation Events](activation-events.md): Learn about all available activation events
- [Extension Lifecycle](extension-lifecycle.md): Understand how extensions activate and deactivate
- [Commands and Menus](../02-core-capabilities/): Add commands and menu items to your extension

# Configuration

Extensions can contribute settings that users configure through VSCode's Settings UI. Configuration allows users to customize extension behavior without modifying code.

## Declaring Configuration

Configuration is declared in the `package.json` manifest under `contributes.configuration`:

```json
{
  "contributes": {
    "configuration": {
      "title": "My Extension",
      "properties": {
        "myExtension.enableFeature": {
          "type": "boolean",
          "default": true,
          "description": "Enable the special feature"
        },
        "myExtension.maxItems": {
          "type": "number",
          "default": 100,
          "description": "Maximum number of items to display"
        }
      }
    }
  }
}
```

## Configuration Properties

### title

The title appears as the section heading in the Settings UI:

```json
{
  "title": "Python Tools"
}
```

### properties

An object containing the settings your extension provides. Each property key is the setting ID, and the value describes the setting.

### Setting ID

Setting IDs should include your extension name as a prefix to avoid conflicts:

```json
{
  "myExtension.enableFeature": { ... },
  "myExtension.maxItems": { ... }
}
```

### Setting Properties

Each setting has these properties:

**type**: The data type. Valid types are:

- `boolean`: True/false values
- `string`: Text values
- `number`: Numeric values
- `integer`: Integer values
- `array`: Lists of values
- `object`: Nested objects

**default**: The default value when the user has not configured the setting.

**description**: Help text shown in the Settings UI. Write clear, concise descriptions that explain what the setting does.

**enum**: An array of allowed values. Creates a dropdown in the Settings UI:

```json
{
  "myExtension.linter": {
    "type": "string",
    "default": "pylint",
    "enum": ["pylint", "flake8", "mypy"],
    "description": "The linter to use for Python files"
  }
}
```

**enumDescriptions**: Descriptions for each enum value:

```json
{
  "myExtension.linter": {
    "type": "string",
    "default": "pylint",
    "enum": ["pylint", "flake8", "mypy"],
    "enumDescriptions": [
      "Pylint - comprehensive analysis",
      "Flake8 - fast style checking",
      "MyPy - static type checking"
    ]
  }
}
```

**minimum / maximum**: For numeric types, constrain the range:

```json
{
  "myExtension.port": {
    "type": "integer",
    "default": 8080,
    "minimum": 1024,
    "maximum": 65535,
    "description": "The port number for the server"
  }
}
```

**minLength / maxLength**: For string types, constrain the length:

```json
{
  "myExtension.prefix": {
    "type": "string",
    "default": "",
    "maxLength": 10,
    "description": "Prefix for generated names"
  }
}
```

**pattern**: A regular expression that string values must match:

```json
{
  "myExtension.apiKey": {
    "type": "string",
    "default": "",
    "pattern": "^[A-Za-z0-9]{32}$",
    "description": "API key (32 alphanumeric characters)"
  }
}
```

**items**: For array types, describe the item type:

```json
{
  "myExtension.excludePatterns": {
    "type": "array",
    "items": {
      "type": "string"
    },
    "default": ["node_modules", ".git"],
    "description": "Patterns to exclude from analysis"
  }
}
```

**properties**: For object types, describe nested properties:

```json
{
  "myExtension.server": {
    "type": "object",
    "default": {
      "host": "localhost",
      "port": 8080
    },
    "properties": {
      "host": {
        "type": "string",
        "description": "Server hostname"
      },
      "port": {
        "type": "integer",
        "description": "Server port"
      }
    }
  }
}
```

**scope**: Where the setting applies:

- `application`: Global across all workspaces (default)
- `window`: Per-window (workspace)
- `resource`: Per-file or folder
- `language-overridable`: Per-language (resource scope with language overrides)

```json
{
  "myExtension.formatOnSave": {
    "type": "boolean",
    "default": false,
    "scope": "resource",
    "description": "Format files when saving"
  }
}
```

**deprecationMessage**: Mark a setting as deprecated:

```json
{
  "myExtension.oldSetting": {
    "type": "boolean",
    "default": true,
    "deprecationMessage": "Use myExtension.newSetting instead"
  }
}
```

**markdownDescription**: Use Markdown in the description:

```json
{
  "myExtension.path": {
    "type": "string",
    "default": "",
    "markdownDescription": "Path to the executable. Use `${workspaceFolder}` for the current workspace."
  }
}
```

**order**: Control the order of settings in the UI (lower numbers appear first):

```json
{
  "myExtension.enable": {
    "type": "boolean",
    "default": true,
    "order": 1
  },
  "myExtension.advanced": {
    "type": "boolean",
    "default": false,
    "order": 2
  }
}
```

## Reading Configuration

Access configuration values using the `vscode.workspace.getConfiguration` API:

```typescript
const config = vscode.workspace.getConfiguration("myExtension");

const enabled = config.get<boolean>("enableFeature", true);
const maxItems = config.get<number>("maxItems", 100);
const linter = config.get<string>("linter", "pylint");
```

### The get Method

The `get` method retrieves a configuration value:

```typescript
config.get<T>(section: string, defaultValue?: T): T | undefined
```

Always provide a default value to handle cases where the setting is not configured:

```typescript
const port = config.get<number>("port", 8080);
```

### Scoped Configuration

For resource-scoped settings, pass a URI to get configuration for a specific file:

```typescript
const document = vscode.window.activeTextEditor?.document;
if (document) {
  const config = vscode.workspace.getConfiguration("myExtension", document.uri);
  const formatOnSave = config.get<boolean>("formatOnSave", false);
}
```

### Inspecting Configuration

Inspect a setting to see its value at different scopes:

```typescript
const inspect = vscode.workspace
  .getConfiguration("myExtension")
  .inspect<boolean>("enableFeature");

console.log("Default:", inspect?.defaultValue);
console.log("Global:", inspect?.globalValue);
console.log("Workspace:", inspect?.workspaceValue);
console.log("Workspace Folder:", inspect?.workspaceFolderValue);
```

## Updating Configuration

Programmatically update configuration values:

```typescript
const config = vscode.workspace.getConfiguration("myExtension");

// Update a setting
await config.update("enableFeature", false);

// Update at a specific scope
await config.update("enableFeature", false, vscode.ConfigurationTarget.Global);
await config.update(
  "enableFeature",
  false,
  vscode.ConfigurationTarget.Workspace,
);
```

### Configuration Targets

- `vscode.ConfigurationTarget.Global`: User settings (applies everywhere)
- `vscode.ConfigurationTarget.Workspace`: Workspace settings (applies to current workspace)
- `vscode.ConfigurationTarget.WorkspaceFolder`: Folder settings (applies to specific folder)

### Updating Resource-Scoped Settings

For resource-scoped settings, pass a URI:

```typescript
const config = vscode.workspace.getConfiguration("myExtension", document.uri);
await config.update("formatOnSave", true);
```

## Listening for Configuration Changes

React to configuration changes using the `onDidChangeConfiguration` event:

```typescript
context.subscriptions.push(
  vscode.workspace.onDidChangeConfiguration((e) => {
    if (e.affectsConfiguration("myExtension.enableFeature")) {
      const newValue = vscode.workspace
        .getConfiguration("myExtension")
        .get<boolean>("enableFeature");
      console.log("enableFeature changed to:", newValue);
    }
  }),
);
```

### Checking Specific Scopes

Check if a change affects a specific resource:

```typescript
vscode.workspace.onDidChangeConfiguration((e) => {
  const document = vscode.window.activeTextEditor?.document;
  if (
    document &&
    e.affectsConfiguration("myExtension.formatOnSave", document.uri)
  ) {
    // Format on save setting changed for this document
  }
});
```

## Configuration Examples

### Boolean Setting

```json
{
  "myExtension.enableLinting": {
    "type": "boolean",
    "default": true,
    "description": "Enable linting for Python files"
  }
}
```

```typescript
const config = vscode.workspace.getConfiguration("myExtension");
if (config.get<boolean>("enableLinting", true)) {
  runLinter();
}
```

### String with Enum

```json
{
  "myExtension.logLevel": {
    "type": "string",
    "default": "info",
    "enum": ["debug", "info", "warning", "error"],
    "enumDescriptions": [
      "Detailed debugging information",
      "General information",
      "Warning messages only",
      "Error messages only"
    ],
    "description": "The minimum log level to display"
  }
}
```

### Array Setting

```json
{
  "myExtension.ignorePatterns": {
    "type": "array",
    "items": {
      "type": "string"
    },
    "default": ["**/node_modules/**", "**/dist/**"],
    "description": "Glob patterns for files to ignore"
  }
}
```

```typescript
const patterns = vscode.workspace
  .getConfiguration("myExtension")
  .get<string[]>("ignorePatterns", []);
```

### Object Setting

```json
{
  "myExtension.shortcuts": {
    "type": "object",
    "default": {
      "run": "ctrl+r",
      "test": "ctrl+t"
    },
    "description": "Custom keyboard shortcuts"
  }
}
```

```typescript
const shortcuts = vscode.workspace
  .getConfiguration("myExtension")
  .get<Record<string, string>>("shortcuts", {});
```

### Language-Specific Settings

Users can override settings per language through VSCode's UI. Extensions can read language-specific settings:

```typescript
const document = vscode.window.activeTextEditor?.document;
if (document) {
  const config = vscode.workspace.getConfiguration("myExtension", document);
  const setting = config.get<string>("someSetting");
}
```

## Multiple Configuration Sections

For extensions with many settings, split configuration into multiple sections:

```json
{
  "contributes": {
    "configuration": [
      {
        "title": "Python Tools - General",
        "properties": {
          "myExtension.enableLinting": { ... },
          "myExtension.enableFormatting": { ... }
        }
      },
      {
        "title": "Python Tools - Linting",
        "properties": {
          "myExtension.linter": { ... },
          "myExtension.lintOnSave": { ... }
        }
      }
    ]
  }
}
```

## Best Practices

**Use descriptive names**: Setting names should clearly indicate what they control. Use camelCase for consistency.

**Provide sensible defaults**: Choose defaults that work well for most users. Avoid requiring configuration for basic functionality.

**Write clear descriptions**: Explain what each setting does and, when helpful, how different values affect behavior.

**Use appropriate scopes**: Mark settings as `resource` scope when they should be configurable per-folder or per-file.

**Handle missing values**: Always provide default values when reading configuration to handle cases where settings are not configured.

**React to changes**: Listen for configuration changes and update extension behavior accordingly, rather than requiring a restart.

**Document settings**: Include a section in your README that describes available settings and their effects.

## Next Steps

- [Commands](commands.md): Create commands that use configuration values
- [Extension Manifest](../01-extension-anatomy/extension-manifest.md): Understand all manifest options

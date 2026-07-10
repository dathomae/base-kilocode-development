# Getting Started with VSCode Extension Development

This guide walks through setting up a development environment and creating your first VSCode extension. By the end, you will have a working extension that registers a command and displays a message.

## Prerequisites

Before developing VSCode extensions, ensure you have the following installed:

**Node.js**: Version 18 or later. Node.js provides the runtime for extension development tools and the Extension Host process. Download it from [nodejs.org](https://nodejs.org/).

**Visual Studio Code**: A recent version of VSCode. You will develop extensions using VSCode itself, taking advantage of its built-in debugging and development features.

**Git**: Version control for managing extension source code. Many extension templates include Git initialization.

**A code editor**: VSCode itself serves this purpose, providing TypeScript/JavaScript IntelliSense, debugging, and integrated terminal access.

## Installing the Extension Generator

Microsoft provides an official extension generator that creates a starter project with all necessary configuration. Install it globally using npm:

```bash
npm install -g yo generator-code
```

The `yo` tool runs Yeoman generators, and `generator-code` is the official VSCode extension template generator.

## Creating Your First Extension

Run the generator to create a new extension project:

```bash
yo code
```

The generator prompts for several configuration options:

### Extension Name

Enter a name for your extension. This name appears in the VSCode marketplace and serves as the extension's display name. Use a descriptive name that indicates the extension's purpose, such as "markdown-linter" or "python-test-runner".

### Extension Identifier

The identifier is a unique string that distinguishes your extension. By default, it combines your publisher ID with the extension name, formatted as `publisher.extension-name`. Use lowercase letters, numbers, and hyphens.

### Description

Provide a brief description of what the extension does. This description appears in the marketplace and helps users understand the extension's purpose.

### Initialize a Git Repository

Choose whether to initialize a Git repository. Version control is recommended for tracking changes and collaborating with others.

### Bundle with Webpack

The generator can configure webpack to bundle your extension. Bundling reduces the extension size and improves load times. Select "Yes" for production extensions.

### Package Manager

Choose between npm and yarn for managing dependencies. Both work well; npm comes bundled with Node.js.

### Extension Type

Select the type of extension to create:

- **New Extension (TypeScript)**: A general-purpose extension using TypeScript. This is the recommended choice for most extensions.
- **New Extension (JavaScript)**: A general-purpose extension using JavaScript. Choose this if you prefer not to use TypeScript.
- **New Color Theme**: Creates a color theme extension.
- **New Language Support**: Creates a basic language support extension with syntax highlighting.

For this guide, select "New Extension (TypeScript)".

### TypeScript Configuration

If you chose TypeScript, the generator creates a `tsconfig.json` with appropriate settings for extension development.

## Project Structure

The generator creates a project with the following structure:

```
my-extension/
├── .vscode/
│   ├── launch.json     # Debug configuration
│   ├── tasks.json      # Build tasks
│   └── extensions.json # Recommended extensions
├── src/
│   └── extension.ts    # Extension entry point
├── package.json        # Extension manifest
├── tsconfig.json       # TypeScript configuration
└── README.md           # Extension documentation
```

### The Extension Manifest

The `package.json` file serves as the extension manifest. It contains metadata and declares the extension's capabilities:

```json
{
  "name": "my-extension",
  "displayName": "My Extension",
  "description": "A sample VSCode extension",
  "version": "0.0.1",
  "engines": {
    "vscode": "^1.74.0"
  },
  "categories": ["Other"],
  "activationEvents": [],
  "main": "./out/extension.js",
  "contributes": {
    "commands": [
      {
        "command": "my-extension.helloWorld",
        "title": "Hello World"
      }
    ]
  },
  "scripts": {
    "vscode:prepublish": "npm run compile",
    "compile": "tsc -p ./",
    "watch": "tsc -watch -p ./"
  },
  "devDependencies": {
    "@types/node": "^18.0.0",
    "@types/vscode": "^1.74.0",
    "typescript": "^4.9.0"
  }
}
```

Key fields include:

- **engines.vscode**: Specifies the minimum VSCode version the extension supports
- **activationEvents**: Defines when the extension activates
- **main**: Points to the compiled JavaScript entry point
- **contributes**: Declares the extension's contributions (commands, menus, etc.)

### The Extension Entry Point

The `src/extension.ts` file contains the extension's activation and deactivation logic:

```typescript
import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  console.log('Extension "my-extension" is now active!');

  const disposable = vscode.commands.registerCommand(
    "my-extension.helloWorld",
    () => {
      vscode.window.showInformationMessage("Hello World from My Extension!");
    },
  );

  context.subscriptions.push(disposable);
}

export function deactivate() {}
```

The `activate` function runs when the extension activates. It registers a command that displays a message when invoked. The `deactivate` function runs when the extension deactivates, allowing cleanup of resources.

## Running and Debugging the Extension

VSCode provides first-class support for debugging extensions. The generated project includes a debug configuration in `.vscode/launch.json`.

### Starting the Debug Session

1. Open the extension project in VSCode
2. Press `F5` or select "Run > Start Debugging" from the menu
3. A new VSCode window opens with the extension loaded

This new window is called the "Extension Development Host." It runs with your extension loaded and provides a clean environment for testing.

### Testing the Command

1. In the Extension Development Host window, open the Command Palette (`Ctrl+Shift+P` or `Cmd+Shift+P` on macOS)
2. Type "Hello World" to find your command
3. Press Enter to execute the command
4. An information message appears in the bottom-right corner

### Setting Breakpoints

You can set breakpoints in your TypeScript source code:

1. In the original VSCode window, open `src/extension.ts`
2. Click in the left margin next to a line of code to set a breakpoint
3. Trigger the code path in the Extension Development Host
4. Execution pauses at the breakpoint, allowing inspection of variables

### The Debug Console

The Debug Console shows output from your extension, including `console.log` statements. Use it to view diagnostic information during development.

## Making Changes

When you modify the source code, the changes do not automatically appear in the Extension Development Host. You have two options:

### Recompile and Restart

1. Stop the debug session (`Shift+F5`)
2. Run the compile task: `npm run compile` or use the build task (`Ctrl+Shift+B`)
3. Start debugging again (`F5`)

### Watch Mode

Run the TypeScript compiler in watch mode to automatically recompile when files change:

```bash
npm run watch
```

With watch mode running, restart the debug session to load the latest changes.

## Understanding Activation Events

The generated extension uses `onCommand` activation, meaning it activates when the user executes the command. The manifest declares this:

```json
{
  "activationEvents": ["onCommand:my-extension.helloWorld"]
}
```

Other common activation events include:

- `onLanguage:<languageId>`: Activate when opening a file of the specified language
- `onView:<viewId>`: Activate when a specific view becomes visible
- `onFileSystem:<scheme>`: Activate when accessing a file with the specified scheme
- `workspaceContains:<glob>`: Activate when the workspace contains files matching the pattern
- `*`: Activate on VSCode startup (use sparingly)

## Adding a Keybinding

To make the command accessible via keyboard shortcut, add a keybinding to the manifest:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "my-extension.helloWorld",
        "title": "Hello World"
      }
    ],
    "keybindings": [
      {
        "command": "my-extension.helloWorld",
        "key": "ctrl+f1",
        "mac": "cmd+f1",
        "when": "editorTextFocus"
      }
    ]
  }
}
```

The `when` clause specifies conditions under which the keybinding is active. `editorTextFocus` means the keybinding works only when a text editor has focus.

## Next Steps

Now that you have a working extension, explore these topics to add more functionality:

- [Extension Anatomy](01-extension-anatomy/): Understand the manifest, activation events, and lifecycle
- [Commands, Menus, and Keybindings](02-core-capabilities/): Add more ways for users to interact with your extension
- [Language Features](03-language-features/): Add code completion, hover information, and diagnostics
- [UI Components](04-ui-components/): Create custom views and webviews

## Troubleshooting

### Extension Does Not Appear

If the extension does not load in the Extension Development Host:

1. Check the Debug Console for errors
2. Verify the `main` field in `package.json` points to the correct output file
3. Ensure TypeScript compiled successfully

### Command Not Found

If the command does not appear in the Command Palette:

1. Verify the command is registered in `package.json` under `contributes.commands`
2. Check that the command ID in `registerCommand` matches the manifest
3. Ensure the extension activated (check the Debug Console for the activation message)

### Changes Not Reflected

If changes do not appear after modifying code:

1. Ensure you restarted the debug session after compilation
2. Check that watch mode is running or manually compile
3. Verify you edited the correct file

## External Resources

- [VSCode Extension API Documentation](https://code.visualstudio.com/api)
- [Your First Extension Guide](https://code.visualstudio.com/api/get-started/your-first-extension)
- [Extension Guidelines](https://code.visualstudio.com/api/references/extension-guidelines)
- [Extension Samples Repository](https://github.com/microsoft/vscode-extension-samples)

# Terminal Integration

VSCode provides a terminal API that allows extensions to create terminals, send text to them, and create terminal profiles for custom development environments.

## Creating Terminals

Create a terminal using `vscode.window.createTerminal`:

```typescript
import * as vscode from "vscode";

// Create a basic terminal
const terminal = vscode.window.createTerminal("My Terminal");
terminal.show();

// Create a terminal with options
const terminal = vscode.window.createTerminal({
  name: "Build Terminal",
  shellPath: "/bin/bash",
  shellArgs: ["--login"],
  cwd: "/path/to/project",
  env: { NODE_ENV: "development" },
});
```

### Terminal Options

```typescript
interface TerminalOptions {
  // Terminal name shown in the dropdown
  name?: string;

  // Shell executable path
  shellPath?: string;

  // Arguments for the shell
  shellArgs?: string | string[];

  // Working directory
  cwd?: string | vscode.Uri;

  // Environment variables
  env?: { [key: string]: string | null };

  // Use environment variables from the extension host
  strictEnv?: boolean;

  // Terminal location: editor or panel
  location?: vscode.TerminalLocation;

  // Icon for the terminal
  iconPath?:
    | vscode.Uri
    | { light: vscode.Uri; dark: vscode.Uri }
    | vscode.ThemeIcon;

  // Color for the terminal tab
  color?: vscode.ThemeColor;

  // Whether the terminal can be interacted with
  isTransient?: boolean;
}
```

## Terminal Profiles

Terminal profiles appear in the terminal dropdown and allow users to quickly create specialized terminals.

### Registering a Terminal Profile Provider

```typescript
vscode.window.registerTerminalProfileProvider("myExtension.devProfile", {
  provideTerminalProfile(
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<vscode.TerminalProfile> {
    return {
      options: {
        name: "Dev Environment",
        shellPath: "/bin/bash",
        shellArgs: ["--rcfile", ".devrc"],
        env: {
          PATH: "/custom/bin:${env:PATH}",
          DEV_MODE: "true",
        },
      },
    };
  },
});
```

### Declaring in Manifest

```json
{
  "contributes": {
    "terminal": {
      "profiles": [
        {
          "id": "myExtension.devProfile",
          "title": "Dev Environment",
          "icon": "$(terminal)"
        }
      ]
    }
  }
}
```

## Sending Text to Terminals

Send commands to terminals:

```typescript
// Send text without executing
terminal.sendText("npm run build");

// Send text and execute (add newline)
terminal.sendText("npm run build\n");

// Send to the active terminal
vscode.window.activeTerminal?.sendText('echo "Hello"\n');
```

## Terminal Events

Listen for terminal events:

```typescript
// Terminal opened
vscode.window.onDidOpenTerminal((terminal) => {
  console.log("Terminal opened:", terminal.name);
});

// Terminal closed
vscode.window.onDidCloseTerminal((terminal) => {
  console.log("Terminal closed:", terminal.name);
});

// Active terminal changed
vscode.window.onDidChangeActiveTerminal((terminal) => {
  console.log("Active terminal:", terminal?.name);
});

// Terminal state changed
vscode.window.onDidChangeTerminalState((event) => {
  console.log("Terminal state:", event.terminal.state);
});
```

## Pseudoterminal

Create custom terminals with `Pseudoterminal`:

```typescript
class MyPseudoterminal implements vscode.Pseudoterminal {
  private writeEmitter = new vscode.EventEmitter<string>();
  onDidWrite: vscode.Event<string> = this.writeEmitter.event;

  private closeEmitter = new vscode.EventEmitter<number>();
  onDidClose: vscode.Event<number> = this.closeEmitter.event;

  private dimensions: vscode.TerminalDimensions | undefined;

  open(initialDimensions: vscode.TerminalDimensions | undefined): void {
    this.dimensions = initialDimensions;
    this.writeEmitter.fire("Welcome to My Terminal\r\n");
    this.writeEmitter.fire('Type "help" for available commands\r\n');
    this.writeEmitter.fire("$ ");
  }

  close(): void {
    // Clean up resources
  }

  handleInput(data: string): void {
    // Handle user input
    if (data === "\r") {
      // Enter pressed - process command
      this.writeEmitter.fire("\r\n");
      this.processCommand();
    } else if (data === "\x7f") {
      // Backspace
      this.writeEmitter.fire("\b \b");
    } else {
      // Echo the character
      this.writeEmitter.fire(data);
    }
  }

  setDimensions(dimensions: vscode.TerminalDimensions): void {
    this.dimensions = dimensions;
  }

  private currentLine: string = "";

  private processCommand(): void {
    const command = this.currentLine.trim();
    this.currentLine = "";

    switch (command) {
      case "help":
        this.writeEmitter.fire("Available commands:\r\n");
        this.writeEmitter.fire("  help - Show this help\r\n");
        this.writeEmitter.fire("  exit - Close terminal\r\n");
        break;
      case "exit":
        this.closeEmitter.fire(0);
        return;
      default:
        if (command) {
          this.writeEmitter.fire(`Unknown command: ${command}\r\n`);
        }
    }

    this.writeEmitter.fire("$ ");
  }
}

// Create a terminal with the pseudoterminal
const terminal = vscode.window.createTerminal({
  name: "My Custom Terminal",
  pty: new MyPseudoterminal(),
});
terminal.show();
```

## Terminal Location

Control where terminals appear:

```typescript
// Terminal in the panel (default)
vscode.window.createTerminal({
  name: "Panel Terminal",
  location: vscode.TerminalLocation.Panel,
});

// Terminal in the editor area
vscode.window.createTerminal({
  name: "Editor Terminal",
  location: vscode.TerminalLocation.Editor,
});

// Terminal split from an existing terminal
vscode.window.createTerminal({
  name: "Split Terminal",
  location: { parentTerminal: existingTerminal },
});
```

## Terminal Groups

Work with terminal groups:

```typescript
// Get all terminals
const terminals = vscode.window.terminals;

// Get terminal groups
const groups = vscode.window.terminals.reduce((groups, terminal) => {
  // Terminals in the same group share a tab
  return groups;
}, [] as vscode.Terminal[][]);

// Show terminal in a specific column
terminal.show(true); // Preserve focus
terminal.show(false); // Take focus
```

## Environment Variables

Control environment variables for terminals:

```typescript
// Set environment variables for all terminals
context.environmentVariableCollection.append("MY_VAR", "my-value");
context.environmentVariableCollection.prepend("PATH", "/custom/bin:");
context.environmentVariableCollection.replace("EDITOR", "code");

// Remove an environment variable
context.environmentVariableCollection.delete("MY_VAR");

// Clear all custom environment variables
context.environmentVariableCollection.clear();

// Scope to specific workspace folders
context.environmentVariableCollection.prepend("PATH", "/project/bin:", {
  applyAtShellIntegration: true,
  scope: vscode.EnvironmentVariableScope.Workspace,
});
```

## Shell Integration

Detect when shell integration is available:

```typescript
vscode.window.onDidChangeTerminalShellIntegration((event) => {
  console.log("Shell integration changed:", event);
});

// Access shell integration
const terminal = vscode.window.activeTerminal;
if (terminal?.shellIntegration) {
  // Execute command with shell integration
  terminal.shellIntegration.executeCommand("npm test");
}
```

## Terminal Links

Create clickable links in terminals:

```typescript
vscode.window.registerTerminalLinkProvider({
  provideTerminalLinks(context, token) {
    // Find links in the terminal line
    const links: vscode.TerminalLink[] = [];
    const regex = /error:\s+(.+):(\d+):(\d+)/g;
    let match;

    while ((match = regex.exec(context.line)) !== null) {
      links.push({
        startIndex: match.index,
        length: match[0].length,
        tooltip: "Open file",
        data: {
          file: match[1],
          line: parseInt(match[2]),
          column: parseInt(match[3]),
        },
      });
    }

    return links;
  },

  handleTerminalLink(link: vscode.TerminalLink) {
    const { file, line, column } = link.data;
    vscode.window.showTextDocument(vscode.Uri.file(file), {
      selection: new vscode.Range(line - 1, column - 1, line - 1, column - 1),
    });
  },
});
```

## Best Practices

**Use terminal profiles**: Provide pre-configured terminals for common workflows.

**Clean up resources**: Dispose of terminals when no longer needed.

**Handle shell differences**: Account for different shells (bash, zsh, PowerShell, cmd).

**Provide helpful names**: Name terminals descriptively so users can identify them.

**Consider transient terminals**: Use `isTransient` for terminals that don't need to persist.

**Test on all platforms**: Terminal behavior can differ between Windows, macOS, and Linux.

## Common Patterns

### Run Command in Terminal

```typescript
async function runInTerminal(
  command: string,
  name: string = "Command",
): Promise<void> {
  const terminal = vscode.window.createTerminal({ name });
  terminal.show();
  terminal.sendText(command);
}
```

### Development Environment Terminal

```typescript
vscode.window.registerTerminalProfileProvider("myExtension.devEnv", {
  provideTerminalProfile() {
    const workspace = vscode.workspace.workspaceFolders?.[0];
    return {
      options: {
        name: "Dev Environment",
        cwd: workspace?.uri.fsPath,
        env: {
          NODE_ENV: "development",
          PATH: `${workspace?.uri.fsPath}/node_modules/.bin:${process.env.PATH}`,
        },
        shellArgs: ["-c", "source .env && exec bash"],
      },
    };
  },
});
```

### Interactive Terminal

```typescript
class InteractiveTerminal implements vscode.Pseudoterminal {
  private writeEmitter = new vscode.EventEmitter<string>();
  onDidWrite = this.writeEmitter.event;

  private closeEmitter = new vscode.EventEmitter<number>();
  onDidClose = this.closeEmitter.event;

  private buffer: string = "";

  open() {
    this.writeEmitter.fire("Interactive Terminal\r\n");
    this.writeEmitter.fire("> ");
  }

  close() {}

  handleInput(data: string) {
    if (data === "\r") {
      this.writeEmitter.fire("\r\n");
      this.executeCommand(this.buffer);
      this.buffer = "";
      this.writeEmitter.fire("> ");
    } else if (data === "\x03") {
      this.closeEmitter.fire(0);
    } else {
      this.buffer += data;
      this.writeEmitter.fire(data);
    }
  }

  private executeCommand(command: string) {
    // Process command
    this.writeEmitter.fire(`Executed: ${command}\r\n`);
  }
}
```

## Next Steps

- [Tasks](tasks.md): Define and run tasks
- [Debug Adapters](debug-adapters.md): Implement debug support
- [VSCode Terminal Documentation](https://code.visualstudio.com/api/extension-guides/terminal)

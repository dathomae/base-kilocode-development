# Tasks

Tasks allow extensions to define and run build scripts, linters, and other tools. Users can run tasks through the Task Explorer or Command Palette, and tasks can integrate with VSCode's problem matching system.

## Task Provider

Register a task provider to contribute tasks:

```typescript
import * as vscode from "vscode";

class MyTaskProvider implements vscode.TaskProvider {
  provideTasks(
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<vscode.Task[]> {
    return [
      new vscode.Task(
        { type: "myTask" },
        vscode.TaskScope.Workspace,
        "Build",
        "myExtension",
        new vscode.ShellExecution("npm run build"),
      ),
    ];
  }

  resolveTask(
    task: vscode.Task,
    token: vscode.CancellationToken,
  ): vscode.ProviderResult<vscode.Task> {
    // Resolve a task definition to a full task
    return task;
  }
}

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.tasks.registerTaskProvider("myTask", new MyTaskProvider()),
  );
}
```

## Task Definition

Define task types in `package.json`:

```json
{
  "contributes": {
    "taskDefinitions": [
      {
        "type": "myTask",
        "required": ["command"],
        "properties": {
          "command": {
            "type": "string",
            "description": "The command to run"
          },
          "args": {
            "type": "array",
            "items": { "type": "string" },
            "description": "Command arguments"
          },
          "cwd": {
            "type": "string",
            "description": "Working directory"
          }
        }
      }
    ]
  }
}
```

## Creating Tasks

### Task Structure

```typescript
const task = new vscode.Task(
  // Task definition
  { type: "myTask", command: "build" },

  // Task scope (Workspace or specific folder)
  vscode.TaskScope.Workspace,

  // Task name
  "Build Project",

  // Source (shown in task list)
  "myExtension",

  // Execution
  new vscode.ShellExecution("npm", ["run", "build"]),

  // Problem matcher (optional)
  "$eslint-compact",
);
```

### Shell Execution

Run shell commands:

```typescript
// Simple command
new vscode.ShellExecution("npm run build");

// Command with arguments
new vscode.ShellExecution("npm", ["run", "build"]);

// With options
new vscode.ShellExecution("npm run build", {
  cwd: "/path/to/project",
  env: { NODE_ENV: "production" },
  shellArgs: ["-c"],
  executable: "/bin/bash",
});
```

### Process Execution

Run processes directly:

```typescript
new vscode.ProcessExecution("/usr/bin/node", ["script.js"], {
  cwd: "/path/to/project",
  env: { NODE_ENV: "production" },
});
```

### Custom Execution

Run custom code:

```typescript
new vscode.CustomExecution(
  async (
    resolvedDefinition: vscode.TaskDefinition,
  ): Promise<vscode.Pseudoterminal> => {
    return new MyPseudoterminal();
  },
);
```

## Task Properties

### Group

Organize tasks into groups:

```typescript
task.group = vscode.TaskGroup.Build;
// or
task.group = vscode.TaskGroup.Test;
// or
task.group = vscode.TaskGroup.Clean;
```

### Presentation

Control how task output is displayed:

```typescript
task.presentationOptions = {
  reveal: vscode.TaskRevealKind.Always, // Always, Silent, Never
  panel: vscode.TaskPanelKind.Shared, // Shared, Dedicated, New
  clear: true, // Clear terminal before running
  showReuseMessage: false, // Show "Terminal will be reused" message
  focus: false, // Focus terminal when task starts
};
```

### Problem Matchers

Match errors and warnings in task output:

```typescript
// Built-in problem matchers
task.problemMatchers = ["$eslint-compact", "$tsc"];

// Custom problem matcher
task.problemMatchers = ["$myProblemMatcher"];
```

### Detail

Provide additional information:

```typescript
task.detail = "Builds the project for production";
```

### Identifier

Set a unique identifier for the task:

```typescript
task.definition.identifier = "build-production";
```

## Problem Matchers

Define custom problem matchers in `package.json`:

```json
{
  "contributes": {
    "problemMatchers": [
      {
        "name": "myMatcher",
        "owner": "myExtension",
        "fileLocation": ["relative", "${workspaceFolder}"],
        "pattern": {
          "regexp": "^(.*):(\\d+):(\\d+)\\s+-\\s+(.*)$",
          "file": 1,
          "line": 2,
          "column": 3,
          "message": 4
        }
      }
    ]
  }
}
```

### Pattern Properties

The pattern defines how to parse output lines:

```json
{
  "pattern": {
    "regexp": "^(.*):(\\d+):(\\d+):\\s+(warning|error):\\s+(.*)$",
    "file": 1,
    "line": 2,
    "column": 3,
    "severity": 4,
    "message": 5
  }
}
```

### Multi-Line Patterns

Match patterns across multiple lines:

```json
{
  "pattern": [
    {
      "regexp": "^([^\\s].*)$",
      "file": 1
    },
    {
      "regexp": "^\\s+(\\d+):(\\d+)\\s+-\\s+(.*)$",
      "line": 1,
      "column": 2,
      "message": 3,
      "loop": true
    }
  ]
}
```

### Background Problem Matchers

For tasks that run continuously:

```json
{
  "name": "myBackgroundMatcher",
  "owner": "myExtension",
  "background": {
    "activeOnStart": true,
    "beginsPattern": "^Starting build",
    "endsPattern": "^Build complete"
  },
  "pattern": {
    "regexp": "^(.*):(\\d+):\\s+(.*)$",
    "file": 1,
    "line": 2,
    "message": 3
  }
}
```

## Running Tasks

### Execute a Task

```typescript
const task = new vscode.Task(/* ... */);

const execution = await vscode.tasks.executeTask(task);

// Wait for completion
const promise = new Promise<void>((resolve) => {
  const disposable = vscode.tasks.onDidEndTask((e) => {
    if (e.execution === execution) {
      disposable.dispose();
      resolve();
    }
  });
});

await promise;
```

### Execute with Terminal

```typescript
const execution = await vscode.tasks.executeTask(task);

// The terminal is created automatically
// Access it through the execution
const terminal = execution.task.definition.terminal;
```

## Task Events

Listen for task events:

```typescript
// Task started
vscode.tasks.onDidStartTask((e) => {
  console.log("Task started:", e.execution.task.name);
});

// Task completed
vscode.tasks.onDidEndTask((e) => {
  console.log("Task ended:", e.execution.task.name);
});

// Task process started
vscode.tasks.onDidStartTaskProcess((e) => {
  console.log("Process started:", e.processId);
});

// Task process ended
vscode.tasks.onDidEndTaskProcess((e) => {
  console.log("Process ended with code:", e.exitCode);
});
```

## Fetching Tasks

Get available tasks:

```typescript
const tasks = await vscode.tasks.fetchTasks({
  type: "myTask",
});

tasks.forEach((task) => {
  console.log(task.name);
});
```

## Custom Terminal

Implement `Pseudoterminal` for custom task execution:

```typescript
class MyPseudoterminal implements vscode.Pseudoterminal {
  private writeEmitter = new vscode.EventEmitter<string>();
  onDidWrite: vscode.Event<string> = this.writeEmitter.event;

  private closeEmitter = new vscode.EventEmitter<number>();
  onDidClose: vscode.Event<number> = this.closeEmitter.event;

  open(initialDimensions: vscode.TerminalDimensions | undefined): void {
    this.writeEmitter.fire("Starting task...\r\n");

    // Do work
    this.doWork()
      .then(() => {
        this.writeEmitter.fire("Task completed\r\n");
        this.closeEmitter.fire(0);
      })
      .catch((err) => {
        this.writeEmitter.fire(`Error: ${err.message}\r\n`);
        this.closeEmitter.fire(1);
      });
  }

  close(): void {
    // Clean up
  }

  handleInput(data: string): void {
    // Handle user input
    this.writeEmitter.fire(`Received: ${data}\r\n`);
  }

  private async doWork(): Promise<void> {
    // Implementation
  }
}
```

## Best Practices

**Use appropriate execution types**: ShellExecution for commands, ProcessExecution for direct process execution, CustomExecution for complex scenarios.

**Define problem matchers**: Help users see errors and warnings in their code.

**Group related tasks**: Use task groups to organize tasks in the UI.

**Provide clear names**: Task names should indicate what the task does.

**Handle cancellation**: Support task cancellation when possible.

**Test on different platforms**: Ensure shell commands work on Windows, macOS, and Linux.

## Common Patterns

### Build Task

```typescript
const buildTask = new vscode.Task(
  { type: "myTask", command: "build" },
  vscode.TaskScope.Workspace,
  "Build",
  "myExtension",
  new vscode.ShellExecution("npm", ["run", "build"]),
  "$tsc",
);
buildTask.group = vscode.TaskGroup.Build;
buildTask.presentationOptions = { reveal: vscode.TaskRevealKind.Silent };
```

### Watch Task

```typescript
const watchTask = new vscode.Task(
  { type: "myTask", command: "watch" },
  vscode.TaskScope.Workspace,
  "Watch",
  "myExtension",
  new vscode.ShellExecution("npm", ["run", "watch"]),
  "$tsc-watch",
);
watchTask.group = vscode.TaskGroup.Build;
watchTask.isBackground = true;
watchTask.presentationOptions = { reveal: vscode.TaskRevealKind.Never };
```

### Test Task

```typescript
const testTask = new vscode.Task(
  { type: "myTask", command: "test" },
  vscode.TaskScope.Workspace,
  "Run Tests",
  "myExtension",
  new vscode.ShellExecution("npm", ["test"]),
);
testTask.group = vscode.TaskGroup.Test;
```

## Next Steps

- [Debug Adapters](debug-adapters.md): Implement debug support
- [Terminal](terminal.md): Create terminal profiles
- [VSCode Tasks Documentation](https://code.visualstudio.com/api/extension-guides/task-provider)

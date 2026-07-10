# Advanced Topics

This section covers advanced extension features including debug adapters, task providers, terminal integration, and file system providers.

## Table of Contents

**[debug-adapters.md](debug-adapters.md)**
: Implementing debug adapters using the Debug Adapter Protocol

**[tasks.md](tasks.md)**
: Providing custom task definitions and problem matchers

**[terminal.md](terminal.md)**
: Creating terminal profiles and integrating with the terminal

**[file-system.md](file-system.md)**
: Implementing custom file system providers

## Quick Reference

### Debug Adapter

```typescript
vscode.debug.registerDebugAdapterDescriptorFactory("myDebug", {
  createDebugAdapterDescriptor(session) {
    return new vscode.DebugAdapterExecutable(process.execPath, [
      "--require",
      "./out/debugAdapter.js",
    ]);
  },
});
```

### Task Provider

```typescript
vscode.tasks.registerTaskProvider("myTask", {
  provideTasks() {
    return [
      new vscode.Task(
        { type: "myTask" },
        vscode.TaskScope.Workspace,
        "Build",
        "myExtension",
        new vscode.ShellExecution("npm run build"),
      ),
    ];
  },
  resolveTask(task) {
    return task;
  },
});
```

### Terminal Profile

```typescript
vscode.window.registerTerminalProfileProvider("myExtension.terminal", {
  provideTerminalProfile(token) {
    return {
      options: {
        name: "My Terminal",
        shellPath: "/bin/bash",
        shellArgs: ["--login"],
      },
    };
  },
});
```

### File System Provider

```typescript
vscode.workspace.registerFileSystemProvider(
  "myfs",
  new MyFileSystemProvider(),
  {
    isCaseSensitive: true,
  },
);
```

# Extension Lifecycle

Understanding the extension lifecycle is essential for properly initializing and cleaning up resources. Extensions go through distinct phases from discovery to deactivation, and handling each phase correctly ensures a smooth user experience.

## Lifecycle Phases

An extension progresses through these phases:

1. **Discovery**: VSCode reads the extension manifest
2. **Activation**: An activation event triggers extension loading
3. **Execution**: The extension provides its functionality
4. **Deactivation**: The extension cleans up and terminates

## The activate Function

The `activate` function is the entry point for extension code. VSCode calls this function when an activation event occurs:

```typescript
export function activate(
  context: vscode.ExtensionContext,
): void | Thenable<void> {
  // Extension initialization code
  console.log("Extension is now active");

  // Register commands, providers, and other contributions
  const disposable = vscode.commands.registerCommand(
    "myExtension.hello",
    () => {
      vscode.window.showInformationMessage("Hello!");
    },
  );

  // Add to subscriptions for automatic cleanup
  context.subscriptions.push(disposable);
}
```

### Function Signature

The `activate` function receives an `ExtensionContext` object and can return either void or a Promise:

```typescript
export function activate(context: vscode.ExtensionContext): void {
  // Synchronous activation
}

export async function activate(
  context: vscode.ExtensionContext,
): Promise<void> {
  // Asynchronous activation
  await initializeExtension();
}
```

When returning a Promise, VSCode waits for the Promise to resolve before considering the extension fully activated. Other extensions that depend on this one will wait for activation to complete.

### The ExtensionContext Object

The `ExtensionContext` provides access to essential extension resources:

```typescript
interface ExtensionContext {
  // Unique identifier for this extension
  readonly extensionId: string;

  // Absolute path to extension directory
  readonly extensionPath: string;

  // Path to extension's log directory
  readonly logPath: string;

  // Path to extension's global storage directory
  readonly globalStoragePath: string;

  // Path to extension's workspace storage directory
  readonly storagePath: string | undefined;

  // Persistent key-value storage (global)
  readonly globalState: Memento;

  // Persistent key-value storage (workspace-scoped)
  readonly workspaceState: Memento;

  // Resources to dispose on deactivation
  subscriptions: { dispose(): any }[];

  // Extension's package.json
  readonly extension: Extension<any>;

  // Environment variables collection
  readonly environmentVariableCollection: EnvironmentVariableCollection;

  // Secrets storage
  readonly secrets: SecretStorage;
}
```

### Storage Paths

The context provides several storage paths for different purposes:

**extensionPath**: The directory where the extension is installed. Use this to access bundled files:

```typescript
const iconPath = vscode.Uri.file(
  path.join(context.extensionPath, "images", "icon.png"),
);
```

**globalStoragePath**: A directory for storing data that persists across workspaces. Create files here for extension data that should survive workspace changes:

```typescript
const dataFile = path.join(context.globalStoragePath, "cache.json");
await fs.promises.writeFile(dataFile, JSON.stringify(data));
```

**storagePath**: A directory for workspace-specific data. This path is undefined when no workspace is open:

```typescript
if (context.storagePath) {
  const workspaceData = path.join(context.storagePath, "workspace-cache.json");
}
```

**logPath**: A directory for extension log files. Use this for diagnostic logs:

```typescript
const logFile = path.join(context.logPath, "extension.log");
```

### Persistent State

The `globalState` and `workspaceState` Memento objects provide key-value storage:

```typescript
// Store a value
await context.globalState.update("lastRunDate", new Date().toISOString());

// Retrieve a value
const lastRun = context.globalState.get<string>("lastRunDate");

// Store with a default value
const count = context.globalState.get<number>("runCount", 0);
await context.globalState.update("runCount", count + 1);

// Keys with a specific prefix
const keys = context.globalState.keys().filter((k) => k.startsWith("item-"));
```

**globalState**: Persists across all workspaces. Use for user preferences, cached data, and settings that apply everywhere.

**workspaceState**: Persists only for the current workspace. Use for workspace-specific data like project configuration or cached analysis results.

### Secrets Storage

The `secrets` property provides secure storage for sensitive data:

```typescript
// Store a secret
await context.secrets.store("apiToken", "secret-value");

// Retrieve a secret
const token = await context.secrets.get("apiToken");

// Delete a secret
await context.secrets.delete("apiToken");

// Listen for secret changes
context.secrets.onDidChange((e) => {
  if (e.key === "apiToken") {
    console.log("API token was updated");
  }
});
```

Secrets are stored securely using the operating system's credential storage.

### Subscriptions

The `subscriptions` array holds disposable resources that VSCode cleans up automatically when the extension deactivates:

```typescript
export function activate(context: vscode.ExtensionContext) {
  // Register a command
  const commandDisposable = vscode.commands.registerCommand(
    "myExtension.hello",
    helloHandler,
  );
  context.subscriptions.push(commandDisposable);

  // Register a completion provider
  const completionDisposable = vscode.languages.registerCompletionItemProvider(
    { language: "javascript" },
    new MyCompletionProvider(),
    ".",
  );
  context.subscriptions.push(completionDisposable);

  // Register an event listener
  const saveDisposable = vscode.workspace.onDidSaveTextDocument((doc) => {
    console.log("Saved:", doc.uri);
  });
  context.subscriptions.push(saveDisposable);
}
```

All objects added to `subscriptions` must implement the `Disposable` interface with a `dispose()` method.

## The deactivate Function

The `deactivate` function is called when the extension is being unloaded:

```typescript
export function deactivate(): void | Thenable<void> {
  console.log("Extension is deactivating");
  // Clean up resources not handled by subscriptions
}
```

### When Deactivate is Called

Deactivation occurs when:

- VSCode is closing
- The extension is uninstalled
- The extension is updated (old version deactivates)
- The extension is disabled

### What to Clean Up

If you use the `subscriptions` pattern, most cleanup happens automatically. The `deactivate` function is needed for:

**Background processes**: Stop any running servers, timers, or long-running operations:

```typescript
let server: Server | undefined;

export function activate(context: vscode.ExtensionContext) {
  server = startServer();
}

export async function deactivate() {
  if (server) {
    await server.stop();
  }
}
```

**External resources**: Close database connections, file handles, or network connections:

```typescript
let db: Database;

export async function deactivate() {
  if (db) {
    await db.close();
  }
}
```

**State persistence**: Save any unsaved state before shutdown:

```typescript
export async function deactivate() {
  await savePendingChanges();
}
```

### Deactivate is Optional

If your extension only uses subscriptions for resource management, you can omit the `deactivate` function entirely:

```typescript
export function activate(context: vscode.ExtensionContext) {
  // All resources go into subscriptions
  context.subscriptions.push(
    vscode.commands.registerCommand("myExtension.hello", () => {}),
  );
}

// No deactivate function needed
```

## Async Activation

For extensions that need to perform asynchronous initialization:

```typescript
export async function activate(context: vscode.ExtensionContext) {
  // Show progress during initialization
  await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Window,
      title: "Loading extension...",
      cancellable: false,
    },
    async (progress) => {
      progress.report({ message: "Initializing..." });
      await initializeLanguageServer();

      progress.report({ message: "Loading configuration..." });
      await loadConfiguration();

      progress.report({ message: "Ready!" });
    },
  );

  // Register commands after initialization
  registerCommands(context);
}
```

## Handling Extension Updates

When an extension updates, the old version deactivates and the new version activates. Handle migration between versions:

```typescript
export async function activate(context: vscode.ExtensionContext) {
  const currentVersion = context.extension.packageJSON.version;
  const previousVersion = context.globalState.get<string>("version");

  if (previousVersion && previousVersion !== currentVersion) {
    await migrateFrom(previousVersion, currentVersion);
  }

  await context.globalState.update("version", currentVersion);
}

async function migrateFrom(from: string, to: string): Promise<void> {
  console.log(`Migrating from ${from} to ${to}`);
  // Perform migration logic
}
```

## Checking Activation State

You can check if an extension is active:

```typescript
const extension = vscode.extensions.getExtension("publisher.extensionName");
if (extension) {
  if (extension.isActive) {
    // Extension is active
    const api = extension.exports;
  } else {
    // Extension is not yet active
    await extension.activate();
  }
}
```

## Exports API

Extensions can export an API for other extensions to consume:

```typescript
// In your extension
export function activate(context: vscode.ExtensionContext): MyApi {
  return {
    doSomething: () => {
      /* ... */
    },
    getValue: () => 42,
  };
}

// In another extension
const extension = vscode.extensions.getExtension("publisher.myExtension");
if (extension?.isActive) {
  const api = extension.exports as MyApi;
  api.doSomething();
}
```

## Lifecycle Events

VSCode provides events related to extension lifecycle:

**onDidChangeActiveEditor**: Fires when the active editor changes
**onDidChangeTextEditorSelection**: Fires when editor selection changes
**onDidOpenTextDocument**: Fires when a document opens
**onDidCloseTextDocument**: Fires when a document closes
**onDidChangeConfiguration**: Fires when settings change

These events are useful for extensions that need to respond to user activity.

## Best Practices

**Keep activation fast**: The `activate` function should return quickly. Defer expensive operations until needed.

**Use subscriptions**: Add all disposables to `context.subscriptions` for automatic cleanup.

**Handle deactivation**: Clean up all resources, especially background processes and external connections.

**Persist state appropriately**: Use `globalState` for cross-workspace data, `workspaceState` for workspace-specific data, and `secrets` for sensitive information.

**Show progress for long initialization**: If activation requires significant work, show a progress indicator.

**Test lifecycle behavior**: Verify that your extension activates correctly and cleans up properly when deactivated.

## Common Patterns

### Lazy Initialization

Defer expensive initialization until features are actually used:

```typescript
let initialized = false;

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.commands.registerCommand("myExtension.analyze", async () => {
      if (!initialized) {
        await initialize();
        initialized = true;
      }
      performAnalysis();
    }),
  );
}
```

### Singleton Services

Create services that persist throughout the extension's lifetime:

```typescript
let myService: MyService | undefined;

export function activate(context: vscode.ExtensionContext) {
  myService = new MyService(context);
  context.subscriptions.push(myService);
}

export function deactivate() {
  myService = undefined;
}
```

### Configuration-Driven Activation

Initialize based on user settings:

```typescript
export function activate(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration("myExtension");

  if (config.get<boolean>("enableFeature")) {
    enableFeature(context);
  }

  context.subscriptions.push(
    vscode.workspace.onDidChangeConfiguration((e) => {
      if (e.affectsConfiguration("myExtension.enableFeature")) {
        const newValue = config.get<boolean>("enableFeature");
        if (newValue) {
          enableFeature(context);
        } else {
          disableFeature();
        }
      }
    }),
  );
}
```

## Next Steps

- [Extension Manifest](extension-manifest.md): Configure your extension's metadata and contributions
- [Activation Events](activation-events.md): Control when your extension activates
- [Commands](../02-core-capabilities/commands.md): Register commands in your extension

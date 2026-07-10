# Activation Events

Activation events determine when VSCode loads and activates an extension. VSCode uses lazy activation to keep startup fast, loading extensions only when their functionality is needed.

## How Activation Works

When VSCode starts, it reads all extension manifests but does not immediately load the extension code. Instead, it monitors for activation events. When an event matches an extension's declared activation events, VSCode loads the extension and calls its `activate` function.

This lazy loading approach ensures that extensions do not slow down VSCode startup. Extensions should declare specific activation events rather than activating immediately.

## Declaring Activation Events

Activation events are declared in the `package.json` manifest:

```json
{
  "activationEvents": ["onLanguage:python", "onCommand:myExtension.format"]
}
```

An extension activates when any of its declared activation events occur. Multiple activation events are combined with OR logic.

## Event Types

### onLanguage

Activates when a file with the specified language ID opens:

```json
{
  "activationEvents": ["onLanguage:python"]
}
```

Common language IDs include `javascript`, `typescript`, `python`, `java`, `cpp`, `json`, `markdown`, and many more. This event is appropriate for language support extensions that provide features like completion, hover, or diagnostics.

### onCommand

Activates when a specific command is invoked:

```json
{
  "activationEvents": ["onCommand:myExtension.sayHello"]
}
```

The command must also be declared in the `contributes.commands` section. This event ensures the extension loads only when the user explicitly invokes its command.

### onDebug

Activates when starting a debug session:

```json
{
  "activationEvents": ["onDebug"]
}
```

For more specific activation, use:

- `onDebugInitialConfigurations`: Activate when providing initial debug configurations
- `onDebugResolve:<type>`: Activate when resolving debug configurations for a specific type

### onView

Activates when a specific view becomes visible:

```json
{
  "activationEvents": ["onView:myExtension.explorer"]
}
```

The view must be declared in `contributes.views`. This event is useful for extensions that provide custom views in the Explorer or other containers.

### onFileSystem

Activates when accessing a file with the specified scheme:

```json
{
  "activationEvents": ["onFileSystem:sftp"]
}
```

This event is used by extensions that provide custom file system providers, such as remote file access or version control systems.

### onSearch

Activates when performing a search with the specified scheme:

```json
{
  "activationEvents": ["onSearch:sftp"]
}
```

### workspaceContains

Activates when the workspace contains files matching a glob pattern:

```json
{
  "activationEvents": [
    "workspaceContains:**/.eslintrc.json",
    "workspaceContains:package.json"
  ]
}
```

This event is useful for tools that apply to specific project types. The glob pattern supports `*` for single-level wildcards and `**` for multi-level wildcards.

### onStartupFinished

Activates after VSCode completes startup:

```json
{
  "activationEvents": ["onStartupFinished"]
}
```

This event activates the extension shortly after VSCode becomes ready. Use this sparingly, as it affects startup performance. Prefer more specific activation events when possible.

### onWebviewPanel

Activates when a webview panel with the specified view type is restored:

```json
{
  "activationEvents": ["onWebviewPanel:myExtension.preview"]
}
```

### onTerminalProfile

Activates when a terminal profile is created:

```json
{
  "activationEvents": ["onTerminalProfile:myExtension.terminal"]
}
```

### onAuthenticationRequest

Activates when an authentication session for the specified provider is requested:

```json
{
  "activationEvents": ["onAuthenticationRequest:github"]
}
```

### onRender

Activates when rendering markdown with the specified extension:

```json
{
  "activationEvents": ["onRender:myExtension.markdown"]
}
```

### onTaskType

Activates when resolving a task of the specified type:

```json
{
  "activationEvents": ["onTaskType:myExtension.build"]
}
```

### onWalkthrough

Activates when a walkthrough is opened:

```json
{
  "activationEvents": ["onWalkthrough:myExtension.welcome"]
}
```

## The Star Activation Event

The `*` activation event causes the extension to activate on VSCode startup:

```json
{
  "activationEvents": ["*"]
}
```

This event should be used sparingly. Extensions that activate immediately impact VSCode startup time. Only use this when the extension must run regardless of user activity, such as:

- Extensions that monitor file changes across all files
- Extensions that provide background services
- Extensions that modify VSCode's appearance globally

If you must use `*`, consider whether your extension can defer some initialization until actually needed.

## Implicit Activation Events

Some contribution points implicitly add activation events:

**commands**: Declaring a command in `contributes.commands` implicitly adds `onCommand:<commandId>` if no activation events are specified.

**languages**: Declaring a language in `contributes.languages` implicitly adds `onLanguage:<languageId>`.

**debuggers**: Declaring a debugger in `contributes.debuggers` implicitly adds `onDebug`.

These implicit events are added only when the `activationEvents` array is empty or not present. Explicitly declaring activation events overrides implicit behavior.

## Activation Event Combinations

Extensions often need multiple activation events to cover all use cases:

```json
{
  "activationEvents": [
    "onLanguage:javascript",
    "onLanguage:typescript",
    "onCommand:myExtension.fixAll",
    "workspaceContains:**/.eslintrc.*"
  ]
}
```

This extension activates when:

- Opening a JavaScript file
- Opening a TypeScript file
- Running the fix-all command
- Opening a workspace with an ESLint configuration file

## Checking Activation State

Extensions can check if they are activated programmatically:

```typescript
const extension = vscode.extensions.getExtension("myPublisher.myExtension");
if (extension) {
  const isActive = extension.isActive;
  const exports = extension.exports; // Available after activation
}
```

## Activation Timing

The time between an activation event occurring and the `activate` function completing affects user experience. Extensions should:

**Minimize activation time**: Defer expensive initialization until needed. The `activate` function should return quickly.

**Show progress for long operations**: If activation requires significant work, show a progress indicator:

```typescript
export async function activate(context: vscode.ExtensionContext) {
  await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Window,
      title: "Initializing extension...",
    },
    async () => {
      // Perform initialization
    },
  );
}
```

**Use async activation**: The `activate` function can return a Promise. VSCode waits for this Promise to resolve before considering the extension fully activated.

## Debugging Activation

To debug activation issues:

**Check the Activation Times**: Open the Command Palette and run "Developer: Show Running Extensions". This displays activation times for all extensions.

**Use Console Logging**: Add logging to the `activate` function to verify when activation occurs:

```typescript
export function activate(context: vscode.ExtensionContext) {
  console.log("Extension activated at:", new Date().toISOString());
}
```

**Inspect Extension Host**: Use "Developer: Open Webview Developer Tools" to access the Extension Host's developer tools.

## Best Practices

**Be specific**: Declare the most specific activation events possible. Avoid `*` unless absolutely necessary.

**Cover all use cases**: Ensure activation events cover all scenarios where your extension provides functionality.

**Consider user workflow**: Think about when users need your extension's features and activate at that point.

**Test activation**: Verify that your extension activates at the right time by testing in the Extension Development Host.

**Document requirements**: If your extension requires specific conditions to work, document these clearly.

## Common Patterns

### Language Extension

```json
{
  "activationEvents": ["onLanguage:myLanguage", "onCommand:myExtension.compile"]
}
```

### Tool Integration

```json
{
  "activationEvents": [
    "workspaceContains:**/config.yml",
    "onCommand:myExtension.runTool"
  ]
}
```

### Theme Extension

```json
{
  "activationEvents": ["*"]
}
```

Theme extensions typically activate immediately to apply the theme.

### Debugger Extension

```json
{
  "activationEvents": ["onDebug", "onDebugResolve:myRuntime"]
}
```

## Next Steps

- [Extension Lifecycle](extension-lifecycle.md): Learn about the activate and deactivate functions
- [Extension Manifest](extension-manifest.md): Understand all manifest configuration options
- [Commands](../02-core-capabilities/commands.md): Register commands that trigger activation

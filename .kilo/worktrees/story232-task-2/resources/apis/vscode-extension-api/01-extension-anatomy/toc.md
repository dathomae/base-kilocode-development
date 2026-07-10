# Extension Anatomy

This section covers the fundamental structure of VSCode extensions, including the manifest file, activation events, and extension lifecycle.

## Table of Contents

**[extension-manifest.md](extension-manifest.md)**
: The package.json manifest file and its configuration options

**[activation-events.md](activation-events.md)**
: How and when extensions activate in VSCode

**[extension-lifecycle.md](extension-lifecycle.md)**
: The lifecycle of an extension from activation to deactivation

## Quick Reference

### Essential Manifest Fields

Every extension must include these fields in `package.json`:

- `name`: The extension's identifier (lowercase, no spaces)
- `displayName`: The human-readable name shown in the marketplace
- `version`: Semantic version number
- `engines.vscode`: Minimum VSCode version supported
- `main`: Entry point file path
- `activationEvents`: When the extension should activate

### Common Activation Events

| Event                      | Description                                            |
| -------------------------- | ------------------------------------------------------ |
| `onCommand:<commandId>`    | Activate when a specific command is invoked            |
| `onLanguage:<languageId>`  | Activate when opening a file of the specified language |
| `onView:<viewId>`          | Activate when a specific view becomes visible          |
| `workspaceContains:<glob>` | Activate when workspace contains matching files        |
| `onFileSystem:<scheme>`    | Activate when accessing files with the scheme          |
| `onDebug`                  | Activate when starting a debug session                 |
| `*`                        | Activate on VSCode startup (use sparingly)             |

### Lifecycle Functions

```typescript
export function activate(context: vscode.ExtensionContext): void {
  // Called when extension activates
  // Register commands, providers, and subscriptions
}

export function deactivate(): void {
  // Called when extension deactivates
  // Clean up resources (optional if using subscriptions)
}
```

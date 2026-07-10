# IntelliJ Platform Architecture

The IntelliJ Platform is a modular framework designed for building IDEs and developer tools. Understanding its architecture helps you create plugins that integrate well with the platform.

## Platform Overview

The platform consists of several layers, each building on the one below:

### Core Layer

The core layer provides fundamental services that everything else depends on:

- **Application**: The singleton representing the running IDE instance
- **Project**: A user-opened project with its configuration and files
- **Module**: A logical grouping of source files and dependencies within a project
- **Virtual File System**: An abstraction over files regardless of their location

### Component Model

The platform uses a component-based architecture. Components are classes managed by the platform with defined lifecycles. There are three levels of components:

| Level       | Scope       | Created When   | Disposed When  |
| ----------- | ----------- | -------------- | -------------- |
| Application | Global      | IDE starts     | IDE closes     |
| Project     | Per-project | Project opens  | Project closes |
| Module      | Per-module  | Module created | Module removed |

Components are registered in `plugin.xml` and instantiated by the platform. The platform handles dependency injection between components.

### Service Architecture

Modern plugins use services instead of components. Services provide a cleaner API and better testability. Like components, services have three scopes:

- **Application-level**: Shared across all projects
- **Project-level**: One instance per project
- **Module-level**: One instance per module

## Threading Model

The platform has specific threading rules that plugins must follow.

### Event Dispatch Thread (EDT)

All UI operations must run on the EDT. This includes:

- Creating or modifying UI components
- Showing dialogs or notifications
- Updating the editor

Use `ApplicationManager.getApplication().invokeLater()` to run code on the EDT:

```kotlin
ApplicationManager.getApplication().invokeLater {
    // UI code here
}
```

### Background Threads

Long-running operations should not block the EDT. Use background threads for:

- File I/O
- Network requests
- Expensive computations

The platform provides utilities for running background tasks:

```kotlin
ProgressManager.getInstance().run(object : Task.Backgroundable(project, "Processing") {
    override fun run(indicator: ProgressIndicator) {
        // Background work here
    }
})
```

### Read and Write Actions

Accessing platform data structures requires proper read/write locks:

- **Read Action**: Required to read PSI, virtual files, or project structure
- **Write Action**: Required to modify PSI, virtual files, or project structure

```kotlin
// Read action
ReadAction.compute<Unit, Exception> {
    // Read PSI or other platform data
}

// Write action
WriteAction.run<Exception> {
    // Modify PSI or other platform data
}
```

Write actions must run on the EDT. Read actions can run on any thread but must not conflict with write actions.

## Plugin Loading

Plugins are loaded in a specific order based on dependencies:

1. The platform loads core components
2. Plugins with no dependencies load first
3. Plugins with dependencies load after their dependencies
4. Each plugin's components and services are initialized

The loading process is controlled by the `plugin.xml` configuration. Dependencies are declared using the `<depends>` element.

## Message Bus

The platform uses a publish-subscribe model for event communication. The message bus provides a decoupled way for components to communicate.

### Publishing Events

```kotlin
val bus = project.messageBus
val publisher = bus.syncPublisher(MyTopic)
publisher.onEvent(eventData)
```

### Subscribing to Events

```kotlin
project.messageBus.connect().subscribe(MyTopic, object : MyListener {
    override fun onEvent(data: EventData) {
        // Handle event
    }
})
```

Subscriptions are automatically disposed when the connection is disposed, preventing memory leaks.

## Extension System

The platform is extensible through extension points. An extension point defines a place where plugins can contribute functionality.

### Built-in Extension Points

The platform provides hundreds of extension points for common tasks:

- `com.intellij.action` - Register actions
- `com.intellij.service` - Register services
- `com.intellij.toolWindow` - Register tool windows
- `com.intellij.fileType` - Register file types

### Custom Extension Points

Plugins can define their own extension points for other plugins to use:

```xml
<extensionPoints>
    <extensionPoint name="myExtension"
                    interface="com.example.MyExtensionInterface"/>
</extensionPoints>
```

Other plugins can then contribute implementations:

```xml
<extensions defaultExtensionNs="com.example">
    <myExtension implementation="com.otherplugin.MyExtensionImpl"/>
</extensions>
```

## Disposables

The platform uses a disposable pattern for resource management. Every component, service, and subscription has an associated disposable parent.

### Creating Disposables

```kotlin
val disposable = Disposer.newDisposable()
try {
    // Use resources
} finally {
    Disposer.dispose(disposable)
}
```

### Hierarchical Disposal

Disposables form a tree structure. When a parent is disposed, all children are disposed automatically:

```kotlin
val parent = Disposer.newDisposable()
val child = Disposer.newDisposable(parent)

Disposer.dispose(parent) // Also disposes child
```

This pattern ensures resources are cleaned up when projects close or plugins unload.

## Related Topics

- [Plugin Lifecycle](plugin-lifecycle.md) - How plugins are loaded and unloaded
- [Services](services.md) - Creating and using services
- [Extension Points](extension-points.md) - Detailed extension point usage

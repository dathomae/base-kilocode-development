# Workspace Model

The Workspace Model is the modern API for managing project structure in the IntelliJ Platform. It provides a unified, transactional model for representing projects, modules, and their dependencies.

## Overview

### What Is the Workspace Model?

The Workspace Model replaces the legacy project model APIs with a more robust, consistent interface. It provides:

- **Unified Model**: Single source of truth for project structure
- **Transaction Support**: Atomic modifications with rollback
- **Event System**: Consistent notifications for changes
- **Extensibility**: Custom entities for plugin-specific data

### When to Use Workspace Model

Use the Workspace Model for:

- Creating or modifying modules
- Managing dependencies
- Working with content roots
- Storing custom project data

The legacy APIs (`ModuleManager`, `ModuleRootManager`) still work but the Workspace Model is preferred for new development.

## Core Concepts

### Workspace Model Access

```kotlin
val workspaceModel = WorkspaceModel.getInstance(project)
```

### Entity-Based Model

The Workspace Model uses entities to represent project structure:

| Entity                    | Description                         |
| ------------------------- | ----------------------------------- |
| `ModuleEntity`            | A module in the project             |
| `ContentRootEntity`       | A content root directory            |
| `SourceRootEntity`        | A source root within a content root |
| `LibraryDependencyEntity` | Dependency on a library             |
| `ModuleDependencyEntity`  | Dependency on another module        |

### Entity Sources

Entities track their origin through `EntitySource`:

```kotlin
val moduleEntity: ModuleEntity = ...
val source = moduleEntity.entitySource

// Check if from project file
if (source is ProjectFileEntitySource) {
    // Entity defined in project file
}
```

## Reading Project Structure

### Getting Modules

```kotlin
val workspaceModel = WorkspaceModel.getInstance(project)
val modules = workspaceModel.entities(ModuleEntity::class.java).toList()

for (module in modules) {
    println("Module: ${module.name}")
}
```

### Getting Content Roots

```kotlin
val module: ModuleEntity = ...
val contentRoots = module.contentRoots

for (root in contentRoots) {
    println("Content root: ${root.url}")
}
```

### Getting Source Roots

```kotlin
val contentRoot: ContentRootEntity = ...
val sourceRoots = contentRoot.sourceRoots

for (sourceRoot in sourceRoots) {
    println("Source root: ${sourceRoot.url}, type: ${sourceRoot.rootType}")
}
```

### Getting Dependencies

```kotlin
val module: ModuleEntity = ...
val dependencies = module.dependencies

for (dep in dependencies) {
    when (dep) {
        is ModuleDependencyEntity -> println("Module: ${dep.module.name}")
        is LibraryDependencyEntity -> println("Library: ${dep.library.name}")
    }
}
```

## Modifying Project Structure

### Using Write Access

Modifications require write access:

```kotlin
WriteAction.run<Exception> {
    // Workspace model modifications
}
```

### Creating a Module

```kotlin
WriteAction.run<Exception> {
    val workspaceModel = WorkspaceModel.getInstance(project)

    workspaceModel.updateProjectModel("Create module") { updater ->
        val moduleEntity = updater.addEntity(
            ModuleEntity("my-module", emptyList(), project.defaultEntitySource) {
                this.dependencies = emptyList()
            }
        )
    }
}
```

### Adding a Content Root

```kotlin
WriteAction.run<Exception> {
    val workspaceModel = WorkspaceModel.getInstance(project)
    val moduleEntity = workspaceModel.entities(ModuleEntity::class.java).first()

    workspaceModel.updateProjectModel("Add content root") { updater ->
        updater.addEntity(
            ContentRootEntity(
                "file:///path/to/root",
                emptyList(),
                project.defaultEntitySource
            ) {
                this.module = moduleEntity
            }
        )
    }
}
```

### Adding a Source Root

```kotlin
WriteAction.run<Exception> {
    val workspaceModel = WorkspaceModel.getInstance(project)
    val contentRoot = workspaceModel.entities(ContentRootEntity::class.java).first()

    workspaceModel.updateProjectModel("Add source root") { updater ->
        updater.addEntity(
            SourceRootEntity(
                "file:///path/to/src",
                "java-source",
                project.defaultEntitySource
            ) {
                this.contentRoot = contentRoot
            }
        )
    }
}
```

### Adding a Dependency

```kotlin
WriteAction.run<Exception> {
    val workspaceModel = WorkspaceModel.getInstance(project)
    val module = workspaceModel.entities(ModuleEntity::class.java).first()
    val library = workspaceModel.entities(LibraryEntity::class.java).first()

    workspaceModel.updateProjectModel("Add dependency") { updater ->
        updater.addEntity(
            LibraryDependencyEntity(project.defaultEntitySource) {
                this.module = module
                this.library = library
            }
        )
    }
}
```

### Modifying Entities

```kotlin
WriteAction.run<Exception> {
    val workspaceModel = WorkspaceModel.getInstance(project)
    val module = workspaceModel.entities(ModuleEntity::class.java).first()

    workspaceModel.updateProjectModel("Rename module") { updater ->
        updater.modifyEntity(module) {
            this.name = "new-name"
        }
    }
}
```

### Removing Entities

```kotlin
WriteAction.run<Exception> {
    val workspaceModel = WorkspaceModel.getInstance(project)
    val module = workspaceModel.entities(ModuleEntity::class.java).first()

    workspaceModel.updateProjectModel("Remove module") { updater ->
        updater.removeEntity(module)
    }
}
```

## Transactions

### Transaction Basics

All modifications happen within transactions:

```kotlin
workspaceModel.updateProjectModel("Description") { updater ->
    // All changes in this block are atomic
    updater.addEntity(...)
    updater.modifyEntity(...)
    updater.removeEntity(...)
}
```

### Nested Transactions

Transactions can be nested:

```kotlin
workspaceModel.updateProjectModel("Outer") { updater ->
    // Outer transaction

    workspaceModel.updateProjectModel("Inner") { innerUpdater ->
        // Inner transaction
    }
}
```

### Transaction Events

Listen for transaction completion:

```kotlin
workspaceModel.changes.subscribe(parentDisposable, object : ProjectModelChangeListener {
    override fun onChanged(event: ProjectModelEvent) {
        // Handle changes
    }
})
```

## Custom Entities

### Defining Custom Entities

Extend the workspace model with custom entities:

```kotlin
interface MyCustomEntity : WorkspaceEntity {
    var customProperty: String

    companion object {
        val customProperty: EntityProperty<MyCustomEntity, String> =
            EntityProperty("customProperty")
    }
}
```

### Registering Custom Entities

```xml
<extensions defaultExtensionNs="com.intellij">
    <workspaceModel.extension
        implementation="com.example.MyCustomEntityExtension"/>
</extensions>
```

### Using Custom Entities

```kotlin
// Add custom entity
workspaceModel.updateProjectModel("Add custom") { updater ->
    updater.addEntity(
        MyCustomEntityImpl {
            this.customProperty = "value"
        }
    )
}

// Query custom entities
val customEntities = workspaceModel.entities(MyCustomEntity::class.java).toList()
```

## Integration with Legacy APIs

### Converting Between Models

The Workspace Model integrates with legacy APIs:

```kotlin
// From legacy Module to ModuleEntity
val module: Module = ...
val moduleEntity = ModuleEntity.find(module)

// From ModuleEntity to legacy Module
val moduleEntity: ModuleEntity = ...
val module = moduleEntity.module
```

### Bridge APIs

Use bridge APIs for compatibility:

```kotlin
// Get content roots using bridge
val contentRoots = ModuleContentRootsManager.getInstance(module).contentRoots
```

## Event Handling

### Subscribing to Changes

```kotlin
val connection = project.messageBus.connect()
connection.subscribe(WorkspaceModelChangeListener.TOPIC, object : WorkspaceModelChangeListener {
    override fun changed(event: VersionedStorageChange) {
        // Handle changes
        val addedModules = event.getChanges(ModuleEntity::class.java)
            .filterIsInstance<ModuleEntityChange.Added>()

        for (addition in addedModules) {
            println("Module added: ${addition.newEntity.name}")
        }
    }
})
```

### Change Types

| Change Type | Description         |
| ----------- | ------------------- |
| `Added`     | Entity was added    |
| `Removed`   | Entity was removed  |
| `Modified`  | Entity was modified |

### Filtering Changes

```kotlin
event.getChanges(ModuleEntity::class.java)
    .filterIsInstance<ModuleEntityChange.Modified>()
    .forEach { change ->
        println("Module ${change.newEntity.name} changed")
    }
```

## Best Practices

### Use Transactions

Always use transactions for modifications:

```kotlin
// WRONG: Direct modification
module.name = "new-name"  // Not allowed

// CORRECT: Use transaction
workspaceModel.updateProjectModel("Rename") { updater ->
    updater.modifyEntity(module) {
        this.name = "new-name"
    }
}
```

### Batch Operations

Batch related operations in a single transaction:

```kotlin
workspaceModel.updateProjectModel("Setup module") { updater ->
    val module = updater.addEntity(ModuleEntity(...))
    val contentRoot = updater.addEntity(ContentRootEntity(...) { this.module = module })
    updater.addEntity(SourceRootEntity(...) { this.contentRoot = contentRoot })
}
```

### Handle Entity Absence

Entities may not exist:

```kotlin
val module = workspaceModel.entities(ModuleEntity::class.java)
    .find { it.name == "my-module" }

if (module != null) {
    // Work with module
}
```

## Related Topics

- [Project Model](project-model.md) - Legacy project model APIs
- [Virtual File System](virtual-file-system.md) - File access abstraction
- [Extension Points](../01-core-concepts/extension-points.md) - Extending the platform

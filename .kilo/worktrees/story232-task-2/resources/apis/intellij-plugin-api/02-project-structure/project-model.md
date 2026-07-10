# Project Model

The IntelliJ Platform organizes code using a hierarchical project model. Understanding this model is essential for plugins that work with projects, modules, or their contents.

## Project Hierarchy

### Project

A project is the top-level container in the IntelliJ Platform. It represents a complete development context including:

- Source files and resources
- Configuration settings
- Build configurations
- Modules

Projects are represented by the `Project` interface:

```kotlin
interface Project {
    val name: String
    val basePath: String?
    val projectFile: VirtualFile?
    val workspaceFile: VirtualFile?

    fun getService(serviceClass: Class<T>): T
    fun dispose()
}
```

### Module

A module is a logical grouping of source files and dependencies within a project. Modules enable:

- Separation of concerns within a project
- Different dependency sets for different parts
- Independent compilation units

```kotlin
interface Module {
    val name: String
    val project: Project
    val moduleFile: VirtualFile?

    fun getService(serviceClass: Class<T>): T
}
```

### Content Root

A content root is a directory containing files that belong to a module. Each module can have multiple content roots. Content roots are the base directories for source files, resources, and excluded directories.

### Source Root

Source roots are directories within content roots marked for specific purposes:

| Type          | Description                |
| ------------- | -------------------------- |
| Source        | Production source code     |
| Test Source   | Test source code           |
| Resource      | Production resources       |
| Test Resource | Test resources             |
| Excluded      | Not indexed or processed   |
| Generated     | Auto-generated source code |

## Accessing Project Components

### Getting the Current Project

From an action:

```kotlin
class MyAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        // Work with project
    }
}
```

From a service:

```kotlin
class MyProjectService(val project: Project) {
    fun doSomething() {
        // project is available as a property
    }
}
```

From a data context:

```kotlin
val project = e.getData(CommonDataKeys.PROJECT)
```

### Getting Modules

```kotlin
// All modules in a project
val modules = ModuleManager.getInstance(project).modules

// Find module by name
val module = ModuleManager.getInstance(project).findModuleByName("my-module")

// Module for a file
val file = ... // VirtualFile
val module = ModuleUtilCore.findModuleForFile(file, project)
```

### Getting Content Roots

```kotlin
val module: Module = ...

// Get content roots
val contentRoots = ModuleRootManager.getInstance(module).contentRoots

// Get source roots
val sourceRoots = ModuleRootManager.getInstance(module).getSourceRoots(false)

// Get test source roots
val testRoots = ModuleRootManager.getInstance(module).getSourceRoots(true)
```

## Project Types

### Directory-Based Projects

Modern projects use a directory-based structure with configuration in the `.idea` directory:

```
my-project/
├── .idea/
│   ├── modules.xml
│   ├── vcs.xml
│   └── ...
├── src/
└── test/
```

### File-Based Projects

Legacy projects use a single `.ipr` file for configuration. This format is deprecated but still supported.

## Project Lifecycle

### Opening Projects

Projects are opened through the platform:

```kotlin
// Open an existing project
val projectPath = "/path/to/project"
val project = ProjectManager.getInstance().openProject(projectPath)
```

### Project Events

Listen for project lifecycle events:

```kotlin
project.messageBus.connect().subscribe(
    ProjectManager.TOPIC,
    object : ProjectManagerListener {
        override fun projectOpened(project: Project) {
            // Project opened
        }

        override fun projectClosed(project: Project) {
            // Project closed
        }

        override fun projectClosing(project: Project) {
            // Project about to close
        }
    }
)
```

### Default Project

The platform provides a default project for settings:

```kotlin
val defaultProject = ProjectManager.getInstance().defaultProject
```

The default project is used for:

- Template settings
- Actions without an open project
- Configuration defaults

## Module Operations

### Creating Modules

```kotlin
val modifiableModel = ModuleManager.getInstance(project).getModifiableModel()
val module = modifiableModel.newModule("my-module", "file://path/to/module.iml")
modifiableModel.commit()
```

### Modifying Modules

```kotlin
val model = ModuleRootManager.getInstance(module).getModifiableModel()

// Add content root
val contentEntry = model.addContentEntry(virtualFile)
contentEntry.addSourceFolder(sourcePath, false)

model.commit()
```

### Deleting Modules

```kotlin
val model = ModuleManager.getInstance(project).getModifiableModel()
model.removeModule(module)
model.commit()
```

## Dependencies

### Module Dependencies

Modules can depend on other modules:

```kotlin
val model = ModuleRootManager.getInstance(module).getModifiableModel()

// Add module dependency
model.addModuleDependency(otherModule, DependencyScope.COMPILE, false)

model.commit()
```

### Library Dependencies

```kotlin
val model = ModuleRootManager.getInstance(module).getModifiableModel()

// Add library dependency
val library = LibraryTablesRegistrar.getInstance().libraryTable.getLibraryByName("my-lib")
model.addLibraryEntry(library)

model.commit()
```

### Dependency Scopes

| Scope      | Description                        |
| ---------- | ---------------------------------- |
| `COMPILE`  | Available at compile and runtime   |
| `TEST`     | Available only for tests           |
| `RUNTIME`  | Available only at runtime          |
| `PROVIDED` | Available at compile, not packaged |

## Project Configuration

### Accessing Project Settings

```kotlin
// Get project settings
val settings = project.getService(MyProjectSettings::class.java)

// Persistent settings
val state = settings.state
```

### Modifying Project Settings

```kotlin
val settings = project.getService(MyProjectSettings::class.java)
settings.someOption = "new value"
```

## Working with Project Files

### Project Configuration Directory

```kotlin
val ideaDir = project.projectFile?.parent
// or
val ideaDir = VfsUtil.findFile(File(project.basePath!!, ".idea"), true)
```

### Configuration Files

```kotlin
// Get specific configuration file
val vcsConfig = ideaDir?.findChild("vcs.xml")
```

## Thread Safety

### Read Access

Reading project structure requires read access:

```kotlin
ReadAction.run<Exception> {
    val modules = ModuleManager.getInstance(project).modules
    // Process modules
}
```

### Write Access

Modifying project structure requires write access:

```kotlin
WriteAction.run<Exception> {
    val model = ModuleManager.getInstance(project).getModifiableModel()
    // Modify model
    model.commit()
}
```

## Common Patterns

### Finding Files in a Module

```kotlin
fun findFilesInModule(module: Module, extension: String): List<VirtualFile> {
    val result = mutableListOf<VirtualFile>()

    for (root in ModuleRootManager.getInstance(module).getSourceRoots(false)) {
        VfsUtil.iterateChildrenRecursively(root, null) { file ->
            if (file.extension == extension) {
                result.add(file)
            }
            true
        }
    }

    return result
}
```

### Checking if File Belongs to Project

```kotlin
fun isInProject(file: VirtualFile, project: Project): Boolean {
    return ProjectFileIndex.getInstance(project).isInContent(file)
}
```

### Getting Source Root for File

```kotlin
fun getSourceRoot(file: VirtualFile, project: Project): VirtualFile? {
    return ProjectFileIndex.getInstance(project).getSourceRootForFile(file)
}
```

## Related Topics

- [Virtual File System](virtual-file-system.md) - File access abstraction
- [Workspace Model](workspace-model.md) - Modern project structure API
- [Indexing](indexing.md) - Searching project files

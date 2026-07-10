# File Indexing

The IntelliJ Platform maintains indexes of project files to enable fast searching and analysis. Understanding indexing is essential for plugins that need to search or analyze code efficiently.

## Overview

### What Is Indexing?

Indexing is the process of scanning project files and building data structures that enable fast lookups. The platform indexes:

- File names
- File contents
- Symbol definitions
- Symbol references
- Custom data from plugins

### When Indexing Occurs

Indexing happens:

- When a project is opened
- When files are added or modified
- When the IDE starts
- When explicitly triggered

During indexing, some features may be unavailable or return incomplete results.

## Index Structure

### File-Based Indexes

File-based indexes map file IDs to indexed data. Each file is processed independently:

```kotlin
abstract class FileBasedIndexExtension<K, V> {
    abstract val name: ID<K, V>
    abstract val version: Int
    abstract val inputFilter: DataInputFilter?
    abstract val keyDescriptor: KeyDescriptor<K>
    abstract val valueExternalizer: DataExternalizer<V>
    abstract val indexer: DataIndexer<K, V, FileContent>
}
```

### Index IDs

Each index has a unique ID:

```kotlin
companion object {
    val NAME: ID<String, MyData> = ID.create("my.index.name")
}
```

## Creating a Custom Index

### Define the Index Extension

```kotlin
class MyIndexExtension : FileBasedIndexExtension<String, MyData>() {
    companion object {
        val NAME: ID<String, MyData> = ID.create("my.plugin.myindex")
    }

    override val name: ID<String, MyData> = NAME
    override val version = 1

    override val inputFilter: DataInputFilter =
        DefaultFileTypeSpecificInputFilter(PlainTextFileType.INSTANCE)

    override val keyDescriptor: KeyDescriptor<String> =
        EnumeratorStringDescriptor.INSTANCE

    override val valueExternalizer: DataExternalizer<MyData> =
        object : DataExternalizer<MyData> {
            override fun save(out: DataOutput, value: MyData) {
                out.writeUTF(value.name)
                out.writeInt(value.count)
            }

            override fun read(input: DataInput): MyData {
                return MyData(input.readUTF(), input.readInt())
            }
        }

    override val indexer: DataIndexer<String, MyData, FileContent> =
        DataIndexer { content ->
            val map = mutableMapOf<String, MyData>()
            // Extract data from content
            map["key"] = MyData(content.fileName, content.contentAsText.length)
            map
        }
}
```

### Register the Index

```xml
<extensions defaultExtensionNs="com.intellij">
    <fileBasedIndex implementation="com.example.MyIndexExtension"/>
</extensions>
```

## Querying Indexes

### Basic Query

```kotlin
val values = FileBasedIndex.getInstance()
    .getValues(MyIndexExtension.NAME, "key", GlobalSearchScope.projectScope(project))
```

### Getting All Keys

```kotlin
val keys = FileBasedIndex.getInstance()
    .getAllKeys(MyIndexExtension.NAME, project)
```

### Getting Files by Key

```kotlin
val files = FileBasedIndex.getInstance()
    .getContainingFiles(MyIndexExtension.NAME, "key", GlobalSearchScope.projectScope(project))
```

### Processing Indexed Data

```kotlin
FileBasedIndex.getInstance()
    .getFileData(MyIndexExtension.NAME, virtualFile, project)
    .forEach { (key, value) ->
        // Process each key-value pair
    }
```

## Built-in Indexes

### File Name Index

Find files by name:

```kotlin
val files = FilenameIndex.getVirtualFilesByName(
    project,
    "FileName.java",
    GlobalSearchScope.projectScope(project)
)
```

### File Type Index

Find files by type:

```kotlin
val files = FileTypeIndex.getFiles(
    PlainTextFileType.INSTANCE,
    GlobalSearchScope.projectScope(project)
)
```

### Symbol Index

Find symbol definitions:

```kotlin
val symbols = PsiShortNamesCache.getInstance(project)
    .getClassesByName("MyClass", GlobalSearchScope.projectScope(project))
```

## Search Scopes

### Scope Types

| Scope                                              | Description          |
| -------------------------------------------------- | -------------------- |
| `GlobalSearchScope.projectScope(project)`          | All files in project |
| `GlobalSearchScope.moduleScope(module)`            | Files in a module    |
| `GlobalSearchScope.directoryScope(dir, recursive)` | Files in a directory |
| `GlobalSearchScope.fileScope(file)`                | A single file        |

### Combining Scopes

```kotlin
// Union
val combined = scope1.union(scope2)

// Intersection
val intersection = scope1.intersectWith(scope2)

// Exclusion
val excluded = scope1.intersectWith(scope2.not)
```

### Custom Scopes

```kotlin
val customScope = object : GlobalSearchScope(project) {
    override fun contains(file: VirtualFile): Boolean {
        return file.extension == "java"
    }

    override fun isSearchInModuleContent(module: Module): Boolean = true

    override fun isSearchInLibraries(): Boolean = false
}
```

## Indexing Events

### Dumb Mode

During indexing, the IDE is in "dumb mode" where indexes are unavailable:

```kotlin
if (DumbService.getInstance(project).isDumb) {
    // Indexing in progress, indexes unavailable
}
```

### Dumb Mode Listener

```kotlin
project.messageBus.connect().subscribe(
    DumbService.DUMB_MODE,
    object : DumbService.DumbModeListener {
        override fun enteredDumbMode() {
            // Indexing started
        }

        override fun exitDumbMode() {
            // Indexing finished
        }
    }
)
```

### Running in Dumb Mode

Some operations can run during dumb mode:

```kotlin
DumbService.getInstance(project).runWhenSmart {
    // Runs when indexing completes
}
```

### Running with Index Access

```kotlin
DumbService.getInstance(project).withAlternativeResolveStrategy {
    // Code that needs index access
}
```

## Index Updates

### Rebuilding Indexes

```kotlin
FileBasedIndex.getInstance().requestRebuild(MyIndexExtension.NAME)
```

### Reindexing Files

```kotlin
FileBasedIndex.getInstance().requestReindex(virtualFile)
```

### Version Updates

When index structure changes, increment the version:

```kotlin
override val version = 2  // Increment when structure changes
```

The platform automatically rebuilds indexes when the version changes.

## Performance Guidelines

### Minimize Index Size

Keep indexed data small:

```kotlin
// WRONG: Storing large data
override val indexer = DataIndexer { content ->
    mapOf("key" to content.contentAsText.toString())  // Too large
}

// CORRECT: Store only essential data
override val indexer = DataIndexer { content ->
    mapOf("key" to content.contentAsText.length.toString())  // Compact
}
```

### Use Efficient Keys

Keys should be efficiently comparable:

```kotlin
// Use strings for simple keys
override val keyDescriptor = EnumeratorStringDescriptor.INSTANCE

// Use integers for numeric keys
override val keyDescriptor = IntKeyDescriptor.INSTANCE
```

### Filter Input Files

Only index relevant files:

```kotlin
override val inputFilter = DefaultFileTypeSpecificInputFilter(
    JavaFileType.INSTANCE,
    KotlinFileType.INSTANCE
)
```

### Avoid Expensive Processing

Keep indexing fast:

```kotlin
override val indexer = DataIndexer { content ->
    // WRONG: Expensive parsing
    val ast = parseFullAst(content.contentAsText)

    // CORRECT: Quick extraction
    val imports = extractImportsQuickly(content.contentAsText)
    mapOf("imports" to imports)
}
```

## Common Patterns

### Symbol Index

```kotlin
class SymbolIndexExtension : FileBasedIndexExtension<String, Void>() {
    override val indexer = DataIndexer { content ->
        val symbols = extractSymbols(content.contentAsText)
        symbols.associateWith { null }
    }

    // ... other methods
}

// Usage
fun findSymbolDefinition(name: String, project: Project): VirtualFile? {
    return FileBasedIndex.getInstance()
        .getContainingFiles(SymbolIndexExtension.NAME, name,
            GlobalSearchScope.projectScope(project))
        .firstOrNull()
}
```

### Configuration Index

```kotlin
class ConfigIndexExtension : FileBasedIndexExtension<String, String>() {
    override val inputFilter = DefaultFileTypeSpecificInputFilter(
        FileTypeManager.getInstance().getFileTypeByExtension("json")
    )

    override val indexer = DataIndexer { content ->
        val config = JsonParser.parseString(content.contentAsText.toString())
        config.keys.associateWith { config[it].toString() }
    }

    // ... other methods
}
```

## Related Topics

- [Virtual File System](virtual-file-system.md) - File access abstraction
- [PSI Files](../03-psi/psi-files.md) - PSI representation of files
- [Performance Tips](../11-reference/performance-tips.md) - General performance guidelines

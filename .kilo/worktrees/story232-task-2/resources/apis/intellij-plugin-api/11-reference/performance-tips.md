# Performance Tips

This document covers performance optimization strategies for IntelliJ plugins.

## Overview

### Why Performance Matters

Poor performance causes:

- IDE slowdowns
- User frustration
- Plugin rejection

### Performance Areas

| Area              | Impact              |
| ----------------- | ------------------- |
| Startup           | Plugin loading time |
| PSI operations    | Code analysis speed |
| UI responsiveness | User experience     |
| Memory usage      | Stability           |

## Startup Performance

### Defer Initialization

```kotlin
// BAD: Initialize immediately
class MyService {
    private val data = loadData() // Slow!
}

// GOOD: Lazy initialization
class MyService {
    private val data by lazy { loadData() }
}
```

### Use StartupActivity

```kotlin
class MyStartupActivity : StartupActivity {
    override fun runActivity(project: Project) {
        // Run after project opened
        // Keep it fast!
    }
}

// For background initialization
class MyBackgroundStartupActivity : StartupActivity.Background {
    override fun runActivity(project: Project) {
        // Run in background
    }
}
```

### Post-Startup Initialization

```kotlin
class MyService(val project: Project) {
    init {
        // Schedule after startup
        ApplicationManager.getApplication().invokeLater {
            if (!project.isDisposed) {
                initialize()
            }
        }
    }
}
```

## PSI Performance

### Use Cached Values

```kotlin
class MyPsiCache {
    private val cache = CachedValuesManager.getManager(project)

    fun getValue(element: PsiElement): MyData {
        return cache.getCachedValue(element) {
            val data = computeData(element)
            CachedValueProvider.Result.create(
                data,
                element, // Dependency
                ModificationTracker.EVER_CHANGED
            )
        }
    }
}
```

### Avoid Deep Recursion

```kotlin
// BAD: Recursive visitor
fun processElements(element: PsiElement) {
    for (child in element.children) {
        processElements(child) // Stack overflow risk
    }
}

// GOOD: Iterative visitor
fun processElements(element: PsiElement) {
    element.accept(object : PsiRecursiveElementVisitor() {
        override fun visitElement(element: PsiElement) {
            // Process element
            super.visitElement(element)
        }
    })
}
```

### Use Indexes

```kotlin
// BAD: Iterate all files
fun findClass(name: String): PsiClass? {
    return FilenameIndex.getAllFiles(project)
        .mapNotNull { PsiManager.getInstance(project).findFile(it) }
        .flatMap { it.children.toList() }
        .filterIsInstance<PsiClass>()
        .firstOrNull { it.name == name }
}

// GOOD: Use index
fun findClass(name: String): PsiClass? {
    return JavaPsiFacade.getInstance(project)
        .findClass(name, GlobalSearchScope.projectScope(project))
}
```

## UI Performance

### Run on Correct Thread

```kotlin
// BAD: Block EDT
fun loadData() {
    val data = fetchFromNetwork() // Blocks UI!
    updateUI(data)
}

// GOOD: Background task
fun loadData() {
    ProgressManager.getInstance().run(object : Task.Backgroundable(project, "Loading") {
        private var data: List<Item>? = null

        override fun run(indicator: ProgressIndicator) {
            data = fetchFromNetwork()
        }

        override fun onSuccess() {
            updateUI(data!!)
        }
    })
}
```

### Batch UI Updates

```kotlin
// BAD: Update per item
fun updateList(items: List<Item>) {
    for (item in items) {
        model.addElement(item) // Multiple updates
    }
}

// GOOD: Batch update
fun updateList(items: List<Item>) {
    model.removeAllElements()
    model.addAll(items) // Single update
}
```

### Use Smart Update

```kotlin
class MyTreeModel : TreeModel {
    override fun valueForPathChanged(path: TreePath, newValue: Any?) {
        // Only update changed node
        val node = path.lastPathComponent
        if (node != newValue) {
            node.value = newValue
            fireTreeNodesChanged(path)
        }
    }
}
```

## Memory Management

### Avoid Memory Leaks

```kotlin
// BAD: Static reference
object MyCache {
    val data = mutableMapOf<Project, Any>() // Leaks projects!
}

// GOOD: Use project service
class MyCache(val project: Project) {
    private val data = mutableMapOf<String, Any>()

    companion object {
        fun getInstance(project: Project): MyCache {
            return project.getService(MyCache::class.java)
        }
    }
}
```

### Use Disposables

```kotlin
class MyComponent(val project: Project) : Disposable {
    private val listeners = mutableListOf<Listener>()

    fun addListener(listener: Listener) {
        listeners.add(listener)
    }

    override fun dispose() {
        listeners.clear()
    }
}
```

### Release Resources

```kotlin
class MyService : Disposable {
    private var connection: MessageBusConnection? = null

    fun connect() {
        connection = project.messageBus.connect()
    }

    override fun dispose() {
        connection?.disconnect()
        connection = null
    }
}
```

## Indexing Performance

### Efficient Indexing

```kotlin
class MyIndex : FileBasedIndexExtension<String, String>() {
    override fun getIndexer(): DataIndexer<String, String, FileContent> {
        return DataIndexer { content ->
            val result = mutableMapOf<String, String>()

            // Only process relevant files
            if (content.fileType == MyFileType) {
                // Extract data efficiently
                result["key"] = extractValue(content)
            }

            result
        }
    }
}
```

### Use Index Efficiently

```kotlin
// BAD: Process all files
fun findFiles(): List<VirtualFile> {
    return FilenameIndex.getAllFiles(project)
        .filter { it.extension == "my" }
}

// GOOD: Use index
fun findFiles(): List<VirtualFile> {
    return FileBasedIndex.getInstance()
        .getContainingFiles(MyIndex.ID, "key", GlobalSearchScope.projectScope(project))
        .toList()
}
```

## Completion Performance

### Limit Results

```kotlin
class MyCompletionContributor : CompletionContributor() {
    override fun fillCompletionVariants(
        parameters: CompletionParameters,
        result: CompletionResultSet
    ) {
        // Limit results
        val items = computeItems()
            .take(100) // Limit to 100

        result.addAllElements(items)
    }
}
```

### Use Prefix Matching

```kotlin
override fun fillCompletionVariants(
    parameters: CompletionParameters,
    result: CompletionResultSet
) {
    val prefix = result.prefixMatcher.prefix

    // Only compute matching items
    val items = computeMatchingItems(prefix)

    result.addAllElements(items)
}
```

## Caching Strategies

### Application-Level Cache

```kotlin
class ApplicationCache {
    private val cache = ConcurrentHashMap<String, Any>()

    fun get(key: String, loader: () -> Any): Any {
        return cache.getOrPut(key) { loader() }
    }

    fun clear() {
        cache.clear()
    }

    companion object {
        fun getInstance(): ApplicationCache {
            return ApplicationManager.getApplication()
                .getService(ApplicationCache::class.java)
        }
    }
}
```

### Project-Level Cache

```kotlin
class ProjectCache(val project: Project) {
    private val cache = ConcurrentHashMap<String, Any>()

    fun get(key: String, loader: () -> Any): Any {
        return cache.getOrPut(key) { loader() }
    }

    companion object {
        fun getInstance(project: Project): ProjectCache {
            return project.getService(ProjectCache::class.java)
        }
    }
}
```

## Best Practices

### Performance Guidelines

1. **Profile before optimizing**: Measure actual performance
2. **Use lazy initialization**: Defer expensive operations
3. **Cache wisely**: Cache expensive computations
4. **Run off EDT**: Keep UI responsive
5. **Limit results**: Don't return too much data

### Profiling Tools

1. VisualVM
2. YourKit
3. IntelliJ Profiler
4. JProfiler

## Related Topics

- [Common Patterns](common-patterns.md) - Code patterns
- [Troubleshooting](troubleshooting.md) - Common issues
- [PSI Performance](../03-psi/psi-performance.md) - PSI optimization

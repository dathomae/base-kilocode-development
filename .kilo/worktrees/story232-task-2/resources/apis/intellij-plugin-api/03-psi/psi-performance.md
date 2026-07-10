# PSI Performance

Working with PSI efficiently is critical for plugin performance. This guide covers optimization strategies and common pitfalls.

## Understanding PSI Performance

### Why PSI Can Be Slow

PSI operations can be expensive because:

1. **Parsing**: Files are parsed on demand
2. **Tree Traversal**: Walking large trees is costly
3. **Reference Resolution**: Resolving references may search the entire project
4. **Index Access**: Some operations require index lookups

### Key Principles

- Minimize PSI tree access
- Use indexes instead of tree traversal when possible
- Cache results appropriately
- Avoid repeated resolution

## Caching Strategies

### Local Caching

Cache results within a single operation:

```kotlin
// WRONG: Resolve same reference multiple times
fun processReferences(element: PsiElement) {
    val references = element.references
    for (ref in references) {
        val target = ref.resolve()  // First resolution
        if (target != null) {
            val name = target.text
            val target2 = ref.resolve()  // Unnecessary second resolution
        }
    }
}

// CORRECT: Cache resolved target
fun processReferencesOptimized(element: PsiElement) {
    val references = element.references
    for (ref in references) {
        val target = ref.resolve()  // Single resolution
        if (target != null) {
            val name = target.text
            // Use cached target
        }
    }
}
```

### Service-Level Caching

Use services for longer-lived caches:

```kotlin
class MyCacheService(val project: Project) {
    private val cache = ConcurrentHashMap<String, PsiClass>()

    fun findClass(name: String): PsiClass? {
        return cache.getOrPut(name) {
            JavaPsiFacade.getInstance(project)
                .findClass(name, GlobalSearchScope.projectScope(project))
        }
    }

    fun clearCache() {
        cache.clear()
    }
}
```

### Using CachedValues

Use `CachedValuesManager` for automatic cache invalidation:

```kotlin
fun getMethodCount(psiClass: PsiClass): Int {
    return CachedValuesManager.getManager(psiClass.project).getCachedValue(psiClass) {
        val count = psiClass.methods.size
        CachedValueProvider.Result.create(
            count,
            PsiModificationTracker.MODIFICATION_COUNT  // Invalidate on PSI changes
        )
    }
}
```

## Using Indexes

### Prefer Indexes Over Tree Traversal

```kotlin
// WRONG: Traverse all files to find classes
fun findAllClassesSlow(project: Project): List<PsiClass> {
    val classes = mutableListOf<PsiClass>()
    val scope = GlobalSearchScope.projectScope(project)

    for (file in FilenameIndex.getAllFilesByExt(project, "java", scope)) {
        val psiFile = PsiManager.getInstance(project).findFile(file)
        psiFile?.accept(object : PsiRecursiveElementVisitor() {
            override fun visitElement(element: PsiElement) {
                if (element is PsiClass) classes.add(element)
                super.visitElement(element)
            }
        })
    }

    return classes
}

// CORRECT: Use class index
fun findAllClassesFast(project: Project): Collection<PsiClass> {
    return JavaPsiFacade.getInstance(project)
        .findClasses("*", GlobalSearchScope.projectScope(project))
}
```

### Use Appropriate Indexes

| Task               | Index                       |
| ------------------ | --------------------------- |
| Find files by name | `FilenameIndex`             |
| Find files by type | `FileTypeIndex`             |
| Find classes       | `JavaPsiFacade.findClass()` |
| Find methods       | `PsiShortNamesCache`        |
| Find usages        | `ReferencesSearch`          |

## Efficient Tree Traversal

### Use Specific Visitors

```kotlin
// WRONG: Generic visitor with type checking
psiFile.accept(object : PsiRecursiveElementVisitor() {
    override fun visitElement(element: PsiElement) {
        if (element is PsiMethod) {
            // Process method
        }
        super.visitElement(element)
    }
})

// CORRECT: Use PsiTreeUtil for specific types
val methods = PsiTreeUtil.findChildrenOfType(psiFile, PsiMethod::class.java)
```

### Limit Traversal Depth

```kotlin
// Stop recursion when condition met
psiFile.accept(object : PsiRecursiveElementVisitor() {
    var found = false

    override fun visitElement(element: PsiElement) {
        if (found) return  // Stop processing

        if (element is PsiMethod && element.name == "targetMethod") {
            found = true
            return  // Don't visit children
        }

        super.visitElement(element)
    }
})
```

### Process Only What You Need

```kotlin
// WRONG: Process entire file for one method
fun findMethodByName(psiFile: PsiFile, name: String): PsiMethod? {
    var result: PsiMethod? = null
    psiFile.accept(object : PsiRecursiveElementVisitor() {
        override fun visitElement(element: PsiElement) {
            if (element is PsiMethod && element.name == name) {
                result = element
            }
            super.visitElement(element)
        }
    })
    return result
}

// CORRECT: Use targeted search
fun findMethodByNameFast(psiFile: PsiFile, name: String): PsiMethod? {
    val javaFile = psiFile as? PsiJavaFile ?: return null
    return javaFile.classes
        .flatMap { it.methods.toList() }
        .find { it.name == name }
}
```

## Reference Resolution

### Batch Resolution

```kotlin
// WRONG: Resolve references one by one in a loop
fun processAllReferences(project: Project) {
    val scope = GlobalSearchScope.projectScope(project)
    for (file in FilenameIndex.getAllFilesByExt(project, "java", scope)) {
        val psiFile = PsiManager.getInstance(project).findFile(file) ?: continue
        for (ref in psiFile.references) {
            ref.resolve()  // Each resolution may trigger indexing
        }
    }
}

// CORRECT: Batch process with proper scope
fun processAllReferencesOptimized(project: Project) {
    val scope = GlobalSearchScope.projectScope(project)
    // Use DumbService for batch operations
    DumbService.getInstance(project).runReadActionInSmartMode {
        // All references resolved in one pass
    }
}
```

### Avoid Unnecessary Resolution

```kotlin
// WRONG: Always resolve
fun isReferenceTo(element: PsiElement, targetName: String): Boolean {
    val ref = element.reference ?: return false
    val resolved = ref.resolve() ?: return false
    return resolved.text == targetName
}

// CORRECT: Check text first
fun isReferenceToOptimized(element: PsiElement, targetName: String): Boolean {
    if (element.text != targetName) return false
    val ref = element.reference ?: return false
    return ref.resolve() != null
}
```

## Read/Write Actions

### Minimize Write Action Duration

```kotlin
// WRONG: Long write action
WriteAction.run<Exception> {
    // Prepare data (slow)
    val data = prepareData()

    // Modify PSI
    element.add(newElement)
}

// CORRECT: Prepare outside write action
val data = prepareData()  // Outside write action

WriteAction.run<Exception> {
    // Only PSI modification in write action
    element.add(newElement)
}
```

### Use Smart Mode

```kotlin
// Run only when indexes are ready
DumbService.getInstance(project).runWhenSmart {
    // Safe to use indexes
}

// Run in smart mode, wait if necessary
DumbService.getInstance(project).runReadActionInSmartMode {
    // Indexes available
}
```

## Memory Management

### Avoid Holding PSI References

```kotlin
// WRONG: Store PSI elements in collection
class MyService {
    private val elements = mutableListOf<PsiElement>()

    fun addElement(element: PsiElement) {
        elements.add(element)  // Memory leak!
    }
}

// CORRECT: Use smart pointers
class MyService(val project: Project) {
    private val pointers = mutableListOf<SmartPsiElementPointer<PsiElement>>()

    fun addElement(element: PsiElement) {
        val pointer = SmartPointerManager.getInstance(project)
            .createSmartPsiElementPointer(element)
        pointers.add(pointer)
    }

    fun getElements(): List<PsiElement> {
        return pointers.mapNotNull { it.element }.filter { it.isValid }
    }
}
```

### Clear Caches on Disposal

```kotlin
class MyCacheService(val project: Project) : Disposable {
    private val cache = ConcurrentHashMap<String, Any>()

    override fun dispose() {
        cache.clear()
    }
}
```

## Background Processing

### Use Background Tasks

```kotlin
// WRONG: Process on EDT
fun processLargeFile(psiFile: PsiFile) {
    psiFile.accept(object : PsiRecursiveElementVisitor() {
        // Blocks UI
    })
}

// CORRECT: Process in background
fun processLargeFileInBackground(psiFile: PsiFile) {
    ProgressManager.getInstance().run(
        object : Task.Backgroundable(psiFile.project, "Processing") {
            override fun run(indicator: ProgressIndicator) {
                ReadAction.run<Exception> {
                    psiFile.accept(object : PsiRecursiveElementVisitor() {
                        override fun visitElement(element: PsiElement) {
                            indicator.checkCanceled()
                            // Process element
                            super.visitElement(element)
                        }
                    })
                }
            }
        }
    )
}
```

### Check for Cancellation

```kotlin
psiFile.accept(object : PsiRecursiveElementVisitor() {
    override fun visitElement(element: PsiElement) {
        ProgressManager.checkCanceled()  // Allow user to cancel
        // Process element
        super.visitElement(element)
    }
})
```

## Common Pitfalls

### Pitfall: Repeated File Lookups

```kotlin
// WRONG
for (i in 0..100) {
    val file = PsiManager.getInstance(project).findFile(virtualFile)
    // Process file
}

// CORRECT
val file = PsiManager.getInstance(project).findFile(virtualFile)
for (i in 0..100) {
    // Process file
}
```

### Pitfall: Not Checking Validity

```kotlin
// WRONG
val element = getSomeElement()
Thread.sleep(1000)  // Element may become invalid
element.text  // May throw exception

// CORRECT
val element = getSomeElement()
Thread.sleep(1000)
if (element.isValid) {
    element.text
}
```

### Pitfall: Large Result Sets

```kotlin
// WRONG: Load all results into memory
val allReferences = ReferencesSearch.search(element, scope).findAll()

// CORRECT: Process incrementally
ReferencesSearch.search(element, scope).forEach { ref ->
    // Process one at a time
}
```

## Related Topics

- [Indexing](../02-project-structure/indexing.md) - Using indexes efficiently
- [Modifying PSI](modifying-psi.md) - Safe modifications
- [Performance Tips](../11-reference/performance-tips.md) - General performance guidelines

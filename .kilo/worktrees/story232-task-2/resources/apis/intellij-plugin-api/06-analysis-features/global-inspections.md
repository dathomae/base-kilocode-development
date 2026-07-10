# Global Inspections

Global inspections analyze code across the entire project, detecting issues that span multiple files.

## Overview

### What Are Global Inspections?

Global inspections can:

- Analyze cross-file dependencies
- Detect project-wide patterns
- Find unused code across the project
- Validate consistency between files

### Local vs. Global

| Aspect      | Local Inspection  | Global Inspection |
| ----------- | ----------------- | ----------------- |
| Scope       | Single file       | Entire project    |
| Performance | Fast              | Slower            |
| Use case    | File-level issues | Cross-file issues |

## Creating a Global Inspection

### Basic Global Inspection

```kotlin
class MyGlobalInspection : GlobalInspectionTool() {
    override fun getDisplayName(): String = "My global inspection"

    override fun getGroupDisplayName(): String = "My plugin"

    override fun getShortName(): String = "MyGlobalInspection"

    override fun isEnabledByDefault(): Boolean = true

    override fun runInspection(
        scope: AnalysisScope,
        manager: InspectionManager,
        globalContext: GlobalInspectionContext,
        problemDescriptions: MutableSet<ProblemDescriptor>
    ) {
        scope.accept { element ->
            if (element is PsiFile) {
                analyzeFile(element, manager, globalContext, problemDescriptions)
            }
            true
        }
    }
}
```

### Registering a Global Inspection

```xml
<extensions defaultExtensionNs="com.intellij">
    <globalInspection
        shortName="MyGlobalInspection"
        displayName="My global inspection"
        groupName="My plugin"
        enabledByDefault="true"
        implementationClass="com.example.MyGlobalInspection"/>
</extensions>
```

## Cross-File Analysis

### Finding Unused Declarations

```kotlin
class UnusedDeclarationInspection : GlobalInspectionTool() {
    override fun getDisplayName() = "Unused declaration"

    override fun runInspection(
        scope: AnalysisScope,
        manager: InspectionManager,
        globalContext: GlobalInspectionContext,
        problemDescriptions: MutableSet<ProblemDescriptor>
    ) {
        // Collect all declarations
        val declarations = mutableMapOf<String, MutableList<MyNamedElement>>()

        scope.accept { element ->
            if (element is MyNamedElement && element.isPublic) {
                val name = element.name ?: return@accept true
                declarations.getOrPut(name) { mutableListOf() }.add(element)
            }
            true
        }

        // Check each declaration for usages
        for ((_, elements) in declarations) {
            for (element in elements) {
                if (!isUsed(element, scope)) {
                    problemDescriptions.add(
                        manager.createProblemDescriptor(
                            element.nameIdentifier ?: element,
                            "'${element.name}' is never used",
                            RemoveUnusedFix(element),
                            ProblemHighlightType.LIKE_UNUSED_SYMBOL,
                            false
                        )
                    )
                }
            }
        }
    }

    private fun isUsed(element: MyNamedElement, scope: AnalysisScope): Boolean {
        return ReferencesSearch.search(element, scope.toGlobalSearchScope())
            .findFirst() != null
    }
}
```

### Detecting Duplicate Code

```kotlin
class DuplicateCodeInspection : GlobalInspectionTool() {
    override fun getDisplayName() = "Duplicate code"

    override fun runInspection(
        scope: AnalysisScope,
        manager: InspectionManager,
        globalContext: GlobalInspectionContext,
        problemDescriptions: MutableSet<ProblemDescriptor>
    ) {
        val codeBlocks = mutableMapOf<String, MutableList<PsiElement>>()

        scope.accept { element ->
            if (element is MyBlock && element.textLength > 100) {
                val normalized = normalizeCode(element.text)
                codeBlocks.getOrPut(normalized) { mutableListOf() }.add(element)
            }
            true
        }

        // Report duplicates
        for ((_, duplicates) in codeBlocks) {
            if (duplicates.size > 1) {
                for (duplicate in duplicates) {
                    problemDescriptions.add(
                        manager.createProblemDescriptor(
                            duplicate,
                            "Duplicate code found",
                            ExtractMethodFix(duplicates),
                            ProblemHighlightType.WARNING,
                            false
                        )
                    )
                }
            }
        }
    }

    private fun normalizeCode(text: String): String {
        return text.replace(Regex("\\s+"), " ").trim()
    }
}
```

## Graph-Based Analysis

### Dependency Analysis

```kotlin
class CircularDependencyInspection : GlobalInspectionTool() {
    override fun getDisplayName() = "Circular dependency"

    override fun runInspection(
        scope: AnalysisScope,
        manager: InspectionManager,
        globalContext: GlobalInspectionContext,
        problemDescriptions: MutableSet<ProblemDescriptor>
    ) {
        // Build dependency graph
        val graph = buildDependencyGraph(scope)

        // Find cycles
        val cycles = findCycles(graph)

        // Report cycles
        for (cycle in cycles) {
            for (element in cycle) {
                problemDescriptions.add(
                    manager.createProblemDescriptor(
                        element,
                        "Circular dependency: ${cycle.joinToString(" -> ")}",
                        BreakCycleFix(element),
                        ProblemHighlightType.ERROR,
                        false
                    )
                )
            }
        }
    }

    private fun buildDependencyGraph(scope: AnalysisScope): Map<MyClass, Set<MyClass>> {
        val graph = mutableMapOf<MyClass, MutableSet<MyClass>>()

        scope.accept { element ->
            if (element is MyClass) {
                val dependencies = mutableSetOf<MyClass>()
                element.accept(object : PsiRecursiveElementVisitor() {
                    override fun visitElement(el: PsiElement) {
                        if (el is MyIdentifier) {
                            val resolved = el.reference?.resolve()
                            if (resolved is MyClass && resolved != element) {
                                dependencies.add(resolved)
                            }
                        }
                        super.visitElement(el)
                    }
                })
                graph[element] = dependencies
            }
            true
        }

        return graph
    }

    private fun findCycles(graph: Map<MyClass, Set<MyClass>>): List<List<MyClass>> {
        val cycles = mutableListOf<List<MyClass>>()
        val visited = mutableSetOf<MyClass>()
        val path = mutableListOf<MyClass>()

        fun dfs(node: MyClass) {
            if (node in path) {
                val cycleStart = path.indexOf(node)
                cycles.add(path.subList(cycleStart, path.size).toList() + node)
                return
            }
            if (node in visited) return

            visited.add(node)
            path.add(node)

            for (dependency in graph[node] ?: emptySet()) {
                dfs(dependency)
            }

            path.removeAt(path.size - 1)
        }

        for (node in graph.keys) {
            dfs(node)
        }

        return cycles
    }
}
```

## Inspection Context

### Using GlobalInspectionContext

```kotlin
override fun runInspection(
    scope: AnalysisScope,
    manager: InspectionManager,
    globalContext: GlobalInspectionContext,
    problemDescriptions: MutableSet<ProblemDescriptor>
) {
    // Get project
    val project = globalContext.project

    // Get progress indicator
    val progress = globalContext.progressIndicator

    // Check for cancellation
    if (progress.isCanceled) return

    // Get extension
    val extension = globalContext.getExtension(MyExtensionKey)
}
```

## Problem Descriptors

### Creating Global Problem Descriptors

```kotlin
// Standard problem
manager.createProblemDescriptor(
    element,
    "Problem description",
    MyQuickFix(),
    ProblemHighlightType.WARNING,
    false
)

// Problem with multiple locations
manager.createProblemDescriptor(
    arrayOf(element1, element2),
    "Related problem",
    MyQuickFix(),
    ProblemHighlightType.WARNING,
    false
)
```

## Performance Considerations

### Efficient Global Analysis

```kotlin
override fun runInspection(
    scope: AnalysisScope,
    manager: InspectionManager,
    globalContext: GlobalInspectionContext,
    problemDescriptions: MutableSet<ProblemDescriptor>
) {
    // Use indexed data
    val index = FileBasedIndex.getInstance()

    // Process in batches
    val batchSize = 100
    var processed = 0

    scope.accept { element ->
        // Check for cancellation periodically
        if (processed++ % batchSize == 0) {
            if (globalContext.progressIndicator.isCanceled) {
                return@accept false
            }
        }

        // Process element
        analyzeElement(element, problemDescriptions)
        true
    }
}
```

### Caching Results

```kotlin
class MyCachedGlobalInspection : GlobalInspectionTool() {
    private val cache = ConcurrentHashMap<PsiElement, Boolean>()

    override fun runInspection(
        scope: AnalysisScope,
        manager: InspectionManager,
        globalContext: GlobalInspectionContext,
        problemDescriptions: MutableSet<ProblemDescriptor>
    ) {
        scope.accept { element ->
            val hasProblem = cache.getOrPut(element) {
                analyzeElement(element)
            }

            if (hasProblem) {
                problemDescriptions.add(createProblem(element, manager))
            }
            true
        }
    }
}
```

## Related Topics

- [Inspections](inspections.md) - Local inspections
- [Inspection Profiles](inspection-profiles.md) - Managing settings
- [Find Usages](../05-language-support/find-usages.md) - Usage search

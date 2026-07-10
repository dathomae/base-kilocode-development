# Safe Delete

Safe Delete ensures that elements are only removed when they are no longer used. This prevents accidental deletion of code that is still referenced.

## Overview

### Safe Delete Process

1. User invokes Safe Delete (Alt+Delete)
2. Platform checks for usages
3. If usages exist, user is warned
4. User can choose to delete anyway or cancel
5. Element and related code are removed

### Safe Delete Components

| Component                    | Purpose                                   |
| ---------------------------- | ----------------------------------------- |
| `RefactoringSupportProvider` | Enables safe delete for elements          |
| `SafeDeleteProcessor`        | Handles deletion logic                    |
| `UsageInfo`                  | Represents a usage that would be affected |

## Implementing Safe Delete

### Refactoring Support Provider

```kotlin
class MyRefactoringSupportProvider : RefactoringSupportProvider() {
    override fun isSafeDeleteAvailable(element: PsiElement): Boolean {
        return element is MyClass ||
               element is MyFunction ||
               element is MyField
    }

    override fun findUsagesForSafeDelete(
        element: PsiElement,
        usages: MutableList<UsageInfo>
    ): Boolean {
        if (element !is MyNamedElement) return false

        // Find all references
        val references = ReferencesSearch.search(
            element,
            GlobalSearchScope.projectScope(element.project)
        )

        for (reference in references) {
            usages.add(UsageInfo(reference.element))
        }

        return usages.isNotEmpty()
    }
}
```

### Registering the Provider

```xml
<extensions defaultExtensionNs="com.intellij">
    <lang.refactoringSupport
        language="MyLanguage"
        implementationClass="com.example.MyRefactoringSupportProvider"/>
</extensions>
```

## Custom Safe Delete Handler

### Implementing Safe Delete

```kotlin
class MySafeDeleteHandler : BaseRefactoringHandler() {
    override fun invoke(project: Project, elements: Array<PsiElement>, dataContext: DataContext?) {
        val usages = mutableListOf<UsageInfo>()

        // Check for usages
        for (element in elements) {
            findUsages(element, usages)
        }

        if (usages.isNotEmpty()) {
            // Show usage conflicts
            showConflicts(project, usages, elements)
        } else {
            // Safe to delete
            performDelete(elements)
        }
    }

    private fun findUsages(element: PsiElement, usages: MutableList<UsageInfo>) {
        if (element !is MyNamedElement) return

        ReferencesSearch.search(
            element,
            GlobalSearchScope.projectScope(element.project)
        ).forEach { reference ->
            usages.add(UsageInfo(reference.element))
        }
    }

    private fun showConflicts(project: Project, usages: List<UsageInfo>, elements: Array<PsiElement>) {
        val conflicts = usages.map { usage ->
            val file = usage.file
            val line = file?.let {
                PsiDocumentManager.getInstance(project)
                    .getDocument(it)
                    ?.getLineNumber(usage.navigationOffset) ?: 0
            } ?: 0
            "Used in ${file?.name}:${line + 1}"
        }

        val result = Messages.showYesNoDialog(
            project,
            "Element is used in ${usages.size} places:\n${conflicts.take(5).joinToString("\n")}\n\nDelete anyway?",
            "Safe Delete",
            "Delete Anyway",
            "Cancel",
            Messages.getWarningIcon()
        )

        if (result == Messages.YES) {
            performDelete(elements)
        }
    }

    private fun performDelete(elements: Array<PsiElement>) {
        WriteAction.run<Exception> {
            for (element in elements) {
                element.delete()
            }
        }
    }
}
```

## Related Element Deletion

### Deleting Related Elements

```kotlin
class MySafeDeleteProcessor : SafeDeleteProcessor {
    override fun prepareForDeletion(elements: Array<PsiElement>): List<UsageInfo> {
        val usages = mutableListOf<UsageInfo>()

        for (element in elements) {
            when (element) {
                is MyClass -> {
                    // Check for subclasses
                    val subclasses = ClassInheritorsSearch.search(
                        element,
                        GlobalSearchScope.projectScope(element.project),
                        true
                    ).findAll()

                    for (subclass in subclasses) {
                        usages.add(UsageInfo(subclass, "Subclass ${subclass.name}"))
                    }

                    // Check for instances
                    findClassUsages(element, usages)
                }
                is MyFunction -> {
                    // Check for overrides
                    findOverridingMethods(element).forEach { override ->
                        usages.add(UsageInfo(override, "Override in ${override.containingClass?.name}"))
                    }

                    // Check for calls
                    findMethodCalls(element, usages)
                }
                is MyField -> {
                    // Check for field access
                    findFieldAccesses(element, usages)
                }
            }
        }

        return usages
    }

    private fun findClassUsages(clazz: MyClass, usages: MutableList<UsageInfo>) {
        ReferencesSearch.search(clazz, GlobalSearchScope.projectScope(clazz.project))
            .forEach { ref -> usages.add(UsageInfo(ref.element)) }
    }

    private fun findOverridingMethods(method: MyFunction): List<MyFunction> {
        val containingClass = PsiTreeUtil.getParentOfType(method, MyClass::class.java) ?: return emptyList()

        return ClassInheritorsSearch.search(
            containingClass,
            GlobalSearchScope.projectScope(method.project),
            true
        ).flatMap { inheritor ->
            inheritor.methods.filter { it.name == method.name }
        }.toList()
    }
}
```

## Usage Detection

### Detecting Different Usage Types

```kotlin
class MyUsageDetector {
    fun detectUsages(element: PsiElement): List<UsageInfo> {
        val usages = mutableListOf<UsageInfo>()

        when (element) {
            is MyClass -> detectClassUsages(element, usages)
            is MyFunction -> detectFunctionUsages(element, usages)
            is MyField -> detectFieldUsages(element, usages)
            is MyVariable -> detectVariableUsages(element, usages)
        }

        return usages
    }

    private fun detectClassUsages(clazz: MyClass, usages: MutableList<UsageInfo>) {
        val project = clazz.project
        val scope = GlobalSearchScope.projectScope(project)

        // Type references
        ReferencesSearch.search(clazz, scope).forEach { ref ->
            usages.add(UsageInfo(ref.element, "Type reference"))
        }

        // Inheritance
        ClassInheritorsSearch.search(clazz, scope, true).forEach { inheritor ->
            usages.add(UsageInfo(inheritor, "Inherits from ${clazz.name}"))
        }

        // Instance creation
        findConstructorCalls(clazz).forEach { call ->
            usages.add(UsageInfo(call, "Creates instance"))
        }
    }

    private fun detectFunctionUsages(function: MyFunction, usages: MutableList<UsageInfo>) {
        val project = function.project
        val scope = GlobalSearchScope.projectScope(project)

        // Direct calls
        ReferencesSearch.search(function, scope).forEach { ref ->
            usages.add(UsageInfo(ref.element, "Function call"))
        }

        // Overrides
        findOverridingMethods(function).forEach { override ->
            usages.add(UsageInfo(override, "Overrides"))
        }
    }

    private fun detectFieldUsages(field: MyField, usages: MutableList<UsageInfo>) {
        val project = field.project
        val scope = GlobalSearchScope.projectScope(project)

        ReferencesSearch.search(field, scope).forEach { ref ->
            val context = ref.element.parent
            val usageType = when (context) {
                is MyAssignment -> "Write access"
                else -> "Read access"
            }
            usages.add(UsageInfo(ref.element, usageType))
        }
    }
}
```

## Search in Comments and Strings

### Finding Textual References

```kotlin
fun findTextualReferences(element: MyNamedElement): List<UsageInfo> {
    val usages = mutableListOf<UsageInfo>()
    val name = element.name ?: return usages
    val project = element.project

    // Search in comments
    val commentFiles = FileTypeIndex.getFiles(
        MyFileType.INSTANCE,
        GlobalSearchScope.projectScope(project)
    )

    for (file in commentFiles) {
        val psiFile = PsiManager.getInstance(project).findFile(file) ?: continue

        psiFile.accept(object : PsiRecursiveElementVisitor() {
            override fun visitComment(comment: PsiComment) {
                if (comment.text.contains(name)) {
                    usages.add(UsageInfo(comment, "Referenced in comment"))
                }
            }

            override fun visitElement(element: PsiElement) {
                if (element is PsiLiteralExpression) {
                    val value = element.value as? String
                    if (value?.contains(name) == true) {
                        usages.add(UsageInfo(element, "Referenced in string"))
                    }
                }
                super.visitElement(element)
            }
        })
    }

    return usages
}
```

## Programmatic Safe Delete

### Safe Delete from Code

```kotlin
fun safeDelete(element: PsiElement, project: Project): Boolean {
    val usages = mutableListOf<UsageInfo>()

    // Check for usages
    if (element is MyNamedElement) {
        ReferencesSearch.search(
            element,
            GlobalSearchScope.projectScope(project)
        ).forEach { ref ->
            usages.add(UsageInfo(ref.element))
        }
    }

    if (usages.isNotEmpty()) {
        return false  // Cannot safely delete
    }

    // Perform deletion
    WriteAction.run<Exception> {
        element.delete()
    }

    return true
}
```

## Related Topics

- [Find Usages](find-usages.md) - Finding references
- [Rename Refactoring](rename-refactoring.md) - Rename operations
- [Reference Resolution](reference-resolution.md) - Resolving references

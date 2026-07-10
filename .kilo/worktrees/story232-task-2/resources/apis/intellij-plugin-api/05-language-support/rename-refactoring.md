# Rename Refactoring

Rename refactoring allows users to rename symbols and update all references automatically. This requires implementing rename handlers and validators.

## Overview

### Rename Process

1. User invokes rename (Shift+F6)
2. In-place rename or dialog appears
3. New name is validated
4. All references are updated
5. Changes are applied atomically

### Rename Components

| Component                   | Purpose                 |
| --------------------------- | ----------------------- |
| `RenameHandler`             | Initiates rename action |
| `RenamePsiElementProcessor` | Handles rename logic    |
| `NameValidator`             | Validates new names     |

## Implementing Rename Support

### Basic Rename Processor

```kotlin
class MyRenameProcessor : RenamePsiElementProcessor() {
    override fun canProcessElement(element: PsiElement): Boolean {
        return element is MyNamedElement
    }

    override fun renameElement(
        element: PsiElement,
        newName: String,
        state: RenamePsiElementProcessor.State
    ) {
        if (element !is MyNamedElement) return

        // Find all references
        val references = ReferencesSearch.search(
            element,
            GlobalSearchScope.projectScope(element.project)
        )

        // Rename the element
        element.setName(newName)

        // Update all references
        for (reference in references) {
            val refElement = reference.element
            if (refElement is MyIdentifier) {
                refElement.setName(newName)
            }
        }
    }

    override fun prepareRenaming(
        element: PsiElement,
        newName: String,
        allRenames: MutableMap<PsiElement, String>,
        scope: SearchScope
    ) {
        // Add related elements to rename
        if (element is MyClass) {
            // Rename related file if needed
            val file = element.containingFile
            if (file.name == "${element.name}.myf") {
                allRenames[file] = "$newName.myf"
            }
        }
    }
}
```

### Registering the Processor

```xml
<extensions defaultExtensionNs="com.intellij">
    <renamePsiElementProcessor
        implementation="com.example.MyRenameProcessor"/>
</extensions>
```

## In-Place Rename

### Enabling In-Place Rename

```kotlin
class MyRenameProcessor : RenamePsiElementProcessor() {
    override fun isInplaceRenameSupported(): Boolean = true

    override fun substituteElementToRename(element: PsiElement, editor: Editor?): PsiElement? {
        // Return the element to rename (e.g., name identifier)
        if (element is MyNamedElement) {
            return element.nameIdentifier ?: element
        }
        return element
    }
}
```

### Rename Input Validator

```kotlin
class MyRenameInputValidator : RenameInputValidator {
    override fun isInputValid(newName: String, element: PsiElement, context: Project): Boolean {
        // Validate name format
        if (newName.isEmpty()) return false
        if (newName.first().isDigit()) return false
        if (!newName.all { it.isLetterOrDigit() || it == '_' }) return false

        // Check for reserved words
        if (newName in RESERVED_WORDS) return false

        return true
    }

    override fun getErrorMessage(newName: String, context: Project): String? {
        if (newName.isEmpty()) return "Name cannot be empty"
        if (newName.first().isDigit()) return "Name cannot start with a digit"
        if (!newName.all { it.isLetterOrDigit() || it == '_' }) {
            return "Name can only contain letters, digits, and underscores"
        }
        if (newName in RESERVED_WORDS) return "'$newName' is a reserved word"
        return null
    }

    companion object {
        private val RESERVED_WORDS = setOf(
            "if", "else", "for", "while", "return", "class", "function"
        )
    }
}
```

### Registering the Validator

```xml
<extensions defaultExtensionNs="com.intellij">
    <renameInputValidator
        implementation="com.example.MyRenameInputValidator"/>
</extensions>
```

## Rename Handler

### Custom Rename Handler

```kotlin
class MyRenameHandler : RenameHandler {
    override fun isAvailableOnDataContext(dataContext: DataContext): Boolean {
        val element = dataContext.getData(CommonDataKeys.PSI_ELEMENT)
        return element is MyNamedElement
    }

    override fun isRenaming(dataContext: DataContext): Boolean {
        return isAvailableOnDataContext(dataContext)
    }

    override fun invoke(project: Project, editor: Editor, file: PsiFile, dataContext: DataContext) {
        val element = dataContext.getData(CommonDataKeys.PSI_ELEMENT) as? MyNamedElement
            ?: return

        // Start in-place rename
        val rename = InplaceRefactoring.from(element)
        rename?.performInplaceRename()
    }

    override fun invoke(project: Project, elements: Array<PsiElement>, dataContext: DataContext) {
        for (element in elements) {
            if (element is MyNamedElement) {
                invokeRenameDialog(project, element)
            }
        }
    }

    private fun invokeRenameDialog(project: Project, element: MyNamedElement) {
        val dialog = RenameDialog(project, element, null, null)
        dialog.show()
    }
}
```

### Registering the Handler

```xml
<actions>
    <action id="MyRenameAction" class="com.example.MyRenameAction">
        <keyboard-shortcut keymap="$default" first-keystroke="shift F6"/>
        <add-to-group group-id="RefactoringMenu" anchor="first"/>
    </action>
</actions>
```

## Related Elements

### Renaming Related Elements

```kotlin
override fun prepareRenaming(
    element: PsiElement,
    newName: String,
    allRenames: MutableMap<PsiElement, String>,
    scope: SearchScope
) {
    when (element) {
        is MyClass -> {
            // Rename constructor
            element.constructors.forEach { constructor ->
                allRenames[constructor] = newName
            }

            // Rename file if matches class name
            val file = element.containingFile
            if (file.name == "${element.name}.myf") {
                allRenames[file.virtualFile] = "$newName.myf"
            }
        }
        is MyFunction -> {
            // Rename override methods
            findOverridingMethods(element).forEach { override ->
                allRenames[override] = newName
            }
        }
    }
}

private fun findOverridingMethods(method: MyFunction): List<MyFunction> {
    val containingClass = PsiTreeUtil.getParentOfType(method, MyClass::class.java) ?: return emptyList()

    return ClassInheritorsSearch.search(containingClass, GlobalSearchScope.projectScope(method.project), true)
        .flatMap { inheritor ->
            inheritor.methods.filter { it.name == method.name }
        }
        .toList()
}
```

## Rename Conflict Detection

### Detecting Conflicts

```kotlin
class MyRenameConflictDetector : RenamePsiElementProcessor() {
    override fun findExistingNameConflicts(
        element: PsiElement,
        newName: String,
        conflicts: MutableSet<String>
    ) {
        val scope = getSearchScope(element)

        // Check for existing elements with same name
        when (element) {
            is MyClass -> {
                val existingClass = findClass(newName, element.project, scope)
                if (existingClass != null && existingClass != element) {
                    conflicts.add("Class '$newName' already exists")
                }
            }
            is MyFunction -> {
                val containingClass = PsiTreeUtil.getParentOfType(element, MyClass::class.java)
                val existingMethod = containingClass?.methods?.find { it.name == newName && it != element }
                if (existingMethod != null) {
                    conflicts.add("Method '$newName' already exists in class")
                }
            }
            is MyVariable -> {
                val scope = getLocalScope(element)
                val existingVar = scope.variables.find { it.name == newName && it != element }
                if (existingVar != null) {
                    conflicts.add("Variable '$newName' already exists in scope")
                }
            }
        }
    }
}
```

## Programmatic Rename

### Renaming Programmatically

```kotlin
fun renameElement(element: MyNamedElement, newName: String) {
    WriteAction.run<Exception> {
        RenameProcessor(element.project, element, newName, false, false).run()
    }
}
```

### Rename with Callback

```kotlin
fun renameWithCallback(element: MyNamedElement, newName: String, callback: () -> Unit) {
    val processor = RenameProcessor(element.project, element, newName, false, false)
    processor.setCallback {
        callback()
    }

    WriteAction.run<Exception> {
        processor.run()
    }
}
```

## Undo Support

Rename operations are automatically undoable. The platform maintains a command stack that allows users to undo rename operations.

## Related Topics

- [Find Usages](find-usages.md) - Finding references
- [Reference Resolution](reference-resolution.md) - Resolving references
- [Safe Delete](safe-delete.md) - Safe deletion

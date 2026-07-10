# Modifying PSI

Modifying PSI trees requires careful attention to threading, transactions, and validity. This guide covers safe patterns for modifying code programmatically.

## Overview

### Modification Requirements

PSI modifications require:

1. **Write Action**: All modifications must occur within a write action
2. **Valid Elements**: Elements must be valid when modified
3. **Document Commit**: Documents must be committed before PSI access

### The Modification Process

```kotlin
WriteAction.run<Exception> {
    // 1. Ensure document is committed
    PsiDocumentManager.getInstance(project).commitAllDocuments()

    // 2. Perform PSI modifications
    element.add(newElement)

    // 3. Post-process if needed
    JavaCodeStyleManager.getInstance(project).optimizeImports(file)
}
```

## Creating Elements

### Using PsiElementFactory

```kotlin
val factory = PsiElementFactory.getInstance(project)

// Create a class
val newClass = factory.createClass("MyClass")

// Create a method
val method = factory.createMethod("myMethod", PsiType.VOID)

// Create a field
val field = factory.createField("myField", PsiType.INT)

// Create a parameter
val parameter = factory.createParameter("param", PsiType.INT)

// Create a statement from text
val statement = factory.createStatementFromText("int x = 0;", null)

// Create an expression from text
val expression = factory.createExpressionFromText("1 + 2", null)
```

### Creating from Text

```kotlin
// Create a class from text
val classText = """
    public class MyClass {
        public void myMethod() {
        }
    }
""".trimIndent()
val classFromFile = factory.createClassFromText(classText, null)

// Create a method from text
val methodText = "public void myMethod() { System.out.println(\"Hello\"); }"
val methodFromFile = factory.createMethodFromText(methodText, null)
```

## Adding Elements

### Adding to a Class

```kotlin
WriteAction.run<Exception> {
    val psiClass: PsiClass = ...
    val newMethod = PsiElementFactory.getInstance(project).createMethod("newMethod", PsiType.VOID)

    psiClass.add(newMethod)
}
```

### Adding at Specific Position

```kotlin
WriteAction.run<Exception> {
    val psiClass: PsiClass = ...
    val newMethod = ...

    // Add before existing method
    val anchorMethod = psiClass.methods.first()
    psiClass.addBefore(newMethod, anchorMethod)

    // Add after existing method
    psiClass.addAfter(newMethod, anchorMethod)
}
```

### Adding to a Code Block

```kotlin
WriteAction.run<Exception> {
    val codeBlock: PsiCodeBlock = ...
    val statement = PsiElementFactory.getInstance(project)
        .createStatementFromText("int x = 0;", null)

    codeBlock.add(statement)

    // Add at beginning
    codeBlock.addBefore(statement, codeBlock.firstChild)

    // Add at end (before closing brace)
    codeBlock.addBefore(statement, codeBlock.lastChild)
}
```

## Modifying Elements

### Changing Modifiers

```kotlin
WriteAction.run<Exception> {
    val method: PsiMethod = ...

    // Add modifier
    method.modifierList?.setModifierProperty(PsiModifier.STATIC, true)

    // Remove modifier
    method.modifierList?.setModifierProperty(PsiModifier.PUBLIC, false)
}
```

### Changing Names

```kotlin
WriteAction.run<Exception> {
    val variable: PsiVariable = ...
    variable.setName("newName")

    val method: PsiMethod = ...
    method.setName("newMethodName")

    val psiClass: PsiClass = ...
    psiClass.setName("NewClassName")
}
```

### Changing Types

```kotlin
WriteAction.run<Exception> {
    val variable: PsiVariable = ...
    variable.setType(PsiType.INT)

    val method: PsiMethod = ...
    method.returnType = PsiType.VOID
}
```

### Replacing Elements

```kotlin
WriteAction.run<Exception> {
    val oldElement: PsiElement = ...
    val newElement: PsiElement = ...

    oldElement.replace(newElement)
}
```

## Deleting Elements

### Basic Deletion

```kotlin
WriteAction.run<Exception> {
    element.delete()
}
```

### Conditional Deletion

```kotlin
WriteAction.run<Exception> {
    val method: PsiMethod = ...

    if (method.hasModifierProperty(PsiModifier.PRIVATE)) {
        method.delete()
    }
}
```

### Deleting Multiple Elements

```kotlin
WriteAction.run<Exception> {
    val methodsToDelete = psiClass.methods.filter {
        it.hasModifierProperty(PsiModifier.PRIVATE)
    }

    for (method in methodsToDelete) {
        method.delete()
    }
}
```

## Working with Documents

### Document vs. PSI Modification

You can modify code either through PSI or through the document:

```kotlin
// PSI modification (preferred)
WriteAction.run<Exception> {
    psiClass.add(newMethod)
}

// Document modification (for text-level changes)
WriteAction.run<Exception> {
    val document = PsiDocumentManager.getInstance(project).getDocument(psiFile)
    document?.insertString(0, "// Comment\n")
    PsiDocumentManager.getInstance(project).commitDocument(document!!)
}
```

### Committing Documents

Always commit documents before PSI operations:

```kotlin
WriteAction.run<Exception> {
    // Commit any pending document changes
    PsiDocumentManager.getInstance(project).commitAllDocuments()

    // Now safe to modify PSI
    element.add(newElement)
}
```

## Code Style and Formatting

### Reformatting Code

```kotlin
WriteAction.run<Exception> {
    // Make changes
    psiClass.add(newMethod)

    // Reformat
    CodeStyleManager.getInstance(project).reformat(psiClass)
}
```

### Optimizing Imports

```kotlin
WriteAction.run<Exception> {
    JavaCodeStyleManager.getInstance(project).optimizeImports(psiFile)
}
```

### Shortening Class References

```kotlin
WriteAction.run<Exception> {
    JavaCodeStyleManager.getInstance(project)
        .shortenClassReferences(psiFile)
}
```

## Transactions

### Using Write Actions

```kotlin
// Run with exception handling
WriteAction.run<Exception> {
    // Modifications
}

// Run with result
val result = WriteAction.compute<PsiElement, Exception> {
    element.add(newElement)
}

// Run asynchronously
WriteAction.runLater {
    // Modifications on EDT
}
```

### Command Groups

Group related modifications:

```kotlin
WriteAction.run<Exception> {
    val command = "Add getter and setter"

    ApplicationManager.getApplication().runWriteAction {
        // All changes in this block are undoable as one unit
        psiClass.add(getterMethod)
        psiClass.add(setterMethod)
    }
}
```

## Smart Pointers

### Tracking Elements Across Modifications

```kotlin
// Create pointer before modification
val pointer = SmartPointerManager.getInstance(project)
    .createSmartPsiElementPointer(element)

WriteAction.run<Exception> {
    // Modify PSI
    element.parent.add(newElement)

    // Get updated element
    val updatedElement = pointer.element
    if (updatedElement?.isValid == true) {
        // Work with updated element
    }
}
```

### Range Markers

```kotlin
val document = PsiDocumentManager.getInstance(project).getDocument(psiFile)
val marker = document?.createRangeMarker(element.textRange!!)

WriteAction.run<Exception> {
    // Modify document
    document.insertString(0, "// Comment\n")

    // Marker is updated automatically
    val newOffset = marker?.startOffset
}
```

## Common Patterns

### Adding a Method to a Class

```kotlin
fun addMethodToClass(
    psiClass: PsiClass,
    name: String,
    returnType: PsiType,
    body: String
) {
    WriteAction.run<Exception> {
        PsiDocumentManager.getInstance(psiClass.project).commitAllDocuments()

        val factory = PsiElementFactory.getInstance(psiClass.project)
        val method = factory.createMethod(name, returnType)
        val methodBody = factory.createCodeBlockFromText(body, method)
        method.body?.replace(methodBody)

        psiClass.add(method)

        CodeStyleManager.getInstance(psiClass.project).reformat(method)
    }
}
```

### Renaming a Class and Updating References

```kotlin
fun renameClass(psiClass: PsiClass, newName: String) {
    WriteAction.run<Exception> {
        PsiDocumentManager.getInstance(psiClass.project).commitAllDocuments()

        // Rename the class
        psiClass.setName(newName)

        // References are updated automatically by the refactoring framework
    }
}
```

### Wrapping a Statement in a Block

```kotlin
fun wrapInIfStatement(statement: PsiStatement, condition: String) {
    WriteAction.run<Exception> {
        val factory = PsiElementFactory.getInstance(statement.project)

        val ifStatement = factory.createStatementFromText(
            "if ($condition) { }",
            statement.context
        ) as PsiIfStatement

        val thenBlock = factory.createCodeBlock()
        thenBlock.add(statement.copy())
        ifStatement.thenBlock?.replace(thenBlock)

        statement.replace(ifStatement)
    }
}
```

## Error Handling

### Checking Validity

```kotlin
WriteAction.run<Exception> {
    if (!element.isValid) {
        return@run  // Element no longer valid
    }

    // Safe to modify
    element.add(newElement)
}
```

### Handling Conflicts

```kotlin
WriteAction.run<Exception> {
    try {
        element.add(newElement)
    } catch (e: IncorrectOperationException) {
        // Handle invalid operation
        println("Cannot add element: ${e.message}")
    }
}
```

## Related Topics

- [PSI Elements](psi-elements.md) - Understanding element types
- [PSI Performance](psi-performance.md) - Performance optimization
- [Code Formatting](../04-editor-features/code-formatting.md) - Formatting APIs

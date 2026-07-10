# PSI Elements

PSI elements are the nodes in the PSI tree. Each element represents a syntactic construct in the source code, such as a class, method, statement, or expression.

## Overview

### The PsiElement Interface

All PSI elements implement the `PsiElement` interface:

```kotlin
interface PsiElement {
    val parent: PsiElement?
    val children: Array<PsiElement>
    val containingFile: PsiFile
    val text: String
    val textRange: TextRange?
    val startOffsetInParent: Int

    fun <T : PsiElement> getContextOfType(clazz: Class<T>): T?
    fun accept(visitor: PsiElementVisitor)
    fun <R> accept(visitor: PsiElementVisitor<R>): R
}
```

### Element Hierarchy

PSI elements form a tree structure:

```
PsiFile
  PsiClass
    PsiField
    PsiMethod
      PsiParameterList
        PsiParameter
      PsiCodeBlock
        PsiStatement
          PsiExpression
```

## Common Element Types

### Java Elements

| Element            | Description                           |
| ------------------ | ------------------------------------- |
| `PsiClass`         | Class, interface, enum, or annotation |
| `PsiMethod`        | Method declaration                    |
| `PsiField`         | Field declaration                     |
| `PsiParameter`     | Method parameter                      |
| `PsiLocalVariable` | Local variable                        |
| `PsiExpression`    | Expression (many subtypes)            |
| `PsiStatement`     | Statement (many subtypes)             |
| `PsiComment`       | Comment                               |
| `PsiModifierList`  | Modifiers (public, static, etc.)      |

### Kotlin Elements

| Element        | Description          |
| -------------- | -------------------- |
| `KtClass`      | Class declaration    |
| `KtFunction`   | Function declaration |
| `KtProperty`   | Property declaration |
| `KtParameter`  | Function parameter   |
| `KtExpression` | Expression           |
| `KtStatement`  | Statement            |

### XML Elements

| Element        | Description   |
| -------------- | ------------- |
| `XmlFile`      | XML file      |
| `XmlTag`       | XML tag       |
| `XmlAttribute` | XML attribute |
| `XmlText`      | Text content  |

## Accessing Elements

### From a File

```kotlin
val psiFile: PsiFile = ...

// Get root element
val root = psiFile

// Get all children
val children = psiFile.children
```

### From an Action

```kotlin
class MyAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val element = e.getData(CommonDataKeys.PSI_ELEMENT)
    }
}
```

### From an Editor

```kotlin
val editor: Editor = ...
val offset = editor.caretModel.offset
val psiFile = PsiDocumentManager.getInstance(project).getPsiFile(editor.document)
val element = psiFile?.findElementAt(offset)
```

## Element Properties

### Basic Properties

```kotlin
val text = element.text              // Text content
val textRange = element.textRange    // Range in document
val startOffset = element.textRange?.startOffset
val endOffset = element.textRange?.endOffset
val parent = element.parent          // Parent element
val children = element.children      // Direct children
```

### Navigation Properties

```kotlin
val containingFile = element.containingFile
val prevSibling = element.prevSibling
val nextSibling = element.nextSibling
val firstChild = element.firstChild
val lastChild = element.lastChild
```

### Type Information

```kotlin
// For typed elements (Java)
val type = (element as? PsiTypedElement)?.type

// For variables
val variableType = (element as? PsiVariable)?.type

// For methods
val returnType = (element as? PsiMethod)?.returnType
```

## Finding Elements

### By Type

```kotlin
// Find parent of type
val parentClass = PsiTreeUtil.getParentOfType(element, PsiClass::class.java)

// Find child of type
val childMethod = PsiTreeUtil.findChildOfType(element, PsiMethod::class.java)

// Find all children of type
val allMethods = PsiTreeUtil.findChildrenOfType(psiFile, PsiMethod::class.java)
```

### By Condition

```kotlin
// Find first matching
val method = PsiTreeUtil.findFirstParent(element) { it is PsiMethod }

// Find all matching
val allPublic = PsiTreeUtil.findChildrenOfType(psiFile, PsiMethod::class.java)
    .filter { it.hasModifierProperty(PsiModifier.PUBLIC) }
```

### At Offset

```kotlin
val element = psiFile.findElementAt(offset)

// Get element of specific type at offset
val method = PsiTreeUtil.getParentOfType(
    psiFile.findElementAt(offset),
    PsiMethod::class.java
)
```

## Working with Specific Elements

### Classes

```kotlin
val psiClass: PsiClass = ...

// Basic info
val name = psiClass.name
val qualifiedName = psiClass.qualifiedName
val isInterface = psiClass.isInterface
val isEnum = psiClass.isEnum
val isAnnotationType = psiClass.isAnnotationType

// Members
val methods = psiClass.methods
val fields = psiClass.fields
val innerClasses = psiClass.innerClasses

// Inheritance
val superClass = psiClass.superClass
val interfaces = psiClass.interfaces
```

### Methods

```kotlin
val method: PsiMethod = ...

// Basic info
val name = method.name
val returnType = method.returnType
val parameterList = method.parameterList
val parameters = method.parameterList.parameters

// Modifiers
val isStatic = method.hasModifierProperty(PsiModifier.STATIC)
val isPublic = method.hasModifierProperty(PsiModifier.PUBLIC)

// Body
val body = method.body
```

### Expressions

```kotlin
val expression: PsiExpression = ...

// Type
val type = expression.type

// Specific types
when (expression) {
    is PsiLiteralExpression -> {
        val value = expression.value
    }
    is PsiReferenceExpression -> {
        val reference = expression.reference
        val resolved = reference?.resolve()
    }
    is PsiMethodCallExpression -> {
        val method = expression.resolveMethod()
    }
    is PsiBinaryExpression -> {
        val operator = expression.operationSign
        val left = expression.lOperand
        val right = expression.rOperand
    }
}
```

### Statements

```kotlin
val statement: PsiStatement = ...

when (statement) {
    is PsiIfStatement -> {
        val condition = statement.condition
        val thenBranch = statement.thenBranch
        val elseBranch = statement.elseBranch
    }
    is PsiForStatement -> {
        val initialization = statement.initialization
        val condition = statement.condition
        val update = statement.update
        val body = statement.body
    }
    is PsiReturnStatement -> {
        val returnValue = statement.returnValue
    }
}
```

## Element Visitors

### Using Visitors

```kotlin
psiFile.accept(object : PsiRecursiveElementVisitor() {
    override fun visitElement(element: PsiElement) {
        when (element) {
            is PsiClass -> println("Class: ${element.name}")
            is PsiMethod -> println("Method: ${element.name}")
        }
        super.visitElement(element)
    }
})
```

### Java-Specific Visitors

```kotlin
psiFile.accept(object : JavaRecursiveElementVisitor() {
    override fun visitClass(aClass: PsiClass) {
        println("Visiting class: ${aClass.name}")
        super.visitClass(aClass)
    }

    override fun visitMethod(method: PsiMethod) {
        println("Visiting method: ${method.name}")
        super.visitMethod(method)
    }

    override fun visitField(field: PsiField) {
        println("Visiting field: ${field.name}")
        super.visitField(field)
    }
})
```

### Kotlin-Specific Visitors

```kotlin
ktFile.accept(object : KtTreeVisitor<Void>() {
    override fun visitClass(klass: KtClass, data: Void?): Void? {
        println("Visiting class: ${klass.name}")
        return super.visitClass(klass, data)
    }

    override fun visitFunction(function: KtFunction, data: Void?): Void? {
        println("Visiting function: ${function.name}")
        return super.visitFunction(function, data)
    }
})
```

## Element Modification

### Creating Elements

```kotlin
val factory = PsiElementFactory.getInstance(project)

// Create a class
val newClass = factory.createClass("NewClass")

// Create a method
val newMethod = factory.createMethod("newMethod", PsiType.VOID)

// Create a field
val newField = factory.createField("newField", PsiType.INT)

// Create a statement
val statement = factory.createStatementFromText("int x = 0;", null)
```

### Adding Elements

```kotlin
WriteAction.run<Exception> {
    val added = psiClass.add(newMethod)
}
```

### Modifying Elements

```kotlin
WriteAction.run<Exception> {
    val modifierList = method.modifierList
    modifierList.setModifierProperty(PsiModifier.STATIC, true)
}
```

### Deleting Elements

```kotlin
WriteAction.run<Exception> {
    element.delete()
}
```

## Related Topics

- [What Is PSI](what-is-psi.md) - PSI overview
- [Navigating PSI](navigating-psi.md) - Traversal patterns
- [Modifying PSI](modifying-psi.md) - Safe modifications
- [PSI References](psi-references.md) - Reference resolution

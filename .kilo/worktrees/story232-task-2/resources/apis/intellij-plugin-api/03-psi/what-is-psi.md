# What Is PSI?

The Program Structure Interface (PSI) is the IntelliJ Platform's representation of source code structure. It provides a parsed, semantic view of code that enables analysis, navigation, and modification.

## Overview

### PSI as an Abstract Syntax Tree

PSI represents source code as a tree of elements. Each element corresponds to a syntactic construct in the source language:

```
Source Code:
class MyClass {
    void myMethod() {
        System.out.println("Hello");
    }
}

PSI Tree:
PsiClass: MyClass
  PsiMethod: myMethod
    PsiCodeBlock
      PsiExpressionStatement
        PsiMethodCallExpression
          PsiReferenceExpression: System.out.println
          PsiLiteralExpression: "Hello"
```

### Language Support

PSI is language-agnostic at its core. Each language plugin provides:

- **Parser**: Converts text to PSI tree
- **Element Types**: Language-specific PSI interfaces
- **Reference Resolver**: Connects references to definitions

The platform provides PSI implementations for:

- Java
- Kotlin
- XML
- Properties
- JSON
- And many more through plugins

## PSI vs. Text

### Why Use PSI?

Working with PSI instead of raw text provides:

| Aspect       | Text                | PSI              |
| ------------ | ------------------- | ---------------- |
| Parsing      | Manual              | Automatic        |
| Structure    | Implicit            | Explicit tree    |
| Navigation   | By position         | By semantic role |
| Modification | String manipulation | Tree operations  |
| Validation   | Manual              | Built-in         |

### Example: Finding a Method

Text-based approach:

```kotlin
// Fragile, depends on exact formatting
val text = file.text
val methodStart = text.indexOf("void myMethod(")
val methodEnd = text.indexOf("}", methodStart)
```

PSI-based approach:

```kotlin
// Robust, works regardless of formatting
val psiClass = PsiTreeUtil.findChildOfType(psiFile, PsiClass::class.java)
val method = psiClass?.findMethodsByName("myMethod", false)?.firstOrNull()
```

## PSI Architecture

### Core Interfaces

All PSI elements implement `PsiElement`:

```kotlin
interface PsiElement {
    val parent: PsiElement?
    val children: Array<PsiElement>
    val containingFile: PsiFile
    val text: String
    val textRange: TextRange?

    fun <T : PsiElement> getContextOfType(clazz: Class<T>): T?
    fun accept(visitor: PsiElementVisitor)
}
```

### Language-Specific Interfaces

Each language extends the core interfaces:

```kotlin
// Java
interface PsiClass : PsiClassOwner, PsiModifierListOwner {
    val name: String
    val methods: Array<PsiMethod>
    val fields: Array<PsiField>
    val innerClasses: Array<PsiClass>
}

// Kotlin
interface KtClass : KtClassOrObject {
    val name: String?
    val declarations: List<KtDeclaration>
    val primaryConstructor: KtPrimaryConstructor?
}
```

### PSI Files

A `PsiFile` represents a complete source file:

```kotlin
interface PsiFile : PsiElement {
    val name: String
    val virtualFile: VirtualFile?
    val fileType: FileType
    val viewProvider: FileViewProvider
}
```

## PSI Creation

### From a File

```kotlin
val virtualFile: VirtualFile = ...
val psiFile = PsiManager.getInstance(project).findFile(virtualFile)
```

### From an Editor

```kotlin
val editor: Editor = ...
val psiFile = PsiDocumentManager.getInstance(project).getPsiFile(editor.document)
```

### From a Data Context

```kotlin
val psiFile = e.getData(CommonDataKeys.PSI_FILE)
val psiElement = e.getData(CommonDataKeys.PSI_ELEMENT)
```

## PSI Validity

### Validity Checks

PSI elements can become invalid when:

- The underlying file is deleted
- The file is modified externally
- A write action modifies the PSI

Always check validity:

```kotlin
if (psiElement.isValid) {
    // Safe to use
}
```

### Smart Pointers

Use smart pointers to track PSI elements across modifications:

```kotlin
val pointer = SmartPointerManager.getInstance(project)
    .createSmartPsiElementPointer(psiElement)

// Later, after potential modifications
val element = pointer.element
if (element != null && element.isValid) {
    // Element still exists and is valid
}
```

## PSI and Documents

### Document Relationship

PSI and documents are synchronized:

```kotlin
// PSI to Document
val document = PsiDocumentManager.getInstance(project).getDocument(psiFile)

// Document to PSI
val psiFile = PsiDocumentManager.getInstance(project).getPsiFile(document)
```

### Committing Documents

Before accessing PSI after document changes:

```kotlin
val document: Document = ...
PsiDocumentManager.getInstance(project).commitDocument(document)
// Now PSI is synchronized with document
```

## Thread Safety

### Read Access

Reading PSI requires read access:

```kotlin
ReadAction.compute<PsiElement?, Exception> {
    PsiManager.getInstance(project).findFile(virtualFile)
}
```

### Write Access

Modifying PSI requires write access:

```kotlin
WriteAction.run<Exception> {
    psiElement.delete()
}
```

### Document vs. PSI Modifications

- Document modifications can happen on any thread (with write action)
- PSI modifications must commit documents first

```kotlin
WriteAction.run<Exception> {
    // Modify document
    document.insertString(0, "// comment\n")

    // Commit to sync PSI
    PsiDocumentManager.getInstance(project).commitDocument(document)

    // Now PSI reflects changes
}
```

## Common Use Cases

### Finding Elements

```kotlin
// Find class by name
val psiClass = JavaPsiFacade.getInstance(project)
    .findClass("com.example.MyClass", GlobalSearchScope.projectScope(project))

// Find element at offset
val element = psiFile.findElementAt(offset)

// Find parent of type
val parentClass = PsiTreeUtil.getParentOfType(element, PsiClass::class.java)
```

### Analyzing Code

```kotlin
// Get all methods in a class
val methods = psiClass.methods

// Check if method is static
val isStatic = method.hasModifierProperty(PsiModifier.STATIC)

// Get method signature
val signature = method.getSignature(PsiSubstitutor.EMPTY)
```

### Resolving References

```kotlin
// Resolve a reference
val reference = element.reference
val resolved = reference?.resolve()

// Get resolved element info
if (resolved is PsiClass) {
    println("Resolved to class: ${resolved.name}")
}
```

## Related Topics

- [PSI Files](psi-files.md) - Working with file-level PSI
- [PSI Elements](psi-elements.md) - Understanding element types
- [Navigating PSI](navigating-psi.md) - Traversal patterns
- [Modifying PSI](modifying-psi.md) - Safe modifications

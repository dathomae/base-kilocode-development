# PSI Files

A PSI file represents the parsed structure of a source file. It is the root of the PSI tree for a single file and provides access to all elements within that file.

## Overview

### What Is a PsiFile?

`PsiFile` is the interface representing a complete source file. It extends `PsiElement` and serves as the root of the PSI tree:

```kotlin
interface PsiFile : PsiElement {
    val name: String
    val virtualFile: VirtualFile?
    val fileType: FileType
    val viewProvider: FileViewProvider
    val language: Language
}
```

### Language-Specific File Types

Each language provides specialized file types:

| Language   | File Type        |
| ---------- | ---------------- |
| Java       | `PsiJavaFile`    |
| Kotlin     | `KtFile`         |
| XML        | `XmlFile`        |
| Properties | `PropertiesFile` |
| JSON       | `JsonFile`       |

## Getting a PsiFile

### From a VirtualFile

```kotlin
val virtualFile: VirtualFile = ...
val psiFile = PsiManager.getInstance(project).findFile(virtualFile)
```

### From a Document

```kotlin
val document: Document = ...
val psiFile = PsiDocumentManager.getInstance(project).getPsiFile(document)
```

### From an Action

```kotlin
class MyAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val psiFile = e.getData(CommonDataKeys.PSI_FILE)
    }
}
```

### From an Editor

```kotlin
val editor: Editor = ...
val psiFile = PsiDocumentManager.getInstance(project).getPsiFile(editor.document)
```

## File Properties

### Basic Properties

```kotlin
val name = psiFile.name           // "MyClass.java"
val language = psiFile.language   // Language.JAVA
val fileType = psiFile.fileType   // JavaFileType.INSTANCE
val virtualFile = psiFile.virtualFile  // VirtualFile
```

### Text Content

```kotlin
val text = psiFile.text
val textLength = psiFile.textLength
val textRange = psiFile.textRange
```

### Containing Directory

```kotlin
val directory = psiFile.containingDirectory
val packageName = (psiFile as? PsiJavaFile)?.packageName
```

## Working with Java Files

### Package Information

```kotlin
val javaFile = psiFile as PsiJavaFile
val packageName = javaFile.packageName
val packageStatement = javaFile.packageStatement
```

### Import Statements

```kotlin
val javaFile = psiFile as PsiJavaFile
val imports = javaFile.importList

for (importStatement in imports?.allImportStatements ?: emptyArray()) {
    println(importStatement.importReference?.text)
}
```

### Classes

```kotlin
val javaFile = psiFile as PsiJavaFile
val classes = javaFile.classes

for (psiClass in classes) {
    println("Class: ${psiClass.name}")
    println("Methods: ${psiClass.methods.map { it.name }}")
}
```

## Working with Kotlin Files

### Declarations

```kotlin
val ktFile = psiFile as KtFile
val declarations = ktFile.declarations

for (declaration in declarations) {
    when (declaration) {
        is KtClass -> println("Class: ${declaration.name}")
        is KtFunction -> println("Function: ${declaration.name}")
        is KtProperty -> println("Property: ${declaration.name}")
    }
}
```

### Package and Imports

```kotlin
val ktFile = psiFile as KtFile
val packageFqName = ktFile.packageFqName
val importDirectives = ktFile.importDirectives
```

## Working with XML Files

### Root Tag

```kotlin
val xmlFile = psiFile as XmlFile
val rootTag = xmlFile.rootTag

println("Root tag: ${rootTag?.name}")
```

### Sub-tags

```kotlin
val rootTag = xmlFile.rootTag
val subTags = rootTag?.subTags

for (tag in subTags ?: emptyArray()) {
    println("Tag: ${tag.name}, value: ${tag.value?.text}")
}
```

### Attributes

```kotlin
val tag: XmlTag = ...
val attributes = tag.attributes

for (attr in attributes) {
    println("${attr.name} = ${attr.value}")
}
```

## File Creation

### Creating a New File

```kotlin
val directory: PsiDirectory = ...
val fileName = "NewClass.java"
val content = """
    package ${directory.virtualFile.path};

    public class NewClass {
    }
""".trimIndent()

WriteAction.run<Exception> {
    val newFile = directory.createFile(fileName)
    val document = PsiDocumentManager.getInstance(project).getDocument(newFile)
    document?.setText(content)
    PsiDocumentManager.getInstance(project).commitDocument(document!!)
}
```

### Creating from Template

```kotlin
val template = FileTemplateManager.getInstance(project)
    .getInternalTemplate("JavaClass")

val properties = Properties()
properties.setProperty("PACKAGE_NAME", packageName)
properties.setProperty("CLASS_NAME", className)

val content = template.getText(properties)

WriteAction.run<Exception> {
    val newFile = directory.createFile("$className.java")
    val document = PsiDocumentManager.getInstance(project).getDocument(newFile)
    document?.setText(content)
}
```

## File Modification

### Modifying Content

```kotlin
WriteAction.run<Exception> {
    val document = PsiDocumentManager.getInstance(project).getDocument(psiFile)
    document?.insertString(0, "// Added comment\n")
    PsiDocumentManager.getInstance(project).commitDocument(document!!)
}
```

### Adding Imports

```kotlin
val javaFile = psiFile as PsiJavaFile

WriteAction.run<Exception> {
    JavaCodeStyleManager.getInstance(project)
        .addImport(javaFile, "java.util.List")
}
```

### Optimizing Imports

```kotlin
WriteAction.run<Exception> {
    JavaCodeStyleManager.getInstance(project)
        .optimizeImports(psiFile)
}
```

## File Events

### Listening for Changes

```kotlin
val connection = PsiManager.getInstance(project).addPsiTreeChangeListener(
    object : PsiTreeChangeListener {
        override fun beforeChildAddition(event: PsiTreeChangeEvent) {}
        override fun childAdded(event: PsiTreeChangeEvent) {}
        override fun beforeChildRemoval(event: PsiTreeChangeEvent) {}
        override fun childRemoved(event: PsiTreeChangeEvent) {}
        override fun beforeChildReplacement(event: PsiTreeChangeEvent) {}
        override fun childReplaced(event: PsiTreeChangeEvent) {}
        override fun beforeChildMovement(event: PsiTreeChangeEvent) {}
        override fun childMoved(event: PsiTreeChangeEvent) {}
        override fun beforeChildrenChange(event: PsiTreeChangeEvent) {}
        override fun childrenChanged(event: PsiTreeChangeEvent) {}
        override fun beforePropertyChange(event: PsiTreeChangeEvent) {}
        override fun propertyChanged(event: PsiTreeChangeEvent) {}
    },
    parentDisposable
)
```

## File Validation

### Checking Validity

```kotlin
if (psiFile.isValid) {
    // File exists and is valid
}
```

### Checking if Physical

```kotlin
if (psiFile.isPhysical) {
    // File exists on disk
} else {
    // File is in-memory only (e.g., for preview)
}
```

## Common Patterns

### Finding Element at Offset

```kotlin
val offset = editor.caretModel.offset
val element = psiFile.findElementAt(offset)

// Find parent of specific type
val method = PsiTreeUtil.getParentOfType(element, PsiMethod::class.java)
```

### Getting All Elements of Type

```kotlin
val allMethods = PsiTreeUtil.findChildrenOfType(psiFile, PsiMethod::class.java)
val allClasses = PsiTreeUtil.findChildrenOfType(psiFile, PsiClass::class.java)
```

### Processing File Elements

```kotlin
psiFile.accept(object : PsiRecursiveElementVisitor() {
    override fun visitElement(element: PsiElement) {
        when (element) {
            is PsiClass -> println("Found class: ${element.name}")
            is PsiMethod -> println("Found method: ${element.name}")
        }
        super.visitElement(element)
    }
})
```

## Related Topics

- [What Is PSI](what-is-psi.md) - PSI overview
- [File View Providers](file-view-providers.md) - Multi-language files
- [PSI Elements](psi-elements.md) - Element types
- [Modifying PSI](modifying-psi.md) - Safe modifications

# Code Folding

Code folding allows users to collapse and expand regions of code. Plugins can define custom folding regions for specific languages or file types.

## Overview

### What Is Code Folding?

Code folding provides:

- Collapsible code regions
- Custom placeholder text
- Visual indicators in the gutter
- Keyboard shortcuts for folding/unfolding

### Folding Architecture

The folding system consists of:

- **FoldingBuilder**: Creates folding descriptors
- **FoldingDescriptor**: Describes a foldable region
- **FoldingModel**: Manages folding state in the editor

## Creating a Folding Builder

### Basic Folding Builder

```kotlin
class MyFoldingBuilder : FoldingBuilderEx() {
    override fun buildFoldRegions(root: PsiElement, document: Document, quick: Boolean): Array<FoldingDescriptor> {
        val descriptors = mutableListOf<FoldingDescriptor>()

        root.accept(object : PsiRecursiveElementVisitor() {
            override fun visitElement(element: PsiElement) {
                if (isFoldable(element)) {
                    descriptors.add(FoldingDescriptor(element, element.textRange))
                }
                super.visitElement(element)
            }
        })

        return descriptors.toTypedArray()
    }

    override fun getPlaceholderText(node: ASTNode): String? {
        return when (node.elementType) {
            MyTokenTypes.BLOCK -> "{...}"
            MyTokenTypes.COMMENT -> "/*...*/"
            else -> "..."
        }
    }

    override fun isCollapsedByDefault(node: ASTNode): Boolean {
        return node.elementType == MyTokenTypes.IMPORT_LIST
    }

    private fun isFoldable(element: PsiElement): Boolean {
        return element.node?.elementType in setOf(
            MyTokenTypes.BLOCK,
            MyTokenTypes.COMMENT,
            MyTokenTypes.IMPORT_LIST
        )
    }
}
```

### Registering the Folding Builder

```xml
<extensions defaultExtensionNs="com.intellij">
    <lang.foldingBuilder
        language="MyLanguage"
        implementationClass="com.example.MyFoldingBuilder"/>
</extensions>
```

## Folding Descriptors

### Creating Descriptors

```kotlin
// Basic descriptor
FoldingDescriptor(element, element.textRange)

// With custom placeholder
FoldingDescriptor(element, element.textRange, null, "custom placeholder")

// With collapsed state
FoldingDescriptor(element, element.textRange, null, "...", true)
```

### Descriptor Properties

| Property             | Description                |
| -------------------- | -------------------------- |
| `element`            | The PSI element to fold    |
| `range`              | The text range to collapse |
| `group`              | Group for related foldings |
| `placeholderText`    | Text shown when collapsed  |
| `collapsedByDefault` | Initial collapsed state    |

## Placeholder Text

### Dynamic Placeholders

```kotlin
override fun getPlaceholderText(node: ASTNode): String? {
    return when (node.elementType) {
        MyTokenTypes.CLASS -> {
            val className = node.findChildByType(MyTokenTypes.IDENTIFIER)?.text
            "class $className {...}"
        }
        MyTokenTypes.METHOD -> {
            val methodName = node.findChildByType(MyTokenTypes.IDENTIFIER)?.text
            "$methodName() {...}"
        }
        MyTokenTypes.IMPORT_LIST -> "import ..."
        else -> "..."
    }
}
```

### Placeholder with Element Count

```kotlin
override fun getPlaceholderText(node: ASTNode): String? {
    if (node.elementType == MyTokenTypes.BLOCK) {
        val statementCount = countStatements(node)
        return "{...$statementCount statements...}"
    }
    return "..."
}

private fun countStatements(node: ASTNode): Int {
    var count = 0
    var child = node.firstChildNode
    while (child != null) {
        if (child.elementType == MyTokenTypes.STATEMENT) {
            count++
        }
        child = child.treeNext
    }
    return count
}
```

## Collapsed by Default

### Setting Default State

```kotlin
override fun isCollapsedByDefault(node: ASTNode): Boolean {
    return when (node.elementType) {
        // Always collapse imports
        MyTokenTypes.IMPORT_LIST -> true

        // Collapse long comments
        MyTokenTypes.COMMENT -> node.textLength > 100

        // Collapse generated code
        MyTokenTypes.GENERATED_BLOCK -> true

        // Don't collapse by default
        else -> false
    }
}
```

### User Preferences

```kotlin
override fun isCollapsedByDefault(node: ASTNode): Boolean {
    val settings = CodeFoldingSettings.getInstance()

    return when (node.elementType) {
        MyTokenTypes.IMPORT_LIST -> settings.COLLAPSE_IMPORTS
        MyTokenTypes.COMMENT -> settings.COLLAPSE_DOC_COMMENTS
        MyTokenTypes.METHOD -> settings.COLLAPSE_METHODS
        else -> false
    }
}
```

## Common Patterns

### Folding Methods

```kotlin
class MethodFoldingBuilder : FoldingBuilderEx() {
    override fun buildFoldRegions(root: PsiElement, document: Document, quick: Boolean): Array<FoldingDescriptor> {
        val descriptors = mutableListOf<FoldingDescriptor>()

        root.accept(object : JavaRecursiveElementVisitor() {
            override fun visitMethod(method: PsiMethod) {
                val body = method.body ?: return
                descriptors.add(FoldingDescriptor(method, body.textRange))
                super.visitMethod(method)
            }
        })

        return descriptors.toTypedArray()
    }

    override fun getPlaceholderText(node: ASTNode): String? {
        val method = node.psi as? PsiMethod ?: return "..."
        val params = method.parameterList.parameters
        return "${method.name}(${params.joinToString { it.name }}) {...}"
    }

    override fun isCollapsedByDefault(node: ASTNode): Boolean = false
}
```

### Folding Imports

```kotlin
class ImportFoldingBuilder : FoldingBuilderEx() {
    override fun buildFoldRegions(root: PsiElement, document: Document, quick: Boolean): Array<FoldingDescriptor> {
        val descriptors = mutableListOf<FoldingDescriptor>()

        if (root is PsiJavaFile) {
            val importList = root.importList
            if (importList != null && importList.allImportStatements.size > 1) {
                descriptors.add(FoldingDescriptor(importList, importList.textRange))
            }
        }

        return descriptors.toTypedArray()
    }

    override fun getPlaceholderText(node: ASTNode): String = "import ..."

    override fun isCollapsedByDefault(node: ASTNode): Boolean = true
}
```

### Folding XML Tags

```kotlin
class XmlFoldingBuilder : FoldingBuilderEx() {
    override fun buildFoldRegions(root: PsiElement, document: Document, quick: Boolean): Array<FoldingDescriptor> {
        val descriptors = mutableListOf<FoldingDescriptor>()

        root.accept(object : XmlRecursiveElementVisitor() {
            override fun visitTag(tag: XmlTag) {
                if (tag.value.text.isNotEmpty()) {
                    descriptors.add(FoldingDescriptor(tag, tag.textRange))
                }
                super.visitTag(tag)
            }
        })

        return descriptors.toTypedArray()
    }

    override fun getPlaceholderText(node: ASTNode): String? {
        val tag = node.psi as? XmlTag ?: return "..."
        return "<${tag.name}>...</${tag.name}>"
    }

    override fun isCollapsedByDefault(node: ASTNode): Boolean = false
}
```

### Folding Comments

```kotlin
class CommentFoldingBuilder : FoldingBuilderEx() {
    override fun buildFoldRegions(root: PsiElement, document: Document, quick: Boolean): Array<FoldingDescriptor> {
        val descriptors = mutableListOf<FoldingDescriptor>()

        root.accept(object : PsiRecursiveElementVisitor() {
            override fun visitComment(comment: PsiComment) {
                if (comment.textLength > 50) {
                    descriptors.add(FoldingDescriptor(comment, comment.textRange))
                }
            }
        })

        return descriptors.toTypedArray()
    }

    override fun getPlaceholderText(node: ASTNode): String? {
        val comment = node.psi as? PsiComment ?: return "..."
        val firstLine = comment.text.lines().firstOrNull()?.take(30) ?: "..."
        return "$firstLine..."
    }

    override fun isCollapsedByDefault(node: ASTNode): Boolean = false
}
```

## Programmatic Folding Control

### Getting Folding State

```kotlin
val foldingModel = editor.foldingModel
val allFoldRegions = foldingModel.allFoldRegions

for (region in allFoldRegions) {
    println("Region: ${region.placeholderText}, collapsed: ${region.isExpanded}")
}
```

### Collapsing/Expanding Regions

```kotlin
// Collapse all
editor.foldingModel.runBatchFoldingOperation {
    for (region in editor.foldingModel.allFoldRegions) {
        region.isExpanded = false
    }
}

// Expand all
editor.foldingModel.runBatchFoldingOperation {
    for (region in editor.foldingModel.allFoldRegions) {
        region.isExpanded = true
    }
}
```

### Folding at Offset

```kotlin
fun foldAtOffset(editor: Editor, offset: Int) {
    editor.foldingModel.runBatchFoldingOperation {
        val region = editor.foldingModel.getCollapsedRegionAtOffset(offset)
        region?.isExpanded = false
    }
}
```

## Performance Considerations

### Quick Mode

```kotlin
override fun buildFoldRegions(root: PsiElement, document: Document, quick: Boolean): Array<FoldingDescriptor> {
    if (quick) {
        // Return minimal set for quick display
        return buildQuickFoldRegions(root)
    }
    // Return full set
    return buildFullFoldRegions(root)
}
```

### Caching

```kotlin
class MyFoldingBuilder : FoldingBuilderEx() {
    private val cache = ConcurrentHashMap<PsiElement, Array<FoldingDescriptor>>()

    override fun buildFoldRegions(root: PsiElement, document: Document, quick: Boolean): Array<FoldingDescriptor> {
        return cache.getOrPut(root) {
            buildFoldRegionsInternal(root)
        }
    }
}
```

## Related Topics

- [Editor Basics](editor-basics.md) - Editor architecture
- [Document Model](document-model.md) - Working with documents
- [Syntax Highlighting](../05-language-support/syntax-highlighting.md) - Visual styling

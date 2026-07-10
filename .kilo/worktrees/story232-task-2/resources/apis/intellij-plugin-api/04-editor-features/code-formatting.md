# Code Formatting

Code formatting ensures consistent style across a codebase. Plugins can provide custom formatting rules for specific languages or file types.

## Overview

### Formatting Architecture

The formatting system consists of:

- **FormattingModel**: Describes formatting structure
- **Block**: Represents a formatted region
- **Spacing**: Controls whitespace between blocks
- **Indent**: Controls indentation
- **Wrap**: Controls line wrapping

### Formatting Process

1. Build a tree of blocks representing the code structure
2. Define spacing rules between blocks
3. Define indentation rules
4. Define wrap rules
5. Apply formatting to the document

## Creating a Formatter

### Formatting Model Builder

```kotlin
class MyFormattingModelBuilder : FormattingModelBuilder {
    override fun createModel(formattingContext: FormattingContext): FormattingModel {
        val element = formattingContext.psiElement
        val settings = formattingContext.codeStyleSettings

        val rootBlock = MyRootBlock(element, settings)
        return FormattingModelProvider.createFormattingModelForPsiFile(
            element.containingFile,
            rootBlock,
            settings
        )
    }
}
```

### Registering the Formatter

```xml
<extensions defaultExtensionNs="com.intellij">
    <lang.formatter
        language="MyLanguage"
        implementationClass="com.example.MyFormattingModelBuilder"/>
</extensions>
```

## Defining Blocks

### Basic Block

```kotlin
class MyBlock(
    element: ASTNode,
    wrap: Wrap?,
    alignment: Alignment?,
    private val settings: CodeStyleSettings
) : AbstractBlock(element, wrap, alignment) {

    override fun buildChildren(): List<Block> {
        val children = mutableListOf<Block>()
        var child = myNode.firstChildNode

        while (child != null) {
            if (!isWhitespaceOrComment(child)) {
                children.add(MyBlock(child, null, null, settings))
            }
            child = child.treeNext
        }

        return children
    }

    override fun getSpacing(child1: Block?, child2: Block): Spacing? {
        // Define spacing between blocks
        return Spacing.getReadOnlySpacing()
    }

    override fun getIndent(): Indent? {
        return Indent.getNoneIndent()
    }

    private fun isWhitespaceOrComment(node: ASTNode): Boolean {
        return node.elementType in setOf(TokenType.WHITE_SPACE, MyTokenTypes.COMMENT)
    }
}
```

### Root Block

```kotlin
class MyRootBlock(
    element: ASTNode,
    settings: CodeStyleSettings
) : MyBlock(element, null, null, settings) {

    override fun getIndent(): Indent? = Indent.getNoneIndent()
}
```

## Spacing Rules

### Basic Spacing

```kotlin
override fun getSpacing(child1: Block?, child2: Block): Spacing? {
    return when {
        // Single space between elements
        isBetweenKeywords(child1, child2) -> Spacing.createSpacing(1, 1, 0, false, 0)

        // No space
        isBetweenOperator(child1, child2) -> Spacing.createSpacing(0, 0, 0, false, 0)

        // Line break
        isBetweenStatements(child1, child2) -> Spacing.createSpacing(0, 0, 1, true, 0)

        // Default
        else -> Spacing.getReadOnlySpacing()
    }
}
```

### Spacing Factory Methods

| Method                                                                              | Description               |
| ----------------------------------------------------------------------------------- | ------------------------- |
| `createSpacing(minSpaces, maxSpaces, minLineFeeds, keepLineBreaks, keepBlankLines)` | Custom spacing            |
| `getReadOnlySpacing()`                                                              | Preserve existing spacing |
| `getEmptySpacing()`                                                                 | No spacing                |
| `createSafeSpacing(false, 0)`                                                       | No line breaks            |

### Spacing Examples

```kotlin
// Single space, no line breaks
Spacing.createSpacing(1, 1, 0, false, 0)

// No space
Spacing.createSpacing(0, 0, 0, false, 0)

// At least one line break
Spacing.createSpacing(0, Int.MAX_VALUE, 1, true, 1)

// Preserve existing line breaks
Spacing.createSpacing(0, Int.MAX_VALUE, 0, true, 0)
```

## Indentation

### Indent Types

| Type                             | Description              |
| -------------------------------- | ------------------------ |
| `Indent.getNoneIndent()`         | No indentation           |
| `Indent.getNormalIndent()`       | Standard indentation     |
| `Indent.getSpaceIndent(size)`    | Fixed space indentation  |
| `Indent.getContinuationIndent()` | Continuation indentation |

### Applying Indentation

```kotlin
class MyBlock : AbstractBlock {
    override fun getIndent(): Indent? {
        return when (myNode.elementType) {
            MyTokenTypes.STATEMENT -> Indent.getNormalIndent()
            MyTokenTypes.CONTINUATION -> Indent.getContinuationIndent()
            else -> Indent.getNoneIndent()
        }
    }
}
```

### Child Indentation

```kotlin
override fun getChildIndent(childIndex: Int): Indent? {
    return when (myNode.elementType) {
        MyTokenTypes.BLOCK -> Indent.getNormalIndent()
        else -> Indent.getNoneIndent()
    }
}
```

## Wrapping

### Wrap Types

```kotlin
// Always wrap
Wrap.createWrap(WrapType.ALWAYS, true)

// Wrap if long
Wrap.createWrap(WrapType.NORMAL, true)

// Never wrap
Wrap.createWrap(WrapType.NONE, true)

// Chop down if long
Wrap.createWrap(WrapType.CHOP_DOWN_IF_LONG, true)
```

### Applying Wraps

```kotlin
class MyBlock : AbstractBlock {
    private val wrap: Wrap? = when (myNode.elementType) {
        MyTokenTypes.PARAMETER -> Wrap.createWrap(WrapType.NORMAL, true)
        else -> null
    }

    constructor(node: ASTNode, settings: CodeStyleSettings) : super(node, null, null) {
        // Wrap is applied in constructor
    }
}
```

## Alignment

### Creating Alignment

```kotlin
class MyBlock : AbstractBlock {
    private val alignment: Alignment? = when (myNode.elementType) {
        MyTokenTypes.ASSIGNMENT -> Alignment.createAlignment()
        else -> null
    }

    constructor(node: ASTNode, settings: CodeStyleSettings) : super(node, null, alignment) {
        // Alignment is applied in constructor
    }
}
```

### Shared Alignment

```kotlin
class MyFormattingModelBuilder : FormattingModelBuilder {
    private val sharedAlignment = Alignment.createAlignment()

    override fun createModel(formattingContext: FormattingContext): FormattingModel {
        // Pass shared alignment to blocks that should align
    }
}
```

## Code Style Settings

### Accessing Settings

```kotlin
val settings = CodeStyleSettingsManager.getSettings(project)

// Common settings
val indentSize = settings.indentOptions.INDENT_SIZE
val tabSize = settings.indentOptions.TAB_SIZE
val useTab = settings.indentOptions.USE_TAB_CHARACTER
val rightMargin = settings.getRightMargin(fileType)
```

### Custom Settings

```kotlin
class MyCodeStyleSettings(container: CodeStyleSettings) : CustomCodeStyleSettings("MySettings", container) {
    @JvmField var SPACE_AROUND_OPERATOR = true
    @JvmField var KEEP_BLANK_LINES = 2
}
```

### Registering Custom Settings

```xml
<extensions defaultExtensionNs="com.intellij">
    <codeStyleSettingsProvider
        implementation="com.example.MyCodeStyleSettingsProvider"/>
    <langCodeStyleSettingsProvider
        implementation="com.example.MyLanguageCodeStyleSettingsProvider"/>
</extensions>
```

## Programmatic Formatting

### Formatting a File

```kotlin
fun formatFile(psiFile: PsiFile) {
    WriteAction.run<Exception> {
        CodeStyleManager.getInstance(psiFile.project)
            .reformat(psiFile)
    }
}
```

### Formatting a Range

```kotlin
fun formatRange(psiFile: PsiFile, start: Int, end: Int) {
    WriteAction.run<Exception> {
        CodeStyleManager.getInstance(psiFile.project)
            .reformatText(psiFile, start, end)
    }
}
```

### Formatting with Options

```kotlin
fun formatWithOptimizations(psiFile: PsiFile) {
    WriteAction.run<Exception> {
        val manager = CodeStyleManager.getInstance(psiFile.project)

        // Reformat
        manager.reformat(psiFile)

        // Optimize imports
        JavaCodeStyleManager.getInstance(psiFile.project)
            .optimizeImports(psiFile)

        // Shorten references
        JavaCodeStyleManager.getInstance(psiFile.project)
            .shortenClassReferences(psiFile)
    }
}
```

## Common Patterns

### Formatting a Block Statement

```kotlin
class BlockStatementBlock(node: ASTNode, settings: CodeStyleSettings) : AbstractBlock(node, null, null) {
    override fun buildChildren(): List<Block> {
        val children = mutableListOf<Block>()
        var child = myNode.firstChildNode

        while (child != null) {
            when (child.elementType) {
                TokenType.WHITE_SPACE -> { /* skip */ }
                MyTokenTypes.LBRACE, MyTokenTypes.RBRACE -> {
                    children.add(MyBlock(child, null, null, settings))
                }
                else -> {
                    children.add(MyBlock(child, null, null, settings).apply {
                        // Indent child statements
                    })
                }
            }
            child = child.treeNext
        }

        return children
    }

    override fun getSpacing(child1: Block?, child2: Block): Spacing? {
        return when {
            child1 is MyBlock && child1.myNode.elementType == MyTokenTypes.LBRACE ->
                Spacing.createSpacing(0, 0, 1, false, 0)  // Line break after {
            child2 is MyBlock && child2.myNode.elementType == MyTokenTypes.RBRACE ->
                Spacing.createSpacing(0, 0, 1, false, 0)  // Line break before }
            else -> Spacing.getReadOnlySpacing()
        }
    }

    override fun getIndent(): Indent? = Indent.getNoneIndent()
}
```

## Related Topics

- [Document Model](document-model.md) - Working with documents
- [Modifying PSI](../03-psi/modifying-psi.md) - Safe modifications
- [Language Support](../05-language-support/) - Custom language features

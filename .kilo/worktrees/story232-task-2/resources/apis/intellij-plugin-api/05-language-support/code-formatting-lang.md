# Language-Specific Code Formatting

This guide covers implementing custom formatting rules for your language, building on the general formatting concepts.

## Overview

### Formatting for Custom Languages

Custom language formatting involves:

- Defining formatting model builder
- Creating block structure
- Specifying spacing and indentation rules
- Handling wrapping behavior

## Formatting Model Builder

### Language-Specific Builder

```kotlin
class MyFormattingModelBuilder : FormattingModelBuilder {
    override fun createModel(formattingContext: FormattingContext): FormattingModel {
        val element = formattingContext.psiElement
        val settings = formattingContext.codeStyleSettings

        val rootBlock = MyRootBlock(element.node, settings)
        return FormattingModelProvider.createFormattingModelForPsiFile(
            element.containingFile,
            rootBlock,
            settings
        )
    }
}
```

### Registering the Builder

```xml
<extensions defaultExtensionNs="com.intellij">
    <lang.formatter
        language="MyLanguage"
        implementationClass="com.example.MyFormattingModelBuilder"/>
</extensions>
```

## Block Implementation

### Language-Specific Blocks

```kotlin
class MyBlock(
    node: ASTNode,
    wrap: Wrap?,
    alignment: Alignment?,
    private val settings: CodeStyleSettings
) : AbstractBlock(node, wrap, alignment) {

    override fun buildChildren(): List<Block> {
        val children = mutableListOf<Block>()
        var child = myNode.firstChildNode

        while (child != null) {
            if (!isWhitespaceOrComment(child)) {
                children.add(createBlock(child))
            }
            child = child.treeNext
        }

        return children
    }

    private fun createBlock(node: ASTNode): MyBlock {
        return when (node.elementType) {
            MyElementTypes.CLASS -> MyClassBlock(node, settings)
            MyElementTypes.FUNCTION -> MyFunctionBlock(node, settings)
            MyElementTypes.BLOCK -> MyCodeBlockBlock(node, settings)
            else -> MyBlock(node, null, null, settings)
        }
    }

    override fun getSpacing(child1: Block?, child2: Block): Spacing? {
        return MySpacingBuilder(settings).getSpacing(this, child1, child2)
    }

    override fun getIndent(): Indent? {
        return when (myNode.elementType) {
            MyElementTypes.STATEMENT -> Indent.getNormalIndent()
            else -> Indent.getNoneIndent()
        }
    }

    private fun isWhitespaceOrComment(node: ASTNode): Boolean {
        return node.elementType in setOf(
            TokenType.WHITE_SPACE,
            MyTokenTypes.LINE_COMMENT,
            MyTokenTypes.BLOCK_COMMENT
        )
    }
}
```

## Spacing Builder

### Language-Specific Spacing

```kotlin
class MySpacingBuilder(private val settings: CodeStyleSettings) {

    fun getSpacing(parent: MyBlock, child1: Block?, child2: Block): Spacing? {
        val parentType = parent.node.elementType
        val child1Type = (child1 as? MyBlock)?.node?.elementType
        val child2Type = (child2 as? MyBlock)?.node?.elementType

        return when {
            // Class body
            parentType == MyElementTypes.CLASS_BODY -> {
                spacingBetweenMembers(child1Type, child2Type)
            }

            // Function parameters
            parentType == MyElementTypes.PARAMETER_LIST -> {
                spacingBetweenParameters(child1Type, child2Type)
            }

            // Binary expressions
            parentType == MyElementTypes.BINARY_EXPRESSION -> {
                spacingAroundOperators(child1Type, child2Type)
            }

            // Statements
            parentType == MyElementTypes.BLOCK -> {
                spacingBetweenStatements(child1Type, child2Type)
            }

            else -> Spacing.getReadOnlySpacing()
        }
    }

    private fun spacingBetweenMembers(child1: IElementType?, child2: IElementType?): Spacing? {
        if (child1 == null || child2 == null) return null

        // Blank line between class members
        return Spacing.createSpacing(0, 0, settings.BLANK_LINES_AROUND_FIELD + 1, true, 0)
    }

    private fun spacingBetweenParameters(child1: IElementType?, child2: IElementType?): Spacing? {
        if (child1 == MyTokenTypes.COMMA) {
            // Space after comma
            return Spacing.createSpacing(1, 1, 0, false, 0)
        }
        return Spacing.getReadOnlySpacing()
    }

    private fun spacingAroundOperators(child1: IElementType?, child2: IElementType?): Spacing? {
        if (child1 in MyTokenSets.OPERATORS || child2 in MyTokenSets.OPERATORS) {
            // Space around operators
            return Spacing.createSpacing(1, 1, 0, false, 0)
        }
        return Spacing.getReadOnlySpacing()
    }

    private fun spacingBetweenStatements(child1: IElementType?, child2: IElementType?): Spacing? {
        // Line break between statements
        return Spacing.createSpacing(0, 0, 1, true, 0)
    }
}
```

## Indentation Rules

### Language-Specific Indentation

```kotlin
class MyIndentProcessor(private val settings: CodeStyleSettings) {

    fun getChildIndent(parent: ASTNode, child: ASTNode): Indent? {
        return when (parent.elementType) {
            MyElementTypes.CLASS_BODY -> Indent.getNormalIndent()
            MyElementTypes.BLOCK -> Indent.getNormalIndent()
            MyElementTypes.IF_STATEMENT -> getIfIndent(child)
            MyElementTypes.FOR_STATEMENT -> getForIndent(child)
            else -> Indent.getNoneIndent()
        }
    }

    private fun getIfIndent(child: ASTNode): Indent? {
        return when (child.elementType) {
            MyElementTypes.THEN_BLOCK -> Indent.getNormalIndent()
            MyElementTypes.ELSE_BLOCK -> Indent.getNormalIndent()
            else -> Indent.getNoneIndent()
        }
    }

    private fun getForIndent(child: ASTNode): Indent? {
        return when (child.elementType) {
            MyElementTypes.FOR_BODY -> Indent.getNormalIndent()
            else -> Indent.getNoneIndent()
        }
    }
}
```

## Wrapping Rules

### Language-Specific Wrapping

```kotlin
class MyWrapProcessor {

    fun getWrap(parent: ASTNode, child: ASTNode): Wrap? {
        return when (parent.elementType) {
            MyElementTypes.PARAMETER_LIST -> {
                if (shouldWrapParameters(parent)) {
                    Wrap.createWrap(WrapType.NORMAL, true)
                } else null
            }
            MyElementTypes.ARGUMENT_LIST -> {
                Wrap.createWrap(WrapType.CHOP_DOWN_IF_LONG, true)
            }
            MyElementTypes.BINARY_EXPRESSION -> {
                Wrap.createWrap(WrapType.NORMAL, false)
            }
            else -> null
        }
    }

    private fun shouldWrapParameters(node: ASTNode): Boolean {
        // Check if parameters exceed line length
        val text = node.text
        return text.length > 80
    }
}
```

## Code Style Settings

### Custom Settings

```kotlin
class MyCodeStyleSettings(container: CodeStyleSettings) :
    CustomCodeStyleSettings("MyLanguageSettings", container) {

    @JvmField var SPACE_AROUND_OPERATORS = true
    @JvmField var SPACE_AFTER_COMMA = true
    @JvmField var SPACE_BEFORE_COMMA = false
    @JvmField var BLANK_LINES_BETWEEN_METHODS = 1
    @JvmField var BLANK_LINES_BETWEEN_CLASSES = 2
    @JvmField var WRAP_PARAMETERS = false
    @JvmField var INDENT_SIZE = 4
    @JvmField var TAB_SIZE = 4
    @JvmField var USE_TAB_CHARACTER = false
}
```

### Settings Provider

```kotlin
class MyCodeStyleSettingsProvider : CodeStyleSettingsProvider() {
    override fun createCustomSettings(settings: CodeStyleSettings): CustomCodeStyleSettings? {
        return MyCodeStyleSettings(settings)
    }

    override fun getConfigurableDisplayName(): String = "My Language"
}
```

### Language Settings Provider

```kotlin
class MyLanguageCodeStyleSettingsProvider : LanguageCodeStyleSettingsProvider() {
    override fun getLanguage(): Language = MyLanguage.INSTANCE

    override fun getCodeSample(settingsType: SettingsType): String? {
        return when (settingsType) {
            SettingsType.INDENT_SETTINGS -> INDENT_SAMPLE
            SettingsType.SPACING_SETTINGS -> SPACING_SAMPLE
            SettingsType.WRAPPING_AND_BRACES_SETTINGS -> WRAPPING_SAMPLE
            else -> null
        }
    }

    override fun customizeDefaults(commonSettings: CommonCodeStyleSettings, indentOptions: CommonCodeStyleSettings.IndentOptions) {
        indentOptions.INDENT_SIZE = 4
        indentOptions.TAB_SIZE = 4
        indentOptions.USE_TAB_CHARACTER = false
    }

    companion object {
        private const val INDENT_SAMPLE = """
class MyClass {
    function myFunction() {
        if (true) {
            return 42;
        }
    }
}
"""

        private const val SPACING_SAMPLE = """
function myFunction(param1, param2) {
    return param1 + param2;
}
"""

        private const val WRAPPING_SAMPLE = """
class MyClass {
    function longFunction(
        param1,
        param2,
        param3
    ) {
        // body
    }
}
"""
    }
}
```

### Registering Settings

```xml
<extensions defaultExtensionNs="com.intellij">
    <codeStyleSettingsProvider
        implementation="com.example.MyCodeStyleSettingsProvider"/>
    <langCodeStyleSettingsProvider
        implementation="com.example.MyLanguageCodeStyleSettingsProvider"/>
</extensions>
```

## Related Topics

- [Code Formatting](../04-editor-features/code-formatting.md) - General formatting
- [Language Registration](language-registration.md) - Language definition
- [Lexer and Parser](lexer-and-parser.md) - Parsing for formatting

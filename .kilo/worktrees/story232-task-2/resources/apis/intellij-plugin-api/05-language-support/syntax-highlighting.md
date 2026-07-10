# Syntax Highlighting

Syntax highlighting colors tokens and elements in the editor. This guide covers implementing custom syntax highlighting for your language.

## Overview

### Highlighting Layers

| Layer                  | Purpose                      |
| ---------------------- | ---------------------------- |
| Lexer highlighting     | Colors tokens from the lexer |
| Parser highlighting    | Colors PSI elements          |
| Annotator highlighting | Semantic highlighting        |

### Highlighting Process

1. Lexer produces tokens
2. Syntax highlighter maps tokens to text attributes
3. Editor applies attributes to text

## Creating a Syntax Highlighter

### SyntaxHighlighter Implementation

```kotlin
class MySyntaxHighlighter : SyntaxHighlighterBase() {
    override fun getHighlightingLexer(): Lexer {
        return MyLexer()
    }

    override fun getTokenHighlights(tokenType: IElementType): Array<TextAttributesKey> {
        return when (tokenType) {
            // Keywords
            MyTokenTypes.IF, MyTokenTypes.ELSE,
            MyTokenTypes.FOR, MyTokenTypes.WHILE,
            MyTokenTypes.RETURN -> KEYWORD_KEYS

            // Identifiers
            MyTokenTypes.IDENTIFIER -> IDENTIFIER_KEYS

            // Literals
            MyTokenTypes.STRING -> STRING_KEYS
            MyTokenTypes.NUMBER -> NUMBER_KEYS

            // Comments
            MyTokenTypes.LINE_COMMENT, MyTokenTypes.BLOCK_COMMENT -> COMMENT_KEYS

            // Operators
            MyTokenTypes.PLUS, MyTokenTypes.MINUS,
            MyTokenTypes.EQ, MyTokenTypes.ASSIGN -> OPERATOR_KEYS

            // Brackets
            MyTokenTypes.LPAREN, MyTokenTypes.RPAREN -> PAREN_KEYS
            MyTokenTypes.LBRACE, MyTokenTypes.RBRACE -> BRACE_KEYS

            else -> EMPTY_KEYS
        }
    }

    companion object {
        val KEYWORD = TextAttributesKey.createTextAttributesKey(
            "MY_KEYWORD",
            DefaultLanguageHighlighterColors.KEYWORD
        )
        val IDENTIFIER = TextAttributesKey.createTextAttributesKey(
            "MY_IDENTIFIER",
            DefaultLanguageHighlighterColors.IDENTIFIER
        )
        val STRING = TextAttributesKey.createTextAttributesKey(
            "MY_STRING",
            DefaultLanguageHighlighterColors.STRING
        )
        val NUMBER = TextAttributesKey.createTextAttributesKey(
            "MY_NUMBER",
            DefaultLanguageHighlighterColors.NUMBER
        )
        val COMMENT = TextAttributesKey.createTextAttributesKey(
            "MY_COMMENT",
            DefaultLanguageHighlighterColors.LINE_COMMENT
        )
        val OPERATOR = TextAttributesKey.createTextAttributesKey(
            "MY_OPERATOR",
            DefaultLanguageHighlighterColors.OPERATION_SIGN
        )
        val PAREN = TextAttributesKey.createTextAttributesKey(
            "MY_PAREN",
            DefaultLanguageHighlighterColors.PARENTHESES
        )
        val BRACE = TextAttributesKey.createTextAttributesKey(
            "MY_BRACE",
            DefaultLanguageHighlighterColors.BRACES
        )

        val KEYWORD_KEYS = arrayOf(KEYWORD)
        val IDENTIFIER_KEYS = arrayOf(IDENTIFIER)
        val STRING_KEYS = arrayOf(STRING)
        val NUMBER_KEYS = arrayOf(NUMBER)
        val COMMENT_KEYS = arrayOf(COMMENT)
        val OPERATOR_KEYS = arrayOf(OPERATOR)
        val PAREN_KEYS = arrayOf(PAREN)
        val BRACE_KEYS = arrayOf(BRACE)
        val EMPTY_KEYS = emptyArray<TextAttributesKey>()
    }
}
```

### SyntaxHighlighterFactory

```kotlin
class MySyntaxHighlighterFactory : SyntaxHighlighterFactory() {
    override fun getSyntaxHighlighter(project: Project?, virtualFile: VirtualFile?): SyntaxHighlighter {
        return MySyntaxHighlighter()
    }
}
```

### Registering the Highlighter

```xml
<extensions defaultExtensionNs="com.intellij">
    <lang.syntaxHighlighterFactory
        language="MyLanguage"
        implementationClass="com.example.MySyntaxHighlighterFactory"/>
</extensions>
```

## Color Settings

### Color Settings Page

```kotlin
class MyColorSettingsPage : ColorSettingsPage {
    override fun getHighlightDisplayName(): String = "My Language"

    override fun getIcon(): Icon = MyFileType.INSTANCE.icon

    override fun getAttributeDescriptors(): Array<AttributesDescriptor> {
        return arrayOf(
            AttributesDescriptor("Keyword", MySyntaxHighlighter.KEYWORD),
            AttributesDescriptor("Identifier", MySyntaxHighlighter.IDENTIFIER),
            AttributesDescriptor("String", MySyntaxHighlighter.STRING),
            AttributesDescriptor("Number", MySyntaxHighlighter.NUMBER),
            AttributesDescriptor("Comment", MySyntaxHighlighter.COMMENT),
            AttributesDescriptor("Operator", MySyntaxHighlighter.OPERATOR),
            AttributesDescriptor("Parentheses", MySyntaxHighlighter.PAREN),
            AttributesDescriptor("Braces", MySyntaxHighlighter.BRACE)
        )
    }

    override fun getColorDescriptors(): Array<ColorDescriptor> {
        return ColorDescriptor.EMPTY_ARRAY
    }

    override fun getDemoText(): String {
        return """
            // This is a comment
            class MyClass {
                function myFunction(param) {
                    if (param == 42) {
                        return "The answer";
                    }
                    return 0;
                }
            }
        """.trimIndent()
    }

    override fun getAdditionalHighlightingTagToDescriptorMap(): Map<String, TextAttributesKey>? {
        return null
    }

    override fun getDisplayName(): String = "My Language"
}
```

### Registering Color Settings

```xml
<extensions defaultExtensionNs="com.intellij">
    <colorSettingsPage
        implementation="com.example.MyColorSettingsPage"/>
</extensions>
```

## Annotator-Based Highlighting

### Creating an Annotator

```kotlin
class MyAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        when (element) {
            is MyClass -> highlightClass(element, holder)
            is MyFunction -> highlightFunction(element, holder)
            is MyIdentifier -> highlightIdentifier(element, holder)
        }
    }

    private fun highlightClass(element: MyClass, holder: AnnotationHolder) {
        val nameIdentifier = element.nameIdentifier ?: return
        holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
            .range(nameIdentifier)
            .textAttributes(MySyntaxHighlighter.CLASS_NAME)
            .create()
    }

    private fun highlightFunction(element: MyFunction, holder: AnnotationHolder) {
        val nameIdentifier = element.nameIdentifier ?: return
        holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
            .range(nameIdentifier)
            .textAttributes(MySyntaxHighlighter.FUNCTION_NAME)
            .create()
    }

    private fun highlightIdentifier(element: MyIdentifier, holder: AnnotationHolder) {
        // Check if identifier is a reference
        val reference = element.reference
        if (reference != null) {
            val resolved = reference.resolve()
            when (resolved) {
                is MyClass -> {
                    holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
                        .range(element)
                        .textAttributes(MySyntaxHighlighter.CLASS_REFERENCE)
                        .create()
                }
                is MyFunction -> {
                    holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
                        .range(element)
                        .textAttributes(MySyntaxHighlighter.FUNCTION_REFERENCE)
                        .create()
                }
            }
        }
    }
}
```

### Registering the Annotator

```xml
<extensions defaultExtensionNs="com.intellij">
    <annotator
        language="MyLanguage"
        implementationClass="com.example.MyAnnotator"/>
</extensions>
```

## Text Attributes

### Creating Custom Attributes

```kotlin
companion object {
    val CLASS_NAME = TextAttributesKey.createTextAttributesKey(
        "MY_CLASS_NAME",
        DefaultLanguageHighlighterColors.CLASS_NAME
    )

    val FUNCTION_NAME = TextAttributesKey.createTextAttributesKey(
        "MY_FUNCTION_NAME",
        DefaultLanguageHighlighterColors.FUNCTION_DECLARATION
    )

    val CLASS_REFERENCE = TextAttributesKey.createTextAttributesKey(
        "MY_CLASS_REFERENCE",
        DefaultLanguageHighlighterColors.CLASS_REFERENCE
    )

    val FUNCTION_REFERENCE = TextAttributesKey.createTextAttributesKey(
        "MY_FUNCTION_REFERENCE",
        DefaultLanguageHighlighterColors.FUNCTION_CALL
    )
}
```

### Default Colors

| Default Key            | Description           |
| ---------------------- | --------------------- |
| `KEYWORD`              | Language keywords     |
| `IDENTIFIER`           | General identifiers   |
| `STRING`               | String literals       |
| `NUMBER`               | Numeric literals      |
| `LINE_COMMENT`         | Single-line comments  |
| `BLOCK_COMMENT`        | Multi-line comments   |
| `OPERATION_SIGN`       | Operators             |
| `PARENTHESES`          | Parentheses           |
| `BRACES`               | Curly braces          |
| `BRACKETS`             | Square brackets       |
| `CLASS_NAME`           | Class declarations    |
| `FUNCTION_DECLARATION` | Function declarations |
| `CLASS_REFERENCE`      | Class usages          |
| `FUNCTION_CALL`        | Function calls        |

## Highlight Info

### Creating Highlight Info

```kotlin
holder.newAnnotation(HighlightSeverity.ERROR, "Error message")
    .range(element)
    .highlightType(ProblemHighlightType.ERROR)
    .create()

holder.newAnnotation(HighlightSeverity.WARNING, "Warning message")
    .range(element)
    .highlightType(ProblemHighlightType.WARNING)
    .create()

holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
    .range(element)
    .textAttributes(MySyntaxHighlighter.CUSTOM_ATTRIBUTE)
    .create()
```

### Annotation Builder Options

```kotlin
holder.newAnnotation(HighlightSeverity.WARNING, "Unused variable")
    .range(element)
    .highlightType(ProblemHighlightType.LIKE_UNUSED_SYMBOL)
    .withFix(MyQuickFix())
    .afterEndOfLine()
    .create()
```

## Related Topics

- [Lexer and Parser](lexer-and-parser.md) - Implementing parsing
- [Annotators](../06-analysis-features/annotators.md) - Code analysis
- [Editor Basics](../04-editor-features/editor-basics.md) - Editor architecture

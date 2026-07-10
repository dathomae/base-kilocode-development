# Syntax Errors

Syntax errors are displayed when the parser encounters invalid code. This guide covers how to display and manage syntax errors in custom languages.

## Overview

### Syntax Error Display

Syntax errors appear as:

- Red underlines in the editor
- Error markers in the gutter
- Entries in the Problems view
- Error highlighting in the structure view

### Error Sources

| Source           | Description                              |
| ---------------- | ---------------------------------------- |
| Parser errors    | Invalid syntax detected during parsing   |
| Error elements   | PSI elements representing invalid code   |
| Annotator errors | Semantic errors detected during analysis |

## Parser Error Handling

### Creating Error Elements

```kotlin
class MyParser : PsiParser {
    override fun parse(root: IElementType, builder: PsiBuilder): ASTNode {
        val rootMarker = builder.mark()

        while (!builder.eof()) {
            if (!parseStatement(builder)) {
                // Create error element
                val errorMarker = builder.mark()
                builder.advanceLexer()
                errorMarker.error("Unexpected token")
            }
        }

        rootMarker.done(root)
        return builder.treeBuilt
    }
}
```

### Error Recovery

```kotlin
private fun parseStatement(builder: PsiBuilder): Boolean {
    return when (builder.tokenType) {
        MyTokenTypes.IF -> parseIfStatement(builder)
        MyTokenTypes.FOR -> parseForStatement(builder)
        MyTokenTypes.WHILE -> parseWhileStatement(builder)
        MyTokenTypes.RETURN -> parseReturnStatement(builder)
        MyTokenTypes.IDENTIFIER -> parseExpressionStatement(builder)
        else -> {
            // Error recovery: skip to next statement boundary
            builder.error("Expected statement")
            skipToStatementBoundary(builder)
            false
        }
    }
}

private fun skipToStatementBoundary(builder: PsiBuilder) {
    while (!builder.eof() && builder.tokenType !in STATEMENT_BOUNDARIES) {
        builder.advanceLexer()
    }
}

companion object {
    private val STATEMENT_BOUNDARIES = setOf(
        MyTokenTypes.SEMICOLON,
        MyTokenTypes.RBRACE,
        MyTokenTypes.IF,
        MyTokenTypes.FOR,
        MyTokenTypes.WHILE,
        MyTokenTypes.RETURN
    )
}
```

## Error Elements

### Accessing Error Elements

```kotlin
fun findErrors(file: PsiFile): List<PsiErrorElement> {
    val errors = mutableListOf<PsiErrorElement>()

    file.accept(object : PsiRecursiveElementVisitor() {
        override fun visitErrorElement(element: PsiErrorElement) {
            errors.add(element)
            super.visitErrorElement(element)
        }
    })

    return errors
}
```

### Error Element Properties

```kotlin
val error: PsiErrorElement = ...

// Error description
val description = error.errorDescription

// Error location
val range = error.textRange
val startOffset = error.textRange?.startOffset
val endOffset = error.textRange?.endOffset
```

## Custom Syntax Error Annotator

### Annotating Parser Errors

```kotlin
class MySyntaxErrorAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        if (element is PsiErrorElement) {
            holder.newAnnotation(
                HighlightSeverity.ERROR,
                element.errorDescription ?: "Syntax error"
            )
                .range(element)
                .create()
        }
    }
}
```

### Enhanced Error Messages

```kotlin
class MyEnhancedErrorAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        if (element !is PsiErrorElement) return

        val message = enhanceErrorMessage(element)

        holder.newAnnotation(HighlightSeverity.ERROR, message)
            .range(element)
            .withFix(SuggestFixAction(element))
            .create()
    }

    private fun enhanceErrorMessage(error: PsiErrorElement): String {
        val baseMessage = error.errorDescription ?: "Syntax error"
        val context = error.parent

        return when {
            context is MyFunction && error.prevSibling?.text == "(" ->
                "$baseMessage - expected parameter list"
            context is MyClass && error.prevSibling?.text == "{" ->
                "$baseMessage - expected class body"
            else -> baseMessage
        }
    }
}
```

## Expected Token Errors

### Creating Expected Token Errors

```kotlin
private fun expect(builder: PsiBuilder, expected: IElementType, message: String): Boolean {
    if (builder.tokenType == expected) {
        builder.advanceLexer()
        return true
    }

    builder.error(message)
    return false
}

// Usage
private fun parseIfStatement(builder: PsiBuilder): Boolean {
    if (builder.tokenType != MyTokenTypes.IF) return false

    val marker = builder.mark()
    builder.advanceLexer() // 'if'

    if (!expect(builder, MyTokenTypes.LPAREN, "Expected '(' after 'if'")) {
        marker.drop()
        return false
    }

    parseExpression(builder)

    if (!expect(builder, MyTokenTypes.RPAREN, "Expected ')' after condition")) {
        // Try to recover
        skipToToken(builder, MyTokenTypes.LBRACE)
    }

    parseBlock(builder)

    if (builder.tokenType == MyTokenTypes.ELSE) {
        builder.advanceLexer()
        parseBlock(builder)
    }

    marker.done(MyElementTypes.IF_STATEMENT)
    return true
}
```

## Error Highlighting

### Custom Error Highlighting

```kotlin
class MyErrorHighlighter : ErrorHighlighter {
    override fun highlightError(element: PsiErrorElement, holder: AnnotationHolder) {
        val attributes = TextAttributes().apply {
            effectType = EffectType.WAVE_UNDERSCORE
            effectColor = Color.RED
        }

        holder.newSilentAnnotation(HighlightSeverity.ERROR)
            .range(element)
            .enforcedTextAttributes(attributes)
            .create()
    }
}
```

## Error Quick Fixes

### Adding Quick Fixes to Errors

```kotlin
class MyErrorQuickFix : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        if (element is PsiErrorElement) {
            val fixes = suggestFixes(element)

            val builder = holder.newAnnotation(
                HighlightSeverity.ERROR,
                element.errorDescription ?: "Syntax error"
            ).range(element)

            for (fix in fixes) {
                builder.withFix(fix)
            }

            builder.create()
        }
    }

    private fun suggestFixes(error: PsiErrorElement): List<LocalQuickFix> {
        val fixes = mutableListOf<LocalQuickFix>()

        // Suggest inserting missing token
        val expected = inferExpectedToken(error)
        if (expected != null) {
            fixes.add(InsertTokenFix(expected))
        }

        // Suggest removing unexpected token
        if (error.text.isNotEmpty()) {
            fixes.add(DeleteTokenFix())
        }

        return fixes
    }

    private fun inferExpectedToken(error: PsiErrorElement): IElementType? {
        val prev = error.prevSibling?.lastChild?.node?.elementType
        return when (prev) {
            MyTokenTypes.IF -> MyTokenTypes.LPAREN
            MyTokenTypes.FOR -> MyTokenTypes.LPAREN
            MyTokenTypes.WHILE -> MyTokenTypes.LPAREN
            else -> null
        }
    }
}
```

### Insert Token Fix

```kotlin
class InsertTokenFix(private val tokenType: IElementType) : LocalQuickFix {
    override fun getFamilyName(): String = "Insert missing token"

    override fun applyFix(project: Project, descriptor: ProblemDescriptor) {
        val element = descriptor.psiElement
        val text = getTokenText(tokenType)

        WriteAction.run<Exception> {
            val factory = PsiElementFactory.getInstance(project)
            val newElement = factory.createDummyHolder(text, tokenType, null)
            element.parent.addBefore(newElement, element)
        }
    }

    private fun getTokenText(type: IElementType): String {
        return when (type) {
            MyTokenTypes.LPAREN -> "("
            MyTokenTypes.RPAREN -> ")"
            MyTokenTypes.LBRACE -> "{"
            MyTokenTypes.RBRACE -> "}"
            MyTokenTypes.SEMICOLON -> ";"
            else -> ""
        }
    }
}
```

## Related Topics

- [Annotators](annotators.md) - Semantic analysis
- [Inspections](inspections.md) - Code quality checks
- [Lexer and Parser](../05-language-support/lexer-and-parser.md) - Parser implementation

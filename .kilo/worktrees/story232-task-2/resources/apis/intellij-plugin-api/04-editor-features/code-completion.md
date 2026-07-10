# Code Completion

Code completion provides suggestions as users type. Plugins can contribute custom completion providers for specific languages or contexts.

## Overview

### Completion Types

| Type       | Trigger          | Description           |
| ---------- | ---------------- | --------------------- |
| Basic      | Ctrl+Space       | Standard completion   |
| Smart      | Ctrl+Shift+Space | Type-aware completion |
| Class Name | Ctrl+Space twice | Class name completion |
| Line       | Ctrl+Shift+Enter | Complete current line |

### Completion Architecture

Completion involves:

- **CompletionContributor**: Provides completion items
- **LookupElement**: A completion suggestion
- **CompletionResultSet**: Collects completion items
- **CompletionParameters**: Context for completion

## Creating a Completion Contributor

### Basic Contributor

```kotlin
class MyCompletionContributor : CompletionContributor() {
    init {
        extend(
            CompletionType.BASIC,
            PlatformPatterns.psiElement(),
            MyCompletionProvider()
        )
    }
}
```

### Registering the Contributor

```xml
<extensions defaultExtensionNs="com.intellij">
    <completion.contributor
        language="JAVA"
        implementationClass="com.example.MyCompletionContributor"/>
</extensions>
```

## Completion Provider

### Implementing a Provider

```kotlin
class MyCompletionProvider : CompletionProvider<CompletionParameters>() {
    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        // Add completion items
        result.addElement(LookupElementBuilder.create("suggestion"))
    }
}
```

### Using Patterns

```kotlin
init {
    // Complete after specific element
    extend(
        CompletionType.BASIC,
        PlatformPatterns.psiElement()
            .afterLeaf(PlatformPatterns.psiElement(JavaTokenType.DOT)),
        MyCompletionProvider()
    )

    // Complete in specific context
    extend(
        CompletionType.BASIC,
        PlatformPatterns.psiElement()
            .inside(PsiClass::class.java),
        MyCompletionProvider()
    )

    // Complete with language
    extend(
        CompletionType.BASIC,
        PlatformPatterns.psiElement()
            .withLanguage(JavaLanguage.INSTANCE),
        MyCompletionProvider()
    )
}
```

## Lookup Elements

### Basic Lookup Element

```kotlin
result.addElement(
    LookupElementBuilder.create("myMethod")
)
```

### Rich Lookup Elements

```kotlin
result.addElement(
    LookupElementBuilder.create("myMethod")
        .withTypeText("void")
        .withTailText("(String param)", true)
        .withIcon(AllIcons.Nodes.Method)
        .withStrikeoutness(false)
        .withBoldness(true)
)
```

### Lookup Element with Insertion Handler

```kotlin
result.addElement(
    LookupElementBuilder.create("myMethod")
        .withInsertHandler { context, item ->
            // Add parentheses after method name
            context.document.insertString(
                context.tailOffset,
                "()"
            )
            // Move caret inside parentheses
            context.editor.caretModel.moveToOffset(context.tailOffset - 1)
        }
)
```

### Prioritizing Items

```kotlin
result.addElement(
    PrioritizedLookupElement.withPriority(
        LookupElementBuilder.create("important"),
        100.0  // Higher priority appears first
    )
)
```

## Completion Parameters

### Accessing Context

```kotlin
override fun addCompletions(
    parameters: CompletionParameters,
    context: ProcessingContext,
    result: CompletionResultSet
) {
    val position = parameters.position
    val originalFile = parameters.originalFile
    val offset = parameters.offset
    val invocationCount = parameters.invocationCount

    // Get element at position
    val element = position.parent

    // Get surrounding context
    val containingClass = PsiTreeUtil.getParentOfType(element, PsiClass::class.java)
}
```

### Original Position

```kotlin
// Position where completion was triggered
val originalPosition = parameters.originalPosition

// Current position (may be different after typing)
val currentPosition = parameters.position
```

## Completion Result Set

### Adding Elements

```kotlin
// Single element
result.addElement(element)

// Multiple elements
result.addAllElements(elements)

// With prefix matching
result.withPrefixMatcher("prefix").addElement(element)
```

### Case Sensitivity

```kotlin
// Case-insensitive matching
result.caseInsensitive()
    .addElement(LookupElementBuilder.create("MyClass"))
```

### Stopping Completion

```kotlin
// Stop further contributors
result.stopHere()
```

## Common Patterns

### Method Completion

```kotlin
class MethodCompletionProvider : CompletionProvider<CompletionParameters>() {
    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        val psiClass = PsiTreeUtil.getParentOfType(
            parameters.position,
            PsiClass::class.java
        ) ?: return

        for (method in psiClass.methods) {
            result.addElement(
                LookupElementBuilder.create(method)
                    .withIcon(AllIcons.Nodes.Method)
                    .withTypeText(method.returnType?.presentableText)
                    .withTailText(formatParameters(method))
                    .withInsertHandler { context, _ ->
                        insertMethodCall(context, method)
                    }
            )
        }
    }

    private fun formatParameters(method: PsiMethod): String {
        return method.parameterList.parameters
            .joinToString(", ", "(", ")") {
                it.type.presentableText
            }
    }

    private fun insertMethodCall(context: InsertionContext, method: PsiMethod) {
        val hasParams = method.parameterList.parameters.isNotEmpty()
        context.document.insertString(context.tailOffset, "()")
        if (hasParams) {
            context.editor.caretModel.moveToOffset(context.tailOffset + 1)
        }
    }
}
```

### Keyword Completion

```kotlin
class KeywordCompletionProvider : CompletionProvider<CompletionParameters>() {
    private val keywords = listOf("if", "else", "for", "while", "return", "break", "continue")

    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        for (keyword in keywords) {
            result.addElement(
                LookupElementBuilder.create(keyword)
                    .withBoldness(true)
                    .withInsertHandler { context, _ ->
                        // Add space after keyword
                        context.document.insertString(context.tailOffset, " ")
                        context.editor.caretModel.moveToOffset(context.tailOffset + 1)
                    }
            )
        }
    }
}
```

### Variable Completion

```kotlin
class VariableCompletionProvider : CompletionProvider<CompletionParameters>() {
    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        val scope = PsiTreeUtil.getParentOfType(
            parameters.position,
            PsiCodeBlock::class.java
        ) ?: return

        scope.accept(object : JavaRecursiveElementVisitor() {
            override fun visitLocalVariable(variable: PsiLocalVariable) {
                result.addElement(
                    LookupElementBuilder.create(variable)
                        .withIcon(AllIcons.Nodes.Variable)
                        .withTypeText(variable.type.presentableText)
                )
            }
        })
    }
}
```

## Performance Considerations

### Lazy Completion

```kotlin
result.addElement(
    LookupElementBuilder.create(object : LookupElementPresentation() {
        // Presentation is computed lazily
    })
)
```

### Limiting Results

```kotlin
// Only add if prefix matches
if (result.prefixMatcher.prefixMatches("suggestion")) {
    result.addElement(LookupElementBuilder.create("suggestion"))
}
```

### Background Processing

Long-running completion should use background threads:

```kotlin
override fun addCompletions(
    parameters: CompletionParameters,
    context: ProcessingContext,
    result: CompletionResultSet
) {
    // Check for cancellation
    ProgressManager.checkCanceled()

    // Process in background if needed
    result.restartCompletionOnAnyPrefixChange()
}
```

## Related Topics

- [Editor Basics](editor-basics.md) - Editor architecture
- [Intentions](intentions.md) - Quick fixes
- [Language Support](../05-language-support/) - Custom language features

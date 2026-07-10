# Language-Specific Code Completion

This guide covers implementing code completion for custom languages, building on the general completion concepts.

## Overview

### Completion for Custom Languages

Custom language completion involves:

- Identifying completion contexts
- Providing relevant suggestions
- Handling insertion and formatting

### Completion Contexts

Common completion contexts:

- After keywords
- In expressions
- In declarations
- After dot (member access)

## Basic Completion Contributor

### Language-Specific Contributor

```kotlin
class MyCompletionContributor : CompletionContributor() {
    init {
        // Keyword completion
        extend(
            CompletionType.BASIC,
            PlatformPatterns.psiElement().withLanguage(MyLanguage.INSTANCE),
            KeywordCompletionProvider()
        )

        // Function name completion
        extend(
            CompletionType.BASIC,
            PlatformPatterns.psiElement()
                .inside(MyCallExpression::class.java)
                .withLanguage(MyLanguage.INSTANCE),
            FunctionCompletionProvider()
        )

        // Variable completion
        extend(
            CompletionType.BASIC,
            PlatformPatterns.psiElement()
                .inside(MyBlock::class.java)
                .withLanguage(MyLanguage.INSTANCE),
            VariableCompletionProvider()
        )
    }
}
```

### Registering the Contributor

```xml
<extensions defaultExtensionNs="com.intellij">
    <completion.contributor
        language="MyLanguage"
        implementationClass="com.example.MyCompletionContributor"/>
</extensions>
```

## Keyword Completion

### Keyword Provider

```kotlin
class KeywordCompletionProvider : CompletionProvider<CompletionParameters>() {
    private val keywords = listOf(
        "if", "else", "for", "while", "return",
        "class", "function", "var", "const"
    )

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

## Declaration Completion

### Class Members

```kotlin
class MemberCompletionProvider : CompletionProvider<CompletionParameters>() {
    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        val position = parameters.position
        val myClass = PsiTreeUtil.getParentOfType(position, MyClass::class.java) ?: return

        // Add methods
        for (method in myClass.methods) {
            result.addElement(
                LookupElementBuilder.create(method)
                    .withIcon(AllIcons.Nodes.Method)
                    .withTypeText(method.returnType?.text)
                    .withTailText(formatParameters(method))
            )
        }

        // Add fields
        for (field in myClass.fields) {
            result.addElement(
                LookupElementBuilder.create(field)
                    .withIcon(AllIcons.Nodes.Field)
                    .withTypeText(field.type?.text)
            )
        }
    }
}
```

## Reference Completion

### Completing References

```kotlin
class ReferenceCompletionProvider : CompletionProvider<CompletionParameters>() {
    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        val position = parameters.position
        val reference = position.parent?.reference as? MyReference ?: return

        // Get completion variants from reference
        val variants = reference.getVariants()

        for (variant in variants) {
            result.addElement(
                LookupElementBuilder.create(variant)
                    .withIcon(getIcon(variant))
                    .withTypeText(getTypeText(variant))
            )
        }
    }
}
```

### Reference with Variants

```kotlin
class MyReference(element: MyIdentifier) : PsiReferenceBase<MyIdentifier>(element) {
    override fun resolve(): PsiElement? {
        val name = element.text
        return findDefinition(name, element)
    }

    override fun getVariants(): Array<Any> {
        val scope = getResolutionScope(element)
        val variants = mutableListOf<LookupElement>()

        // Add local variables
        for (variable in scope.getLocalVariables()) {
            variants.add(
                LookupElementBuilder.create(variable)
                    .withIcon(AllIcons.Nodes.Variable)
                    .withTypeText(variable.type?.text)
            )
        }

        // Add functions
        for (function in scope.getFunctions()) {
            variants.add(
                LookupElementBuilder.create(function)
                    .withIcon(AllIcons.Nodes.Method)
                    .withTailText(formatParameters(function))
            )
        }

        // Add classes
        for (clazz in scope.getClasses()) {
            variants.add(
                LookupElementBuilder.create(clazz)
                    .withIcon(AllIcons.Nodes.Class)
            )
        }

        return variants.toTypedArray()
    }
}
```

## Dot Completion

### Member Access Completion

```kotlin
class DotCompletionProvider : CompletionProvider<CompletionParameters>() {
    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        val position = parameters.position

        // Check if we're after a dot
        val dotExpr = PsiTreeUtil.getParentOfType(position, MyDotExpression::class.java)
            ?: return

        // Resolve the qualifier type
        val qualifier = dotExpr.qualifier ?: return
        val type = resolveType(qualifier) ?: return

        // Get members of the type
        val members = getMembersOfType(type)

        for (member in members) {
            result.addElement(
                LookupElementBuilder.create(member.name)
                    .withIcon(getMemberIcon(member))
                    .withTypeText(member.type?.text)
                    .withInsertHandler { context, _ ->
                        if (member.isFunction) {
                            context.document.insertString(context.tailOffset, "()")
                            context.editor.caretModel.moveToOffset(context.tailOffset + 1)
                        }
                    }
            )
        }
    }
}
```

## Context-Sensitive Completion

### Statement Completion

```kotlin
class StatementCompletionProvider : CompletionProvider<CompletionParameters>() {
    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        val position = parameters.position
        val parent = position.parent

        when (parent) {
            is MyBlock -> {
                // Inside a block - suggest statements
                addStatementKeywords(result)
                addLocalVariables(parent, result)
                addLocalFunctions(parent, result)
            }
            is MyIfStatement -> {
                // Inside if condition
                addExpressionCompletions(result)
            }
            is MyForStatement -> {
                // Inside for loop
                addForLoopCompletions(result)
            }
        }
    }

    private fun addStatementKeywords(result: CompletionResultSet) {
        val keywords = listOf("if", "for", "while", "return", "var", "const")
        for (keyword in keywords) {
            result.addElement(
                LookupElementBuilder.create(keyword)
                    .withBoldness(true)
            )
        }
    }
}
```

## Template Completion

### Live Templates

```kotlin
class MyLiveTemplateContributor : CompletionContributor() {
    init {
        extend(
            CompletionType.BASIC,
            PlatformPatterns.psiElement().withLanguage(MyLanguage.INSTANCE),
            TemplateCompletionProvider()
        )
    }
}

class TemplateCompletionProvider : CompletionProvider<CompletionParameters>() {
    override fun addCompletions(
        parameters: CompletionParameters,
        context: ProcessingContext,
        result: CompletionResultSet
    ) {
        // Add template suggestions
        result.addElement(
            LookupElementBuilder.create("fori")
                .withTypeText("for (int i = 0; i < ; i++)")
                .withInsertHandler { context, _ ->
                    val template = TemplateBuilderFactory.getInstance()
                        .createTemplate("fori", "MyLanguage", "for (int \$i$ = 0; \$i$ < \$limit$; \$i$++) {\n$END$\n}")
                    TemplateManager.getInstance(context.project).startTemplate(context.editor, template)
                }
        )
    }
}
```

## Priority and Sorting

### Prioritized Elements

```kotlin
override fun addCompletions(
    parameters: CompletionParameters,
    context: ProcessingContext,
    result: CompletionResultSet
) {
    // High priority - exact matches
    result.addElement(
        PrioritizedLookupElement.withPriority(
            LookupElementBuilder.create("exactMatch"),
            100.0
        )
    )

    // Normal priority
    result.addElement(
        PrioritizedLookupElement.withPriority(
            LookupElementBuilder.create("normalMatch"),
            50.0
        )
    )

    // Low priority - less relevant
    result.addElement(
        PrioritizedLookupElement.withPriority(
            LookupElementBuilder.create("lessRelevant"),
            0.0
        )
    )
}
```

## Related Topics

- [Code Completion](../04-editor-features/code-completion.md) - General completion
- [Reference Resolution](reference-resolution.md) - Resolving references
- [Lexer and Parser](lexer-and-parser.md) - Parsing for context

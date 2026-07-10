# Annotators

Annotators provide semantic analysis and highlighting during the editor's highlighting pass. They can add errors, warnings, and visual decorations to code.

## Overview

### What Are Annotators?

Annotators run during the highlighting pass and can:

- Add error/warning annotations
- Apply text attributes for highlighting
- Add gutter icons
- Provide quick fixes

### Annotator vs. Inspection

| Aspect      | Annotator             | Inspection             |
| ----------- | --------------------- | ---------------------- |
| Timing      | Highlighting pass     | Analysis pass          |
| Scope       | Single file           | Single file or project |
| Purpose     | Semantic highlighting | Code quality           |
| Performance | Must be fast          | Can be slower          |

## Creating an Annotator

### Basic Annotator

```kotlin
class MyAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        // Check element type
        if (element !is MyIdentifier) return

        // Check if unresolved
        val reference = element.reference
        if (reference == null || reference.resolve() == null) {
            holder.newAnnotation(
                HighlightSeverity.ERROR,
                "Cannot resolve symbol '${element.text}'"
            )
                .range(element)
                .withFix(CreateSymbolFix(element))
                .create()
        }
    }
}
```

### Registering an Annotator

```xml
<extensions defaultExtensionNs="com.intellij">
    <annotator
        language="MyLanguage"
        implementationClass="com.example.MyAnnotator"/>
</extensions>
```

## Annotation Types

### Error Annotations

```kotlin
holder.newAnnotation(HighlightSeverity.ERROR, "This is an error")
    .range(element)
    .create()
```

### Warning Annotations

```kotlin
holder.newAnnotation(HighlightSeverity.WARNING, "This is a warning")
    .range(element)
    .create()
```

### Information Annotations

```kotlin
holder.newAnnotation(HighlightSeverity.INFORMATION, "Information message")
    .range(element)
    .create()
```

### Silent Annotations (Highlighting Only)

```kotlin
holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
    .range(element)
    .textAttributes(MyAttributes.KEYWORD)
    .create()
```

## Annotation Builder

### Full Annotation Options

```kotlin
holder.newAnnotation(HighlightSeverity.WARNING, "Unused variable")
    .range(element)
    .highlightType(ProblemHighlightType.LIKE_UNUSED_SYMBOL)
    .withFix(RemoveUnusedFix(element))
    .withFix(CommentOutFix(element))
    .afterEndOfLine()  // Show message at end of line
    .create()
```

### Highlight Types

| Type                       | Description           |
| -------------------------- | --------------------- |
| `ERROR`                    | Error highlighting    |
| `WARNING`                  | Warning highlighting  |
| `GENERIC_ERROR_OR_WARNING` | Generic error/warning |
| `LIKE_UNUSED_SYMBOL`       | Unused symbol style   |
| `LIKE_UNKNOWN_SYMBOL`      | Unknown symbol style  |
| `LIKE_DEPRECATED`          | Deprecated style      |
| `INFORMATION`              | Information style     |

## Common Annotator Patterns

### Unresolved Reference Annotator

```kotlin
class UnresolvedReferenceAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        if (element !is MyIdentifier) return

        val reference = element.reference ?: return

        if (reference.resolve() == null) {
            holder.newAnnotation(
                HighlightSeverity.ERROR,
                "Cannot resolve '${element.text}'"
            )
                .range(element)
                .highlightType(ProblemHighlightType.LIKE_UNKNOWN_SYMBOL)
                .withFix(CreateVariableFix(element))
                .withFix(CreateFunctionFix(element))
                .create()
        }
    }
}
```

### Unused Symbol Annotator

```kotlin
class UnusedSymbolAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        when (element) {
            is MyVariable -> checkUnusedVariable(element, holder)
            is MyFunction -> checkUnusedFunction(element, holder)
            is MyField -> checkUnusedField(element, holder)
        }
    }

    private fun checkUnusedVariable(variable: MyVariable, holder: AnnotationHolder) {
        if (isUsed(variable)) return

        holder.newAnnotation(
            HighlightSeverity.WARNING,
            "Variable '${variable.name}' is never used"
        )
            .range(variable.nameIdentifier ?: variable)
            .highlightType(ProblemHighlightType.LIKE_UNUSED_SYMBOL)
            .withFix(RemoveUnusedFix(variable))
            .create()
    }

    private fun isUsed(element: PsiNamedElement): Boolean {
        return ReferencesSearch.search(
            element,
            GlobalSearchScope.fileScope(element.containingFile)
        ).findFirst() != null
    }
}
```

### Type Mismatch Annotator

```kotlin
class TypeMismatchAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        if (element !is MyAssignment) return

        val leftType = inferType(element.left)
        val rightType = inferType(element.right)

        if (leftType != null && rightType != null && !isAssignable(leftType, rightType)) {
            holder.newAnnotation(
                HighlightSeverity.ERROR,
                "Type mismatch: expected ${leftType.name}, found ${rightType.name}"
            )
                .range(element.right)
                .withFix(CastTypeFix(element.right, leftType))
                .create()
        }
    }

    private fun isAssignable(target: MyType, source: MyType): Boolean {
        return target == source || source.isSubtypeOf(target)
    }
}
```

### Deprecated Symbol Annotator

```kotlin
class DeprecatedSymbolAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        if (element !is MyIdentifier) return

        val reference = element.reference ?: return
        val resolved = reference.resolve() ?: return

        if (resolved is MyAnnotatedElement && resolved.isDeprecated) {
            val message = resolved.deprecationMessage ?: "Deprecated"

            holder.newAnnotation(
                HighlightSeverity.WARNING,
                "'${element.text}' is deprecated: $message"
            )
                .range(element)
                .highlightType(ProblemHighlightType.LIKE_DEPRECATED)
                .create()
        }
    }
}
```

## Gutter Icons

### Adding Gutter Icons

```kotlin
class MyGutterAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        if (element !is MyFunction) return

        // Check if function overrides another
        val superMethod = findSuperMethod(element)
        if (superMethod != null) {
            holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
                .range(element.nameIdentifier ?: element)
                .gutterIconRenderer(MyGutterRenderer(element, superMethod))
                .create()
        }
    }
}

class MyGutterRenderer(
    private val method: MyFunction,
    private val superMethod: MyFunction
) : GutterIconRenderer() {
    override val icon: Icon = AllIcons.Gutter.OverridingMethod

    override val tooltipText: String = "Overrides ${superMethod.name}"

    override fun getClickAction(): AnAction {
        return object : AnAction() {
            override fun actionPerformed(e: AnActionEvent) {
                navigateToElement(superMethod)
            }
        }
    }

    override fun equals(other: Any?): Boolean {
        return other is MyGutterRenderer && method == other.method
    }

    override fun hashCode(): Int = method.hashCode()
}
```

## Text Attributes

### Custom Highlighting

```kotlin
class MyHighlightingAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        when (element) {
            is MyClass -> highlightClass(element, holder)
            is MyFunction -> highlightFunction(element, holder)
            is MyVariable -> highlightVariable(element, holder)
        }
    }

    private fun highlightClass(element: MyClass, holder: AnnotationHolder) {
        val nameIdentifier = element.nameIdentifier ?: return

        holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
            .range(nameIdentifier)
            .textAttributes(MyAttributes.CLASS_NAME)
            .create()
    }

    private fun highlightFunction(element: MyFunction, holder: AnnotationHolder) {
        val nameIdentifier = element.nameIdentifier ?: return

        holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
            .range(nameIdentifier)
            .textAttributes(MyAttributes.FUNCTION_NAME)
            .create()
    }
}
```

## Performance Considerations

### Fast Annotators

```kotlin
class MyFastAnnotator : Annotator {
    override fun annotate(element: PsiElement, holder: AnnotationHolder) {
        // Quick type check first
        if (element.node.elementType != MyTokenTypes.IDENTIFIER) return

        // Only process relevant elements
        val parent = element.parent
        if (parent !is MyReference) return

        // Minimal work in annotator
        val name = element.text
        if (name.startsWith("_")) return

        // Do the annotation
        // ...
    }
}
```

### Avoiding Expensive Operations

```kotlin
// WRONG: Expensive operation in annotator
override fun annotate(element: PsiElement, holder: AnnotationHolder) {
    val allFiles = getAllProjectFiles(element.project)  // Expensive!
    for (file in allFiles) {
        // Process each file
    }
}

// CORRECT: Use cached data
override fun annotate(element: PsiElement, holder: AnnotationHolder) {
    val cachedData = MyCacheService.getInstance(element.project).getCachedData()
    // Use cached data
}
```

## Related Topics

- [Inspections](inspections.md) - Code quality checks
- [Syntax Errors](syntax-errors.md) - Parser errors
- [Syntax Highlighting](../05-language-support/syntax-highlighting.md) - Highlighting configuration

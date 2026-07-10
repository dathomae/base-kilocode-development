# Intentions

Intentions are quick fixes or enhancements suggested by the IDE. They appear as lightbulb icons in the editor gutter and can modify code automatically.

## Overview

### What Are Intentions?

Intentions are actions that can:

- Fix code problems
- Improve code quality
- Transform code patterns
- Add missing code

### Intention Types

| Type                         | Description              |
| ---------------------------- | ------------------------ |
| Quick Fix                    | Fixes a specific problem |
| Intention Action             | General code improvement |
| Intention Action with Choice | Offers multiple options  |

## Creating an Intention

### Basic Intention

```kotlin
class MyIntention : IntentionAction {
    override fun getText(): String = "My intention"

    override fun getFamilyName(): String = "My plugin"

    override fun isAvailable(project: Project, editor: Editor?, file: PsiFile?): Boolean {
        // Check if intention is applicable
        return file != null && editor != null
    }

    override fun invoke(project: Project, editor: Editor?, file: PsiFile?) {
        // Perform the action
    }

    override fun startInWriteAction(): Boolean = true
}
```

### Registering an Intention

```xml
<extensions defaultExtensionNs="com.intellij">
    <intentionAction>
        <className>com.example.MyIntention</className>
        <category>My plugin</category>
    </intentionAction>
</extensions>
```

## Intention Context

### Checking Availability

```kotlin
override fun isAvailable(project: Project, editor: Editor?, file: PsiFile?): Boolean {
    if (editor == null || file == null) return false

    // Get element at caret
    val offset = editor.caretModel.offset
    val element = file.findElementAt(offset) ?: return false

    // Check if applicable
    return element.parent is PsiIfStatement
}
```

### Getting Context

```kotlin
override fun invoke(project: Project, editor: Editor?, file: PsiFile?) {
    val editor = editor ?: return
    val file = file ?: return

    val offset = editor.caretModel.offset
    val element = file.findElementAt(offset) ?: return

    // Work with element
    val parent = element.parent
}
```

## Common Patterns

### Wrapping Code

```kotlin
class WrapInTryCatchIntention : IntentionAction {
    override fun getText() = "Wrap in try-catch"
    override fun getFamilyName() = "Code wrapping"

    override fun isAvailable(project: Project, editor: Editor?, file: PsiFile?): Boolean {
        val element = getElementAtCaret(project, editor, file) ?: return false
        return element is PsiStatement && element !is PsiTryStatement
    }

    override fun invoke(project: Project, editor: Editor?, file: PsiFile?) {
        val statement = getElementAtCaret(project, editor, file) as? PsiStatement ?: return
        val factory = PsiElementFactory.getInstance(project)

        val tryStatement = factory.createStatementFromText(
            "try { } catch (Exception e) { }",
            null
        ) as PsiTryStatement

        tryStatement.tryBlock.add(statement.copy())
        statement.replace(tryStatement)
    }

    private fun getElementAtCaret(project: Project, editor: Editor?, file: PsiFile?): PsiElement? {
        val offset = editor?.caretModel?.offset ?: return null
        return file?.findElementAt(offset)
    }

    override fun startInWriteAction() = true
}
```

### Adding Missing Code

```kotlin
class AddMissingOverrideIntention : IntentionAction {
    override fun getText() = "Add @Override annotation"
    override fun getFamilyName() = "Java annotations"

    override fun isAvailable(project: Project, editor: Editor?, file: PsiFile?): Boolean {
        val method = getMethodAtCaret(project, editor, file) ?: return false
        return method.findSuperMethods().isNotEmpty() &&
               !method.modifierList.hasAnnotation("java.lang.Override")
    }

    override fun invoke(project: Project, editor: Editor?, file: PsiFile?) {
        val method = getMethodAtCaret(project, editor, file) ?: return
        val factory = PsiElementFactory.getInstance(project)
        val annotation = factory.createAnnotationFromText("@Override", method)
        method.modifierList?.add(annotation)
    }

    private fun getMethodAtCaret(project: Project, editor: Editor?, file: PsiFile?): PsiMethod? {
        val offset = editor?.caretModel?.offset ?: return null
        val element = file?.findElementAt(offset) ?: return null
        return PsiTreeUtil.getParentOfType(element, PsiMethod::class.java)
    }

    override fun startInWriteAction() = true
}
```

### Code Transformation

```kotlin
class ConvertToStringBuilderIntention : IntentionAction {
    override fun getText() = "Convert to StringBuilder"
    override fun getFamilyName() = "String operations"

    override fun isAvailable(project: Project, editor: Editor?, file: PsiFile?): Boolean {
        val expression = getExpressionAtCaret(project, editor, file) ?: return false
        return expression is PsiPolyadicExpression &&
               expression.operationTokenType == JavaTokenType.PLUS
    }

    override fun invoke(project: Project, editor: Editor?, file: PsiFile?) {
        val expression = getExpressionAtCaret(project, editor, file) as? PsiPolyadicExpression ?: return
        val factory = PsiElementFactory.getInstance(project)

        val operands = expression.operands
        val builder = StringBuilder("new StringBuilder()")

        for (operand in operands) {
            builder.append(".append(${operand.text})")
        }
        builder.append(".toString()")

        val newExpression = factory.createExpressionFromText(builder.toString(), expression.context)
        expression.replace(newExpression)
    }

    private fun getExpressionAtCaret(project: Project, editor: Editor?, file: PsiFile?): PsiExpression? {
        val offset = editor?.caretModel?.offset ?: return null
        val element = file?.findElementAt(offset) ?: return null
        return PsiTreeUtil.getParentOfType(element, PsiExpression::class.java)
    }

    override fun startInWriteAction() = true
}
```

## Intention with Choices

### Implementing Choice Provider

```kotlin
class MyIntentionWithChoice : IntentionAction, HighPriorityAction {
    override fun getText() = "Convert visibility"
    override fun getFamilyName() = "Visibility modifiers"

    override fun isAvailable(project: Project, editor: Editor?, file: PsiFile?): Boolean {
        val method = getMethodAtCaret(project, editor, file) ?: return false
        return method.hasModifierProperty(PsiModifier.PRIVATE) ||
               method.hasModifierProperty(PsiModifier.PROTECTED)
    }

    override fun invoke(project: Project, editor: Editor?, file: PsiFile?) {
        val method = getMethodAtCaret(project, editor, file) ?: return

        val options = listOf("public", "protected", "private")
        val chosen = Messages.showEditableChooseDialog(
            options.toTypedArray(),
            "Choose visibility",
            Messages.getQuestionIcon(),
            "public",
            null
        ) ?: return

        method.modifierList?.setModifierProperty(PsiModifier.PUBLIC, false)
        method.modifierList?.setModifierProperty(PsiModifier.PROTECTED, false)
        method.modifierList?.setModifierProperty(PsiModifier.PRIVATE, false)
        method.modifierList?.setModifierProperty(chosen, true)
    }

    private fun getMethodAtCaret(project: Project, editor: Editor?, file: PsiFile?): PsiMethod? {
        val offset = editor?.caretModel?.offset ?: return null
        val element = file?.findElementAt(offset) ?: return null
        return PsiTreeUtil.getParentOfType(element, PsiMethod::class.java)
    }

    override fun startInWriteAction() = true
}
```

## Quick Fixes

### Quick Fix for Inspection

```kotlin
class MyQuickFix : LocalQuickFix {
    override fun getFamilyName(): String = "My quick fix"

    override fun applyFix(project: Project, descriptor: ProblemDescriptor) {
        val element = descriptor.psiElement
        // Fix the problem
    }
}
```

### Using Quick Fix in Inspection

```kotlin
class MyInspection : AbstractBaseJavaLocalInspectionTool() {
    override fun checkClass(aClass: PsiClass, manager: InspectionManager, isOnTheFly: Boolean): Array<ProblemDescriptor>? {
        val problems = mutableListOf<ProblemDescriptor>()

        if (aClass.name?.startsWith("_") == true) {
            problems.add(
                manager.createProblemDescriptor(
                    aClass.nameIdentifier!!,
                    "Class name should not start with underscore",
                    MyQuickFix(),
                    ProblemHighlightType.WARNING,
                    isOnTheFly
                )
            )
        }

        return problems.toTypedArray()
    }
}
```

## Intention Priority

### Priority Interfaces

```kotlin
// High priority - shown first
class HighPriorityIntention : IntentionAction, HighPriorityAction {
    // ...
}

// Low priority - shown last
class LowPriorityIntention : IntentionAction, LowPriorityAction {
    // ...
}
```

## Best Practices

### Clear Names

```kotlin
// WRONG: Vague name
override fun getText() = "Fix"

// CORRECT: Descriptive name
override fun getText() = "Add missing @Override annotation"
```

### Precise Availability Check

```kotlin
override fun isAvailable(project: Project, editor: Editor?, file: PsiFile?): Boolean {
    // Check all conditions
    if (editor == null || file == null) return false
    if (file !is PsiJavaFile) return false

    val element = file.findElementAt(editor.caretModel.offset) ?: return false
    val method = PsiTreeUtil.getParentOfType(element, PsiMethod::class.java) ?: return false

    return method.findSuperMethods().isNotEmpty()
}
```

### Safe Modifications

```kotlin
override fun invoke(project: Project, editor: Editor?, file: PsiFile?) {
    val element = getElementAtCaret(project, editor, file) ?: return

    if (!element.isValid) return  // Check validity

    WriteAction.run<Exception> {
        // Perform modification
    }
}
```

## Related Topics

- [Inspections](../06-analysis-features/inspections.md) - Code analysis
- [Code Completion](code-completion.md) - Completion features
- [Modifying PSI](../03-psi/modifying-psi.md) - Safe modifications

# Inspections

Inspections analyze code for potential problems and code quality issues. They can provide warnings, errors, and quick fixes.

## Overview

### What Are Inspections?

Inspections are code analysis tools that:

- Detect potential bugs
- Identify code style issues
- Suggest improvements
- Provide quick fixes

### Inspection Types

| Type   | Scope        | Description               |
| ------ | ------------ | ------------------------- |
| Local  | Single file  | Fast, file-level analysis |
| Global | Project-wide | Cross-file analysis       |

## Creating a Local Inspection

### Basic Inspection

```kotlin
class MyInspection : LocalInspectionTool() {
    override fun getDisplayName(): String = "My inspection"

    override fun getGroupDisplayName(): String = "My plugin"

    override fun getShortName(): String = "MyInspection"

    override fun isEnabledByDefault(): Boolean = true

    override fun checkFile(
        file: PsiFile,
        manager: InspectionManager,
        isOnTheFly: Boolean
    ): Array<ProblemDescriptor>? {
        val problems = mutableListOf<ProblemDescriptor>()

        file.accept(object : PsiRecursiveElementVisitor() {
            override fun visitElement(element: PsiElement) {
                if (shouldReport(element)) {
                    problems.add(
                        manager.createProblemDescriptor(
                            element,
                            "Problem description",
                            MyQuickFix(),
                            ProblemHighlightType.WARNING,
                            isOnTheFly
                        )
                    )
                }
                super.visitElement(element)
            }
        })

        return problems.toTypedArray()
    }
}
```

### Registering an Inspection

```xml
<extensions defaultExtensionNs="com.intellij">
    <localInspection
        language="MyLanguage"
        shortName="MyInspection"
        displayName="My inspection"
        groupName="My plugin"
        enabledByDefault="true"
        implementationClass="com.example.MyInspection"/>
</extensions>
```

## Problem Descriptors

### Creating Problem Descriptors

```kotlin
// Basic problem
manager.createProblemDescriptor(
    element,
    "Problem description",
    isOnTheFly
)

// With quick fix
manager.createProblemDescriptor(
    element,
    "Problem description",
    MyQuickFix(),
    ProblemHighlightType.WARNING,
    isOnTheFly
)

// With multiple fixes
manager.createProblemDescriptor(
    element,
    "Problem description",
    arrayOf(Fix1(), Fix2()),
    ProblemHighlightType.ERROR,
    isOnTheFly,
    false
)
```

### Highlight Types

| Type                       | Description      |
| -------------------------- | ---------------- |
| `ERROR`                    | Error level      |
| `WARNING`                  | Warning level    |
| `GENERIC_ERROR_OR_WARNING` | Generic          |
| `LIKE_UNUSED_SYMBOL`       | Unused style     |
| `LIKE_UNKNOWN_SYMBOL`      | Unknown style    |
| `LIKE_DEPRECATED`          | Deprecated style |
| `INFORMATION`              | Information only |

## Quick Fixes

### Implementing Quick Fixes

```kotlin
class MyQuickFix : LocalQuickFix {
    override fun getFamilyName(): String = "Fix the problem"

    override fun applyFix(project: Project, descriptor: ProblemDescriptor) {
        val element = descriptor.psiElement

        WriteAction.run<Exception> {
            // Apply the fix
            element.replace(createFixedElement(element))
        }
    }
}
```

### Quick Fix with Options

```kotlin
class MyQuickFixWithOptions : LocalQuickFix {
    override fun getFamilyName(): String = "Fix with options"

    override fun applyFix(project: Project, descriptor: ProblemDescriptor) {
        val element = descriptor.psiElement

        // Show options dialog
        val options = showOptionsDialog(project) ?: return

        WriteAction.run<Exception> {
            applyWithOptions(element, options)
        }
    }

    private fun showOptionsDialog(project: Project): MyOptions? {
        // Show dialog and return options
        return null
    }
}
```

## Common Inspection Patterns

### Naming Convention Inspection

```kotlin
class NamingConventionInspection : LocalInspectionTool() {
    override fun getDisplayName() = "Naming convention"

    override fun checkFile(
        file: PsiFile,
        manager: InspectionManager,
        isOnTheFly: Boolean
    ): Array<ProblemDescriptor>? {
        val problems = mutableListOf<ProblemDescriptor>()

        file.accept(object : PsiRecursiveElementVisitor() {
            override fun visitElement(element: PsiElement) {
                when (element) {
                    is MyClass -> checkClassName(element, problems, manager, isOnTheFly)
                    is MyFunction -> checkFunctionName(element, problems, manager, isOnTheFly)
                    is MyVariable -> checkVariableName(element, problems, manager, isOnTheFly)
                }
                super.visitElement(element)
            }
        })

        return problems.toTypedArray()
    }

    private fun checkClassName(
        element: MyClass,
        problems: MutableList<ProblemDescriptor>,
        manager: InspectionManager,
        isOnTheFly: Boolean
    ) {
        val name = element.name ?: return
        if (!name.first().isUpperCase()) {
            problems.add(
                manager.createProblemDescriptor(
                    element.nameIdentifier ?: element,
                    "Class name should start with uppercase",
                    RenameFix(element),
                    ProblemHighlightType.WARNING,
                    isOnTheFly
                )
            )
        }
    }
}
```

### Unused Code Inspection

```kotlin
class UnusedCodeInspection : LocalInspectionTool() {
    override fun getDisplayName() = "Unused code"

    override fun checkFile(
        file: PsiFile,
        manager: InspectionManager,
        isOnTheFly: Boolean
    ): Array<ProblemDescriptor>? {
        val problems = mutableListOf<ProblemDescriptor>()

        file.accept(object : PsiRecursiveElementVisitor() {
            override fun visitElement(element: PsiElement) {
                if (element is MyNamedElement && !isUsed(element)) {
                    problems.add(
                        manager.createProblemDescriptor(
                            element.nameIdentifier ?: element,
                            "${element.name} is never used",
                            arrayOf(RemoveFix(element), CommentOutFix(element)),
                            ProblemHighlightType.LIKE_UNUSED_SYMBOL,
                            isOnTheFly
                        )
                    )
                }
                super.visitElement(element)
            }
        })

        return problems.toTypedArray()
    }

    private fun isUsed(element: MyNamedElement): Boolean {
        return ReferencesSearch.search(
            element,
            GlobalSearchScope.fileScope(element.containingFile)
        ).findFirst() != null
    }
}
```

### Code Style Inspection

```kotlin
class CodeStyleInspection : LocalInspectionTool() {
    override fun getDisplayName() = "Code style issues"

    override fun checkFile(
        file: PsiFile,
        manager: InspectionManager,
        isOnTheFly: Boolean
    ): Array<ProblemDescriptor>? {
        val problems = mutableListOf<ProblemDescriptor>()

        file.accept(object : PsiRecursiveElementVisitor() {
            override fun visitElement(element: PsiElement) {
                when (element) {
                    is MyIfStatement -> checkIfStatement(element, problems, manager, isOnTheFly)
                    is MyBlock -> checkBlock(element, problems, manager, isOnTheFly)
                }
                super.visitElement(element)
            }
        })

        return problems.toTypedArray()
    }

    private fun checkIfStatement(
        element: MyIfStatement,
        problems: MutableList<ProblemDescriptor>,
        manager: InspectionManager,
        isOnTheFly: Boolean
    ) {
        // Check for missing braces
        val thenBlock = element.thenBlock
        if (thenBlock !is MyBlock) {
            problems.add(
                manager.createProblemDescriptor(
                    thenBlock,
                    "Missing braces around if body",
                    AddBracesFix(element),
                    ProblemHighlightType.WARNING,
                    isOnTheFly
                )
            )
        }
    }
}
```

## Inspection Options

### Configurable Inspection

```kotlin
class MyConfigurableInspection : LocalInspectionTool() {
    @JvmField var reportWarnings: Boolean = true
    @JvmField var maxLineLength: Int = 120

    override fun getDisplayName() = "Configurable inspection"

    override fun createOptionsPanel(): JComponent? {
        val panel = JPanel()
        panel.layout = BorderLayout()

        val warningCheckbox = JCheckBox("Report warnings", reportWarnings)
        warningCheckbox.addActionListener { reportWarnings = warningCheckbox.isSelected }
        panel.add(warningCheckbox, BorderLayout.NORTH)

        val lengthField = JTextField(maxLineLength.toString())
        lengthField.addActionListener {
            maxLineLength = lengthField.text.toIntOrNull() ?: 120
        }
        panel.add(lengthField, BorderLayout.CENTER)

        return panel
    }
}
```

## Inspection Severity

### Setting Severity

```kotlin
override fun checkFile(
    file: PsiFile,
    manager: InspectionManager,
    isOnTheFly: Boolean
): Array<ProblemDescriptor>? {
    val severity = if (isCritical(element)) {
        ProblemHighlightType.ERROR
    } else {
        ProblemHighlightType.WARNING
    }

    problems.add(
        manager.createProblemDescriptor(
            element,
            "Problem description",
            MyQuickFix(),
            severity,
            isOnTheFly
        )
    )
}
```

## Related Topics

- [Global Inspections](global-inspections.md) - Project-wide analysis
- [Annotators](annotators.md) - Semantic highlighting
- [Inspection Profiles](inspection-profiles.md) - Managing inspection settings

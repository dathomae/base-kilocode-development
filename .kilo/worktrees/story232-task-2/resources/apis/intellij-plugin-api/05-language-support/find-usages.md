# Find Usages

Find Usages allows users to locate all references to a symbol. Implementing this feature requires defining usage providers and handlers.

## Overview

### Usage Search Components

| Component             | Purpose                         |
| --------------------- | ------------------------------- |
| `UsageTargetProvider` | Identifies what can be searched |
| `FindUsagesProvider`  | Describes elements for search   |
| `UsageTypeProvider`   | Categorizes usage types         |

### Usage Search Process

1. User invokes Find Usages on an element
2. Platform collects usage targets
3. Search runs across project scope
4. Results are displayed in Usage view

## Implementing FindUsagesProvider

### Basic Provider

```kotlin
class MyFindUsagesProvider : FindUsagesProvider() {
    override fun canFindUsagesFor(element: PsiElement): Boolean {
        return element is MyNamedElement
    }

    override fun getHelpId(element: PsiElement): String? = null

    override fun getType(element: PsiElement): String {
        return when (element) {
            is MyClass -> "class"
            is MyFunction -> "function"
            is MyVariable -> "variable"
            is MyField -> "field"
            else -> "element"
        }
    }

    override fun getDescriptiveName(element: PsiElement): String {
        return when (element) {
            is MyClass -> element.name ?: "anonymous class"
            is MyFunction -> "${element.name}()"
            is MyVariable -> element.name ?: "variable"
            is MyField -> element.name ?: "field"
            else -> "unknown"
        }
    }

    override fun getNodeText(element: PsiElement, useFullName: Boolean): String {
        return when (element) {
            is MyClass -> element.name ?: "class"
            is MyFunction -> element.name ?: "function"
            is MyVariable -> element.name ?: "variable"
            else -> element.text
        }
    }
}
```

### Registering the Provider

```xml
<extensions defaultExtensionNs="com.intellij">
    <lang.findUsagesProvider
        language="MyLanguage"
        implementationClass="com.example.MyFindUsagesProvider"/>
</extensions>
```

## Usage Types

### Custom Usage Types

```kotlin
class MyUsageTypeProvider : UsageTypeProvider {
    override fun getUsageType(element: PsiElement): UsageType? {
        return when (element) {
            is MyClass -> CLASS_USAGE
            is MyFunction -> FUNCTION_USAGE
            is MyVariable -> VARIABLE_USAGE
            is MyField -> FIELD_USAGE
            else -> null
        }
    }

    companion object {
        val CLASS_USAGE = UsageType("Class usage")
        val FUNCTION_USAGE = UsageType("Function usage")
        val VARIABLE_USAGE = UsageType("Variable usage")
        val FIELD_USAGE = UsageType("Field usage")
    }
}
```

### Registering Usage Types

```xml
<extensions defaultExtensionNs="com.intellij">
    <usageTypeProvider
        implementation="com.example.MyUsageTypeProvider"/>
</extensions>
```

## Usage Target Provider

### Identifying Search Targets

```kotlin
class MyUsageTargetProvider : UsageTargetProvider {
    override fun getTargets(editor: Editor, file: PsiFile): Array<PsiElementUsageTarget>? {
        val offset = editor.caretModel.offset
        val element = file.findElementAt(offset) ?: return null

        val target = PsiTreeUtil.getParentOfType(
            element,
            MyClass::class.java,
            MyFunction::class.java,
            MyVariable::class.java
        ) ?: return null

        return arrayOf(PsiElement2UsageTargetAdapter(target))
    }

    override fun getTargets(psiElement: PsiElement): Array<PsiElementUsageTarget>? {
        val target = PsiTreeUtil.getParentOfType(
            psiElement,
            MyClass::class.java,
            MyFunction::class.java,
            MyVariable::class.java
        ) ?: return null

        return arrayOf(PsiElement2UsageTargetAdapter(target))
    }
}
```

## Customizing Search

### Find Usages Handler

```kotlin
class MyFindUsagesHandler(
    element: PsiElement,
    factory: FindUsagesHandlerFactory
) : FindUsagesHandler(element, factory) {

    override fun processElementUsages(
        element: PsiElement,
        processor: Processor<in UsageInfo>,
        options: FindUsagesOptions
    ): Boolean {
        // Custom search logic
        val references = ReferencesSearch.search(
            element,
            options.searchScope
        )

        for (reference in references) {
            val usageInfo = UsageInfo(reference.element)
            if (!processor.process(usageInfo)) {
                return false
            }
        }

        return true
    }

    override fun getPrimaryElements(): Array<PsiElement> {
        return arrayOf(element)
    }

    override fun getSecondaryElements(): Array<PsiElement> {
        // Related elements to search
        return emptyArray()
    }
}
```

### Handler Factory

```kotlin
class MyFindUsagesHandlerFactory : FindUsagesHandlerFactory() {
    override fun canFindUsages(element: PsiElement): Boolean {
        return element is MyNamedElement
    }

    override fun createFindUsagesHandler(element: PsiElement, forHighlightUsages: Boolean): FindUsagesHandler {
        return MyFindUsagesHandler(element, this)
    }
}
```

### Registering the Handler

```xml
<extensions defaultExtensionNs="com.intellij">
    <findUsagesHandlerFactory
        implementation="com.example.MyFindUsagesHandlerFactory"/>
</extensions>
```

## Usage Info

### Custom Usage Info

```kotlin
class MyUsageInfo(
    element: PsiElement,
    val usageType: UsageType
) : UsageInfo(element) {
    override fun getNavigationOffset(): Int {
        return element?.textRange?.startOffset ?: -1
    }
}
```

### Usage Info with Range

```kotlin
class MyRangeUsageInfo(
    element: PsiElement,
    range: TextRange
) : UsageInfo(element, range.startOffset, range.endOffset) {
    // Usage with specific range within element
}
```

## Search Options

### Custom Find Usages Options

```kotlin
class MyFindUsagesOptions(project: Project) : FindUsagesOptions(project) {
    var searchInComments: Boolean = false
    var searchInStrings: Boolean = false
    var includeOverloads: Boolean = false
}
```

### Options Panel

```kotlin
class MyFindUsagesDialog(
    project: Project,
    element: PsiElement,
    options: MyFindUsagesOptions
) : FindUsagesDialog(project, element, options) {

    override fun createFindUsagesOptionsPanel(): FindUsagesOptionsPanel {
        return object : FindUsagesOptionsPanel() {
            private val searchInCommentsCheckbox = JCheckBox("Search in comments")
            private val searchInStringsCheckbox = JCheckBox("Search in string literals")

            init {
                add(searchInCommentsCheckbox)
                add(searchInStringsCheckbox)
            }

            override fun getFindUsagesOptions(): FindUsagesOptions {
                val opts = options as MyFindUsagesOptions
                opts.searchInComments = searchInCommentsCheckbox.isSelected
                opts.searchInStrings = searchInStringsCheckbox.isSelected
                return opts
            }
        }
    }
}
```

## Programmatic Usage Search

### Finding Usages Programmatically

```kotlin
fun findUsages(element: PsiElement, project: Project): List<UsageInfo> {
    val usages = mutableListOf<UsageInfo>()
    val options = FindUsagesOptions(project)

    FindUsagesManager.getInstance(project).findUsages(
        element,
        options,
        object : FindUsagesHandler() {
            override fun processElementUsages(
                element: PsiElement,
                processor: Processor<in UsageInfo>,
                options: FindUsagesOptions
            ): Boolean {
                ReferencesSearch.search(element, options.searchScope).forEach { ref ->
                    usages.add(UsageInfo(ref.element))
                }
                return true
            }
        },
        false
    )

    return usages
}
```

### Checking if Element Is Used

```kotlin
fun isElementUsed(element: PsiNamedElement, project: Project): Boolean {
    return ReferencesSearch.search(
        element,
        GlobalSearchScope.projectScope(project)
    ).findFirst() != null
}
```

## Related Topics

- [Reference Resolution](reference-resolution.md) - Resolving references
- [Rename Refactoring](rename-refactoring.md) - Rename operations
- [PSI References](../03-psi/psi-references.md) - Reference architecture

# Reference Resolution

Reference resolution connects usages of symbols to their definitions. This enables navigation, refactoring, and code analysis features.

## Overview

### What Is Reference Resolution?

Reference resolution maps a usage (like a variable name) to its definition. This enables:

- Go to definition (Ctrl+Click)
- Find usages
- Rename refactoring
- Code analysis

### Resolution Process

1. Identify reference elements in PSI
2. Implement `PsiReference` for each reference
3. Resolve to target element
4. Provide completion variants

## Creating References

### Reference Provider

```kotlin
class MyReferenceContributor : PsiReferenceContributor() {
    override fun registerReferenceProviders(registrar: PsiReferenceRegistrar) {
        registrar.registerReferenceProvider(
            PlatformPatterns.psiElement(MyIdentifier::class.java)
                .withLanguage(MyLanguage.INSTANCE),
            MyReferenceProvider()
        )
    }
}

class MyReferenceProvider : PsiReferenceProvider() {
    override fun getReferencesByElement(
        element: PsiElement,
        context: ProcessingContext
    ): Array<PsiReference> {
        return arrayOf(MyReference(element as MyIdentifier))
    }
}
```

### Registering References

```xml
<extensions defaultExtensionNs="com.intellij">
    <psi.referenceContributor
        language="MyLanguage"
        implementation="com.example.MyReferenceContributor"/>
</extensions>
```

## Implementing PsiReference

### Basic Reference

```kotlin
class MyReference(element: MyIdentifier) : PsiReferenceBase<MyIdentifier>(element) {
    override fun resolve(): PsiElement? {
        val name = element.text
        val scope = getResolutionScope(element)

        // Search in local scope
        scope.getLocalVariables().find { it.name == name }?.let { return it }
        scope.getFunctions().find { it.name == name }?.let { return it }

        // Search in class scope
        val containingClass = PsiTreeUtil.getParentOfType(element, MyClass::class.java)
        containingClass?.fields?.find { it.name == name }?.let { return it }

        return null
    }

    override fun getVariants(): Array<Any> {
        val scope = getResolutionScope(element)
        val variants = mutableListOf<LookupElement>()

        // Add local variables
        for (variable in scope.getLocalVariables()) {
            variants.add(createLookupElement(variable))
        }

        // Add functions
        for (function in scope.getFunctions()) {
            variants.add(createLookupElement(function))
        }

        return variants.toTypedArray()
    }

    private fun getResolutionScope(element: PsiElement): MyScope {
        // Walk up the tree to find enclosing scope
        var current: PsiElement? = element
        while (current != null) {
            if (current is MyFunction || current is MyClass) {
                return MyScope(current)
            }
            current = current.parent
        }
        return MyScope(element.containingFile)
    }
}
```

### Poly-Variant Reference

```kotlin
class MyPolyReference(element: MyIdentifier) : PsiPolyVariantReferenceBase<MyIdentifier>(element) {
    override fun multiResolve(incompleteCode: Boolean): Array<ResolveResult> {
        val name = element.text
        val results = mutableListOf<ResolveResult>()

        // Find all possible targets
        val targets = findAllTargets(name)

        for (target in targets) {
            results.add(PsiElementResolveResult(target))
        }

        return results.toTypedArray()
    }

    override fun resolve(): PsiElement? {
        val results = multiResolve(false)
        return results.firstOrNull()?.element
    }

    private fun findAllTargets(name: String): List<PsiElement> {
        val targets = mutableListOf<PsiElement>()

        // Search in project
        val scope = GlobalSearchScope.projectScope(element.project)
        val files = FilenameIndex.getVirtualFilesByName(
            element.project,
            name + ".myf",
            scope
        )

        for (file in files) {
            val psiFile = PsiManager.getInstance(element.project).findFile(file)
            psiFile?.accept(object : PsiRecursiveElementVisitor() {
                override fun visitElement(element: PsiElement) {
                    if (element is MyNamedElement && element.name == name) {
                        targets.add(element)
                    }
                    super.visitElement(element)
                }
            })
        }

        return targets
    }
}
```

## Scope-Based Resolution

### Scope Implementation

```kotlin
class MyScope(private val element: PsiElement) {
    fun getLocalVariables(): List<MyVariable> {
        val variables = mutableListOf<MyVariable>()

        when (element) {
            is MyFunction -> {
                // Parameters
                variables.addAll(element.parameters)
                // Local variables in body
                element.body?.accept(object : PsiRecursiveElementVisitor() {
                    override fun visitElement(el: PsiElement) {
                        if (el is MyVariable) {
                            variables.add(el)
                        }
                        super.visitElement(el)
                    }
                })
            }
            is MyClass -> {
                // Fields
                variables.addAll(element.fields)
            }
        }

        return variables
    }

    fun getFunctions(): List<MyFunction> {
        val functions = mutableListOf<MyFunction>()

        when (element) {
            is MyClass -> {
                functions.addAll(element.methods)
            }
            is MyFile -> {
                element.accept(object : PsiRecursiveElementVisitor() {
                    override fun visitElement(el: PsiElement) {
                        if (el is MyFunction) {
                            functions.add(el)
                        }
                        super.visitElement(el)
                    }
                })
            }
        }

        return functions
    }
}
```

### Nested Scope Resolution

```kotlin
class MyNestedScopeReference(element: MyIdentifier) : PsiReferenceBase<MyIdentifier>(element) {
    override fun resolve(): PsiElement? {
        // Start from innermost scope
        var current: PsiElement? = element.parent

        while (current != null) {
            // Check current scope
            when (current) {
                is MyBlock -> {
                    val variable = findVariableInBlock(current, element.text)
                    if (variable != null) return variable
                }
                is MyFunction -> {
                    val param = current.parameters.find { it.name == element.text }
                    if (param != null) return param
                }
                is MyClass -> {
                    val field = current.fields.find { it.name == element.text }
                    if (field != null) return field

                    val method = current.methods.find { it.name == element.text }
                    if (method != null) return method
                }
            }

            current = current.parent
        }

        return null
    }

    private fun findVariableInBlock(block: MyBlock, name: String): MyVariable? {
        for (statement in block.statements) {
            if (statement is MyVariableDeclaration && statement.name == name) {
                return statement.variable
            }
        }
        return null
    }
}
```

## Qualified References

### Dot Expression Resolution

```kotlin
class MyDotReference(element: MyIdentifier) : PsiReferenceBase<MyIdentifier>(element) {
    override fun resolve(): PsiElement? {
        val dotExpr = PsiTreeUtil.getParentOfType(element, MyDotExpression::class.java)
            ?: return null

        val qualifier = dotExpr.qualifier ?: return null
        val qualifierType = resolveType(qualifier) ?: return null

        // Find member in type
        return findMember(qualifierType, element.text)
    }

    private fun resolveType(expression: MyExpression): MyType? {
        return when (expression) {
            is MyIdentifier -> {
                val resolved = expression.reference?.resolve()
                when (resolved) {
                    is MyVariable -> resolved.type
                    is MyField -> resolved.type
                    else -> null
                }
            }
            is MyCallExpression -> {
                val function = expression.reference?.resolve() as? MyFunction
                function?.returnType
            }
            is MyDotExpression -> {
                val member = expression.reference?.resolve()
                when (member) {
                    is MyField -> member.type
                    is MyFunction -> member.returnType
                    else -> null
                }
            }
            else -> null
        }
    }

    private fun findMember(type: MyType, name: String): PsiElement? {
        val classDef = findClassDefinition(type.name, element.project) ?: return null

        // Search fields
        classDef.fields.find { it.name == name }?.let { return it }

        // Search methods
        classDef.methods.find { it.name == name }?.let { return it }

        return null
    }
}
```

## Import Resolution

### Resolving Imported Symbols

```kotlin
class MyImportReference(element: MyIdentifier) : PsiReferenceBase<MyIdentifier>(element) {
    override fun resolve(): PsiElement? {
        val file = element.containingFile as? MyFile ?: return null
        val name = element.text

        // Check imports
        for (import in file.imports) {
            if (import.importedName == name) {
                return resolveImport(import)
            }
        }

        // Check default imports
        val defaultImports = getDefaultImports(element.project)
        for (imported in defaultImports) {
            if (imported.name == name) {
                return imported
            }
        }

        return null
    }

    private fun resolveImport(import: MyImport): PsiElement? {
        val importedPath = import.importedPath
        val parts = importedPath.split(".")

        // Find file by path
        val scope = GlobalSearchScope.projectScope(element.project)
        val fileName = parts.last() + ".myf"
        val files = FilenameIndex.getVirtualFilesByName(element.project, fileName, scope)

        for (file in files) {
            val psiFile = PsiManager.getInstance(element.project).findFile(file) as? MyFile
            val exported = psiFile?.exports?.find { it.name == import.importedName }
            if (exported != null) return exported
        }

        return null
    }
}
```

## Performance Optimization

### Caching Resolution Results

```kotlin
class MyCachedReference(element: MyIdentifier) : PsiReferenceBase<MyIdentifier>(element) {
    private val cachedResult: PsiElement? by lazy {
        resolveInternal()
    }

    override fun resolve(): PsiElement? = cachedResult

    private fun resolveInternal(): PsiElement? {
        // Expensive resolution logic
        return findTarget(element.text)
    }
}
```

### Using CachedValuesManager

```kotlin
class MyOptimizedReference(element: MyIdentifier) : PsiReferenceBase<MyIdentifier>(element) {
    override fun resolve(): PsiElement? {
        return CachedValuesManager.getCachedValue(element) {
            val result = resolveInternal()
            CachedValueProvider.Result.create(
                result,
                PsiModificationTracker.MODIFICATION_COUNT
            )
        }
    }
}
```

## Related Topics

- [PSI References](../03-psi/psi-references.md) - Reference architecture
- [Code Completion](code-completion-lang.md) - Completion integration
- [Find Usages](find-usages.md) - Usage search

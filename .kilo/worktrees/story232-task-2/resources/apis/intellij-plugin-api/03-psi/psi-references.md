# PSI References

References connect usages of symbols to their definitions. Understanding references is essential for implementing navigation, refactoring, and code analysis features.

## Overview

### What Is a Reference?

A reference represents a usage of a symbol that can be resolved to its definition. For example:

```java
MyClass myObject = new MyClass();  // "MyClass" references the class definition
myObject.myMethod();               // "myMethod" references the method definition
```

### The PsiReference Interface

```kotlin
interface PsiReference {
    val element: PsiElement
    val rangeInElement: TextRange
    val canonicalText: String

    fun resolve(): PsiElement?
    fun isReferenceTo(element: PsiElement): Boolean
    fun handleElementRename(newElementName: String): PsiElement
    fun bindToElement(element: PsiElement): PsiElement
}
```

## Getting References

### From an Element

```kotlin
// Get all references from an element
val references = element.references

// Get single reference (most common case)
val reference = element.reference

// Get reference of specific type
val ref = element.getReference(PsiReference::class.java)
```

### From an Expression

```kotlin
val expression: PsiReferenceExpression = ...
val reference = expression.reference

// Resolve to the target
val target = reference?.resolve()
```

## Resolving References

### Basic Resolution

```kotlin
val reference = element.reference
val resolved = reference?.resolve()

if (resolved != null) {
    println("Resolved to: ${resolved.text}")
} else {
    println("Could not resolve reference")
}
```

### Resolution with Type Checking

```kotlin
val reference = element.reference
val resolved = reference?.resolve()

when (resolved) {
    is PsiClass -> println("Class: ${resolved.name}")
    is PsiMethod -> println("Method: ${resolved.name}")
    is PsiField -> println("Field: ${resolved.name}")
    is PsiVariable -> println("Variable: ${resolved.name}")
    null -> println("Unresolved")
}
```

### Poly-References

Some references can resolve to multiple targets:

```kotlin
val polyReference = element.reference as? PsiPolyVariantReference
val results = polyReference?.multiResolve(false)

results?.forEach { resolveResult ->
    val element = resolveResult.element
    println("Candidate: ${element?.text}")
}
```

## Creating References

### Reference Provider

Create references using a `PsiReferenceProvider`:

```kotlin
class MyReferenceProvider : PsiReferenceProvider() {
    override fun getReferencesByElement(
        element: PsiElement,
        context: ProcessingContext
    ): Array<PsiReference> {
        return arrayOf(MyReference(element))
    }
}
```

### Reference Contributor

Register references using a `PsiReferenceContributor`:

```kotlin
class MyReferenceContributor : PsiReferenceContributor() {
    override fun registerReferenceProviders(registrar: PsiReferenceRegistrar) {
        registrar.registerReferenceProvider(
            PlatformPatterns.psiElement(PsiLiteralExpression::class.java),
            MyReferenceProvider()
        )
    }
}
```

### Register the Contributor

```xml
<extensions defaultExtensionNs="com.intellij">
    <psi.referenceContributor
        language="JAVA"
        implementation="com.example.MyReferenceContributor"/>
</extensions>
```

## Custom Reference Implementation

### Basic Reference

```kotlin
class MyReference(element: PsiElement) : PsiReferenceBase<PsiElement>(element) {
    override fun resolve(): PsiElement? {
        // Find the target element
        val name = element.text
        return findTargetElement(name)
    }

    override fun getVariants(): Array<Any> {
        // Provide completion variants
        return getAllPossibleTargets().map { it.name }.toTypedArray()
    }

    private fun findTargetElement(name: String): PsiElement? {
        // Implementation to find target
        return null
    }
}
```

### Poly-Variant Reference

```kotlin
class MyPolyReference(element: PsiElement) : PsiPolyVariantReferenceBase<PsiElement>(element) {
    override fun multiResolve(incompleteCode: Boolean): Array<ResolveResult> {
        val targets = findTargetElements(element.text)
        return targets.map { PsiElementResolveResult(it) }.toTypedArray()
    }

    override fun resolve(): PsiElement? {
        val results = multiResolve(false)
        return results.firstOrNull()?.element
    }
}
```

## Reference Types

### Java References

| Type                          | Description                                 |
| ----------------------------- | ------------------------------------------- |
| `PsiReferenceExpression`      | Expression referencing a variable or method |
| `PsiJavaCodeReferenceElement` | Reference to a class or package             |
| `PsiImportStatement`          | Import statement reference                  |
| `PsiIdentifier`               | Identifier within a reference               |

### Resolving Java References

```kotlin
// Resolve a reference expression
val refExpr: PsiReferenceExpression = ...
val target = refExpr.resolve()

// Resolve a class reference
val classRef: PsiJavaCodeReferenceElement = ...
val resolvedClass = classRef.resolve()

// Check if reference is to a specific element
if (refExpr.isReferenceTo(targetElement)) {
    // Reference points to targetElement
}
```

## Finding References

### ReferencesSearch

```kotlin
// Find all references to an element
val references = ReferencesSearch.search(
    psiElement,
    GlobalSearchScope.projectScope(project)
).findAll()

for (reference in references) {
    println("Reference in: ${reference.element.containingFile.name}")
}
```

### Usage Search

```kotlin
// Find usages of an element
val usages = UsageInfoSearcher(
    psiElement,
    GlobalSearchScope.projectScope(project)
).search()

usages.forEach { usage ->
    println("Usage at: ${usage.file.name}:${usage.navigationOffset}")
}
```

## Reference Manipulation

### Renaming

```kotlin
// Handle element rename
val reference = element.reference
reference?.handleElementRename("newName")
```

### Binding

```kotlin
// Bind reference to a different element
val reference = element.reference
reference?.bindToElement(newTargetElement)
```

## Reference Patterns

### Using Patterns for Registration

```kotlin
override fun registerReferenceProviders(registrar: PsiReferenceRegistrar) {
    // Match specific element types
    registrar.registerReferenceProvider(
        PlatformPatterns.psiElement(PsiLiteralExpression::class.java),
        MyReferenceProvider()
    )

    // Match with conditions
    registrar.registerReferenceProvider(
        PlatformPatterns.psiElement(PsiLiteralExpression::class.java)
            .withParent(PsiMethodCallExpression::class.java),
        MyMethodReferenceProvider()
    )

    // Match with language
    registrar.registerReferenceProvider(
        PlatformPatterns.psiElement()
            .withLanguage(JavaLanguage.INSTANCE),
        MyJavaReferenceProvider()
    )
}
```

## Common Patterns

### Finding All References to a Class

```kotlin
fun findClassReferences(psiClass: PsiClass, project: Project): List<PsiReference> {
    return ReferencesSearch.search(
        psiClass,
        GlobalSearchScope.projectScope(project)
    ).findAll().toList()
}
```

### Finding Method Call Sites

```kotlin
fun findMethodCallSites(method: PsiMethod, project: Project): List<PsiElement> {
    return ReferencesSearch.search(method, GlobalSearchScope.projectScope(project))
        .mapping { it.element }
        .toList()
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

### Resolving Qualified References

```kotlin
fun resolveQualifiedReference(expression: PsiQualifiedReference): PsiElement? {
    val qualifier = expression.qualifier
    val qualifierTarget = qualifier?.reference?.resolve()

    if (qualifierTarget is PsiClass) {
        // Resolve member within class
        return qualifierTarget.findFieldByName(expression.referenceName, false)
            ?: qualifierTarget.findMethodsByName(expression.referenceName, false).firstOrNull()
    }

    return null
}
```

## Performance Considerations

### Lazy Resolution

References should resolve lazily:

```kotlin
class MyReference(element: PsiElement) : PsiReferenceBase<PsiElement>(element) {
    private val resolvedElement by lazy {
        // Expensive resolution happens only when needed
        performResolution()
    }

    override fun resolve(): PsiElement? = resolvedElement
}
```

### Caching Results

```kotlin
class MyReference(element: PsiElement) : PsiReferenceBase<PsiElement>(element) {
    private var cachedResult: PsiElement? = null
    private var cached = false

    override fun resolve(): PsiElement? {
        if (!cached) {
            cachedResult = performResolution()
            cached = true
        }
        return cachedResult
    }
}
```

## Related Topics

- [Navigating PSI](navigating-psi.md) - Traversal patterns
- [PSI Elements](psi-elements.md) - Understanding element types
- [Find Usages](../05-language-support/find-usages.md) - Usage search features
- [Rename Refactoring](../05-language-support/rename-refactoring.md) - Rename operations

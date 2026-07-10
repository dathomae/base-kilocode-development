# Navigating PSI

Navigating the PSI tree is a fundamental skill for plugin development. This guide covers patterns for traversing, searching, and finding elements in PSI trees.

## Tree Traversal

### Parent Navigation

```kotlin
// Get immediate parent
val parent = element.parent

// Find parent of specific type
val containingClass = PsiTreeUtil.getParentOfType(element, PsiClass::class.java)

// Find parent matching condition
val containingMethod = PsiTreeUtil.findFirstParent(element) {
    it is PsiMethod
}
```

### Child Navigation

```kotlin
// Get all direct children
val children = element.children

// Get first/last child
val firstChild = element.firstChild
val lastChild = element.lastChild

// Find child of specific type
val method = PsiTreeUtil.findChildOfType(element, PsiMethod::class.java)

// Find all children of type
val allMethods = PsiTreeUtil.findChildrenOfType(element, PsiMethod::class.java)
```

### Sibling Navigation

```kotlin
// Get siblings
val prevSibling = element.prevSibling
val nextSibling = element.nextSibling

// Skip whitespace siblings
val prevCodeSibling = PsiTreeUtil.getPrevSiblingOfType(element, PsiStatement::class.java)
val nextCodeSibling = PsiTreeUtil.getNextSiblingOfType(element, PsiStatement::class.java)
```

## Using PsiTreeUtil

### Finding Elements

```kotlin
// Find single element
val method = PsiTreeUtil.findChildOfType(psiFile, PsiMethod::class.java)

// Find all elements of type
val allClasses = PsiTreeUtil.findChildrenOfType(psiFile, PsiClass::class.java)

// Find element in parent chain
val parentClass = PsiTreeUtil.getParentOfType(element, PsiClass::class.java)

// Find element at offset
val elementAtOffset = PsiTreeUtil.findElementOfClassAtOffset(
    psiFile,
    offset,
    PsiIdentifier::class.java,
    false
)
```

### Collecting Elements

```kotlin
// Collect all elements of type
val methods = mutableListOf<PsiMethod>()
PsiTreeUtil.processElements(psiFile) { element ->
    if (element is PsiMethod) {
        methods.add(element)
    }
    true  // Continue processing
}
```

### Finding Common Parent

```kotlin
val element1: PsiElement = ...
val element2: PsiElement = ...
val commonParent = PsiTreeUtil.findCommonParent(element1, element2)
```

## Searching for Elements

### By Name

```kotlin
// Find class by qualified name
val psiClass = JavaPsiFacade.getInstance(project)
    .findClass("com.example.MyClass", GlobalSearchScope.projectScope(project))

// Find method by name in class
val method = psiClass?.findMethodsByName("myMethod", false)?.firstOrNull()

// Find field by name
val field = psiClass?.findFieldByName("myField", false)
```

### By Pattern

```kotlin
// Find elements matching pattern
val results = mutableListOf<PsiElement>()
psiFile.accept(object : PsiRecursiveElementVisitor() {
    override fun visitElement(element: PsiElement) {
        if (element.text.contains("TODO")) {
            results.add(element)
        }
        super.visitElement(element)
    }
})
```

### Using References

```kotlin
// Find all references to an element
val references = ReferencesSearch.search(psiElement, GlobalSearchScope.projectScope(project))
    .findAll()

for (reference in references) {
    println("Reference at: ${reference.element.textRange}")
}
```

## Scope-Based Search

### Search Scopes

```kotlin
// Project scope
val projectScope = GlobalSearchScope.projectScope(project)

// Module scope
val moduleScope = GlobalSearchScope.moduleScope(module)

// Directory scope
val dirScope = GlobalSearchScope.directoryScope(directory, true)

// File scope
val fileScope = GlobalSearchScope.fileScope(psiFile)

// Library scope
val libScope = GlobalSearchScope.allScope(project)
```

### Combining Scopes

```kotlin
// Union
val combined = scope1.union(scope2)

// Intersection
val intersection = scope1.intersectWith(scope2)

// Exclusion
val excluded = scope1.intersectWith(scope2.not)
```

### Searching Within Scopes

```kotlin
// Find classes in scope
val classes = JavaPsiFacade.getInstance(project)
    .findClasses("MyClass", searchScope)

// Search for references in scope
val references = ReferencesSearch.search(element, searchScope)
```

## Element Visitors

### Recursive Visitor

```kotlin
psiFile.accept(object : PsiRecursiveElementVisitor() {
    override fun visitElement(element: PsiElement) {
        // Process element
        super.visitElement(element)  // Continue recursion
    }
})
```

### Java-Specific Visitor

```kotlin
psiFile.accept(object : JavaRecursiveElementVisitor() {
    override fun visitClass(aClass: PsiClass) {
        println("Class: ${aClass.name}")
        super.visitClass(aClass)
    }

    override fun visitMethod(method: PsiMethod) {
        println("Method: ${method.name}")
        super.visitMethod(method)
    }

    override fun visitReferenceExpression(expression: PsiReferenceExpression) {
        println("Reference: ${expression.text}")
        super.visitReferenceExpression(expression)
    }
})
```

### Stopping Recursion

```kotlin
psiFile.accept(object : PsiRecursiveElementVisitor() {
    override fun visitElement(element: PsiElement) {
        if (shouldSkip(element)) {
            return  // Don't visit children
        }
        // Process element
        super.visitElement(element)
    }
})
```

## Query System

### Using Queries

```kotlin
// Search for references
val query = ReferencesSearch.search(psiElement, searchScope)
query.forEach { reference ->
    // Process reference
}

// Search for implementations
val impls = ClassInheritorsSearch.search(psiClass, searchScope, true)
impls.forEach { implementation ->
    // Process implementation
}

// Search for usages
val usages = QueryFactory.getInstance().search(UsageSearchParameters)
```

### Custom Queries

```kotlin
// Create a custom query
val query = object : Query<PsiElement>() {
    override fun forEach(consumer: Processor<PsiElement>): Boolean {
        // Custom search logic
        return true
    }
}
```

## Navigation Utilities

### Getting Offset Information

```kotlin
// Get element at offset
val element = psiFile.findElementAt(offset)

// Get element range
val range = element.textRange

// Check if offset is within element
if (range?.contains(offset) == true) {
    // Offset is within element
}
```

### Getting Line Information

```kotlin
val document = PsiDocumentManager.getInstance(project).getDocument(psiFile)
val lineNumber = document?.getLineNumber(element.textRange?.startOffset ?: 0)
val columnNumber = element.textRange?.startOffset?.let { offset ->
    offset - document?.getLineStartOffset(lineNumber ?: 0)!!
}
```

### Navigation to Element

```kotlin
// Open element in editor
val descriptor = OpenFileDescriptor(
    project,
    element.containingFile.virtualFile!!,
    element.textRange?.startOffset ?: 0
)
FileEditorManager.getInstance(project).openTextEditor(descriptor, true)
```

## Common Patterns

### Finding All Methods in a Class

```kotlin
fun getAllMethods(psiClass: PsiClass): List<PsiMethod> {
    val methods = mutableListOf<PsiMethod>()

    // Include inherited methods
    val allMethods = psiClass.allMethods

    // Or just declared methods
    val declaredMethods = psiClass.methods

    return methods
}
```

### Finding All Implementations

```kotlin
fun findImplementations(psiClass: PsiClass, project: Project): List<PsiClass> {
    return ClassInheritorsSearch.search(
        psiClass,
        GlobalSearchScope.projectScope(project),
        true
    ).findAll().toList()
}
```

### Finding References to a Method

```kotlin
fun findMethodUsages(method: PsiMethod, project: Project): List<PsiReference> {
    return ReferencesSearch.search(
        method,
        GlobalSearchScope.projectScope(project)
    ).findAll().toList()
}
```

### Finding Element at Caret

```kotlin
fun getElementAtCaret(editor: Editor, project: Project): PsiElement? {
    val offset = editor.caretModel.offset
    val psiFile = PsiDocumentManager.getInstance(project).getPsiFile(editor.document)
    return psiFile?.findElementAt(offset)
}

fun getTypedElementAtCaret(editor: Editor, project: Project, type: Class<T>): T? {
    val element = getElementAtCaret(editor, project) ?: return null
    return PsiTreeUtil.getParentOfType(element, type)
}
```

### Walking Up the Tree

```kotlin
fun walkToRoot(element: PsiElement): List<PsiElement> {
    val path = mutableListOf<PsiElement>()
    var current: PsiElement? = element

    while (current != null) {
        path.add(current)
        current = current.parent
    }

    return path
}
```

## Performance Considerations

### Avoid Deep Recursion

```kotlin
// WRONG: Deep recursion on large files
fun countElements(element: PsiElement): Int {
    var count = 1
    for (child in element.children) {
        count += countElements(child)
    }
    return count
}

// CORRECT: Use iterative approach
fun countElementsIterative(root: PsiElement): Int {
    var count = 0
    PsiTreeUtil.processElements(root) {
        count++
        true
    }
    return count
}
```

### Use Smart Pointers for Long Operations

```kotlin
val pointer = SmartPointerManager.getInstance(project)
    .createSmartPsiElementPointer(element)

// After long operation
val element = pointer.element
if (element?.isValid == true) {
    // Use element
}
```

## Related Topics

- [PSI Elements](psi-elements.md) - Understanding element types
- [PSI References](psi-references.md) - Reference resolution
- [PSI Performance](psi-performance.md) - Performance optimization

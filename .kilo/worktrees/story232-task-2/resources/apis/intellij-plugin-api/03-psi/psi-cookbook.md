# PSI Cookbook

This page provides common patterns and recipes for working with PSI. Each recipe addresses a specific task with working code examples.

## Finding Elements

### Find a Class by Qualified Name

```kotlin
fun findClass(project: Project, qualifiedName: String): PsiClass? {
    return JavaPsiFacade.getInstance(project)
        .findClass(qualifiedName, GlobalSearchScope.projectScope(project))
}
```

### Find All Classes in a Package

```kotlin
fun findClassesInPackage(project: Project, packageName: String): List<PsiClass> {
    val packageFacade = JavaPsiFacade.getInstance(project)
        .findPackage(packageName) ?: return emptyList()

    return packageFacade.getClasses(GlobalSearchScope.projectScope(project)).toList()
}
```

### Find Method by Name in Class

```kotlin
fun findMethod(psiClass: PsiClass, name: String): PsiMethod? {
    return psiClass.findMethodsByName(name, false).firstOrNull()
}

// Include inherited methods
fun findMethodIncludingInherited(psiClass: PsiClass, name: String): PsiMethod? {
    return psiClass.findMethodsByName(name, true).firstOrNull()
}
```

### Find Element at Caret

```kotlin
fun getElementAtCaret(editor: Editor, project: Project): PsiElement? {
    val offset = editor.caretModel.offset
    val psiFile = PsiDocumentManager.getInstance(project).getPsiFile(editor.document)
    return psiFile?.findElementAt(offset)
}

fun getTypedElementAtCaret(editor: Editor, project: Project): PsiElement? {
    val element = getElementAtCaret(editor, project) ?: return null

    // Walk up to find meaningful element
    return PsiTreeUtil.getParentOfType(
        element,
        PsiClass::class.java,
        PsiMethod::class.java,
        PsiField::class.java,
        PsiLocalVariable::class.java
    )
}
```

### Find All Implementations of an Interface

```kotlin
fun findImplementations(psiClass: PsiClass, project: Project): List<PsiClass> {
    if (!psiClass.isInterface) return emptyList()

    return ClassInheritorsSearch.search(
        psiClass,
        GlobalSearchScope.projectScope(project),
        true
    ).findAll().toList()
}
```

## Analyzing Code

### Check if Method Overrides Another

```kotlin
fun isOverride(method: PsiMethod): Boolean {
    return method.findSuperMethods().isNotEmpty()
}

fun getSuperMethod(method: PsiMethod): PsiMethod? {
    return method.findSuperMethods().firstOrNull()
}
```

### Check if Class Implements Interface

```kotlin
fun implementsInterface(psiClass: PsiClass, interfaceName: String): Boolean {
    return psiClass.interfaces.any { it.qualifiedName == interfaceName }
}
```

### Get Method Signature

```kotlin
fun getMethodSignature(method: PsiMethod): String {
    val builder = StringBuilder()

    // Modifiers
    val modifiers = method.modifierList?.text ?: ""
    if (modifiers.isNotEmpty()) {
        builder.append(modifiers).append(" ")
    }

    // Return type
    method.returnType?.let { builder.append(it.presentableText).append(" ") }

    // Name and parameters
    builder.append(method.name)
    builder.append("(")
    builder.append(method.parameterList.parameters.joinToString(", ") {
        "${it.type.presentableText} ${it.name}"
    })
    builder.append(")")

    return builder.toString()
}
```

### Count Lines of Code in a Class

```kotlin
fun countLinesOfCode(psiClass: PsiClass): Int {
    var lines = 0

    psiClass.accept(object : JavaRecursiveElementVisitor() {
        override fun visitMethod(method: PsiMethod) {
            method.body?.let { body ->
                lines += body.text.lines().filter { it.isNotBlank() }.size
            }
            super.visitMethod(method)
        }
    })

    return lines
}
```

## Modifying Code

### Add a Method to a Class

```kotlin
fun addMethod(
    psiClass: PsiClass,
    name: String,
    returnType: PsiType,
    parameters: List<Pair<String, PsiType>>,
    body: String
): PsiMethod {
    return WriteAction.compute<PsiMethod, Exception> {
        val factory = PsiElementFactory.getInstance(psiClass.project)

        // Create method
        val method = factory.createMethod(name, returnType)

        // Add parameters
        parameters.forEach { (paramName, paramType) ->
            val param = factory.createParameter(paramName, paramType)
            method.parameterList.add(param)
        }

        // Set body
        val codeBlock = factory.createCodeBlockFromText(body, method)
        method.body?.replace(codeBlock)

        // Add to class
        psiClass.add(method) as PsiMethod
    }
}
```

### Add a Field to a Class

```kotlin
fun addField(
    psiClass: PsiClass,
    name: String,
    type: PsiType,
    initializer: String? = null,
    modifiers: String = "private"
): PsiField {
    return WriteAction.compute<PsiField, Exception> {
        val factory = PsiElementFactory.getInstance(psiClass.project)

        val field = factory.createField(name, type)

        // Set modifiers
        val modifierList = factory.createModifierList(modifiers)
        field.modifierList?.replace(modifierList)

        // Set initializer
        initializer?.let {
            val initExpr = factory.createExpressionFromText(it, field)
            field.initializer = initExpr
        }

        psiClass.add(field) as PsiField
    }
}
```

### Add Import Statement

```kotlin
fun addImport(psiFile: PsiFile, qualifiedName: String) {
    WriteAction.run<Exception> {
        val javaFile = psiFile as? PsiJavaFile ?: return@run
        JavaCodeStyleManager.getInstance(psiFile.project)
            .addImport(javaFile, qualifiedName)
    }
}
```

### Remove Unused Imports

```kotlin
fun removeUnusedImports(psiFile: PsiFile) {
    WriteAction.run<Exception> {
        JavaCodeStyleManager.getInstance(psiFile.project)
            .optimizeImports(psiFile)
    }
}
```

### Generate Getter and Setter

```kotlin
fun generateGetterSetter(psiClass: PsiClass, field: PsiField) {
    WriteAction.run<Exception> {
        val factory = PsiElementFactory.getInstance(psiClass.project)

        // Generate getter
        val getterName = "get${field.name.replaceFirstChar { it.uppercase() }}"
        val getter = factory.createMethod(getterName, field.type)
        getter.modifierList?.setModifierProperty(PsiModifier.PUBLIC, true)
        getter.body?.add(factory.createStatementFromText("return ${field.name};", getter))

        // Generate setter
        val setterName = "set${field.name.replaceFirstChar { it.uppercase() }}"
        val setter = factory.createMethod(setterName, PsiType.VOID)
        setter.modifierList?.setModifierProperty(PsiModifier.PUBLIC, true)
        val param = factory.createParameter(field.name, field.type)
        setter.parameterList.add(param)
        setter.body?.add(factory.createStatementFromText("this.${field.name} = ${field.name};", setter))

        psiClass.add(getter)
        psiClass.add(setter)
    }
}
```

## Working with References

### Find All Usages of an Element

```kotlin
fun findUsages(element: PsiElement, project: Project): List<PsiReference> {
    return ReferencesSearch.search(
        element,
        GlobalSearchScope.projectScope(project)
    ).findAll().toList()
}

fun findUsageLocations(element: PsiElement, project: Project): List<String> {
    return findUsages(element, project).map { ref ->
        val file = ref.element.containingFile
        val line = PsiDocumentManager.getInstance(project)
            .getDocument(file)
            ?.getLineNumber(ref.element.textRange?.startOffset ?: 0) ?: 0
        "${file.name}:${line + 1}"
    }
}
```

### Check if Element Is Used

```kotlin
fun isUsed(element: PsiNamedElement, project: Project): Boolean {
    return ReferencesSearch.search(
        element,
        GlobalSearchScope.projectScope(project)
    ).findFirst() != null
}
```

### Find Dead Code

```kotlin
fun findUnusedPrivateMethods(psiClass: PsiClass, project: Project): List<PsiMethod> {
    return psiClass.methods
        .filter { it.hasModifierProperty(PsiModifier.PRIVATE) }
        .filter { !isUsed(it, project) }
}
```

## Working with Comments

### Add Javadoc to a Method

```kotlin
fun addJavadoc(method: PsiMethod, comment: String) {
    WriteAction.run<Exception> {
        val factory = PsiElementFactory.getInstance(method.project)
        val docComment = factory.createDocCommentFromText("/** $comment */")
        method.addBefore(docComment, method.firstChild)
    }
}
```

### Extract TODO Comments

```kotlin
fun findTodos(psiFile: PsiFile): List<Pair<String, Int>> {
    val todos = mutableListOf<Pair<String, Int>>()
    val document = PsiDocumentManager.getInstance(psiFile.project).getDocument(psiFile)

    psiFile.accept(object : PsiRecursiveElementVisitor() {
        override fun visitComment(comment: PsiComment) {
            if (comment.text.contains("TODO", ignoreCase = true)) {
                val line = document?.getLineNumber(comment.textRange?.startOffset ?: 0) ?: 0
                todos.add(comment.text to line + 1)
            }
        }
    })

    return todos
}
```

## Working with Annotations

### Check for Annotation

```kotlin
fun hasAnnotation(element: PsiModifierListOwner, annotationFqn: String): Boolean {
    return element.modifierList?.findAnnotation(annotationFqn) != null
}

fun getAnnotationValue(element: PsiModifierListOwner, annotationFqn: String, attribute: String): String? {
    val annotation = element.modifierList?.findAnnotation(annotationFqn) ?: return null
    return annotation.findAttributeValue(attribute)?.text?.removeSurrounding("\"")
}
```

### Add Annotation

```kotlin
fun addAnnotation(element: PsiModifierListOwner, annotationFqn: String) {
    WriteAction.run<Exception> {
        val factory = PsiElementFactory.getInstance(element.project)
        val annotation = factory.createAnnotationFromText("@$annotationFqn", element)
        element.modifierList?.add(annotation)
    }
}
```

## Working with Expressions

### Evaluate Constant Expression

```kotlin
fun evaluateConstant(expression: PsiExpression): Any? {
    val constant = ConstantExpressionEvaluator.computeConstantExpression(expression, false)
    return constant
}
```

### Find Method Calls

```kotlin
fun findMethodCalls(psiFile: PsiFile, methodName: String): List<PsiMethodCallExpression> {
    val calls = mutableListOf<PsiMethodCallExpression>()

    psiFile.accept(object : JavaRecursiveElementVisitor() {
        override fun visitMethodCallExpression(expression: PsiMethodCallExpression) {
            if (expression.methodExpression.referenceName == methodName) {
                calls.add(expression)
            }
            super.visitMethodCallExpression(expression)
        }
    })

    return calls
}
```

## Working with Files

### Create a New Java File

```kotlin
fun createJavaFile(
    directory: PsiDirectory,
    packageName: String,
    className: String,
    content: String
): PsiFile {
    return WriteAction.compute<PsiFile, Exception> {
        val factory = PsiFileFactory.getInstance(directory.project)
        val fileName = "$className.java"
        val file = factory.createFileFromText(fileName, JavaLanguage.INSTANCE, content)
        directory.add(file) as PsiFile
    }
}
```

### Copy a File

```kotlin
fun copyFile(source: PsiFile, targetDirectory: PsiDirectory, newName: String): PsiFile {
    return WriteAction.compute<PsiFile, Exception> {
        val copy = source.copy() as PsiFile
        copy.setName(newName)
        targetDirectory.add(copy) as PsiFile
    }
}
```

## Related Topics

- [PSI Elements](psi-elements.md) - Understanding element types
- [Modifying PSI](modifying-psi.md) - Safe modifications
- [PSI References](psi-references.md) - Reference resolution

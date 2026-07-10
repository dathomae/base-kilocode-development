# File View Providers

File View Providers manage PSI trees for files that contain multiple languages. They enable the platform to handle mixed-language files like HTML with embedded JavaScript or JSP files.

## Overview

### What Is a FileViewProvider?

A `FileViewProvider` manages one or more PSI trees for a single file. When a file contains multiple languages, the view provider creates separate PSI trees for each language.

```kotlin
interface FileViewProvider {
    val virtualFile: VirtualFile
    val file: PsiFile?
    val languages: Set<Language>
    val baseLanguage: Language

    fun getPsi(language: Language): PsiFile?
    fun getContents(): CharSequence
}
```

### When View Providers Are Needed

View providers handle:

- HTML files with embedded CSS and JavaScript
- JSP files with Java code
- PHP files with HTML
- Template files with multiple languages
- Any file with language injection

## Accessing View Providers

### From a PsiFile

```kotlin
val viewProvider = psiFile.viewProvider
```

### From a VirtualFile

```kotlin
val viewProvider = PsiManager.getInstance(project)
    .findViewProvider(virtualFile)
```

### From an Element

```kotlin
val viewProvider = psiElement.containingFile.viewProvider
```

## Working with Multiple Languages

### Getting Available Languages

```kotlin
val viewProvider = psiFile.viewProvider
val languages = viewProvider.languages

for (language in languages) {
    println("Language: ${language.id}")
}
```

### Getting PSI for a Language

```kotlin
val viewProvider = psiFile.viewProvider

// Get HTML PSI
val htmlFile = viewProvider.getPsi(HTMLLanguage.INSTANCE)

// Get JavaScript PSI
val jsFile = viewProvider.getPsi(JavaScriptLanguage.INSTANCE)
```

### Getting the Base Language

```kotlin
val baseLanguage = viewProvider.baseLanguage
val baseFile = viewProvider.getPsi(baseLanguage)
```

## Language Injection

### What Is Language Injection?

Language injection embeds one language inside another. For example:

- SQL strings in Java code
- Regular expressions in string literals
- HTML in template strings

### Accessing Injected PSI

```kotlin
val element: PsiElement = ...  // A string literal containing SQL

// Get injected PSI
val injectedFile = InjectedLanguageManager.getInstance(project)
    .getInjectedPsiFiles(element)

if (injectedFile != null) {
    for ((injectedPsi, range) in injectedFile) {
        println("Injected language: ${injectedPsi.language}")
    }
}
```

### Creating Injected Language

```kotlin
// Inject SQL into a string literal
val element: PsiLiteralExpression = ...
val injectedLanguage = Language.findLanguageByID("SQL")

InjectedLanguageManager.getInstance(project)
    .startInjectedLanguageEditing(element, injectedLanguage)
```

## Creating Custom View Providers

### When to Create a Custom Provider

Create a custom view provider when:

- Supporting a new template language
- Handling files with custom embedding rules
- Providing language-specific PSI for embedded code

### Implementing a View Provider

```kotlin
class MyViewProvider(
    manager: PsiManager,
    file: VirtualFile,
    eventSystemEnabled: Boolean
) : FileViewProviderBase(manager, file, eventSystemEnabled) {

    override val baseLanguage: Language = MyLanguage.INSTANCE

    override val languages: Set<Language>
        get() = setOf(baseLanguage, XMLLanguage.INSTANCE)

    override fun createFile(language: Language): PsiFile? {
        return when (language) {
            baseLanguage -> MyFileImpl(this)
            XMLLanguage.INSTANCE -> XmlFileImpl(this)
            else -> null
        }
    }
}
```

### Registering a View Provider

```xml
<extensions defaultExtensionNs="com.intellij">
    <fileType.viewProvider
        filetype="MyFileType"
        implementationClass="com.example.MyViewProvider"/>
</extensions>
```

## View Provider Factory

### Creating a Factory

```kotlin
class MyViewProviderFactory : FileViewProviderFactory {
    override fun createFileViewProvider(
        file: VirtualFile,
        language: Language,
        manager: PsiManager,
        eventSystemEnabled: Boolean
    ): FileViewProvider {
        return MyViewProvider(manager, file, eventSystemEnabled)
    }
}
```

### Registering the Factory

```xml
<extensions defaultExtensionNs="com.intellij">
    <lang.fileViewProviderFactory
        language="MyLanguage"
        implementationClass="com.example.MyViewProviderFactory"/>
</extensions>
```

## Document Synchronization

### Getting the Document

```kotlin
val document = viewProvider.document
val contents = viewProvider.contents
```

### Committing Changes

```kotlin
// Commit document changes to PSI
viewProvider.commitAllDocuments()

// Check if commit is needed
if (viewProvider.isEventSystemEnabled) {
    // Changes will be automatically committed
}
```

## Element Mapping

### Mapping Between PSI Trees

Elements in different PSI trees can be mapped to each other:

```kotlin
// Get element at offset in specific language
val offset = 100
val htmlElement = viewProvider.findElementAt(offset, HTMLLanguage.INSTANCE)
val jsElement = viewProvider.findElementAt(offset, JavaScriptLanguage.INSTANCE)
```

### Getting PSI Elements

```kotlin
// Get element at offset in base language
val element = viewProvider.findElementAt(offset)

// Get element at offset in specific language
val element = viewProvider.findElementAt(offset, language)
```

## Common Patterns

### Processing All Languages

```kotlin
fun processAllLanguages(psiFile: PsiFile) {
    val viewProvider = psiFile.viewProvider

    for (language in viewProvider.languages) {
        val file = viewProvider.getPsi(language) ?: continue

        println("Processing ${language.id}")
        file.accept(object : PsiRecursiveElementVisitor() {
            override fun visitElement(element: PsiElement) {
                // Process element
                super.visitElement(element)
            }
        })
    }
}
```

### Finding Elements Across Languages

```kotlin
fun findElementInAnyLanguage(
    viewProvider: FileViewProvider,
    offset: Int
): PsiElement? {
    for (language in viewProvider.languages) {
        val element = viewProvider.findElementAt(offset, language)
        if (element != null) return element
    }
    return null
}
```

### Working with Injected SQL

```kotlin
fun processSqlInJavaFile(javaFile: PsiJavaFile) {
    val injectedManager = InjectedLanguageManager.getInstance(javaFile.project)

    javaFile.accept(object : PsiRecursiveElementVisitor() {
        override fun visitElement(element: PsiElement) {
            if (element is PsiLiteralExpression) {
                val injected = injectedManager.getInjectedPsiFiles(element)
                injected?.forEach { (psiFile, _) ->
                    if (psiFile.language.id == "SQL") {
                        processSqlFile(psiFile)
                    }
                }
            }
            super.visitElement(element)
        }
    })
}
```

## Related Topics

- [PSI Files](psi-files.md) - Working with PSI files
- [Language Registration](../05-language-support/language-registration.md) - Registering languages
- [Syntax Highlighting](../05-language-support/syntax-highlighting.md) - Highlighting mixed-language files

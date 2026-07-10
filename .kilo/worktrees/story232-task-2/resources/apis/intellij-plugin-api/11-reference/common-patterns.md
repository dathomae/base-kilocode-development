# Common Patterns

This document covers common code patterns used in IntelliJ plugin development.

## Service Patterns

### Application Service

```kotlin
// Service interface
interface MyService {
    fun processData(data: String): String
}

// Service implementation
class MyServiceImpl : MyService {
    override fun processData(data: String): String {
        return data.uppercase()
    }
}

// Registration
// plugin.xml:
// <applicationService serviceInterface="com.example.MyService"
//                     serviceImplementation="com.example.MyServiceImpl"/>

// Usage
val service = ApplicationManager.getApplication().getService(MyService::class.java)
```

### Project Service

```kotlin
class MyProjectService(val project: Project) {
    private val data = mutableMapOf<String, Any>()

    fun getData(key: String): Any? = data[key]

    fun setData(key: String, value: Any) {
        data[key] = value
    }

    companion object {
        fun getInstance(project: Project): MyProjectService {
            return project.getService(MyProjectService::class.java)
        }
    }
}
```

## Action Patterns

### Simple Action

```kotlin
class MyAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return

        // Perform action
        Messages.showInfoMessage(project, "Action executed", "Info")
    }

    override fun update(e: AnActionEvent) {
        val project = e.project
        e.presentation.isEnabled = project != null
    }
}
```

### Toggle Action

```kotlin
class MyToggleAction : ToggleAction("Toggle Feature") {
    private var enabled = false

    override fun isSelected(e: AnActionEvent): Boolean = enabled

    override fun setSelected(e: AnActionEvent, state: Boolean) {
        enabled = state
        // Handle state change
    }
}
```

## PSI Patterns

### Finding Elements

```kotlin
fun findClass(project: Project, className: String): PsiClass? {
    val scope = GlobalSearchScope.projectScope(project)
    return JavaPsiFacade.getInstance(project)
        .findClass(className, scope)
}

fun findFile(project: Project, fileName: String): PsiFile? {
    val scope = GlobalSearchScope.projectScope(project)
    val files = FilenameIndex.getVirtualFilesByName(project, fileName, scope)
    return files.firstOrNull()?.let {
        PsiManager.getInstance(project).findFile(it)
    }
}
```

### Navigating PSI

```kotlin
fun findParentClass(element: PsiElement): PsiClass? {
    var current: PsiElement? = element
    while (current != null) {
        if (current is PsiClass) {
            return current
        }
        current = current.parent
    }
    return null
}

fun findChildrenOfType(element: PsiElement, type: Class<out PsiElement>): List<PsiElement> {
    val result = mutableListOf<PsiElement>()
    element.accept(object : PsiRecursiveElementVisitor() {
        override fun visitElement(element: PsiElement) {
            if (type.isInstance(element)) {
                result.add(element)
            }
            super.visitElement(element)
        }
    })
    return result
}
```

### Modifying PSI

```kotlin
fun addMethod(psiClass: PsiClass, methodText: String) {
    WriteAction.run<Exception> {
        val factory = PsiElementFactory.getInstance(psiClass.project)
        val method = factory.createMethodFromText(methodText, psiClass)
        psiClass.add(method)
    }
}

fun replaceElement(oldElement: PsiElement, newText: String) {
    WriteAction.run<Exception> {
        val factory = PsiElementFactory.getInstance(oldElement.project)
        val newElement = factory.createExpressionFromText(newText, oldElement.parent)
        oldElement.replace(newElement)
    }
}
```

## Editor Patterns

### Getting Current Editor

```kotlin
fun getCurrentEditor(project: Project): Editor? {
    val fileEditorManager = FileEditorManager.getInstance(project)
    return fileEditorManager.selectedTextEditor
}

fun getCurrentFile(project: Project): PsiFile? {
    val editor = getCurrentEditor(project) ?: return null
    val document = editor.document
    return PsiDocumentManager.getInstance(project).getPsiFile(document)
}
```

### Modifying Document

```kotlin
fun insertText(editor: Editor, offset: Int, text: String) {
    WriteAction.run<Exception> {
        editor.document.insertString(offset, text)
    }
}

fun replaceText(editor: Editor, start: Int, end: Int, text: String) {
    WriteAction.run<Exception> {
        editor.document.replaceString(start, end, text)
    }
}
```

## Virtual File Patterns

### Finding Files

```kotlin
fun findVirtualFile(project: Project, path: String): VirtualFile? {
    val baseDir = project.baseDir ?: return null
    return baseDir.findFileByRelativePath(path)
}

fun listFiles(dir: VirtualFile): List<VirtualFile> {
    return dir.children.filter { it.isFile }
}

fun listDirectories(dir: VirtualFile): List<VirtualFile> {
    return dir.children.filter { it.isDirectory }
}
```

### File Operations

```kotlin
fun createFile(parent: VirtualFile, name: String, content: String): VirtualFile? {
    return WriteAction.compute<VirtualFile?, Exception> {
        val file = parent.createChildData(null, name)
        VfsUtil.saveText(file, content)
        file
    }
}

fun deleteFile(file: VirtualFile) {
    WriteAction.run<Exception> {
        file.delete(null)
    }
}
```

## Thread Patterns

### Running on EDT

```kotlin
fun runOnEdt(action: () -> Unit) {
    if (EventQueue.isDispatchThread()) {
        action()
    } else {
        ApplicationManager.getApplication().invokeLater(action)
    }
}

fun runOnEdtAndWait(action: () -> Unit) {
    if (EventQueue.isDispatchThread()) {
        action()
    } else {
        ApplicationManager.getApplication().invokeAndWait(action)
    }
}
```

### Background Task

```kotlin
fun runInBackground(project: Project, title: String, action: (ProgressIndicator) -> Unit) {
    ProgressManager.getInstance().run(object : Task.Backgroundable(project, title) {
        override fun run(indicator: ProgressIndicator) {
            action(indicator)
        }
    })
}
```

## Notification Patterns

### Showing Notifications

```kotlin
fun showInfo(project: Project, title: String, message: String) {
    NotificationGroupManager.getInstance()
        .getNotificationGroup("MyPlugin")
        .createNotification(title, message, NotificationType.INFORMATION)
        .notify(project)
}

fun showError(project: Project, title: String, message: String) {
    NotificationGroupManager.getInstance()
        .getNotificationGroup("MyPlugin")
        .createNotification(title, message, NotificationType.ERROR)
        .notify(project)
}
```

## Caching Patterns

### Data Caching

```kotlin
class CachedData<T>(private val loader: () -> T) {
    private var cachedValue: T? = null
    private var lastModified: Long = 0

    fun get(): T {
        val currentModified = getCurrentModification()

        if (cachedValue == null || lastModified != currentModified) {
            cachedValue = loader()
            lastModified = currentModified
        }

        return cachedValue!!
    }

    private fun getCurrentModification(): Long {
        return System.currentTimeMillis()
    }
}
```

### PSI Caching

```kotlin
class PsiCache<T : PsiElement> {
    private val cache = ConcurrentHashMap<PsiFile, T>()

    fun get(file: PsiFile, loader: (PsiFile) -> T): T {
        return cache.getOrPut(file) {
            loader(file)
        }
    }

    fun invalidate(file: PsiFile) {
        cache.remove(file)
    }
}
```

## Listener Patterns

### Message Bus

```kotlin
class MyListener(val project: Project) : Disposable {
    private val connection = project.messageBus.connect()

    init {
        connection.subscribe(
            AppTopics.FILE_DOCUMENT_SYNC,
            object : FileDocumentManagerListener {
                override fun fileDocumentSaved(file: VirtualFile, document: Document) {
                    // Handle save
                }
            }
        )
    }

    override fun dispose() {
        connection.disconnect()
    }
}
```

## Best Practices

### Pattern Guidelines

1. **Use services**: Encapsulate logic in services
2. **Handle threads**: Run UI on EDT, background in tasks
3. **Use disposables**: Clean up resources properly
4. **Cache wisely**: Cache expensive computations
5. **Listen efficiently**: Use message bus for events

## Related Topics

- [Services](../01-core-concepts/services.md) - Service implementation
- [Actions System](../01-core-concepts/actions-system.md) - Action patterns
- [Performance Tips](performance-tips.md) - Performance optimization

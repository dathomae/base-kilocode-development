# Virtual File System

The Virtual File System (VFS) is an abstraction layer that provides uniform access to files regardless of their physical location. It is a fundamental component of the IntelliJ Platform.

## Overview

### What Is VFS?

VFS provides a consistent API for working with files that may be:

- On the local filesystem
- Inside JAR or ZIP files
- In version control systems
- Generated dynamically
- Remote files (via plugins)

The core interface is `VirtualFile`, which represents a file or directory in the VFS.

### VirtualFile Interface

```kotlin
interface VirtualFile {
    val name: String
    val extension: String?
    val path: String
    val url: String
    val parent: VirtualFile?
    val isDirectory: Boolean
    val exists: Boolean
    val length: Long
    val timeStamp: Long
    val modificationStamp: Long

    fun getChildren(): Array<VirtualFile>?
    fun findChild(name: String): VirtualFile?
    fun contentsToByteArray(): ByteArray
    fun inputStream(): InputStream
    fun outputStream(requestor: Any?): OutputStream
}
```

## Accessing Files

### By Path

```kotlin
// Find file by path
val file = LocalFileSystem.getInstance().findFileByPath("/path/to/file.txt")

// Refresh and find
val file = LocalFileSystem.getInstance().refreshAndFindFileByPath("/path/to/file.txt")
```

### By URL

```kotlin
val file = VirtualFileManager.getInstance().findFileByUrl("file:///path/to/file.txt")
val jarFile = VirtualFileManager.getInstance().findFileByUrl("jar:///path/to/file.jar!/internal/path")
```

### From an Action

```kotlin
class MyAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val file = e.getData(CommonDataKeys.VIRTUAL_FILE)
        val files = e.getData(CommonDataKeys.VIRTUAL_FILE_ARRAY)
    }
}
```

### From PSI

```kotlin
val psiFile: PsiFile = ...
val virtualFile = psiFile.virtualFile
```

## File Operations

### Reading Files

```kotlin
// Read as string
val content = String(file.contentsToByteArray(), StandardCharsets.UTF_8)

// Read via stream
file.inputStream().use { stream ->
    // Process stream
}
```

### Writing Files

Writing requires a write action:

```kotlin
WriteAction.run<Exception> {
    file.outputStream(this).use { stream ->
        stream.write(content.toByteArray())
    }
}
```

### Creating Files

```kotlin
val parent: VirtualFile = ...

WriteAction.run<Exception> {
    val newFile = parent.createChildData(this, "newfile.txt")
    newFile.outputStream(this).use { stream ->
        stream.write(content.toByteArray())
    }
}
```

### Creating Directories

```kotlin
val parent: VirtualFile = ...

WriteAction.run<Exception> {
    val newDir = parent.createChildDirectory(this, "newdir")
}
```

### Deleting Files

```kotlin
WriteAction.run<Exception> {
    file.delete(this)
}
```

### Renaming Files

```kotlin
WriteAction.run<Exception> {
    file.rename(this, "newname.txt")
}
```

### Moving Files

```kotlin
val newParent: VirtualFile = ...

WriteAction.run<Exception> {
    file.move(this, newParent)
}
```

## File System Events

### Listening for Changes

```kotlin
val connection = VirtualFileManager.getInstance().addVirtualFileListener(
    object : VirtualFileListener {
        override fun contentsChanged(event: VirtualFileEvent) {
            // File contents changed
        }

        override fun fileCreated(event: VirtualFileEvent) {
            // File created
        }

        override fun fileDeleted(event: VirtualFileEvent) {
            // File deleted
        }

        override fun fileMoved(event: VirtualFileMoveEvent) {
            // File moved
        }

        override fun fileRenamed(event: VirtualFilePropertyEvent) {
            // File renamed
        }
    },
    parentDisposable
)
```

### Bulk Changes

For bulk operations, use `VfsBatchListener`:

```kotlin
VirtualFileManager.getInstance().addVirtualFileListener(
    object : VirtualFileListener {
        override fun beforePropertyChange(event: VirtualFilePropertyEvent) {
            // Before property change
        }
    }
)
```

## Refreshing Files

### Refreshing a Single File

```kotlin
file.refresh(false, false)  // asynchronous, not recursive
```

### Refreshing a Directory

```kotlin
directory.refresh(true, true)  // asynchronous, recursive
```

### Refreshing with Callback

```kotlin
directory.refresh(true, true, Runnable {
    // Called after refresh completes
})
```

### Synchronous Refresh

```kotlin
VfsUtil.markDirtyAndRefresh(false, true, true, directory)
```

## File Attributes

### Basic Attributes

```kotlin
val name = file.name           // "file.txt"
val extension = file.extension // "txt"
val path = file.path           // "/path/to/file.txt"
val url = file.url             // "file:///path/to/file.txt"
val isDirectory = file.isDirectory
val exists = file.exists
val length = file.length
```

### Timestamps

```kotlin
val created = file.timeStamp
val modified = file.modificationStamp
```

### File System

```kotlin
val fileSystem = file.fileSystem
val fsType = fileSystem.protocol  // "file", "jar", etc.
```

## Special File Systems

### Local File System

```kotlin
val localFs = LocalFileSystem.getInstance()
val file = localFs.findFileByPath("/local/path")
```

### JAR File System

```kotlin
val jarUrl = "jar:///path/to/archive.jar!/path/inside"
val file = VirtualFileManager.getInstance().findFileByUrl(jarUrl)
```

### Temp File System

For testing, use the in-memory temp file system:

```kotlin
val tempFs = VirtualFileManager.getInstance().getFileSystem("temp")
val tempFile = tempFs.findFileByPath("/temp/path")
```

## Performance Considerations

### Caching

VFS caches file information. Avoid unnecessary refreshes:

```kotlin
// WRONG: Refreshing every time
fun getFileContent(path: String): String? {
    val file = LocalFileSystem.getInstance().refreshAndFindFileByPath(path)
    return file?.let { String(it.contentsToByteArray()) }
}

// CORRECT: Use cached version, refresh only when needed
fun getFileContent(path: String): String? {
    val file = LocalFileSystem.getInstance().findFileByPath(path)
    return file?.let { String(it.contentsToByteArray()) }
}
```

### Bulk Operations

For bulk file operations, batch them:

```kotlin
WriteAction.run<Exception> {
    for (file in files) {
        file.delete(this)
    }
}
```

### Avoiding Memory Leaks

Do not hold references to `VirtualFile` objects longer than necessary. Files can be invalidated:

```kotlin
// Check if file is still valid
if (file.isValid) {
    // Use file
}
```

## Common Patterns

### Finding Files Recursively

```kotlin
fun findFilesRecursively(root: VirtualFile, extension: String): List<VirtualFile> {
    val result = mutableListOf<VirtualFile>()
    VfsUtilCore.iterateChildrenRecursively(root, null) { file ->
        if (!file.isDirectory && file.extension == extension) {
            result.add(file)
        }
        true
    }
    return result
}
```

### Getting Relative Path

```kotlin
val basePath = "/project/root"
val file = LocalFileSystem.getInstance().findFileByPath("/project/root/src/Main.java")
val relativePath = VfsUtil.getRelativePath(file, baseDir, '/')
// Result: "src/Main.java"
```

### Copying Files

```kotlin
val source: VirtualFile = ...
val targetDir: VirtualFile = ...

WriteAction.run<Exception> {
    VfsUtilCore.copyFile(this, source, targetDir)
}
```

### Checking File Type

```kotlin
val fileType = FileTypeManager.getInstance().getFileTypeByFile(file)
if (fileType == FileTypeManager.getInstance().getFileTypeByExtension("java")) {
    // Java file
}
```

## Thread Safety

### Read Access

Reading VFS requires read access:

```kotlin
ReadAction.compute<VirtualFile?, Exception> {
    LocalFileSystem.getInstance().findFileByPath("/path/to/file")
}
```

### Write Access

Modifying VFS requires write access:

```kotlin
WriteAction.run<Exception> {
    file.delete(this)
}
```

### File System Operations

Some operations automatically handle threading:

```kotlin
// Refresh handles threading internally
VfsUtil.markDirtyAndRefresh(false, true, true, directory)
```

## Related Topics

- [Project Model](project-model.md) - How files relate to projects
- [Indexing](indexing.md) - How VFS integrates with indexing
- [PSI Files](../03-psi/psi-files.md) - PSI representation of files

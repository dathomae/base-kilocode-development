# Document Model

Documents represent the text content being edited. They are separate from editors and can be shared between multiple editor instances.

## Overview

### What Is a Document?

A `Document` is an interface representing text content:

```kotlin
interface Document {
    val text: CharSequence
    val textLength: Int
    val lineNumber: Int
    val modificationStamp: Long

    fun getLineNumber(offset: Int): Int
    fun getLineStartOffset(line: Int): Int
    fun getLineEndOffset(line: Int): Int
    fun insertString(offset: Int, s: CharSequence)
    fun deleteString(startOffset: Int, endOffset: Int)
    fun replaceString(startOffset: Int, endOffset: Int, s: CharSequence)
}
```

### Document vs. PSI

Documents and PSI are synchronized views of the same content:

| Aspect     | Document        | PSI             |
| ---------- | --------------- | --------------- |
| Level      | Text            | Syntax tree     |
| Operations | Character-based | Element-based   |
| Updates    | Immediate       | Requires commit |
| Use case   | Text editing    | Code analysis   |

## Getting a Document

### From a VirtualFile

```kotlin
val virtualFile: VirtualFile = ...
val document = FileDocumentManager.getInstance().getDocument(virtualFile)
```

### From a PsiFile

```kotlin
val psiFile: PsiFile = ...
val document = PsiDocumentManager.getInstance(project).getDocument(psiFile)
```

### From an Editor

```kotlin
val editor: Editor = ...
val document = editor.document
```

### Creating a Document

```kotlin
val factory = EditorFactory.getInstance()
val document = factory.createDocument("Initial content")
```

## Reading Document Content

### Getting Text

```kotlin
// Full text
val text = document.text

// Text range
val substring = document.getText(TextRange(start, end))

// Lines
val lineCount = document.lineCount
val lineText = document.getText(
    TextRange(
        document.getLineStartOffset(lineNumber),
        document.getLineEndOffset(lineNumber)
    )
)
```

### Line Information

```kotlin
// Get line number from offset
val lineNumber = document.getLineNumber(offset)

// Get line boundaries
val lineStart = document.getLineStartOffset(lineNumber)
val lineEnd = document.getLineEndOffset(lineNumber)

// Line separator
val lineSeparator = document.getLineSeparator()

// Check if line is blank
val isBlank = document.getText(
    TextRange(lineStart, lineEnd)
).isBlank()
```

### Offset Conversion

```kotlin
// Offset to logical position
val line = document.getLineNumber(offset)
val column = offset - document.getLineStartOffset(line)

// Logical position to offset
val offset = document.getLineStartOffset(line) + column
```

## Modifying Documents

### Basic Modifications

```kotlin
// Insert text
WriteAction.run<Exception> {
    document.insertString(offset, "inserted text")
}

// Delete text
WriteAction.run<Exception> {
    document.deleteString(startOffset, endOffset)
}

// Replace text
WriteAction.run<Exception> {
    document.replaceString(startOffset, endOffset, "replacement")
}
```

### Setting Full Text

```kotlin
WriteAction.run<Exception> {
    document.setText("New content")
}
```

### Modifying Lines

```kotlin
fun insertLine(document: Document, lineNumber: Int, text: String) {
    WriteAction.run<Exception> {
        val offset = document.getLineStartOffset(lineNumber)
        document.insertString(offset, text + document.lineSeparator)
    }
}

fun deleteLine(document: Document, lineNumber: Int) {
    WriteAction.run<Exception> {
        val start = document.getLineStartOffset(lineNumber)
        val end = document.getLineEndOffset(lineNumber) + document.lineSeparator.length
        document.deleteString(start, minOf(end, document.textLength))
    }
}
```

## Document Events

### Adding a Listener

```kotlin
document.addDocumentListener(object : DocumentListener {
    override fun beforeDocumentChange(event: DocumentEvent) {
        // Called before change
        println("About to change: ${event.oldFragment}")
    }

    override fun documentChanged(event: DocumentEvent) {
        // Called after change
        println("Changed to: ${event.newFragment}")
    }
}, parentDisposable)
```

### Document Event Properties

```kotlin
override fun documentChanged(event: DocumentEvent) {
    val document = event.document
    val offset = event.offset
    val oldLength = event.oldLength
    val newLength = event.newLength
    val oldFragment = event.oldFragment
    val newFragment = event.newFragment
}
```

### Bulk Changes

```kotlin
// Suppress events during bulk operation
document.setSuppressEvents(true)
try {
    // Multiple modifications
    document.insertString(0, "header\n")
    document.insertString(document.textLength, "\nfooter")
} finally {
    document.setSuppressEvents(false)
}
```

## Document and PSI Synchronization

### Committing Documents

Before accessing PSI after document changes:

```kotlin
// Commit document to sync PSI
PsiDocumentManager.getInstance(project).commitDocument(document)

// Commit all documents
PsiDocumentManager.getInstance(project).commitAllDocuments()
```

### Getting PSI from Document

```kotlin
val psiFile = PsiDocumentManager.getInstance(project).getPsiFile(document)
```

### Getting Document from PSI

```kotlin
val document = PsiDocumentManager.getInstance(project).getDocument(psiFile)
```

## File Document Manager

### Saving Documents

```kotlin
val fileDocumentManager = FileDocumentManager.getInstance()

// Save specific document
fileDocumentManager.saveDocument(document)

// Save all documents
fileDocumentManager.saveAllDocuments()

// Check if document is modified
val isModified = fileDocumentManager.isDocumentUnsaved(document)
```

### Document and File Association

```kotlin
// Get file for document
val file = FileDocumentManager.getInstance().getFile(document)

// Get document for file
val document = FileDocumentManager.getInstance().getDocument(virtualFile)

// Request document save
FileDocumentManager.getInstance().requestSaving(document)
```

## Read-Only Documents

### Checking Read-Only Status

```kotlin
val isReadOnly = !document.isWritable
```

### Making Documents Read-Only

```kotlin
document.setReadOnly(true)
```

### Handling Read-Only Files

```kotlin
if (!document.isWritable) {
    // Check if file is read-only on disk
    val file = FileDocumentManager.getInstance().getFile(document)
    if (file?.isWritable == false) {
        // File is read-only on disk
        // Offer to make writable
        val result = Messages.showYesNoDialog(
            project,
            "File is read-only. Make writable?",
            "Read-Only File",
            Messages.getQuestionIcon()
        )
        if (result == Messages.YES) {
            file.isWritable = true
        }
    }
}
```

## Document Modification Stamp

### Tracking Changes

```kotlin
// Get current stamp
val stamp = document.modificationStamp

// Check if changed since last check
if (document.modificationStamp != lastStamp) {
    // Document was modified
    lastStamp = document.modificationStamp
}
```

## Common Patterns

### Appending Text

```kotlin
fun appendText(document: Document, text: String) {
    WriteAction.run<Exception> {
        document.insertString(document.textLength, text)
    }
}
```

### Prepending Text

```kotlin
fun prependText(document: Document, text: String) {
    WriteAction.run<Exception> {
        document.insertString(0, text)
    }
}
```

### Wrapping Selection

```kotlin
fun wrapSelection(editor: Editor, prefix: String, suffix: String) {
    val document = editor.document
    val selectionModel = editor.selectionModel

    if (selectionModel.hasSelection()) {
        WriteAction.run<Exception> {
            val start = selectionModel.selectionStart
            val end = selectionModel.selectionEnd

            document.insertString(end, suffix)
            document.insertString(start, prefix)

            // Adjust selection
            selectionModel.setSelection(start, end + prefix.length + suffix.length)
        }
    }
}
```

### Finding Text

```kotlin
fun findText(document: Document, searchText: String, startOffset: Int = 0): Int {
    val text = document.text.toString()
    return text.indexOf(searchText, startOffset)
}

fun findAllOccurrences(document: Document, searchText: String): List<Int> {
    val text = document.text.toString()
    val occurrences = mutableListOf<Int>()
    var index = 0

    while (true) {
        index = text.indexOf(searchText, index)
        if (index == -1) break
        occurrences.add(index)
        index++
    }

    return occurrences
}
```

### Replacing All Occurrences

```kotlin
fun replaceAll(document: Document, search: String, replace: String) {
    WriteAction.run<Exception> {
        val text = document.text.toString()
        val newText = text.replace(search, replace)
        document.setText(newText)
    }
}
```

## Thread Safety

### Write Actions Required

All document modifications require write actions:

```kotlin
// WRONG: No write action
document.insertString(0, "text")  // Throws exception

// CORRECT: With write action
WriteAction.run<Exception> {
    document.insertString(0, "text")
}
```

### Read Actions for Reading

Reading documents is thread-safe but should use read actions for consistency:

```kotlin
val text = ReadAction.compute<String, Exception> {
    document.text.toString()
}
```

## Related Topics

- [Editor Basics](editor-basics.md) - Editor architecture
- [Modifying PSI](../03-psi/modifying-psi.md) - PSI modifications
- [Code Formatting](code-formatting.md) - Formatting documents

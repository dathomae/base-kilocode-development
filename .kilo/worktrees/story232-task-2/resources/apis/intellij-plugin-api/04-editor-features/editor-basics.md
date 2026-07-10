# Editor Basics

The editor is the central component for displaying and editing text in the IntelliJ Platform. Understanding the editor architecture is essential for plugins that interact with text content.

## Editor Architecture

### Core Components

The editor consists of several key components:

| Component        | Description                |
| ---------------- | -------------------------- |
| `Editor`         | Main editor interface      |
| `Document`       | Text content model         |
| `CaretModel`     | Cursor position management |
| `SelectionModel` | Text selection management  |
| `ScrollingModel` | Viewport scrolling         |
| `FoldingModel`   | Code folding regions       |
| `MarkupModel`    | Highlighting and markers   |

### Editor Hierarchy

```
Editor
├── Document (text content)
├── CaretModel (cursor positions)
├── SelectionModel (selections)
├── ScrollingModel (viewport)
├── FoldingModel (collapsed regions)
├── MarkupModel (highlighters)
└── InlayModel (inline elements)
```

## Getting an Editor

### From an Action

```kotlin
class MyAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val editor = e.getData(CommonDataKeys.EDITOR) ?: return
        // Work with editor
    }
}
```

### From a File

```kotlin
val file: VirtualFile = ...
val editor = FileEditorManager.getInstance(project).getSelectedEditor(file)

// Get text editor specifically
val textEditor = FileEditorManager.getInstance(project)
    .getSelectedEditor(file) as? TextEditor
```

### From a Document

```kotlin
val document: Document = ...
val editors = EditorFactory.getInstance().getEditors(document, project)
```

## Editor Properties

### Basic Properties

```kotlin
val document = editor.document
val project = editor.project
val isViewer = editor.isViewer  // Read-only mode
val component = editor.component  // Swing component
val contentComponent = editor.contentComponent
```

### Visual Properties

```kotlin
val lineHeight = editor.lineHeight
val columnMode = editor.isColumnMode
val oneLineMode = editor.isOneLineMode
val fontMetrics = editor.fontMetrics
```

## Caret Management

### Getting Caret Position

```kotlin
val caretModel = editor.caretModel

// Current position
val offset = caretModel.offset
val line = caretModel.logicalPosition.line
val column = caretModel.logicalPosition.column

// Visual position (accounts for folding)
val visualLine = caretModel.visualPosition.line
val visualColumn = caretModel.visualPosition.column
```

### Moving the Caret

```kotlin
// Move to offset
caretModel.moveToOffset(offset)

// Move to logical position
caretModel.moveToLogicalPosition(LogicalPosition(line, column))

// Move to visual position
caretModel.moveToVisualPosition(VisualPosition(line, column))
```

### Multiple Carets

```kotlin
// Get all carets
val carets = caretModel.carets

// Add a caret
caretModel.addCaret(VisualPosition(line, column))

// Remove a caret
caretModel.removeCaret(caret)

// Remove all secondary carets
caretModel.removeSecondaryCarets()
```

## Selection Management

### Getting Selection

```kotlin
val selectionModel = editor.selectionModel

val selectionStart = selectionModel.selectionStart
val selectionEnd = selectionModel.selectionEnd
val selectedText = selectionModel.selectedText

// Check if selection exists
val hasSelection = selectionModel.hasSelection()
```

### Setting Selection

```kotlin
// Select range
selectionModel.setSelection(start, end)

// Select all
selectionModel.selectAll()

// Clear selection
selectionModel.removeSelection()

// Select line
selectionModel.selectLineAtCaret()
```

### Block Selection

```kotlin
// Enable column mode
editor.isColumnMode = true

// Get block selection
val blockStart = selectionModel.blockSelectionStarts
val blockEnd = selectionModel.blockSelectionEnds
```

## Scrolling

### Getting Viewport

```kotlin
val scrollingModel = editor.scrollingModel

val visibleArea = scrollingModel.visibleArea
val firstVisibleLine = scrollingModel.visibleArea.y / editor.lineHeight
```

### Scrolling

```kotlin
// Scroll to offset
scrollingModel.scrollTo(ScrollType.RELATIVE)

// Scroll to caret
scrollingModel.scrollToCaret(ScrollType.CENTER)

// Scroll to logical position
scrollingModel.scrollTo(LogicalPosition(line, column), ScrollType.CENTER)
```

### Scroll Types

| Type                  | Description                         |
| --------------------- | ----------------------------------- |
| `RELATIVE`            | Scroll minimally to make visible    |
| `CENTER`              | Center the position in viewport     |
| `MAKE_VISIBLE`        | Ensure position is visible          |
| `KEEP_CARET_CENTERED` | Keep caret centered while scrolling |

## Markup and Highlighting

### Adding Highlighters

```kotlin
val markupModel = editor.markupModel

// Add text highlighter
val highlighter = markupModel.addRangeHighlighter(
    startOffset,
    endOffset,
    HighlighterLayer.SELECTION,
    TextAttributes().apply {
        backgroundColor = Color.YELLOW
    },
    HighlighterTargetArea.EXACT_RANGE
)

// Remove highlighter
markupModel.removeHighlighter(highlighter)
```

### Highlighter Layers

| Layer               | Description               |
| ------------------- | ------------------------- |
| `BACKGROUND`        | Background effects        |
| `SELECTION`         | Selection highlighting    |
| `ERROR`             | Error highlighting        |
| `WARNING`           | Warning highlighting      |
| `SYNTAX`            | Syntax highlighting       |
| `ADDITIONAL_SYNTAX` | Additional syntax         |
| `CARET_ROW`         | Current line highlighting |

### Gutter Markers

```kotlin
// Add gutter icon
val gutterRenderer = object : ActiveGutterRenderer {
    override fun paint(editor: Editor, g: Graphics, r: Rectangle) {
        // Custom painting
    }

    override fun getClickAction(editor: Editor, line: Int): AnAction? {
        return MyAction()
    }
}

markupModel.addLineHighlighter(
    line,
    HighlighterLayer.ERROR,
    null
).gutterIconRenderer = gutterRenderer
```

## Inlay Hints

### Adding Inlay Elements

```kotlin
val inlayModel = editor.inlayModel

// Add inline element
val inlay = inlayModel.addInlineElement(
    offset,
    true,
    MyInlayRenderer()
)

// Add block element
val blockInlay = inlayModel.addBlockElement(
    offset,
    true,
    true,
    0,
    MyBlockRenderer()
)
```

### Custom Inlay Renderer

```kotlin
class MyInlayRenderer(private val text: String) : InlayRenderer {
    override fun paint(
        inlay: Inlay<*>,
        g: Graphics,
        targetRegion: Rectangle,
        textAttributes: TextAttributes
    ) {
        g.color = Color.GRAY
        g.drawString(text, targetRegion.x, targetRegion.y + g.fontMetrics.ascent)
    }

    override fun calcWidthInPixels(inlay: Inlay<*>): Int {
        return inlay.editor.fontMetrics.stringWidth(text)
    }
}
```

## Editor Events

### Caret Listener

```kotlin
editor.caretModel.addCaretListener(object : CaretListener {
    override fun caretPositionChanged(event: CaretEvent) {
        // Caret moved
    }

    override fun caretAdded(event: CaretEvent) {
        // Caret added
    }

    override fun caretRemoved(event: CaretEvent) {
        // Caret removed
    }
}, parentDisposable)
```

### Selection Listener

```kotlin
editor.selectionModel.addSelectionListener(object : SelectionListener {
    override fun selectionChanged(e: SelectionEvent) {
        // Selection changed
    }
}, parentDisposable)
```

### Document Listener

```kotlin
editor.document.addDocumentListener(object : DocumentListener {
    override fun documentChanged(event: DocumentEvent) {
        // Document changed
    }
}, parentDisposable)
```

## Editor Actions

### Executing Editor Actions

```kotlin
// Execute built-in action
val actionManager = ActionManager.getInstance()
val action = actionManager.getAction("EditorEnter")
actionManager.tryToExecute(action, e.inputEvent, null, null, false)
```

### Common Editor Actions

| Action ID         | Description               |
| ----------------- | ------------------------- |
| `EditorEnter`     | Insert newline            |
| `EditorBackSpace` | Delete previous character |
| `EditorDelete`    | Delete next character     |
| `EditorTab`       | Insert tab                |
| `EditorLineStart` | Move to line start        |
| `EditorLineEnd`   | Move to line end          |
| `EditorSelectAll` | Select all                |
| `EditorCopy`      | Copy selection            |
| `EditorPaste`     | Paste                     |

## Related Topics

- [Document Model](document-model.md) - Working with documents
- [Code Completion](code-completion.md) - Completion features
- [Tool Windows](../07-ui-components/tool-windows.md) - Editor integration

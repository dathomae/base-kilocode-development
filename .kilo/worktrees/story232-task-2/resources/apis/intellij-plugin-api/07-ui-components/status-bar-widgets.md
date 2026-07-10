# Status Bar Widgets

Status bar widgets display information and provide quick actions in the IDE status bar.

## Overview

### What Are Status Bar Widgets?

Status bar widgets are small UI components that:

- Display status information
- Provide quick access to settings
- Show progress indicators
- Offer context-sensitive actions

### Standard Widgets

| Widget             | Description            |
| ------------------ | ---------------------- |
| Position           | Line and column number |
| Encoding           | File encoding          |
| Line Separator     | Line ending style      |
| Inspection Profile | Current profile name   |
| Memory             | Memory usage indicator |

## Creating a Status Bar Widget

### Basic Widget

```kotlin
class MyStatusBarWidget : StatusBarWidget, StatusBarWidget.TextPresentation {
    private var statusBar: StatusBar? = null
    private var text: String = "Ready"

    override fun ID(): String = "MyWidget"

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar
        updateText()
    }

    override fun dispose() {
        statusBar = null
    }

    override fun getText(): String = text

    override fun getTooltipText(): String? = "My Widget Tooltip"

    override fun getIcon(): Icon? = AllIcons.General.Information

    override fun getClickConsumer(): Consumer<MouseEvent>? {
        return Consumer { event ->
            // Handle click
            showPopup(event)
        }
    }

    private fun updateText() {
        text = computeStatusText()
        statusBar?.updateWidget(ID())
    }

    private fun showPopup(event: MouseEvent) {
        val popup = JBPopupFactory.getInstance()
            .createMessage("Widget clicked!")
        popup.showUnderneathOf(event.component as JComponent)
    }
}
```

### Registering a Widget

```xml
<extensions defaultExtensionNs="com.intellij">
    <statusBarWidgetFactory
        id="MyWidget"
        order="after Position"
        implementationClass="com.example.MyStatusBarWidgetFactory"/>
</extensions>
```

### Widget Factory

```kotlin
class MyStatusBarWidgetFactory : StatusBarWidgetFactory {
    override fun getId(): String = "MyWidget"

    override fun getDisplayName(): String = "My Widget"

    override fun isAvailable(project: Project): Boolean {
        // Control visibility
        return true
    }

    override fun createWidget(project: Project): StatusBarWidget {
        return MyStatusBarWidget(project)
    }

    override fun canCreateOnStatusBar(statusBar: StatusBar): Boolean = true
}
```

## Widget Types

### Text Widget

```kotlin
class TextWidget(project: Project) : StatusBarWidget, StatusBarWidget.TextPresentation {
    private var statusBar: StatusBar? = null
    private var text: String = ""

    override fun ID(): String = "TextWidget"

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar

        // Listen for changes
        project.messageBus.connect().subscribe(
            AppTopics.FILE_DOCUMENT_SYNC,
            object : FileDocumentManagerListener {
                override fun fileDocumentSaved(file: VirtualFile, document: Document) {
                    updateText()
                }
            }
        )

        updateText()
    }

    override fun dispose() {
        statusBar = null
    }

    override fun getText(): String = text

    override fun getTooltipText(): String? = "Current status"

    override fun getIcon(): Icon? = null

    override fun getClickConsumer(): Consumer<MouseEvent>? = null

    private fun updateText() {
        text = computeText()
        statusBar?.updateWidget(ID())
    }
}
```

### Multiple Text Values

```kotlin
class MultiTextWidget : StatusBarWidget, StatusBarWidget.MultipleTextValuesPresentation {
    private var statusBar: StatusBar? = null
    private var values: List<String> = listOf("Option 1", "Option 2", "Option 3")
    private var selectedValue: String = values[0]

    override fun ID(): String = "MultiTextWidget"

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar
    }

    override fun dispose() {
        statusBar = null
    }

    override fun getText(): String = selectedValue

    override fun getTooltipText(): String? = "Select an option"

    override fun getPopupStep(): ListPopup? {
        return JBPopupFactory.getInstance()
            .createPopupChooserBuilder(values)
            .setTitle("Select Option")
            .setItemChosenCallback { selected ->
                selectedValue = selected
                statusBar?.updateWidget(ID())
            }
            .createPopup()
    }

    override fun getIcon(): Icon? = AllIcons.General.ArrowDown
}
```

### Icon Widget

```kotlin
class IconWidget : StatusBarWidget, StatusBarWidget.IconPresentation {
    private var statusBar: StatusBar? = null
    private var icon: Icon = AllIcons.General.Information

    override fun ID(): String = "IconWidget"

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar
    }

    override fun dispose() {
        statusBar = null
    }

    override fun getIcon(): Icon = icon

    override fun getTooltipText(): String? = "Icon widget"

    override fun getClickConsumer(): Consumer<MouseEvent>? {
        return Consumer { event ->
            // Handle click
            toggleIcon()
        }
    }

    private fun toggleIcon() {
        icon = if (icon === AllIcons.General.Information) {
            AllIcons.General.Warning
        } else {
            AllIcons.General.Information
        }
        statusBar?.updateWidget(ID())
    }
}
```

## Widget Updates

### Updating Widget Content

```kotlin
class UpdatableWidget(private val project: Project) : StatusBarWidget, StatusBarWidget.TextPresentation {
    private var statusBar: StatusBar? = null
    private var text: String = "Initializing..."

    override fun ID(): String = "UpdatableWidget"

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar

        // Subscribe to updates
        project.messageBus.connect().subscribe(
            MyTopic.UPDATES,
            object : MyUpdateListener {
                override fun onUpdated(newValue: String) {
                    text = newValue
                    statusBar?.updateWidget(ID())
                }
            }
        )
    }

    override fun dispose() {
        statusBar = null
    }

    override fun getText(): String = text

    override fun getTooltipText(): String? = "Current value: $text"

    override fun getIcon(): Icon? = null

    override fun getClickConsumer(): Consumer<MouseEvent>? = null
}
```

### Periodic Updates

```kotlin
class PeriodicWidget : StatusBarWidget, StatusBarWidget.TextPresentation {
    private var statusBar: StatusBar? = null
    private var text: String = ""
    private var updateTask: Timer? = null

    override fun ID(): String = "PeriodicWidget"

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar

        // Start periodic updates
        updateTask = Timer()
        updateTask?.scheduleAtFixedRate(object : TimerTask() {
            override fun run() {
                updateText()
            }
        }, 0, 1000)
    }

    override fun dispose() {
        updateTask?.cancel()
        updateTask = null
        statusBar = null
    }

    override fun getText(): String = text

    override fun getTooltipText(): String? = null

    override fun getIcon(): Icon? = null

    override fun getClickConsumer(): Consumer<MouseEvent>? = null

    private fun updateText() {
        text = computeCurrentStatus()
        statusBar?.updateWidget(ID())
    }
}
```

## Widget Ordering

### Position in Status Bar

```xml
<extensions defaultExtensionNs="com.intellij">
    <!-- Position after specific widget -->
    <statusBarWidgetFactory
        id="MyWidget"
        order="after Position"
        implementationClass="com.example.MyWidgetFactory"/>

    <!-- Position before specific widget -->
    <statusBarWidgetFactory
        id="AnotherWidget"
        order="before Encoding"
        implementationClass="com.example.AnotherWidgetFactory"/>

    <!-- Position first -->
    <statusBarWidgetFactory
        id="FirstWidget"
        order="first"
        implementationClass="com.example.FirstWidgetFactory"/>
</extensions>
```

### Standard Widget IDs

| Widget ID           | Description         |
| ------------------- | ------------------- |
| `Position`          | Line/column display |
| `Encoding`          | File encoding       |
| `LineSeparator`     | Line endings        |
| `ReadOnlyAttribute` | Read-only indicator |
| `InspectionProfile` | Inspection profile  |
| `MemoryUsage`       | Memory indicator    |

## Context-Sensitive Widgets

### Project-Specific Widget

```kotlin
class ProjectSpecificWidgetFactory : StatusBarWidgetFactory {
    override fun getId(): String = "ProjectWidget"

    override fun getDisplayName(): String = "Project Widget"

    override fun isAvailable(project: Project): Boolean {
        // Only show for specific project types
        return project.isMyProjectType()
    }

    override fun createWidget(project: Project): StatusBarWidget {
        return ProjectWidget(project)
    }

    override fun canCreateOnStatusBar(statusBar: StatusBar): Boolean = true
}
```

### Editor-Sensitive Widget

```kotlin
class EditorSensitiveWidget(private val project: Project) : StatusBarWidget, StatusBarWidget.TextPresentation {
    private var statusBar: StatusBar? = null
    private var text: String = ""

    override fun ID(): String = "EditorWidget"

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar

        // Listen for editor changes
        project.messageBus.connect().subscribe(
            FileEditorManagerListener.FILE_EDITOR_MANAGER,
            object : FileEditorManagerListener {
                override fun selectionChanged(event: FileEditorManagerEvent) {
                    updateForFile(event.newFile)
                }
            }
        )
    }

    override fun dispose() {
        statusBar = null
    }

    override fun getText(): String = text

    override fun getTooltipText(): String? = null

    override fun getIcon(): Icon? = null

    override fun getClickConsumer(): Consumer<MouseEvent>? = null

    private fun updateForFile(file: VirtualFile?) {
        text = if (file != null) {
            "File: ${file.name}"
        } else {
            "No file"
        }
        statusBar?.updateWidget(ID())
    }
}
```

## Best Practices

### Widget Guidelines

1. **Keep text short**: Status bar space is limited
2. **Provide tooltips**: Explain widget purpose
3. **Handle clicks appropriately**: Provide useful actions
4. **Update efficiently**: Avoid excessive updates
5. **Clean up resources**: Dispose properly

### Example: Complete Widget

```kotlin
class MyCompleteWidget(private val project: Project) : StatusBarWidget, StatusBarWidget.TextPresentation {
    private var statusBar: StatusBar? = null
    private var status: String = "Ready"
    private var count: Int = 0

    override fun ID(): String = "MyCompleteWidget"

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar

        // Subscribe to events
        project.messageBus.connect().subscribe(
            MyTopic.COUNT_CHANGES,
            object : CountChangeListener {
                override fun onCountChanged(newCount: Int) {
                    count = newCount
                    updateStatus()
                }
            }
        )

        updateStatus()
    }

    override fun dispose() {
        statusBar = null
    }

    override fun getText(): String = "$status ($count)"

    override fun getTooltipText(): String? = "My Widget: $count items"

    override fun getIcon(): Icon? = when {
        count > 10 -> AllIcons.General.Warning
        count > 0 -> AllIcons.General.Information
        else -> null
    }

    override fun getClickConsumer(): Consumer<MouseEvent>? {
        return Consumer { event ->
            showDetailsPopup(event)
        }
    }

    private fun updateStatus() {
        status = when {
            count == 0 -> "Empty"
            count < 10 -> "Normal"
            else -> "High"
        }
        statusBar?.updateWidget(ID())
    }

    private fun showDetailsPopup(event: MouseEvent) {
        val panel = JPanel(BorderLayout()).apply {
            add(JLabel("Count: $count"), BorderLayout.CENTER)
        }

        JBPopupFactory.getInstance()
            .createComponentPopupBuilder(panel, null)
            .setTitle("Widget Details")
            .createPopup()
            .showUnderneathOf(event.component as JComponent)
    }
}
```

## Related Topics

- [Tool Windows](tool-windows.md) - Dockable panels
- [Progress Indicators](progress-indicators.md) - Progress tracking
- [UI Guidelines](ui-guidelines.md) - Design best practices

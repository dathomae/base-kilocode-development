# UI Guidelines

Guidelines for creating consistent, user-friendly IntelliJ plugin interfaces.

## Overview

### Why Follow Guidelines?

Following UI guidelines ensures:

- Consistency with IDE conventions
- Better user experience
- Professional appearance
- Accessibility compliance

### Core Principles

1. **Consistency**: Match IDE look and feel
2. **Simplicity**: Keep interfaces clean and focused
3. **Feedback**: Provide clear visual feedback
4. **Accessibility**: Support all users

## IntelliJ UI Components

### Using JB Components

Prefer IntelliJ components over standard Swing components:

| Standard    | IntelliJ              |
| ----------- | --------------------- |
| JPanel      | JPanel (with JBColor) |
| JLabel      | JLabel (with JBColor) |
| JButton     | JButton (with Action) |
| JTextField  | JBTextField           |
| JComboBox   | ComboBox              |
| JList       | JBList                |
| JTree       | JTree (with renderer) |
| JTable      | JBTable               |
| JScrollPane | JBScrollPane          |

### Colors

```kotlin
// Use JBColor for theme-aware colors
val background = JBColor.PanelBackground
val foreground = JBColor.foreground()
val selection = JBColor.selectionBackground()

// Custom theme-aware color
val myColor = JBColor(
    Color(0xRRGGBB), // Light theme
    Color(0xRRGGBB)  // Dark theme
)
```

### Icons

```kotlin
// Use built-in icons
val icon = AllIcons.Actions.Refresh

// Load custom icon
val myIcon = IconLoader.getIcon("/META-INF/myIcon.svg", javaClass)

// Icon sizes
val smallIcon = IconLoader.scale(icon, 0.75f) // 13x13
val normalIcon = icon // 16x16
```

## Layout

### Using IntelliJ Layout DSL

```kotlin
import com.intellij.ui.layout.panel

class MyDialog : DialogWrapper(project) {
    override fun createCenterPanel(): JComponent? {
        return panel {
            row("Name:") {
                textField()
            }
            row("Type:") {
                comboBox(arrayOf("Type 1", "Type 2"))
            }
            row {
                checkBox("Enable feature")
            }
            buttonsGroup {
                row {
                    radioButton("Option A", true)
                }
                row {
                    radioButton("Option B", false)
                }
            }
        }
    }
}
```

### Layout Guidelines

| Element | Guideline                   |
| ------- | --------------------------- |
| Margins | Use 12px margins            |
| Spacing | Use 6px between elements    |
| Labels  | Right-align, end with colon |
| Buttons | Right-align in dialogs      |

### BorderLayout Example

```kotlin
val panel = JPanel(BorderLayout()).apply {
    border = JBEmptyBorder(12)

    add(toolbar, BorderLayout.NORTH)
    add(content, BorderLayout.CENTER)
    add(statusBar, BorderLayout.SOUTH)
}
```

## Dialogs

### Dialog Structure

```kotlin
class MyDialog(project: Project) : DialogWrapper(project) {
    private val nameField = JTextField()

    init {
        title = "Dialog Title"
        setOKButtonText("Save")
        init()
    }

    override fun createCenterPanel(): JComponent? {
        return panel {
            row("Name:") {
                nameField(growX)
            }
        }
    }

    override fun getPreferredFocusedComponent(): JComponent = nameField
}
```

### Dialog Guidelines

1. **Clear title**: Describe the action
2. **Logical order**: Most important fields first
3. **Validation**: Show errors inline
4. **Default button**: OK should be default
5. **Cancel support**: Allow canceling changes

## Tool Windows

### Tool Window Guidelines

1. **Appropriate anchor**: Place where users expect
2. **Clear icon**: Use recognizable icon
3. **Toolbar actions**: Common actions in toolbar
4. **Context menu**: Right-click actions
5. **State persistence**: Remember user preferences

### Tool Window Structure

```kotlin
class MyToolWindowPanel(project: Project) : JPanel(BorderLayout()) {
    init {
        // Toolbar at top
        add(createToolbar(), BorderLayout.NORTH)

        // Main content
        add(JBScrollPane(content), BorderLayout.CENTER)

        // Optional status bar
        add(statusBar, BorderLayout.SOUTH)
    }

    private fun createToolbar(): JComponent {
        val group = DefaultActionGroup().apply {
            add(RefreshAction())
            addSeparator()
            add(SettingsAction())
        }

        return ActionManager.getInstance()
            .createActionToolbar("MyToolbar", group, true)
            .component
    }
}
```

## Actions

### Action Guidelines

1. **Clear text**: Use action verbs
2. **Mnemonic keys**: Add keyboard shortcuts
3. **Icons**: Use appropriate icons
4. **Context sensitivity**: Enable/disable appropriately
5. **Tooltips**: Provide helpful descriptions

### Action Presentation

```kotlin
class MyAction : AnAction("My Action", "Action description", AllIcons.Actions.Refresh) {
    override fun update(e: AnActionEvent) {
        val presentation = e.presentation

        // Set visibility
        presentation.isVisible = shouldShow()

        // Set enabled state
        presentation.isEnabled = canPerform()

        // Dynamic text
        presentation.text = getDynamicText()
    }
}
```

## Notifications

### Notification Guidelines

1. **Appropriate type**: Use correct severity
2. **Clear message**: Be concise
3. **Actionable**: Provide actions when relevant
4. **Don't overuse**: Reserve for important events

### Notification Types

| Type        | Use Case            |
| ----------- | ------------------- |
| INFORMATION | Success messages    |
| WARNING     | Potential issues    |
| ERROR       | Errors and failures |
| IDE_UPDATE  | IDE updates         |

```kotlin
fun showNotification(project: Project) {
    NotificationGroupManager.getInstance()
        .getNotificationGroup("MyPlugin")
        .createNotification(
            "Title",
            "Message content",
            NotificationType.INFORMATION
        )
        .addAction(object : NotificationAction("Action") {
            override fun actionPerformed(e: AnActionEvent, notification: Notification) {
                // Handle action
                notification.expire()
            }
        })
        .notify(project)
}
```

## Tables and Lists

### Table Guidelines

1. **Column headers**: Clear, concise labels
2. **Row height**: Comfortable for reading
3. **Sorting**: Enable where appropriate
4. **Selection**: Clear visual feedback
5. **Empty state**: Show helpful message

### Empty State

```kotlin
class EmptyStatePanel : JPanel(BorderLayout()) {
    init {
        add(JLabel("No items found").apply {
            horizontalAlignment = SwingConstants.CENTER
            foreground = JBColor.GRAY
        }, BorderLayout.CENTER)
    }
}

// Use in table panel
if (items.isEmpty()) {
    add(EmptyStatePanel(), BorderLayout.CENTER)
} else {
    add(JBScrollPane(table), BorderLayout.CENTER)
}
```

## Progress Indication

### Progress Guidelines

1. **Show progress**: For operations > 1 second
2. **Allow cancel**: Provide cancellation option
3. **Update text**: Describe current operation
4. **Accurate fraction**: Show real progress

```kotlin
ProgressManager.getInstance().run(object : Task.Backgroundable(project, "Processing") {
    override fun run(indicator: ProgressIndicator) {
        indicator.text = "Starting..."

        for (i in 1..100) {
            ProgressManager.checkCanceled()
            indicator.fraction = i / 100.0
            indicator.text = "Processing item $i"
            processItem(i)
        }
    }
})
```

## Accessibility

### Accessibility Guidelines

1. **Keyboard navigation**: All actions accessible via keyboard
2. **Screen readers**: Provide accessible descriptions
3. **Color contrast**: Ensure sufficient contrast
4. **Focus indicators**: Clear focus visibility

### Accessible Components

```kotlin
val button = JButton("Save").apply {
    toolTipText = "Save changes (Ctrl+S)"
    mnemonic = KeyEvent.VK_S
    accessibleContext.accessibleDescription = "Save the current file"
}
```

## Performance

### UI Performance Guidelines

1. **Avoid EDT blocking**: Long operations in background
2. **Lazy loading**: Load data on demand
3. **Debounce updates**: Batch rapid updates
4. **Virtual scrolling**: For large lists

### Background Operations

```kotlin
// Bad: Blocks EDT
fun loadData() {
    val data = fetchFromNetwork() // Blocks!
    updateUI(data)
}

// Good: Background task
fun loadData() {
    ProgressManager.getInstance().run(object : Task.Backgroundable(project, "Loading") {
        private var data: List<Item>? = null

        override fun run(indicator: ProgressIndicator) {
            data = fetchFromNetwork()
        }

        override fun onSuccess() {
            data?.let { updateUI(it) }
        }
    })
}
```

## Testing UI

### UI Testing Guidelines

1. **Test on different themes**: Light and dark
2. **Test scaling**: Different DPI settings
3. **Test accessibility**: Keyboard navigation
4. **Test responsiveness**: Different window sizes

## Related Topics

- [Tool Windows](tool-windows.md) - Tool window creation
- [Dialogs and Popups](dialogs-and-popups.md) - Dialog creation
- [Actions System](../01-core-concepts/actions-system.md) - Action registration

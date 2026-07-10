# Tool Windows

Tool windows are child windows of the main IDE window that provide specialized functionality.

## Overview

### What Are Tool Windows?

Tool windows are dockable panels that:

- Appear on the edges of the IDE window
- Provide specialized views (Project, Structure, Debug, etc.)
- Can be shown, hidden, and moved by users
- Support content switching via tabs

### Standard Tool Windows

| Tool Window | Description            |
| ----------- | ---------------------- |
| Project     | File system view       |
| Structure   | Code structure outline |
| Editor      | Main editing area      |
| Debug       | Debugging controls     |
| Problems    | Issues and warnings    |
| Terminal    | Integrated terminal    |

## Creating a Tool Window

### Tool Window Factory

```kotlin
class MyToolWindowFactory : ToolWindowFactory {
    override fun createToolWindowContent(
        project: Project,
        toolWindow: ToolWindow
    ) {
        // Create content
        val contentManager = toolWindow.contentManager

        // Create UI component
        val panel = createToolWindowPanel(project)

        // Add content
        val content = ContentFactory.getInstance()
            .createContent(panel, "", false)

        contentManager.addContent(content)
    }

    private fun createToolWindowPanel(project: Project): JPanel {
        return JPanel(BorderLayout()).apply {
            add(JLabel("My Tool Window"), BorderLayout.NORTH)
            add(createContent(project), BorderLayout.CENTER)
        }
    }

    override fun init(toolWindow: ToolWindow) {
        // Initialize tool window state
    }

    override fun shouldBeAvailable(project: Project): Boolean {
        // Control initial visibility
        return true
    }
}
```

### Registering a Tool Window

```xml
<extensions defaultExtensionNs="com.intellij">
    <toolWindow
        id="MyToolWindow"
        anchor="right"
        icon="/META-INF/pluginIcon.svg"
        factoryClass="com.example.MyToolWindowFactory"
        secondary="false"
        canCloseContents="true"/>
</extensions>
```

### Tool Window Attributes

| Attribute          | Description                   |
| ------------------ | ----------------------------- |
| `id`               | Unique identifier             |
| `anchor`           | Position: left, right, bottom |
| `icon`             | Icon path (16x16 or 13x13)    |
| `secondary`        | Less prominent position       |
| `canCloseContents` | Allow closing tabs            |
| `icon`             | Tool window icon              |

## Tool Window Content

### Adding Multiple Tabs

```kotlin
override fun createToolWindowContent(
    project: Project,
    toolWindow: ToolWindow
) {
    val contentFactory = ContentFactory.getInstance()
    val contentManager = toolWindow.contentManager

    // Add first tab
    val tab1 = contentFactory.createContent(
        createFirstPanel(project),
        "First Tab",
        false
    )
    contentManager.addContent(tab1)

    // Add second tab
    val tab2 = contentFactory.createContent(
        createSecondPanel(project),
        "Second Tab",
        false
    )
    contentManager.addContent(tab2)
}
```

### Dynamic Content

```kotlin
class DynamicToolWindowFactory : ToolWindowFactory {
    override fun createToolWindowContent(
        project: Project,
        toolWindow: ToolWindow
    ) {
        val contentManager = toolWindow.contentManager

        // Listen for changes
        project.messageBus.connect().subscribe(
            MyTopic.CHANGES,
            object : MyChangeListener {
                override fun onChanged() {
                    updateContent(contentManager, project)
                }
            }
        )

        // Initial content
        updateContent(contentManager, project)
    }

    private fun updateContent(
        contentManager: ContentManager,
        project: Project
    ) {
        // Remove old content
        contentManager.removeAllContents(true)

        // Add new content
        val content = ContentFactory.getInstance()
            .createContent(createPanel(project), "", false)
        contentManager.addContent(content)
    }
}
```

## Tool Window Actions

### Adding Toolbar Actions

```kotlin
override fun createToolWindowContent(
    project: Project,
    toolWindow: ToolWindow
) {
    val contentManager = toolWindow.contentManager

    // Create panel with toolbar
    val panel = JPanel(BorderLayout())
    panel.add(createToolbar(), BorderLayout.NORTH)
    panel.add(createContent(project), BorderLayout.CENTER)

    val content = ContentFactory.getInstance()
        .createContent(panel, "", false)
    contentManager.addContent(content)
}

private fun createToolbar(): JComponent {
    val actionManager = ActionManager.getInstance()
    val actionGroup = actionManager.getAction("MyToolWindowActions") as ActionGroup

    return ActionManager.getInstance()
        .createActionToolbar("MyToolWindow", actionGroup, true)
        .component
}
```

### Registering Tool Window Actions

```xml
<actions>
    <group id="MyToolWindowActions">
        <action id="MyToolWindow.Refresh"
                class="com.example.RefreshAction"
                icon="AllIcons.Actions.Refresh"/>
        <action id="MyToolWindow.Settings"
                class="com.example.SettingsAction"
                icon="AllIcons.General.GearPlain"/>
    </group>
</actions>
```

## Tool Window State

### Managing State

```kotlin
class MyToolWindowFactory : ToolWindowFactory, DumbAware {
    override fun createToolWindowContent(
        project: Project,
        toolWindow: ToolWindow
    ) {
        val state = ToolWindowState.getInstance(project)

        // Restore state
        val panel = createPanel(project, state)

        // Save state on changes
        panel.addPropertyChangeListener { event ->
            state.saveState(panel)
        }

        val content = ContentFactory.getInstance()
            .createContent(panel, "", false)
        toolWindow.contentManager.addContent(content)
    }
}

class ToolWindowState : PersistentStateComponent<ToolWindowState> {
    private var state = ToolWindowState()

    companion object {
        fun getInstance(project: Project): ToolWindowState {
            return project.getService(ToolWindowState::class.java)
        }
    }

    override fun getState(): ToolWindowState = state

    override fun loadState(state: ToolWindowState) {
        this.state = state
    }

    fun saveState(panel: JPanel) {
        // Save panel state
    }
}

data class ToolWindowState(
    var selectedTab: Int = 0,
    var filter: String = ""
)
```

## Tool Window Activation

### Conditional Availability

```kotlin
class ConditionalToolWindowFactory : ToolWindowFactory {
    override fun shouldBeAvailable(project: Project): Boolean {
        // Only show for specific project types
        return project.isMyProjectType()
    }

    override fun createToolWindowContent(
        project: Project,
        toolWindow: ToolWindow
    ) {
        // Create content
    }
}
```

### Programmatic Activation

```kotlin
fun activateToolWindow(project: Project) {
    val toolWindowManager = ToolWindowManager.getInstance(project)
    val toolWindow = toolWindowManager.getToolWindow("MyToolWindow")

    toolWindow?.activate {
        // Callback when activated
        println("Tool window activated")
    }
}

fun showToolWindow(project: Project) {
    val toolWindowManager = ToolWindowManager.getInstance(project)
    val toolWindow = toolWindowManager.getToolWindow("MyToolWindow")

    toolWindow?.setAvailable(true)
    toolWindow?.show()
}

fun hideToolWindow(project: Project) {
    val toolWindowManager = ToolWindowManager.getInstance(project)
    val toolWindow = toolWindowManager.getToolWindow("MyToolWindow")

    toolWindow?.hide()
}
```

## Tool Window Anchors

### Anchor Positions

| Anchor   | Position    | Typical Use        |
| -------- | ----------- | ------------------ |
| `left`   | Left edge   | Project, Structure |
| `right`  | Right edge  | Database, Gradle   |
| `bottom` | Bottom edge | Terminal, Problems |
| `top`    | Top edge    | Rarely used        |

### Secondary Tool Windows

```xml
<extensions defaultExtensionNs="com.intellij">
    <!-- Primary tool window -->
    <toolWindow
        id="PrimaryToolWindow"
        anchor="right"
        secondary="false"/>

    <!-- Secondary tool window (less prominent) -->
    <toolWindow
        id="SecondaryToolWindow"
        anchor="right"
        secondary="true"/>
</extensions>
```

## Best Practices

### UI Guidelines

1. **Use appropriate anchor**: Place tool windows where users expect them
2. **Provide meaningful icons**: Use clear, recognizable icons
3. **Support keyboard navigation**: Ensure accessibility
4. **Persist state**: Remember user preferences
5. **Handle dumb mode**: Disable or simplify during indexing

### Example: Complete Tool Window

```kotlin
class MyToolWindowFactory : ToolWindowFactory, DumbAware {
    override fun shouldBeAvailable(project: Project): Boolean = true

    override fun createToolWindowContent(
        project: Project,
        toolWindow: ToolWindow
    ) {
        val panel = MyToolWindowPanel(project)

        val content = ContentFactory.getInstance()
            .createContent(panel, "", false)
            .apply {
                preferredFocusableComponent = panel.preferredFocusComponent
            }

        toolWindow.contentManager.addContent(content)
    }
}

class MyToolWindowPanel(project: Project) : JPanel(BorderLayout()) {
    val preferredFocusComponent: JComponent

    init {
        val toolbar = createToolbar()
        add(toolbar, BorderLayout.NORTH)

        val content = createContent(project)
        add(content, BorderLayout.CENTER)

        preferredFocusComponent = content
    }

    private fun createToolbar(): JComponent {
        val group = DefaultActionGroup().apply {
            add(RefreshAction())
            addSeparator()
            add(SettingsAction())
        }

        return ActionManager.getInstance()
            .createActionToolbar("MyToolWindow", group, true)
            .component
    }

    private fun createContent(project: Project): JComponent {
        return JBScrollPane(JBList<String>()).apply {
            preferredSize = Dimension(200, 400)
        }
    }
}
```

## Related Topics

- [Actions System](../01-core-concepts/actions-system.md) - Action registration
- [Dialogs and Popups](dialogs-and-popups.md) - Dialog windows
- [Status Bar Widgets](status-bar-widgets.md) - Status bar integration

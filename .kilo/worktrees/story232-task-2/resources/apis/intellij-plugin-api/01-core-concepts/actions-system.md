# Actions System

Actions are the primary mechanism for plugins to add user-triggered functionality. The actions system provides a unified way to define, register, and handle commands from menus, toolbars, and keyboard shortcuts.

## Action Basics

### What Is an Action?

An action represents a user-triggered operation. Actions can be invoked from:

- Main menu items
- Toolbar buttons
- Context menus (right-click)
- Keyboard shortcuts
- Search everywhere

### Creating an Action

All actions extend `AnAction` and implement the `actionPerformed` method:

```kotlin
class MyAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        // Perform the action
    }
}
```

### Action Events

The `AnActionEvent` parameter provides context about the action invocation:

| Property       | Description                      |
| -------------- | -------------------------------- |
| `project`      | The current project, if any      |
| `dataContext`  | Provides access to platform data |
| `inputEvent`   | The triggering input event       |
| `place`        | Where the action was invoked     |
| `presentation` | The action's UI presentation     |

## Action Visibility and State

### The update() Method

Override `update()` to control action visibility and enabled state:

```kotlin
class MyAction : AnAction() {
    override fun update(e: AnActionEvent) {
        val project = e.project
        val file = e.getData(CommonDataKeys.VIRTUAL_FILE)

        // Enable only for specific file types
        e.presentation.isEnabled = project != null &&
            file?.extension == "txt"

        // Optionally hide the action
        e.presentation.isVisible = project != null
    }

    override fun actionPerformed(e: AnActionEvent) {
        // Action logic
    }
}
```

### Presentation Properties

The `Presentation` object controls how the action appears:

| Property      | Description                       |
| ------------- | --------------------------------- |
| `isEnabled`   | Whether the action can be invoked |
| `isVisible`   | Whether the action is shown       |
| `text`        | The display text                  |
| `description` | Tooltip text                      |
| `icon`        | The action icon                   |

### Performance Considerations

The `update()` method is called frequently. Keep it fast:

```kotlin
// WRONG: Expensive operation in update
override fun update(e: AnActionEvent) {
    val file = e.getData(CommonDataKeys.VIRTUAL_FILE)
    e.presentation.isEnabled = file?.let {
        calculateExpensiveCondition(it)
    } ?: false
}

// CORRECT: Cache expensive results
private val conditionCache = ConcurrentHashMap<VirtualFile, Boolean>()

override fun update(e: AnActionEvent) {
    val file = e.getData(CommonDataKeys.VIRTUAL_FILE)
    e.presentation.isEnabled = file?.let {
        conditionCache.getOrPut(it) { calculateExpensiveCondition(it) }
    } ?: false
}
```

## Registering Actions

### Declarative Registration

Register actions in `plugin.xml`:

```xml
<actions>
    <action id="com.example.MyAction"
            class="com.example.MyAction"
            text="My Action"
            description="Performs my action">
        <add-to-group group-id="ToolsMenu" anchor="first"/>
        <keyboard-shortcut keymap="$default" first-keystroke="ctrl alt M"/>
    </action>
</actions>
```

### Action Attributes

| Attribute     | Description                      |
| ------------- | -------------------------------- |
| `id`          | Unique identifier for the action |
| `class`       | Fully qualified class name       |
| `text`        | Display text                     |
| `description` | Tooltip description              |
| `icon`        | Icon path or class reference     |

### Adding to Groups

Actions are placed in groups using `<add-to-group>`:

```xml
<add-to-group group-id="ToolsMenu" anchor="first"/>
```

Common group IDs:

- `ToolsMenu` - The Tools menu
- `EditorPopupMenu` - Editor context menu
- `ProjectViewPopupMenu` - Project view context menu
- `MainToolBar` - Main toolbar
- `NavbarToolBar` - Navigation bar toolbar

Anchor values:

- `first` - At the beginning
- `last` - At the end
- `before` - Before a specific action
- `after` - After a specific action

### Programmatic Registration

Actions can be registered at runtime:

```kotlin
val actionManager = ActionManager.getInstance()
val action = MyAction()
actionManager.registerAction("com.example.MyAction", action)

val group = actionManager.getAction("ToolsMenu") as DefaultActionGroup
group.add(action)
```

## Action Groups

Groups organize related actions into menus and toolbars.

### Creating a Group

```xml
<actions>
    <group id="com.example.MyGroup" text="My Plugin" popup="true">
        <action id="com.example.Action1" class="com.example.Action1" text="Action 1"/>
        <action id="com.example.Action2" class="com.example.Action2" text="Action 2"/>
        <separator/>
        <action id="com.example.Action3" class="com.example.Action3" text="Action 3"/>
        <add-to-group group-id="ToolsMenu" anchor="last"/>
    </group>
</actions>
```

### Group Types

| Type                 | Description                       |
| -------------------- | --------------------------------- |
| `DefaultActionGroup` | Standard group with fixed actions |
| `CompactActionGroup` | Compact toolbar group             |
| `PopupActionGroup`   | Shows as a popup menu             |

### Dynamic Groups

Groups can provide actions dynamically:

```kotlin
class MyDynamicGroup : ActionGroup() {
    override fun getChildren(e: AnActionEvent?): Array<AnAction> {
        return if (e?.project != null) {
            arrayOf(Action1(), Action2())
        } else {
            emptyArray()
        }
    }
}
```

## Keyboard Shortcuts

### Defining Shortcuts

```xml
<keyboard-shortcut keymap="$default" first-keystroke="ctrl alt M"/>
<keyboard-shortcut keymap="$default" first-keystroke="ctrl alt M" second-keystroke="shift N"/>
```

### Keymap Scopes

- `$default` - Default keymap (used as base for all keymaps)
- `$primary` - Primary keymap for the current IDE
- Specific keymap names like `Mac OS X`, `Windows`, `Eclipse`

### Removing Shortcuts

To remove a shortcut from another action:

```xml
<keyboard-shortcut keymap="$default" first-keystroke="ctrl S" remove="true"/>
```

### Mouse Shortcuts

Actions can respond to mouse clicks:

```xml
<mouse-shortcut keymap="$default" keystroke="button1 doubleClick"/>
```

## Action Context

### Data Context

The data context provides access to platform data:

```kotlin
override fun actionPerformed(e: AnActionEvent) {
    val project = e.getData(CommonDataKeys.PROJECT)
    val file = e.getData(CommonDataKeys.VIRTUAL_FILE)
    val editor = e.getData(CommonDataKeys.EDITOR)
    val psiFile = e.getData(CommonDataKeys.PSI_FILE)
    val element = e.getData(CommonDataKeys.PSI_ELEMENT)

    // Use the data
}
```

### Common Data Keys

| Key            | Description                        |
| -------------- | ---------------------------------- |
| `PROJECT`      | Current project                    |
| `VIRTUAL_FILE` | Selected file                      |
| `EDITOR`       | Current editor                     |
| `PSI_FILE`     | PSI representation of current file |
| `PSI_ELEMENT`  | PSI element at cursor              |
| `NAVIGATABLE`  | Navigatable element                |
| `MODULE`       | Current module                     |

### Custom Data Keys

Plugins can define custom data keys:

```kotlin
companion object {
    val MY_DATA_KEY = DataKey.create<MyData>("myData")
}
```

## Toggle Actions

Toggle actions represent on/off states:

```kotlin
class MyToggleAction : ToggleAction("My Toggle") {
    override fun isSelected(e: AnActionEvent): Boolean {
        val settings = service<MySettings>()
        return settings.isEnabled
    }

    override fun setSelected(e: AnActionEvent, state: Boolean) {
        val settings = service<MySettings>()
        settings.isEnabled = state
    }
}
```

## Action Templates

Templates allow reusing action configuration:

```xml
<actions>
    <action id="com.example.TemplateAction"
            class="com.example.MyAction"
            text="Template"
            icon="AllIcons.Actions.Execute"/>

    <action id="com.example.DerivedAction"
            class="com.example.MyAction"
            text="Derived"
            use-shortcut-of="com.example.TemplateAction"/>
</actions>
```

## Background Actions

Long-running actions should run in the background:

```kotlin
class MyLongAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return

        ProgressManager.getInstance().run(
            object : Task.Backgroundable(project, "Processing", true) {
                override fun run(indicator: ProgressIndicator) {
                    indicator.text = "Working..."
                    indicator.fraction = 0.0

                    // Long-running work
                    for (i in 1..100) {
                        if (indicator.isCanceled) return
                        indicator.fraction = i / 100.0
                        Thread.sleep(50)
                    }
                }

                override fun onSuccess() {
                    NotificationUtil.showInfo(project, "Complete!")
                }
            }
        )
    }
}
```

## Related Topics

- [Action Groups](../07-ui-components/action-groups.md) - Detailed group configuration
- [Tool Windows](../07-ui-components/tool-windows.md) - Adding tool windows
- [Keymap Management](../10-advanced-topics/external-tools.md) - Custom keymap integration

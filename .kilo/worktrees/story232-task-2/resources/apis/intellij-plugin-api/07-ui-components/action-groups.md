# Action Groups

Action groups organize actions into menus, toolbars, and popup menus.

## Overview

### What Are Action Groups?

Action groups provide:

- Logical organization of related actions
- Hierarchical menu structures
- Toolbar button groups
- Context-sensitive popup menus

### Group Types

| Type               | Description                |
| ------------------ | -------------------------- |
| DefaultActionGroup | Static group of actions    |
| CompactActionGroup | Compact toolbar group      |
| PopupActionGroup   | Popup menu group           |
| ToggleActionGroup  | Mutually exclusive actions |

## Creating Action Groups

### Default Action Group

```kotlin
class MyActionGroup : DefaultActionGroup() {
    init {
        add(Action1())
        add(Action2())
        addSeparator()
        add(Action3())
    }
}
```

### Registering Action Groups

```xml
<actions>
    <group id="MyGroup"
           class="com.example.MyActionGroup"
           text="My Actions"
           description="My action group"
           icon="AllIcons.Actions.Group">
        <add-to-group group-id="MainMenu" anchor="last"/>
    </group>
</actions>
```

### Inline Group Definition

```xml
<actions>
    <group id="MyInlineGroup" text="My Actions">
        <action id="Action1" class="com.example.Action1"/>
        <action id="Action2" class="com.example.Action2"/>
        <separator/>
        <action id="Action3" class="com.example.Action3"/>
    </group>
</actions>
```

## Adding to Existing Groups

### Adding to Main Menu

```xml
<actions>
    <group id="MyMenu" text="My Plugin">
        <action id="MyAction" class="com.example.MyAction"/>
        <add-to-group group-id="MainMenu" anchor="last"/>
    </group>
</actions>
```

### Adding to Editor Popup

```xml
<actions>
    <action id="MyEditorAction" class="com.example.MyEditorAction">
        <add-to-group group-id="EditorPopupMenu" anchor="first"/>
    </action>
</actions>
```

### Adding to Toolbar

```xml
<actions>
    <group id="MyToolbarGroup">
        <action id="MyToolbarAction" class="com.example.MyToolbarAction"/>
        <add-to-group group-id="MainToolBar" anchor="after" relative-to-action="HelpTopics"/>
    </group>
</actions>
```

## Standard Group IDs

### Main Menu Groups

| Group ID          | Description   |
| ----------------- | ------------- |
| `MainMenu`        | Main menu bar |
| `FileMenu`        | File menu     |
| `EditMenu`        | Edit menu     |
| `ViewMenu`        | View menu     |
| `NavigateMenu`    | Navigate menu |
| `CodeMenu`        | Code menu     |
| `RefactoringMenu` | Refactor menu |
| `ToolsMenu`       | Tools menu    |

### Editor Groups

| Group ID                | Description             |
| ----------------------- | ----------------------- |
| `EditorPopupMenu`       | Editor right-click menu |
| `EditorTabPopupMenu`    | Editor tab menu         |
| `EditorGutterPopupMenu` | Gutter menu             |

### Project View Groups

| Group ID                          | Description       |
| --------------------------------- | ----------------- |
| `ProjectViewPopupMenu`            | Project view menu |
| `ProjectViewPopupMenuModifyGroup` | Modify group      |
| `ProjectViewPopupMenuRunGroup`    | Run group         |

## Dynamic Action Groups

### Conditionally Visible Actions

```xml
<actions>
    <action id="MyConditionalAction" class="com.example.MyConditionalAction">
        <add-to-group group-id="EditorPopupMenu"/>
    </action>
</actions>
```

```kotlin
class MyConditionalAction : AnAction() {
    override fun update(e: AnActionEvent) {
        val project = e.project
        val file = e.getData(CommonDataKeys.PSI_FILE)

        e.presentation.isVisible = file?.fileType == MyFileType
        e.presentation.isEnabled = project != null
    }

    override fun actionPerformed(e: AnActionEvent) {
        // Action logic
    }
}
```

### Dynamic Group Content

```kotlin
class MyDynamicActionGroup : ActionGroup() {
    override fun getChildren(e: AnActionEvent?): Array<AnAction> {
        val actions = mutableListOf<AnAction>()

        // Add actions dynamically
        for (item in getDynamicItems()) {
            actions.add(MyItemAction(item))
        }

        return actions.toTypedArray()
    }

    override fun update(e: AnActionEvent) {
        e.presentation.isVisible = hasDynamicItems()
    }
}
```

### Lazy-Loaded Actions

```kotlin
class LazyActionGroup : ActionGroup() {
    private var children: Array<AnAction>? = null

    override fun getChildren(e: AnActionEvent?): Array<AnAction> {
        if (children == null) {
            children = loadChildren()
        }
        return children!!
    }

    private fun loadChildren(): Array<AnAction> {
        // Load actions on demand
        return arrayOf(
            Action1(),
            Action2(),
            Action3()
        )
    }
}
```

## Popup Menus

### Creating Popup Menus

```kotlin
fun showPopupMenu(component: JComponent) {
    val group = DefaultActionGroup().apply {
        add(CutAction())
        add(CopyAction())
        add(PasteAction())
        addSeparator()
        add(DeleteAction())
    }

    val popup = ActionManager.getInstance()
        .createActionPopupMenu("MyPopup", group)

    popup.component.show(component, 0, component.height)
}
```

### Context-Sensitive Popup

```kotlin
class ContextPopupAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val element = e.getData(CommonDataKeys.PSI_ELEMENT) ?: return

        val group = createContextGroup(element)

        val popup = JBPopupFactory.getInstance()
            .createActionGroupPopup(
                "Actions",
                group,
                e.dataContext,
                JBPopupFactory.ActionSelectionAid.SPEEDSEARCH,
                true
            )

        popup.showInBestPositionFor(e.dataContext)
    }

    private fun createContextGroup(element: PsiElement): ActionGroup {
        return DefaultActionGroup().apply {
            when (element) {
                is PsiClass -> {
                    add(GoToClassAction())
                    add(FindUsagesAction())
                }
                is PsiMethod -> {
                    add(GoToImplementationAction())
                    add(FindUsagesAction())
                }
                else -> {
                    add(FindUsagesAction())
                }
            }
        }
    }
}
```

## Toolbar Actions

### Creating Toolbar

```kotlin
fun createToolbar(): JComponent {
    val group = DefaultActionGroup().apply {
        add(RefreshAction())
        addSeparator()
        add(SettingsAction())
    }

    return ActionManager.getInstance()
        .createActionToolbar("MyToolbar", group, true)
        .component
}
```

### Toolbar with Toggle Actions

```kotlin
class MyToggleAction : ToggleAction("Toggle", "Toggle description", AllIcons.Actions.Checked) {
    private var selected = false

    override fun isSelected(e: AnActionEvent): Boolean = selected

    override fun setSelected(e: AnActionEvent, state: Boolean) {
        selected = state
        // Handle state change
    }
}
```

## Action Constraints

### Anchor Positions

| Anchor   | Description            |
| -------- | ---------------------- |
| `first`  | First in group         |
| `last`   | Last in group          |
| `before` | Before relative action |
| `after`  | After relative action  |

### Example with Anchors

```xml
<actions>
    <action id="MyAction" class="com.example.MyAction">
        <add-to-group group-id="EditorPopupMenu" anchor="before" relative-to-action="CopyReference"/>
    </action>
</actions>
```

## Nested Groups

### Creating Nested Groups

```xml
<actions>
    <group id="MyNestedGroup" text="Nested Actions">
        <action id="Action1" class="com.example.Action1"/>

        <group id="SubGroup" text="Sub Actions">
            <action id="SubAction1" class="com.example.SubAction1"/>
            <action id="SubAction2" class="com.example.SubAction2"/>
        </group>

        <action id="Action2" class="com.example.Action2"/>
    </group>
</actions>
```

### Programmatic Nested Groups

```kotlin
class NestedActionGroup : DefaultActionGroup() {
    init {
        add(Action1())

        add(DefaultActionGroup("Sub Actions", true).apply {
            add(SubAction1())
            add(SubAction2())
        })

        add(Action2())
    }
}
```

## Best Practices

### Action Group Guidelines

1. **Group related actions**: Keep logically related actions together
2. **Use separators**: Separate different action categories
3. **Limit group size**: Avoid overwhelming users
4. **Use icons consistently**: Maintain visual consistency
5. **Provide descriptions**: Help users understand actions

### Example: Complete Action Group

```kotlin
class MyCompleteActionGroup : DefaultActionGroup() {
    init {
        templatePresentation.text = "My Plugin"
        templatePresentation.icon = AllIcons.General.Plugin

        // Primary actions
        add(OpenAction())
        add(SaveAction())
        addSeparator()

        // Edit actions
        add(EditAction())
        add(DeleteAction())
        addSeparator()

        // Tools submenu
        add(DefaultActionGroup("Tools", true).apply {
            add(Tool1Action())
            add(Tool2Action())
            addSeparator()
            add(ConfigureToolsAction())
        })
        addSeparator()

        // Help
        add(HelpAction())
    }
}
```

## Related Topics

- [Actions System](../01-core-concepts/actions-system.md) - Action basics
- [Tool Windows](tool-windows.md) - Tool window toolbars
- [UI Guidelines](ui-guidelines.md) - Design best practices

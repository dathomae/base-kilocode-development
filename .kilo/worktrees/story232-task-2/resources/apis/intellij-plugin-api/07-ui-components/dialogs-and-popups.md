# Dialogs and Popups

Dialogs and popups provide user interaction for configuration, confirmation, and information display.

## Overview

### Types of Dialogs

| Type             | Description                     |
| ---------------- | ------------------------------- |
| Modal Dialog     | Blocks interaction until closed |
| Non-modal Dialog | Allows interaction with IDE     |
| Popup            | Lightweight, context-sensitive  |
| Notification     | Non-blocking messages           |

## Dialogs

### Basic Dialog

```kotlin
class MyDialog(project: Project) : DialogWrapper(project) {
    private val textField = JTextField()

    init {
        title = "My Dialog"
        init()
    }

    override fun createCenterPanel(): JComponent? {
        return panel {
            row("Name:") {
                textField(growX)
            }
        }
    }

    override fun doOKAction() {
        val name = textField.text
        if (name.isBlank()) {
            showValidationError("Name cannot be empty")
            return
        }
        super.doOKAction()
    }

    private fun showValidationError(message: String) {
        setErrorText(message)
    }

    fun getName(): String = textField.text
}

// Usage
val dialog = MyDialog(project)
if (dialog.showAndGet()) {
    val name = dialog.getName()
    // Process result
}
```

### Dialog with Validation

```kotlin
class ValidatedDialog(project: Project) : DialogWrapper(project) {
    private val nameField = JTextField()
    private val emailField = JTextField()

    init {
        title = "User Information"
        setOKButtonText("Save")
        init()
    }

    override fun createCenterPanel(): JComponent? {
        return panel {
            row("Name:") {
                nameField(growX)
                    .withValidationOnInput { validateName() }
            }
            row("Email:") {
                emailField(growX)
                    .withValidationOnInput { validateEmail() }
            }
        }
    }

    private fun validateName(): ValidationInfo? {
        val name = nameField.text
        return when {
            name.isBlank() -> ValidationInfo("Name is required", nameField)
            name.length < 2 -> ValidationInfo("Name too short", nameField)
            else -> null
        }
    }

    private fun validateEmail(): ValidationInfo? {
        val email = emailField.text
        return when {
            email.isBlank() -> ValidationInfo("Email is required", emailField)
            !email.contains("@") -> ValidationInfo("Invalid email format", emailField)
            else -> null
        }
    }
}
```

### Dialog with Multiple Actions

```kotlin
class MultiActionDialog(project: Project) : DialogWrapper(project) {
    init {
        title = "Confirm Action"
        setOKButtonText("Save")
        setCancelButtonText("Don't Save")

        // Add custom action
        myHelpAction.putValue(Action.NAME, "Help")
    }

    override fun createActions(): Array<Action> {
        return arrayOf(
            okAction,
            cancelAction,
            helpAction
        )
    }

    override fun createCenterPanel(): JComponent? {
        return JLabel("Do you want to save changes?")
    }

    override fun doHelpAction() {
        // Show help
    }
}
```

### File Chooser Dialog

```kotlin
fun selectFile(project: Project): VirtualFile? {
    val descriptor = FileChooserDescriptor(
        /* chooseFiles = */ true,
        /* chooseFolders = */ false,
        /* chooseJars = */ false,
        /* chooseJarsAsFiles = */ false,
        /* chooseJarContents = */ false,
        /* chooseMultiple = */ false
    ).apply {
        title = "Select File"
        description = "Select a file to process"
    }

    return FileChooser.chooseFile(descriptor, project, null)
}

fun selectDirectory(project: Project): VirtualFile? {
    val descriptor = FileChooserDescriptorFactory
        .createSingleFolderDescriptor()
        .apply {
            title = "Select Directory"
            description = "Select a directory"
        }

    return FileChooser.chooseFile(descriptor, project, null)
}
```

## Popups

### Simple Popup

```kotlin
fun showSimplePopup(component: JComponent) {
    val popup = JBPopupFactory.getInstance()
        .createMessage("This is a popup message")

    popup.showUnderneathOf(component)
}
```

### List Popup

```kotlin
fun showListPopup(project: Project, component: JComponent) {
    val items = listOf(
        "Option 1",
        "Option 2",
        "Option 3"
    )

    val popup = JBPopupFactory.getInstance()
        .createPopupChooserBuilder(items)
        .setTitle("Select Option")
        .setItemChosenCallback { selectedItem ->
            // Handle selection
            println("Selected: $selectedItem")
        }
        .createPopup()

    popup.showUnderneathOf(component)
}
```

### Action Popup

```kotlin
fun showActionPopup(component: JComponent) {
    val group = DefaultActionGroup().apply {
        add(Action1())
        add(Action2())
        addSeparator()
        add(Action3())
    }

    val popup = JBPopupFactory.getInstance()
        .createActionGroupPopup(
            "Actions",
            group,
            DataManager.getInstance().getDataContext(component),
            JBPopupFactory.ActionSelectionAid.SPEEDSEARCH,
            true
        )

    popup.showUnderneathOf(component)
}
```

### Custom Popup Content

```kotlin
fun showCustomPopup(component: JComponent) {
    val panel = JPanel(BorderLayout()).apply {
        add(JLabel("Custom Content"), BorderLayout.NORTH)
        add(JBScrollPane(JBList(listOf("Item 1", "Item 2"))), BorderLayout.CENTER)
        preferredSize = Dimension(300, 200)
    }

    val popup = JBPopupFactory.getInstance()
        .createComponentPopupBuilder(panel, null)
        .setTitle("Custom Popup")
        .setResizable(true)
        .setMovable(true)
        .setRequestFocus(true)
        .createPopup()

    popup.showUnderneathOf(component)
}
```

## Notifications

### Showing Notifications

```kotlin
fun showInfoNotification(project: Project, title: String, content: String) {
    val notification = NotificationGroupManager.getInstance()
        .getNotificationGroup("MyPlugin Notifications")
        .createNotification(title, content, NotificationType.INFORMATION)

    notification.notify(project)
}

fun showErrorNotification(project: Project, title: String, content: String) {
    val notification = NotificationGroupManager.getInstance()
        .getNotificationGroup("MyPlugin Notifications")
        .createNotification(title, content, NotificationType.ERROR)

    notification.notify(project)
}
```

### Registering Notification Group

```xml
<extensions defaultExtensionNs="com.intellij">
    <notificationGroup
        id="MyPlugin Notifications"
        displayType="BALLOON"
        key="notification.group.name"/>
</extensions>
```

### Notification with Actions

```kotlin
fun showNotificationWithAction(project: Project) {
    val notification = NotificationGroupManager.getInstance()
        .getNotificationGroup("MyPlugin Notifications")
        .createNotification(
            "Action Required",
            "Click to perform action",
            NotificationType.WARNING
        )
        .addAction(object : NotificationAction("Perform Action") {
            override fun actionPerformed(
                e: AnActionEvent,
                notification: Notification
            ) {
                // Perform action
                notification.expire()
            }
        })

    notification.notify(project)
}
```

## Input Dialogs

### Simple Input Dialog

```kotlin
fun getUserName(): String? {
    return Messages.showInputDialog(
        project,
        "Enter your name:",
        "Name Input",
        Messages.getQuestionIcon()
    )
}
```

### Input with Validation

```kotlin
fun getValidatedInput(project: Project): String? {
    return Messages.showInputDialog(
        project,
        "Enter a value (min 3 characters):",
        "Validated Input",
        Messages.getQuestionIcon(),
        "",
        InputValidator { text ->
            text != null && text.length >= 3
        }
    )
}
```

### Password Input

```kotlin
fun getPassword(): String? {
    return Messages.showPasswordDialog(
        project,
        "Enter your password:",
        "Password"
    )
}
```

## Confirmation Dialogs

### Yes/No Dialog

```kotlin
fun confirmAction(project: Project): Boolean {
    val result = Messages.showYesNoDialog(
        project,
        "Are you sure you want to proceed?",
        "Confirm Action",
        Messages.getQuestionIcon()
    )

    return result == Messages.YES
}
```

### Yes/No/Cancel Dialog

```kotlin
fun confirmWithCancel(project: Project): Int {
    return Messages.showYesNoCancelDialog(
        project,
        "Save changes?",
        "Save",
        "Save",
        "Don't Save",
        "Cancel",
        Messages.getQuestionIcon()
    )
}
```

### Custom Options Dialog

```kotlin
fun showCustomOptions(project: Project): Int {
    val options = arrayOf("Overwrite", "Merge", "Skip")

    return Messages.showDialog(
        project,
        "File already exists. What would you like to do?",
        "File Conflict",
        options,
        0, // Default selection
        Messages.getWarningIcon()
    )
}
```

## Best Practices

### Dialog Guidelines

1. **Use appropriate modality**: Modal for critical decisions, non-modal for reference
2. **Provide clear actions**: Label buttons clearly
3. **Validate input**: Show errors before closing
4. **Remember size and position**: Persist dialog state
5. **Support keyboard navigation**: Ensure accessibility

### Example: Complete Dialog

```kotlin
class ConfigurationDialog(
    project: Project,
    private val config: MyConfiguration
) : DialogWrapper(project) {
    private val nameField = JTextField(config.name)
    private val enabledCheckbox = JCheckBox("Enabled", config.enabled)
    private val pathField = TextFieldWithBrowseButton()

    init {
        title = "Configuration"
        pathField.text = config.path

        // Add browse action
        pathField.addBrowseFolderListener(
            "Select Path",
            null,
            project,
            FileChooserDescriptorFactory.createSingleFolderDescriptor()
        )

        init()
    }

    override fun createCenterPanel(): JComponent? {
        return panel {
            row("Name:") {
                nameField(growX)
                    .withValidationOnInput { validateName() }
            }
            row {
                enabledCheckbox()
            }
            row("Path:") {
                pathField(growX)
                    .withValidationOnInput { validatePath() }
            }
        }
    }

    override fun getPreferredFocusedComponent(): JComponent = nameField

    private fun validateName(): ValidationInfo? {
        val name = nameField.text
        return when {
            name.isBlank() -> ValidationInfo("Name is required", nameField)
            name.length > 50 -> ValidationInfo("Name too long", nameField)
            else -> null
        }
    }

    private fun validatePath(): ValidationInfo? {
        val path = pathField.text
        return when {
            path.isBlank() -> ValidationInfo("Path is required", pathField)
            !File(path).exists() -> ValidationInfo("Path does not exist", pathField)
            else -> null
        }
    }

    override fun doOKAction() {
        config.name = nameField.text
        config.enabled = enabledCheckbox.isSelected
        config.path = pathField.text
        super.doOKAction()
    }
}
```

## Related Topics

- [Tool Windows](tool-windows.md) - Dockable panels
- [Actions System](../01-core-concepts/actions-system.md) - Action registration
- [UI Guidelines](ui-guidelines.md) - Design best practices

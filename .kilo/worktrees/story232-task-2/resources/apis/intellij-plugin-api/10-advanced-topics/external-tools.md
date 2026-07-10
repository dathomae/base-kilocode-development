# External Tools

External tools integration allows plugins to run and manage external processes.

## Overview

### What Are External Tools?

External tools are:

- Command-line utilities
- Build tools
- Linters and formatters
- Custom scripts

### Integration Points

| Integration      | Description          |
| ---------------- | -------------------- |
| Tool Window      | Output display       |
| Actions          | Menu/toolbar actions |
| Before Run Tasks | Pre-execution tasks  |
| File Watchers    | Automatic execution  |

## Running External Processes

### Basic Process Execution

```kotlin
fun runExternalTool(
    executable: String,
    arguments: List<String>,
    workingDirectory: String
): String {
    val process = ProcessBuilder()
        .command(listOf(executable) + arguments)
        .directory(File(workingDirectory))
        .redirectErrorStream(true)
        .start()

    val output = process.inputStream.bufferedReader().readText()
    process.waitFor()

    return output
}
```

### Using GeneralCommandLine

```kotlin
fun runCommand(): String {
    val commandLine = GeneralCommandLine()
        .withExePath("mytool")
        .withParameters("--format", "json")
        .withWorkDirectory(project.basePath)

    val processHandler = OSProcessHandler(commandLine)
    val output = StringBuilder()

    processHandler.addProcessListener(object : ProcessAdapter() {
        override fun onTextAvailable(event: ProcessEvent, outputType: Key<*>) {
            output.append(event.text)
        }
    })

    processHandler.startNotify()
    processHandler.waitFor()

    return output.toString()
}
```

### Capturing Output

```kotlin
class ProcessOutputCapture {
    private val stdout = StringBuilder()
    private val stderr = StringBuilder()

    fun capture(processHandler: ProcessHandler) {
        processHandler.addProcessListener(object : ProcessAdapter() {
            override fun onTextAvailable(event: ProcessEvent, outputType: Key<*>) {
                if (outputType == ProcessOutputTypes.STDOUT) {
                    stdout.append(event.text)
                } else if (outputType == ProcessOutputTypes.STDERR) {
                    stderr.append(event.text)
                }
            }
        })
    }

    fun getStdout(): String = stdout.toString()
    fun getStderr(): String = stderr.toString()
}
```

## Tool Output Display

### Console View

```kotlin
class MyToolConsole(project: Project) {
    private val consoleView: ConsoleView = TextConsoleBuilderFactory.getInstance()
        .createBuilder(project)
        .build()

    fun attachToProcess(processHandler: ProcessHandler) {
        consoleView.attachToProcess(processHandler)
    }

    fun print(text: String, attributes: ConsoleViewContentType) {
        consoleView.print(text, attributes)
    }

    fun clear() {
        consoleView.clear()
    }

    fun getComponent(): JComponent = consoleView.component
}
```

### Using ConsoleView

```kotlin
class MyToolAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return

        // Create console
        val console = MyToolConsole(project)

        // Create process
        val commandLine = GeneralCommandLine("mytool")
        val processHandler = OSProcessHandler(commandLine)

        // Attach console
        console.attachToProcess(processHandler)

        // Show in tool window
        showInToolWindow(project, console.getComponent())

        // Start process
        processHandler.startNotify()
    }
}
```

## Tool Configuration

### Configurable Tool Settings

```kotlin
class MyToolSettings : PersistentStateComponent<MyToolSettings.State> {
    var executablePath: String = ""
    var defaultArguments: String = ""
    var workingDirectory: String = ""

    data class State(
        var executablePath: String = "",
        var defaultArguments: String = "",
        var workingDirectory: String = ""
    )

    override fun getState(): State {
        return State(executablePath, defaultArguments, workingDirectory)
    }

    override fun loadState(state: State) {
        executablePath = state.executablePath
        defaultArguments = state.defaultArguments
        workingDirectory = state.workingDirectory
    }

    companion object {
        fun getInstance(): MyToolSettings {
            return ApplicationManager.getApplication()
                .getService(MyToolSettings::class.java)
        }
    }
}
```

### Settings UI

```kotlin
class MyToolConfigurable : Configurable {
    private val executableField = TextFieldWithBrowseButton()
    private val argumentsField = JTextField()

    override fun createComponent(): JComponent? {
        executableField.addBrowseFolderListener(
            "Select Executable",
            null,
            null,
            FileChooserDescriptorFactory.createSingleFileDescriptor()
        )

        return panel {
            row("Executable:") {
                executableField(growX)
            }
            row("Arguments:") {
                argumentsField(growX)
            }
        }
    }

    override fun isModified(): Boolean {
        val settings = MyToolSettings.getInstance()
        return executableField.text != settings.executablePath ||
               argumentsField.text != settings.defaultArguments
    }

    override fun apply() {
        val settings = MyToolSettings.getInstance()
        settings.executablePath = executableField.text
        settings.defaultArguments = argumentsField.text
    }

    override fun reset() {
        val settings = MyToolSettings.getInstance()
        executableField.text = settings.executablePath
        argumentsField.text = settings.defaultArguments
    }

    override fun getDisplayName() = "My Tool"
}
```

## File Watcher Integration

### Running on File Changes

```kotlin
class MyFileWatcherListener : BulkFileListener {
    override fun after(events: MutableList<out VFileEvent>) {
        for (event in events) {
            if (event is VFileContentChangeEvent) {
                val file = event.file
                if (shouldProcess(file)) {
                    runToolOnFile(file)
                }
            }
        }
    }

    private fun shouldProcess(file: VirtualFile): Boolean {
        return file.extension == "my"
    }

    private fun runToolOnFile(file: VirtualFile) {
        val settings = MyToolSettings.getInstance()
        val commandLine = GeneralCommandLine()
            .withExePath(settings.executablePath)
            .withParameters(file.path)

        val processHandler = OSProcessHandler(commandLine)
        processHandler.startNotify()
    }
}
```

### Registering Listener

```kotlin
class MyProjectService(val project: Project) : Disposable {
    private val connection = project.messageBus.connect()

    init {
        connection.subscribe(VirtualFileManager.VFS_CHANGES, MyFileWatcherListener())
    }

    override fun dispose() {
        connection.disconnect()
    }
}
```

## Before Run Tasks

### Tool Before Run Task

```kotlin
class MyToolBeforeRunTask : BeforeRunTask<MyToolBeforeRunTask>(MyToolTaskProvider.ID) {
    var enabled: Boolean = true
}

class MyToolTaskProvider : BeforeRunTaskProvider<MyToolBeforeRunTask>() {
    companion object {
        val ID = Key.create<MyToolBeforeRunTask>("MyToolTask")
    }

    override fun getId() = ID
    override fun getName() = "Run My Tool"

    override fun createTask(runConfiguration: RunConfiguration): MyToolBeforeRunTask? {
        return MyToolBeforeRunTask()
    }

    override fun executeTask(
        configuration: RunConfiguration,
        task: MyToolBeforeRunTask,
        env: ExecutionEnvironment
    ): Boolean {
        val settings = MyToolSettings.getInstance()
        val commandLine = GeneralCommandLine()
            .withExePath(settings.executablePath)
            .withParameters(settings.defaultArguments)

        val processHandler = OSProcessHandler(commandLine)
        processHandler.startNotify()
        processHandler.waitFor()

        return processHandler.exitCode == 0
    }
}
```

## Error Handling

### Handling Process Errors

```kotlin
fun runWithErrorHandling(): Result<String> {
    return try {
        val commandLine = GeneralCommandLine("mytool")
        val processHandler = OSProcessHandler(commandLine)
        val output = ProcessOutputCapture()

        output.capture(processHandler)
        processHandler.startNotify()
        processHandler.waitFor()

        if (processHandler.exitCode == 0) {
            Result.success(output.getStdout())
        } else {
            Result.failure(Exception(output.getStderr()))
        }
    } catch (e: Exception) {
        Result.failure(e)
    }
}
```

### User Notifications

```kotlin
fun notifyError(project: Project, message: String) {
    NotificationGroupManager.getInstance()
        .getNotificationGroup("MyTool")
        .createNotification(
            "Tool Error",
            message,
            NotificationType.ERROR
        )
        .notify(project)
}
```

## Best Practices

### External Tool Guidelines

1. **Validate configuration**: Check tool exists and is executable
2. **Handle timeouts**: Don't hang indefinitely
3. **Show progress**: Indicate when running
4. **Capture output**: Log for debugging
5. **Handle cancellation**: Allow user to stop

### Example: Complete Tool Integration

```kotlin
class MyCompleteTool(val project: Project) {
    private val settings = MyToolSettings.getInstance()

    fun runOnFile(file: VirtualFile): Result<String> {
        // Validate settings
        if (settings.executablePath.isBlank()) {
            return Result.failure(Exception("Tool executable not configured"))
        }

        // Build command
        val commandLine = GeneralCommandLine()
            .withExePath(settings.executablePath)
            .withParameters(settings.defaultArguments.split(" "))
            .withParameters(file.path)
            .withWorkDirectory(settings.workingDirectory.ifEmpty { project.basePath })

        // Run with progress
        return runWithProgress(commandLine)
    }

    private fun runWithProgress(commandLine: GeneralCommandLine): Result<String> {
        var result: Result<String> = Result.failure(Exception("Not completed"))

        ProgressManager.getInstance().run(object : Task.Backgroundable(project, "Running tool") {
            override fun run(indicator: ProgressIndicator) {
                indicator.text = "Executing..."

                val processHandler = OSProcessHandler(commandLine)
                val output = ProcessOutputCapture()
                output.capture(processHandler)

                processHandler.startNotify()

                while (!processHandler.isProcessTerminated) {
                    if (indicator.isCanceled) {
                        processHandler.destroyProcess()
                        result = Result.failure(Exception("Cancelled"))
                        return
                    }
                    Thread.sleep(100)
                }

                if (processHandler.exitCode == 0) {
                    result = Result.success(output.getStdout())
                } else {
                    result = Result.failure(Exception(output.getStderr()))
                }
            }
        })

        return result
    }
}
```

## Related Topics

- [Run Configurations](run-configurations.md) - Run configuration setup
- [Progress Indicators](../07-ui-components/progress-indicators.md) - Progress tracking
- [Tool Windows](../07-ui-components/tool-windows.md) - Tool window creation

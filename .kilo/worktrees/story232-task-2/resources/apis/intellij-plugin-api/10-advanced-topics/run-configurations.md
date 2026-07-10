# Run Configurations

Run configurations define how applications are executed and debugged in the IDE.

## Overview

### What Are Run Configurations?

Run configurations provide:

- Execution parameters
- Environment setup
- Debugging support
- Build dependencies

### Configuration Types

| Type        | Description          |
| ----------- | -------------------- |
| Application | Standard application |
| Tests       | Test execution       |
| Remote      | Remote debugging     |
| Custom      | Plugin-defined       |

## Creating a Run Configuration

### Configuration Type

```kotlin
class MyRunConfigurationType : ConfigurationType {
    override fun getDisplayName(): String = "My Configuration"

    override fun getConfigurationTypeDescription(): String =
        "Runs my custom application"

    override fun getIcon(): Icon = AllIcons.RunConfigurations.Application

    override fun getId(): String = "MyRunConfiguration"

    override fun getConfigurationFactories(): Array<ConfigurationFactory> {
        return arrayOf(MyConfigurationFactory(this))
    }
}
```

### Configuration Factory

```kotlin
class MyConfigurationFactory(type: ConfigurationType) : ConfigurationFactory(type) {
    override fun getId(): String = "MyFactory"

    override fun createTemplateConfiguration(project: Project): RunConfiguration {
        return MyRunConfiguration(project, this)
    }

    override fun createConfiguration(name: String, template: RunConfiguration): RunConfiguration {
        val configuration = super.createConfiguration(name, template) as MyRunConfiguration
        configuration.name = name
        return configuration
    }
}
```

### Run Configuration Implementation

```kotlin
class MyRunConfiguration(
    project: Project,
    factory: ConfigurationFactory
) : RunConfigurationBase<MyRunConfiguration>(project, factory, "My Configuration") {

    var mainClass: String = ""
    var arguments: String = ""
    var workingDirectory: String = ""
    var environmentVariables: Map<String, String> = emptyMap()

    override fun getConfigurationEditor(): SettingsEditor<out RunConfiguration> {
        return MySettingsEditor()
    }

    override fun getState(
        executor: Executor,
        environment: ExecutionEnvironment
    ): RunProfileState? {
        return MyCommandLineState(environment)
    }

    override fun checkConfiguration() {
        if (mainClass.isBlank()) {
            throw RuntimeConfigurationError("Main class not specified")
        }
    }
}
```

## Configuration Editor

### Settings Editor

```kotlin
class MySettingsEditor : SettingsEditor<MyRunConfiguration>() {
    private val mainClassField = JTextField()
    private val argumentsField = JTextField()
    private val workingDirectoryField = TextFieldWithBrowseButton()

    init {
        workingDirectoryField.addBrowseFolderListener(
            "Select Working Directory",
            null,
            null,
            FileChooserDescriptorFactory.createSingleFolderDescriptor()
        )
    }

    override fun createEditor(): JComponent {
        return panel {
            row("Main Class:") {
                mainClassField(growX)
            }
            row("Arguments:") {
                argumentsField(growX)
            }
            row("Working Directory:") {
                workingDirectoryField(growX)
            }
        }
    }

    override fun resetEditorFrom(configuration: MyRunConfiguration) {
        mainClassField.text = configuration.mainClass
        argumentsField.text = configuration.arguments
        workingDirectoryField.text = configuration.workingDirectory
    }

    override fun applyEditorTo(configuration: MyRunConfiguration) {
        configuration.mainClass = mainClassField.text
        configuration.arguments = argumentsField.text
        configuration.workingDirectory = workingDirectoryField.text
    }
}
```

## Execution

### Command Line State

```kotlin
class MyCommandLineState(
    private val environment: ExecutionEnvironment
) : CommandLineState(environment) {

    override fun startProcess(): ProcessHandler {
        val configuration = environment.runProfile as MyRunConfiguration

        val commandLine = GeneralCommandLine()
            .withExePath("java")
            .withParameters("-cp", getClasspath())
            .withParameters(configuration.mainClass)
            .withParameters(configuration.arguments.split(" "))
            .withWorkDirectory(configuration.workingDirectory)

        // Add environment variables
        val env = commandLine.environment
        env.putAll(configuration.environmentVariables)

        return OSProcessHandler(commandLine)
    }

    private fun getClasspath(): String {
        // Build classpath
        return "path/to/classes"
    }
}
```

### Java Command Line

```kotlin
class MyJavaCommandLineState(
    private val environment: ExecutionEnvironment
) : JavaCommandLineState(environment) {

    override fun createJavaParameters(): JavaParameters {
        val configuration = environment.runProfile as MyRunConfiguration
        val params = JavaParameters()

        // Set JDK
        params.jdk = getJdk()

        // Set classpath
        params.classPath.add(getClasspath())

        // Set main class
        params.mainClass = configuration.mainClass

        // Set arguments
        params.programParametersList.addParametersString(configuration.arguments)

        // Set working directory
        params.workingDirectory = configuration.workingDirectory

        // Set environment
        params.env.putAll(configuration.environmentVariables)

        return params
    }
}
```

## Run Configuration Producer

### Creating from Context

```kotlin
class MyRunConfigurationProducer : RunConfigurationProducer<MyRunConfiguration>(MyConfigurationType()) {

    override fun setupConfigurationFromContext(
        configuration: MyRunConfiguration,
        context: ConfigurationContext,
        sourceElement: Ref<PsiElement>
    ): Boolean {
        val element = context.psiLocation ?: return false

        // Check if element is a main method
        val mainMethod = findMainMethod(element) ?: return false

        sourceElement.set(mainMethod)
        configuration.mainClass = getQualifiedName(mainMethod)
        configuration.name = "Run ${configuration.mainClass}"

        return true
    }

    override fun isConfigurationFromContext(
        configuration: MyRunConfiguration,
        context: ConfigurationContext
    ): Boolean {
        val element = context.psiLocation ?: return false
        val mainMethod = findMainMethod(element) ?: return false

        return configuration.mainClass == getQualifiedName(mainMethod)
    }

    private fun findMainMethod(element: PsiElement): PsiMethod? {
        // Find main method
        return null
    }

    private fun getQualifiedName(method: PsiMethod): String {
        val containingClass = method.containingClass ?: return ""
        return "${containingClass.qualifiedName}"
    }
}
```

## Registering Run Configuration

### plugin.xml Registration

```xml
<extensions defaultExtensionNs="com.intellij">
    <configurationType
        implementation="com.example.MyRunConfigurationType"/>

    <runConfigurationProducer
        implementation="com.example.MyRunConfigurationProducer"/>
</extensions>
```

## Run Options

### Before Run Tasks

```kotlin
class MyBeforeRunTask : BeforeRunTask<MyBeforeRunTask>(MyBeforeRunTaskProvider.ID) {
    var enabled: Boolean = true
}

class MyBeforeRunTaskProvider : BeforeRunTaskProvider<MyBeforeRunTask>() {
    companion object {
        val ID = Key.create<MyBeforeRunTask>("MyBeforeRunTask")
    }

    override fun getId(): Key<MyBeforeRunTask> = ID

    override fun getName(): String = "My Task"

    override fun getDescription(task: MyBeforeRunTask): String =
        "Execute my task before run"

    override fun createTask(runConfiguration: RunConfiguration): MyBeforeRunTask? {
        return MyBeforeRunTask()
    }

    override fun executeTask(
        context: RunConfiguration,
        configuration: MyBeforeRunTask,
        env: ExecutionEnvironment
    ): Boolean {
        // Execute before run task
        return true
    }
}
```

### Registering Before Run Task

```xml
<extensions defaultExtensionNs="com.intellij">
    <beforeRunTask
        implementation="com.example.MyBeforeRunTaskProvider"/>
</extensions>
```

## Debug Support

### Enabling Debug

```kotlin
class MyCommandLineState(
    private val environment: ExecutionEnvironment
) : CommandLineState(environment) {

    override fun startProcess(): ProcessHandler {
        val configuration = environment.runProfile as MyRunConfiguration
        val commandLine = buildCommandLine(configuration)

        val handler = OSProcessHandler(commandLine)

        // Attach debugger if needed
        if (environment.executor.id == DefaultDebugExecutor.EXECUTOR_ID) {
            attachDebugger(handler)
        }

        return handler
    }

    private fun attachDebugger(handler: ProcessHandler) {
        // Attach debugger
    }
}
```

## Best Practices

### Run Configuration Guidelines

1. **Validate configuration**: Check all required fields
2. **Provide defaults**: Set sensible default values
3. **Support context creation**: Create from right-click
4. **Handle errors gracefully**: Show clear error messages
5. **Persist settings**: Save configuration state

### Example: Complete Run Configuration

```kotlin
// Configuration Type
class MyCompleteRunConfigurationType : ConfigurationType {
    override fun getDisplayName() = "My Application"
    override fun getConfigurationTypeDescription() = "Run my application"
    override fun getIcon() = AllIcons.RunConfigurations.Application
    override fun getId() = "MyCompleteRunConfiguration"

    override fun getConfigurationFactories() = arrayOf(
        MyCompleteConfigurationFactory(this)
    )
}

// Configuration
class MyCompleteRunConfiguration(
    project: Project,
    factory: ConfigurationFactory
) : RunConfigurationBase<MyCompleteRunConfiguration>(project, factory, "My App") {

    var mainClass: String = ""
    var arguments: String = ""
    var workingDirectory: String = project.basePath ?: ""
    var envVars: Map<String, String> = emptyMap()

    override fun getConfigurationEditor() = MyCompleteSettingsEditor()

    override fun getState(executor: Executor, env: ExecutionEnvironment) =
        MyCompleteCommandLineState(env)

    override fun checkConfiguration() {
        if (mainClass.isBlank()) {
            throw RuntimeConfigurationError("Main class required")
        }
    }

    override fun readExternal(element: Element) {
        super.readExternal(element)
        mainClass = element.getAttributeValue("mainClass") ?: ""
        arguments = element.getAttributeValue("arguments") ?: ""
        workingDirectory = element.getAttributeValue("workingDir") ?: ""
    }

    override fun writeExternal(element: Element) {
        super.writeExternal(element)
        element.setAttribute("mainClass", mainClass)
        element.setAttribute("arguments", arguments)
        element.setAttribute("workingDir", workingDirectory)
    }
}
```

## Related Topics

- [Debuggers](debuggers.md) - Debugger integration
- [External Tools](external-tools.md) - External tool integration
- [Actions System](../01-core-concepts/actions-system.md) - Action registration

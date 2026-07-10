# Extension Points

Extension points define places where plugins can contribute functionality to the IntelliJ Platform. They are the primary mechanism for extensibility in the platform.

## Understanding Extension Points

### What Is an Extension Point?

An extension point is a named location in the platform where plugins can register implementations. The platform defines hundreds of extension points, and plugins can define their own for other plugins to extend.

### Extension Point Types

Extension points accept different types of extensions:

| Type      | Description                                  |
| --------- | -------------------------------------------- |
| Interface | Extensions implement a Java/Kotlin interface |
| Class     | Extensions extend a base class               |
| Bean      | Extensions are defined as XML configuration  |

## Using Extension Points

### Registering Extensions

Extensions are registered in `plugin.xml`:

```xml
<extensions defaultExtensionNs="com.intellij">
    <toolWindow id="MyToolWindow"
                factoryClass="com.example.MyToolWindowFactory"
                anchor="right"/>
</extensions>
```

The `defaultExtensionNs` attribute specifies the namespace:

- `com.intellij` - Platform extension points
- Your plugin ID - Your plugin's extension points

### Extension Attributes

Each extension point defines its own attributes:

```xml
<extensions defaultExtensionNs="com.intellij">
    <fileType name="My File Type"
              implementationClass="com.example.MyFileType"
              extensions="myf;myft"
              fieldName="INSTANCE"/>
</extensions>
```

### Multiple Extensions

Register multiple extensions of the same type:

```xml
<extensions defaultExtensionNs="com.intellij">
    <inspectionToolProvider implementation="com.example.InspectionProvider1"/>
    <inspectionToolProvider implementation="com.example.InspectionProvider2"/>
</extensions>
```

## Common Platform Extension Points

### Actions

```xml
<extensions defaultExtensionNs="com.intellij">
    <actionPromoter implementation="com.example.MyActionPromoter"/>
</extensions>
```

### File Types

```xml
<extensions defaultExtensionNs="com.intellij">
    <fileType name="My File"
              implementationClass="com.example.MyFileType"
              extensions="myf"/>
</extensions>
```

### Services

```xml
<extensions defaultExtensionNs="com.intellij">
    <applicationService serviceImplementation="com.example.MyService"/>
    <projectService serviceImplementation="com.example.MyProjectService"/>
</extensions>
```

### Tool Windows

```xml
<extensions defaultExtensionNs="com.intellij">
    <toolWindow id="MyTool"
                factoryClass="com.example.MyToolWindowFactory"
                anchor="right"
                icon="/icons/toolWindow.svg"/>
</extensions>
```

### Inspections

```xml
<extensions defaultExtensionNs="com.intellij">
    <localInspection language="JAVA"
                     shortName="MyInspection"
                     displayName="My Inspection"
                     groupPath="Java"
                     groupName="Code style"
                     enabledByDefault="true"
                     implementationClass="com.example.MyInspection"/>
</extensions>
```

### Annotators

```xml
<extensions defaultExtensionNs="com.intellij">
    <annotator language="JAVA" implementationClass="com.example.MyAnnotator"/>
</extensions>
```

### Code Completion

```xml
<extensions defaultExtensionNs="com.intellij">
    <completion.contributor language="JAVA"
                            implementationClass="com.example.MyCompletionContributor"/>
</extensions>
```

## Creating Custom Extension Points

### Defining Extension Points

Define extension points in `plugin.xml`:

```xml
<extensionPoints>
    <extensionPoint name="myExtension"
                    interface="com.example.MyExtensionInterface"/>
</extensionPoints>
```

### Extension Point Attributes

| Attribute   | Description                                            |
| ----------- | ------------------------------------------------------ |
| `name`      | The extension point name                               |
| `interface` | Interface extensions must implement                    |
| `beanClass` | Bean class for configuration-based extensions          |
| `area`      | Scope: `IDEA_PROJECT`, `IDEA_MODULE`, or `APPLICATION` |

### Interface-Based Extension Points

Define an interface that extensions implement:

```kotlin
interface MyExtensionInterface {
    fun process(data: String): String
    val name: String
}
```

Extensions register implementations:

```xml
<extensions defaultExtensionNs="com.example.my-plugin">
    <myExtension implementation="com.otherplugin.MyExtensionImpl"/>
</extensions>
```

### Bean-Based Extension Points

For configuration without code, use bean classes:

```kotlin
class MyExtensionBean : AbstractExtensionPointBean() {
    @Attribute("name")
    var name: String? = null

    @Attribute("value")
    var value: String? = null
}
```

Define the extension point:

```xml
<extensionPoints>
    <extensionPoint name="myConfigExtension"
                    beanClass="com.example.MyExtensionBean"/>
</extensionPoints>
```

Extensions provide configuration:

```xml
<extensions defaultExtensionNs="com.example.my-plugin">
    <myConfigExtension name="option1" value="value1"/>
    <myConfigExtension name="option2" value="value2"/>
</extensions>
```

## Accessing Extensions

### Extension List

Get all registered extensions:

```kotlin
val extensions = MyExtensionInterface.EP_NAME.extensionList
extensions.forEach { extension ->
    println(extension.name)
}
```

### Extension Point Name

Define the extension point name in your interface or companion object:

```kotlin
interface MyExtensionInterface {
    companion object {
        val EP_NAME = ExtensionPointName<MyExtensionInterface>(
            "com.example.my-plugin.myExtension"
        )
    }

    fun process(data: String): String
    val name: String
}
```

### Extension by Name

Find a specific extension:

```kotlin
val extension = MyExtensionInterface.EP_NAME.extensionList
    .find { it.name == "specificName" }
```

### Lazy Extension Access

For better performance, use lazy access:

```kotlin
private val extensions by lazy {
    MyExtensionInterface.EP_NAME.extensionList
}
```

## Extension Point Best Practices

### Naming Conventions

Use consistent naming for extension points:

- Use camelCase for the extension point name
- Include the feature area in the name
- Make the purpose clear

```xml
<!-- Good -->
<extensionPoint name="codeGenerator" interface="..."/>
<extensionPoint name="fileProcessor" interface="..."/>

<!-- Avoid -->
<extensionPoint name="ext1" interface="..."/>
<extensionPoint name="processor" interface="..."/>
```

### Documentation

Document extension points with the `description` attribute:

```xml
<extensionPoints>
    <extensionPoint name="myExtension"
                    interface="com.example.MyExtensionInterface">
        <with attribute="implementationClass"
              implements="com.example.MyExtensionInterface"/>
    </extensionPoint>
</extensionPoints>
```

### Versioning

Mark extension points with version information:

```xml
<extensionPoints>
    <extensionPoint name="myExtension"
                    interface="com.example.MyExtensionInterface"
                    deprecated="true"
                    deprecationMessage="Use newExtension instead"/>
</extensionPoints>
```

## Dynamic Extensions

### Handling Extension Changes

Listen for extension changes:

```kotlin
class MyExtensionManager : Disposable {
    private val connection = ApplicationManager.getApplication().messageBus.connect()

    init {
        MyExtensionInterface.EP_NAME.addChangeListener(this) {
            // Extensions changed, reload
            reloadExtensions()
        }
    }

    override fun dispose() {
        connection.dispose()
    }
}
```

### Dynamic Plugin Support

For dynamic plugin support, properly handle extension registration and unregistration:

```kotlin
class MyExtensionManager {
    private val activeExtensions = mutableMapOf<String, MyExtensionInterface>()

    fun registerExtension(extension: MyExtensionInterface) {
        activeExtensions[extension.name] = extension
    }

    fun unregisterExtension(extension: MyExtensionInterface) {
        activeExtensions.remove(extension.name)
    }
}
```

## Extension Point Dependencies

Extensions can depend on other plugins being available:

```xml
<extensions defaultExtensionNs="com.intellij">
    <fileType name="My File"
              implementationClass="com.example.MyFileType"
              extensions="myf"
              dependencies="com.intellij.modules.java"/>
</extensions>
```

## Advanced Patterns

### Extension with Attributes

```xml
<extensionPoints>
    <extensionPoint name="processor"
                    beanClass="com.example.ProcessorBean">
        <with attribute="implementationClass"
              implements="com.example.Processor"/>
    </extensionPoint>
</extensionPoints>
```

```kotlin
class ProcessorBean : AbstractExtensionPointBean() {
    @Attribute("implementationClass")
    var implementationClass: String? = null

    @Attribute("priority")
    var priority: Int = 0

    fun getInstance(): Processor {
        return instantiateClass(implementationClass!!, plugin.pluginClassLoader)
    }
}
```

### Grouped Extensions

```xml
<extensionPoints>
    <extensionPoint name="processorGroup">
        <extensionPoint name="preProcessor" interface="com.example.Processor"/>
        <extensionPoint name="postProcessor" interface="com.example.Processor"/>
    </extensionPoint>
</extensionPoints>
```

## Related Topics

- [Services](services.md) - Services often work with extension points
- [Plugin Configuration](../09-deployment/plugin-configuration.md) - Plugin XML structure
- [Dynamic Plugins](../09-deployment/dynamic-plugins.md) - Dynamic plugin loading

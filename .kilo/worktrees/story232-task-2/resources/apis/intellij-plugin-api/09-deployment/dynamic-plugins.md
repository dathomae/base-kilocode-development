# Dynamic Plugins

Dynamic plugins can be loaded, unloaded, and updated without restarting the IDE.

## Overview

### What Are Dynamic Plugins?

Dynamic plugins support:

- Hot reloading during development
- Installation without restart
- Update without restart
- Uninstallation without restart

### Requirements

For dynamic loading, plugins must:

- Declare `dynamic="true"` in plugin.xml
- Handle load/unload events
- Clean up resources properly
- Avoid static state

## Enabling Dynamic Loading

### Plugin Declaration

```xml
<idea-plugin>
    <!-- Enable dynamic loading -->
    <id>com.example.myplugin</id>
    <name>My Plugin</name>
    <version>1.0.0</version>

    <!-- Mark as dynamic -->
    <extensions defaultExtensionNs="com.intellij">
        <!-- Extensions must support dynamic loading -->
    </extensions>
</idea-plugin>
```

### Extension Point Configuration

```xml
<extensionPoints>
    <extensionPoint
        name="myExtension"
        interface="com.example.MyExtension"
        dynamic="true"/>
</extensionPoints>
```

## Plugin Lifecycle

### Load Events

```kotlin
class MyPluginInitializer : PluginInitializer {
    override fun initialize(pluginId: String) {
        // Called when plugin is loaded
        println("Plugin $pluginId loaded")
    }
}
```

### Unload Events

```kotlin
class MyPluginDisposable : Disposable {
    override fun dispose() {
        // Called when plugin is unloaded
        cleanup()
    }
}
```

## Handling Dynamic Changes

### Service Cleanup

```kotlin
class MyService : Disposable {
    private val listeners = mutableListOf<Listener>()

    fun addListener(listener: Listener) {
        listeners.add(listener)
    }

    override fun dispose() {
        // Clean up listeners
        listeners.clear()
    }
}
```

### Extension Point Changes

```kotlin
class MyExtensionPointListener : ExtensionPointListener<MyExtension> {
    override fun extensionAdded(extension: MyExtension, pluginDescriptor: PluginDescriptor) {
        // Handle extension added
    }

    override fun extensionRemoved(extension: MyExtension, pluginDescriptor: PluginDescriptor) {
        // Handle extension removed
    }
}

// Register listener
ExtensionPointName.create("com.example.myExtension")
    .addExtensionPointListener(MyExtensionPointListener(), /* disposable */ parentDisposable)
```

## Avoiding Static State

### Problem: Static State

```kotlin
// BAD: Static state prevents unloading
object MyGlobalState {
    var data: String = ""
}

class MyService {
    fun process() {
        MyGlobalState.data = "processed"
    }
}
```

### Solution: Instance State

```kotlin
// GOOD: Instance state in service
class MyService : Disposable {
    private var data: String = ""

    fun process() {
        data = "processed"
    }

    override fun dispose() {
        data = ""
    }
}
```

## Disposables

### Using Disposables

```kotlin
class MyComponent {
    private val disposable = Disposer.newDisposable()

    init {
        // Register child disposable
        val child = Disposer.newDisposable(disposable)

        // Will be disposed with parent
        Disposer.register(disposable, child)
    }

    fun dispose() {
        Disposer.dispose(disposable)
    }
}
```

### Project-Level Disposables

```kotlin
class MyProjectService(val project: Project) {
    private val disposable = Disposer.newDisposable()

    init {
        // Register with project
        Disposer.register(project, disposable)

        // Subscribe to events
        project.messageBus.connect(disposable)
            .subscribe(MyTopic.CHANGES, MyChangeListener())
    }
}
```

## Message Bus

### Dynamic Subscriptions

```kotlin
class MyDynamicSubscriber(val project: Project) {
    private var connection: MessageBusConnection? = null

    fun subscribe() {
        connection = project.messageBus.connect()
        connection?.subscribe(MyTopic.CHANGES, object : MyChangeListener {
            override fun onChanged() {
                // Handle change
            }
        })
    }

    fun unsubscribe() {
        connection?.disconnect()
        connection = null
    }
}
```

### Using Disposables with Message Bus

```kotlin
class MyService(val project: Project) : Disposable {
    init {
        // Connection disposed automatically
        project.messageBus.connect(this)
            .subscribe(MyTopic.CHANGES, MyChangeListener())
    }

    override fun dispose() {
        // Connection already disposed
    }
}
```

## Actions and Dynamic Loading

### Dynamic Action Registration

```kotlin
class MyDynamicAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        // Action logic
    }
}

// Register dynamically
val actionManager = ActionManager.getInstance()
val action = MyDynamicAction()
actionManager.registerAction("MyDynamicAction", action, pluginDisposable)

// Unregister automatically when disposed
```

### Action Group Updates

```kotlin
class MyDynamicActionGroup : DefaultActionGroup() {
    override fun getChildren(e: AnActionEvent?): Array<AnAction> {
        // Dynamically compute children
        return computeActions()
    }
}
```

## Testing Dynamic Loading

### Manual Testing

1. Build plugin
2. Install in IDE
3. Modify plugin code
4. Build again
5. Update plugin without restart
6. Verify changes work

### Automated Testing

```kotlin
class DynamicLoadingTest : BasePlatformTestCase() {
    fun testDynamicLoad() {
        val pluginId = PluginId.getId("com.example.myplugin")

        // Load plugin
        val plugin = PluginManagerCore.getPlugin(pluginId)
        assertNotNull(plugin)

        // Verify services
        val service = ApplicationManager.getApplication()
            .getService(MyService::class.java)
        assertNotNull(service)
    }
}
```

## Common Issues

### Memory Leaks

```kotlin
// BAD: Listener not removed
class MyComponent {
    init {
        ApplicationManager.getApplication()
            .messageBus
            .connect()
            .subscribe(AppTopics.FILE_DOCUMENT_SYNC, MyListener())
    }
}

// GOOD: Listener removed on dispose
class MyComponent : Disposable {
    private val disposable = Disposer.newDisposable()

    init {
        ApplicationManager.getApplication()
            .messageBus
            .connect(disposable)
            .subscribe(AppTopics.FILE_DOCUMENT_SYNC, MyListener())
    }

    override fun dispose() {
        Disposer.dispose(disposable)
    }
}
```

### Thread Safety

```kotlin
// BAD: Not thread-safe
class MyService {
    var state: String = ""
}

// GOOD: Thread-safe with volatile
class MyService {
    @Volatile
    var state: String = ""
}
```

## Best Practices

### Dynamic Plugin Guidelines

1. **Mark as dynamic**: Set `dynamic="true"` in plugin.xml
2. **Use disposables**: Clean up resources properly
3. **Avoid static state**: Use services instead
4. **Handle events**: Respond to load/unload
5. **Test thoroughly**: Verify hot reload works

### Example: Dynamic Plugin

```kotlin
// Service with proper cleanup
class MyDynamicService : Disposable {
    private val data = mutableMapOf<String, Any>()
    private val listeners = mutableListOf<ChangeListener>()

    fun getData(key: String): Any? = data[key]

    fun setData(key: String, value: Any) {
        data[key] = value
        notifyListeners(key, value)
    }

    fun addListener(listener: ChangeListener) {
        listeners.add(listener)
    }

    private fun notifyListeners(key: String, value: Any) {
        listeners.forEach { it.onChanged(key, value) }
    }

    override fun dispose() {
        data.clear()
        listeners.clear()
    }
}

// Plugin configuration
<idea-plugin>
    <id>com.example.myplugin</id>
    <name>My Dynamic Plugin</name>

    <extensions defaultExtensionNs="com.intellij">
        <applicationService
            serviceImplementation="com.example.MyDynamicService"/>
    </extensions>
</idea-plugin>
```

## Related Topics

- [Plugin Configuration](plugin-configuration.md) - Plugin manifest
- [Plugin Lifecycle](../01-core-concepts/plugin-lifecycle.md) - Lifecycle events
- [Services](../01-core-concepts/services.md) - Service implementation

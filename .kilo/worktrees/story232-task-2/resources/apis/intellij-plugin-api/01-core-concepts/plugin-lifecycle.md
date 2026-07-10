# Plugin Lifecycle

Understanding the plugin lifecycle is essential for proper resource management and initialization. Plugins go through distinct phases from loading to unloading.

## Lifecycle Phases

### Loading Phase

The platform loads plugins during IDE startup in this order:

1. **Discovery**: The platform scans the plugin directory for `plugin.xml` files
2. **Dependency Resolution**: Dependencies between plugins are resolved
3. **Class Loading**: Plugin classes are loaded into isolated classloaders
4. **Component Creation**: Components and services are instantiated

### Initialization Phase

After loading, the platform initializes plugins:

1. **Service Instantiation**: Application-level services are created
2. **Extension Registration**: Extensions are registered with the platform
3. **Action Registration**: Actions are added to menus and toolbars

The initialization happens before any project is opened. Plugins should not assume a project exists during this phase.

### Runtime Phase

During normal operation:

1. **Project Opening**: Project-level services are created when a project opens
2. **User Interaction**: Actions are executed, services are called
3. **Event Processing**: Listeners receive and process events

### Shutdown Phase

When the IDE closes or a plugin is unloaded:

1. **Project Closing**: Project-level services are disposed
2. **Application Shutdown**: Application-level services are disposed
3. **Plugin Unloading**: Plugin resources are released

## Component Initialization

### Application-Level Components

Application-level components are created once when the IDE starts. They remain active until the IDE closes.

```xml
<application-service serviceImplementation="com.example.MyApplicationService"/>
```

```kotlin
class MyApplicationService {
    init {
        // Initialization code runs once at startup
    }

    fun dispose() {
        // Cleanup when IDE closes
    }
}
```

### Project-Level Components

Project-level components are created when a project opens and disposed when it closes.

```xml
<project-service serviceImplementation="com.example.MyProjectService"/>
```

```kotlin
class MyProjectService(val project: Project) : Disposable {
    init {
        // Initialization code runs for each project
    }

    override fun dispose() {
        // Cleanup when project closes
    }
}
```

## Plugin State

Plugins can persist state across IDE sessions using the state storage mechanism.

### Persistent State Component

Implement `PersistentStateComponent` to save and restore state:

```kotlin
class MySettings : PersistentStateComponent<MySettings.State> {
    private var state = State()

    data class State(
        var enabled: Boolean = true,
        var lastUsedPath: String? = null
    )

    override fun getState(): State = state

    override fun loadState(state: State) {
        this.state = state
    }

    companion object {
        fun getInstance(): MySettings =
            service<MySettings>()
    }
}
```

### Storage Location

Specify where state is stored in `plugin.xml`:

```xml
<application-service
    serviceImplementation="com.example.MySettings"
    storage="com.example.MySettings">
    <option name="storage" value="file:$APP_CONFIG$/my_settings.xml"/>
</application-service>
```

Common storage locations:

- `$APP_CONFIG$`: IDE configuration directory
- `$PROJECT_CONFIG$`: Project-specific `.idea` directory

## Initialization Order

### Plugin Dependencies

When plugins depend on each other, initialization order is determined by dependencies:

```xml
<depends>com.example.other-plugin</depends>
```

The dependent plugin loads after its dependencies. Services from dependencies are available during initialization.

### Service Dependencies

Services can depend on other services. Use constructor injection or service retrieval:

```kotlin
class MyService {
    private val otherService = service<OtherService>()

    // Or use constructor injection
    constructor() {
        val anotherService = service<AnotherService>()
    }
}
```

## Dynamic Plugin Loading

Modern versions of the IntelliJ Platform support dynamic plugin loading. This allows plugins to be loaded and unloaded without restarting the IDE.

### Requirements for Dynamic Loading

To support dynamic loading, plugins must:

1. Properly implement `Disposable` for cleanup
2. Release all resources in `dispose()` methods
3. Unregister listeners and subscriptions
4. Not hold references to disposed components

### Dynamic Unloading

When a plugin is unloaded:

1. All project-level services are disposed
2. All application-level services are disposed
3. Extension points are unregistered
4. Actions are removed from UI

```kotlin
class MyDynamicService : Disposable {
    private val connection = ApplicationManager.getApplication().messageBus.connect()

    init {
        connection.subscribe(MyTopic, myListener)
    }

    override fun dispose() {
        // Connection is automatically disposed
        // Clean up any other resources here
    }
}
```

## Startup Activities

Plugins can run code during IDE startup using startup activities.

### Registering a Startup Activity

```xml
<extensions defaultExtensionNs="com.intellij">
    <postStartupActivity implementation="com.example.MyStartupActivity"/>
</extensions>
```

```kotlin
class MyStartupActivity : ProjectActivity {
    override suspend fun execute(project: Project) {
        // Runs after project is opened
        // Runs in background thread
    }
}
```

### Background vs. UI Thread

Startup activities run on a background thread. To perform UI operations:

```kotlin
class MyStartupActivity : ProjectActivity {
    override suspend fun execute(project: Project) {
        withContext(Dispatchers.EDT) {
            // UI operations here
        }
    }
}
```

## Common Pitfalls

### Memory Leaks

Memory leaks occur when plugins hold references to disposed components:

```kotlin
// WRONG: Static reference to project
object MyHelper {
    var project: Project? = null  // Leaks when project closes
}

// CORRECT: Use project-level service
class MyHelper(val project: Project) {
    // Automatically disposed with project
}
```

### Blocking Initialization

Long-running operations during initialization slow IDE startup:

```kotlin
// WRONG: Blocking initialization
class MyService {
    init {
        Thread.sleep(5000)  // Blocks startup
    }
}

// CORRECT: Lazy initialization
class MyService {
    private val data by lazy {
        // Computed when first accessed
        loadData()
    }
}
```

### Missing Disposal

Resources must be properly disposed:

```kotlin
// WRONG: Not disposing resources
class MyService {
    val timer = Timer()
    init {
        timer.schedule(task, 1000, 1000)
    }
}

// CORRECT: Dispose resources
class MyService : Disposable {
    val timer = Timer()

    override fun dispose() {
        timer.cancel()
    }
}
```

## Related Topics

- [IntelliJ Platform](intellij-platform.md) - Platform architecture overview
- [Services](services.md) - Creating and managing services
- [Plugin Configuration](../09-deployment/plugin-configuration.md) - Configuring plugin metadata

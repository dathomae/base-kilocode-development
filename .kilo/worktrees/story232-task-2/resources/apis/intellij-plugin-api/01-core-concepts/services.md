# Services

Services are the primary way to organize business logic in IntelliJ Platform plugins. They provide a clean architecture for encapsulating functionality and managing state.

## Service Overview

A service is a singleton or per-project class that provides specific functionality. Services offer several benefits:

- **Managed Lifecycle**: The platform creates and disposes services automatically
- **Dependency Injection**: Services can depend on other services
- **Testability**: Services can be easily mocked in tests
- **State Management**: Services can persist state across sessions

## Service Scopes

### Application-Level Services

Application-level services exist once for the entire IDE instance. They are created on first access and disposed when the IDE closes.

```xml
<application-service serviceImplementation="com.example.MyApplicationService"/>
```

```kotlin
class MyApplicationService {
    fun doSomething() {
        // Application-wide functionality
    }
}
```

Access application services:

```kotlin
val service = service<MyApplicationService>()
// or
val service = ApplicationManager.getApplication().getService(MyApplicationService::class.java)
```

### Project-Level Services

Project-level services exist once per open project. They are created when the project opens and disposed when it closes.

```xml
<project-service serviceImplementation="com.example.MyProjectService"/>
```

```kotlin
class MyProjectService(val project: Project) {
    fun doSomething() {
        // Project-specific functionality
    }
}
```

Access project services:

```kotlin
val service = project.service<MyProjectService>()
// or
val service = project.getService(MyProjectService::class.java)
```

### Module-Level Services

Module-level services exist once per module within a project. They are less commonly used but available for module-specific functionality.

```xml
<module-service serviceImplementation="com.example.MyModuleService"/>
```

```kotlin
class MyModuleService(val module: Module) {
    fun doSomething() {
        // Module-specific functionality
    }
}
```

Access module services:

```kotlin
val service = module.getService(MyModuleService::class.java)
```

## Service Registration

### Basic Registration

The simplest registration specifies only the implementation class:

```xml
<application-service serviceImplementation="com.example.MyService"/>
```

### Interface and Implementation

Register with a separate interface and implementation:

```xml
<application-service
    serviceInterface="com.example.MyService"
    serviceImplementation="com.example.MyServiceImpl"/>
```

This pattern allows:

- Multiple implementations for different scenarios
- Easy mocking in tests
- Clean API separation

### Overriding Services

Plugins can override services from other plugins:

```xml
<application-service
    serviceInterface="com.example.MyService"
    serviceImplementation="com.otherplugin.MyCustomServiceImpl"
    overrides="true"/>
```

## Service Lifecycle

### Initialization

Services are created lazily on first access. To initialize eagerly:

```xml
<application-service
    serviceImplementation="com.example.MyService"
    preload="await"/>
```

Preload options:

- `await` - Load during startup, wait for completion
- `true` - Load during startup, don't wait

### Disposal

Services that need cleanup should implement `Disposable`:

```kotlin
class MyService : Disposable {
    private val resources = mutableListOf<Resource>()

    fun addResource(resource: Resource) {
        resources.add(resource)
    }

    override fun dispose() {
        resources.forEach { it.close() }
        resources.clear()
    }
}
```

The platform calls `dispose()` when the service scope ends.

## Persistent State

Services can persist state across IDE sessions by implementing `PersistentStateComponent`.

### State Class

Define a state class to hold persisted data:

```kotlin
data class MyState(
    var enabled: Boolean = true,
    var items: List<String> = emptyList(),
    var lastAccessed: Long = 0
)
```

### Persistent Service Implementation

```kotlin
class MySettings : PersistentStateComponent<MyState> {
    private var state = MyState()

    override fun getState(): MyState = state

    override fun loadState(state: MyState) {
        this.state = state
    }

    // Convenience methods
    var enabled: Boolean
        get() = state.enabled
        set(value) { state.enabled = value }
}
```

### Storage Configuration

Configure storage location in `plugin.xml`:

```xml
<application-service
    serviceImplementation="com.example.MySettings">
    <option name="storage" value="file:$APP_CONFIG$/my_settings.xml"/>
</application-service>
```

Storage paths:

- `$APP_CONFIG$` - IDE configuration directory
- `$PROJECT_CONFIG$` - Project `.idea` directory
- `$PROJECT_FILE$` - Project file (for directory-based projects)

### Roaming Settings

Settings can roam across devices:

```xml
<application-service
    serviceImplementation="com.example.MySettings">
    <option name="storage" value="roamingType:DEFAULT"/>
</application-service>
```

Roaming types:

- `DEFAULT` - Standard roaming
- `PER_OS` - Separate per operating system
- `DISABLED` - No roaming

## Service Dependencies

### Constructor Injection

Services can depend on other services through constructor injection:

```kotlin
class MyService(
    private val settings: MySettings,
    private val helper: HelperService
) {
    // Use injected services
}
```

The platform automatically resolves dependencies. Register normally:

```xml
<application-service serviceImplementation="com.example.MyService"/>
```

### Lazy Dependencies

For optional or circular dependencies, use lazy access:

```kotlin
class MyService {
    private val optionalService: OptionalService? by lazy {
        serviceIfCreated<OptionalService>()
    }
}
```

## Lightweight Services

For simple services without complex lifecycle needs, use lightweight services:

```kotlin
@Service(Service.Level.APP)
class MyLightweightService {
    fun doSomething() {
        // Simple functionality
    }
}
```

Lightweight services:

- Are registered automatically via annotation
- Do not require `plugin.xml` registration
- Support constructor injection

## Thread Safety

Services must be thread-safe. The platform may access services from multiple threads.

### Synchronized Access

```kotlin
class MyService {
    private val lock = ReentrantLock()
    private val data = mutableListOf<String>()

    fun addItem(item: String) {
        lock.withLock {
            data.add(item)
        }
    }

    fun getItems(): List<String> {
        lock.withLock {
            return data.toList()
        }
    }
}
```

### Concurrent Collections

```kotlin
class MyService {
    private val data = ConcurrentHashMap<String, String>()

    fun put(key: String, value: String) {
        data[key] = value
    }

    fun get(key: String): String? = data[key]
}
```

### Read/Write Actions

When accessing platform data, use read/write actions:

```kotlin
class MyService {
    fun processFile(file: VirtualFile): String? {
        return ReadAction.compute<String, Exception> {
            // Read PSI or other platform data
            val psiFile = PsiManager.getInstance(file.fileSystem.project)
                .findFile(file)
            psiFile?.text
        }
    }
}
```

## Testing Services

### Mocking Services

In tests, replace services with mocks:

```kotlin
class MyServiceTest {
    @BeforeEach
    fun setup() {
        // Register mock service
        val mockService = mock<MyService>()
        ApplicationManager.getApplication()
            .registerServiceInstance(MyService::class.java, mockService)
    }
}
```

### Test-Friendly Design

Design services for testability:

```kotlin
class MyService(
    private val dependency: DependencyService = service()  // Default, overridable
) {
    fun process(input: String): String {
        return dependency.transform(input)
    }
}
```

## Common Patterns

### Settings Service

```kotlin
class AppSettings : PersistentStateComponent<AppSettings.State> {
    private var state = State()

    data class State(
        var theme: String = "default",
        var autoSave: Boolean = true,
        var recentFiles: List<String> = emptyList()
    )

    override fun getState() = state
    override fun loadState(state: State) {
        this.state = state
    }

    companion object {
        fun getInstance() = service<AppSettings>()
    }
}
```

### Cache Service

```kotlin
class CacheService {
    private val cache = ConcurrentHashMap<String, CachedItem>()

    fun get(key: String): CachedItem? = cache[key]

    fun put(key: String, value: CachedItem) {
        cache[key] = value
    }

    fun invalidate(key: String) {
        cache.remove(key)
    }

    fun clear() {
        cache.clear()
    }
}
```

### Event Publisher Service

```kotlin
class EventService(val project: Project) {
    private val bus = project.messageBus

    fun publish(event: MyEvent) {
        bus.syncPublisher(MyTopic).onEvent(event)
    }

    fun subscribe(listener: MyListener): Disposable {
        return bus.connect().subscribe(MyTopic, listener)
    }
}
```

## Related Topics

- [Plugin Lifecycle](plugin-lifecycle.md) - Service lifecycle details
- [Extension Points](extension-points.md) - Extending services
- [Testing](../08-testing/) - Testing services

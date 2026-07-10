# Troubleshooting

This document covers common issues and their solutions in IntelliJ plugin development.

## Common Issues

### Plugin Not Loading

**Symptoms:**

- Plugin doesn't appear in IDE
- No error messages

**Solutions:**

1. Check plugin.xml location:

```
src/main/resources/META-INF/plugin.xml
```

2. Verify plugin ID:

```xml
<idea-plugin>
    <id>com.example.myplugin</id>
</idea-plugin>
```

3. Check dependencies:

```xml
<depends>com.intellij.modules.platform</depends>
```

### Class Not Found

**Symptoms:**

- `ClassNotFoundException`
- `NoClassDefFoundError`

**Solutions:**

1. Check class is in correct package
2. Verify dependency declarations
3. Check build configuration:

```kotlin
// build.gradle.kts
intellij {
    type.set("IC")
    version.set("2023.1")
}
```

### Action Not Appearing

**Symptoms:**

- Action doesn't show in menu
- Keyboard shortcut doesn't work

**Solutions:**

1. Check action registration:

```xml
<actions>
    <action id="MyAction" class="com.example.MyAction">
        <add-to-group group-id="ToolsMenu" anchor="last"/>
    </action>
</actions>
```

2. Verify action update:

```kotlin
override fun update(e: AnActionEvent) {
    e.presentation.isVisible = true
    e.presentation.isEnabled = true
}
```

### PSI Access Issues

**Symptoms:**

- `AssertionError: Write access is required`
- `Read access is required`

**Solutions:**

1. Use proper access actions:

```kotlin
// Read access
ReadAction.run {
    // Read PSI
}

// Write access
WriteAction.run<Exception> {
    // Modify PSI
}
```

2. Check thread:

```kotlin
if (ApplicationManager.getApplication().isReadAccessAllowed) {
    // Read access available
}

if (ApplicationManager.getApplication().isWriteAccessAllowed) {
    // Write access available
}
```

### Index Not Ready

**Symptoms:**

- `IndexNotReadyException`
- Empty search results

**Solutions:**

1. Check dumb mode:

```kotlin
if (DumbService.getInstance(project).isDumb) {
    DumbService.getInstance(project).runWhenSmart {
        // Run when indexing complete
    }
} else {
    // Run now
}
```

2. Use dumb-aware:

```kotlin
class MyService : DumbAware {
    // Service works during indexing
}
```

### Memory Issues

**Symptoms:**

- OutOfMemoryError
- Slow performance

**Solutions:**

1. Increase heap size:

```bash
# idea.vmoptions
-Xmx2g
```

2. Check for memory leaks:

```kotlin
// Use disposables
class MyComponent : Disposable {
    override fun dispose() {
        // Clean up
    }
}
```

### UI Thread Issues

**Symptoms:**

- `AssertionError: EventQueue.isDispatchThread()`
- UI freezes

**Solutions:**

1. Run UI on EDT:

```kotlin
ApplicationManager.getApplication().invokeLater {
    // UI updates
}
```

2. Run background tasks:

```kotlin
ProgressManager.getInstance().run(object : Task.Backgroundable(project, "Task") {
    override fun run(indicator: ProgressIndicator) {
        // Background work
    }
})
```

## Debugging Tips

### Enable Debug Logging

```kotlin
// Enable debug logging
Logger.getInstance(MyClass::class.java).setLevel(LogLevel.DEBUG)
```

### Check Plugin State

```kotlin
// Check if plugin is loaded
val plugin = PluginManagerCore.getPlugin(PluginId.getId("com.example.myplugin"))
println("Plugin loaded: ${plugin?.isEnabled}")
```

### Inspect PSI

```kotlin
// Print PSI tree
fun printPsiTree(element: PsiElement, indent: Int = 0) {
    println("  ".repeat(indent) + element.javaClass.simpleName + ": " + element.text)
    for (child in element.children) {
        printPsiTree(child, indent + 1)
    }
}
```

### Debug Actions

```kotlin
override fun update(e: AnActionEvent) {
    println("Action update called")
    println("Project: ${e.project}")
    println("File: ${e.getData(CommonDataKeys.PSI_FILE)}")
}
```

## Error Messages

### "Plugin is incompatible"

**Solution:**
Update plugin.xml:

```xml
<idea-version since-build="231.0" until-build="233.*"/>
```

### "Extension point not found"

**Solution:**
Check extension point name:

```xml
<extensions defaultExtensionNs="com.intellij">
    <!-- Correct namespace -->
</extensions>
```

### "Service not registered"

**Solution:**
Register service:

```xml
<extensions defaultExtensionNs="com.intellij">
    <applicationService
        serviceImplementation="com.example.MyService"/>
</extensions>
```

## Performance Issues

### Slow Startup

**Solutions:**

1. Defer initialization:

```kotlin
class MyService {
    private val data: Map<String, Any> by lazy {
        loadData()
    }
}
```

2. Use background loading:

```kotlin
class MyInitializer : StartupActivity {
    override fun runActivity(project: Project) {
        // Run in background
    }
}
```

### Slow Completion

**Solutions:**

1. Limit results:

```kotlin
override fun addCompletions(
    parameters: CompletionParameters,
    context: ProcessingContext,
    result: CompletionResultSet
) {
    // Limit to 100 results
    val items = computeItems().take(100)
    result.addAllElements(items)
}
```

2. Use indexes:

```kotlin
// Use file-based index
FileBasedIndex.getInstance().processFilesContainingAllKeys(
    MyIndex.ID,
    keys,
    scope,
    null,
    { file -> true }
)
```

## Testing Issues

### Tests Failing

**Solutions:**

1. Check test data path:

```kotlin
override fun getTestDataPath(): String = "src/test/testData"
```

2. Verify fixture setup:

```kotlin
override fun setUp() {
    super.setUp()
    // Setup code
}
```

### Fixture Errors

**Solutions:**

1. Check file creation:

```kotlin
myFixture.configureByText("test.my", "content")
```

2. Verify PSI access:

```kotlin
val file = myFixture.file
assertNotNull(file)
```

## Best Practices

### Troubleshooting Guidelines

1. **Check logs**: Look in IDE log file
2. **Use debugger**: Set breakpoints
3. **Isolate issue**: Create minimal reproduction
4. **Check documentation**: Review API docs
5. **Search issues**: Check JetBrains YouTrack

### Getting Help

1. JetBrains Platform Slack
2. IntelliJ SDK Documentation
3. JetBrains YouTrack Issues
4. Stack Overflow `[intellij-plugin]` tag

## Related Topics

- [Performance Tips](performance-tips.md) - Performance optimization
- [Common Patterns](common-patterns.md) - Code patterns
- [External Resources](external-resources.md) - External documentation

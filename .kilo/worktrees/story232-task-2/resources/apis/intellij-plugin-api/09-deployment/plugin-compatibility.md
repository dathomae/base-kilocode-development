# Plugin Compatibility

Plugin compatibility ensures your plugin works across different IDE versions and products.

## Overview

### Why Compatibility Matters

Compatibility considerations:

- IDE version updates
- Different JetBrains products
- Platform API changes
- Dependency management

### Compatibility Types

| Type          | Description               |
| ------------- | ------------------------- |
| Backward      | Works with older versions |
| Forward       | Works with newer versions |
| Cross-product | Works across IDEs         |

## Version Targeting

### Specifying Versions

```xml
<idea-plugin>
    <idea-version
        since-build="231.0"    <!-- Minimum version -->
        until-build="233.*"/>  <!-- Maximum version -->
</idea-plugin>
```

### Build Numbers

| IDE Version | Build Number |
| ----------- | ------------ |
| 2022.2      | 222.\*       |
| 2022.3      | 223.\*       |
| 2023.1      | 231.\*       |
| 2023.2      | 232.\*       |
| 2023.3      | 233.\*       |

### Version Ranges

```xml
<!-- Single major version -->
<idea-version since-build="231.0" until-build="231.*"/>

<!-- Multiple major versions -->
<idea-version since-build="223.0" until-build="233.*"/>

<!-- Open-ended (not recommended) -->
<idea-version since-build="231.0"/>
```

## API Compatibility

### Handling API Changes

```kotlin
// Check API availability
if (ApplicationInfo.getInstance().build >= BuildNumber.fromString("231.0")) {
    // Use new API
    NewApi.process()
} else {
    // Use old API
    OldApi.process()
}
```

### Using Utility Methods

```kotlin
object ApiCompatibility {
    fun isAtLeast(major: Int, minor: Int = 0): Boolean {
        val build = ApplicationInfo.getInstance().build
        return build.baselineVersion >= major && build.buildNumber >= minor
    }
}

// Usage
if (ApiCompatibility.isAtLeast(231)) {
    // Use 2023.1+ API
}
```

### Deprecated API Handling

```kotlin
@Suppress("DEPRECATION")
fun useDeprecatedApi() {
    // Use deprecated API with suppression
    OldApi.process()
}

@Deprecated("Use newMethod instead", ReplaceWith("newMethod()"))
fun oldMethod() {
    // Old implementation
}

fun newMethod() {
    // New implementation
}
```

## Cross-Product Compatibility

### Product-Specific Features

```xml
<!-- Optional dependency on Java module -->
<depends optional="true" config-file="java-support.xml">
    com.intellij.modules.java
</depends>
```

```kotlin
// Check for module availability
fun hasJavaSupport(): Boolean {
    return PluginManagerCore.isPluginInstalled(
        PluginId.getId("com.intellij.modules.java")
    )
}

// Conditional feature
if (hasJavaSupport()) {
    enableJavaFeatures()
}
```

### Product Detection

```kotlin
fun detectProduct(): String {
    val appInfo = ApplicationInfo.getInstance()
    return when {
        appInfo.build.productCode == "IC" -> "IntelliJ Community"
        appInfo.build.productCode == "IU" -> "IntelliJ Ultimate"
        appInfo.build.productCode == "PS" -> "PhpStorm"
        appInfo.build.productCode == "WS" -> "WebStorm"
        appInfo.build.productCode == "AI" -> "Android Studio"
        else -> "Unknown"
    }
}
```

## Dependency Management

### Module Dependencies

```xml
<idea-plugin>
    <!-- Required -->
    <depends>com.intellij.modules.platform</depends>

    <!-- Optional -->
    <depends optional="true" config-file="java.xml">
        com.intellij.modules.java
    </depends>

    <!-- Plugin dependency -->
    <depends optional="true">com.example.otherplugin</depends>
</idea-plugin>
```

### Checking Dependencies

```kotlin
fun checkDependencies(): Boolean {
    val required = listOf(
        "com.intellij.modules.platform",
        "com.intellij.modules.java"
    )

    return required.all {
        PluginManagerCore.isPluginInstalled(PluginId.getId(it))
    }
}
```

## Migration Strategies

### API Migration

```kotlin
// Migration helper
object ApiMigration {
    inline fun <T> withNewApi(newApi: () -> T, oldApi: () -> T): T {
        return try {
            newApi()
        } catch (e: NoSuchMethodError) {
            oldApi()
        }
    }
}

// Usage
val result = ApiMigration.withNewApi(
    newApi = { NewClass.newMethod() },
    oldApi = { OldClass.oldMethod() }
)
```

### Version-Specific Code

```kotlin
fun getVersionSpecificImplementation(): MyInterface {
    return if (ApiCompatibility.isAtLeast(231)) {
        NewImplementation()
    } else {
        LegacyImplementation()
    }
}
```

## Testing Compatibility

### Multi-Version Testing

```kotlin
// build.gradle.kts
tasks {
    runPluginVerifier {
        ideVersions.set(listOf(
            "IC-2022.3",
            "IC-2023.1",
            "IC-2023.2",
            "IC-2023.3"
        ))
    }
}
```

### Compatibility Test

```kotlin
class CompatibilityTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testApiAvailability() {
        // Test that required APIs exist
        assertNotNull(findClass("com.intellij.openapi.application.Application"))
    }

    private fun findClass(name: String): Class<*>? {
        return try {
            Class.forName(name)
        } catch (e: ClassNotFoundException) {
            null
        }
    }
}
```

## Common Compatibility Issues

### Missing Classes

```
Problem: ClassNotFoundException: com.intellij.openapi.old.OldClass
Solution: Use new class or add fallback
```

### Method Signature Changes

```
Problem: NoSuchMethodError: OldClass.oldMethod()
Solution: Update to new method signature
```

### Package Moves

```
Problem: Package not found: com.intellij.old.package
Solution: Import from new package location
```

## Best Practices

### Compatibility Guidelines

1. **Test multiple versions**: Verify with all supported versions
2. **Use Plugin Verifier**: Catch issues early
3. **Handle missing APIs**: Provide fallbacks
4. **Document requirements**: Specify minimum version
5. **Update regularly**: Keep up with platform changes

### Example: Compatibility Layer

```kotlin
// Compatibility.kt
object Compatibility {
    private val isAtLeast231: Boolean by lazy {
        ApplicationInfo.getInstance().build.baselineVersion >= 231
    }

    fun getEditorFactory(): EditorFactory {
        return if (isAtLeast231) {
            EditorFactory.getInstance()
        } else {
            @Suppress("DEPRECATION")
            LegacyEditorFactory.getInstance()
        }
    }

    fun createEditor(document: Document): Editor {
        return getEditorFactory().createEditor(document)
    }
}

// Usage
val editor = Compatibility.createEditor(document)
```

## Related Topics

- [Plugin Configuration](plugin-configuration.md) - Plugin manifest
- [Plugin Verifier](plugin-verifier.md) - Verification tool
- [Publishing](publishing.md) - Publishing to Marketplace

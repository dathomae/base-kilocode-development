# API Changes

This document covers API changes between platform versions and migration strategies.

## Overview

### Why APIs Change

APIs change for:

- Performance improvements
- Security fixes
- Feature enhancements
- Deprecation of old patterns

### Change Types

| Type             | Impact | Description                |
| ---------------- | ------ | -------------------------- |
| Deprecation      | Low    | Marked for removal         |
| Removal          | High   | API removed                |
| Signature change | Medium | Method signature modified  |
| Package move     | Low    | Class moved to new package |

## Migration Strategies

### Handling Deprecation

```kotlin
// Old API (deprecated)
@Deprecated("Use newMethod instead", ReplaceWith("newMethod()"))
fun oldMethod() {
    // Old implementation
}

// New API
fun newMethod() {
    // New implementation
}

// Migration
@Suppress("DEPRECATION")
fun migrate() {
    // Use old API temporarily
    oldMethod()

    // Or use new API
    newMethod()
}
```

### Version-Specific Code

```kotlin
object ApiCompatibility {
    fun isAtLeast(major: Int): Boolean {
        val build = ApplicationInfo.getInstance().build
        return build.baselineVersion >= major
    }
}

// Usage
fun getImplementation(): MyInterface {
    return if (ApiCompatibility.isAtLeast(231)) {
        NewApiImplementation()
    } else {
        LegacyImplementation()
    }
}
```

## Common Migrations

### Application.runWriteAction

```kotlin
// Old API
ApplicationManager.getApplication().runWriteAction {
    // Write operation
}

// New API
WriteAction.run<Exception> {
    // Write operation
}

// With result
val result = WriteAction.compute<Exception, String> {
    // Write operation returning result
    "result"
}
```

### Application.runReadAction

```kotlin
// Old API
ApplicationManager.getApplication().runReadAction {
    // Read operation
}

// New API
ReadAction.run {
    // Read operation
}

// With result
val result = ReadAction.compute<String> {
    // Read operation returning result
    psiFile.text
}
```

### FileDocumentManager

```kotlin
// Old API
FileDocumentManager.getInstance().saveDocument(document)

// New API
FileDocumentManager.getInstance().saveDocument(document)
// Same API, but use with WriteAction
WriteAction.run {
    FileDocumentManager.getInstance().saveDocument(document)
}
```

### DumbService

```kotlin
// Old API
DumbService.getInstance(project).runWhenSmart {
    // Run when smart
}

// New API
DumbService.getInstance(project).runWhenSmart {
    // Same API
}

// With smart mode check
if (DumbService.getInstance(project).isDumb) {
    DumbService.getInstance(project).runWhenSmart {
        // Run when smart
    }
} else {
    // Run now
}
```

## Removed APIs

### Common Removals

| Old API                                      | Replacement                                  | Version |
| -------------------------------------------- | -------------------------------------------- | ------- |
| `Application.runWriteAction()`               | `WriteAction.run()`                          | 2023.1  |
| `Application.runReadAction()`                | `ReadAction.run()`                           | 2023.1  |
| `PsiManager.getInstance(project).findFile()` | `PsiManager.getInstance(project).findFile()` | -       |

### Handling Removed APIs

```kotlin
// Check if API exists
fun hasNewApi(): Boolean {
    return try {
        Class.forName("com.intellij.openapi.application.WriteAction")
        true
    } catch (e: ClassNotFoundException) {
        false
    }
}

// Use appropriate API
fun performWrite() {
    if (hasNewApi()) {
        WriteAction.run<Exception> {
            // Write operation
        }
    } else {
        @Suppress("DEPRECATION")
        ApplicationManager.getApplication().runWriteAction {
            // Write operation
        }
    }
}
```

## Package Moves

### Common Package Changes

| Old Package                | New Package |
| -------------------------- | ----------- |
| `com.intellij.openapi.vfs` | Same        |
| `com.intellij.psi`         | Same        |

### Handling Package Moves

```kotlin
// Update imports
import com.intellij.openapi.application.WriteAction
import com.intellij.openapi.application.ReadAction
```

## Breaking Changes by Version

### 2023.1 Changes

- `Application.runWriteAction()` deprecated
- `Application.runReadAction()` deprecated
- New `WriteAction` and `ReadAction` utilities

### 2023.2 Changes

- Additional deprecations
- Performance improvements

### 2023.3 Changes

- API cleanup
- New extension points

## Migration Tools

### Plugin Verifier

```bash
# Run verifier
./gradlew runPluginVerifier

# Check specific versions
./gradlew runPluginVerifier -PideVersions=IC-2023.1,IC-2023.2
```

### Inspection for Deprecated API

Enable inspection in IDE:

1. Settings → Editor → Inspections
2. Enable "Deprecated API usage"
3. Review warnings

## Best Practices

### API Migration Guidelines

1. **Use new APIs**: Prefer new APIs when available
2. **Suppress warnings**: Use `@Suppress` for temporary use
3. **Test compatibility**: Verify with Plugin Verifier
4. **Document changes**: Note migration in comments
5. **Update gradually**: Migrate incrementally

### Example: Complete Migration

```kotlin
object MigrationHelper {
    private val isNewApiAvailable: Boolean by lazy {
        try {
            Class.forName("com.intellij.openapi.application.WriteAction")
            true
        } catch (e: ClassNotFoundException) {
            false
        }
    }

    fun runWriteAction(action: () -> Unit) {
        if (isNewApiAvailable) {
            WriteAction.run<Exception> { action() }
        } else {
            @Suppress("DEPRECATION")
            ApplicationManager.getApplication().runWriteAction(action)
        }
    }

    fun <T> computeReadAction(computation: () -> T): T {
        return if (isNewApiAvailable) {
            ReadAction.compute<T> { computation() }
        } else {
            @Suppress("DEPRECATION")
            ApplicationManager.getApplication().runReadAction<T>(computation)
        }
    }
}
```

## Related Topics

- [Plugin Compatibility](../09-deployment/plugin-compatibility.md) - Version compatibility
- [Plugin Verifier](../09-deployment/plugin-verifier.md) - Verification tool
- [Troubleshooting](troubleshooting.md) - Common issues

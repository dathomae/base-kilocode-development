# Plugin Verifier

The Plugin Verifier tool checks plugin compatibility with different IDE versions.

## Overview

### What Is Plugin Verifier?

Plugin Verifier:

- Checks binary compatibility
- Detects API changes
- Identifies missing dependencies
- Validates plugin structure

### Why Use Plugin Verifier?

- Catch compatibility issues early
- Ensure plugin works across versions
- Identify deprecated API usage
- Validate plugin before release

## Running Plugin Verifier

### Command Line

```bash
# Download verifier
curl -L -o verifier.jar https://github.com/JetBrains/intellij-plugin-verifier/releases/download/1.265/verifier-1.265.jar

# Run verification
java -jar verifier.jar check-plugin \
    /path/to/plugin.zip \
    /path/to/ide \
    --team-city
```

### Gradle Integration

```kotlin
// build.gradle.kts
tasks {
    val verifyPlugin by verifying {
        // Automatically run during build
    }

    runPluginVerifier {
        ideVersions.set(listOf(
            "IC-2023.1",
            "IC-2023.2",
            "IC-2023.3"
        ))

        failureLevel.set(listOf(
            FailureLevel.COMPATIBILITY_WARNINGS,
            FailureLevel.COMPATIBILITY_PROBLEMS,
            FailureLevel.MISSING_DEPENDENCIES
        ))
    }
}
```

### Running from Gradle

```bash
# Run verification
./gradlew runPluginVerifier

# Check specific versions
./gradlew runPluginVerifier -PideVersions=IC-2023.1,IC-2023.2
```

## Verification Results

### Output Format

```
Plugin MyPlugin-1.0.0 against IC-231.8109.175:
  Compatible: true
  Warnings: 2
  Problems: 0

Warnings:
  1. Deprecated API usage: OldClass.oldMethod()
     at com.example.MyClass.myMethod(MyClass.kt:42)
     Solution: Use NewClass.newMethod() instead

  2. Experimental API usage: ExperimentalFeature
     at com.example.MyClass.useFeature(MyClass.kt:100)
     Solution: Annotate with @OptIn(ExperimentalFeature::class)
```

### Problem Types

| Problem             | Severity | Description                |
| ------------------- | -------- | -------------------------- |
| Missing dependency  | Error    | Required plugin not found  |
| Incompatible change | Error    | Breaking API change        |
| Deprecated API      | Warning  | Using deprecated method    |
| Experimental API    | Warning  | Using experimental feature |

## Common Issues

### Missing Dependencies

```
Problem: Missing dependency: com.intellij.modules.java
Solution: Add to plugin.xml:
  <depends>com.intellij.modules.java</depends>
```

### Incompatible API

```
Problem: Method removed: OldUtil.process()
Solution: Use NewUtil.process() instead
```

### Deprecated API

```
Problem: Deprecated method: Application.runWriteAction()
Solution: Use WriteAction.run() instead
```

## Fixing Issues

### Binary Compatibility

```kotlin
// Problem: Using removed API
val result = OldApi.process()

// Solution: Use new API
val result = NewApi.process()
```

### Missing Dependencies

```xml
<!-- Problem: Missing dependency -->
<idea-plugin>
    <depends>com.intellij.modules.platform</depends>
</idea-plugin>

<!-- Solution: Add required dependency -->
<idea-plugin>
    <depends>com.intellij.modules.platform</depends>
    <depends>com.intellij.modules.java</depends>
</idea-plugin>
```

### Deprecated Methods

```kotlin
// Problem: Deprecated method
ApplicationManager.getApplication().runWriteAction {
    // ...
}

// Solution: Use new API
WriteAction.run<Exception> {
    // ...
}
```

## Verification Levels

### Failure Levels

```kotlin
runPluginVerifier {
    failureLevel.set(listOf(
        // Fail on compatibility problems
        FailureLevel.COMPATIBILITY_PROBLEMS,

        // Fail on missing dependencies
        FailureLevel.MISTING_DEPENDENCIES,

        // Fail on deprecated API warnings
        FailureLevel.DEPRECATED_API_USAGES,

        // Fail on experimental API usage
        FailureLevel.EXPERIMENTAL_API_USAGES
    ))
}
```

### Ignoring Warnings

```kotlin
runPluginVerifier {
    // Only fail on errors
    failureLevel.set(listOf(
        FailureLevel.COMPATIBILITY_PROBLEMS,
        FailureLevel.MISSING_DEPENDENCIES
    ))

    // Ignore specific warnings
    ignoredProblems.set(listOf(
        "Deprecated method in MyOldClass"
    ))
}
```

## CI Integration

### GitHub Actions

```yaml
name: Verify Plugin

on: [push, pull_request]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Java
        uses: actions/setup-java@v3
        with:
          java-version: 17

      - name: Run Plugin Verifier
        run: ./gradlew runPluginVerifier

      - name: Upload Results
        uses: actions/upload-artifact@v3
        with:
          name: verification-results
          path: build/reports/pluginVerifier/
```

### TeamCity

```kotlin
// TeamCity configuration
project {
    buildType {
        id("Verify")
        name = "Verify Plugin"

        steps {
            gradle {
                tasks = "runPluginVerifier"
            }
        }

        artifactRules = "build/reports/pluginVerifier/**"
    }
}
```

## IDE Versions

### Specifying Versions

```kotlin
runPluginVerifier {
    ideVersions.set(listOf(
        // IntelliJ Community
        "IC-2022.3",
        "IC-2023.1",
        "IC-2023.2",

        // IntelliJ Ultimate
        "IU-2023.1",

        // Android Studio
        "AI-2022.3.1"
    ))
}
```

### Version Aliases

| Alias | Description        |
| ----- | ------------------ |
| `IC`  | IntelliJ Community |
| `IU`  | IntelliJ Ultimate  |
| `AI`  | Android Studio     |
| `PS`  | PhpStorm           |
| `WS`  | WebStorm           |

## Best Practices

### Verification Guidelines

1. **Run before release**: Always verify before publishing
2. **Test multiple versions**: Check all supported IDE versions
3. **Fix all errors**: Address compatibility problems
4. **Review warnings**: Evaluate deprecated API usage
5. **Automate**: Integrate with CI pipeline

### Example: Complete Verification Setup

```kotlin
// build.gradle.kts
intellij {
    type.set("IC")
    version.set("2023.1")
}

tasks {
    runPluginVerifier {
        ideVersions.set(listOf(
            "IC-2022.3",
            "IC-2023.1",
            "IC-2023.2",
            "IC-2023.3"
        ))

        failureLevel.set(listOf(
            FailureLevel.COMPATIBILITY_PROBLEMS,
            FailureLevel.MISSING_DEPENDENCIES
        ))

        // Ignore known issues
        ignoredProblems.set(listOf(
            "Deprecated API in legacy support"
        ))
    }

    // Run verification as part of build
    build {
        dependsOn(runPluginVerifier)
    }
}
```

## Related Topics

- [Plugin Compatibility](plugin-compatibility.md) - Version compatibility
- [Plugin Configuration](plugin-configuration.md) - Plugin manifest
- [Publishing](publishing.md) - Publishing to Marketplace

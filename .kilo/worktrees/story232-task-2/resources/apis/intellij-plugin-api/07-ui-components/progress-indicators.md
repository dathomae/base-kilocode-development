# Progress Indicators

Progress indicators show operation progress and allow cancellation.

## Overview

### What Are Progress Indicators?

Progress indicators provide:

- Visual feedback for long-running operations
- Cancellation support
- Background task execution
- Progress percentage display

### Types of Indicators

| Type      | Description              |
| --------- | ------------------------ |
| Modal     | Blocks UI until complete |
| Non-modal | Runs in background       |
| Silent    | No visible indicator     |

## Basic Progress

### Running with Progress

```kotlin
fun runWithProgress(project: Project) {
    ProgressManager.getInstance().run(object : Task.Backgroundable(project, "Processing") {
        override fun run(indicator: ProgressIndicator) {
            indicator.text = "Starting..."
            indicator.fraction = 0.0

            for (i in 1..100) {
                // Check for cancellation
                if (indicator.isCanceled) {
                    return
                }

                // Update progress
                indicator.text = "Processing item $i"
                indicator.fraction = i / 100.0

                // Do work
                processItem(i)
            }

            indicator.fraction = 1.0
            indicator.text = "Complete"
        }
    })
}
```

### Modal Progress

```kotlin
fun runModalProgress(project: Project) {
    ProgressManager.getInstance().run(object : Task.Modal(project, "Processing", true) {
        override fun run(indicator: ProgressIndicator) {
            indicator.text = "Processing..."

            for (i in 1..100) {
                if (indicator.isCanceled) {
                    return
                }

                indicator.fraction = i / 100.0
                processItem(i)
            }
        }

        override fun onSuccess() {
            // Called on success
            Notifications.showInfo(project, "Processing complete")
        }

        override fun onThrowable(error: Throwable) {
            // Called on error
            Notifications.showError(project, "Error: ${error.message}")
        }
    })
}
```

## Progress Indicator API

### Setting Progress

```kotlin
fun updateProgress(indicator: ProgressIndicator) {
    // Set text
    indicator.text = "Main text"
    indicator.text2 = "Secondary text"

    // Set fraction (0.0 to 1.0)
    indicator.fraction = 0.5

    // Set indeterminate
    indicator.isIndeterminate = true
}
```

### Checking Cancellation

```kotlin
fun runWithCancellation(indicator: ProgressIndicator) {
    for (item in items) {
        // Check cancellation
        if (indicator.isCanceled) {
            return
        }

        // Alternative: throw exception
        ProgressManager.checkCanceled()

        processItem(item)
    }
}
```

## Background Tasks

### Background Task

```kotlin
fun runBackgroundTask(project: Project) {
    ProgressManager.getInstance().run(object : Task.Backgroundable(
        project,
        "Background Task",
        true,  // Can cancel
        PerformInBackgroundOption.ALWAYS_BACKGROUND
    ) {
        override fun run(indicator: ProgressIndicator) {
            indicator.text = "Processing..."

            // Long-running operation
            for (i in 1..1000) {
                ProgressManager.checkCanceled()
                indicator.fraction = i / 1000.0
                processItem(i)
            }
        }

        override fun onSuccess() {
            // Called on success
        }

        override fun onThrowable(error: Throwable) {
            // Called on error
        }

        override fun onFinished() {
            // Always called
        }
    })
}
```

### Conditional Background

```kotlin
fun runConditionalBackground(project: Project) {
    val task = object : Task.Backgroundable(project, "Task") {
        override fun run(indicator: ProgressIndicator) {
            // Work
        }
    }

    // Run in background if estimated time > threshold
    ProgressManager.getInstance().run(task)
}
```

## Progress in Dialogs

### Progress Dialog

```kotlin
fun runWithProgressDialog(project: Project): Boolean {
    return ProgressManager.getInstance().runProcessWithProgressSynchronously(
        { indicator ->
            indicator.text = "Processing..."
            for (i in 1..100) {
                ProgressManager.checkCanceled()
                indicator.fraction = i / 100.0
                processItem(i)
            }
        },
        "Processing",
        true,  // Can cancel
        project
    )
}
```

### Progress with Result

```kotlin
fun <T> runWithResult(project: Project, operation: (ProgressIndicator) -> T): T? {
    var result: T? = null

    ProgressManager.getInstance().run(object : Task.WithResult<T, Exception>(project, "Processing", true) {
        override fun compute(indicator: ProgressIndicator): T {
            return operation(indicator)
        }
    })

    return result
}
```

## Nested Progress

### Splitting Progress

```kotlin
fun runNestedProgress(indicator: ProgressIndicator) {
    val items = getItems()
    val total = items.size.toDouble()

    items.forEachIndexed { index, item ->
        ProgressManager.checkCanceled()

        // Update main progress
        indicator.text = "Processing ${item.name}"
        indicator.fraction = index / total

        // Process sub-items
        val subItems = item.subItems
        subItems.forEachIndexed { subIndex, subItem ->
            ProgressManager.checkCanceled()

            indicator.text2 = "Sub-item ${subIndex + 1}/${subItems.size}"

            // Calculate combined fraction
            val subFraction = (index + subIndex / subItems.size.toDouble()) / total
            indicator.fraction = subFraction

            processSubItem(subItem)
        }
    }
}
```

### Progress Wrapper

```kotlin
class ProgressWrapper(
    private val indicator: ProgressIndicator,
    private val startFraction: Double,
    private val endFraction: Double
) : ProgressIndicator {
    override fun setFraction(fraction: Double) {
        indicator.fraction = startFraction + (endFraction - startFraction) * fraction
    }

    override fun getFraction(): Double = indicator.fraction

    override fun setText(text: String?) {
        indicator.text = text
    }

    override fun getText(): String? = indicator.text

    override fun isCanceled(): Boolean = indicator.isCanceled

    override fun cancel() {
        indicator.cancel()
    }

    // Delegate other methods...
}
```

## Silent Progress

### Running Silently

```kotlin
fun runSilently(project: Project) {
    ProgressManager.getInstance().runProcessWithProgressAsynchronously(
        object : Task.Backgroundable(project, "Task", false) {
            override fun run(indicator: ProgressIndicator) {
                // Work without visible progress
            }
        },
        EmptyProgressIndicator()
    )
}
```

### Empty Progress Indicator

```kotlin
fun runWithEmptyProgress() {
    val indicator = EmptyProgressIndicator()

    // Can still check cancellation
    if (indicator.isCanceled) {
        return
    }

    // But no visual feedback
}
```

## Thread Safety

### Running on EDT

```kotlin
override fun run(indicator: ProgressIndicator) {
    // Background thread

    // Update UI on EDT
    ApplicationManager.getApplication().invokeLater {
        // UI updates
    }

    // Or use invokeAndWait
    ApplicationManager.getApplication().invokeAndWait {
        // UI updates that must complete before continuing
    }
}
```

### Read/Write Actions

```kotlin
override fun run(indicator: ProgressIndicator) {
    // Read action
    val result = ReadAction.compute<Exception> {
        // Read PSI
        psiFile.text
    }

    // Write action
    WriteAction.run<Exception> {
        // Modify PSI
        psiFile.delete()
    }
}
```

## Best Practices

### Progress Guidelines

1. **Check cancellation frequently**: Allow users to cancel
2. **Update progress regularly**: Provide visual feedback
3. **Use appropriate text**: Describe current operation
4. **Handle exceptions**: Show meaningful errors
5. **Clean up resources**: Ensure proper cleanup

### Example: Complete Progress Task

```kotlin
class MyProgressTask(
    project: Project,
    private val items: List<Item>
) : Task.Backgroundable(project, "Processing Items", true) {
    private var processedCount = 0
    private var failedCount = 0

    override fun run(indicator: ProgressIndicator) {
        indicator.isIndeterminate = false
        indicator.text = "Starting..."

        val total = items.size.toDouble()

        for ((index, item) in items.withIndex()) {
            // Check cancellation
            ProgressManager.checkCanceled()

            // Update progress
            indicator.text = "Processing ${item.name}"
            indicator.text2 = "Item ${index + 1} of ${items.size}"
            indicator.fraction = index / total

            try {
                // Process item
                processItem(item)
                processedCount++
            } catch (e: Exception) {
                failedCount++
                // Log error but continue
                Logger.getInstance(javaClass).warn("Failed to process ${item.name}", e)
            }
        }

        indicator.fraction = 1.0
        indicator.text = "Complete"
    }

    private fun processItem(item: Item) {
        // Item processing logic
    }

    override fun onSuccess() {
        Notifications.showInfo(
            project,
            "Processed $processedCount items (${failedCount} failed)"
        )
    }

    override fun onThrowable(error: Throwable) {
        Notifications.showError(
            project,
            "Error: ${error.message}"
        )
    }

    override fun onFinished() {
        // Cleanup
    }
}

// Usage
fun startProcessing(project: Project, items: List<Item>) {
    ProgressManager.getInstance().run(MyProgressTask(project, items))
}
```

## Related Topics

- [Services](../01-core-concepts/services.md) - Background services
- [Tool Windows](tool-windows.md) - Tool window content
- [Status Bar Widgets](status-bar-widgets.md) - Status display

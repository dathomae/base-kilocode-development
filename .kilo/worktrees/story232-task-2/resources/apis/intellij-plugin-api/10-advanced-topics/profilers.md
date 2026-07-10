# Profilers

Profiler integration allows plugins to provide performance analysis tools.

## Overview

### What Are Profilers?

Profilers provide:

- CPU usage analysis
- Memory allocation tracking
- Thread analysis
- Performance metrics

### Profiler Types

| Type            | Description           |
| --------------- | --------------------- |
| CPU Profiler    | Method execution time |
| Memory Profiler | Allocation tracking   |
| Thread Profiler | Thread state analysis |

## Profiler Integration

### Profiler Action

```kotlin
class MyProfilerAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return

        // Start profiling
        val profiler = MyProfilerService.getInstance(project)
        profiler.startProfiling()
    }

    override fun update(e: AnActionEvent) {
        val project = e.project
        e.presentation.isEnabled = project != null
    }
}
```

### Profiler Service

```kotlin
class MyProfilerService(val project: Project) {
    private var isProfiling = false
    private val data = mutableListOf<ProfileData>()

    fun startProfiling() {
        if (isProfiling) return

        isProfiling = true
        data.clear()

        // Attach profiler
        attachProfiler()
    }

    fun stopProfiling(): List<ProfileData> {
        if (!isProfiling) return emptyList()

        isProfiling = false

        // Detach profiler
        detachProfiler()

        return data.toList()
    }

    private fun attachProfiler() {
        // Attach to process
    }

    private fun detachProfiler() {
        // Detach from process
    }

    companion object {
        fun getInstance(project: Project): MyProfilerService {
            return project.getService(MyProfilerService::class.java)
        }
    }
}
```

## CPU Profiling

### Capturing CPU Data

```kotlin
class CpuProfiler {
    private val samples = mutableListOf<MethodSample>()

    fun sample(): MethodSample {
        // Capture stack trace
        val stackTrace = Thread.currentThread().stackTrace

        return MethodSample(
            threadId = Thread.currentThread().id,
            stackTrace = stackTrace.map { it.toFrame() },
            timestamp = System.nanoTime()
        )
    }

    fun startSampling(intervalMs: Long) {
        Thread {
            while (true) {
                samples.add(sample())
                Thread.sleep(intervalMs)
            }
        }.start()
    }
}

data class MethodSample(
    val threadId: Long,
    val stackTrace: List<StackFrame>,
    val timestamp: Long
)

data class StackFrame(
    val className: String,
    val methodName: String,
    val fileName: String?,
    val lineNumber: Int
)
```

### Analyzing Results

```kotlin
class CpuAnalysis(private val samples: List<MethodSample>) {
    fun getHotMethods(): List<MethodStats> {
        val methodCounts = mutableMapOf<String, Int>()

        for (sample in samples) {
            for (frame in sample.stackTrace) {
                val key = "${frame.className}.${frame.methodName}"
                methodCounts[key] = methodCounts.getOrDefault(key, 0) + 1
            }
        }

        return methodCounts.map { (method, count) ->
            MethodStats(
                method = method,
                sampleCount = count,
                percentage = count.toDouble() / samples.size * 100
            )
        }.sortedByDescending { it.sampleCount }
    }

    fun getCallTree(): CallTreeNode {
        val root = CallTreeNode("root")

        for (sample in samples) {
            var node = root
            for (frame in sample.stackTrace.reversed()) {
                node = node.getOrCreateChild(frame.methodName)
                node.sampleCount++
            }
        }

        return root
    }
}

data class MethodStats(
    val method: String,
    val sampleCount: Int,
    val percentage: Double
)

class CallTreeNode(val name: String) {
    val children = mutableMapOf<String, CallTreeNode>()
    var sampleCount = 0

    fun getOrCreateChild(name: String): CallTreeNode {
        return children.getOrPut(name) { CallTreeNode(name) }
    }
}
```

## Memory Profiling

### Memory Snapshot

```kotlin
class MemoryProfiler {
    fun captureSnapshot(): MemorySnapshot {
        val runtime = Runtime.getRuntime()

        return MemorySnapshot(
            totalMemory = runtime.totalMemory(),
            freeMemory = runtime.freeMemory(),
            usedMemory = runtime.totalMemory() - runtime.freeMemory(),
            maxMemory = runtime.maxMemory(),
            timestamp = System.currentTimeMillis()
        )
    }

    fun forceGc() {
        System.gc()
    }
}

data class MemorySnapshot(
    val totalMemory: Long,
    val freeMemory: Long,
    val usedMemory: Long,
    val maxMemory: Long,
    val timestamp: Long
)
```

### Object Allocation Tracking

```kotlin
class AllocationTracker {
    private val allocations = mutableListOf<AllocationEvent>()

    fun recordAllocation(
        className: String,
        size: Long,
        stackTrace: List<StackFrame>
    ) {
        allocations.add(AllocationEvent(
            className = className,
            size = size,
            stackTrace = stackTrace,
            timestamp = System.nanoTime()
        ))
    }

    fun getAllocationSummary(): Map<String, AllocationSummary> {
        return allocations
            .groupBy { it.className }
            .mapValues { (_, events) ->
                AllocationSummary(
                    className = events.first().className,
                    count = events.size,
                    totalSize = events.sumOf { it.size }
                )
            }
    }
}

data class AllocationEvent(
    val className: String,
    val size: Long,
    val stackTrace: List<StackFrame>,
    val timestamp: Long
)

data class AllocationSummary(
    val className: String,
    val count: Int,
    val totalSize: Long
)
```

## Profiler UI

### Displaying Results

```kotlin
class ProfilerResultsPanel(
    private val project: Project
) : JPanel(BorderLayout()) {
    private val tree = JBTree()

    fun displayResults(results: List<MethodStats>) {
        val root = DefaultMutableTreeNode("Hot Methods")

        for (stat in results.take(20)) {
            root.add(DefaultMutableTreeNode(
                "${stat.method} (${String.format("%.1f", stat.percentage)}%)"
            ))
        }

        tree.model = DefaultTreeModel(root)
    }
}
```

### Flame Graph

```kotlin
class FlameGraphPanel(
    private val data: List<MethodSample>
) : JPanel() {
    override fun paintComponent(g: Graphics) {
        super.paintComponent(g)

        val g2d = g as Graphics2D
        val width = width
        val height = height

        // Draw flame graph
        drawFlameGraph(g2d, width, height)
    }

    private fun drawFlameGraph(g: Graphics2D, width: Int, height: Int) {
        val callTree = CpuAnalysis(data).getCallTree()
        drawNode(g, callTree, 0, width, 0)
    }

    private fun drawNode(
        g: Graphics2D,
        node: CallTreeNode,
        x: Int,
        width: Int,
        depth: Int
    ) {
        val rowHeight = 20
        val y = depth * rowHeight

        // Draw rectangle
        g.color = getColorForMethod(node.name)
        g.fillRect(x, y, width, rowHeight - 1)

        // Draw label
        g.color = Color.BLACK
        g.drawString(node.name, x + 2, y + rowHeight - 5)

        // Draw children
        var childX = x
        for ((_, child) in node.children) {
            val childWidth = (child.sampleCount.toDouble() / node.sampleCount * width).toInt()
            drawNode(g, child, childX, childWidth, depth + 1)
            childX += childWidth
        }
    }

    private fun getColorForMethod(name: String): Color {
        // Generate color based on method name
        val hash = name.hashCode()
        return Color(
            (hash and 0xFF0000 shr 16) % 128 + 128,
            (hash and 0x00FF00 shr 8) % 128 + 128,
            (hash and 0x0000FF) % 128 + 128
        )
    }
}
```

## Integration with IDE

### Tool Window

```kotlin
class ProfilerToolWindowFactory : ToolWindowFactory {
    override fun createToolWindowContent(
        project: Project,
        toolWindow: ToolWindow
    ) {
        val panel = ProfilerPanel(project)
        val content = ContentFactory.getInstance()
            .createContent(panel, "", false)

        toolWindow.contentManager.addContent(content)
    }
}
```

### Registering Profiler

```xml
<extensions defaultExtensionNs="com.intellij">
    <toolWindow
        id="Profiler"
        anchor="bottom"
        factoryClass="com.example.ProfilerToolWindowFactory"/>

    <projectService
        serviceImplementation="com.example.MyProfilerService"/>
</extensions>
```

## Best Practices

### Profiler Guidelines

1. **Minimize overhead**: Keep profiling impact low
2. **Sample efficiently**: Use appropriate sampling rates
3. **Present clearly**: Make results easy to understand
4. **Support export**: Allow saving results
5. **Handle large data**: Support large profiling sessions

### Example: Complete Profiler

```kotlin
class MyCompleteProfiler(val project: Project) {
    private val cpuProfiler = CpuProfiler()
    private val memoryProfiler = MemoryProfiler()
    private var isRunning = false

    fun start() {
        if (isRunning) return
        isRunning = true

        cpuProfiler.startSampling(10)
    }

    fun stop(): ProfileResults {
        isRunning = false

        val cpuResults = cpuProfiler.getResults()
        val memorySnapshot = memoryProfiler.captureSnapshot()

        return ProfileResults(
            cpuData = cpuResults,
            memoryData = memorySnapshot
        )
    }
}

data class ProfileResults(
    val cpuData: List<MethodSample>,
    val memoryData: MemorySnapshot
)
```

## Related Topics

- [Run Configurations](run-configurations.md) - Run configuration setup
- [Debuggers](debuggers.md) - Debugger integration
- [Tool Windows](../07-ui-components/tool-windows.md) - Tool window creation

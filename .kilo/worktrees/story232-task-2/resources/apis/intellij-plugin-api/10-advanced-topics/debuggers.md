# Debuggers

Debugger integration allows plugins to provide custom debugging support for languages and runtimes.

## Overview

### What Is Debugger Integration?

Debugger integration provides:

- Breakpoint management
- Variable inspection
- Step execution
- Expression evaluation
- Stack frame navigation

### Debugger Architecture

| Component       | Description      |
| --------------- | ---------------- |
| Debug Process   | Running debuggee |
| Suspend Context | Suspended state  |
| Stack Frame     | Call stack entry |
| Variable        | Debug value      |

## Basic Debugger Support

### Debug Process

```kotlin
class MyDebugProcess(
    private val session: XDebugSession
) : XDebugProcess(session) {

    override fun startPausing() {
        // Pause execution
    }

    override fun startResuming() {
        // Resume execution
    }

    override fun startStepOver(context: XSuspendContext?) {
        // Step over
    }

    override fun startStepInto(context: XSuspendContext?) {
        // Step into
    }

    override fun startStepOut(context: XSuspendContext?) {
        // Step out
    }

    override fun stop() {
        // Stop debugging
    }
}
```

### Debug Process Stopped

```kotlin
class MySuspendContext(
    private val stackFrames: List<XStackFrame>
) : XSuspendContext() {

    override fun getActiveExecutionStack(): XExecutionStack? {
        return MyExecutionStack(stackFrames)
    }

    override fun getExecutionStacks(): Array<XExecutionStack> {
        return arrayOf(getActiveExecutionStack()!!)
    }
}
```

## Stack Frames

### Custom Stack Frame

```kotlin
class MyStackFrame(
    private val frame: MyFrameInfo
) : XStackFrame() {

    override fun getSourcePosition(): XSourcePosition? {
        return XSourcePositionImpl.create(
            frame.file,
            frame.lineNumber
        )
    }

    override fun getEvaluator(): XValueEvaluator? {
        return MyValueEvaluator(frame)
    }

    override fun computeChildren(node: CompositeNode) {
        val children = XValueChildrenList()

        for (variable in frame.variables) {
            children.add(MyValue(variable))
        }

        node.addChildren(children, true)
    }

    override fun getPresentation(): XValuePresentation {
        return XValuePresentation() {
            it.append(frame.methodName)
            it.append("()")
        }
    }
}
```

### Execution Stack

```kotlin
class MyExecutionStack(
    private val frames: List<XStackFrame>
) : XExecutionStack("My Stack") {

    override fun computeStackFrames(
        firstFrameIndex: Int,
        container: XStackFrameContainer
    ) {
        container.addStackFrames(frames, true)
    }

    override fun getTopFrame(): XStackFrame? {
        return frames.firstOrNull()
    }
}
```

## Variables

### Custom Value

```kotlin
class MyValue(
    private val variable: MyVariableInfo
) : XValue() {

    override fun getPresentation(): XValuePresentation {
        return XValuePresentation() {
            it.append(variable.name)
            it.append(" = ")
            it.append(variable.value)
        }
    }

    override fun canNavigateToSource(): Boolean = true

    override fun computeSourcePosition(): XSourcePosition? {
        return XSourcePositionImpl.create(
            variable.declarationFile,
            variable.declarationLine
        )
    }

    override fun computeChildren(node: CompositeNode) {
        if (variable.hasChildren) {
            val children = XValueChildrenList()
            for (child in variable.children) {
                children.add(MyValue(child))
            }
            node.addChildren(children, true)
        } else {
            node.setEmpty()
        }
    }

    override fun getEvaluator(): XValueEvaluator? {
        return MyValueEvaluator(variable)
    }
}
```

### Value Evaluation

```kotlin
class MyValueEvaluator(
    private val variable: MyVariableInfo
) : XValueEvaluator {

    override fun evaluate(expression: String, callback: XValueCallback) {
        try {
            val result = evaluateExpression(expression)
            callback.evaluated(MyValue(result))
        } catch (e: Exception) {
            callback.errorOccurred(e.message ?: "Evaluation error")
        }
    }

    private fun evaluateExpression(expression: String): MyVariableInfo {
        // Evaluate expression
        return MyVariableInfo(expression, "result")
    }
}
```

## Breakpoints

### Custom Breakpoint Type

```kotlin
class MyBreakpointType : XBreakpointType<MyBreakpointProperties>(
    "MyBreakpoint",
    "My Breakpoint"
) {
    override fun getEnabledIcon(): Icon = AllIcons.Debugger.Db_set_breakpoint

    override fun createBreakpointProperties(): MyBreakpointProperties {
        return MyBreakpointProperties()
    }

    override fun canPutAt(
        file: VirtualFile,
        line: Int,
        project: Project
    ): Boolean {
        // Check if breakpoint can be set
        return file.extension == "my"
    }
}

class MyBreakpointProperties : XBreakpointProperties<MyBreakpointProperties>() {
    var condition: String = ""

    override fun getState(): MyBreakpointProperties = this

    override fun loadState(state: MyBreakpointProperties) {
        condition = state.condition
    }
}
```

### Breakpoint Handler

```kotlin
class MyBreakpointHandler(
    private val debugProcess: MyDebugProcess
) : XBreakpointHandler<XLineBreakpoint<MyBreakpointProperties>>(
    MyBreakpointType::class.java
) {
    private val breakpoints = mutableMapOf<XLineBreakpoint<*>, Int>()

    override fun registerBreakpoint(breakpoint: XLineBreakpoint<MyBreakpointProperties>) {
        val position = breakpoint.sourcePosition ?: return
        val id = debugProcess.addBreakpoint(position)
        breakpoints[breakpoint] = id
    }

    override fun unregisterBreakpoint(
        breakpoint: XLineBreakpoint<MyBreakpointProperties>,
        temporary: Boolean
    ) {
        val id = breakpoints.remove(breakpoint) ?: return
        debugProcess.removeBreakpoint(id)
    }
}
```

## Debug Session

### Creating Debug Session

```kotlin
class MyDebugProcessFactory : XDebugProcessFactory() {
    override fun createDebugProcess(
        session: XDebugSession,
        runProfile: RunProfile
    ): XDebugProcess? {
        if (runProfile !is MyRunConfiguration) return null

        return MyDebugProcess(session)
    }
}

class MyDebugProcess(session: XDebugSession) : XDebugProcess(session) {
    private val breakpointHandler = MyBreakpointHandler(this)

    override fun getBreakpointHandlers(): Array<XBreakpointHandler<*>> {
        return arrayOf(breakpointHandler)
    }

    override fun sessionInitialized() {
        // Start debugging
        startDebugging()
    }

    private fun startDebugging() {
        // Launch process
        val handler = startProcess()

        // Wait for connection
        handler.addProcessListener(object : ProcessAdapter() {
            override fun onTextAvailable(event: ProcessEvent, outputType: Key<*>) {
                if (event.text.contains("Debugger connected")) {
                    session.setBreakpointMuted(false)
                }
            }
        })
    }
}
```

### Registering Debugger

```xml
<extensions defaultExtensionNs="com.intellij">
    <xdebugger.breakpointType
        implementation="com.example.MyBreakpointType"/>

    <xdebugger.debugProcessFactory
        implementation="com.example.MyDebugProcessFactory"/>
</extensions>
```

## Best Practices

### Debugger Guidelines

1. **Handle errors gracefully**: Provide clear error messages
2. **Support async operations**: Use callbacks appropriately
3. **Cache values**: Avoid repeated evaluation
4. **Handle disconnection**: Clean up on disconnect
5. **Support hot reload**: Allow code changes during debug

### Example: Complete Debug Process

```kotlin
class MyCompleteDebugProcess(session: XDebugSession) : XDebugProcess(session) {
    private val breakpointHandler = MyBreakpointHandler(this)
    private var processHandler: ProcessHandler? = null

    override fun getBreakpointHandlers() = arrayOf<XBreakpointHandler<*>>(breakpointHandler)

    override fun startPausing() {
        sendCommand("pause")
    }

    override fun startResuming() {
        sendCommand("continue")
    }

    override fun startStepOver(context: XSuspendContext?) {
        sendCommand("stepOver")
    }

    override fun startStepInto(context: XSuspendContext?) {
        sendCommand("stepInto")
    }

    override fun startStepOut(context: XSuspendContext?) {
        sendCommand("stepOut")
    }

    override fun stop() {
        sendCommand("stop")
        processHandler?.destroyProcess()
    }

    fun handleBreakpointHit(frameInfo: MyFrameInfo) {
        val frames = frameInfo.stackFrames.map { MyStackFrame(it) }
        val context = MySuspendContext(frames)
        session.positionReached(context)
    }

    private fun sendCommand(command: String) {
        processHandler?.processInput?.writer()?.println(command)
    }
}
```

## Related Topics

- [Run Configurations](run-configurations.md) - Run configuration setup
- [External Tools](external-tools.md) - External tool integration
- [Actions System](../01-core-concepts/actions-system.md) - Action registration

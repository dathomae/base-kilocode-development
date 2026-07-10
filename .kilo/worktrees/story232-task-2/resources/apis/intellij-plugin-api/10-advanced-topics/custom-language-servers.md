# Custom Language Servers

Language Server Protocol (LSP) integration allows plugins to provide language support via language servers.

## Overview

### What Is LSP?

Language Server Protocol provides:

- Code completion
- Diagnostics
- Go to definition
- Find references
- Rename refactoring

### LSP Benefits

| Benefit           | Description                      |
| ----------------- | -------------------------------- |
| Reusability       | Same server for multiple editors |
| Language agnostic | Any language support             |
| Standard protocol | Well-defined API                 |

## LSP Client Integration

### LSP Server Definition

```kotlin
class MyLanguageServerDefinition : LanguageServerDefinition {
    override fun createServer(
        project: Project,
        client: LanguageClient
    ): LanguageServer {
        val commandLine = GeneralCommandLine()
            .withExePath("my-language-server")
            .withParameters("--stdio")

        val process = commandLine.createProcess()

        return MyLanguageServer(process, client)
    }

    override fun getLanguageIds(): Set<String> {
        return setOf("my-language")
    }

    override fun getCompletionTriggerCharacters(): Set<String> {
        return setOf(".", "@")
    }
}
```

### Language Server Implementation

```kotlin
class MyLanguageServer(
    process: Process,
    client: LanguageClient
) : LanguageServer {
    private val input = process.inputStream
    private val output = process.outputStream

    override fun initialize(params: InitializeParams): CompletableFuture<InitializeResult> {
        val request = Request(
            id = generateId(),
            method = "initialize",
            params = params
        )

        return sendRequest(request).thenApply { response ->
            gson.fromJson(response, InitializeResult::class.java)
        }
    }

    override fun shutdown(): CompletableFuture<Any> {
        val request = Request(
            id = generateId(),
            method = "shutdown",
            params = null
        )

        return sendRequest(request)
    }

    override fun exit() {
        val notification = Notification(
            method = "exit",
            params = null
        )

        sendNotification(notification)
    }

    private fun sendRequest(request: Request): CompletableFuture<JsonElement> {
        // Send request and return future
    }

    private fun sendNotification(notification: Notification) {
        // Send notification
    }
}
```

## LSP Features

### Text Document Synchronization

```kotlin
class MyTextDocumentService {
    fun didOpen(params: DidOpenTextDocumentParams) {
        // Notify server of opened document
    }

    fun didChange(params: DidChangeTextDocumentParams) {
        // Notify server of document changes
    }

    fun didClose(params: DidCloseTextDocumentParams) {
        // Notify server of closed document
    }

    fun didSave(params: DidSaveTextDocumentParams) {
        // Notify server of saved document
    }
}
```

### Completion Support

```kotlin
class MyCompletionSupport {
    fun completion(params: CompletionParams): CompletableFuture<CompletionList> {
        val request = Request(
            id = generateId(),
            method = "textDocument/completion",
            params = params
        )

        return sendRequest(request).thenApply { response ->
            gson.fromJson(response, CompletionList::class.java)
        }
    }

    fun resolveCompletion(item: CompletionItem): CompletableFuture<CompletionItem> {
        val request = Request(
            id = generateId(),
            method = "completionItem/resolve",
            params = item
        )

        return sendRequest(request).thenApply { response ->
            gson.fromJson(response, CompletionItem::class.java)
        }
    }
}
```

### Diagnostics

```kotlin
class MyDiagnosticsHandler {
    fun handleDiagnostics(params: PublishDiagnosticsParams) {
        val file = findFile(params.uri) ?: return
        val diagnostics = params.diagnostics

        // Convert to IntelliJ annotations
        val annotations = diagnostics.map { diagnostic ->
            DiagnosticAnnotation(
                file = file,
                range = diagnostic.range,
                message = diagnostic.message,
                severity = convertSeverity(diagnostic.severity)
            )
        }

        // Apply annotations
        applyAnnotations(file, annotations)
    }

    private fun convertSeverity(severity: DiagnosticSeverity?): HighlightSeverity {
        return when (severity) {
            DiagnosticSeverity.Error -> HighlightSeverity.ERROR
            DiagnosticSeverity.Warning -> HighlightSeverity.WARNING
            DiagnosticSeverity.Information -> HighlightSeverity.INFORMATION
            DiagnosticSeverity.Hint -> HighlightSeverity.INFORMATION
            else -> HighlightSeverity.INFORMATION
        }
    }
}
```

### Go to Definition

```kotlin
class MyDefinitionSupport {
    fun definition(params: DefinitionParams): CompletableFuture<List<Location>> {
        val request = Request(
            id = generateId(),
            method = "textDocument/definition",
            params = params
        )

        return sendRequest(request).thenApply { response ->
            // Parse locations
            parseLocations(response)
        }
    }

    fun typeDefinition(params: TypeDefinitionParams): CompletableFuture<List<Location>> {
        val request = Request(
            id = generateId(),
            method = "textDocument/typeDefinition",
            params = params
        )

        return sendRequest(request).thenApply { response ->
            parseLocations(response)
        }
    }
}
```

### Find References

```kotlin
class MyReferencesSupport {
    fun references(params: ReferenceParams): CompletableFuture<List<Location>> {
        val request = Request(
            id = generateId(),
            method = "textDocument/references",
            params = params
        )

        return sendRequest(request).thenApply { response ->
            parseLocations(response)
        }
    }
}
```

### Rename Refactoring

```kotlin
class MyRenameSupport {
    fun rename(params: RenameParams): CompletableFuture<WorkspaceEdit> {
        val request = Request(
            id = generateId(),
            method = "textDocument/rename",
            params = params
        )

        return sendRequest(request).thenApply { response ->
            gson.fromJson(response, WorkspaceEdit::class.java)
        }
    }

    fun prepareRename(params: PrepareRenameParams): CompletableFuture<Range> {
        val request = Request(
            id = generateId(),
            method = "textDocument/prepareRename",
            params = params
        )

        return sendRequest(request).thenApply { response ->
            gson.fromJson(response, Range::class.java)
        }
    }
}
```

## LSP Communication

### Message Protocol

```kotlin
object LspMessageProtocol {
    fun writeMessage(output: OutputStream, message: String) {
        val content = message.toByteArray(Charsets.UTF_8)
        val header = "Content-Length: ${content.size}\r\n\r\n"

        output.write(header.toByteArray(Charsets.UTF_8))
        output.write(content)
        output.flush()
    }

    fun readMessage(input: InputStream): String {
        // Read headers
        var line = readLine(input)
        var contentLength = 0

        while (line.isNotEmpty()) {
            if (line.startsWith("Content-Length:")) {
                contentLength = line.substring(15).trim().toInt()
            }
            line = readLine(input)
        }

        // Read content
        val buffer = ByteArray(contentLength)
        input.read(buffer)

        return String(buffer, Charsets.UTF_8)
    }

    private fun readLine(input: InputStream): String {
        val buffer = ByteArrayOutputStream()
        var b: Int

        while (input.read().also { b = it } != -1) {
            if (b == '\r'.code) {
                input.read() // Skip \n
                break
            }
            buffer.write(b)
        }

        return buffer.toString(Charsets.UTF_8.name())
    }
}
```

## Registering LSP Support

### plugin.xml Registration

```xml
<extensions defaultExtensionNs="com.intellij">
    <languageServer
        language="my-language"
        implementationClass="com.example.MyLanguageServerDefinition"/>
</extensions>
```

## Best Practices

### LSP Guidelines

1. **Handle server lifecycle**: Start and stop server properly
2. **Manage connections**: Handle connection errors
3. **Cache results**: Avoid redundant requests
4. **Support cancellation**: Cancel pending requests
5. **Handle large files**: Be efficient with large documents

### Example: Complete LSP Integration

```kotlin
class MyCompleteLspIntegration(val project: Project) {
    private var server: LanguageServer? = null
    private val client = MyLanguageClient()

    fun startServer() {
        val definition = MyLanguageServerDefinition()
        server = definition.createServer(project, client)

        // Initialize
        val params = InitializeParams().apply {
            rootUri = project.basePath?.let { "file://$it" }
            capabilities = ClientCapabilities()
        }

        server?.initialize(params)?.thenAccept { result ->
            // Server initialized
            onServerInitialized(result)
        }
    }

    fun stopServer() {
        server?.shutdown()?.thenAccept {
            server?.exit()
            server = null
        }
    }

    private fun onServerInitialized(result: InitializeResult) {
        // Register capabilities
        val capabilities = result.capabilities

        if (capabilities.completionProvider != null) {
            enableCompletion()
        }

        if (capabilities.definitionProvider != null) {
            enableGoToDefinition()
        }
    }

    private fun enableCompletion() {
        // Enable completion support
    }

    private fun enableGoToDefinition() {
        // Enable go to definition
    }
}
```

## Related Topics

- [Language Support](../05-language-support/language-registration.md) - Language registration
- [Code Completion](../04-editor-features/code-completion.md) - Completion implementation
- [Reference Resolution](../05-language-support/reference-resolution.md) - Reference resolution

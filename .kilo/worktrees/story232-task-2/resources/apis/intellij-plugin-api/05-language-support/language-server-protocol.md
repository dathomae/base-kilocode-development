# Language Server Protocol

The Language Server Protocol (LSP) allows plugins to integrate language servers that provide features like completion, diagnostics, and navigation.

## Overview

### What Is LSP?

LSP is a protocol for communication between editors and language servers. It enables:

- Code completion
- Diagnostics (errors/warnings)
- Go to definition
- Find references
- Hover information
- Code actions

### LSP Integration Approaches

| Approach           | Description                           |
| ------------------ | ------------------------------------- |
| LSP4IJ             | JetBrains library for LSP integration |
| Custom Integration | Direct implementation of LSP client   |

## Using LSP4IJ

### Adding LSP4IJ Dependency

```kotlin
// build.gradle.kts
dependencies {
    implementation("com.github.ballerina-platform:lsp4ij:0.1.0")
}
```

### Defining a Language Server

```kotlin
class MyLanguageServerDefinition : LanguageServerDefinition {
    override val id: String = "my-language-server"

    override val name: String = "My Language Server"

    override fun createLanguageServer(
        project: Project,
        client: LanguageClient
    ): CompletableFuture<LanguageServer> {
        val serverPath = findLanguageServerExecutable()
        val builder = ProcessBuilder(serverPath)

        val process = builder.start()
        val server = Launcher.Builder<LanguageServer>()
            .setLocalService(client)
            .setRemoteInterface(LanguageServer::class.java)
            .setInput(process.inputStream)
            .setOutput(process.outputStream)
            .create()

        server.startListening()
        return CompletableFuture.completedFuture(server.remoteProxy)
    }

    override fun getLanguageIds(): Set<String> {
        return setOf("MyLanguage")
    }

    private fun findLanguageServerExecutable(): String {
        // Find or download the language server
        return "/path/to/language-server"
    }
}
```

### Registering the Language Server

```xml
<extensions defaultExtensionNs="com.intellij">
    <languageServerDefinition
        implementation="com.example.MyLanguageServerDefinition"/>
</extensions>
```

## LSP Features

### Text Document Synchronization

```kotlin
class MyTextDocumentService : TextDocumentService {
    override fun didOpen(params: DidOpenTextDocumentParams) {
        // Handle document open
        val uri = params.textDocument.uri
        val text = params.textDocument.text
    }

    override fun didChange(params: DidChangeTextDocumentParams) {
        // Handle document change
        val uri = params.textDocument.uri
        val changes = params.contentChanges
    }

    override fun didClose(params: DidCloseTextDocumentParams) {
        // Handle document close
    }

    override fun didSave(params: DidSaveTextDocumentParams) {
        // Handle document save
    }
}
```

### Completion Support

```kotlin
class MyCompletionService : TextDocumentService {
    override fun completion(params: CompletionParams): CompletableFuture<Either<List<CompletionItem>, CompletionList>> {
        val uri = params.textDocument.uri
        val position = params.position

        // Get completions from language server
        val completions = getCompletions(uri, position)

        return CompletableFuture.completedFuture(
            Either.forRight(CompletionList(completions))
        )
    }

    override fun resolveCompletionItem(params: CompletionItem): CompletableFuture<CompletionItem> {
        // Resolve additional details for completion item
        return CompletableFuture.completedFuture(params)
    }
}
```

### Diagnostics

```kotlin
class MyDiagnosticsHandler {
    fun handleDiagnostics(params: PublishDiagnosticsParams) {
        val uri = params.uri
        val diagnostics = params.diagnostics

        // Convert to IntelliJ annotations
        val file = findFile(uri) ?: return
        val document = PsiDocumentManager.getInstance(file.project).getDocument(file) ?: return

        for (diagnostic in diagnostics) {
            val range = diagnostic.range
            val startOffset = document.getLineStartOffset(range.start.line) + range.start.character
            val endOffset = document.getLineStartOffset(range.end.line) + range.end.character

            val severity = when (diagnostic.severity) {
                DiagnosticSeverity.Error -> HighlightSeverity.ERROR
                DiagnosticSeverity.Warning -> HighlightSeverity.WARNING
                DiagnosticSeverity.Information -> HighlightSeverity.INFORMATION
                DiagnosticSeverity.Hint -> HighlightSeverity.TEXT_ATTRIBUTES
                else -> HighlightSeverity.INFORMATION
            }

            // Create annotation
            createAnnotation(file, startOffset, endOffset, severity, diagnostic.message)
        }
    }
}
```

### Go to Definition

```kotlin
class MyDefinitionService : TextDocumentService {
    override fun definition(params: DefinitionParams): CompletableFuture<Either<List<Location>, List<LocationLink>>> {
        val uri = params.textDocument.uri
        val position = params.position

        // Get definition location from language server
        val locations = getDefinitionLocations(uri, position)

        return CompletableFuture.completedFuture(
            Either.forLeft(locations)
        )
    }
}
```

### Find References

```kotlin
class MyReferenceService : TextDocumentService {
    override fun references(params: ReferenceParams): CompletableFuture<List<Location>> {
        val uri = params.textDocument.uri
        val position = params.position
        val context = params.context

        // Get references from language server
        val references = getReferenceLocations(uri, position, context)

        return CompletableFuture.completedFuture(references)
    }
}
```

### Hover Information

```kotlin
class MyHoverService : TextDocumentService {
    override fun hover(params: HoverParams): CompletableFuture<Hover?> {
        val uri = params.textDocument.uri
        val position = params.position

        // Get hover info from language server
        val hover = getHoverInfo(uri, position)

        return CompletableFuture.completedFuture(hover)
    }
}
```

## Custom LSP Client

### Implementing a Custom Client

```kotlin
class MyLanguageClient(private val project: Project) : LanguageClient {
    override fun publishDiagnostics(params: PublishDiagnosticsParams) {
        // Handle diagnostics from server
        ApplicationManager.getApplication().invokeLater {
            MyDiagnosticsHandler(project).handleDiagnostics(params)
        }
    }

    override fun showMessage(params: MessageParams) {
        // Show message from server
        ApplicationManager.getApplication().invokeLater {
            when (params.type) {
                MessageType.Error -> Messages.showErrorDialog(project, params.message, "Error")
                MessageType.Warning -> Messages.showWarningDialog(project, params.message, "Warning")
                MessageType.Info -> Messages.showInfoMessage(project, params.message, "Info")
                MessageType.Log -> println(params.message)
            }
        }
    }

    override fun showMessageRequest(params: ShowMessageRequestParams): CompletableFuture<MessageActionItem> {
        // Show message with actions
        return CompletableFuture.supplyAsync {
            val options = params.actions.map { it.title }.toTypedArray()
            val result = Messages.showDialog(
                project,
                params.message,
                "Select Action",
                options,
                0,
                Messages.getQuestionIcon()
            )
            MessageActionItem(if (result >= 0) params.actions[result].title else "")
        }
    }

    override fun logMessage(params: MessageParams) {
        // Log message from server
        println("[LSP] ${params.type}: ${params.message}")
    }
}
```

## Server Lifecycle

### Starting the Server

```kotlin
class MyLanguageServerManager(private val project: Project) {
    private var server: LanguageServer? = null

    fun startServer() {
        val definition = MyLanguageServerDefinition()
        val client = MyLanguageClient(project)

        definition.createLanguageServer(project, client).thenAccept { ls ->
            server = ls

            // Initialize server
            val initParams = InitializeParams().apply {
                rootUri = project.basePath?.let { "file://$it" }
                capabilities = ClientCapabilities()
            }

            ls.initialize(initParams).thenAccept { result ->
                // Server initialized
                println("Server capabilities: ${result.capabilities}")
            }
        }
    }

    fun stopServer() {
        server?.shutdown()
        server?.exit()
        server = null
    }
}
```

### Document Synchronization

```kotlin
class MyDocumentSynchronizer(private val project: Project) {
    private val server: LanguageServer? = null

    fun documentOpened(file: PsiFile) {
        val uri = file.virtualFile?.url ?: return
        val document = PsiDocumentManager.getInstance(project).getDocument(file) ?: return

        server?.textDocumentService?.didOpen(
            DidOpenTextDocumentParams().apply {
                textDocument = TextDocumentItem().apply {
                    this.uri = uri
                    this.languageId = file.language.id
                    this.version = 0
                    this.text = document.text
                }
            }
        )
    }

    fun documentChanged(file: PsiFile, event: DocumentEvent) {
        val uri = file.virtualFile?.url ?: return

        server?.textDocumentService?.didChange(
            DidChangeTextDocumentParams().apply {
                textDocument = VersionedTextDocumentIdentifier().apply {
                    this.uri = uri
                    this.version = event.document.modificationStamp.toInt()
                }
                contentChanges = listOf(
                    TextDocumentContentChangeEvent().apply {
                        this.text = event.document.text
                    }
                )
            }
        )
    }

    fun documentClosed(file: PsiFile) {
        val uri = file.virtualFile?.url ?: return

        server?.textDocumentService?.didClose(
            DidCloseTextDocumentParams().apply {
                textDocument = TextDocumentIdentifier(uri)
            }
        )
    }
}
```

## Configuration

### Server Settings

```kotlin
class MyLanguageServerSettings : PersistentStateComponent<MyLanguageServerSettings.State> {
    private var state = State()

    data class State(
        var serverPath: String = "",
        var enabled: Boolean = true,
        var completionEnabled: Boolean = true,
        var diagnosticsEnabled: Boolean = true
    )

    override fun getState(): State = state
    override fun loadState(state: State) {
        this.state = state
    }

    companion object {
        fun getInstance(): MyLanguageServerSettings = service()
    }
}
```

### Configuration UI

```kotlin
class MyLanguageServerConfigurable : Configurable {
    private val panel = JPanel()
    private val serverPathField = JTextField()
    private val enabledCheckbox = JCheckBox("Enable language server")

    init {
        panel.layout = BorderLayout()
        panel.add(JLabel("Server Path:"), BorderLayout.NORTH)
        panel.add(serverPathField, BorderLayout.CENTER)
        panel.add(enabledCheckbox, BorderLayout.SOUTH)
    }

    override fun createComponent(): JComponent = panel

    override fun isModified(): Boolean {
        val settings = MyLanguageServerSettings.getInstance()
        return serverPathField.text != settings.state.serverPath ||
               enabledCheckbox.isSelected != settings.state.enabled
    }

    override fun apply() {
        val settings = MyLanguageServerSettings.getInstance()
        settings.state.serverPath = serverPathField.text
        settings.state.enabled = enabledCheckbox.isSelected
    }

    override fun getDisplayName(): String = "My Language Server"
}
```

## Related Topics

- [Language Registration](language-registration.md) - Language definition
- [Code Completion](code-completion-lang.md) - Completion features
- [Syntax Highlighting](syntax-highlighting.md) - Visual styling

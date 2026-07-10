# Test Data and Fixtures

Test data and fixtures provide the foundation for writing effective plugin tests.

## Overview

### What Are Test Data?

Test data includes:

- Input files for testing
- Expected output files
- Configuration files
- Mock data

### What Are Fixtures?

Fixtures provide:

- Test project setup
- File creation utilities
- Editor manipulation
- PSI access

## Test Data Organization

### Directory Structure

```
src/test/
├── kotlin/
│   └── com/example/
│       └── MyTest.kt
└── testData/
    ├── parsing/
    │   ├── simple.my
    │   └── simple.txt
    ├── completion/
    │   └── basic.my
    ├── highlighting/
    │   └── errors.my
    └── inspection/
        └── unused.my
```

### Naming Conventions

| Test Type    | Input File       | Expected File     |
| ------------ | ---------------- | ----------------- |
| Parser       | `name.my`        | `name.txt`        |
| Highlighting | `name.my`        | (markers in file) |
| Completion   | `name.my`        | (checked in code) |
| Quick Fix    | `name.before.my` | `name.after.my`   |

## Test Data Files

### Creating Test Files

```kotlin
class MyTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData"

    fun testFromFile() {
        // Load file from testData directory
        myFixture.configureByFile("parsing/simple.my")

        // Test logic
        assertNotNull(myFixture.file)
    }
}
```

### Loading Test Data

```kotlin
protected fun loadFile(path: String): String {
    return File(testDataPath, path).readText()
}

protected fun loadTestData(name: String): String {
    return loadFile("parsing/$name.my")
}
```

## Fixtures

### CodeInsightTestFixture

The main fixture interface:

```kotlin
class MyTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testFixture() {
        // Access project
        val project = myFixture.project

        // Access file
        val file = myFixture.file

        // Access editor
        val editor = myFixture.editor

        // Access PSI
        val psiFile = myFixture.file
    }
}
```

### File Operations

```kotlin
fun testFileOperations() {
    // Create file in project
    val file = myFixture.addFileToProject(
        "src/main.my",
        "content"
    )

    // Find file
    val found = myFixture.findFileInTempDir("src/main.my")

    // Configure from file
    myFixture.configureFromExistingVirtualFile(found!!)
}
```

### Editor Operations

```kotlin
fun testEditorOperations() {
    // Configure with text
    myFixture.configureByText("test.my", "content<caret>")

    // Get caret position
    val offset = myFixture.editor.caretModel.offset

    // Move caret
    myFixture.editor.caretModel.moveToOffset(5)

    // Type text
    myFixture.type("inserted")

    // Check result
    myFixture.checkResult("contentinserted")
}
```

## Temp Directory Fixture

### Using TempDirFixture

```kotlin
fun testTempDir() {
    // Create directory
    val dir = myFixture.tempDirFixture.findOrCreateDir("src/main")

    // Create file
    val file = myFixture.tempDirFixture.createFile("test.my", "content")

    // Find file
    val found = myFixture.tempDirFixture.getFile("test.my")

    assertNotNull(found)
}
```

### File Creation

```kotlin
fun testCreateFiles() {
    // Create multiple files
    myFixture.tempDirFixture.createFile("a.my", "content a")
    myFixture.tempDirFixture.createFile("b.my", "content b")

    // Create in subdirectory
    myFixture.tempDirFixture.findOrCreateDir("sub")
    myFixture.tempDirFixture.createFile("sub/c.my", "content c")
}
```

## Caret Markers

### Using Caret Markers

```kotlin
fun testCaretMarker() {
    myFixture.configureByText("test.my", """
        function test() {
            <caret>
        }
    """.trimIndent())

    // Caret is at marked position
    val offset = myFixture.editor.caretModel.offset
    assertTrue(offset > 0)
}
```

### Selection Markers

```kotlin
fun testSelectionMarker() {
    myFixture.configureByText("test.my", """
        function <selection>test</selection>() {}
    """.trimIndent())

    // Selection is marked
    val selection = myFixture.editor.selectionModel
    assertEquals("test", selection.selectedText)
}
```

### Block Selection

```kotlin
fun testBlockSelection() {
    myFixture.configureByText("test.my", """
        line1
        <block>line2
        line3</block>
    """.trimIndent())

    // Block selection is marked
    val selection = myFixture.editor.selectionModel
    assertTrue(selection.hasBlockSelection())
}
```

## Test Data Patterns

### Input/Output Pattern

```kotlin
fun testTransformation() {
    // Load input
    val input = loadTestData("input.my")

    // Perform transformation
    val output = transform(input)

    // Load expected
    val expected = loadTestData("expected.my")

    // Compare
    assertEquals(expected, output)
}
```

### Before/After Pattern

```kotlin
fun testRefactoring() {
    // Load before state
    myFixture.configureByFile("refactoring/before.my")

    // Perform refactoring
    performRefactoring()

    // Check after state
    myFixture.checkResultByFile("refactoring/after.my")
}
```

## Mock Data

### Creating Mock Services

```kotlin
class MyTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun setUp() {
        super.setUp()

        // Register mock service
        val mockService = MockMyService()
        registerApplicationService(MyService::class.java, mockService)
    }

    protected fun <T : Any> registerApplicationService(
        serviceClass: Class<T>,
        implementation: T
    ) {
        ApplicationManager.getApplication()
            .replaceService(serviceClass, implementation, testRootDisposable)
    }
}

class MockMyService : MyService {
    override fun getData(): String = "mock data"
}
```

### Mock Settings

```kotlin
fun testWithMockSettings() {
    // Create mock settings
    val settings = MySettings().apply {
        myOption = "test value"
    }

    registerApplicationService(MySettings::class.java, settings)

    // Test with settings
    // ...
}
```

## Test Configuration

### Module Configuration

```kotlin
fun testWithModule() {
    // Create module
    val module = myFixture.module

    // Configure module roots
    val model = ModuleRootManager.getInstance(module).modifiableModel

    val contentEntry = model.addContentEntry(
        myFixture.tempDirFixture.getFile("")!!
    )

    contentEntry.addSourceFolder(
        myFixture.tempDirFixture.getFile("src")!!,
        false
    )

    model.commit()
}
```

### Library Configuration

```kotlin
fun testWithLibrary() {
    // Create library
    val libraryTable = LibraryTablesRegistrar.getInstance()
        .getLibraryTable(myFixture.project)

    val library = libraryTable.createLibrary("TestLib")
    val model = library.modifiableModel

    model.addRoot(
        "jar://path/to/lib.jar!/",
        OrderRootType.CLASSES
    )

    model.commit()
}
```

## Test Utilities

### File Comparison

```kotlin
fun testFileComparison() {
    // Configure file
    myFixture.configureByFile("input.my")

    // Perform operation
    performOperation()

    // Compare with expected file
    myFixture.checkResultByFile("expected.my")
}
```

### Text Comparison

```kotlin
fun testTextComparison() {
    myFixture.configureByText("test.my", "input")

    performOperation()

    myFixture.checkResult("expected output")
}
```

### Ignoring Whitespace

```kotlin
fun testIgnoringWhitespace() {
    myFixture.configureByText("test.my", "input")

    performOperation()

    // Compare ignoring whitespace differences
    myFixture.checkResult("""
        expected
        output
    """.trimIndent(), /* ignoreTrailingWhitespace = */ true)
}
```

## Best Practices

### Test Data Guidelines

1. **Organize by test type**: Separate directories for different tests
2. **Use descriptive names**: Name files after test cases
3. **Keep files small**: Minimal test data needed
4. **Version control**: Commit test data with tests
5. **Document purpose**: Comment complex test data

### Fixture Guidelines

1. **Clean up in tearDown**: Dispose resources
2. **Use testRootDisposable**: For automatic cleanup
3. **Don't share state**: Tests should be independent
4. **Mock external dependencies**: Isolate from external systems

### Example: Complete Test Setup

```kotlin
class MyCompleteTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData"

    private lateinit var mockService: MockMyService

    override fun setUp() {
        super.setUp()

        // Setup mock service
        mockService = MockMyService()
        registerApplicationService(MyService::class.java, mockService)

        // Create project structure
        myFixture.tempDirFixture.findOrCreateDir("src/main")
        myFixture.addFileToProject("src/main/config.json", "{}")
    }

    override fun tearDown() {
        mockService.reset()
        super.tearDown()
    }

    fun testWithTestData() {
        // Load test file
        myFixture.configureByFile("testData/input.my")

        // Perform test
        val result = myFixture.file.text

        // Verify
        assertTrue(result.isNotEmpty())
    }

    fun testWithInlineTestData() {
        myFixture.configureByText("test.my", """
            function test() {
                <caret>
            }
        """.trimIndent())

        // Test with caret position
        val offset = myFixture.editor.caretModel.offset
        assertTrue(offset > 0)
    }

    protected fun <T : Any> registerApplicationService(
        serviceClass: Class<T>,
        implementation: T
    ) {
        ApplicationManager.getApplication()
            .replaceService(serviceClass, implementation, testRootDisposable)
    }
}

class MockMyService : MyService {
    private var callCount = 0

    override fun getData(): String {
        callCount++
        return "mock data"
    }

    fun reset() {
        callCount = 0
    }

    fun getCallCount(): Int = callCount
}
```

## Related Topics

- [Testing Overview](testing-overview.md) - Testing concepts
- [Light Tests](light-tests.md) - Light test framework
- [Heavy Tests](heavy-tests.md) - Heavy test framework

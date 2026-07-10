# Heavy Tests

Heavy tests provide full IDE integration testing with complete project setup.

## Overview

### What Are Heavy Tests?

Heavy tests provide:

- Full IDE environment
- Complete project structure
- Real module configuration
- Multi-file analysis support

### When to Use Heavy Tests

| Use Case               | Suitable             |
| ---------------------- | -------------------- |
| Multi-file analysis    | Yes                  |
| Project-level features | Yes                  |
| Module dependencies    | Yes                  |
| Build integration      | Yes                  |
| Simple PSI tests       | No (use light tests) |

## Basic Heavy Test

### Test Class

```kotlin
class MyHeavyTest : BasePlatformTestCase() {
    override fun getTestDataPath(): String = "src/test/testData"

    fun testProject() {
        // Create project structure
        myFixture.addFileToProject("src/main.my", "content")

        // Test with full project
        val project = myFixture.project
        assertNotNull(project)
    }
}
```

### Project Setup

```kotlin
class MyProjectTest : BasePlatformTestCase() {
    override fun setUp() {
        super.setUp()

        // Create project structure
        createProjectStructure()
    }

    private fun createProjectStructure() {
        // Create source directories
        myFixture.tempDirFixture.findOrCreateDir("src/main")
        myFixture.tempDirFixture.findOrCreateDir("src/test")

        // Create files
        myFixture.addFileToProject(
            "src/main/main.my",
            "function main() {}"
        )
    }
}
```

## Module Configuration

### Creating Modules

```kotlin
class MyModuleTest : BasePlatformTestCase() {
    fun testWithModule() {
        // Create module
        val module = createModule("myModule")

        // Configure module
        ModuleRootManager.getInstance(module).modifiableModel.apply {
            addContentEntry(myFixture.tempDirFixture.getFile("src")!!)
            commit()
        }

        assertNotNull(module)
    }

    private fun createModule(name: String): Module {
        val moduleManager = ModuleManager.getInstance(myFixture.project)
        return moduleManager.newModule(
            myFixture.tempDirFixture.getFile("$name.iml")!!,
            "MY_MODULE_TYPE"
        )
    }
}
```

### Module Dependencies

```kotlin
fun testModuleDependencies() {
    // Create modules
    val mainModule = createModule("main")
    val libModule = createModule("lib")

    // Add dependency
    ModuleRootModificationUtil.addDependency(mainModule, libModule)

    // Verify dependency
    val dependencies = ModuleRootManager.getInstance(mainModule).dependencies
    assertTrue(dependencies.contains(libModule))
}
```

## Library Configuration

### Adding Libraries

```kotlin
fun testWithLibrary() {
    // Create library
    val library = createLibrary("myLibrary", listOf("path/to/jar.jar"))

    // Add to module
    val module = myFixture.module
    ModuleRootModificationUtil.addDependency(module, library)

    // Verify library
    val libraries = ModuleRootManager.getInstance(module).orderEntries
        .filterIsInstance<LibraryOrderEntry>()
    assertTrue(libraries.any { it.libraryName == "myLibrary" })
}

private fun createLibrary(name: String, classes: List<String>): Library {
    val libraryTable = LibraryTablesRegistrar.getInstance()
        .getLibraryTable(myFixture.project)

    val library = libraryTable.createLibrary(name)
    val model = library.modifiableModel

    for (path in classes) {
        model.addRoot(path, OrderRootType.CLASSES)
    }

    model.commit()
    return library
}
```

## Multi-File Testing

### Cross-File Analysis

```kotlin
class MyMultiFileTest : BasePlatformTestCase() {
    fun testCrossFileReference() {
        // Create files
        myFixture.addFileToProject("src/a.my", """
            export function myFunction() {}
        """.trimIndent())

        myFixture.configureByText("src/b.my", """
            import { myFunction } from './a';
            myFunction();
        """.trimIndent())

        // Test reference resolution
        val reference = myFixture.getReferenceAtCaretPosition()
        val resolved = reference?.resolve()

        assertNotNull(resolved)
    }
}
```

### Project-Wide Search

```kotlin
fun testProjectSearch() {
    // Create multiple files
    myFixture.addFileToProject("src/a.my", "function funcA() {}")
    myFixture.addFileToProject("src/b.my", "function funcB() {}")
    myFixture.addFileToProject("src/c.my", "function funcC() {}")

    // Search across project
    val scope = GlobalSearchScope.projectScope(myFixture.project)
    val results = mutableListOf<PsiElement>()

    FileBasedIndex.getInstance().processFilesContainingAllKeys(
        MyIndex.ID,
        setOf("function"),
        scope,
        null,
        { file ->
            results.add(file)
            true
        }
    )

    assertEquals(3, results.size)
}
```

## Index Testing

### Testing with Index

```kotlin
class MyIndexTest : BasePlatformTestCase() {
    fun testIndex() {
        // Create file
        myFixture.addFileToProject("test.my", "function myFunc() {}")

        // Wait for indexing
        FileBasedIndex.getInstance().waitUntilIndicesAreInitialized()

        // Query index
        val keys = FileBasedIndex.getInstance().getAllKeys(
            MyIndex.ID,
            myFixture.project
        )

        assertTrue(keys.contains("myFunc"))
    }
}
```

### Rebuilding Index

```kotlin
fun testRebuildIndex() {
    // Create file
    myFixture.addFileToProject("test.my", "content")

    // Rebuild index
    FileBasedIndex.getInstance().requestRebuild(MyIndex.ID)
    FileBasedIndex.getInstance().waitUntilIndicesAreInitialized()

    // Test index state
    // ...
}
```

## Run Configuration Testing

```kotlin
class MyRunConfigurationTest : BasePlatformTestCase() {
    fun testRunConfiguration() {
        // Create run configuration
        val factory = MyConfigurationFactory()
        val configuration = factory.createTemplateConfiguration(myFixture.project)

        // Configure
        configuration.name = "Test Run"
        configuration.mainClass = "com.example.Main"

        // Verify
        assertEquals("Test Run", configuration.name)
        assertEquals("com.example.Main", configuration.mainClass)
    }
}
```

## Facet Testing

```kotlin
class MyFacetTest : BasePlatformTestCase() {
    fun testFacet() {
        // Create module
        val module = myFixture.module

        // Add facet
        val facetType = MyFacetType.INSTANCE
        val facet = FacetManager.getInstance(module).addFacet(
            facetType,
            "MyFacet",
            null
        )

        // Configure facet
        facet.configuration.mySetting = "value"

        // Verify
        val loadedFacet = FacetManager.getInstance(module).getFacetByType(MyFacetType.ID)
        assertNotNull(loadedFacet)
        assertEquals("value", loadedFacet?.configuration?.mySetting)
    }
}
```

## SDK Configuration

```kotlin
class MySdkTest : BasePlatformTestCase() {
    fun testSdk() {
        // Create SDK
        val sdk = createSdk("MySDK", "/path/to/sdk")

        // Set project SDK
        ProjectRootManager.getInstance(myFixture.project).projectSdk = sdk

        // Verify
        val projectSdk = ProjectRootManager.getInstance(myFixture.project).projectSdk
        assertEquals("MySDK", projectSdk?.name)
    }

    private fun createSdk(name: String, homePath: String): Sdk {
        val sdkType = MySdkType.INSTANCE
        val sdk = ProjectJdkTable.getInstance().createSdk(name, sdkType)

        val sdkModificator = sdk.sdkModificator
        sdkModificator.homePath = homePath
        sdkModificator.commitChanges()

        return sdk
    }
}
```

## Performance Considerations

### Heavy Test Performance

Heavy tests are slower than light tests. Optimize by:

1. **Minimize project setup**: Only create necessary files
2. **Reuse project state**: Share setup between related tests
3. **Use appropriate scope**: Don't use heavy tests for simple cases

```kotlin
class OptimizedHeavyTest : BasePlatformTestCase() {
    companion object {
        // Shared setup for all tests
        @JvmStatic
        fun setupProject(fixture: CodeInsightTestFixture) {
            fixture.addFileToProject("common.my", "shared content")
        }
    }

    override fun setUp() {
        super.setUp()
        setupProject(myFixture)
    }
}
```

## Best Practices

### Heavy Test Guidelines

1. **Use when necessary**: Only for features requiring full IDE
2. **Minimize setup**: Create only required files and modules
3. **Clean up resources**: Dispose in tearDown
4. **Test isolation**: Ensure tests don't affect each other
5. **Document requirements**: Note any special setup needs

### Example: Complete Heavy Test

```kotlin
class MyCompleteHeavyTest : BasePlatformTestCase() {
    override fun getTestDataPath(): String = "src/test/testData"

    private lateinit var myModule: Module

    override fun setUp() {
        super.setUp()

        // Create module
        myModule = createModule("testModule")

        // Setup project structure
        myFixture.tempDirFixture.findOrCreateDir("src/main")
        myFixture.addFileToProject("src/main/main.my", "function main() {}")
    }

    fun testModuleStructure() {
        // Verify module
        assertNotNull(myModule)

        // Verify files
        val mainFile = myFixture.findFileInTempDir("src/main/main.my")
        assertNotNull(mainFile)
    }

    fun testCrossFileAnalysis() {
        // Create additional file
        myFixture.addFileToProject("src/main/lib.my", "function lib() {}")

        // Test analysis
        val files = FilenameIndex.getVirtualFilesByName(
            myFixture.project,
            "lib.my",
            GlobalSearchScope.moduleScope(myModule)
        )

        assertEquals(1, files.size)
    }

    private fun createModule(name: String): Module {
        val moduleManager = ModuleManager.getInstance(myFixture.project)
        return moduleManager.newModule(
            myFixture.tempDirFixture.getFile("$name.iml")!!,
            "EMPTY_MODULE"
        )
    }
}
```

## Related Topics

- [Testing Overview](testing-overview.md) - Testing concepts
- [Light Tests](light-tests.md) - Fast unit tests
- [Test Data and Fixtures](test-data-and-fixtures.md) - Test data management

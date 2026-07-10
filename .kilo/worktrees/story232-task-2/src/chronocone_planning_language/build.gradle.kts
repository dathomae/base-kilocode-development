/**
 * Gradle build configuration for Chronocone Planning Language (CPL) Parser
 *
 * This build file configures ANTLR4 for Go code generation and sets up
 * the Go module structure for building the parser.
 *
 * There's some conflict between the build/test ecosystems that are intend
 * to be used for antlr and for go, which is the parser target language.
 *
 * Antlr expects to live in the traditional Java ecosystem, for which the
 * usual build support is provided by Gradle. Go, however, expects to live
 * in the Go ecosystem, for which the usual build support is provided by
 * the Go toolchain.
 *
 * To address that this has been set up as a gradle environment with 
 * custom tasks that shell out to the Go toolchain. There's an implicit
 * assumption that we're running under a Unix-like environment (as opposed
 * to Windows). There's also a go support task that depends on the use
 * of git, so if we wanted to build outside of a git repository we'd have
 * to fiddle with this a bit.
 */

plugins {
    // ANTLR plugin for generating parser from grammar files
    antlr
}

repositories {
    // Maven Central for ANTLR dependencies
    mavenCentral()
}

// ANTLR4 version configuration
val antlrVersion = "4.13.1"

dependencies {
    // ANTLR tool for code generation
    antlr("org.antlr:antlr4:$antlrVersion")
}

// Configure generateGrammarSource task for Go output
tasks.generateGrammarSource {
    // Only process the main grammar files, not all .g4 files
    source = fileTree("src/main/antlr") {
        include("ChronoconePlanningLanguage.g4")
        include("ChronoconePlanningLanguageLexer.g4")
    }

    // Set the output directory for generated Go code directly in source tree
    outputDirectory = file("src/main/go/parser/antlr")

    // Configure additional arguments for Go generation
    arguments = listOf(
        "-Dlanguage=Go",
        "-visitor",
        "-listener",
        "-package", "antlr"
    )
}

// Configure source sets for ANTLR
sourceSets {
    main {
        antlr {
            setSrcDirs(setOf(file("src/main/antlr")))
        }
    }
}

// Task to clean generated files
tasks.register<Delete>("cleanGenerated") {
    group = "build"
    description = "Cleans generated Go code in src/main/go/parser/antlr/"

    delete("src/main/go/parser/antlr")
}

// Make clean task also clean generated files
// Note: Even though generated files are committed to git for Go module distribution,
// they are still build artifacts that should be cleaned to force regeneration
tasks.clean {
    dependsOn("cleanGenerated")
}

// Custom task to run Go tests
tasks.register<Exec>("goTest") {
    group = "verification"
    description = "Runs Go tests in the src/main/go directory"

    // Ensure ANTLR files are generated and the compiler binary is built before running tests
    dependsOn("generateGrammarSource", "compileCplCompiler")

    // Check if Go is installed
    onlyIf {
        try {
            exec {
                commandLine("go", "version")
                isIgnoreExitValue = true
            }.exitValue == 0
        } catch (e: Exception) {
            false
        }
    }

    // Run tests from the project root (where go.mod is located)
    workingDir = project.projectDir

    commandLine("go", "test", "-v", "./src/main/go/...")

    doFirst {
        // Ensure generated/cmd exists before running tests; some integration tests
        // create temporary directories under this path and will fail if it is missing.
        file("generated/cmd").mkdirs()
        println("Running Go tests...")
    }

    doLast {
        println("Go tests completed!")
    }
}

// Task to run go vet for static analysis
tasks.register<Exec>("goVet") {
    group = "verification"
    description = "Runs go vet for static analysis"

    dependsOn("generateGrammarSource")

    // Check if Go is installed
    onlyIf {
        try {
            exec {
                commandLine("go", "version")
                isIgnoreExitValue = true
            }.exitValue == 0
        } catch (e: Exception) {
            false
        }
    }

    workingDir = project.projectDir
    // Run go vet on all Go packages. go vet analyzes transitive dependencies,
    // so we cannot exclude parser/antlr by omitting it from the package list.
    // Instead, we run vet on everything and filter out errors from the generated
    // ANTLR code in src/main/go/parser/antlr.
    commandLine("sh", "-c", "OUTPUT=\$(go vet ./src/main/go/... 2>&1); FILTERED=\$(echo \"\$OUTPUT\" | grep -v 'parser/antlr/'); [ -n \"\$FILTERED\" ] && { echo \"\$FILTERED\"; exit 1; } || exit 0")

    doFirst {
        println("Running go vet for static analysis (excluding generated ANTLR code)...")
    }

    doLast {
        println("go vet completed successfully!")
    }
}

// Task to format Go code files
tasks.register<Exec>("goFmt") {
    group = "formatting"
    description = "Formats Go code files using gofmt"

    // Check if Go is installed
    onlyIf {
        try {
            exec {
                commandLine("go", "version")
                isIgnoreExitValue = true
            }.exitValue == 0
        } catch (e: Exception) {
            false
        }
    }

    workingDir = project.projectDir

    // gofmt -w writes formatted output back to the files
    // Use find to exclude generated ANTLR code in src/main/go/parser/antlr
    commandLine("sh", "-c", "find src/main/go -name '*.go' -not -path 'src/main/go/parser/antlr/*' -exec gofmt -w {} + && echo 'Go code formatting completed'")

    doFirst {
        println("Formatting Go code files (excluding generated ANTLR code)...")
    }

    doLast {
        println("Go code formatting completed!")
    }
}

// Task to check if Git is installed
tasks.register("checkGit") {
    group = "verification"
    description = "Checks if Git is installed and available"

    doLast {
        try {
            exec {
                commandLine("git", "--version")
                isIgnoreExitValue = false
            }
            println("Git is available.")
        } catch (e: Exception) {
            throw GradleException("Git is not installed or not available in PATH. Please install Git to run this task.")
        }
    }
}

// Task to verify go.mod is tidy
tasks.register<Exec>("goModTidy") {
    group = "verification"
    description = "Verifies go.mod and go.sum are in sync with imports"

    dependsOn("checkGit")

    // Check if Go is installed
    onlyIf {
        try {
            exec {
                commandLine("go", "version")
                isIgnoreExitValue = true
            }.exitValue == 0
        } catch (e: Exception) {
            false
        }
    }

    workingDir = project.projectDir

    // Run go mod tidy and check if go.mod or go.sum changed
    commandLine("sh", "-c", "go mod tidy && git diff --exit-code go.mod go.sum || (echo 'Error: go.mod or go.sum is out of sync. Run go mod tidy to fix.' && exit 1)")

    doFirst {
        println("Verifying go.mod and go.sum are tidy...")
    }

    doLast {
        println("go.mod verification passed!")
    }
}

// Task to compile the CPL compiler executable
tasks.register<Exec>("compileCplCompiler") {
    group = "build"
    description = "Compiles the CPL compiler Go binary to build/cpl_compiler"

    dependsOn("generateGrammarSource")

    // Check if Go is installed
    onlyIf {
        try {
            exec {
                commandLine("go", "version")
                isIgnoreExitValue = true
            }.exitValue == 0
        } catch (e: Exception) {
            false
        }
    }

    workingDir = project.projectDir

    // Create build/bin directory if it doesn't exist and compile the binary
    commandLine("sh", "-c", "mkdir -p build/bin && go build -o build/bin/cpl_compiler ./src/main/go/cli/compiler/")

    doFirst {
        println("Compiling CPL compiler to build/bin/cpl_compiler...")
    }

    doLast {
        println("CPL compiler compiled successfully to build/bin/cpl_compiler")
    }
}

// Task to compile the parquetgen executable
tasks.register<Exec>("compileParquetGen") {
    group = "build"
    description = "Compiles the parquetgen Go binary to build/bin/parquetgen"

    dependsOn("generateGrammarSource")

    // Check if Go is installed
    onlyIf {
        try {
            exec {
                commandLine("go", "version")
                isIgnoreExitValue = true
            }.exitValue == 0
        } catch (e: Exception) {
            false
        }
    }

    workingDir = project.projectDir

    // Create build/bin directory if it doesn't exist and compile the binary
    commandLine("sh", "-c", "mkdir -p build/bin && go build -o build/bin/parquetgen ./src/main/go/cli/parquetgen/")

    doFirst {
        println("Compiling parquetgen to build/bin/parquetgen...")
    }

    doLast {
        println("parquetgen compiled successfully to build/bin/parquetgen")
    }
}

// Task to clean the build directory
tasks.register<Delete>("cleanBuild") {
    group = "build"
    description = "Cleans the build directory"

    delete("build")
}

// Make clean task also clean build directory
tasks.clean {
    dependsOn("cleanBuild")
}

// Configure build task to include grammar generation and compiler compilation
tasks.build {
    dependsOn("generateGrammarSource", "compileCplCompiler", "compileParquetGen")
}

// Make the standard test task also run Go tests and verification tasks
tasks.test {
    // Ensure ANTLR files are generated and the compiler binary is built before running tests
    dependsOn("generateGrammarSource", "compileCplCompiler", "goTest", "goVet", "goModTidy")
}


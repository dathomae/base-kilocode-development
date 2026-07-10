# Go Testing Best Practices

This document summarizes guidelines for code organization and test construction in Go, based on the [JetBrains comprehensive guide to testing in Go](https://blog.jetbrains.com/go/2022/11/22/comprehensive-guide-to-testing-in-go/). This summary focuses on language-level patterns and practices, omitting IDE-specific features.

## Table of Contents

- [Code Organization](#code-organization)
- [Test Structure Basics](#test-structure-basics)
- [Unit Tests](#unit-tests)
- [Table-Driven Tests](#table-driven-tests)
- [Parallel Tests](#parallel-tests)
- [Subtests](#subtests)
- [Test Utilities and Helpers](#test-utilities-and-helpers)
- [Coverage Tests](#coverage-tests)
- [Benchmark Tests](#benchmark-tests)
- [Fuzz Tests](#fuzz-tests)
- [Using Testify](#using-testify)

---

## Code Organization

### Test File Naming

Go uses a specific naming convention to identify test files:

- Test files must end with `*_test.go`
- The `testing` package is designed to be used alongside the `go test` command
- Test files should be placed in the same package as the code they test

Example directory structure:

```
mypackage/
├── mycode.go       # Source code
├── mycode_test.go  # Test file
└── utils.go        # More source code
```

### Package Organization

Tests typically belong to the same package as the code under test. This allows access to unexported (private) functions and types when needed.

If you need to test only the public API, use the `_test` suffix on the package name:

```go
// In mypackage.go
package mypackage

// In mypackage_test.go
package mypackage_test  // Tests only the public API
```

---

## Test Structure Basics

### Anatomy of a Test Function

A basic test function follows this structure:

```go
func TestXxx(t *testing.T) {
    // Test logic here
}
```

Requirements:

- Function name must start with `Test`
- Function must accept `*testing.T` as its only parameter
- Test files must be named `*_test.go`

### Example: Simple Unit Test

```go
package mypackage

import "testing"

// Function to test
func Fooer(input int) string {
    if input%3 == 0 {
        return "Foo"
    }
    return strconv.Itoa(input)
}

// Basic test
func TestFooer(t *testing.T) {
    result := Fooer(3)
    if result != "Foo" {
        t.Errorf("Result was incorrect, got: %s, want: %s.", result, "Foo")
    }
}
```

---

## Unit Tests

Unit tests verify that individual functions or methods work correctly in isolation.

### Multiple Test Functions

You can create multiple test functions for the same code to test different scenarios:

```go
func TestFooer_Zero(t *testing.T) {
    result := Fooer(0)
    if result != "Foo" {
        t.Errorf("Fooer(0) = %s; want Foo", result)
    }
}

func TestFooer_Positive(t *testing.T) {
    result := Fooer(3)
    if result != "Foo" {
        t.Errorf("Fooer(3) = %s; want Foo", result)
    }
}

func TestFooer_Negative(t *testing.T) {
    result := Fooer(-3)
    if result != "Foo" {
        t.Errorf("Fooer(-3) = %s; want Foo", result)
    }
}
```

### Test Isolation

Each test function runs independently. Tests should not depend on the execution order or share mutable state between them.

---

## Table-Driven Tests

Table-driven tests are a Go idiom for testing multiple scenarios with a single test function. This pattern reduces code duplication and makes it easier to add new test cases.

### Basic Pattern

```go
func TestFooer_TableDriven(t *testing.T) {
    // Define test cases
    testCases := []struct {
        name     string
        input    int
        expected string
    }{
        {
            name:     "zero is divisible by 3",
            input:    0,
            expected: "Foo",
        },
        {
            name:     "three is divisible by 3",
            input:    3,
            expected: "Foo",
        },
        {
            name:     "one is not divisible by 3",
            input:    1,
            expected: "1",
        },
        {
            name:     "negative three is divisible by 3",
            input:    -3,
            expected: "Foo",
        },
    }

    // Execute each test case
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := Fooer(tc.input)
            if result != tc.expected {
                t.Errorf("Fooer(%d) = %s; want %s",
                    tc.input, result, tc.expected)
            }
        })
    }
}
```

### Benefits of Table-Driven Tests

1. **Reduced duplication**: Common test logic is written once
2. **Easy to extend**: Adding a new test case requires only adding a struct entry
3. **Clear intent**: Each test case has a descriptive name
4. **Consistent structure**: All cases execute the same verification logic

### When to Use

- When testing the same function with different inputs and expected outputs
- When you have multiple related test scenarios
- When test cases can be expressed as data (input → expected output)

---

## Parallel Tests

Go supports running tests in parallel, which can significantly reduce test execution time.

### Enabling Parallel Execution

Use `t.Parallel()` at the start of a test function:

```go
func TestFooer_Parallel(t *testing.T) {
    t.Parallel()  // Marks this test as parallelizable

    result := Fooer(3)
    if result != "Foo" {
        t.Errorf("Result was incorrect")
    }
}
```

### Parallel Table-Driven Tests

When combining parallel tests with table-driven tests, the `tc` variable must be captured in the closure:

```go
func TestFooer_ParallelTableDriven(t *testing.T) {
    testCases := []struct {
        name     string
        input    int
        expected string
    }{
        {"zero", 0, "Foo"},
        {"three", 3, "Foo"},
        {"one", 1, "1"},
    }

    for _, tc := range testCases {
        tc := tc  // Capture range variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()  // Each subtest runs in parallel
            result := Fooer(tc.input)
            if result != tc.expected {
                t.Errorf("Fooer(%d) = %s; want %s",
                    tc.input, result, tc.expected)
            }
        })
    }
}
```

**Important**: The `tc := tc` line is essential. Without it, all goroutines would share the same loop variable, causing race conditions.

### Best Practices for Parallel Tests

1. Tests marked with `t.Parallel()` will run simultaneously with other parallel tests
2. Parallel tests execute only after non-parallel tests complete
3. Use `t.Setenv()` instead of `os.Setenv()` in parallel tests (available since Go 1.17)
4. Be careful with shared resources—parallel tests should be independent

---

## Subtests

Subtests allow you to group related tests and share setup/teardown code.

### Creating Subtests

Use `t.Run()` to create subtests:

```go
func TestFooer_WithSubtests(t *testing.T) {
    // Setup code (runs once)
    commonSetup()

    t.Run("divisible by 3", func(t *testing.T) {
        result := Fooer(3)
        if result != "Foo" {
            t.Errorf("Expected Foo, got %s", result)
        }
    })

    t.Run("not divisible by 3", func(t *testing.T) {
        result := Fooer(1)
        if result != "1" {
            t.Errorf("Expected 1, got %s", result)
        }
    })

    // Cleanup (runs after all subtests complete)
}
```

### Subtest with Table-Driven Pattern

The table-driven pattern uses subtests internally via `t.Run()`:

```go
for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
        // This subtest can be run independently
        result := Fooer(tc.input)
        if result != tc.expected {
            t.Errorf("Expected %s, got %s", tc.expected, result)
        }
    })
}
```

### Benefits of Subtests

- Run individual subtests using `go test -run TestFooer/divisible`
- Shared setup/teardown for related tests
- Hierarchical test organization
- Better error reporting with context

---

## Test Utilities and Helpers

### t.Cleanup() for Resource Management

Use `t.Cleanup()` to register cleanup functions that run after each test (including subtests):

```go
func Test_With_Cleanup(t *testing.T) {
    // Create a resource
    resource := setupExpensiveResource()

    // Register cleanup
    t.Cleanup(func() {
        resource.Close()
    })

    // Test code here...
    result := doSomething(resource)
    if result != expected {
        t.Errorf("Unexpected result")
    }

    // Cleanup runs automatically after this function returns
}
```

Benefits of `t.Cleanup()`:

- Cleanup logic is declared near resource creation
- Runs even if the test fails or panics
- Executes in reverse order of registration (LIFO)

### t.Helper() for Better Error Messages

Mark helper functions with `t.Helper()` to improve failure reporting:

```go
func assertEqual(t *testing.T, got, want string) {
    t.Helper()  // Marks this as a helper function
    if got != want {
        t.Errorf("got %q, want %q", got, want)
    }
}

func TestFooer_WithHelper(t *testing.T) {
    result := Fooer(3)
    assertEqual(t, result, "Foo")
}
```

Without `t.Helper()`, failure reports would point to the line inside `assertEqual`. With `t.Helper()`, they point to the call site in the test function.

### t.TempDir() for Temporary Files

Create temporary directories that are automatically cleaned up:

```go
func Test_With_TempDir(t *testing.T) {
    tmpDir := t.TempDir()

    // Use tmpDir for file operations
    filePath := filepath.Join(tmpDir, "test.txt")
    err := os.WriteFile(filePath, []byte("data"), 0644)
    if err != nil {
        t.Fatalf("Failed to write file: %v", err)
    }

    // Directory and contents are deleted automatically after test
}
```

Benefits:

- No manual cleanup required
- Each test gets a unique directory
- Automatically deleted even if test fails

---

## Coverage Tests

Test coverage measures how much of your code is exercised by tests.

### Running Coverage Tests

```bash
# Basic coverage report
go test -cover

# Coverage for all packages including those without test files
go test ./... -coverpkg=./...

# Generate coverage profile file
go test -coverprofile=coverage.out

# View coverage in HTML format
go tool cover -html=coverage.out
```

### Coverage Modes

The `-covermode` flag controls how coverage is calculated:

- `set` (default): Binary coverage—was the statement executed?
- `count`: Count how many times each statement was executed
- `atomic`: Like `count`, but safe for parallel tests

Example:

```bash
go test -covermode=count
go test -covermode=atomic  # For parallel tests
```

### Interpreting Coverage Reports

- Coverage percentage indicates the proportion of statements executed
- Focus on covering critical paths and edge cases, not just maximizing percentage
- Uncovered code may indicate missing tests or unreachable code

---

## Benchmark Tests

Benchmark tests measure the performance of your code.

### Creating Benchmark Tests

Requirements:

- Function name starts with `Benchmark`
- Function accepts `*testing.B` as its only parameter
- Code under test runs inside a loop using `b.N`

```go
func BenchmarkFooer(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Fooer(i)
    }
}
```

### Running Benchmarks

```bash
# Run benchmarks
go test -bench=.

# Run specific benchmark
go test -bench=BenchmarkFooer

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem
```

### Benchmark Output

```
BenchmarkFooer-8    59969790    19.0 ns/op    0 B/op    0 allocs/op
```

Explanation:

- `BenchmarkFooer-8`: Test name (8 = GOMAXPROCS)
- `59969790`: Number of iterations executed
- `19.0 ns/op`: Nanoseconds per operation
- `0 B/op`: Bytes allocated per operation
- `0 allocs/op`: Heap allocations per operation

### Benchmark Best Practices

1. The benchmark loop should only contain the code being measured
2. Avoid allocations in benchmark setup code (before the loop)
3. Run benchmarks multiple times to ensure stable results
4. Compare benchmarks with `benchstat` for statistically significant differences

---

## Fuzz Tests

Fuzz testing uses random input to discover edge cases and bugs.

### Creating Fuzz Tests

Requirements:

- Function name starts with `Fuzz`
- Function accepts `*testing.F` as its only parameter
- Define seed corpus with `f.Add()`
- Define a fuzz target with `f.Fuzz()`

```go
func FuzzFooer(f *testing.F) {
    // Seed corpus: known inputs to start with
    f.Add(3)
    f.Add(0)
    f.Add(-3)
    f.Add(1)

    // Fuzz target: function to test with generated input
    f.Fuzz(func(t *testing.T, a int) {
        result := Fooer(a)
        // Verify the result is valid (not that it matches a specific value)
        if result == "" {
            t.Errorf("Fooer(%d) returned empty string", a)
        }
    })
}
```

### Running Fuzz Tests

```bash
# Run fuzzing for a limited time
go test -fuzz=FuzzFooer -fuzztime=30s

# Run fuzzing until a failure is found
go test -fuzz=FuzzFooer
```

### Fuzz Test Strategies

1. **Input validation**: Ensure the function doesn't panic on unexpected input
2. **Invariants**: Verify properties that should always hold (e.g., non-empty result)
3. **Round-trip testing**: For encoding/decoding pairs, verify encode(decode(x)) == x

### When to Use Fuzzing

- Functions with many branches or complex logic
- Input validation functions
- Parsers and format converters
- Functions that accept user input

---

## Using Testify

Testify is a popular third-party testing toolkit that provides assertions and mocks.

### Installation

```bash
go get github.com/stretchr/testify
```

### Basic Assertions

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestFooer_WithTestify(t *testing.T) {
    // Assert equality
    assert.Equal(t, "Foo", Fooer(0), "0 is divisible by 3")

    // Assert inequality
    assert.NotEqual(t, "Foo", Fooer(1), "1 is not divisible by 3")

    // Assert nil
    var err error
    assert.Nil(t, err)

    // Assert not nil
    err = errors.New("some error")
    assert.NotNil(t, err)
}
```

### Require vs Assert

Testify provides two packages:

- **`assert`**: Continues execution after failure, collects all failures
- **`require`**: Stops execution immediately on failure (fail-fast)

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func Test_WithRequire(t *testing.T) {
    obj := createObject()

    // If this fails, test stops here
    require.NotNil(t, obj, "Object should be created")

    // Only runs if require passes
    assert.Equal(t, "expected", obj.Name)
    assert.Equal(t, 42, obj.Value)
}
```

### Common Assertions

```go
// Equality
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.True(t, condition)
assert.False(t, condition)

// Nil checks
assert.Nil(t, object)
assert.NotNil(t, object)

// Error checks
assert.NoError(t, err)
assert.Error(t, err)
assert.ErrorIs(t, err, targetErr)
assert.ErrorContains(t, err, "substring")

// Collection assertions
assert.Contains(t, slice, element)
assert.Len(t, collection, expectedLen)
assert.Empty(t, collection)
assert.NotEmpty(t, collection)

// String assertions
assert.Contains(t, str, substring)
assert.HasPrefix(t, str, prefix)
assert.HasSuffix(t, str, suffix)
```

### Benefits of Testify

1. **Readability**: Assertions read like natural language
2. **Better error messages**: Shows expected vs actual values clearly
3. **Rich comparisons**: Deep equality for complex types
4. **Mock support**: Mocking framework for interface testing

### When to Use Testify

- When you want more expressive assertions
- When working with complex data structures
- When you need detailed failure messages
- When you require mocking capabilities

---

## Summary of Best Practices

### Code Organization

1. Place test files in the same directory as source code
2. Name test files with the `_test.go` suffix
3. Use the same package for white-box tests, `_test` suffix for black-box tests
4. Keep test code close to the code it tests

### Test Construction

1. **Use table-driven tests** for multiple similar test cases
2. **Write parallel tests** when tests are independent to speed up execution
3. **Use subtests** for organization and shared setup
4. **Capture range variables** in parallel table-driven tests
5. **Use `t.Helper()`** in custom assertion helpers
6. **Use `t.Cleanup()`** for reliable resource cleanup
7. **Use `t.TempDir()`** for temporary file needs

### Test Types

1. **Unit tests** (`TestXxx`): Test individual functions
2. **Benchmarks** (`BenchmarkXxx`): Measure performance
3. **Fuzz tests** (`FuzzXxx`): Find edge cases with random input

### General Guidelines

1. Tests should be independent and isolated
2. Test the public API when possible; test internals only when necessary
3. Use descriptive test names that explain what is being tested
4. Keep test logic simple; complex test logic may indicate a need to refactor the code under test
5. Run tests with race detector (`go test -race`) to catch concurrency issues
6. Aim for meaningful coverage, not just high percentages

---

## Additional Resources

- [Go testing package documentation](https://pkg.go.dev/testing)
- [The Cover Story](https://go.dev/blog/cover) - Understanding test coverage in Go
- [Go Fuzzing Tutorial](https://go.dev/security/fuzz/)
- [Testify on GitHub](https://github.com/stretchr/testify)

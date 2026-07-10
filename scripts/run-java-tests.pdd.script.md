# Run Java Tests

This script runs the Java test suite using Gradle.

## Parameters

| Name      | Type    | Description                                           | Required | Default Value |
| --------- | ------- | ----------------------------------------------------- | -------- | ------------- |
| verbose   | boolean | Whether to run tests in verbose mode                  | No       | true          |
| test_path | string  | Path to specific test class or package to run         | No       | (all tests)   |

## Operations

1. Verify that `java` and `gradle` (or the Gradle wrapper) are installed by running `java --version` and `./gradlew --version`.
2. Construct the test command with the specified parameters:
   - Start with `./gradlew test`
   - Add `--info` if verbose is true
   - If test_path is specified, add `--tests "<test_path>"`
3. Execute the constructed test command.
4. Capture the test execution output including test results.
5. Display the results to the user, referencing the generated HTML report at `build/reports/tests/test/index.html` if available.

## Output

The script outputs:

- Test execution summary (number of tests passed/failed/skipped)
- Any failure details with file and line references
- Path to the generated HTML test report

## Error Handling

- If `java` or `gradle` are not installed, the script will report an error.
- If tests fail, the script will report the failure details.

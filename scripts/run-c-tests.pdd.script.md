# Run C Tests

This script runs the C test suite using the project's Makefile.

## Parameters

| Name      | Type    | Description                                           | Required | Default Value |
| --------- | ------- | ----------------------------------------------------- | -------- | ------------- |
| verbose   | boolean | Whether to run tests in verbose mode                  | No       | true          |

## Operations

1. Verify that `make` and a C compiler (`gcc` or `clang`) are installed by running `which make` and checking for a compiler.
2. Verify that a `Makefile` exists in the project directory.
3. Construct the test command using the Makefile's `test` target:
   - Run `make test`
4. Execute the constructed test command.
5. Capture the test execution output including test results.
6. Display the results to the user.

## Output

The script outputs:

- Test execution summary (pass/fail for each test)
- Any failure details with file and line references
- Build warnings if any

## Error Handling

- If `make` or a C compiler are not installed, the script will report an error.
- If no `Makefile` is found, the script will report an error.
- If the `test` target does not exist in the Makefile, the script will report an error.
- If tests fail, the script will report the failure details.

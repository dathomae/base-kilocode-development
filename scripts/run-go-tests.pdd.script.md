# Run Go Tests

This script runs the Go test suite for the project.

## Parameters

| Name      | Type    | Description                                           | Required | Default Value |
| --------- | ------- | ----------------------------------------------------- | -------- | ------------- |
| verbose   | boolean | Whether to run tests in verbose mode                  | No       | true          |
| test_path | string  | Path to specific test package or directory to run     | No       | ./...         |

## Operations

1. Change to the `src/backend` directory.
2. Verify that `go` is installed by running `go version`.
3. Construct the test command with the specified parameters:
   - Start with `go test`
   - Add `-v` if verbose is true
   - Add the test_path at the end
4. Execute the constructed test command.
5. Capture the test execution output including test results.
6. Display the results to the user.

## Output

The script outputs:

- Test execution summary (number of tests passed/failed/skipped)
- Any failure details with file and line references

## Error Handling

- If `go` is not installed, the script will report an error.
- If tests fail, the script will report the failure details.

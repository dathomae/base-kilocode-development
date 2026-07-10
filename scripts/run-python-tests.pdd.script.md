# Run Python Tests

This script runs the Python test suite using pytest with coverage reporting.

## Parameters

| Name      | Type    | Description                                           | Required | Default Value |
| --------- | ------- | ----------------------------------------------------- | -------- | ------------- |
| verbose   | boolean | Whether to run tests in verbose mode                  | No       | true          |
| test_path | string  | Path to specific test file or directory to run        | No       | tests/        |

## Operations

1. Verify that `python` and `pytest` are installed by running `python --version` and `pytest --version`.
2. Construct the test command with the specified parameters:
   - Start with `pytest`
   - Add `-v` if verbose is true
   - Add `--cov=src` for coverage reporting
   - Add `--cov-report=term-missing` to show uncovered lines
   - Add the test_path at the end
3. Execute the constructed test command.
4. Capture the test execution output including test results and coverage data.
5. Display the results to the user.

## Output

The script outputs:

- Test execution summary (number of tests passed/failed/skipped)
- Coverage report showing covered and missing lines per file
- Any failure details with file and line references

## Error Handling

- If `python` or `pytest` are not installed, the script will report an error.
- If tests fail, the script will report the failure details.

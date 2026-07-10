# Run Rust Tests

This script runs the Rust test suite using Cargo.

## Parameters

| Name      | Type    | Description                                           | Required | Default Value |
| --------- | ------- | ----------------------------------------------------- | -------- | ------------- |
| verbose   | boolean | Whether to run tests in verbose mode                  | No       | true          |
| test_path | string  | Path to specific test module or test function to run  | No       | (all tests)   |

## Operations

1. Verify that `rustc` and `cargo` are installed by running `rustc --version` and `cargo --version`.
2. Construct the test command with the specified parameters:
   - Start with `cargo test`
   - If test_path is specified, add the test path as an argument (e.g., `cargo test <test_name>`)
3. Execute the constructed test command.
4. Capture the test execution output including test results.
5. Display the results to the user.

## Output

The script outputs:

- Test execution summary (number of tests passed/failed/skipped)
- Any failure details with file and line references
- Doc test results if doc tests are present

## Error Handling

- If `rustc` or `cargo` are not installed, the script will report an error.
- If tests fail, the script will report the failure details.
- If the project does not contain a `Cargo.toml` file, the script will report an error.

# Create New PDD Script

This script creates a new Prompt Driven Development (PDD) script file based on the provided parameters.

## Parameters

| Name | Type | Description | Required | Default Value |
|------|------|-------------|----------|---------------|
| script_name | string | The name of the new PDD script file (without .pdd.script.md extension) | Yes | N/A |
| purpose | string | A brief description of what the script does | Yes | N/A |
| steps | list | A list of sequential steps to perform in the script | Yes | N/A |
| output_spec | string | Specification of output items and format | Yes | N/A |

## Operations

1. Validate that all required parameters are provided.
2. Create the script file path using the script_name with .pdd.script.md extension.
3. Write the script header with the purpose description.
4. Add the parameters section formatted as a table.
5. Add the operations section with numbered steps.
6. Add the output specification section.
7. Save the file to the scripts directory.

## Output

The script outputs the path to the newly created PDD script file and a confirmation message indicating success.
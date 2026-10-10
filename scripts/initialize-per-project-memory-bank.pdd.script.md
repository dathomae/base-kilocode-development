# Initialize Per-Project Memory Bank

This script initializes a complete per-project memory bank structure in a specified project directory, including all directories and files referenced in the global organization rules. Normally the project root directory will be the root directory of the project that is open when the command to initialize the memory bank is given.

## Parameters

| Name         | Type   | Description                                                                  | Required | Default Value |
| ------------ | ------ | ---------------------------------------------------------------------------- | -------- | ------------- |
| project_path | string | The path to the project root directory where the memory bank will be created | Yes      | N/A           |

## Operations

1. Create the memory-bank directory in the project root.
2. Create the scripts directory in the project root.
3. Create the resources directory in the project root.
4. Create the templates subdirectory under resources.
5. Create the apis subdirectory under resources.
6. Create toc.md files in templates and apis directories with empty list sections.
7. Create terms.md in memory-bank with empty list section.
8. Create concepts.md in memory-bank with empty list section.
9. Create lessons-learned.md in memory-bank with empty list section.
10. Create bugs.md in memory-bank with empty list section.
11. Create brief.md in memory-bank with placeholder content.
12. Create context.md in memory-bank with placeholder content.
13. Create requirements.md in memory-bank with placeholder content.
14. Create design.md in memory-bank with placeholder content.

## Output

The script outputs the path to the initialized memory bank directory and a confirmation message indicating all files and directories were created successfully.

# Archive Stories

This script automates the process of archiving completed stories by moving them from the active stories directory to the finished-stories directory. It supports archiving a specific story file or archiving all stories where all tasks are marked as "Completed" or "Done". The script verifies story completion, updates the table of contents, and reports which stories were archived.

## Parameters

| Name         | Type   | Description                                                                                                                    | Required | Default Value             |
| ------------ | ------ | ------------------------------------------------------------------------------------------------------------------------------ | -------- | ------------------------- |
| story_file   | string | Specific story file to archive (e.g., "Story001_workspace-setup.md"). If not provided, all completed stories will be archived. | No       | N/A                       |
| project_root | string | The root directory of the project where memory-bank/stories is located                                                         | No       | Current working directory |

## Operations

1. Validate parameter constraints:
   - If story*file is provided, it must match the pattern "StoryNNN*\*.md" where NNN is a three-digit number

2. Construct the stories directory path by combining project_root with `memory-bank/stories`.

3. Construct the finished-stories directory path by combining project_root with `memory-bank/stories/finished-stories`.

4. Verify that the `memory-bank/stories` directory exists. If it does not exist, pause execution and inform the user that there are no stories to archive. Ask if the script should be terminated.

5. Check if the `memory-bank/stories/finished-stories` directory exists. If it does not exist, pause execution and ask the user if the directory should be created or if the script should be terminated.

6. Read the `memory-bank/stories/toc.md` file to identify the current story organization:
   - Parse the "Active Stories" section to extract story entries with their filenames, descriptions, and statuses
   - Parse the "Finished Stories" section to understand the existing archived stories
   - If toc.md does not exist or is missing either section, pause execution and inform the user. Ask if the script should be terminated.

7. Identify stories to archive based on whether story_file was provided:

   **If story_file is provided:**
   a. Verify the story file exists in `memory-bank/stories/`. If not, pause execution and inform the user. Ask if the script should be terminated.
   b. Read the story file content to verify that all tasks are marked as "Completed" or "Done":
   - Locate the "Tasks" section in the story file
   - Check each task entry (typically marked with checkboxes like "[x]" or status indicators)
   - If any task is not marked as complete (e.g., "[ ]", "Not Started", "In Progress"), pause execution and inform the user that the story is not complete. Ask if the script should be terminated.
     c. Add the story to the list of stories to archive.

   **If no story_file is provided:**
   a. Scan all story files in `memory-bank/stories/` that match the pattern "StoryNNN\_\*.md".
   b. For each story file, read its content to check if all tasks are marked as "Completed" or "Done":
   - Locate the "Tasks" section in the story file
   - Check each task entry for completion status
   - Consider a story complete only if ALL tasks are marked as "Completed", "Done", or "[x]"
     c. Create a list of all stories that are fully completed.
     d. If no completed stories are found, report this to the user and terminate the script successfully.

8. Verify user intent before proceeding:
   - Display a summary of the stories that will be archived (filename, description from toc.md)
   - Display the total number of stories to be archived
   - Ask the user to confirm if they want to proceed with archiving these stories
   - If the user declines, terminate the script

9. For each story identified for archiving, perform the following operations:
   a. Move the story file from `memory-bank/stories/` to `memory-bank/stories/finished-stories/`.
   b. If the move operation fails (e.g., permission denied, file already exists), pause execution and ask the user if script execution should continue with remaining stories or if the script should be terminated.

10. Update the `memory-bank/stories/toc.md` file to reflect the archived stories:
    a. For each archived story:
        - Remove the story's entry from the "Active Stories" section. A list entry is a two-line block (bold title line and colon-prefixed description line) separated from adjacent entries by blank lines. Remove the entire block along with any surrounding blank lines to avoid extra whitespace.
        - Add the story's entry to the "Finished Stories" section with status "Done". The entry must follow the list format: a bold linked title line (`**[filename](path)** — _Done_`) followed by a colon-prefixed description line (`: <description>`), separated from other entries by blank lines.
        - Update the story file link in the "Finished Stories" section to use the relative path `finished-stories/<filename>`.
    b. Maintain proper list formatting and ensure the lists remain valid markdown.

11. Verify all operations completed successfully:
    - Confirm all archived story files exist in `memory-bank/stories/finished-stories/`
    - Confirm no archived story files remain in `memory-bank/stories/`
    - Confirm `memory-bank/stories/toc.md` has been updated correctly (entries moved from Active to Finished Stories section)

## Output

The script outputs:

- Summary of the archive operation including:
  - Total number of stories archived
  - List of archived stories with their filename and description
- For each archived story:
  - Confirmation that the story file was moved successfully
  - Verification that all tasks were marked complete
- Confirmation that `memory-bank/stories/toc.md` was updated successfully
- Any warnings or issues encountered during the operation (e.g., stories skipped because they were not complete)
- Final status of the script execution (success or partial success with details)

## Error Handling

If required constraints aren't met, the user will be prompted. The operations in the script WILL NOT be executed until/unless all parameters are valid. During execution, if an error is encountered (e.g., permission denied when moving files, missing directory, file not found), the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

Specific error conditions handled:

- Invalid story*file format (must match "StoryNNN*\*.md" pattern)
- Stories directory does not exist
- finished-stories directory does not exist
- toc.md file is missing or malformed (missing required sections)
- Story file to archive does not exist in memory-bank/stories/
- Story is not complete (tasks not all marked as "Completed" or "Done")
- Permission denied when moving files
- Destination file already exists in finished-stories/
- Failure to update toc.md files

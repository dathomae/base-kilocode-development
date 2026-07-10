# Add Story to GitHub Backlog

This script creates new backlog items (GitHub issues) in the Chronocone Kanban project using the GitHub CLI (`gh`). Each issue represents a story from the local story files and includes the story number, description, goal, and a reference to the story file location.

The script supports processing multiple stories in a single execution. Steps 1-7 (authentication, validation, and setup) are executed once, then steps 8-12 (issue creation) are executed for each story.

## Parameters

| Name         | Type   | Description                                                                                                | Required | Default Value             |
| ------------ | ------ | ---------------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| story_files  | string | List of story file names to process (e.g., "Story008.md,Story009.md" or "Story008.md through Story010.md") | Yes      | N/A                       |
| repository   | string | The GitHub repository in the format "owner/repo" (e.g., "horatiar/chronocone")                             | No       | "horatiar/chronocone"     |
| project      | string | The GitHub project name to add the issue to (e.g., "Chronocone Kanban")                                    | No       | "Chronocone Kanban"       |
| project_root | string | The root directory of the project where memory-bank/stories is located                                     | No       | Current working directory |

## Operations

### Phase 1: Initial Setup (Execute Once)

1. **GitHub CLI Authentication Setup** (Initial step for users who haven't authenticated):
   - Check if `gh` command is available in the system PATH
   - Verify the user is authenticated with `gh auth status`
   - If not authenticated, inform the user they need to run:
     - `gh auth login` to authenticate with GitHub
     - Then `gh auth refresh -s project` to get project access scopes
   - Pause execution and ask if the script should be terminated

2. Validate parameter constraints:
   - story_files must be provided and non-empty
   - Parse story_files to extract individual story file names:
     - Split comma-separated values (e.g., "Story008.md,Story009.md")
     - Expand range syntax (e.g., "Story008.md through Story010.md" → Story008.md, Story009.md, Story010.md)
     - For range syntax, the story numbers are the endpoints (inclusive on both ends)
     - All stories with numbers that fall in the range should be processed
   - Each story file must match the pattern "StoryNNN\_\*.md" where NNN is a three-digit number
   - repository defaults to "horatiar/chronocone" if not provided
   - project defaults to "Chronocone Kanban" if not provided

3. Construct the stories directory path by combining project_root with `memory-bank/stories`.

4. Verify all story files exist at `memory-bank/stories/<story_file>` for each story in the list:
   - For each story file in the list, verify it exists
   - If any story file does not exist, pause execution and inform the user which files are missing
   - Ask if the script should be terminated

5. Verify the repository exists and the user has access:
   - Run `gh repo view <repository>` to verify access
   - Extract the owner from the repository string (format: "owner/repo")
   - If the repository doesn't exist or the user lacks access, pause execution and inform the user. Ask if the script should be terminated.

6. Verify the project exists and get the project number:
   - Run `gh project list --owner <repo_owner>` to list available projects
   - Find the project with the matching name specified in the `project` parameter
   - Extract the project number from the list output
   - If the project doesn't exist or the user lacks access, pause execution and inform the user. Ask if the script should be terminated.

7. **Prepare for processing:**
   - Initialize an empty list to track created issues
   - Display the total number of stories to be processed

### Phase 2: Per-Story Processing (Execute for Each Story)

**For each story_file in the story_files list, repeat steps 8-12:**

8. Read the story file content and construct the issue title:
   - Extract the story number from the filename (e.g., "008" from "Story008_docker-demo-container.md")
   - Extract the story title from the first line (format: "# StoryNNN: Human Readable Title")
   - Extract the goal from the "## Goal" section (the paragraph following the Goal header)
   - Construct the issue title: Format: "Story<NNN>: <Brief Description>"
   - Example: "Story008: Docker Demo Container"

9. Construct the issue description:
   - Include the "## Goal" section content
   - Add a reference to the story file location with annotation
   - Format:

     ```
     ## Goal

     <goal text from story file>

     ## Story File

     Details are available in the story file: `memory-bank/stories/<story_file>`
     ```

10. Create the GitHub issue using `gh issue create`:
    - Use the constructed title
    - Use the constructed description
    - Add the label "story"
    - Add to the specified project using the `--project` flag with the project name
    - Command format: `gh issue create --repo <repository> --title "<title>" --body "<description>" --label "story" --project "<project name>"`

11. Capture the output from the `gh issue create` command to extract the issue URL and issue number.

12. Verify the issue was created successfully:
    - Confirm the issue URL was returned
    - Run `gh issue view <issue_number> --repo <repository>` to verify the issue exists with correct title and labels
    - Add the created issue details (story number, issue URL, issue number) to the tracking list
    - Display confirmation for this story

## Output

The script outputs:

- **Phase 1 Confirmation:**
  - Total number of stories to be processed
  - Repository where issues will be created
  - Project where issues will be added (with project number)

- **Phase 2 - Per Story Confirmation:** For each story processed:
  - Story number (e.g., "Story008")
  - Issue title
  - Issue URL
  - Confirmation that the 'story' label was applied
  - Confirmation that the issue was added to the specified project
  - Summary of the issue content:
    - Story file referenced
    - Goal excerpt (first 200 characters)

- **Final Summary:**
  - Total number of issues successfully created
  - List of all created issues with their URLs
  - Any warnings encountered (e.g., if the 'story' label didn't exist and was created automatically)
  - Any failures that occurred during processing

## Error Handling

If required constraints aren't met, the user will be prompted. The operations in the script WILL NOT be executed until/unless all parameters are valid. During execution, if an error is encountered, the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

### Phase 1 Error Conditions (Setup Phase - Stop Before Processing):

- Missing required parameter (story_files)
- Invalid story file format (each file must match "StoryNNN\_\*.md" pattern)
- One or more story files do not exist in memory-bank/stories/
- GitHub CLI (`gh`) not installed
- User not authenticated with GitHub CLI
- Missing project scopes (requires read:project)
- Repository doesn't exist or user lacks access
- Project doesn't exist or user lacks access

### Phase 2 Error Conditions (Per-Story Processing):

During per-story processing, if an error occurs:

- The error will be logged for the current story
- The user will be asked whether to:
  - **Continue**: Skip the failed story and proceed to the next one
  - **Retry**: Attempt to create the issue for this story again
  - **Terminate**: Stop processing and report results for stories processed so far

Specific per-story error conditions:

- Network connectivity issues when creating the issue
- Failure to create the issue (e.g., permission denied, rate limiting)
- Failure to add issue to project (e.g., project not found, permission denied)
- Label 'story' doesn't exist (gh will create it automatically, but this will be noted)
- Story file cannot be read or parsed (missing Goal section, etc.)

## Example Usage

**Creating backlog items for multiple stories:**

- story_files: "Story008_docker-demo-container.md,Story010_metadata-update-delete.md,Story013_storage-interface.md"
- repository: "horatiar/chronocone" (uses default if not specified)
- project: "Chronocone Kanban" (uses default if not specified)

**Expected output:**

```
Phase 1: Initial Setup
----------------------
✓ GitHub CLI authenticated
✓ 3 story files validated
✓ Repository access confirmed: horatiar/chronocone
✓ Project found: Chronocone Kanban (project #1)

Phase 2: Processing Stories
---------------------------
[1/3] Processing Story008_docker-demo-container.md...
  ✓ Issue created: https://github.com/horatiar/chronocone/issues/42
  ✓ Title: "Story008: Docker Demo Container"
  ✓ Label: "story" applied
  ✓ Added to project: Chronocone Kanban

[2/3] Processing Story010_metadata-update-delete.md...
  ✓ Issue created: https://github.com/horatiar/chronocone/issues/43
  ✓ Title: "Story010: Metadata Update Delete"
  ✓ Label: "story" applied
  ✓ Added to project: Chronocone Kanban

[3/3] Processing Story013_storage-interface.md...
  ✓ Issue created: https://github.com/horatiar/chronocone/issues/44
  ✓ Title: "Story013: Storage Interface"
  ✓ Label: "story" applied
  ✓ Added to project: Chronocone Kanban

Final Summary
-------------
✓ Successfully created 3 issues
  - Story008: https://github.com/horatiar/chronocone/issues/42
  - Story010: https://github.com/horatiar/chronocone/issues/43
  - Story013: https://github.com/horatiar/chronocone/issues/44
```

**Creating a backlog item for a single story (backward compatible):**

- story_files: "Story008_docker-demo-container.md"
- repository: "horatiar/chronocone" (uses default if not specified)
- project: "Chronocone Kanban" (uses default if not specified)

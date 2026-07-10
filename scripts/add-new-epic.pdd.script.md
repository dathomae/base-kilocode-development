# Add New Epic

This script automates the process of adding a new epic to the project. It generates the next available epic number by scanning existing epics in the memory-bank/epics directory, creates a new epic file from the template, updates the epics toc.md, and manages story associations with appropriate linking based on story file existence.

## Parameters

| Name         | Type   | Description                                                                                                                                                                                                                                                                                           | Required | Default Value             |
| ------------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| epic_name    | string | Short descriptive name for the epic (e.g., "database-layer"). Must contain only alphanumeric characters, hyphens, and underscores.                                                                                                                                                                    | Yes      | N/A                       |
| description  | string | Brief description of the epic's ultimate goal (1-3 sentences)                                                                                                                                                                                                                                         | Yes      | N/A                       |
| stories      | list   | List of stories associated with this epic. Each story requires: story_name (string), purpose (string), and optionally associated_ero (string - ERO filename if specifically associated). Story files may not exist yet.                                                                               | Yes      | N/A                       |
| references   | list   | List of reference documents associated with this epic (requirements, architecture, design, or ERO files). Each reference requires: doc_type (string - one of: requirements, architecture, design, ero), doc_path (string - relative path from memory-bank), description (string - brief description). | No       | Empty list                |
| dependencies | list   | List of epic or story dependencies. Each dependency requires: dep_type (string - "epic" or "story"), dep_filename (string), description (string).                                                                                                                                                     | No       | Empty list                |
| project_root | string | The root directory of the project where memory-bank/epics is located                                                                                                                                                                                                                                  | No       | Current working directory |

## Operations

1. Validate parameter constraints:
   - epic_name must be provided and non-empty
   - epic*name must match the pattern `^[a-zA-Z0-9*-]+$` (alphanumeric, hyphens, underscores only)
   - description must be provided and non-empty
   - description should not exceed 300 characters
   - stories must be provided as a list with at least one story
   - Each story must have story*name (non-empty, matches `^[a-zA-Z0-9*-]+$`) and purpose (non-empty, max 200 chars)
   - If associated*ero is provided for a story, it must match the pattern `^ERO\d{3}*[a-zA-Z0-9_-]+\.md$`
   - Each reference doc_type must be one of: requirements, architecture, design, ero
   - Each reference doc_path must be non-empty and use forward slashes

2. Construct the epics directory path by combining project_root with `memory-bank/epics`.

3. Verify that the `memory-bank/epics` directory exists. If it does not exist:
   - Create the directory
   - Create a toc.md file with the standard epics table of contents skeleton:

     ```markdown
     # Epics Table of Contents

     This directory contains epic stories that group related user stories into larger initiatives.
     ```

4. Scan for existing epic files to determine the highest epic number:
   - List all files in `memory-bank/epics/` matching the pattern "EpicNNN\_\*.md" where NNN is a three-digit number
   - Extract the numeric portion (NNN) from each filename
   - Identify the maximum epic number found
   - If no epic files exist, use 0 as the maximum

5. Calculate the next epic number:
   - next_number = max_epic_number + 1
   - Format the number as a three-digit string with leading zeros (e.g., 1 becomes "001", 40 becomes "040")

6. Construct the new epic filename:
   - Format: `Epic<NNN>_<epic_name>.md` (e.g., "Epic001_database-layer.md")

7. Read the epic template from `resources/templates/epics.md`:
   - Verify the template file exists. If not, pause execution and inform the user. Ask if the script should be terminated.
   - Read the template content

8. For each story in the stories list, determine the link format:
   - Construct the expected story filename: `Story<NNN>_<story_name>.md` (note: story number is unknown, use placeholder or check if any story file exists with this name pattern)
   - Search for any existing story files matching pattern `*_<story_name>.md` in `memory-bank/stories/` and `memory-bank/stories/finished-stories/`
   - If found: use link format `[Story<NNN>_<story_name>.md](../stories/Story<NNN>_<story_name>.md)` with the actual story number
   - If not found: use text `StoryXXX_<story_name>.md (not yet defined)` where XXX is a placeholder indicating the story doesn't exist yet
   - If associated_ero is provided for the story, note the ERO reference for inclusion in the story table

9. Process the references list:
   - For each reference, format the link based on doc_type:
     - requirements: `[description](specifications/requirements/<doc_path>)`
     - architecture: `[description](specifications/architecture/<doc_path>)`
     - design: `[description](specifications/design/<doc_path>)`
     - ero: `[description](specifications/design/<doc_path>)` or other appropriate location

10. Process the dependencies list:
    - For each dependency, format the link:
      - epic: `[EpicNNN_<name>.md](../epics/EpicNNN_<name>.md) - description`
      - story: `[StoryNNN_<name>.md](../stories/StoryNNN_<name>.md) - description`
    - If dependencies list is empty, use "None" with a brief explanation

11. Create the new epic file at `memory-bank/epics/<filename>`:
    - Start with the epic header: `# Epic<NNN>: <Human-Readable Epic Name>`
    - Include the description in the Ultimate Goal section
    - Include the formatted References section with all provided references
    - Include the Dependencies section with formatted dependencies
    - Create the Stories table with columns: Story, Status, Description (and ERO column if any stories have associated EROs)
    - For each story, add a row:
      - Story: Use link format if file exists, otherwise "not yet defined" text
      - Status: "Not Started" (default for all stories in a new epic)
      - Description: The purpose provided for the story
      - ERO: If associated_ero is provided, include link `[ERO](path/to/ero)`
    - Include the Story Ordering and Dependencies section with placeholder text for the user to fill in
    - Include the Constraints section with placeholder text
    - Include the Acceptance Criteria section with placeholder text
    - Preserve other template sections as appropriate

12. Read the current `memory-bank/epics/toc.md` file:
    - If the file doesn't exist, create it with the standard epics table of contents skeleton:

      ```markdown
      # Epics Table of Contents

      This directory contains epic stories that group related user stories into larger initiatives.
      ```

13. Add the new epic entry to the epics toc.md:
    - Insert a blank line before the new entry block
    - Insert the entry as a two-line block:
      - Line 1: `**[Epic<NNN>_<epic_name>.md](Epic<NNN>_<epic_name>.md)** — _Not Started_`
      - Line 2: `: <description>` (the provided description parameter, truncated if necessary)
    - Maintain proper list formatting
    - Insert the new entry in epic number order

14. Verify the new epic file was created successfully:
    - Confirm the file exists at the expected path
    - Confirm the file is readable and contains the expected content
    - Verify all story references are properly formatted
    - Verify all document references are properly formatted

15. Verify the toc.md file was updated successfully:
    - Confirm the new epic entry appears in the epics toc.md
    - Confirm the entry has the correct filename, description, and status

16. If any story files referenced in the epic already exist, check if those story files need to be updated:
    - For each existing story file, check if it has an epic reference
    - If not, add a link to this epic in the story's Goal section

## Output

The script outputs:

- Confirmation of the new epic creation including:
  - Epic number assigned (e.g., "Epic001")
  - Full filename (e.g., "Epic001_database-layer.md")
  - Absolute path to the created file
  - The description that was used
- Summary of stories included in the epic:
  - Number of stories defined
  - For each story: name, whether file exists (linked or not yet defined), and associated ERO if any
- Summary of references included:
  - Number of requirement docs
  - Number of architecture docs
  - Number of design docs
  - Number of EROs
- Confirmation that the epic was added to the epics toc.md
- Any warnings encountered (e.g., if story files don't exist yet, they'll be marked as "not yet defined")

## Error Handling

If required constraints aren't met, the user will be prompted. The operations in the script WILL NOT be executed until/unless all parameters are valid. During execution, if an error is encountered, the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

Specific error conditions handled:

- Missing required parameter (epic_name or description)
- Invalid epic_name format (contains characters other than alphanumeric, hyphens, underscores)
- Description exceeds maximum length (300 characters)
- Empty stories list or invalid story entries
- Invalid story_name format in stories list
- Invalid associated_ero format (must match ERONNN_name.md pattern)
- Invalid doc_type in references (not one of: requirements, architecture, design, ero)
- Epic template file (resources/templates/epics.md) is missing
- Failure to create epics directory
- Failure to create or read toc.md
- Epic file already exists (collision with existing filename)
- Permission denied when creating the epic file

## Example Usage

**Creating a new epic with stories and references:**

- epic_name: "data-persistence-layer"
- description: "Implement the complete data persistence layer including relational storage, connection pooling, and migration systems"
- stories:
  - story_name: "relational-schema", purpose: "Define and implement the core relational database schema", associated_ero: "ERO001_database-technology-choice.md"
  - story_name: "connection-pooling", purpose: "Implement connection pooling for database efficiency", associated_ero: null
  - story_name: "migration-system", purpose: "Create the database migration and versioning system", associated_ero: null
- references:
  - doc_type: "architecture", doc_path: "database_design.md", description: "Database Architecture Design"
  - doc_type: "design", doc_path: "storage_interface.md", description: "Storage Interface Design"
  - doc_type: "ero", doc_path: "ERO001_database-technology-choice.md", description: "Database Technology ERO"
- dependencies: []

**Expected output:**

- Epic number 001 assigned
- File created: memory-bank/epics/Epic001_data-persistence-layer.md
- 3 stories defined in epic:
  - relational-schema: linked to existing Story041_relational-schema-implementation.md
  - connection-pooling: not yet defined
  - migration-system: not yet defined
- 3 references included (1 architecture, 1 design, 1 ERO)
- Entry added to toc.md

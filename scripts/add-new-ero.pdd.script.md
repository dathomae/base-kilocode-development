# Add New ERO

This script automates the process of issuing a new Epic Restraining Order (ERO) to the project. It generates the next available ERO number by scanning existing EROs, creates a new ERO file from the template, and updates the relevant toc.md file.

> **Note:** The pre-2026-09-16 EROs were moved to `memory-bank/specifications/design/historical-archive/` and are frozen. This script does not scan or write to that archive. New EROs are created fresh under `memory-bank/specifications/<ero_location>/eros/`, aligned with the current organization and goals; because the live directory starts empty, numbering restarts at ERO001.

## Parameters

| Name                  | Type   | Description                                                                                                                                   | Required | Default Value             |
| --------------------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| ero_name              | string | Short descriptive name for the ERO (e.g., "database-technology-choice"). Must contain only alphanumeric characters, hyphens, and underscores. | Yes      | N/A                       |
| title                 | string | Clear, concise title describing the technical decision being restrained.                                                                      | Yes      | N/A                       |
| purpose               | string | Brief statement (1-3 sentences) explaining what technical decision needs to be made and why.                                                  | Yes      | N/A                       |
| decision_status       | string | Current status of the decision: "Draft", "Decided", or "Superseded".                                                                          | Yes      | N/A                       |
| implementation_status | string | Implementation state: "Not Yet Implemented" or "Implemented".                                                                                 | Yes      | N/A                       |
| related_epics         | list   | List of related epic filenames (e.g., ["Epic001_data-persistence.md"]). May be empty.                                                         | No       | Empty list                |
| related_documents     | list   | List of related documents. Each document requires: doc_path (string - relative path), description (string - brief description).               | No       | Empty list                |
| ero_location          | string | Location for the ERO: "architecture" or "design". Determines whether it goes in specifications/architecture/ or specifications/design/.       | No       | "architecture"            |
| project_root          | string | The root directory of the project where the ERO will be created.                                                                              | No       | Current working directory |

## Operations

1. Validate parameter constraints:
   - ero_name must be provided and non-empty
   - ero*name must match the pattern `^[a-zA-Z0-9*-]+$` (alphanumeric, hyphens, underscores only)
   - title must be provided and non-empty
   - title should not exceed 150 characters
   - purpose must be provided and non-empty
   - purpose should not exceed 300 characters
   - decision_status must be one of: "Draft", "Decided", "Superseded"
   - implementation_status must be one of: "Not Yet Implemented", "Implemented"
   - ero_location must be one of: "architecture", "design"
   - Each epic filename must match the pattern `^Epic\d+_[a-zA-Z0-9_-]+\.md$`

2. Construct the ERO directory path by combining project_root with `memory-bank/specifications/<ero_location>/eros`.

3. Verify that the ERO directory exists. If it does not exist:
   - Create the directory
   - Create a toc.md file with the standard ERO TOC header containing a list-formatted EROs section

4. Scan for existing ERO files to determine the highest ERO number:
   - List all files in the live ERO directory (the frozen `historical-archive/` is never scanned) matching the pattern "ERONNN\_\*.md" where NNN is a three-digit number
   - Extract the numeric portion (NNN) from each filename
   - Identify the maximum ERO number found
   - If no ERO files exist, use 0 as the maximum

5. Calculate the next ERO number:
   - next_number = max_ero_number + 1
   - Format the number as a three-digit string with leading zeros (e.g., 1 becomes "001", 40 becomes "040")

6. Construct the new ERO filename:
   - Format: `ERO<NNN>_<ero_name>.md` (e.g., "ERO001_database-technology-choice.md")

7. Read the ERO template from `resources/templates/ero.md`:
   - Verify the template file exists. If not, pause execution and inform the user. Ask if the script should be terminated.
   - Read the template content

8. Process related epics:
   - If related_epics is empty, prepare an empty table placeholder
   - For each epic filename:
     - Verify the epic file exists in `memory-bank/historical-archive/epics/` (epics are archived and frozen; see `memory-bank/historical-archive/README.md`)
     - Extract the epic number from the filename
     - If found: create link `[Epic<NNN>_<name>.md](../epics/Epic<NNN>_<name>.md) - <status>`
     - If not found: use text `Epic<NNN>_<name>.md (not yet defined)`

9. Process related documents:
   - If related_documents is empty, use "None - No related documents at this time."
   - For each document:
     - Create link `[description](<doc_path>)`

10. Create the new ERO file at the calculated path:
    - Start with the header: `# ERO<NNN>: <title>`
    - Add Title section with the provided title
    - Add Purpose/Goal section with the provided purpose
    - Add Decision Status section with the decision_status
    - Add Implementation Status section:
      - Include implementation_status
      - If related_epics is not empty, add Related Epics subsection with a table
      - If empty, add a note: "None - This ERO documents a decision not tied to specific work items."
    - Add Context section with "To Be Written" placeholder for Situation, Constraints, Requirements, and Assumptions
    - Add Evaluation Criteria / Strategy Fit section with "To Be Written" placeholder
    - Add Options Considered section with "To Be Written" placeholder
    - Add Decision section with "To Be Written" placeholder
    - Add Implications section with "To Be Written" placeholder for Positive Implications, Negative Implications/Trade-offs, Impact on System Qualities, and Future Considerations
    - Add Notes section (optional) with related_documents if any

11. Read the current toc.md file in the ERO directory:
    - If the file doesn't exist or doesn't contain an "EROs" section, create/initialize it with the proper header

12. Add the new ERO entry to the EROs section in toc.md:
    - Insert a new entry in list format:
      - Title line: `**[ERO<NNN>_<ero_name>.md](eros/ERO<NNN>_<ero_name>.md)** — _{decision_status}_`
      - Description line: `: <brief summary from the purpose (truncated if necessary to 100 characters)>`
    - Maintain proper list formatting with a blank line separating entries
    - Insert the new entry in ERO number order

13. Verify the new ERO file was created successfully:
    - Confirm the file exists at the expected path
    - Confirm the file is readable and contains the expected content
    - Verify all section headers are present
    - Verify "To Be Written" placeholders are in place for body sections

14. Verify the toc.md file was updated successfully:
    - Confirm the new ERO entry appears in the EROs section
    - Confirm the entry has the correct filename, description, and decision status

## Output

The script outputs:

- Confirmation of the new ERO creation including:
  - ERO number assigned (e.g., "ERO001")
  - Full filename (e.g., "ERO001_database-technology-choice.md")
  - Absolute path to the created file
  - The location (architecture or design directory)
  - The title that was used
  - The purpose/summary that was used
- Summary of related items:
  - Number of related epics (with count of existing vs not yet defined)
  - Number of related documents
  - Confirmation that the ERO was added to the EROs section in toc.md
- Any warnings encountered (e.g., if referenced stories or epics don't exist yet)

## Error Handling

If required constraints aren't met, the user will be prompted. The operations in the script WILL NOT be executed until/unless all parameters are valid. During execution, if an error is encountered, the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

Specific error conditions handled:

- Missing required parameter (ero_name, title, purpose, decision_status, implementation_status)
- Invalid ero_name format (contains characters other than alphanumeric, hyphens, underscores)
- Title exceeds maximum length (150 characters)
- Purpose exceeds maximum length (300 characters)
- Invalid decision_status (not one of: Draft, Decided, Superseded)
- Invalid implementation_status (not one of: Not Yet Implemented, Implemented)
- Invalid ero_location (not one of: architecture, design)
- ERO template file (resources/templates/ero.md) is missing
- Invalid epic filename format
- Failure to create EROs directory
- Failure to create or read toc.md
- ERO file already exists (collision with existing filename)
- Permission denied when creating the ERO file

## Example Usage

**Creating a new ERO with related epic:**

- ero_name: "database-technology-choice"
- title: "Database Technology Selection for Primary Data Store"
- purpose: "Select a primary relational database technology for storing configuration data, user information, and transactional records."
- decision_status: "Decided"
- implementation_status: "Implemented"
- related_epics: ["Epic001_data-persistence.md"]
- related_documents:
  - doc_path: "specifications/architecture/backend/foundation/database/overview.md", description: "Database Architecture Design"
- ero_location: "architecture"

**Expected output:**

- ERO number 001 assigned
- File created: memory-bank/specifications/architecture/eros/ERO001_database-technology-choice.md
- Location: architecture
- Title: "Database Technology Selection for Primary Data Store"
- Purpose: "Select a primary relational database technology..."
- 1 related epic (1 existing, 0 not yet defined)
- 1 related document
- Entry added to toc.md EROs section

**Creating a new ERO with no related items:**

- ero_name: "api-versioning-strategy"
- title: "API Versioning Strategy"
- purpose: "Define the approach for versioning REST API endpoints to ensure backward compatibility."
- decision_status: "Draft"
- implementation_status: "Not Yet Implemented"
- related_epics: []
- related_documents: []
- ero_location: "architecture"

**Expected output:**

- ERO number 002 assigned
- File created: memory-bank/specifications/architecture/eros/ERO002_api-versioning-strategy.md
- Location: architecture
- Empty Related Epics table created (to be filled in later)
- Entry added to toc.md EROs section with "Draft" status

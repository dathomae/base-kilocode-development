# Add New ADR

This script automates the process of adding a new Architecture Design Record (ADR) to the project. It generates the next available ADR number by scanning existing ADRs, creates a new ADR file from the template, and updates the relevant toc.md file.

## Parameters

| Name                  | Type   | Description                                                                                                                                   | Required | Default Value             |
| --------------------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| adr_name              | string | Short descriptive name for the ADR (e.g., "database-technology-choice"). Must contain only alphanumeric characters, hyphens, and underscores. | Yes      | N/A                       |
| title                 | string | Clear, concise title describing the architectural decision.                                                                                   | Yes      | N/A                       |
| purpose               | string | Brief statement (1-3 sentences) explaining what architectural decision needs to be made and why.                                              | Yes      | N/A                       |
| decision_status       | string | Current status of the decision: "Draft", "Decided", or "Superseded".                                                                          | Yes      | N/A                       |
| implementation_status | string | Implementation state: "Not Yet Implemented" or "Implemented".                                                                                 | Yes      | N/A                       |
| related_stories       | list   | List of related story filenames (e.g., ["Story041_relational-schema-implementation.md"]). May be empty.                                       | No       | Empty list                |
| related_epics         | list   | List of related epic filenames (e.g., ["Epic001_data-persistence.md"]). May be empty.                                                         | No       | Empty list                |
| related_documents     | list   | List of related documents. Each document requires: doc_path (string - relative path), description (string - brief description).               | No       | Empty list                |
| adr_location          | string | Location for the ADR: "architecture" or "design". Determines whether it goes in specifications/architecture/ or specifications/design/.       | No       | "architecture"            |
| project_root          | string | The root directory of the project where the ADR will be created.                                                                              | No       | Current working directory |

## Operations

1. Validate parameter constraints:
   - adr_name must be provided and non-empty
   - adr*name must match the pattern `^[a-zA-Z0-9*-]+$` (alphanumeric, hyphens, underscores only)
   - title must be provided and non-empty
   - title should not exceed 150 characters
   - purpose must be provided and non-empty
   - purpose should not exceed 300 characters
   - decision_status must be one of: "Draft", "Decided", "Superseded"
   - implementation_status must be one of: "Not Yet Implemented", "Implemented"
   - adr_location must be one of: "architecture", "design"
   - Each story filename must match the pattern `^Story\d+_[a-zA-Z0-9_-]+\.md$`
   - Each epic filename must match the pattern `^Epic\d+_[a-zA-Z0-9_-]+\.md$`

2. Construct the ADR directory path by combining project_root with `memory-bank/specifications/<adr_location>/adrs`.

3. Verify that the ADR directory exists. If it does not exist:
   - Create the directory
   - Create a toc.md file with the standard ADR Table of Contents header

4. Scan for existing ADR files to determine the highest ADR number:
   - List all files in the ADR directory matching the pattern "ADRNNN\_\*.md" where NNN is a three-digit number
   - Extract the numeric portion (NNN) from each filename
   - Identify the maximum ADR number found
   - If no ADR files exist, use 0 as the maximum

5. Calculate the next ADR number:
   - next_number = max_adr_number + 1
   - Format the number as a three-digit string with leading zeros (e.g., 1 becomes "001", 40 becomes "040")

6. Construct the new ADR filename:
   - Format: `ADR<NNN>_<adr_name>.md` (e.g., "ADR001_database-technology-choice.md")

7. Read the ADR template from `resources/templates/adr.md`:
   - Verify the template file exists. If not, pause execution and inform the user. Ask if the script should be terminated.
   - Read the template content

8. Get the current date for proposed_date (YYYY-MM-DD format).

9. Set decision_date based on decision_status:
   - If "Decided": use current date (YYYY-MM-DD format)
   - If "Draft" or "Superseded": use "Pending"

10. Process related stories:
    - If related_stories is empty, prepare an empty table placeholder
    - For each story filename:
      - Verify the story file exists in `memory-bank/planning/stories/active-stories/` or `memory-bank/planning/stories/finished-stories/`
      - Extract the story number from the filename
      - If found: create link `[Story<NNN>_<name>.md](../stories/Story<NNN>_<name>.md) - <status>`
      - If not found: use text `Story<NNN>_<name>.md (not yet defined)`

11. Process related epics:
    - If related_epics is empty, prepare an empty table placeholder
    - For each epic filename:
      - Verify the epic file exists in `memory-bank/epics/`
      - Extract the epic number from the filename
      - If found: create link `[Epic<NNN>_<name>.md](../epics/Epic<NNN>_<name>.md) - <status>`
      - If not found: use text `Epic<NNN>_<name>.md (not yet defined)`

12. Process related documents:
    - If related_documents is empty, use "None - No related documents at this time."
    - For each document:
      - Create link `[description](<doc_path>)`

13. Create the new ADR file at the calculated path:
    - Start with the header: `# ADR<NNN>: <title>`
    - Add Title section with the provided title
    - Add Purpose/Goal section with the provided purpose
    - Add Decision Status section with the decision_status
    - Add Implementation Status section:
      - Include implementation_status
      - If related_stories is not empty, add Related Stories subsection with a table
      - If related_epics is not empty, add Related Epics subsection with a table
      - If both are empty, add a note: "None - This ADR documents a decision not tied to specific work items."
    - Add Proposed Date section with proposed_date
    - Add Decision Date section with decision_date
    - Add Context section with "To Be Written" placeholder for Situation, Constraints, Requirements, and Assumptions
    - Add Evaluation Criteria / Strategy Fit section with "To Be Written" placeholder
    - Add Options Considered section with "To Be Written" placeholder
    - Add Decision section with "To Be Written" placeholder
    - Add Implications section with "To Be Written" placeholder for Positive Implications, Negative Implications/Trade-offs, Impact on System Qualities, and Future Considerations
    - Add Notes section (optional) with related_documents if any

14. Read the current toc.md file in the ADR directory:
    - If the file doesn't exist or doesn't contain an "ADRs" section, create/initialize it with the proper header

15. Add the new ADR entry to the ADRs section in toc.md:
    - Insert a new entry in the list using the three-column format:
      - Title line: `**[ADR<NNN>_<adr_name>.md](adrs/ADR<NNN>_<adr_name>.md)** — _<decision_status>_`
      - Description line: `: <description>` where description is a brief summary from the purpose (truncated if necessary to 100 characters)
    - Insert the new entry in ADR number order
    - Ensure entries are separated by blank lines

16. Verify the new ADR file was created successfully:
    - Confirm the file exists at the expected path
    - Confirm the file is readable and contains the expected content
    - Verify all section headers are present
    - Verify "To Be Written" placeholders are in place for body sections

17. Verify the toc.md file was updated successfully:
    - Confirm the new ADR entry appears in the ADRs section
    - Confirm the entry has the correct filename, description, and decision status

## Output

The script outputs:

- Confirmation of the new ADR creation including:
  - ADR number assigned (e.g., "ADR001")
  - Full filename (e.g., "ADR001_database-technology-choice.md")
  - Absolute path to the created file
  - The location (architecture or design directory)
  - The title that was used
  - The purpose/summary that was used
- Summary of related items:
  - Number of related stories (with count of existing vs not yet defined)
  - Number of related epics (with count of existing vs not yet defined)
  - Number of related documents
- Confirmation that the ADR was added to the ADRs section in toc.md
- Any warnings encountered (e.g., if referenced stories or epics don't exist yet)

## Error Handling

If required constraints aren't met, the user will be prompted. The operations in the script WILL NOT be executed until/unless all parameters are valid. During execution, if an error is encountered, the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

Specific error conditions handled:

- Missing required parameter (adr_name, title, purpose, decision_status, implementation_status)
- Invalid adr_name format (contains characters other than alphanumeric, hyphens, underscores)
- Title exceeds maximum length (150 characters)
- Purpose exceeds maximum length (300 characters)
- Invalid decision_status (not one of: Draft, Decided, Superseded)
- Invalid implementation_status (not one of: Not Yet Implemented, Implemented)
- Invalid adr_location (not one of: architecture, design)
- ADR template file (resources/templates/adr.md) is missing
- Invalid story filename format
- Invalid epic filename format
- Failure to create ADRs directory
- Failure to create or read toc.md
- ADR file already exists (collision with existing filename)
- Permission denied when creating the ADR file

## Example Usage

**Creating a new ADR with all related items:**

- adr_name: "database-technology-choice"
- title: "Database Technology Selection for Primary Data Store"
- purpose: "Select a primary relational database technology for storing configuration data, user information, and transactional records."
- decision_status: "Decided"
- implementation_status: "Implemented"
- related_stories: ["Story041_relational-schema-implementation.md"]
- related_epics: []
- related_documents:
  - doc_path: "specifications/architecture/backend/foundation/database/overview.md", description: "Database Architecture Design"
- adr_location: "architecture"

**Expected output:**

- ADR number 001 assigned
- File created: memory-bank/specifications/architecture/adrs/ADR001_database-technology-choice.md
- Location: architecture
- Title: "Database Technology Selection for Primary Data Store"
- Purpose: "Select a primary relational database technology..."
- 1 related story (1 existing, 0 not yet defined)
- 0 related epics
- 1 related document
- Entry added to toc.md ADRs section

**Creating a new ADR with no related items:**

- adr_name: "api-versioning-strategy"
- title: "API Versioning Strategy"
- purpose: "Define the approach for versioning REST API endpoints to ensure backward compatibility."
- decision_status: "Draft"
- implementation_status: "Not Yet Implemented"
- related_stories: []
- related_epics: []
- related_documents: []
- adr_location: "architecture"

**Expected output:**

- ADR number 002 assigned
- File created: memory-bank/specifications/architecture/adrs/ADR002_api-versioning-strategy.md
- Location: architecture
- Empty Related Stories table created (to be filled in later)
- Empty Related Epics table created (to be filled in later)
- Entry added to toc.md ADRs section with "Draft" status

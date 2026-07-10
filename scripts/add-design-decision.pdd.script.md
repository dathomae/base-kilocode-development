# Add Design Decision

This script automates adding a design decision entry to a `*_design_decisions.md` file co-located with a design document. It either creates a new decisions file (from the template) or appends an entry to an existing one, and updates the relevant `toc.md`.

## Parameters

| Name             | Type   | Description                                                                                                                                                     | Required | Default Value |
| ---------------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------- |
| design_doc_path  | string | Relative path from project root to the design document the decision relates to (e.g., "memory-bank/specifications/design/cpl/go-synthesis-target/synthesis.md") | Yes      | N/A           |
| decision_title   | string | Bold header phrase for the decision, easily searchable (e.g., "Use strconv.FormatFloat for literal value formatting")                                           | Yes      | N/A           |
| decision_summary | string | 1-2 sentence description of what was decided                                                                                                                    | Yes      | N/A           |
| alternatives     | string | Alternatives considered and why rejected (bullet list, one per line)                                                                                            | Yes      | N/A           |
| rationale        | string | Why the chosen option was selected (constraints, principles, tradeoffs)                                                                                         | Yes      | N/A           |
| references       | string | References to design docs and/or EROs (relative paths, one per line)                                                                                            | No       | ""            |

## Operations

1. Validate parameter constraints:
   - design_doc_path must be provided and non-empty
   - decision_title must be provided and non-empty
   - decision_summary must be provided and non-empty
   - alternatives must be provided and non-empty
   - rationale must be provided and non-empty

2. Verify that the design document at `design_doc_path` exists. If not, pause execution and inform the user. Ask if the script should be terminated.

3. Construct the design decisions file path:
   - Extract the design doc filename (e.g., `synthesis.md` → `synthesis`)
   - Construct the decisions filename: `<name>_design_decisions.md`
   - Place it in the same directory as the design doc
   - Example: `memory-bank/specifications/design/cpl/go-synthesis-target/synthesis_design_decisions.md`

4. Check if the design decisions file already exists:
   - If it does NOT exist: read the template from [`resources/templates/design_decision.md`](resources/templates/design_decision.md), remove the placeholder entry, and create the new file with the first decision entry
   - If it DOES exist: read the existing file content

5. Format the new decision entry using the template format:

   ```markdown
   **<decision_title>**
   : <decision_summary>

   : **Alternatives considered:**
   : - <alternative line 1>
   : - <alternative line 2>

   : **Rationale:**
   : <rationale>

   : **References:**
   : - [<doc_name>](relative_path)
   ```

6. Append the formatted entry to the design decisions file (after any existing entries, before the final `---` separator if present).

7. If this is a new file (created from template), update the directory's `toc.md`:
   - Read the toc.md in the same directory as the design doc
   - If it contains a table, add a row for the new `*_design_decisions.md` file with a brief description
   - If no toc.md exists in the directory, note this but do not create one (the user should decide whether to add one)

8. Verify the file was created/updated successfully:
   - Confirm the file exists and is readable
   - Confirm the new entry appears in the file

## Output

The script outputs:

- Confirmation of the decision entry creation including:
  - Full path to the design decisions file
  - Whether the file was newly created or appended to
  - The decision title that was added
- If toc.md was updated: confirmation of the toc.md update
- Any warnings (e.g., no toc.md found in directory)

## Error Handling

If required constraints aren't met, the user will be prompted. The operations in the script WILL NOT be executed until/unless all parameters are valid. During execution, if an error is encountered, the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

Specific error conditions handled:

- Missing required parameters
- Design document path does not exist
- Template file (resources/templates/design_decision.md) is missing
- Permission denied when creating or writing the decisions file
- Failure to read or write toc.md

## Notes

- Design decisions reference design docs and EROs, never stories (which are transient artifacts)
- One design decisions file per design document, co-located in the same directory
- Entries use the lessons-learned.md style: bold header phrase followed by colon and detailed explanation

# Create Planning Directories

This script initializes the `memory-bank/planning` directory tree and its `toc.md` skeletons. It creates the plans, stories, and arcs directories, each split into active, finished, and canceled records, and seeds each directory with a `toc.md` in the list format defined by `memory-bank/specifications/toc-format-spec.md`.

> **Note:** This script was formerly named `create-stories-directory.pdd.script.md`. It now initializes the entire planning tree, matching the organization defined in `.kilo/rules/organization.md`.

## Parameters

| Name         | Type   | Description                                                                  | Required | Default Value             |
| ------------ | ------ | ---------------------------------------------------------------------------- | -------- | ------------------------- |
| project_root | string | The root directory of the project where memory-bank/planning will be created | No       | Current working directory |

## Directory Structure

| Directory                                        | Purpose                                       |
| ------------------------------------------------ | --------------------------------------------- |
| `memory-bank/planning/`                          | Root of all current planning documents        |
| `memory-bank/planning/plans/`                    | Targeted plan files                           |
| `memory-bank/planning/plans/active-plans/`       | Plans not started or in progress              |
| `memory-bank/planning/plans/finished-plans/`     | Plans whose objectives have been achieved     |
| `memory-bank/planning/plans/canceled-plans/`     | Plans that were abandoned or superseded       |
| `memory-bank/planning/stories/`                  | Implementation stories                        |
| `memory-bank/planning/stories/active-stories/`   | Stories not started or in progress            |
| `memory-bank/planning/stories/finished-stories/` | Stories whose tasks have all been completed   |
| `memory-bank/planning/stories/canceled-stories/` | Stories that were abandoned or superseded     |
| `memory-bank/planning/arcs/`                     | Implementation arcs (scheduled plan/story sets) |
| `memory-bank/planning/arcs/active-arcs/`         | Arcs not started or in progress               |
| `memory-bank/planning/arcs/finished-arcs/`       | Arcs whose stories have all been completed    |
| `memory-bank/planning/arcs/canceled-arcs/`       | Arcs that were abandoned or superseded        |
| `memory-bank/planning/milestones/`               | Milestones (delivery increments)              |
| `memory-bank/planning/milestones/active-milestones/` | Milestones not started or in progress     |
| `memory-bank/planning/milestones/finished-milestones/` | Milestones finalized for the `agent-main → main` merge |
| `memory-bank/planning/milestones/canceled-milestones/` | Milestones abandoned or superseded      |

Every directory above receives a `toc.md` (sixteen in total).

## Operations

1. Determine the planning root by combining project_root with `memory-bank/planning`.
2. For each directory in the Directory Structure table, create it if it does not already exist (creating parent directories as needed).
3. For each directory, check whether a `toc.md` already exists.
4. If a `toc.md` does not exist, create it with the appropriate skeleton:
   - The planning root, `plans/`, `stories/`, `arcs/`, and `milestones/` tocs are indexes that link to the tocs of the directories beneath them.
   - The active, finished, and canceled tocs list the individual plans, stories, arcs, or milestones and carry the relevant section heading (`## Active Plans`, `## Finished Plans`, `## Canceled Plans`, `## Active Stories`, `## Finished Stories`, `## Canceled Stories`, `## Active Arcs`, `## Finished Arcs`, `## Canceled Arcs`, `## Active Milestones`, `## Finished Milestones`, or `## Canceled Milestones`), the entry format for that directory, and a placeholder line indicating that there are no entries yet.
   - All tocs use the list format defined by `memory-bank/specifications/toc-format-spec.md`.
5. Never overwrite an existing `toc.md` and never delete existing content; the script is idempotent and only fills in missing directories and files.
6. Verify that all sixteen directories and all sixteen `toc.md` files exist.

## Output

The script outputs:

- Confirmation that the `memory-bank/planning` directory tree exists (or was created)
- For each directory: whether it already existed or was created
- For each `toc.md`: whether it already existed or was created
- The path to the planning root

## Error Handling

If required parameters aren't specified, the user should be prompted. The operations in the script WILL NOT be executed until/unless all required parameters have been specified. During execution, if an error is encountered (e.g., permission denied when creating directories or files), the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

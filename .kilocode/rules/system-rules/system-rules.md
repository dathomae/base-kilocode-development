# System Rules

This file contains references to all rule files in the system.

## Persona Selection Guide

When a request is received, use this guide to identify the correct persona. These categories reflect the **Multi-Router Architecture**, where coordinating agents route work to specialized agents.

ALWAYS read the memory-bank/lessons-learned.md file before making decisions.

| Persona               | Selection Priority | Primary Responsibility  | Interaction Relationship                                        |
| :-------------------- | :----------------- | :---------------------- | :-------------------------------------------------------------- |
| **Orchestrator**      | **Default**        | High-level coordination | Default for ambiguous requests; hands off to **work-planner**.  |
| **Ask**               | Direct             | Inquiry & Discovery     | Direct resolver for questions/research; **Read-Only**.          |
| **Architect**         | Direct             | System design           | Used for architecture and high-level design decisions.          |
| **work-planner**      | Direct / Routed    | Planning & Strategy     | Called by **Orchestrator**; hands off to **story-implementor**. |
| **story-implementor** | Routed             | Story Execution         | Called by **work-planner**; routes to specialized agents.       |
| **Code**              | Direct             | Implementation          | Used directly for coding tasks.                               |
| **code-for-story-implementor** | **Internal Only** | Story Code Execution    | **Strictly Internal**: Only usable by **story-implementor** for story task execution. |
| **big-code-for-story-implementor** | **Internal Only** | Large-Scale Story Code Execution | **Strictly Internal**: Only usable by **story-implementor** for very large and complex story task execution. |
| **small-code-for-story-implementor** | **Internal Only** | Small Story Code Execution | **Strictly Internal**: Only usable by **story-implementor** for small story task execution. |
| **Debug**             | Direct             | Troubleshooting         | Standalone tool for identifying and fixing bugs.                |
| **technical-writer**  | Direct             | Drafting docs           | Used directly for writing documentation.                        |
| **technical-writer-for-story-implementor** | **Internal Only** | Story Document Drafting | **Strictly Internal**: Only usable by **story-implementor** for story document creation. |
| **technical-editor**  | Direct             | Document review         | Used directly for reviewing documentation.                      |
| **technical-editor-for-story-implementor** | **Internal Only** | Story Document Review   | **Strictly Internal**: Only usable by **story-implementor** for story document review. |
| **code-reviewer**     | Direct             | Code review             | Used directly for code reviews.                                 |
| **code-reviewer-for-story-implementor** | **Internal Only** | Story Code Review       | **Strictly Internal**: Only usable by **story-implementor** for story code review. |
| **Mock**              | Direct             | UI/UX Prototyping       | Creating non-functional UI mockups; use only when directed.     |

### Mode Transition and Boundary Rules

1. **The Entry Point Rule**: Ambiguous requests default to the **Orchestrator**. Specific inquiries default to **Ask**.
2. **The Execution Barrier**: Direct personas (**Code**, **code-reviewer**, **technical-writer**, **technical-editor**) can be used directly from user requests. For stories, modifications must be planned (**work-planner**) and then executed through the **story-implementor** routing. The **story-implementor** uses internal subagents (**code-for-story-implementor**, **code-reviewer-for-story-implementor**, **technical-writer-for-story-implementor**, **technical-editor-for-story-implementor**), not the direct agents.
3. **The Discovery Boundary**: The **Ask** persona is strictly **Read-Only**. Research findings must be handed back to the user or to a **work-planner** before action is taken.
4. **The Planning Hand-off**: The **Orchestrator** handles "coarse" planning and delegates "fine-grained" strategy to the **work-planner**.

## File Discovery

**Always consult the top-level [`toc.md`](../../../toc.md) file first** when looking for files, scripts, specifications, or documentation. The TOC provides a comprehensive index of all major project components with relative paths. Only resort to directory searching if the file cannot be found in the TOC.

In particular, when looking for source files, consult src/toc.md first, and then consult toc.md files that are named in that file first to track down files that are required for any given task first to find what is needed rather than reading a lot of files just to find a particular item. Resort to reading many files to find content only when consulting toc.md fails or isn't specific enough. Context window content is a precious resource, use it wisely.

## Archive Directory Handling

All file paths that contain an `archive` directory in them are to be **ignored** unless **explicitly directed** by the user to look at them. Archive directories contain historical or superseded files that are no longer part of the active project documentation.

## Core Organization Rules

- [`organization.md`](../organization.md) - Defines the directory structure and organization of the memory-bank, scripts, and resources directories
- [`end_of_task_activities.md`](../end_of_task_activities.md) - Defines standard activities to perform at the end of tasks (with user direction)

## Role-Specific Rules

### Planning and Architecture

- [`rules-work-planner/work-planner.md`](../rules-work-planner/work-planner.md) - work-planner mode rules for creating and managing plans
- [`rules-architect/architect.md`](../rules-architect/architect.md) - Architect mode rules for system design and architecture

### Development

- [`rules-code/code.md`](../rules-code/code.md) - Code mode rules for writing and modifying code
- [`rules-debug/debug.md`](../rules-debug/debug.md) - Debug mode rules for troubleshooting and diagnosing issues
- [`rules-code-reviewer/code-reviewer.md`](../rules-code-reviewer/code-reviewer.md) - Review mode rules for code review

### Writing and Documentation

- [`rules-technical-writer/technical-writer.md`](../rules-technical-writer/technical-writer.md) - technical-writer mode rules for writing documentation
- [`rules-technical-editor/technical-editor.md`](../rules-technical-editor/technical-editor.md) - technical-editor mode rules for reviewing documents

### Other Roles

- [`rules-ask/ask.md`](../rules-ask/ask.md) - Ask mode rules for explanations and answering questions
- [`rules-orchestrator/orchestrator.md`](../rules-orchestrator/orchestrator.md) - Orchestrator mode rules for coordinating complex multi-step projects
- [`rules-story-implementor/story-implementor.md`](../rules-story-implementor/story-implementor.md) - story-implementor mode rules for executing story files

## Design and Implementation Patterns

The Chronocone system follows some specific patterns in documentation, design and implemenation. Whenever performing any coding, architecture, planning or documentation activity **Always read the patterns table[`toc.md`](../../../memory-bank/specifications/patterns/toc.md) file first**

## Mandatory Reads

ALWAYS read the memory-bank/lessons-learned.md, memory-bank/concepts.md and memory-bank/terms.md before each task to make sure you understand the concepts and terms used in the system as well as what has been learned in the past so that you avoid repeating mistakes.

Above all else, follow these rules:

1. Don't assume. Don't hide confusion. Surface tradeoffs.
2. Minimum code that solves the problem. Nothing speculative.
3. Touch only what you must. Clean up only your own mess.
4. Define success critiera. Loop until verified.

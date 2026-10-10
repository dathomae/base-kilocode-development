---
trigger: always_on
---

# Lesson Compliance Rule

To prevent the repetition of past mistakes and respect established architectural patterns, you **must** adhere to the following workflow:

1.  **Initialize:** At the start of any task involving code changes, architectural decisions, or documentation updates about the same, use the `read_file` tool on `memory-bank/lessons-learned.md`.
2.  **Contextualize:** Explicitly state in your internal thought process/plan how the lessons in that file apply to the current task.
3.  **Constraint:** If a user request contradicts a "Lesson Learned," flag the conflict to the user before proceeding.
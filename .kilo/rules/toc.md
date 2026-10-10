# Rules Directory - Table of Contents

System-level rules that apply to **every** persona. These four documents are the only files under
`.kilo/rules/`; role definitions live in `.kilo/personas/` and are registered in `kilo.jsonc`.

**[system-rules.md](system-rules.md)**
: Persona selection guide, mandatory reads, the primary/subagent boundary rules, and file discovery. Injected into every session via the `instructions` config.

**[organization.md](organization.md)**
: Directory structure and organization of `memory-bank/`, `scripts/`, `resources/`, and the agent configuration.

**[end_of_task_activities.md](end_of_task_activities.md)**
: Standard end-of-story and end-of-task activities and who owns them.

**[toc.md](toc.md)**
: This index.

The shared agent resources — the persona registry, escalation policy, change control, artifact
standards, standards manifests, work context, report formats, and git workflow — live in
[`resources/agent/`](../../resources/agent/toc.md).

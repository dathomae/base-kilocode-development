# Agent Resources - Table of Contents

Shared resources for the KiloCode agent hierarchy: the persona registry, the escalation and
change-control policies, artifact standards, report formats, the work-context handoff schema, and the
git workflow the milestone coordinators follow.

**[persona-registry.md](persona-registry.md)**
: Authoritative registry of every agent persona: class, level, registration mode, dispatches, escalations, and standards.

**[escalation-policy.md](escalation-policy.md)**
: The three dispositions, the five escalation categories, retry limits, resolution routing, recording, and pause/hold semantics.

**[change-control.md](change-control.md)**
: The five-rung remediation ladder (R0–R4), decision authority per rung, the cascade order, and revision-vs-replacement rules.

**[artifact-standards.md](artifact-standards.md)**
: Per-level completeness criteria for requirements, architecture, design, plan, and story, and the cross-cutting gates.

**[standards-manifests.md](standards-manifests.md)**
: The explicit standards documents each critic loads, and the inputs each reviewer judges against.

**[work-context.md](work-context.md)**
: The structured handoff schema a coordinator passes to every subagent, replacing the old `story_context`.

**[task-kinds.json](task-kinds.json)**
: Resolved routing registry for core and specialization-defined task kinds, including implementors, reviewers, verifiers, critics, and contract schemas.

**[report-formats.md](report-formats.md)**
: Standard shapes for implementation results, review reports, critic reports, escalation reports, walkthrough additions, and milestone tracking entries.

**[git-workflow.md](git-workflow.md)**
: The `main` / append-only `agent-main` / `story/*` branch model, the merge-based main-sync PR, branch cleanup after a squash-merge, PR and CI rules, force-push constraints, and human gates.

**[agent-setup.md](agent-setup.md)**
: One-time setup of the GitHub App identity and branch rulesets for agent sessions.

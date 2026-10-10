A prompt driven development (PDD) script is a pseudo-executable script intended for "execution" by an
agentic tool. It is analogous to a shell script in the pre-LLM technology world.

A PDD script contains the following sections:

1) A description of the task or operation to be performed, and the persona that executes it.

2) A list of parameters required to execute the script. Parameters may be required or optional,
although all optional parameters MUST have a default value specified in the script. Parameters are
defined as a table with the following columns: Name, Type, Description, Required (Yes/No), Default
Value (if optional).

3) A list of operations that MUST be performed sequentially. The PDD script should stress that
constraint. Operations are numbered sequentially and use imperative language (e.g., "Read the file",
"Update the code").

4) A specification of the output items and output format. Some scripts are executed for their side
effects, so a script whose output is only a success/failure report is permissible; most scripts output
some information about the task (e.g., number of files updated).

5) Error Handling: if required parameters are not specified, the user is prompted. The operations WILL
NOT execute until all required parameters are specified. During execution, if an error is encountered,
execution pauses and the user is asked whether to continue or terminate.

## Coordinator scripts

A coordinator PDD script additionally describes how the coordinator dispatches its subagents. Follow
these conventions.

- **Dispatch mechanism.** A coordinator dispatches a subagent with the Task tool. The agent name is
  the persona's registered name (see `resources/agent/persona-registry.md`). A subagent may itself
  dispatch subagents; top-level (primary) agents may not dispatch other top-level agents.
- **Work context.** Every dispatch passes a `work_context` block
  (see `resources/agent/work-context.md`). The subagent returns it unchanged with its result.
- **Parallelism.** Independent subagents are dispatched together in a single message so they run
  concurrently; dependent ones are dispatched sequentially. Do not describe a control flow the runtime
  does not have (for example, "the subagent creates a new task back to the coordinator") — the
  coordinator receives each subagent's result as a normal tool result.
- **Worktrees.** Parallel task execution uses one git worktree per task. Refer to worktree paths through
  the `worktrees/` symlink, never `.kilo/worktrees/`.
- **Dispositions and escalation.** Each finding gets one of the three dispositions — fix on the spot,
  record, or block and escalate — per `resources/agent/escalation-policy.md`. The coordinator owns the
  disposition; reviewers and critics may propose one. A coordinator resolves findings internal to its
  artifact and escalates findings that would change an upstream artifact.
- **Remediation.** An issue requiring changes above the story level follows the ladder in
  `resources/agent/change-control.md`, executed through `scripts/change-control.pdd.script.md`.
- **Completion gates.** Before declaring an artifact complete, apply the relevant gates in
  `resources/agent/artifact-standards.md` (for plans and stories, an empty `Open Questions` section and
  no figure-it-out items).
- **Reports.** Exchange results in the shapes defined by `resources/agent/report-formats.md`.

See `scripts/story-implementation.pdd.script.md` and `scripts/milestone-generation.pdd.script.md` for worked
examples, and `scripts/create-new-pdd-script.pdd.script.md` for the authoring procedure.

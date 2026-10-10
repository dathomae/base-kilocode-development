# Work Context

The work context is the structured handoff a coordinator passes to every subagent it dispatches. It
generalizes the old `story_context` so the same schema serves coding, writing, review, and critic
tasks at every level. Coordinators must include it; subagents must return it unchanged with their
result so the coordinator can correlate results without re-deriving context.

Related: [`report-formats.md`](report-formats.md), [`persona-registry.md`](persona-registry.md),
[`escalation-policy.md`](escalation-policy.md), [`git-workflow.md`](git-workflow.md).

## Schema

```yaml
work_context:
  work_id: <stable id, e.g. StoryNNN-Task3 or PlanFoo-R2>
  artifact: code | document | plan | story | spec | milestone
  level: spec | plan | milestone | story | task
  assignment: <one-sentence statement of what the subagent must produce>
  source_root: <directory where build/test commands run>
  worktree_path: <path to the worktree, using the worktrees/ symlink, never .kilo/worktrees/>
  base_branch: <branch the work is based on, e.g. agent-main>
  paths:
    primary: [<files/packages this task is expected to touch>]
    read_only: [<files to read for context>]
  standards: [<standards documents the subagent must load; extends the default manifest>]
  upstream_decisions: [<settled decisions the work must implement, by reference>]
  acceptance_criteria: [<verifiable criteria for this work>]
  constraints: [<algorithmic, memory, package, or procedure constraints>]
  story_context: <unchanged passthrough when the work is part of a story>
  expected_return: <what the subagent must return; see report-formats.md>
  blocked_conditions: [<conditions under which the subagent must return Blocked — needs decision>]
```

## Rules

1. **Required fields.** `work_id`, `artifact`, `level`, `assignment`, `paths.primary`,
   `upstream_decisions`, `acceptance_criteria`, `expected_return`. A subagent that is missing any of
   these must terminate and report the omission rather than guess.
2. **Return unchanged.** A subagent returns the `work_context` block unchanged alongside its result.
   The coordinator uses `work_id` to correlate.
3. **Worktree paths.** Always express `worktree_path` through the `worktrees/` symlink, never
   `.kilo/worktrees/`. Dot-directory paths trigger a tool-permission failure that aborts the
   subagent mid-task. The same applies to any path mentioned in task instructions.
4. **Standards.** `standards` extends, never replaces, the critic's default manifest (see
   [`standards-manifests.md`](standards-manifests.md)). Reviewers use `upstream_decisions` and
   `acceptance_criteria` as their intent baseline.
5. **Upstream decisions by reference.** Cite the plan/story/spec section, not a paraphrase, so the
   subagent reads the authoritative text.
6. **Blocked conditions.** State up front what would require a decision the subagent is not authorized
   to make, so the implementor returns `Blocked — needs decision` instead of improvising
   (categories 1/5 of [`escalation-policy.md`](escalation-policy.md)).
7. **No model or persona routing in the context.** Model selection is a session/Agent Manager concern;
   the coordinator selects the persona (subagent type) and passes context only.
8. **Size hints.** A coordinator may set `sizing: small | medium | large` to record its intended
   implementor tier; it is advisory and the coordinator retains the ladder decision.

## Example (coding task inside a story)

```yaml
work_context:
  work_id: Story042-Task2
  artifact: code
  level: task
  assignment: Implement notification delivery retries per the story task.
  source_root: src
  worktree_path: worktrees/story042-task-2
  base_branch: agent-main
  paths:
    primary: [notifications/delivery.py, tests/notifications/test_delivery.py]
    read_only: [notifications/queue.py, memory-bank/specifications/architecture/README.md]
  standards: [resources/code-conventions.md, specifications/patterns/domain-layer-patterns.md]
  upstream_decisions: [plan §Settled decisions D2, story Task 2 acceptance criteria]
  acceptance_criteria:
    - Transient failures are retried up to the configured limit.
    - Permanent failures return a typed delivery error.
  constraints: [do not change the public message schema, preserve idempotency]
  story_context: {story_id: Story042, plan: notification-delivery-plan.md}
  expected_return: implementation summary, files, test results, Blocked status if applicable
  blocked_conditions: [a retry-policy decision beyond the plan's settled D2]
```

## Example (critic pass at story end)

```yaml
work_context:
  work_id: Story042-Critic
  artifact: code
  level: story
  assignment: Critique the Story042 changes and walkthrough against the code standards.
  source_root: src
  worktree_path: worktrees/story042
  base_branch: story/520
  paths:
    primary: [notifications/]
    read_only: [memory-bank/planning/stories/active-stories/Story042_notification-delivery.md]
  standards: []
  upstream_decisions: [plan §Settled decisions]
  acceptance_criteria: [all story tasks complete and verified]
  expected_return: critic report per report-formats.md
  blocked_conditions: []
```

## Example (flow verification at milestone close)

```yaml
work_context:
  work_id: M3-FG1.F1.1
  artifact: milestone
  level: milestone
  assignment: Verify flow FG1/F1.1 against the milestone's flow definition and record evidence.
  source_root: tests/flows
  worktree_path: worktrees/milestone-3
  base_branch: agent-main
  paths:
    primary: [tests/flows/FG1_F1_1_notification_delivery]
    read_only: [memory-bank/planning/milestones/active-milestones/Milestone_M3_....md]
  standards: []
  upstream_decisions: [milestone §4 Flow definitions FG1/F1.1; REQ-012]
  acceptance_criteria:
    - The flow's expected positive and negative outcomes are observed.
    - Evidence records the exact command and the revision.
  expected_return: flow-verification result per report-formats.md
  blocked_conditions: [the flow definition is ambiguous, or the environment cannot run the flow]
```

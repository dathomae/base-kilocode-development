# End of Task Activities

These are the standard activities at the end of a story or task. The **story-completion** activities
(steps 1–4) are performed by `story-implementation-coordinator` as soon as it believes the story is
complete — **before** presenting the result to the user — so the user reviews a finished state.
Arc-level tracking (step 5) is owned by the milestone coordinators. Commits and pushes (step 6) follow
[`resources/agent/git-workflow.md`](../../resources/agent/git-workflow.md), not a blanket rule.

1. **Execute the tests.** If any fail for any reason — even one unrelated to the task just completed —
   STOP and ask the user what to do next.
2. **Update the artifact.** Update the story file's task statuses and, if working from a plan, any plan
   status that changed. Update the story's status in
   `memory-bank/planning/stories/active-stories/toc.md`.
3. **If the story is complete** (all tasks `Completed`), do all of this before asking the user for
   feedback:
   - Move the story file from `memory-bank/planning/stories/active-stories/` to
     `memory-bank/planning/stories/finished-stories/`.
   - Remove the story entry from `active-stories/toc.md` and add it to `finished-stories/toc.md` with
     status `_Done_`; update the link so it is relative to the new `toc.md` (just the filename, since
     the toc and the story now share a directory).
   - Write a **walkthrough** and place it in
     `memory-bank/planning/stories/finished-stories/walkthroughs/`, and add it to that directory's
     `toc.md` (see [Walkthroughs](#walkthroughs)).
   - Update `memory-bank/planning/tracking/system-definition-tracking-status.md` to reflect the
     story's completion (see [Tracking status](#tracking-status)).
4. **If the story is cancelled** (abandoned or superseded rather than completed), move it from
   `active-stories/` to `canceled-stories/`, update the two `toc.md` files, and update
   `system-definition-tracking-status.md`. A cancelled story does **not** get a walkthrough.
   Cancellation is authorized through [`resources/agent/change-control.md`](../../resources/agent/change-control.md)
   and always requires human authorization.
5. **Update the milestone file.** `milestone-generation-coordinator` (or `milestone-implementation-coordinator`
   during execution) updates the milestone file: story status, wave progress, divergences,
   escalation items, and the remediation log.
6. **Commits and pushes** follow [`resources/agent/git-workflow.md`](../../resources/agent/git-workflow.md):
   an agent may commit and push `story/*`, and open pull requests into `agent-main` (story PRs, the
   main-sync PR, and the coordinator's `chore/<milestone>-tracking` PR). It must **never** push `main` and
   must **never** merge `agent-main` into `main` — that merge is always the human's action. This replaces
   the former blanket "never commit or push without the user's explicit direction" rule.

## Milestone close ordering

At **milestone close**, `milestone-implementation-coordinator` completes the **close evaluation and then
the milestone finalization before the merge request**: the pre-merge Merge record (the `main` baseline and
the **pre-finalization** `agent-main` head SHA — the base of the tracking PR that carries the
finalization), `Stage: complete`/`Status: Done`, the finished-tree move, the tocs and
`system-definition-tracking-status.md` updates, and the close record — all landed on `agent-main` through
a dedicated `chore/<milestone>-tracking` PR (App-authored; squash-merged). The head the human merges is
that tracking PR's **merge commit** and is informational only. The human's single
`agent-main → main` merge is the **last action**; the coordinator performs **no post-merge step on
`main`**, does **not** monitor the merge PR or the post-merge CI, and the human reports any merge or
post-merge difficulty ("no news is good news"). `agent-main` is synced at the next milestone's wave start.
See [`resources/agent/git-workflow.md`](../../resources/agent/git-workflow.md).

## Walkthroughs

A walkthrough is a completion record for a finished story. Generate one for **every** completed story,
name it with the story file's stem plus `.walkthrough.md`
(`Story<NNN>_<short-purpose>.walkthrough.md`), save it in
`memory-bank/planning/stories/finished-stories/walkthroughs/`, and list it in that directory's
`toc.md`. Use the template at
[`resources/templates/walkthrough.template.md`](../../resources/templates/walkthrough.template.md).
At minimum it records:

- what the story set out to do, and what actually changed (files created, modified, deleted);
- any deviation from the story and why;
- verification: the tests run and their results;
- decisions made during implementation and the authority for them;
- divergences recorded rather than acted on, and any remediation applied;
- lessons learned / post-mortem — candidate additions to `memory-bank/lessons-learned.md`.

The `code-critic`/`write-critic` read the walkthrough for anomalies: unexplained deviations, silent
scope changes, skipped verification, and decisions made without authority.

## Tracking status

`memory-bank/planning/tracking/system-definition-tracking-status.md` records, per subsystem, whether
its implementation story set is `Planned` (`Done` / `Partial` / `TODO`), how far it has been
`Implemented`, the outstanding `Remaining stories` in implementation order, the plan file(s), and the
open issues. See the column definitions in that file. Update the affected subsystem's row whenever a
story is completed, cancelled, or created:

- **completed:** remove the story from `Remaining stories`; advance `Implemented` (to `Done` when the
  whole planned set is done); advance `Planned` (`Partial` → `Done`) if the full story set now exists.
- **cancelled:** remove the story from `Remaining stories` (or from the planned set) and correct
  `Planned` / `Implemented` so the row stays truthful.
- **created:** add the story to `Remaining stories` in the order it must be implemented, and set
  `Planned` to `Partial` (or `Done` when the full set now exists).

The §4 Resolution log in the same file records human resolutions to escalations. See
[`resources/agent/report-formats.md`](../../resources/agent/report-formats.md).

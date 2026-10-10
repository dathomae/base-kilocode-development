# Walkthrough Template

Use this template for the **completion walkthrough** of a finished story. Save the filled-in
walkthrough in
[`memory-bank/planning/stories/finished-stories/walkthroughs/`](../../memory-bank/planning/stories/finished-stories/walkthroughs/)
under the story file's stem plus `.walkthrough.md`
(`Story<NNN>_<short-purpose>.walkthrough.md`), and list it in that directory's `toc.md`.

`story-implementation-coordinator` writes the walkthrough when it believes the story is complete,
**before** the user is asked for feedback. Keep it factual and concise; it is the durable record of
what happened, not a restatement of the story. `code-critic`/`write-critic` read it specifically for
anomalies: unexplained deviations, silent scope changes, skipped verification, and decisions made
without authority.

The links in the skeleton below are relative to the walkthrough's destination directory
(`finished-stories/walkthroughs/`), not to this template file.

---

# Story<NNN>: <Story Title> — Completion Walkthrough

**Story:** [`Story<NNN>_<short-purpose>.md`](../Story<NNN>_<short-purpose>.md)
**Status:** Done
**Date:** <YYYY-MM-DD>
**Revision:** <n> (if the story was revised under change control)

## 1. Summary

<What the story set out to do, in one short paragraph, and the outcome.>

## 2. What changed

<Files created, modified, or deleted, with a one-line note on each. A table is usually clearest.>

| File | Change |
|---|---|
| `path/to/file` | <created / modified / deleted — what and why> |

## 3. Deviations from the story

<Anything done differently from the story's plan, and why. "None." is a valid answer.>

## 4. Decisions made during implementation

<Decisions made at the coordinator level and the authority for each (story text, plan decision,
escalation resolution). If none, say so.>

## 5. Verification

<Tests run and their results; manual verification performed.>

- `<command>` — <result>

## 6. Review and critic results

<A brief record of the review and critic passes: what they found, what was fixed, and what was
recorded or escalated. Reference the reports.>

## 7. Divergences

<Non-blocking misalignments recorded rather than acted on, with the covering story if any.>

## 8. Remediations

<Any change-control R0/R1 revision applied during the story, with the triggering issue.>

## 9. Post-mortem / lessons learned

<What went well, what went wrong, and what should be added to
[`lessons-learned.md`](../../../../lessons-learned.md). Candidates should be concrete and reusable.>

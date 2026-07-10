# Brief Template

When creating a new `brief.md` for a sub-package, copy the **headings below**
into the new file and write content under each. Do **not** copy the gloss
text or this preamble — write package-specific content from scratch.

Required sections (must always appear, even when there is nothing to say):

- `# <Package Name>` — short, descriptive title.
- `## Responsibilities` — what this package owns; the work that lives here.
- `## Package Boundaries` — what does **not** live here, and which neighbouring
  package owns it instead. Make the negative space explicit.
- `## When you work here` — rules, invariants, and patterns a contributor
  must follow when editing code in this package.

If a required section has no content, include the heading with an
explicit-empty stub on the next line: `_(none — <reason>)_`. For example,
a leaf utility package with no internal-package dependencies might write
`_(none — this package depends only on the standard library)_` under
`## Package Boundaries`. The stub tells an unfocused reader "intentionally
empty" rather than "section forgotten."

Optional sections (include when warranted; omit when empty):

- `## Architecture` — internal structure if non-obvious.
- `## References` — links to specs, patterns, or sibling briefs. Mark
  must-reads as **Critical**.
- `## Lessons Learned` — durable constraints surfaced by past work.

Keep prose tight.

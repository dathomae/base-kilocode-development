# AGENTS.md — Bootstrap Instructions for Codex

## Read Brief Documentation

Before responding to any request:

1. Identify your working directory (where you'll make changes):
   - **Code changes** belong under `src/...`. Traverse the `src/` tree and pick the
     most specific package whose brief's "Responsibilities" / "Package Boundaries" /
     "When you work here" sections fit the task. If a task plausibly spans two
     packages, ask before guessing.
   - **Pure documentation** (specs, patterns, lessons, decision records) belongs
     under `memory-bank/...`. Do not put narrative documentation under `src/`,
     and do not put code under `memory-bank/`.
2. Read all `brief.md` files from project root → to your working directory (in order)
3. Follow external context references marked as "Critical" in those briefs
4. Keep a list of non-critical references and use them at your discretion when the task calls for it

This ensures you understand project goals, constraints, and module-specific patterns without
unnecessary document cascade.

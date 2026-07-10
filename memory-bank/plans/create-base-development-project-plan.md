# Plan: Create Generic Base KiloCode Development Project

## Goal

Transform this repository from a chronocone-specific environment into a generic, reusable base configuration for KiloCode development projects. Strip all chronocone-specific references from rules, scripts, and resources while preserving the methodology framework. Create a new `initialize-project.pdd.script.md` that bootstraps a new project with language-specific scaffolding (including Docker devcontainer setup), a full memory bank, and test-runner scripts for each supported language.

## Steps

### Step 1: Make organization.md generic — Not Started

The `organization.md` currently references `requirements.md` and `design.md` as top-level memory-bank files (items 7 and 8). The chronocone project uses a richer structure where requirements, architecture, and design live under `memory-bank/specifications/`. Update `organization.md` to reflect:

- `memory-bank/specifications/requirements/` as the requirements location
- `memory-bank/specifications/architecture/` as the architecture location
- `memory-bank/specifications/design/` as the design location
- `memory-bank/stories/` with `toc.md` tracking, `finished-stories/`, and `canceled-stories/` subdirectories
- `memory-bank/epics/` for epic stories with `toc.md`
- `memory-bank/plans/` for shorter plan files

No changes needed to `end_of_task_activities.md` — it already references the correct patterns.

### Step 2: Strip chronocone references from .kilocode/rules — Not Started

Three files need fixing:

1. **`system-rules/system-rules.md` line 79**: Replace chronocone-specific sentence with generic guidance.
2. **`rules-mock/mock.md` line 8 and 28**: Replace "Chronocone features (e.g., bitemporal visualizations)" with "project features" and "50,000-plan processing" with "large-scale data processing".
3. **`rules-work-planner/work-planner.md` line 45**: Change "Logical TDD Lifecycle (BTDD)" to "Logical TDD Lifecycle" — drop the BTDD acronym (Bitemporal TDD).

Also update **`rules-code/code.md`** "Related Scripts" section: remove references to the chronocone-specific scripts being deleted (start-core-service, shutdown-servers), replace with the renamed test-runner scripts.

### Step 3: Remove chronocone-specific scripts — Not Started

Delete these 7 scripts (no archival — just remove them from git):

- `scripts/convert-example-to-xlsx.pdd.script.md`
- `scripts/create-new-cpl-example.pdd.script.md`
- `scripts/generate-models.sh`
- `scripts/initialize-database.pdd.script.md`
- `scripts/start-core-service.pdd.script.md`
- `scripts/shutdown-servers.pdd.script.md`
- `scripts/add-story-to-github-backlog-item.pdd.script.md`

### Step 4: Remove ADR script and references — Not Started

Delete `scripts/add-new-adr.pdd.script.md`. The project uses EROs (Epic Restraining Orders) as the primary decision-tracking mechanism via `scripts/add-new-ero.pdd.script.md` and the `resources/templates/ero.md` template.

### Step 5: Rename test runner script and create language-specific variants — Not Started

Rename `scripts/run-pytest-with-coverage.pdd.script.md` to `scripts/run-go-tests.pdd.script.md`. The content already targets Go tests, so only the filename changes (plus any internal header/label updates).

Create language-specific test-runner PDD scripts for the remaining supported languages:

- `scripts/run-python-tests.pdd.script.md` — runs pytest with coverage
- `scripts/run-java-tests.pdd.script.md` — runs Gradle test task
- `scripts/run-kotlin-tests.pdd.script.md` — runs Gradle test task (Kotlin)
- `scripts/run-rust-tests.pdd.script.md` — runs `cargo test`
- `scripts/run-c-tests.pdd.script.md` — runs `make test`

### Step 6: Genericize remaining scripts — Not Started

Two scripts need chronocone/CPL references removed:

- **`scripts/add-new-story.pdd.script.md`**: Remove CPL-specific path conventions (chronocone_planning_language references in lines 5, 53, 69); replace with generic `source_root` guidance.
- **`scripts/story-implementor.pdd.script.md`**: Remove all CPL-specific references (lines 3, 126, 187, 189, 322, 411); make `source_root` detection language-agnostic rather than hardcoded to CPL.

### Step 7: Genericize resources — Not Started

**Remove:**
- `resources/discussions/system_shape_discussion.md` — entirely chronocone-specific

**Update:**
- `resources/templates/stories.md`: Remove CPL-specific path conventions (lines 82-91), replace chronocone_planning_language examples with generic `<source_root>` placeholders, replace "Python 3.11+ with FastAPI" with "the project's chosen language and framework" in the example.

**Create:**
- `resources/toc.md` — A global resources index (exists in chronocone but is missing here)

### Step 8: Enhance initialize-per-project-memory-bank.pdd.script.md — Not Started

Update this script to create the full chronocone-style structure:

- `memory-bank/specifications/requirements/` with `README.md` and `toc.md`
- `memory-bank/specifications/architecture/` with `README.md` and `toc.md`
- `memory-bank/specifications/design/` with `README.md` and `toc.md`
- `memory-bank/stories/` with `toc.md`, `finished-stories/`, and `canceled-stories/` subdirectories
- `memory-bank/epics/` with `toc.md`
- `memory-bank/plans/` directory
- `memory-bank/toc.md` as a master index
- `brief.md` with a placeholder asking the user to describe the project
- `requirements.md` with "TBD - requirements not yet defined" placeholder

### Step 9: Create initialize-project.pdd.script.md — Not Started

Create a new PDD script that bootstraps a complete new project.

**Parameters:**

| Name | Type | Description | Required | Default |
|---|---|---|---|---|
| project_name | string | The name of the project | Yes | N/A |
| language | string | One of: go, python, java, kotlin, rust, c | Yes | N/A |
| description | string | Brief project description | Yes | N/A |
| requirements | string | Requirements text or path to file | No | (creates TBD placeholder) |
| project_path | string | Path where project is created | No | CWD |

**Operations:**

1. Validate all required parameters.
2. Announce the chosen ecosystem defaults to the user for confirmation:
   - Go: module path `github.com/<project_name>`, standard Go layout
   - Python: `pyproject.toml` with setuptools, pytest
   - Java: Gradle with Kotlin DSL, JUnit 5
   - Kotlin: Gradle with Kotlin DSL, Kotest
   - Rust: Cargo with default template
   - C: Simple Makefile, C11 standard
3. Create the language-specific source directory structure:
   - **Go**: `src/` with `go.mod`, `<package>/main.go` stub, `<package>/main_test.go` stub
   - **Python**: `src/<package>/__init__.py`, `pyproject.toml`, `tests/__init__.py`, `tests/test_main.py` stub
   - **Java**: `src/main/java/<package>/Main.java` stub, `src/test/java/<package>/MainTest.java` stub, `build.gradle.kts`, `settings.gradle.kts`
   - **Kotlin**: `src/main/kotlin/<package>/Main.kt` stub, `src/test/kotlin/<package>/MainTest.kt` stub, `build.gradle.kts`, `settings.gradle.kts`
   - **Rust**: `src/main.rs` stub, `Cargo.toml`, `tests/integration_test.rs` stub
   - **C**: `src/main.c` stub, `include/`, `Makefile` (with build/test/clean targets), `tests/test_main.c` stub
4. Create the `.devcontainer/` directory with Docker files genericized from chronocone's pattern:
   - `.devcontainer/.dockerignore` (generic, exclude `.git`, `__pycache__`, etc.)
   - `.devcontainer/Dockerfile` — Ubuntu 24.04 base with language-specific toolchain:
     - Go: Go toolchain + delve debugger
     - Python: Python 3.12, uv package manager, build tools
     - Java/Kotlin: OpenJDK 21, Gradle wrapper
     - Rust: Rust toolchain via rustup, cargo
     - C: GCC, GDB, make, CMake
     - Common across all: git, curl, wget, gh CLI, Node.js 20 + pnpm, vim, emacs-nox, jq, network tools
   - `.devcontainer/devcontainer.json` — genericized name, workspace folder config, VS Code extensions based on language, post-create script hook
   - `.devcontainer/docker-compose.yml` — single app service with workspace mount, env vars (no chronocone-specific db service by default; can be added per-project)
   - `.devcontainer/post-create.sh` — bootstrap script that installs language toolchain and dependencies
5. Invoke `initialize-per-project-memory-bank.pdd.script.md` to create the full memory bank.
6. Populate `memory-bank/brief.md` with the project description.
7. Populate `memory-bank/specifications/requirements/` with TBD placeholder or provided requirements.
8. Report summary of all files created.

**Language-specific .devcontainer details:**

Each language gets a tailored `Dockerfile` and `devcontainer.json`:

- **Go**: Extensions: `golang.go`, `gururagavendra.local-pr-review`. Post-create: downloads latest Go, `go mod download`, installs delve.
- **Python**: Extensions: `ms-python.python`, `ms-python.debugpy`, `gururagavendra.local-pr-review`. Post-create: `uv sync` or `pip install -e ".[dev]"`.
- **Java**: Extensions: `vscjava.vscode-java-pack`, `gururagavendra.local-pr-review`. Post-create: `./gradlew build`.
- **Kotlin**: Extensions: `fwcd.kotlin`, `vscjava.vscode-java-pack`, `gururagavendra.local-pr-review`. Post-create: `./gradlew build`.
- **Rust**: Extensions: `rust-lang.rust-analyzer`, `gururagavendra.local-pr-review`. Post-create: `cargo build`.
- **C**: Extensions: `ms-vscode.cpptools`, `gururagavendra.local-pr-review`. Post-create: `make`.

`Dockerfile` pattern (genericized from chronocone) for each language — always Ubuntu 24.04 base with curl, wget, git, build-essential, vim, emacs-nox, jq, iputils-ping, dnsutils, Node.js 20 + pnpm, gh CLI, plus the language toolchain.

### Step 10: Clean up kilo.jsonc references — Not Started

- The `kilo.jsonc` `code` agent prompt includes fixed script references — these should reference the rules via `.kilocode/rules/rules-code/code.md`, which is already handled by the prompt. No changes needed to `kilo.jsonc` itself.
- The `rules-code/code.md` Related Scripts section is updated in Step 2.

### Step 11: Execute the memory bank initialization on this project — Not Started

Run the enhanced `initialize-per-project-memory-bank.pdd.script.md` against this project's root to populate the `memory-bank/` directory with the standard structure. This validates the script works correctly.

### Step 12: Test the project initialization script — Not Started

Run `initialize-project.pdd.script.md` against `/tmp/test-project` for Go to validate end-to-end: source files, devcontainer, memory bank, test runner all created correctly.

---

## Notes

- The chronocone project has 18 scripts. After Step 3 (remove 7) + Step 4 (remove 1 ADR) = 10 remain, then Step 5 adds 5 new test-runner scripts = 15 total. Renaming `run-pytest-with-coverage` to `run-go-tests` does not change the count.
- The `rules-code/code.md` related-scripts section will be updated in Step 2 to list all 6 test-runner scripts generically, or better yet, point to the scripts directory where the test-runner scripts live.
- EROs are the primary decision-tracking mechanism. ADR scripts and references are removed. Design decisions are captured via `add-design-decision.pdd.script.md` for co-located design decisions and `add-new-ero.pdd.script.md` for standalone architectural decisions.
- Docker devcontainer setup in Step 9 is language-specific but follows the common pattern from chronocone. The docker-compose.yml is simplified — no chronocone-specific database service. Users add services (db, cache, etc.) as their project needs them.

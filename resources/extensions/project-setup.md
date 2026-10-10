# Project Setup by Composition Type

Use this guide to bootstrap a consuming repository from released, version-pinned packs. The four
supported shapes differ only in their pack list; they share the same composition, customization, and
protected-branch procedure.

| Project shape | Packs | Example |
|---|---|---|
| Core-only software | core | A focused software project that needs no domain specialization |
| Hardware/manufacturing | core + hardware/manufacturing | Curta |
| Complex system | core + complex systems | Provenance Reasoning |
| Combined | core + both specializations | Instrumented mason bee house |

The hardware/manufacturing and complex-systems packs are independent siblings. Each pins a tested
core version; neither may depend on the other.

## 1. Prepare the repository and pack inputs

1. Create a normal GitHub repository with `main` as its default branch, then clone it over HTTPS.
2. As the human owner, create the initial commit if the repository is empty.
3. Obtain the exact released core and specialization versions as trusted local checkouts or unpacked
   release artifacts. Keep them outside the consuming project. Do not use a moving `main` checkout as
   a release input.
4. Read each pack's `kilocode-pack.json` and record its exact `name` and `version`. A specialization's
   `requires` entry must match the selected core version.
5. Copy the core pack's `resources/templates/kilocode-composition.template.json` to
   `kilocode-composition.json` at the consuming project root.

The examples below assume sibling directories. Replace every angle-bracket placeholder with the
literal value from the corresponding pack manifest.

## 2. Choose exactly one composition

### Core-only software

Use this for software that needs the general planning and implementation framework but neither domain
specialization.

```json
{
  "schema_version": 1,
  "output": ".",
  "packs": [
    {
      "path": "../base-kilocode-development",
      "name": "base-kilocode-development",
      "version": "<core-version>"
    }
  ]
}
```

### Hardware/manufacturing

Use this for mechanical design, PCB/electronics, or embedded-C work, including projects that use only
a subset of those capabilities.

```json
{
  "schema_version": 1,
  "output": ".",
  "packs": [
    {
      "path": "../base-kilocode-development",
      "name": "base-kilocode-development",
      "version": "<core-version>"
    },
    {
      "path": "../<hardware-pack-directory>",
      "name": "<hardware-pack-name>",
      "version": "<hardware-pack-version>"
    }
  ]
}
```

### Complex system

Use this for Go-first systems that may reuse components such as authentication, authorization,
secrets, serialization/database setup, development containers, or optional infrastructure services.

```json
{
  "schema_version": 1,
  "output": ".",
  "packs": [
    {
      "path": "../base-kilocode-development",
      "name": "base-kilocode-development",
      "version": "<core-version>"
    },
    {
      "path": "../<complex-systems-pack-directory>",
      "name": "<complex-systems-pack-name>",
      "version": "<complex-systems-pack-version>"
    }
  ]
}
```

### Combined hardware and complex system

Use this when the same product contains physical design, electronics or firmware, and a software
system. The sibling specializations both depend on the core and are composed together.

```json
{
  "schema_version": 1,
  "output": ".",
  "packs": [
    {
      "path": "../base-kilocode-development",
      "name": "base-kilocode-development",
      "version": "<core-version>"
    },
    {
      "path": "../<hardware-pack-directory>",
      "name": "<hardware-pack-name>",
      "version": "<hardware-pack-version>"
    },
    {
      "path": "../<complex-systems-pack-directory>",
      "name": "<complex-systems-pack-name>",
      "version": "<complex-systems-pack-version>"
    }
  ]
}
```

## 3. Compose and review the project

Run the composer from the core pack for the first installation, because the project does not contain
the managed copy of the script yet:

```bash
python ../base-kilocode-development/scripts/compose-kilocode-packs.py \
  --composition kilocode-composition.json \
  --check
python ../base-kilocode-development/scripts/compose-kilocode-packs.py \
  --composition kilocode-composition.json
```

Then:

1. Review every reported action and the generated `.kilocode-pack.lock.json`.
2. Confirm that `kilo.jsonc` contains the expected core and specialization agents and that
   `resources/agent/task-kinds.json` contains only the expected task kinds.
3. Commit `kilocode-composition.json`, `.kilocode-pack.lock.json`, and the composed files as the human
   bootstrap commit on `main`.
4. For later updates, acquire the newly pinned pack versions, edit `kilocode-composition.json`, run
   `python scripts/compose-kilocode-packs.py --composition kilocode-composition.json --check`, review
   the plan, then run it without `--check`.

Never edit a managed file to customize the project; the next update will correctly report that as a
conflict. Put project-specific changes in seed files or new project-owned files.

## 4. Customize the project-owned seed files

1. Fill in `memory-bank/brief.md`, `terms.md`, `concepts.md`, `requirements.md`, and `design.md`.
2. Add the project's architecture, design, requirements, and patterns under
   `memory-bank/specifications/`.
3. Set project-specific code conventions and review guidance under `resources/`.
4. Review the default models and permissions in `kilo.jsonc` without changing pack-owned agent
   entries.
5. Define one pre-PR verification command that matches CI and record it where
   `resources/agent/git-workflow.md` requires.
6. Add specialization-specific equipment, service, component, or environment profiles only in the
   paths defined by the selected specialization packs.

## 5. Enable protected milestone implementation

This section is mandatory before an agent implements a milestone, for every composition type.
Composition alone installs the workflow files; it does not apply repository rulesets or configure an
agent identity.

1. Finish and push the human bootstrap commit to `main`.
2. Follow `AGENT_REPOSITORY_SETUP.md` to create or reuse the GitHub App, store its credentials outside
   the repository, apply the `main` and `agent-main` rulesets, and create `agent-main` from `main`.
3. Configure the clone to use the App's HTTPS credential helper and bot commit identity.
4. Run `bash scripts/install-git-guards.sh` to install the local defense-in-depth guard.
5. Run `python scripts/setup-agent-github.py --check --repo <owner>/<repo>` and resolve every failure.
6. Fetch and switch to `agent-main`. Start milestone coordination there, never on `main`.
7. Confirm the operating path before real work: a disposable branch created from `agent-main` can be
   pushed by the App; a direct push to `main` is rejected; and only the human performs the final
   `agent-main` to `main` merge.

During a milestone, agents may create as many branches and worktrees from `agent-main` as the approved
story schedule requires. Story and tracking changes return through PRs to the append-only
`agent-main`; no agent-side confirmation, override, or exception permits work directly on `main`.

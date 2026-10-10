# Agent GitHub Setup

One-time setup so agent sessions act as a **GitHub App** (a distinct non-user actor) and branch
protection gates `main`. After this, agent PRs require your approval while your own PRs do not, and
the milestone coordinators can push story branches, open PRs, and merge into `agent-main`.

Related: [`git-workflow.md`](git-workflow.md) (the branch model and ruleset settings),
[`persona-registry.md`](persona-registry.md).

> **The GitHub App is required** for this protected-branch workflow's commits on
> `agent-main`/`story/*` and for agent-issued PRs; there is no human-credential mode. If you lack admin rights to create the app and apply the
> rulesets, the agent workflow cannot be enabled — do not substitute your own credentials.

> **`--check` semantics (revised 2026-10-04, R3, human-confirmed).** The one-command check passes when
> `main` is **reachable from** `agent-main` (`agent-main` at or ahead of `main`, no drift); it does **not**
> require the two branches to be identical. Under the append-only R3-3 model `agent-main` is normally ahead
> of `main` once a story has merged. On drift it fails, naming the drift and pointing to the wave-start
> main-sync PR; setup never rewrites an existing `agent-main` ref to clear drift.

## Prerequisites

- `gh` authenticated with **admin** on the repository (`gh auth login`).
- `python3` and `openssl` on `PATH`.
- A browser, for the two steps GitHub does not allow scripts to perform.

> **Org-owned Apps.** To share one agent identity across a team, create and own the App under the
> **organization** (`--org`, or the organization's Developer settings), not a personal account. The
> organization must **allow and install** the App before it can be installed on a repository — a
> repository admin alone is insufficient: an org owner may need to enable or approve App installation
> in the organization's settings first.

> **No "previous App was deleted" text.** This runbook makes no "previous App was deleted" assumption:
> the existing-App path reuses the App and re-stores its key.

## Automated setup

```bash
python3 scripts/setup-agent-github.py --repo <owner>/<repo>
```

The script:

1. **Creates the GitHub App** via the manifest flow. It serves a local page that submits the app
   manifest (permissions `contents: write`, `pull_requests: write`, `metadata: read`, and
   `workflows: write` with `--workflows`), opens it in your browser, and captures the callback. You
   click **Create GitHub App** once; the script then exchanges the code for the App ID and private
   key automatically.
2. **Stores the private key outside the repository** at `<config-dir>/github-app.pem` (`0600`), and
   writes `<config-dir>/agent-github.env` (`0600`) with `GITHUB_APP_ID`,
   `GITHUB_APP_PRIVATE_KEY`, and `GITHUB_APP_INSTALLATION_ID`. The default config dir is
   `~/.config/kilocode-agent`.
3. **Installs the app on the repository.** It opens the install page; you click **Install**; the
   script discovers the installation id and records it.
4. **Verifies** by minting a test installation token.
5. **Sets the repository variables** the commit-authorship check compares against
    (`AGENT_BOT_EMAIL`, `AGENT_BOT_LOGIN`).
6. **Applies the branch rulesets** via `scripts/apply-github-rulesets.sh`. The App is **not** a bypass
   actor (no App bypass actor is added; the human `main-protection` bypass is unchanged): it merges its own PRs through the normal 0-approval, green
   `build`/`authorship` gate.
7. **Creates `agent-main` from `main`'s tip only when it is missing.** An **existing** `agent-main` is
   **never rewritten**; a stale/behind ref is left untouched and drift is cleared by the wave-start
   main-sync PR, not by setup.
8. Prints the per-session commands and the remote note.

Options:

| Option | Meaning |
|---|---|
| `--app-name` | App name (must be globally unique; replace the `example-agents` placeholder). |
| `--org` | Create the app under an organization instead of your user. |
| `--workflows` | Also request `workflows: write` (needed if agents edit `.github/workflows/`). |
| `--config-dir` | Where to store the key and env file (default `~/.config/kilocode-agent`). |
| `--existing-app` | App already exists; pass `--app-id` and `--private-key` (skips the manifest flow). |
| `--check` | Validate an existing setup and exit (no changes): rulesets active, `main` reachable from `agent-main` (no drift), authorship variables set. Does **not** require `agent-main` identical to `main`. |
| `--dry-run` | Print the manifest and URLs; change nothing. |
| `--bypass-user-id` | Numeric GitHub user id of the human merge actor allowed to bypass `main-protection` (falls back to `BYPASS_USER_ID`). Required unless `--no-rulesets`; a missing, zero, or non-numeric id fails before any changes are made. |
| `--no-rulesets` | Skip applying the rulesets. |

Re-running is safe: an existing ruleset of the same name is updated, an existing key is replaced, and
an existing `agent-main` ref is left untouched (setup never rewrites it).

## Manual path

If you prefer the UI, or the manifest flow is unavailable:

1. **Create the app:** GitHub → Settings → Developer settings → GitHub Apps → New GitHub App.
   **Homepage URL:** `https://github.com/<owner>/<repo>`. **Webhook URL:** `https://example.com/`
   with **Active** unchecked (`example.com` is a reserved documentation domain; we never use
   webhooks — agents authenticate with an installation token). Repository permissions: **Contents:
   Read and write**, **Pull requests: Read and write**, **Metadata: Read-only** (add **Workflows:
   Read and write** only if agents edit workflow files). No events.
2. **Generate a private key.** Clicking **Generate a private key** downloads a `.pem` immediately to
   your browser's Downloads folder (`<app-slug>.<YYYY-MM-DD>.private-key.pem`). GitHub never shows or
   stores the private key again, so there is no separate download button and a missed download cannot
   be recovered — click **Generate a private key** again to create a new one, and delete the unused
   key from the app settings. Note the **App ID** at the top of the page.
3. **Install the app** on the repository (App → Install App → choose the repo).
4. **Run the setup in existing-app mode** so the key is stored and the installation discovered:
   ```bash
   python3 scripts/setup-agent-github.py --existing-app \
     --app-id <APP_ID> --private-key ~/Downloads/<app>.private-key.pem
   ```

## PEM and env on a clone machine (A5)

Every machine (and every clone) that runs an agent session needs the App credentials, not just the
machine that created the App. Provision two files, both under the config dir (default
`~/.config/kilocode-agent`, or the `--config-dir` you chose):

- `<config-dir>/github-app.pem` — the App private key, mode **`0600`**.
- `<config-dir>/agent-github.env` — mode **`0600`**, naming the PEM path and the App identity:
  `GITHUB_APP_ID`, `GITHUB_APP_PRIVATE_KEY` (the path to the `.pem`), and
  `GITHUB_APP_INSTALLATION_ID`.

**Getting the PEM onto the clone machine.** GitHub shows the private key only once, when it is
generated (or when `setup-agent-github.py` captures it). Transfer it from the machine that generated it
over an encrypted channel — for example `scp`/`rsync` over SSH, or your secrets manager — or re-run
`setup-agent-github.py --existing-app` on the clone machine with a key it can read. Never send it by
email or chat.

**Custody.** Treat the PEM as the critical secret: owner-only `0600`, outside the repository, and never
committed. The env file names the PEM path and is likewise `0600`. **Neither file is ever committed or
placed in the repository**; the config dir lives outside the working tree.

## Per-session use

Each agent session needs a fresh installation token (they expire in about an hour):

```bash
eval "$(python3 scripts/github-app-token.py --export)"   # sets GH_TOKEN
# default: configure a LOCAL credential helper for this clone:
python3 scripts/github-app-token.py --git-local
```

`--git-local` writes `credential.helper` into the clone's **local** git config (never global), so Git
mints a fresh token per operation without a global write. `gh auth setup-git` and its `~/.gitconfig`
write are **not required**; `--gh` remains available as an option if you prefer the global helper.

The helper is the **token-provider seam**: mint on demand, never persist a token, and **re-mint** if a
Git push or API call fails authentication rather than falling back to SSH or a personal access token.

## Remotes: switch to HTTPS

Installation tokens authenticate over **HTTPS**. The remote is often SSH
(`git@github.com:<owner>/<repo>.git`), so an agent using that remote would push with your SSH
key and the PRs would be authored by you — defeating the approval gate. For each agent worktree:

```bash
git remote set-url origin https://github.com/<owner>/<repo>.git
python3 scripts/github-app-token.py --git-local   # writes credential.helper into this clone's LOCAL git config
```

The credential helper is **local-scoped**: `--git-local` writes only the clone's `.git/config`, never
the global `~/.gitconfig`, so `gh auth setup-git` and its global write are **not required**.

## Verify

```bash
python3 scripts/setup-agent-github.py --check
gh api /installation/repositories --jq '.repositories[].full_name'                     # App is installed and can see the repo
gh api repos/<owner>/<repo>/rulesets --jq '.[] | {name, enforcement}'
gh api repos/<owner>/<repo>/commits/main/check-runs --jq '.check_runs[].name'   # CI check contexts
gh variable list --repo <owner>/<repo>                                          # authorship variables
```

`gh api /installation/repositories` proves the **App is installed** and lists the repositories the
installation can reach; `gh api user` does not. The `authorship` check is required on `agent-main`
**only**, not on `main`: planning commits on `main` are human-authored and would fail it.

`--check` passes when both rulesets are active, the authorship variables are set, and `main` is reachable
from `agent-main` (no drift) — not when the two branches are identical. If `main` has advanced past
`agent-main`, `--check` fails and names the drift; the fix is the wave-start main-sync PR, not a rewrite of
`agent-main`.

As a final acceptance check, prove that the App authors a test commit and pull request, and that a
direct push to the protected branch is rejected.

## Troubleshooting

- **Homepage/Webhook URL required:** Homepage = `https://github.com/<owner>/<repo>`; Webhook =
  `https://example.com/` with **Active** unchecked. The manifest supplies both, but the fields can be
  edited before you click Create.
- **Where is the `.pem`?** Clicking **Generate a private key** downloads it to your browser's
  Downloads folder; there is no other way to obtain it. If it is missing, generate a new key (the old
  one becomes unused and can be deleted).
- **App name taken** (HTTP 422 on creation): re-run with `--app-name <unique>`.
- **Callback times out / state mismatch:** re-run; keep the browser and terminal on the same machine.
- **`gh` not authenticated:** `gh auth login` with admin rights on the repo.
- **Ruleset check name mismatch:** pass the exact context to the ruleset script, e.g.
  `CI_CHECK="CI / build" bash scripts/apply-github-rulesets.sh`.
- **Push still attributed to you:** the worktree remote is still SSH; switch it to HTTPS (above).

## What is not automated, and why

- **The two browser clicks** (create the app, install it). GitHub requires an interactive
  confirmation for both; the script opens the pages and waits.
- **Switching remotes.** The script prints the command rather than rewriting your git configuration.
- **Key custody.** The script stores the key under the config dir with `0600`; treat that file as the
  critical secret and never commit it or place it in the repository.

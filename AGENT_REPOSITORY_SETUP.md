# Agent Repository Setup

This setup gives agent sessions a distinct GitHub App identity, an `agent-main` integration branch,
protected story branches, and a local guard against accidental pushes to `main`. It is mandatory
before agents implement milestones; projects that use the toolkit only for planning may omit it.

The operational branch model is defined in [`resources/agent/git-workflow.md`](resources/agent/git-workflow.md).

## Parameters

| Parameter | Meaning |
|---|---|
| `<owner>/<repo>` | Repository on which the App is installed. |
| `<app-id>` | Numeric GitHub App id. |
| `<installation-id>` | App installation id for the repository. |
| `<unique-app-name>` | Globally unique name used when creating a new GitHub App. |
| `<app-slug>` | App slug; its bot login is `<app-slug>[bot]`. |
| `<bot-user-id>` | Bot user id used in the GitHub no-reply email. |
| `<config-dir>` | External credential directory; default `${XDG_CONFIG_HOME:-$HOME/.config}/kilocode-agent`. |
| `<human-user-id>` | Human actor allowed to perform the final merge. |

## Prerequisites

- `gh` authenticated with repository administration permission.
- Python 3 and OpenSSL.
- An HTTPS clone; do not use an SSH remote for an App-authored workflow.

## 1. Confirm the human bootstrap on `main`

```bash
git clone https://github.com/<owner>/<repo>.git <workdir>
cd <workdir>
git remote set-url origin https://github.com/<owner>/<repo>.git
git switch main
git pull --ff-only origin main
```

The composed toolkit, `kilocode-composition.json`, and `.kilocode-pack.lock.json` must already be in a
human-authored bootstrap commit on `main`. Do not start agent work before that commit is pushed.

## 2. Create or reuse the GitHub App

Choose an owner-only directory outside the repository for the App key and configuration:

```bash
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/kilocode-agent"
mkdir -p "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"
```

For an existing App, copy its private key into that directory, lock down its permissions, and run the
existing-App path:

```bash
cp /secure/source/github-app.pem "$CONFIG_DIR/github-app.pem"
chmod 600 "$CONFIG_DIR/github-app.pem"
python3 scripts/setup-agent-github.py \
  --existing-app \
  --app-id <app-id> \
  --private-key "$CONFIG_DIR/github-app.pem" \
  --repo <owner>/<repo> \
  --config-dir "$CONFIG_DIR" \
  --bypass-user-id <human-user-id>
```

For a new App, run the manifest flow instead; the script opens the two GitHub pages that require human
confirmation and stores the returned key in `CONFIG_DIR`:

```bash
python3 scripts/setup-agent-github.py \
  --app-name <unique-app-name> \
  --repo <owner>/<repo> \
  --config-dir "$CONFIG_DIR" \
  --bypass-user-id <human-user-id>
```

Both paths verify the App, write `agent-github.env` with mode `0600`, resolve the App slug and bot
identity, set repository variables, create `agent-main` when absent, and apply the rulesets. Never
place the private key or token inside the repository or send it through chat or email.

## 3. Fetch and select the integration branch

The setup command creates `agent-main` from `main` when it is missing and never rewrites an existing
integration branch. After it succeeds:

```bash
git fetch origin agent-main
git switch --track origin/agent-main 2>/dev/null || git switch agent-main
```

If an existing `agent-main` is behind or has diverged from `main`, stop and follow the main-sync PR
procedure in `resources/agent/git-workflow.md`; never repair it with a force push.

## 4. Set the local commit identity

```bash
git config user.name  "<app-slug>[bot]"
git config user.email "<bot-user-id>+<app-slug>[bot]@users.noreply.github.com"
```

The bot user id is not the same as the App id.

## 5. Configure Git credentials and local guards

```bash
python3 scripts/github-app-token.py \
  --git-local \
  --env-file "$CONFIG_DIR/agent-github.env"
bash scripts/install-git-guards.sh
mkdir -p .kilo/worktrees
ln -sfn .kilo/worktrees worktrees
```

The credential helper mints installation tokens on demand and stores no token in the repository. The
tracked pre-push hook refuses pushes to `main` or `master`; a human can deliberately override it with
`ALLOW_MAIN_PUSH=1` or `--no-verify`.

## 6. Verify

```bash
python3 scripts/setup-agent-github.py \
  --check \
  --repo <owner>/<repo> \
  --config-dir "$CONFIG_DIR"
git config --local --get user.email
git config --local --get core.hooksPath
git remote -v
git branch --show-current
```

Confirm that the App can push a disposable `story/*` branch, the `authorship` check runs, direct pushes
to protected branches fail, `git branch --show-current` reports `agent-main`, and the human can perform
the final `agent-main` to `main` merge. Do not begin milestone implementation until all checks pass.

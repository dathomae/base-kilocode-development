#!/usr/bin/env python3
"""Mint a GitHub App installation access token for agent sessions.

GitHub App installation tokens are short-lived (~1 hour) and are the recommended
credential for agent automation: a distinct non-user actor that branch-protection
approval rules bind, while a different actor (the human) approves.

Environment:
  GITHUB_APP_ID                 required  numeric App ID
  GITHUB_APP_PRIVATE_KEY        required  path to the App's private key (.pem)
  GITHUB_APP_INSTALLATION_ID    optional  installation id; discovered if absent
  GITHUB_REPO                   optional  owner/name, used by --gh for setup-git

Usage:
  scripts/github-app-token.py                 # print the token
  scripts/github-app-token.py --export        # print `export GH_TOKEN=...`
  scripts/github-app-token.py --gh            # configure gh (and git) to use it
  scripts/github-app-token.py --json          # print the full API response
  scripts/github-app-token.py --git-local     # write a local-scoped git credential helper
  scripts/github-app-token.py --git-credential get   # internal helper-protocol mode

The `--git-local` mode writes `credential.helper` into the clone's **local** Git
config only (never `--global`/`--system`): an empty reset value followed by one
helper command that sources the App env file (`--env-file`, default
`~/.config/kilocode-agent/agent-github.env`) and runs this script's internal
`--git-credential` mode. It is the default clone path — onboarding needs no
`gh auth setup-git` and no global Git config write — and a re-run is idempotent.
`--gh` remains available as an alternative.

Uses only the Python standard library plus the `openssl` CLI for RS256 signing,
so it has no third-party dependencies.

Note: installation tokens authenticate over HTTPS. If the repository's remote is
SSH (git@github.com:...), pushes will not use this token; use an HTTPS remote or
a credential helper (the --gh option runs `gh auth setup-git` for HTTPS, or
--git-local configures a local-scoped helper).

Mint on demand per session and re-mint when authentication fails; never persist a token.
"""

import argparse
import base64
import json
import os
import shlex
import subprocess
import sys
import time
import urllib.error
import urllib.request

API = "https://api.github.com"
DEFAULT_ENV_FILE = "~/.config/kilocode-agent/agent-github.env"
REQUIRED_ENV = ("GITHUB_APP_ID", "GITHUB_APP_PRIVATE_KEY")


def b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode()


def make_jwt(app_id: str, key_path: str) -> str:
    now = int(time.time())
    header = {"alg": "RS256", "typ": "JWT"}
    payload = {"iat": now - 60, "exp": now + 540, "iss": app_id}
    signing_input = "{}.{}".format(
        b64url(json.dumps(header, separators=(",", ":")).encode()),
        b64url(json.dumps(payload, separators=(",", ":")).encode()),
    )
    try:
        result = subprocess.run(
            ["openssl", "dgst", "-sha256", "-sign", key_path],
            input=signing_input.encode(),
            capture_output=True,
            check=True,
        )
    except FileNotFoundError:
        sys.exit("openssl not found on PATH")
    except subprocess.CalledProcessError as exc:
        sys.exit("openssl signing failed: {}".format(exc.stderr.decode().strip()))
    return "{}.{}".format(signing_input, b64url(result.stdout))


def api(path: str, method: str = "GET", token: str = None, body=None):
    req = urllib.request.Request(API + path, method=method)
    req.add_header("Accept", "application/vnd.github+json")
    req.add_header("User-Agent", "kilocode-agent-token")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    data = None
    if body is not None:
        data = json.dumps(body).encode()
        req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req, data) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as exc:
        sys.exit("GitHub API {} {} failed: {} {}".format(method, path, exc.code, exc.read().decode()))


def mint() -> dict:
    """Mint a fresh installation access token and return the full API response.

    Validates the provider's environment contract, signs the App JWT, resolves the
    installation (discovered when `GITHUB_APP_INSTALLATION_ID` is absent), and
    exchanges the JWT for an installation token. Prints nothing to stdout.
    """
    app_id = os.environ.get("GITHUB_APP_ID")
    key = os.environ.get("GITHUB_APP_PRIVATE_KEY")
    if not app_id or not key:
        sys.exit("set GITHUB_APP_ID and GITHUB_APP_PRIVATE_KEY (path to the .pem)")
    if not os.path.exists(key):
        sys.exit("private key not found: {}".format(key))

    jwt = make_jwt(app_id, key)

    installation = os.environ.get("GITHUB_APP_INSTALLATION_ID")
    if not installation:
        installs = api("/app/installations", token=jwt)
        if not installs:
            sys.exit("no installations found for this app")
        installation = installs[0]["id"]

    return api("/app/installations/{}/access_tokens".format(installation), method="POST", token=jwt)


def read_env_file(path: str) -> dict:
    """Parse `export KEY=value` lines from an App env file without shell evaluation."""
    values = {}
    with open(path) as handle:
        for line in handle:
            line = line.strip()
            if line.startswith("export "):
                line = line[len("export "):].strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, _, value = line.partition("=")
            value = value.strip()
            if len(value) >= 2 and value[0] == value[-1] and value[0] in "\"'":
                value = value[1:-1]
            values[key.strip()] = value
    return values


def git(*args):
    try:
        return subprocess.run(["git", *args], capture_output=True, text=True)
    except FileNotFoundError:
        sys.exit("git not found on PATH")


def build_helper_command(env_path: str) -> str:
    """Build the shell snippet recorded as the local `credential.helper` value.

    Git runs a `!`-prefixed helper through the shell with the operation appended,
    so the `f() { ...; }; f` wrapper receives `get`/`store`/`erase` as `$1`. The
    snippet sources the App env file (when present) before invoking this script's
    `--git-credential` mode, so a git-invoked helper process has the App
    credentials; an already-exported env also works. Paths are absolute so the
    snippet does not depend on git's invocation cwd.
    """
    script = os.path.realpath(__file__)
    python = sys.executable or "python3"
    return "!" + (
        "f() {{ set -a; [ -f {env} ] && . {env}; set +a; "
        "exec {python} {script} --git-credential \"$@\"; }}; f"
    ).format(
        env=shlex.quote(env_path),
        python=shlex.quote(python),
        script=shlex.quote(script),
    )


def configure_git_local(env_file: str) -> None:
    """Write the local-scoped credential helper into the current repository.

    Writes `credential.helper` into the clone's LOCAL git config only, as an empty
    reset followed by a single helper command, so a re-run is idempotent and the
    inherited global helper list is reset. Never touches global/system config and
    never runs `gh auth setup-git`.
    """
    if git("rev-parse", "--show-toplevel").returncode != 0:
        sys.exit(
            "--git-local must run inside a git repository "
            "(git rev-parse --show-toplevel failed; cd into the agent clone first)"
        )

    env_path = os.path.abspath(os.path.expanduser(env_file))
    if not os.path.isfile(env_path):
        sys.exit(
            "App env file not found: {} (run scripts/setup-agent-github.py to provision it, "
            "or pass --env-file)".format(env_path)
        )
    values = read_env_file(env_path)
    missing = [name for name in REQUIRED_ENV if not values.get(name)]
    if missing:
        sys.exit(
            "App env file {} is missing required variable(s): {}".format(
                env_path, ", ".join(missing)
            )
        )
    key = os.path.expandvars(os.path.expanduser(values["GITHUB_APP_PRIVATE_KEY"]))
    if not os.path.isfile(key):
        sys.exit("private key not found: {} (from {})".format(key, env_path))

    command = build_helper_command(env_path)
    reset = git("config", "--local", "--replace-all", "credential.helper", "")
    if reset.returncode != 0:
        sys.exit("could not write local git config: {}".format(reset.stderr.strip()))
    add = git("config", "--local", "--add", "credential.helper", command)
    if add.returncode != 0:
        sys.exit("could not write local git config: {}".format(add.stderr.strip()))
    print(command)


def git_credential(operation: str) -> None:
    """Implement the git credential-helper protocol for this script.

    Git invokes `<script> --git-credential get` (appending the operation). Only
    `get` emits credentials — `username=x-access-token` and a freshly minted
    installation token as the password — while `store`/`erase` (and any other
    operation git may append) print nothing. No token is persisted.
    """
    sys.stdin.read()  # consume git's credential request / response
    if operation == "get":
        token = mint()["token"]
        sys.stdout.write("username=x-access-token\npassword={}\n".format(token))


def main() -> None:
    parser = argparse.ArgumentParser(description="Mint a GitHub App installation access token.")
    parser.add_argument("--json", action="store_true", help="print the full API response")
    parser.add_argument("--export", action="store_true", help="print an `export GH_TOKEN=...` line")
    parser.add_argument(
        "--gh",
        action="store_true",
        help="also run `gh auth login --with-token` and `gh auth setup-git`",
    )
    parser.add_argument(
        "--git-local",
        action="store_true",
        help="write a local-scoped git credential helper for this clone (no gh, no global config)",
    )
    parser.add_argument(
        "--env-file",
        default=DEFAULT_ENV_FILE,
        help="App env file the local helper sources (default {})".format(DEFAULT_ENV_FILE),
    )
    parser.add_argument(
        "--git-credential",
        action="store_true",
        help="internal git credential-helper protocol mode (invoked by git)",
    )
    parser.add_argument("operation", nargs="?", default=None, help=argparse.SUPPRESS)
    args = parser.parse_args()

    if args.git_credential:
        if args.operation is None:
            parser.error("--git-credential requires an operation (get, store, erase)")
        git_credential(args.operation)
        return

    if args.git_local:
        configure_git_local(args.env_file)
        return

    if args.operation is not None:
        parser.error("unexpected argument: {!r}".format(args.operation))

    resp = mint()

    if args.json:
        print(json.dumps(resp, indent=2))
        return

    token = resp["token"]

    if args.gh:
        subprocess.run(["gh", "auth", "login", "--with-token"], input=token.encode(), check=True)
        subprocess.run(["gh", "auth", "setup-git"], check=True)
        print(
            "gh authenticated as the app installation; token expires {}".format(resp.get("expires_at")),
            file=sys.stderr,
        )

    if args.export:
        print("export GH_TOKEN={}".format(token))
    elif not args.gh:
        print(token)


if __name__ == "__main__":
    main()

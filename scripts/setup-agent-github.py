#!/usr/bin/env python3
"""Set up the GitHub App identity and branch rulesets for agent sessions.

This is a guided, mostly-automated setup assistant. GitHub requires a browser click to
create an app and to install it on a repository, so those two steps are opened for you;
everything else is automated:

  1. create the GitHub App via the manifest flow (returns the App ID and private key),
  2. store the private key outside the repository (0600) and write an env file,
  3. install the app on the repository (browser) and discover the installation id,
  4. mint a test installation token to verify the credentials,
  5. set the repository variables the commit-authorship check compares against,
  6. apply the branch rulesets
     (delegates to scripts/apply-github-rulesets.sh; needs the human merge
     actor's numeric user id via --bypass-user-id or $BYPASS_USER_ID, unless
     --no-rulesets is given),
  7. create refs/heads/agent-main from main if it does not exist (App-authorized),
  8. print the remaining per-session and remote instructions.

Modes:
  default                 full setup via the manifest flow
  --existing-app          you already created the app; pass --app-id and --private-key
  --check                 validate an existing setup without changing anything
  --dry-run               print the manifest and URLs, change nothing

Prerequisites: `gh` authenticated with admin on the repository, `python3`, `openssl`.

Examples:
  python3 scripts/setup-agent-github.py --repo <owner>/<repo>
  python3 scripts/setup-agent-github.py --existing-app --app-id 123456 --private-key ~/Downloads/app.pem
  python3 scripts/setup-agent-github.py --existing-app --app-id 123456 --private-key ~/Downloads/app.pem --bypass-user-id 680936
  python3 scripts/setup-agent-github.py --check
"""

import argparse
import html
import http.server
import importlib.util
import json
import os
import pathlib
import secrets
import shutil
import subprocess
import sys
import threading
import urllib.parse
import webbrowser

HERE = pathlib.Path(__file__).resolve().parent
DEFAULT_CONFIG_DIR = pathlib.Path(os.environ.get("XDG_CONFIG_HOME", pathlib.Path.home() / ".config")) / "kilocode-agent"
DEFAULT_REPO = "example/example-repo"


def load_token_module():
    spec = importlib.util.spec_from_file_location("ghapp", HERE / "github-app-token.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


ghapp = load_token_module()


def run(cmd, **kwargs):
    return subprocess.run(cmd, **kwargs)


def check_prerequisites(repo):
    problems = []
    if shutil.which("gh") is None:
        problems.append("gh CLI not found on PATH")
    else:
        auth = run(["gh", "auth", "status"], capture_output=True, text=True)
        if auth.returncode != 0:
            problems.append("gh is not authenticated; run `gh auth login` with admin rights")
    if shutil.which("openssl") is None:
        problems.append("openssl not found on PATH")
    if not repo or "/" not in repo:
        problems.append("--repo must be owner/name")
    return problems


def build_manifest(app_name, repo, port, workflows, org):
    redirect = "http://127.0.0.1:{}/callback".format(port)
    permissions = {"contents": "write", "pull_requests": "write", "metadata": "read"}
    if workflows:
        permissions["workflows"] = "write"
    return {
        "name": app_name,
        "url": "https://github.com/{}".format(repo),
        "hook_attributes": {"url": "https://example.com/", "active": False},
        "redirect_url": redirect,
        "public": False,
        "default_permissions": permissions,
        "default_events": [],
    }


def manifest_action_url(org):
    if org:
        return "https://github.com/organizations/{}/settings/apps/new".format(org)
    return "https://github.com/settings/apps/new"


def manifest_form_html(manifest, state, org):
    action = "{}?state={}".format(manifest_action_url(org), urllib.parse.quote(state))
    payload = html.escape(json.dumps(manifest), quote=True)
    return (
        "<!doctype html><html><head><meta charset='utf-8'><title>Creating GitHub App…</title></head>"
        "<body onload='document.forms[0].submit()'>"
        "<form action='{}' method='post'>"
        "<input type='hidden' name='manifest' value=\"{}\">"
        "<noscript><button type='submit'>Create GitHub App</button></noscript>"
        "</form></body></html>"
    ).format(action, payload)


def local_manifest_url(port):
    """Return the loopback URL the script serves the manifest form on."""
    return "http://127.0.0.1:{}/".format(port)


def loopback_error_message(port, reason):
    """Return an actionable multi-line message for an unreachable loopback.

    Names the local URL to open manually and explains the loopback
    requirement, so a browser that cannot reach 127.0.0.1:<port> can be
    diagnosed instead of leaving GitHub's blank App page.
    """
    url = local_manifest_url(port)
    return (
        "GitHub App creation was not completed ({reason}).\n"
        "The browser did not deliver the callback to the local page.\n"
        "Open this local page manually to finish creating the App:\n"
        "  {url}\n"
        "It needs a browser able to reach 127.0.0.1:{port} on this host; "
        "if the browser runs elsewhere (remote or containerized), forward the "
        "port to your machine and open the URL there.".format(
            reason=reason, url=url, port=port
        )
    )


class CallbackHandler(http.server.BaseHTTPRequestHandler):
    server_version = "kilocode-agent-setup"
    result = {}
    expected_state = None
    manifest_html = ""

    def log_message(self, *args):  # silence default logging
        pass

    def _send(self, body, status=200):
        data = body.encode()
        self.send_response(status)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        parsed = urllib.parse.urlparse(self.path)
        if parsed.path == "/":
            self._send(self.manifest_html)
            return
        if parsed.path == "/callback":
            params = urllib.parse.parse_qs(parsed.query)
            state = (params.get("state") or [""])[0]
            code = (params.get("code") or [""])[0]
            if self.expected_state and state != self.expected_state:
                self._send("<h2>State mismatch; aborting.</h2>", status=400)
                self.result["error"] = "state mismatch"
            elif not code:
                self._send("<h2>No code returned; aborting.</h2>", status=400)
                self.result["error"] = "no code"
            else:
                self.result["code"] = code
                self._send(
                    "<h2>App created.</h2><p>You can close this tab and return to the terminal.</p>"
                )
            threading.Thread(target=self.server.shutdown, daemon=True).start()
            return
        self._send("<h2>Not found</h2>", status=404)


def create_app_via_manifest(app_name, repo, port, workflows, org, timeout=300):
    manifest = build_manifest(app_name, repo, port, workflows, org)
    state = secrets.token_urlsafe(16)
    CallbackHandler.expected_state = state
    CallbackHandler.manifest_html = manifest_form_html(manifest, state, org)

    server = http.server.ThreadingHTTPServer(("127.0.0.1", port), CallbackHandler)
    server.timeout = 1
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()

    url = local_manifest_url(port)
    print("Opening the GitHub App creation page. If it does not open, visit:")
    print("  " + url)
    try:
        webbrowser.open(url)
    except Exception:
        pass

    print("Waiting for you to click 'Create GitHub App' in the browser…")
    thread.join(timeout)
    server.shutdown()
    if "code" not in CallbackHandler.result:
        raise SystemExit(
            loopback_error_message(
                port, CallbackHandler.result.get("error", "timed out")
            )
        )

    code = CallbackHandler.result["code"]
    print("Exchanging the manifest code for app credentials…")
    resp = ghapp.api("/app-manifests/{}/conversions".format(code), method="POST")
    return resp


def write_secure_config(config_dir, pem_text, app_id, installation_id, slug):
    config_dir = pathlib.Path(config_dir).expanduser()
    config_dir.mkdir(parents=True, exist_ok=True)
    os.chmod(config_dir, 0o700)
    pem_path = config_dir / "github-app.pem"
    pem_path.write_text(pem_text)
    os.chmod(pem_path, 0o600)
    env_path = config_dir / "agent-github.env"
    env_path.write_text(
        "export GITHUB_APP_ID={}\n".format(app_id)
        + "export GITHUB_APP_PRIVATE_KEY={}\n".format(pem_path)
        + "export GITHUB_APP_INSTALLATION_ID={}\n".format(installation_id or "")
    )
    os.chmod(env_path, 0o600)
    return pem_path, env_path


def same_file(src, dst):
    """Return True if src and dst refer to the same file (os.path.samefile semantics)."""
    src = pathlib.Path(src).expanduser()
    dst = pathlib.Path(dst).expanduser()
    if dst.exists():
        try:
            return os.path.samefile(src, dst)
        except OSError:
            return False
    return os.path.realpath(src) == os.path.realpath(dst)


def store_existing_key(config_dir, private_key):
    src = pathlib.Path(private_key).expanduser()
    if not src.exists():
        raise SystemExit("private key not found: {}".format(src))
    config_dir = pathlib.Path(config_dir).expanduser()
    config_dir.mkdir(parents=True, exist_ok=True)
    os.chmod(config_dir, 0o700)
    pem_path = config_dir / "github-app.pem"
    if same_file(src, pem_path):
        os.chmod(pem_path, 0o600)
        return pem_path
    shutil.copyfile(src, pem_path)
    os.chmod(pem_path, 0o600)
    return pem_path


def discover_installation(app_id, pem_path, repo):
    jwt = ghapp.make_jwt(app_id, str(pem_path))
    installs = ghapp.api("/app/installations", token=jwt)
    owner = repo.split("/")[0].lower()
    for inst in installs:
        account = (inst.get("account") or {}).get("login", "").lower()
        if account == owner:
            return inst["id"], inst.get("html_url", "")
    return None, ""


def install_url(slug):
    if not slug:
        return "https://github.com/settings/installations"
    return "https://github.com/apps/{}/installations/new".format(slug)


def verify_token(app_id, pem_path, installation_id):
    jwt = ghapp.make_jwt(app_id, str(pem_path))
    resp = ghapp.api("/app/installations/{}/access_tokens".format(installation_id), method="POST", token=jwt)
    return resp.get("expires_at"), resp.get("token", "")[:12]


def resolve_bypass_user_id(value):
    """Return a validated positive-integer bypass user id, or None if unset.

    ``value`` is the raw ``--bypass-user-id`` flag or ``BYPASS_USER_ID`` value.
    ``None`` or blank means no id was provided. A non-blank value that is not a
    positive integer raises SystemExit naming the flags, so an explicit ``0`` (or
    any non-numeric value) cannot reach the ruleset payload as GitHub's invalid
    ``actor_id``.
    """
    if value is None:
        return None
    text = str(value).strip()
    if not text:
        return None
    if not text.isdigit() or int(text) <= 0:
        raise SystemExit(
            "Invalid bypass user id {!r}: --bypass-user-id/BYPASS_USER_ID must be a "
            "positive integer GitHub user id. Pass --bypass-user-id <id>, set "
            "BYPASS_USER_ID, or pass --no-rulesets to skip applying the rulesets.".format(value)
        )
    return int(text)


def apply_rulesets(repo, bypass_user_id=None):
    script = HERE / "apply-github-rulesets.sh"
    env = dict(os.environ, REPO=repo)
    if bypass_user_id is not None:
        env["BYPASS_USER_ID"] = str(bypass_user_id)
    return run(["bash", str(script)], env=env)


def bot_login(slug):
    """Return the App bot login for a slug: <slug>[bot]."""
    return "{}[bot]".format(slug)


def bot_email(bot_user_id, slug):
    """Return the App bot noreply email.

    GitHub's bot noreply email uses the bot USER id, not the App id:
    <bot-user-id>+<slug>[bot]@users.noreply.github.com
    """
    return "{}+{}[bot]@users.noreply.github.com".format(bot_user_id, slug)


def bot_user_path(slug):
    """Return the API path for the App bot user, with the brackets URL-encoded."""
    return "/users/{}".format(urllib.parse.quote(bot_login(slug), safe=""))


def resolve_app_slug(app_id, pem_path):
    """Resolve the App slug via GET /app using the App JWT.

    Raises SystemExit with an actionable message naming the App id and the failed
    GET /app when the App cannot be read or the response carries no slug.
    """
    try:
        jwt = ghapp.make_jwt(str(app_id), str(pem_path))
        app = ghapp.api("/app", token=jwt)
    except (SystemExit, OSError, ValueError) as exc:
        raise SystemExit(
            "Could not read the App (id {}) via GET /app: {}. "
            "Pass --slug or fix the App credentials.".format(app_id, exc)
        )
    slug = (app or {}).get("slug", "")
    if not slug:
        raise SystemExit(
            "Could not read the App (id {}) via GET /app: the response had no slug. "
            "Pass --slug explicitly.".format(app_id)
        )
    return slug


def set_authorship_variables(repo, slug):
    """Set the repository variables the commit-authorship check compares against."""
    if not slug:
        print(
            "App slug unknown; set repository variables AGENT_BOT_EMAIL and "
            "AGENT_BOT_LOGIN manually (see agent-setup.md).",
            file=sys.stderr,
        )
        return
    resolved = run(
        ["gh", "api", bot_user_path(slug), "--jq", ".id"],
        capture_output=True, text=True,
    )
    bot_id = (resolved.stdout or "").strip()
    if resolved.returncode != 0 or not bot_id:
        print(
            "Could not resolve the App bot user id for '{}'; set AGENT_BOT_EMAIL and "
            "AGENT_BOT_LOGIN manually (see agent-setup.md).".format(bot_login(slug)),
            file=sys.stderr,
        )
        return
    email = bot_email(bot_id, slug)
    login = bot_login(slug)
    run(["gh", "variable", "set", "AGENT_BOT_EMAIL", "--repo", repo, "--body", email])
    run(["gh", "variable", "set", "AGENT_BOT_LOGIN", "--repo", repo, "--body", login])
    print("Set authorship variables: {} <{}>".format(login, email))


def read_env(config_dir):
    env_path = pathlib.Path(config_dir).expanduser() / "agent-github.env"
    values = {}
    if env_path.exists():
        for line in env_path.read_text().splitlines():
            line = line.strip()
            if line.startswith("export ") and "=" in line:
                key, _, val = line[len("export "):].partition("=")
                values[key.strip()] = val.strip()
    return values


def ensure_agent_main(repo, app_id, pem_path, installation_id):
    """Create refs/heads/agent-main from main only when it does not exist.

    Create-only and idempotent: the existing ref is checked first, and an
    existing agent-main is left untouched — never updated, force-reset, or
    rewritten. ``main`` is resolved only when the ref is missing.

    Uses the App installation token so creation is authorized. Drift is not
    reconciled here; ``--check`` detects it (the wave-start main-sync PR clears it).
    """
    jwt = ghapp.make_jwt(str(app_id), str(pem_path))
    token = ghapp.api(
        "/app/installations/{}/access_tokens".format(installation_id), method="POST", token=jwt
    )["token"]
    env = dict(os.environ, GH_TOKEN=token)

    existing = run(
        ["gh", "api", "repos/{}/git/ref/heads/agent-main".format(repo), "--jq", ".object.sha"],
        capture_output=True, text=True, env=env,
    )
    if existing.returncode == 0:
        print("agent-main already exists at {}".format(existing.stdout.strip()[:12]))
        return

    main = run(
        ["gh", "api", "repos/{}/git/ref/heads/main".format(repo), "--jq", ".object.sha"],
        capture_output=True, text=True, env=env,
    )
    main_sha = main.stdout.strip()
    if main.returncode != 0 or not main_sha:
        print("Could not resolve main; create agent-main manually.", file=sys.stderr)
        return

    created = run(
        ["gh", "api", "-X", "POST", "repos/{}/git/refs".format(repo),
         "-f", "ref=refs/heads/agent-main", "-f", "sha={}".format(main_sha)],
        capture_output=True, text=True, env=env,
    )
    if created.returncode != 0:
        print("Could not create agent-main: {}".format(created.stderr.strip()), file=sys.stderr)
        return
    print("Created agent-main at {} (from main)".format(main_sha[:12]))


def is_no_drift(compare_status):
    """Return True when main is reachable from agent-main (no drift).

    The GitHub compare API is called as base=agent-main, head=main, so
    ``identical`` and ``behind`` mean every commit of main is reachable from
    agent-main; ``ahead``/``diverged`` mean main has commits that are not.
    """
    return compare_status in ("identical", "behind")


def drift_report(compare_json):
    """Return an actionable drift message naming the unreachable main commits.

    ``compare_json`` is the parsed ``repos/{repo}/compare/agent-main...main``
    response; its ``commits`` list is exactly the main commits not reachable
    from agent-main. Directs the operator to the wave-start main-sync PR.
    """
    commits = compare_json.get("commits") or []
    shas = [entry.get("sha") for entry in commits if entry.get("sha")]
    detail = ", ".join(shas) if shas else "(no commit details returned)"
    return (
        "drift: main has commits not reachable from agent-main: {}. "
        "Open the wave-start main-sync PR to bring agent-main up to main.".format(detail)
    )


def check_setup(repo):
    """Verify rulesets, agent-main reachability, and the authorship variables.

    Returns a list of problem strings. ``agent-main`` drift is the settled
    no-drift/reachability semantics: every commit of ``main`` must be reachable
    from ``agent-main`` (``identical``/``behind``), never branch identity.
    """
    problems = []

    rules = run(
        ["gh", "api", "repos/{}/rulesets".format(repo), "--jq",
         '[.[] | select(.enforcement=="active") | .name] | join(",")'],
        capture_output=True, text=True,
    )
    if rules.returncode != 0:
        problems.append("could not read rulesets: {}".format(rules.stderr.strip()))
    else:
        active = rules.stdout.strip()
        for name in ("agent-main-protection", "main-protection"):
            if name not in active:
                problems.append("ruleset not active: {}".format(name))

    agent = run(
        ["gh", "api", "repos/{}/git/ref/heads/agent-main".format(repo), "--jq", ".object.sha"],
        capture_output=True, text=True,
    )
    if agent.returncode != 0:
        problems.append("agent-main does not exist")
    else:
        compare = run(
            ["gh", "api", "repos/{}/compare/agent-main...main".format(repo)],
            capture_output=True, text=True,
        )
        if compare.returncode != 0:
            problems.append(
                "could not compare agent-main with main: {}".format(compare.stderr.strip())
            )
        else:
            try:
                compare_json = json.loads(compare.stdout)
            except ValueError as exc:
                problems.append(
                    "could not compare agent-main with main: invalid JSON ({})".format(exc)
                )
            else:
                if not is_no_drift(compare_json.get("status", "")):
                    problems.append(drift_report(compare_json))

    variables = run(["gh", "variable", "list", "--repo", repo], capture_output=True, text=True)
    if variables.returncode != 0:
        problems.append("could not list repository variables: {}".format(variables.stderr.strip()))
    else:
        for name in ("AGENT_BOT_EMAIL", "AGENT_BOT_LOGIN"):
            if name not in variables.stdout:
                problems.append("repository variable missing: {}".format(name))

    return problems


def main():
    parser = argparse.ArgumentParser(
        description="Set up the GitHub App identity and branch rulesets for agent sessions.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("--repo", default=DEFAULT_REPO, help="owner/name (default {})".format(DEFAULT_REPO))
    parser.add_argument("--app-name", default="example-agents", help="GitHub App name (must be globally unique; override the placeholder default)")
    parser.add_argument("--org", default=None, help="organization login, if the app belongs to an org")
    parser.add_argument("--port", type=int, default=8976, help="local callback port (default 8976)")
    parser.add_argument("--config-dir", default=str(DEFAULT_CONFIG_DIR), help="where to store the key and env file")
    parser.add_argument("--workflows", action="store_true", help="also request workflows:write (agents edit workflow files)")
    parser.add_argument("--existing-app", action="store_true", help="app already exists; use --app-id and --private-key")
    parser.add_argument("--app-id", default=None, help="existing App ID")
    parser.add_argument("--private-key", default=None, help="existing private key path")
    parser.add_argument("--slug", default=None, help="existing App slug (resolved via GET /app if omitted)")
    parser.add_argument("--installation-id", default=None, help="existing installation id (discovered if omitted)")
    parser.add_argument("--check", action="store_true", help="validate an existing setup (token, rulesets, agent-main, variables) and exit")
    parser.add_argument("--dry-run", action="store_true", help="print the manifest and URLs, change nothing")
    parser.add_argument("--no-rulesets", action="store_true", help="skip applying the rulesets")
    parser.add_argument(
        "--bypass-user-id",
        default=None,
        help="numeric GitHub user id of the human merge actor allowed to bypass main-protection "
             "(falls back to the BYPASS_USER_ID environment variable; required unless --no-rulesets)",
    )
    args = parser.parse_args()

    config_dir = pathlib.Path(args.config_dir).expanduser()

    if args.check:
        values = read_env(config_dir)
        if not values:
            raise SystemExit("no env file found at {}".format(config_dir / "agent-github.env"))
        pem = values.get("GITHUB_APP_PRIVATE_KEY", "")
        app_id = values.get("GITHUB_APP_ID", "")
        inst = values.get("GITHUB_APP_INSTALLATION_ID", "")
        if not pem or not pathlib.Path(pem).exists():
            raise SystemExit("private key missing at {}".format(pem))
        if not inst:
            inst, _ = discover_installation(app_id, pem, args.repo)
            if not inst:
                raise SystemExit("app is not installed on {}".format(args.repo))
        expires, prefix = verify_token(app_id, pem, inst)
        print("OK: app {} installed as {}; test token {}… expires {}".format(app_id, inst, prefix, expires))
        problems = check_setup(args.repo)
        if problems:
            for problem in problems:
                print("PROBLEM: " + problem, file=sys.stderr)
            raise SystemExit(1)
        print("OK: rulesets active, no drift (main reachable from agent-main), authorship variables set")
        return

    if args.dry_run:
        manifest = build_manifest(args.app_name, args.repo, args.port, args.workflows, args.org)
        print("Manifest:")
        print(json.dumps(manifest, indent=2))
        print("\nCreate page: {}".format(manifest_action_url(args.org)))
        print("Install page (after creation): https://github.com/apps/<slug>/installations/new")
        print("Config dir: {}".format(config_dir))
        return

    # Validate the bypass user id before any side effects: on the manifest path
    # this must fail before the App is created/installed, otherwise an omitted id
    # creates an App and a re-run creates a second one.
    if not args.no_rulesets:
        bypass_user_id = resolve_bypass_user_id(
            args.bypass_user_id
            if args.bypass_user_id is not None
            else os.environ.get("BYPASS_USER_ID")
        )
        if bypass_user_id is None:
            raise SystemExit(
                "Applying branch rulesets needs the human merge actor's numeric GitHub "
                "user id: pass --bypass-user-id <id> (or set BYPASS_USER_ID), or pass "
                "--no-rulesets to skip applying the rulesets."
            )

    problems = check_prerequisites(args.repo)
    if problems:
        for p in problems:
            print("ERROR: " + p, file=sys.stderr)
        raise SystemExit(1)

    if args.existing_app:
        if not args.app_id or not args.private_key:
            raise SystemExit("--existing-app requires --app-id and --private-key")
        pem_path = store_existing_key(config_dir, args.private_key)
        app_id = args.app_id
        slug = args.slug or resolve_app_slug(app_id, pem_path)
        print("Stored private key at {}".format(pem_path))
    else:
        print("Creating GitHub App '{}' via the manifest flow…".format(args.app_name))
        resp = create_app_via_manifest(args.app_name, args.repo, args.port, args.workflows, args.org)
        app_id = resp["id"]
        slug = resp.get("slug", "")
        pem_path, _ = write_secure_config(config_dir, resp["pem"], app_id, "", slug)
        print("Created app id {} (slug {}); private key stored at {}".format(app_id, slug, pem_path))

    # Installation: discover, or guide the user to install.
    installation_id = args.installation_id or None
    if not installation_id:
        installation_id, _ = discover_installation(app_id, pem_path, args.repo)
    if not installation_id:
        url = install_url(slug)
        print("\nThe app is not installed on {} yet.".format(args.repo))
        print("Install it here, then press Enter: {}".format(url))
        try:
            webbrowser.open(url)
        except Exception:
            pass
        input()
        installation_id, _ = discover_installation(app_id, pem_path, args.repo)
        if not installation_id:
            raise SystemExit("still not installed; re-run once it is installed")
    print("Installation id: {}".format(installation_id))

    _, env_path = write_secure_config(config_dir, pem_path.read_text(), app_id, installation_id, slug)
    expires, prefix = verify_token(app_id, pem_path, installation_id)
    print("Verified: test token {}… expires {}".format(prefix, expires))
    print("Wrote {}".format(env_path))

    if not args.no_rulesets:
        print("\nSetting authorship variables…")
        set_authorship_variables(args.repo, slug)
        print("\nApplying branch rulesets…")
        result = apply_rulesets(args.repo, bypass_user_id)
        if result.returncode != 0:
            print("Ruleset script failed; run it manually: bash scripts/apply-github-rulesets.sh", file=sys.stderr)

    print("\nEnsuring agent-main exists…")
    ensure_agent_main(args.repo, app_id, pem_path, installation_id)

    print(
        "\nDone. For each agent session:\n"
        "  eval \"$(python3 scripts/github-app-token.py --export)\"   # or: python3 scripts/github-app-token.py --gh\n"
        "\nNote: installation tokens authenticate over HTTPS. If a worktree's remote is SSH\n"
        "(git@github.com:...), switch it to HTTPS so the app's pushes are attributed to the app:\n"
        "  git remote set-url origin https://github.com/{}.git\n".format(args.repo)
    )


if __name__ == "__main__":
    main()

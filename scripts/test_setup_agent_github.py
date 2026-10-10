"""Unit tests for scripts/setup-agent-github.py."""

import importlib.util
import json
import os
import pathlib
import socket
import sys
import tempfile
import unittest
from unittest import mock

HERE = pathlib.Path(__file__).resolve().parent
SCRIPT_PATH = HERE / "setup-agent-github.py"

spec = importlib.util.spec_from_file_location("setup_agent_github", SCRIPT_PATH)
setup_agent_github = importlib.util.module_from_spec(spec)
spec.loader.exec_module(setup_agent_github)


class DefaultRepoTests(unittest.TestCase):
    """DEFAULT_REPO must be a placeholder example with no real org."""

    def test_default_repo_is_example_placeholder(self):
        self.assertEqual(setup_agent_github.DEFAULT_REPO, "example/example-repo")

class SameFileHelperTests(unittest.TestCase):
    """Tests for the pure same_file(src, dst) helper."""

    def test_same_path_to_itself(self):
        with tempfile.NamedTemporaryFile(delete=False) as f:
            f.write(b"test")
            path = f.name
        try:
            self.assertTrue(setup_agent_github.same_file(path, path))
        finally:
            os.unlink(path)

    def test_two_references_to_same_file(self):
        with tempfile.NamedTemporaryFile(delete=False) as f:
            f.write(b"test")
            path1 = f.name
        path2 = path1 + "_link"
        try:
            os.symlink(path1, path2)
            self.assertTrue(setup_agent_github.same_file(path1, path2))
        finally:
            os.unlink(path1)
            try:
                os.unlink(path2)
            except FileNotFoundError:
                pass

    def test_different_files(self):
        with tempfile.NamedTemporaryFile(delete=False) as f1:
            f1.write(b"a")
            path1 = f1.name
        with tempfile.NamedTemporaryFile(delete=False) as f2:
            f2.write(b"b")
            path2 = f2.name
        try:
            self.assertFalse(setup_agent_github.same_file(path1, path2))
        finally:
            os.unlink(path1)
            os.unlink(path2)

    def test_nonexistent_dst_resolved_paths_equal(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            src = pathlib.Path(tmpdir) / "real.pem"
            src.write_text("key-data")
            dst = pathlib.Path(tmpdir) / "nonexistent" / ".." / "real.pem"
            self.assertFalse(dst.exists())
            self.assertTrue(setup_agent_github.same_file(src, dst))

    def test_nonexistent_dst_resolved_paths_different(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            src = pathlib.Path(tmpdir) / "real.pem"
            src.write_text("key-data")
            dst = pathlib.Path(tmpdir) / "other.pem"
            self.assertFalse(dst.exists())
            self.assertFalse(setup_agent_github.same_file(src, dst))


class StoreExistingKeyTests(unittest.TestCase):
    """Tests for store_existing_key same-file no-op behaviour."""

    def _setup_same_file_fixture(self, tmpdir, initial_mode):
        config_dir = pathlib.Path(tmpdir) / "config"
        config_dir.mkdir()
        pem_path = config_dir / "github-app.pem"
        pem_path.write_text("key-data")
        os.chmod(pem_path, initial_mode)
        return config_dir, pem_path

    def test_same_file_returns_path_without_copying(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir, pem_path = self._setup_same_file_fixture(tmpdir, 0o600)
            result = setup_agent_github.store_existing_key(config_dir, pem_path)
            self.assertEqual(result, pem_path)
            self.assertEqual(os.stat(pem_path).st_mode & 0o777, 0o600)

    def test_same_file_ensures_mode_0600(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir, pem_path = self._setup_same_file_fixture(tmpdir, 0o644)
            result = setup_agent_github.store_existing_key(config_dir, pem_path)
            self.assertEqual(result, pem_path)
            self.assertEqual(os.stat(pem_path).st_mode & 0o777, 0o600)

    def test_different_file_copies_and_sets_mode(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir = pathlib.Path(tmpdir) / "config"
            config_dir.mkdir()
            src = pathlib.Path(tmpdir) / "source.pem"
            src.write_text("key-data")
            os.chmod(src, 0o644)
            result = setup_agent_github.store_existing_key(config_dir, src)
            self.assertEqual(result, config_dir / "github-app.pem")
            self.assertTrue(result.exists())
            self.assertEqual(result.read_text(), "key-data")
            self.assertEqual(os.stat(result).st_mode & 0o777, 0o600)


class BotIdentityHelperTests(unittest.TestCase):
    """Tests for the pure bot login/email/lookup-path construction helpers."""

    def test_bot_login_appends_bot_suffix(self):
        self.assertEqual(
            setup_agent_github.bot_login("example-agents"),
            "example-agents[bot]",
        )

    def test_bot_email_uses_bot_user_id_and_slug(self):
        self.assertEqual(
            setup_agent_github.bot_email("12345", "example-agents"),
            "12345+example-agents[bot]@users.noreply.github.com",
        )

    def test_bot_user_path_encodes_brackets(self):
        self.assertEqual(
            setup_agent_github.bot_user_path("example-agents"),
            "/users/example-agents%5Bbot%5D",
        )


class ResolveAppSlugTests(unittest.TestCase):
    """Tests for resolve_app_slug's GET /app lookup and failure behaviour."""

    def test_success_returns_slug_and_gets_app_with_jwt(self):
        with mock.patch.object(
            setup_agent_github.ghapp, "make_jwt", return_value="jwt"
        ) as make_jwt, mock.patch.object(
            setup_agent_github.ghapp, "api", return_value={"slug": "example-agents"}
        ) as api:
            slug = setup_agent_github.resolve_app_slug("123", "/tmp/app.pem")
        self.assertEqual(slug, "example-agents")
        make_jwt.assert_called_once_with("123", "/tmp/app.pem")
        api.assert_called_once_with("/app", token="jwt")

    def test_api_failure_raises_actionable_system_exit(self):
        with mock.patch.object(
            setup_agent_github.ghapp, "make_jwt", return_value="jwt"
        ), mock.patch.object(
            setup_agent_github.ghapp,
            "api",
            side_effect=SystemExit("GitHub API GET /app failed: 404"),
        ):
            with self.assertRaises(SystemExit) as caught:
                setup_agent_github.resolve_app_slug("123", "/tmp/app.pem")
        message = str(caught.exception)
        self.assertIn("123", message)
        self.assertIn("GET /app", message)

    def test_missing_slug_raises_actionable_system_exit(self):
        with mock.patch.object(
            setup_agent_github.ghapp, "make_jwt", return_value="jwt"
        ), mock.patch.object(
            setup_agent_github.ghapp, "api", return_value={"id": 123}
        ):
            with self.assertRaises(SystemExit) as caught:
                setup_agent_github.resolve_app_slug("123", "/tmp/app.pem")
        message = str(caught.exception)
        self.assertIn("123", message)
        self.assertIn("GET /app", message)

    def test_transport_and_decode_failures_raise_actionable_system_exit(self):
        # ghapp.api may raise OSError/URLError (transport) or ValueError (JSON
        # decoding) rather than SystemExit; both must become the same actionable
        # SystemExit instead of a raw traceback.
        for exc_type in (OSError, ValueError):
            with self.subTest(exc_type=exc_type.__name__):
                with mock.patch.object(
                    setup_agent_github.ghapp, "make_jwt", return_value="jwt"
                ), mock.patch.object(
                    setup_agent_github.ghapp, "api", side_effect=exc_type("boom")
                ):
                    with self.assertRaises(SystemExit) as caught:
                        setup_agent_github.resolve_app_slug("123", "/tmp/app.pem")
                message = str(caught.exception)
                self.assertIn("123", message)
                self.assertIn("GET /app", message)


class LocalManifestUrlTests(unittest.TestCase):
    """Tests for the pure local_manifest_url(port) helper (A1)."""

    def test_builds_loopback_url_for_port(self):
        self.assertEqual(
            setup_agent_github.local_manifest_url(8976),
            "http://127.0.0.1:8976/",
        )

    def test_uses_the_given_port_for_the_loopback_url(self):
        self.assertEqual(
            setup_agent_github.local_manifest_url(12345),
            "http://127.0.0.1:12345/",
        )


class ManifestFormHtmlTests(unittest.TestCase):
    """Regression tests: the form still posts to GitHub's action URL."""

    def test_posts_to_github_action_url_with_escaped_manifest(self):
        manifest = {"name": 'a<b>&"c'}
        html_out = setup_agent_github.manifest_form_html(
            manifest, "st+ate", None
        )
        action = setup_agent_github.manifest_action_url(None)
        self.assertIn("action='{}?state=st%2Bate'".format(action), html_out)
        escaped = setup_agent_github.html.escape(
            json.dumps(manifest), quote=True
        )
        self.assertIn(escaped, html_out)
        self.assertIn("name='manifest' value=\"{}\"".format(escaped), html_out)

    def test_auto_submits_the_form(self):
        html_out = setup_agent_github.manifest_form_html(
            {"name": "x"}, "s", None
        )
        self.assertIn("<body onload='document.forms[0].submit()'>", html_out)

    def test_org_manifest_targets_the_org_action_url(self):
        html_out = setup_agent_github.manifest_form_html(
            {"name": "x"}, "s", "acme"
        )
        self.assertIn(setup_agent_github.manifest_action_url("acme"), html_out)


class LoopbackErrorMessageTests(unittest.TestCase):
    """Tests for the actionable loopback_error_message(port, reason) (A1)."""

    def test_names_local_url_and_reason(self):
        message = setup_agent_github.loopback_error_message(8976, "timed out")
        self.assertIn("http://127.0.0.1:8976/", message)
        self.assertIn("timed out", message)

    def test_names_the_given_reason(self):
        message = setup_agent_github.loopback_error_message(
            1, "state mismatch"
        )
        self.assertIn("http://127.0.0.1:1/", message)
        self.assertIn("state mismatch", message)

    def test_explains_the_loopback_requirement(self):
        message = setup_agent_github.loopback_error_message(8976, "timed out")
        self.assertIn("127.0.0.1:8976", message)


class _TrackingHTTPServer(setup_agent_github.http.server.ThreadingHTTPServer):
    """A real loopback server that records itself so tests can close it."""

    instances = []

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        _TrackingHTTPServer.instances.append(self)


class CreateAppViaManifestLoopbackTests(unittest.TestCase):
    """The manifest flow opens the local page and exits on a bad loopback."""

    def _free_port(self):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as probe:
            probe.bind(("127.0.0.1", 0))
            return probe.getsockname()[1]

    def test_opens_local_page_and_exits_non_zero_on_timeout(self):
        port = self._free_port()
        opened = []
        _TrackingHTTPServer.instances = []
        setup_agent_github.CallbackHandler.result = {}
        setup_agent_github.CallbackHandler.expected_state = None
        try:
            with mock.patch.object(
                setup_agent_github.webbrowser,
                "open",
                side_effect=lambda url: opened.append(url),
            ), mock.patch.object(
                setup_agent_github.http.server,
                "ThreadingHTTPServer",
                _TrackingHTTPServer,
            ):
                with self.assertRaises(SystemExit) as caught:
                    setup_agent_github.create_app_via_manifest(
                        "example-agents",
                        "owner/repo",
                        port,
                        False,
                        None,
                        timeout=0,
                    )
        finally:
            for server in _TrackingHTTPServer.instances:
                server.server_close()

        local_url = "http://127.0.0.1:{}/".format(port)
        self.assertEqual(opened, [local_url])
        message = str(caught.exception)
        self.assertIn(local_url, message)
        self.assertIn("timed out", message)
        self.assertNotEqual(caught.exception.code, 0)


class _FakeCompleted:
    """A minimal stand-in for subprocess.CompletedProcess."""

    def __init__(self, returncode=0, stdout="", stderr=""):
        self.returncode = returncode
        self.stdout = stdout
        self.stderr = stderr


def _join(cmd):
    return " ".join(str(part) for part in cmd)


class NoDriftClassifierTests(unittest.TestCase):
    """Pure A3 classifier: no drift means main reachable from agent-main (R2)."""

    def test_identical_is_no_drift(self):
        self.assertTrue(setup_agent_github.is_no_drift("identical"))

    def test_behind_is_no_drift(self):
        self.assertTrue(setup_agent_github.is_no_drift("behind"))

    def test_ahead_is_drift(self):
        self.assertFalse(setup_agent_github.is_no_drift("ahead"))

    def test_diverged_is_drift(self):
        self.assertFalse(setup_agent_github.is_no_drift("diverged"))


class DriftReportTests(unittest.TestCase):
    """Pure A3 drift_report(compare_json) names the unreachable main commits."""

    def test_names_unreachable_main_commit_shas_and_main_sync_pr(self):
        compare_json = {
            "status": "ahead",
            "commits": [
                {"sha": "1111111111111111111111111111111111111111"},
                {"sha": "2222222222222222222222222222222222222222"},
            ],
        }
        report = setup_agent_github.drift_report(compare_json)
        self.assertIn("1111111111111111111111111111111111111111", report)
        self.assertIn("2222222222222222222222222222222222222222", report)
        self.assertIn("main-sync PR", report)

    def test_handles_no_commits_gracefully(self):
        report = setup_agent_github.drift_report({"status": "ahead", "commits": []})
        self.assertIsInstance(report, str)
        self.assertIn("main-sync PR", report)


class EnsureAgentMainTests(unittest.TestCase):
    """A3: ensure_agent_main is create-only and checks the ref first."""

    def _patch_ghapp(self):
        return (
            mock.patch.object(setup_agent_github.ghapp, "make_jwt", return_value="jwt"),
            mock.patch.object(
                setup_agent_github.ghapp, "api", return_value={"token": "tok"}
            ),
        )

    def test_existing_ref_is_left_untouched(self):
        calls = []

        def handler(cmd, **kwargs):
            calls.append(cmd)
            joined = _join(cmd)
            if "git/ref/heads/agent-main" in joined:
                return _FakeCompleted(0, "abc123")
            if "git/ref/heads/main" in joined:
                return _FakeCompleted(0, "mainsha")
            return _FakeCompleted(0, "")

        key_patch, api_patch = self._patch_ghapp()
        with key_patch, api_patch, mock.patch.object(
            setup_agent_github, "run", side_effect=handler
        ):
            setup_agent_github.ensure_agent_main(
                "owner/repo", "1", "/tmp/app.pem", "42"
            )

        joined_calls = [_join(c) for c in calls]
        self.assertFalse(
            any("POST" in c and "git/refs" in c for c in joined_calls),
            "an existing agent-main must never be created/re-written",
        )
        self.assertFalse(
            any("git/ref/heads/main" in c for c in joined_calls),
            "main must not be resolved once the existing ref is found",
        )

    def test_missing_ref_is_created_from_main_tip(self):
        calls = []
        main_sha = "abcdef1234567890"

        def handler(cmd, **kwargs):
            calls.append(cmd)
            joined = _join(cmd)
            if "git/ref/heads/agent-main" in joined:
                return _FakeCompleted(1, "", "not found")
            if "git/ref/heads/main" in joined:
                return _FakeCompleted(0, main_sha)
            return _FakeCompleted(0, "")

        key_patch, api_patch = self._patch_ghapp()
        with key_patch, api_patch, mock.patch.object(
            setup_agent_github, "run", side_effect=handler
        ):
            setup_agent_github.ensure_agent_main(
                "owner/repo", "1", "/tmp/app.pem", "42"
            )

        joined_calls = [_join(c) for c in calls]
        create_calls = [c for c in calls if "git/refs" in _join(c)]
        self.assertEqual(len(create_calls), 1)
        create_cmd = _join(create_calls[0])
        self.assertIn("POST", create_cmd)
        self.assertIn("ref=refs/heads/agent-main", create_cmd)
        self.assertIn("sha={}".format(main_sha), create_cmd)

        agent_idx = next(
            i for i, c in enumerate(joined_calls) if "git/ref/heads/agent-main" in c
        )
        main_idx = next(
            i for i, c in enumerate(joined_calls) if "git/ref/heads/main" in c
        )
        create_idx = next(
            i for i, c in enumerate(joined_calls) if "git/refs" in c
        )
        self.assertLess(agent_idx, main_idx)
        self.assertLess(main_idx, create_idx)


class CheckSetupDriftTests(unittest.TestCase):
    """A3: check_setup uses the compare API and flags main-unreachable drift."""

    REPO = "owner/repo"

    def _run_with_compare(self, compare_result):
        calls = []

        def handler(cmd, **kwargs):
            calls.append(cmd)
            joined = _join(cmd)
            if "/rulesets" in joined:
                return _FakeCompleted(0, "agent-main-protection,main-protection")
            if "git/ref/heads/agent-main" in joined:
                return _FakeCompleted(0, "abc123")
            if "/compare/agent-main...main" in joined:
                return compare_result
            if "variable" in joined:
                return _FakeCompleted(0, "AGENT_BOT_EMAIL\nAGENT_BOT_LOGIN")
            return _FakeCompleted(0, "")

        return calls, handler

    def test_drifted_compare_appends_problem_naming_commits_and_sync_pr(self):
        sha = "1234567890abcdef1234567890abcdef12345678"
        compare = _FakeCompleted(
            0,
            json.dumps(
                {"status": "ahead", "commits": [{"sha": sha}]}
            ),
        )
        calls, handler = self._run_with_compare(compare)
        with mock.patch.object(setup_agent_github, "run", side_effect=handler):
            problems = setup_agent_github.check_setup(self.REPO)

        self.assertTrue(
            any(sha in p and "main-sync PR" in p for p in problems),
            "drift problem must name the unreachable commit and the main-sync PR",
        )
        self.assertTrue(
            any("/compare/agent-main...main" in _join(c) for c in calls),
            "check_setup must use the GitHub compare API",
        )

    def test_no_drift_compare_yields_no_drift_problem(self):
        compare = _FakeCompleted(
            0, json.dumps({"status": "behind", "commits": []})
        )
        calls, handler = self._run_with_compare(compare)
        with mock.patch.object(setup_agent_github, "run", side_effect=handler):
            problems = setup_agent_github.check_setup(self.REPO)

        self.assertFalse(
            any("main-sync PR" in p or "drift" in p.lower() for p in problems)
        )

    def test_compare_nonzero_yields_actionable_problem(self):
        compare = _FakeCompleted(1, "", "api error")
        calls, handler = self._run_with_compare(compare)
        with mock.patch.object(setup_agent_github, "run", side_effect=handler):
            problems = setup_agent_github.check_setup(self.REPO)

        self.assertTrue(
            any("could not compare" in p for p in problems),
            "non-zero compare must yield an actionable problem",
        )

    def test_compare_nonjson_stdout_yields_actionable_problem(self):
        compare = _FakeCompleted(0, "not-json")
        calls, handler = self._run_with_compare(compare)
        with mock.patch.object(setup_agent_github, "run", side_effect=handler):
            problems = setup_agent_github.check_setup(self.REPO)

        self.assertTrue(
            any("could not compare" in p for p in problems),
            "non-JSON compare stdout must yield an actionable problem",
        )


class ApplyRulesetsTests(unittest.TestCase):
    """apply_rulesets must delegate the human bypass user id (F1.1)."""

    def _capture_run(self):
        captured = {}

        def handler(cmd, **kwargs):
            captured["cmd"] = cmd
            captured["env"] = kwargs.get("env")
            return _FakeCompleted(0, "", "")

        return captured, handler

    def test_passes_bypass_user_id_into_delegated_env(self):
        captured, handler = self._capture_run()
        with mock.patch.object(setup_agent_github, "run", side_effect=handler):
            result = setup_agent_github.apply_rulesets("owner/repo", "680936")

        self.assertEqual(result.returncode, 0)
        self.assertEqual(captured["env"]["REPO"], "owner/repo")
        self.assertEqual(captured["env"]["BYPASS_USER_ID"], "680936")
        self.assertTrue(
            any("apply-github-rulesets.sh" in str(part) for part in captured["cmd"])
        )

    def test_omits_bypass_user_id_from_env_when_none(self):
        captured, handler = self._capture_run()
        with mock.patch.dict(os.environ, {}, clear=False):
            os.environ.pop("BYPASS_USER_ID", None)
            with mock.patch.object(setup_agent_github, "run", side_effect=handler):
                setup_agent_github.apply_rulesets("owner/repo", None)

        self.assertEqual(captured["env"]["REPO"], "owner/repo")
        self.assertNotIn("BYPASS_USER_ID", captured["env"])


class ResolveBypassUserIdTests(unittest.TestCase):
    """resolve_bypass_user_id validates and coerces the raw id (F1.1)."""

    def test_none_is_unset(self):
        self.assertIsNone(setup_agent_github.resolve_bypass_user_id(None))

    def test_blank_is_unset(self):
        self.assertIsNone(setup_agent_github.resolve_bypass_user_id(""))
        self.assertIsNone(setup_agent_github.resolve_bypass_user_id("   "))

    def test_positive_integer_is_returned_as_int(self):
        self.assertEqual(
            setup_agent_github.resolve_bypass_user_id("680936"), 680936
        )
        self.assertEqual(setup_agent_github.resolve_bypass_user_id(42), 42)

    def test_non_positive_or_non_numeric_is_actionable(self):
        for bad in ("0", "-3", "abc", "12.5", " 0 "):
            with self.subTest(bad=bad):
                with self.assertRaises(SystemExit) as caught:
                    setup_agent_github.resolve_bypass_user_id(bad)
                message = str(caught.exception)
                self.assertIn("--bypass-user-id", message)
                self.assertIn("BYPASS_USER_ID", message)
                self.assertIn("--no-rulesets", message)


class MainBypassUserIdTests(unittest.TestCase):
    """main() resolves/validates the bypass id before any side effects (F1.1)."""

    def _run_main(self, extra_args, env_patch=None, observed=None):
        """Run main() in existing-app mode with all side effects mocked.

        ``observed`` (a dict) receives the mocks before main() runs, so they can
        still be asserted when main() raises SystemExit.
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            pem = pathlib.Path(tmpdir) / "app.pem"
            pem.write_text("key-data")
            config_dir = pathlib.Path(tmpdir) / "config"
            argv = [
                "setup-agent-github.py",
                "--existing-app",
                "--app-id", "1",
                "--private-key", str(pem),
                "--slug", "example-agents",
                "--installation-id", "42",
                "--config-dir", str(config_dir),
                "--repo", "owner/repo",
            ] + list(extra_args)
            apply_mock = mock.MagicMock(return_value=_FakeCompleted(0, "", ""))
            set_mock = mock.MagicMock()
            ensure_mock = mock.MagicMock()
            store_mock = mock.MagicMock(return_value=pem)
            write_mock = mock.MagicMock(
                return_value=(pem, config_dir / "agent-github.env")
            )
            if observed is not None:
                observed.update(
                    apply=apply_mock,
                    set_vars=set_mock,
                    ensure=ensure_mock,
                    store=store_mock,
                    write=write_mock,
                )
            with mock.patch.dict(os.environ, {}, clear=False):
                os.environ.pop("BYPASS_USER_ID", None)
                if env_patch:
                    os.environ.update(env_patch)
                with mock.patch.object(sys, "argv", argv), mock.patch.object(
                    setup_agent_github, "check_prerequisites", return_value=[]
                ), mock.patch.object(
                    setup_agent_github, "verify_token", return_value=("exp", "tok")
                ), mock.patch.object(
                    setup_agent_github, "set_authorship_variables", set_mock
                ), mock.patch.object(
                    setup_agent_github, "ensure_agent_main", ensure_mock
                ), mock.patch.object(
                    setup_agent_github, "apply_rulesets", apply_mock
                ), mock.patch.object(
                    setup_agent_github, "store_existing_key", store_mock
                ), mock.patch.object(
                    setup_agent_github, "write_secure_config", write_mock
                ):
                    setup_agent_github.main()
            return apply_mock, set_mock

    def _assert_actionable(self, exc):
        message = str(exc)
        self.assertIn("--bypass-user-id", message)
        self.assertIn("BYPASS_USER_ID", message)
        self.assertIn("--no-rulesets", message)

    def test_missing_id_fails_before_any_side_effects(self):
        observed = {}
        with self.assertRaises(SystemExit) as caught:
            self._run_main([], observed=observed)
        self._assert_actionable(caught.exception)
        observed["apply"].assert_not_called()
        observed["set_vars"].assert_not_called()
        observed["store"].assert_not_called()
        observed["write"].assert_not_called()

    def test_explicit_zero_id_fails_actionably(self):
        observed = {}
        with self.assertRaises(SystemExit) as caught:
            self._run_main(["--bypass-user-id", "0"], observed=observed)
        self._assert_actionable(caught.exception)
        observed["apply"].assert_not_called()
        observed["store"].assert_not_called()

    def test_non_numeric_env_id_fails_actionably(self):
        with self.assertRaises(SystemExit) as caught:
            self._run_main([], env_patch={"BYPASS_USER_ID": "abc"})
        self._assert_actionable(caught.exception)

    def test_env_var_fallback_supplies_bypass_user_id(self):
        observed = {}
        self._run_main(
            [], env_patch={"BYPASS_USER_ID": "680936"}, observed=observed
        )
        observed["apply"].assert_called_once()
        args, _ = observed["apply"].call_args
        self.assertEqual(args[0], "owner/repo")
        self.assertEqual(args[1], 680936)

    def test_cli_flag_overrides_env_var(self):
        observed = {}
        self._run_main(
            ["--bypass-user-id", "42"],
            env_patch={"BYPASS_USER_ID": "680936"},
            observed=observed,
        )
        self.assertEqual(observed["apply"].call_args[0][1], 42)

    def test_no_rulesets_completes_without_id_and_skips_apply(self):
        observed = {}
        self._run_main(["--no-rulesets"], observed=observed)
        observed["apply"].assert_not_called()
        observed["set_vars"].assert_not_called()
        observed["ensure"].assert_called_once()


if __name__ == "__main__":
    unittest.main()

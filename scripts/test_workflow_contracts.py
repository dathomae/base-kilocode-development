"""Regression tests for the protected milestone workflow documentation."""

from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[1]


class WorkflowContractTests(unittest.TestCase):
    def test_story_implementation_has_no_main_branch_escape_hatch(self) -> None:
        script = (ROOT / "scripts/story-implementation.pdd.script.md").read_text(encoding="utf-8")
        self.assertIn("no agent-side or\n   user-confirmation override", script)
        self.assertNotIn("Do not proceed without explicit user confirmation", script)

    def test_project_setup_covers_all_composition_types(self) -> None:
        guide = (ROOT / "resources/extensions/project-setup.md").read_text(encoding="utf-8")
        for heading in (
            "### Core-only software",
            "### Hardware/manufacturing",
            "### Complex system",
            "### Combined hardware and complex system",
        ):
            self.assertIn(heading, guide)
        self.assertIn("mandatory before an agent implements a milestone", guide)

    def test_agent_setup_creates_agent_main_before_selecting_it(self) -> None:
        guide = (ROOT / "AGENT_REPOSITORY_SETUP.md").read_text(encoding="utf-8")
        configure = guide.index("## 2. Create or reuse the GitHub App")
        select_branch = guide.index("## 3. Fetch and select the integration branch")
        self.assertLess(configure, select_branch)
        self.assertIn("mandatory\nbefore agents implement milestones", guide)


if __name__ == "__main__":
    unittest.main()

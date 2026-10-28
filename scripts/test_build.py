"""Exercise build and package rollback paths with controlled tool failures."""
import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


class BuildFailureTests(unittest.TestCase):
    def check_failure(self, stage):
        root = Path(__file__).resolve().parent.parent
        task = Path(os.environ.get("INLINEEXPLAIN_TEST_DIR", Path.home() / "Documents/tmp/inlineExplain"))
        task.mkdir(parents=True, exist_ok=True)
        with tempfile.TemporaryDirectory(prefix="build-regression-", dir=task) as temporary:
            fixture = Path(temporary) / "project with spaces"
            shutil.copytree(root, fixture, ignore=shutil.ignore_patterns("build", ".git", "__pycache__"))
            app = fixture / "build/InlineExplain.app"
            binary = app / "Contents/MacOS/inlineexplain"
            binary.parent.mkdir(parents=True)
            binary.write_text("last working executable")
            (app / "Contents/Info.plist").write_bytes((root / "packaging/Info.plist").read_bytes())
            before = {str(p.relative_to(app)): hashlib.sha256(p.read_bytes()).hexdigest() for p in app.rglob("*") if p.is_file()}
            tools = Path(temporary) / "tools"
            tools.mkdir()
            commands = {
                "go": """#!/bin/zsh
if [[ "$1" == build ]]; then
  while (( $# )); do
    if [[ "$1" == -o ]]; then shift; print -r -- 'new executable' > "$1"; exit 0; fi
    shift
  done
fi
exit 0
""",
                "xcrun": """#!/bin/zsh
if [[ "$INLINEEXPLAIN_BUILD_FAILURE" == helper ]]; then print -u2 injected-helper; exit 17; fi
while (( $# )); do
  if [[ "$1" == -o ]]; then shift; print -r -- 'new helper' > "$1"; exit 0; fi
  shift
done
exit 0
""",
                "codesign": "#!/bin/zsh\nprint -u2 injected-signing\nexit 18\n",
            }
            for name, content in commands.items():
                command = tools / name
                command.write_text(content)
                command.chmod(0o755)
            environment = dict(os.environ, PATH=str(tools) + os.pathsep + os.environ["PATH"], INLINEEXPLAIN_GO=str(tools / "go"), INLINEEXPLAIN_BUILD_FAILURE=stage)
            result = subprocess.run(["zsh", "scripts/build.sh"], cwd=fixture, env=environment, capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0, "The injected build failure must be observed")
            self.assertIn("injected-" + stage, result.stderr, "The test must reach the selected failure stage")
            after = {str(p.relative_to(app)): hashlib.sha256(p.read_bytes()).hexdigest() for p in app.rglob("*") if p.is_file()}
            self.assertEqual(before, after, "A failed build replaced the previous working app")
            self.assertFalse(list((fixture / "build").glob(".InlineExplain-build.*")), "Failed staging folders must be cleaned")

    def test_helper_failure_preserves_previous_app(self):
        self.check_failure("helper")

    def test_signing_failure_preserves_previous_app(self):
        self.check_failure("signing")


class PackageFailureTests(unittest.TestCase):
    def check_failure(self, stage):
        root = Path(__file__).resolve().parent.parent
        task = Path(os.environ.get("INLINEEXPLAIN_TEST_DIR", Path.home() / "Documents/tmp/inlineExplain"))
        task.mkdir(parents=True, exist_ok=True)
        with tempfile.TemporaryDirectory(prefix="package-regression-", dir=task) as temporary:
            fixture = Path(temporary) / "project with spaces"
            shutil.copytree(root, fixture, ignore=shutil.ignore_patterns("build", ".git", "__pycache__"))
            app = fixture / "build/InlineExplain.app"
            app.mkdir(parents=True)
            (app / "fixture").write_text("The build stage is tested separately.")
            (fixture / "scripts/build.sh").write_text("#!/bin/zsh\nexit 0\n")
            release = fixture / "build/release"
            release.mkdir()
            (release / "previous.dmg").write_bytes(b"previous verified release")
            before = {p.name: p.read_bytes() for p in release.iterdir()}
            tools = Path(temporary) / "tools"
            tools.mkdir()
            hdiutil = tools / "hdiutil"
            hdiutil.write_text("""#!/bin/zsh
case "$1" in
  create)
    if [[ "$INLINEEXPLAIN_PACKAGE_FAILURE" == create ]]; then print -u2 injected-create; exit 19; fi
    print -r -- 'unverified image' > "${@[-1]}"
    ;;
  verify) print -u2 injected-verify; exit 20 ;;
  *) exit 21 ;;
esac
""")
            hdiutil.chmod(0o755)
            environment = dict(os.environ, PATH=str(tools) + os.pathsep + os.environ["PATH"], INLINEEXPLAIN_PACKAGE_FAILURE=stage)
            result = subprocess.run([str(fixture / "scripts/package.sh")], cwd=temporary, env=environment, capture_output=True, text=True)
            self.assertNotEqual(result.returncode, 0, "The injected package failure must be observed")
            self.assertIn("injected-" + stage, result.stderr, "The test must reach the selected failure stage")
            self.assertEqual(before, {p.name: p.read_bytes() for p in release.iterdir()}, "A failed package replaced the previous release")
            self.assertFalse(list((fixture / "build").glob(".InlineExplain-package.*")), "Failed package staging must be cleaned")

    def test_creation_failure_preserves_previous_release(self):
        self.check_failure("create")

    def test_verification_failure_preserves_previous_release(self):
        self.check_failure("verify")


if __name__ == "__main__":
    unittest.main()

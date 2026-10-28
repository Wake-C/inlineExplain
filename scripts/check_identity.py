"""Check the installed product identity and permission-facing app names."""
import argparse
import pathlib
import plistlib
import re
import subprocess
import sys


def check_source(root):
    with (root / "packaging/Info.plist").open("rb") as stream:
        info = plistlib.load(stream)
    check_info(info)
    legacy = re.compile(
        r"(?:allow(?:ing)?|reopen|quit|launch|start)\s+Peekturn\b"
        r"|Peekturn\s+(?:Desktop|Settings)"
        r"|<title>Peekturn</title>",
        re.IGNORECASE,
    )
    for path in (root / "frontend").iterdir():
        if path.suffix in {".html", ".js", ".css"}:
            assert not legacy.search(path.read_text()), f"Old product name in {path.relative_to(root)}"
    return info


def check_info(info):
    assert info["CFBundleName"] == "InlineExplain", "Wrong bundle name"
    assert info["CFBundleDisplayName"] == "InlineExplain", "Wrong display name"
    assert info["CFBundleIdentifier"] == "com.inlineexplain.desktop", "Wrong bundle identifier"
    assert info["CFBundleExecutable"] == "inlineexplain", "Wrong executable name"
    assert re.fullmatch(r"\d+\.\d+\.\d+", info["CFBundleShortVersionString"]), "Use a three-part release version"
    assert re.fullmatch(r"[1-9]\d*", info["CFBundleVersion"]), "Use a positive build number"
    assert info["CFBundleIconFile"] == "InlineExplain.icns", "Wrong bundle icon"
    assert info["LSMinimumSystemVersion"] == "15.0", "Wrong minimum macOS version"


def check_app(path, expected):
    with (path / "Contents/Info.plist").open("rb") as stream:
        info = plistlib.load(stream)
    check_info(info)
    for key in ("CFBundleShortVersionString", "CFBundleVersion", "CFBundleIconFile", "LSMinimumSystemVersion"):
        assert info[key] == expected[key], f"Packaged {key} differs from packaging/Info.plist"
    assert (path / "Contents/MacOS/InlineExplainOCR").is_file(), "Missing renamed OCR helper"
    assert not (path / "Contents/MacOS/PeekturnOCR").exists(), "Old OCR helper remains"
    executable = (path / "Contents/MacOS/inlineexplain").read_bytes()
    assert b"After allowing Peekturn" not in executable, "Old permission message remains in the app"
    assert b"After allowing InlineExplain" in executable, "Current permission message is not bundled"
    resources = path / "Contents/Resources"
    icon = (resources / info["CFBundleIconFile"]).read_bytes()
    assert icon[:4] == b"icns" and int.from_bytes(icon[4:8], "big") == len(icon) and len(icon) > 1024, "Invalid bundle icon"
    root = pathlib.Path(__file__).resolve().parent.parent
    assert (resources / "NOTICE").read_bytes() == (root / "NOTICE").read_bytes(), "Missing or mismatched component notices"
    if (root / "LICENSE").is_file():
        assert (resources / "LICENSE").read_bytes() == (root / "LICENSE").read_bytes(), "Missing or mismatched project license"
    for locale in ("en", "zh-Hans"):
        assert (resources / f"{locale}.lproj/InfoPlist.strings").is_file(), f"Missing permission localization: {locale}"
    for name in ("inlineexplain", "InlineExplainOCR"):
        architecture = subprocess.run(["lipo", "-archs", str(path / "Contents/MacOS" / name)], capture_output=True, text=True, check=True)
        assert architecture.stdout.strip() == "arm64", f"Wrong architecture for {name}"
    for binary, identifier in (
        (path, "com.inlineexplain.desktop"),
        (path / "Contents/MacOS/InlineExplainOCR", "com.inlineexplain.desktop.ocr"),
    ):
        result = subprocess.run(["codesign", "-dv", "--verbose=2", str(binary)], capture_output=True, text=True, check=True)
        assert f"Identifier={identifier}\n" in result.stderr, "Wrong code-signing identity"
    subprocess.run(["codesign", "--verify", "--deep", "--strict", str(path)], check=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--app", type=pathlib.Path)
    arguments = parser.parse_args()
    try:
        info = check_source(pathlib.Path(__file__).resolve().parent.parent)
        if arguments.app:
            check_app(arguments.app, info)
    except (AssertionError, KeyError, OSError, subprocess.CalledProcessError) as error:
        print(f"FAIL: {error}", file=sys.stderr)
        sys.exit(1)
    print("PASS: InlineExplain identity, version and package metadata" + (", ARM64 binaries, icon and signatures" if arguments.app else ""))

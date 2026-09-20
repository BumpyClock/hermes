"""Explicitly prepare the synthetic engine-side bundle with the selected local producer."""

import argparse
import importlib.util
import sys
from pathlib import Path


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--producer-root", type=Path, required=True)
    parser.add_argument("--producer-revision", required=True)
    args = parser.parse_args()
    tools = args.producer_root.absolute() / "scripts/bundle"
    sys.path.insert(0, str(tools))
    spec = importlib.util.spec_from_file_location("selected_producer", tools / "bundle.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    module.pinned_revision(args.producer_root, args.producer_revision)
    files = {
        "definitions/example.yaml": b"""schema: 1
site: synthetic-example
hosts: [example.com]
metadata:
  title:
    - text: h1
content:
  groups:
    - [article]
  default_cleaner: false
""",
        "fixtures/example/article.html": b"""<html><head><title>Generic decoy</title></head><body>
<h1>Synthetic candidate headline</h1><article><p>The synthetic report verifies the independent candidate protocol without fetching any article or DNS record.</p></article>
</body></html>
""",
    }
    files["conformance.json"] = module.encode({"protocol": 1, "cases": [{
        "id": "synthetic-example", "definition": "definitions/example.yaml",
        "fixture": "fixtures/example/article.html", "url": "https://www.example.com/report",
        "synthetic": True, "expect": {"title": "Synthetic candidate headline"},
        "exactly_once": ["The synthetic report verifies"], "exclude": ["Generic decoy"],
    }]})
    archive = module.archive_bytes(files)
    archive_hash = module.digest(archive)
    manifest = {
        "protocol": 1, "version": "synthetic-1", "scope": "pilot",
        "source": {"repository": "https://github.com/BumpyClock/hermes-definitions",
                   "revision": args.producer_revision, "state": "overlay",
                   "inputs_sha256": module.digest(module.encode(module.file_records(files)))},
        "definition_schema": 1,
        "engine": {"contract": module.CONTRACT, "operations": [
            "content.default_cleaner", "content.groups", "hosts.exact-www", "metadata.text"],
            "algorithms": []},
        "archive": {"name": f"definitions-synthetic-1-{archive_hash}.tar", "format": "ustar-v1",
                    "sha256": archive_hash, "size": len(archive)},
        "files": module.file_records(files),
    }
    output = Path(__file__).parent / "pinned"
    output.mkdir(exist_ok=True)
    (output / "bundle.tar").write_bytes(archive)
    manifest_data = module.encode(manifest)
    (output / "manifest.json").write_bytes(manifest_data)
    pins = {"archive_sha256": archive_hash, "manifest_sha256": module.digest(manifest_data)}
    (output / "pins.json").write_bytes(module.encode(pins))
    print(module.encode(pins).decode(), end="")


if __name__ == "__main__":
    main()

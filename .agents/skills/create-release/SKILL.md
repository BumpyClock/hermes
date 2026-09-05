---
name: create-release
description: Prepare, publish, or verify a Hermes Go module release. Also use to improve the repository release procedure.
---

# Create Hermes release

Read [the release procedure](../../../docs/RELEASING.md) before release work.
Use that document as the canonical procedure for preparation, publication, recovery, and evidence.

Require explicit authorization for the exact `v1.X.Y` version and publication before remote mutations.
Treat process improvements as local changes unless the user separately authorizes publication.
Preserve unrelated changes and report release blockers without unauthorized repairs.

Use `scripts/release_check.py` for preflight, prepared-commit, and CI checks.
Use `scripts/verify-release.sh` only for authorized public consumer checks.
Do not replace those checks with assumptions about local tags, cached modules, or another commit's CI result.

For interrupted publication, use the recovery table before any mutation.
Never replace a published tag.
Report completion only after the canonical evidence requirements pass.

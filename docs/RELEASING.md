# Release Hermes

This procedure publishes stable `v1.X.Y` releases of `github.com/BumpyClock/hermes` from `master`.
The default GitHub release has source archives and no binary assets.
A `v2` release requires a separate Go module path migration and compatibility plan.
This procedure does not authorize that migration.

## Prerequisites and authorization

Use Bash for the command examples.
Run each command only after the preceding command succeeds.
Install `python3`, `git`, `gh`, `file`, `strings`, `go`, `curl`, `make`, and `golangci-lint`.
Use Python 3.11 or later for release-script tests.
Use the Go version in `go.mod` and golangci-lint v2.13.2.
The linter build must support that Go toolchain.
Check `gh auth status` for the intended BumpyClock account.
Check that `origin` identifies `BumpyClock/hermes` and its default branch is `master`.

Obtain approval for the exact tag, release commit, branch push, tag push, and GitHub release creation.
Obtain approval for public Go proxy, checksum database, and pkg.go.dev queries.
An explicit request that covers these actions satisfies this requirement.
Approval for process improvements does not authorize publication.

Stop on unrelated changes, branch divergence, authentication failures, or uncertain remote state.
Never include unrelated files, such as `.claude/settings.local.json`, in the release commit.
Do not delete or regenerate tracked binaries without separate approval.
Never force-push, move, delete, or recreate a published tag.

## Now: Prepare the release

1. Inspect the release state and select a version.

   Read `AGENTS.md`, `go.mod`, `CHANGELOG.md`, `cmd/hermes/main.go`, `Makefile`, and `.github/workflows/ci.yml`.
   Inspect commits and public API changes since the latest remote stable tag.
   Use a patch version for compatible fixes and a minor version for compatible features or exported API additions.
   If compatibility is uncertain, ask the user before any version edit.
   Set the approved version:

   ```bash
   TAG=v1.X.Y
   VERSION=${TAG#v}
   python3 scripts/release_check.py preflight "$VERSION"
   ```

   The check requires a clean, synchronized `master` checkout and a newer, unused stable v1 tag.
   It checks the module, CLI version, changelog, and tracked binary version consistency.
   It reads remote state but does not publish or change repository files.
   A stale tracked binary blocks preparation until the user approves its removal or target-platform regeneration.
   After an approved binary repair, commit and synchronize that repair before a new preflight check.

2. Define the exact release paths and inspect additional version surfaces.

   ```bash
   RELEASE_PATHS=(CHANGELOG.md cmd/hermes/main.go)
   git grep -n -E 'v[0-9]+\.[0-9]+\.[0-9]+|[0-9]+\.[0-9]+\.[0-9]+' -- ':!go.sum'
   printf '%s\n' "${RELEASE_PATHS[@]}"
   ```

   Add each relevant version surface or approved binary artifact to `RELEASE_PATHS` explicitly.
   Keep historical changelog entries and dependency versions unchanged.
   Do not invent version files.

3. Prepare the approved version changes.

   Add a dated `## [X.Y.Z] - YYYY-MM-DD` section at the top of `CHANGELOG.md`.
   Describe externally visible changes and list exported API additions under `Added`.
   Set CLI version output to `Hermes $TAG` in `cmd/hermes/main.go`.
   Update the other approved release paths.
   Format each changed Go file with `golangci-lint fmt path/to/file.go`.
   Review the complete diff and preserve unrelated prior edits.
   If any changed path falls outside `RELEASE_PATHS`, stop before the stage operation.

4. Run the local acceptance gates.

   ```bash
   go mod download
   make verify
   test "$(./bin/hermes version | sed -n '1p')" = "Hermes $TAG"
   git diff --check
   git diff -- "${RELEASE_PATHS[@]}"
   git status --short
   ```

   `make verify` runs release-script tests, lint, race tests with `coverage.out`, and the CLI build.
   For changes with plausible performance effects, also run `make benchmark`.
   Record commands and results.
   If a gate fails, investigate its cause before a retry.
   Stop on failures outside the approved release scope.

5. Commit only the approved release paths.

   ```bash
   git add -- "${RELEASE_PATHS[@]}"
   git diff --cached --check
   git diff --cached --name-only
   git diff --cached -- "${RELEASE_PATHS[@]}"
   ```

   If the index contains other paths, stop before the commit.
   After index review, create the approved release commit:

   ```bash
   git commit -m "chore(release): prepare $TAG"
   RELEASE_COMMIT=$(git rev-parse HEAD)
   ```

   Record `RELEASE_COMMIT` outside the shell session for recovery.

## Later: Publish and verify

1. Push the approved release commit and require its successful CI result.

   Reconfirm publication authorization before remote mutations.
   Check that `master` still identifies `RELEASE_COMMIT` and the tree is clean.

   ```bash
   test "$(git branch --show-current)" = master
   test "$(git rev-parse HEAD)" = "$RELEASE_COMMIT"
   test -z "$(git status --porcelain)"
   git push origin master
   python3 scripts/release_check.py ready "$VERSION"
   python3 scripts/release_check.py ci "$RELEASE_COMMIT"
   ```

   The `ready` check requires synchronized `master` and target-version source, changelog, and binary consistency.
   The CI check requires the latest `ci.yml` push run on `master` for that exact commit to succeed.
   A successful pull request run or another commit's result does not satisfy this gate.
   For absent or active CI, check at intervals of at most 30 seconds for no more than ten minutes.
   Report status between checks.
   On a failed run, inspect `gh run view RUN_ID --log-failed` and stop publication.
   On timeout, retain the commit and resume later through the recovery table.

2. Create and push an annotated tag for the accepted commit.

   Recheck `ready` and `ci` immediately before tag creation.
   If any tag or release already exists, use the recovery table instead.

   ```bash
   python3 scripts/release_check.py ready "$VERSION"
   python3 scripts/release_check.py ci "$RELEASE_COMMIT"
   test "$(git rev-parse HEAD)" = "$RELEASE_COMMIT"
   git tag -a "$TAG" "$RELEASE_COMMIT" -m "Release $TAG"
   TAG_OBJECT=$(git rev-parse "$TAG")
   test "$(git cat-file -t "$TAG_OBJECT")" = tag
   test "$(git rev-parse "$TAG^{}")" = "$RELEASE_COMMIT"
   git push origin "refs/tags/$TAG"
   ```

   Record `TAG_OBJECT` outside the shell session.
   Run the tag identity checks below before release creation.
   Release policy makes published tags immutable but does not configure server-side tag protection.

3. Create the source-only GitHub release.

   Extract the approved version's changelog section into a temporary file:

   ```bash
   RELEASE_NOTES=$(mktemp "${TMPDIR:-/tmp}/hermes-release-notes.XXXXXX")
   awk -v heading="## [$VERSION] - " '
     index($0, heading) == 1 { found = 1; next }
     found && /^## / { exit }
     found { print; if (NF) content = 1 }
     END { if (!content) exit 1 }
   ' CHANGELOG.md > "$RELEASE_NOTES"
   test -s "$RELEASE_NOTES"
   cat "$RELEASE_NOTES"
   ```

   Stop if extraction fails or the file is empty.
   Review the notes against the approved version and actual changes before publication.
   Check `gh release view "$TAG" --repo BumpyClock/hermes` before creation.
   Treat failure as absence only when the diagnostic explicitly states that the release was not found.
   Stop on any other error.

   ```bash
   gh release create "$TAG" \
     --repo BumpyClock/hermes \
     --verify-tag \
     --target "$RELEASE_COMMIT" \
     --title "$TAG" \
     --notes-file "$RELEASE_NOTES"
   gh release view "$TAG" --repo BumpyClock/hermes --json url,tagName,isDraft,isPrerelease,assets
   ```

   Require the expected tag, a published stable release, and no assets unless separately approved.
   Do not edit or recreate an existing release during recovery.

4. Test the published Go module as an isolated consumer.

   ```bash
   bash scripts/verify-release.sh "$TAG"
   ```

   The script checks proxy artifacts and checksums with fresh temporary caches.
   It compiles a consumer outside this checkout and installs the tagged CLI.
   It asserts the installed CLI version.
   Record the retained temporary paths and command results.
   For propagation failures, retry at most five times with waits of at most 30 seconds.
   Investigate compile, checksum, or CLI version failures instead of unchanged retries.
   Never alter a tag to fix proxy propagation.
   The script also checks pkg.go.dev and reports its status separately.
   If documentation remains unavailable, retry only that query:

   ```bash
   curl --fail --silent --show-error --max-time 30 -I "https://pkg.go.dev/github.com/BumpyClock/hermes@$TAG"
   ```

   Apply the same five-attempt limit to documentation queries.
   Report a delayed documentation index as pending if module and checksum checks succeed.

5. Report release evidence and final repository state.

   Record the tag, commit SHA, annotated tag object SHA, remote peeled target, and GitHub release URL.
   Include local gate results, CI run URL, binary inventory, public consumer checks, and any approved artifact changes.
   Report `coverage.out`, temporary verification paths, proxy artifacts, checksum results, and documentation status.
   Check `git status --short` and equality of local and remote `master`.
   Do not declare completion before commit, tag identity, GitHub release, and public consumer checks succeed.

## Tag identity checks

Use the recorded release commit and annotated tag object for every continuation.
Fetch remote tag references only after the initial clean-tree check.
Do not overwrite an existing local tag if fetch reports a conflict.

```bash
test "$(git cat-file -t "$TAG_OBJECT")" = tag
test "$(git rev-parse "$TAG")" = "$TAG_OBJECT"
test "$(git rev-parse "$TAG^{}")" = "$RELEASE_COMMIT"
REMOTE_TAG_OBJECT=$(git ls-remote --tags origin "refs/tags/$TAG" | cut -f1)
REMOTE_TAG_TARGET=$(git ls-remote --tags origin "refs/tags/$TAG^{}" | cut -f1)
test "$REMOTE_TAG_OBJECT" = "$TAG_OBJECT"
test "$REMOTE_TAG_TARGET" = "$RELEASE_COMMIT"
```

Stop if any check fails or any expected identity lacks reliable evidence.
A matching tag name alone is insufficient.

## Recovery after partial publication

Inspect remote state before any retry after a network error.
Preserve successful artifacts and their recorded identities.
Require the original authorization and unchanged release contents before continuation.
Require a clean `master` tree and exact local and remote release commit agreement before tag or release continuation.
The new-release `preflight` and `ready` checks intentionally reject existing tags or releases.
Use the identity checks above for an existing tag instead of those new-release checks.

| Observed state | Safe next action |
| --- | --- |
| Local commit exists, branch push failed | Check remote `master`. Retry a normal push only if remote history permits it and the commit is unchanged. |
| Branch push succeeded, CI absent or active | Resume exact-commit CI checks within the stated budget. Do not create a tag. |
| CI failed | Inspect the failed run. Obtain approval for any correction outside release scope. Revalidate a changed commit before a tag. |
| Local annotated tag exists, remote tag absent | Check recorded object and commit identities, source state, and exact-commit CI. Push the existing tag without recreation. |
| Tag push result uncertain | Query remote object and peeled target. Continue only if both match recorded identities, or retry the existing tag if absent. |
| Remote tag exists, GitHub release absent | Validate tag identities and exact-commit CI. Create only the missing release with `--verify-tag`. |
| GitHub release already exists | Validate tag identities and published stable release metadata. Continue public consumer checks without a release edit. |
| Proxy or documentation propagation is delayed | Resume bounded public checks. Preserve the commit, tag, and release. |
| Commit, tag object, target, or release metadata differs | Stop. Do not replace published artifacts. Resolve the discrepancy before a separately approved next version. |

## Automation boundaries

`scripts/release_check.py` provides read-only release gates, not a publisher.
Each external command in that script has a 30-second timeout.
`scripts/verify-release.sh` makes public queries and writes only temporary consumer files and caches.
Neither script grants authorization or replaces API compatibility review, explicit path review, or release notes review.
Process-only changes require local tests and document checks, not an actual publication.

See the [GitHub CLI release reference](https://cli.github.com/manual/gh_release_create) for `--verify-tag` semantics.
See [Go module release guidance](https://go.dev/doc/modules/release-workflow) for module version and publication requirements.

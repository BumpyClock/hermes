---
name: create-release
description: Prepare and publish a Hermes Go module release end-to-end, including changelog, CLI version, GitHub release, and Go module publication verification. Use when asked to cut, tag, publish, or verify a Hermes release.
---

# Create Hermes release

Prepare and publish a release of the `github.com/BumpyClock/hermes` Go module. Work from the repository root. The release branch is `master`.

Use a conservative, stop-on-ambiguity workflow. Do not mutate tags, releases, branches, or remote state until the user has explicitly authorized the exact version and publication, unless that authorization is already unambiguous in the current request.

## 1. Inspect the release state

Read these files before choosing a version or editing:

```sh
cat AGENTS.md
cat go.mod
cat CHANGELOG.md
cat cmd/hermes/main.go
cat Makefile
git status --short
git remote -v
git branch --show-current
git remote show origin
git log --oneline --decorate -20
git tag --sort=-v:refname | head -20
```

Also inspect all tracked text version surfaces, including source, docs, packaging, workflows, and scripts. Exclude `go.sum`, which contains dependency versions rather than project release versions:

```sh
git grep -n -E 'v[0-9]+\.[0-9]+\.[0-9]+|[0-9]+\.[0-9]+\.[0-9]+' -- ':!go.sum'
```

Identify tracked executable artifacts with `git ls-files` and `file`. For every executable, PE, Mach-O, or ELF artifact, inspect embedded SemVer strings with `strings` when available:

```sh
git ls-files -z | while IFS= read -r -d '' path; do
  kind=$(file -b "$path")
  case "$kind" in
    *executable*|PE32*|PE32+*|Mach-O*|ELF*)
      printf '%s: %s\n' "$path" "$kind"
      if command -v strings >/dev/null; then
        strings "$path" | rg -n 'v[0-9]+\.[0-9]+\.[0-9]+' || true
      fi
      ;;
  esac
done
```

Stop if any tracked binary artifact embeds a version that disagrees with the proposed release version or source version surfaces. Do not silently delete binaries. Regenerate an artifact deliberately for its target platform, or obtain separate approval to remove it.

Confirm all of the following. Stop and report the first failure; do not repair unrelated release state without authorization.

- The repository is Hermes and `go.mod` declares `module github.com/BumpyClock/hermes`.
- The authenticated GitHub account is the intended BumpyClock account (`gh auth status`).
- `origin` is the intended Hermes remote and its default branch is `master`.
- The checked-out branch is `master`.
- `git status --short` is empty. Any existing change blocks the release. In particular, `.claude/settings.local.json` must be reported as a blocking, unrelated local change; never sweep it into a release commit.
- Only after the clean-tree check passes, `git fetch --prune --tags origin` succeeds and `git rev-list --left-right --count master...origin/master` returns exactly `0 0`. Do not release from a divergent or behind branch.

Use explicit checks:

```sh
git fetch --prune --tags origin
test "$(git rev-list --count master...origin/master)" = 0
```

Treat `gh release view` exit status 1 as "release absent" only after confirming its diagnostic says the release was not found. Any authentication, network, or other error is a stop condition.

## 2. Select and confirm the version

1. Find the latest **remote** stable SemVer tag, not merely the latest local tag. Exclude prereleases unless the user explicitly authorizes a prerelease workflow:

   ```sh
   git ls-remote --tags --refs origin 'v*' \
     | cut -f2 \
     | sed 's#refs/tags/##' \
     | grep -E '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$' \
     | sort -V \
     | tail -1
   ```

2. Inspect commits since that tag and all public API changes. For Go, exported additions and changes in the root package are public API. Use `git diff --stat`, `git log`, and focused inspection of exported identifiers.
3. Apply SemVer:
   - patch: backward-compatible fixes only;
   - minor: backward-compatible exported API additions or features;
   - major: breaking public API changes.
4. Normalize the selected numeric version to `X.Y.Z`, then establish `TAG=v$VERSION`. Reject malformed or prerelease values unless the user expressly requests a prerelease workflow:

   ```sh
   VERSION=X.Y.Z
   TAG="v$VERSION"
   git ls-remote --tags origin "refs/tags/$TAG"
   git show-ref --verify --quiet "refs/tags/$TAG" && printf 'local tag exists\n'
   gh release view "$TAG"
   ```

   Stop if that tag exists locally, remotely, or as a GitHub release. Treat `gh release view` exit status 1 as "release absent" only after confirming its diagnostic says the release was not found. Any authentication, network, or other error is a stop condition.
5. If commit intent or API compatibility makes the bump ambiguous, present the evidence and ask the user to choose before changing files.
6. Before any mutation, obtain confirmation of the exact `TAG` and whether the user authorizes push, tagging, GitHub release creation, public Go proxy and pkg.go.dev queries, and Go publication. If public queries are not explicitly authorized, ask before running them. An exact version plus this explicit authorization in the current request satisfies this confirmation.

Never treat an existing changelog heading, branch name, or local uncommitted version change as authority to publish that version.

## 3. Prepare the release commit

After version confirmation, define the exact intended release path set before editing or staging. It normally includes `CHANGELOG.md` and `cmd/hermes/main.go`, plus every discovered version surface and any explicitly approved regenerated binary artifact. Do not invent version files when none exists.

```sh
RELEASE_PATHS=(CHANGELOG.md cmd/hermes/main.go)
# Append each discovered version surface or approved regenerated artifact explicitly.
# RELEASE_PATHS+=(path/to/approved-artifact)
printf '%s\n' "${RELEASE_PATHS[@]}"
```

1. Add a dated Keep a Changelog section for `VERSION` at the top of `CHANGELOG.md`. Include externally visible changes. List all exported API additions in `Added`; do not hide public API changes under internal refactors.
2. Update the CLI version output in `cmd/hermes/main.go` to `Hermes $TAG`.
3. Update every discovered version surface in `RELEASE_PATHS`. If a tracked binary was stale, stop until the user explicitly authorizes target-platform regeneration or removal; add its path only after that authorization.
4. Format Go changes and review the complete intended set. `.golangci.yml` enables both `gofmt` and `goimports` (with the Hermes local prefix), so do not rely on `gofmt` alone:

   ```sh
   GO_RELEASE_PATHS=()
   for path in "${RELEASE_PATHS[@]}"; do
     case "$path" in *.go) GO_RELEASE_PATHS+=("$path");; esac
   done
   if [ "${#GO_RELEASE_PATHS[@]}" -gt 0 ]; then
     command -v goimports >/dev/null || { printf 'goimports is required for changed Go release files\n' >&2; exit 1; }
     gofmt -w "${GO_RELEASE_PATHS[@]}"
     goimports -w "${GO_RELEASE_PATHS[@]}"
     git diff -- "${GO_RELEASE_PATHS[@]}"
   fi
   git grep -n -F "Hermes $TAG" -- cmd/hermes/main.go
   git diff --check
   git diff -- "${RELEASE_PATHS[@]}"
   git status --short
   ```

If the working tree contains paths outside `RELEASE_PATHS` at this point, stop. Do not stage them.

## 4. Run the local release gates

Run these commands from a clean, prepared release tree. Record each exact command and its result.

```sh
go mod download
make lint
make build
test "$(./bin/hermes version | sed -n '1p')" = "Hermes $TAG"
make test-race
git diff --check
```

For changes that plausibly affect performance, also run:

```sh
make benchmark
```

The CI workflow covers pushes to `master` and pull requests that target `master`. Report the CI result for the release commit separately from local gates.

If a required gate fails, stop before staging or publishing. Investigate and fix only failures attributable to the release scope, then rerun the failed gate and affected prior gates. Do not mask failures, skip required checks, or call a partial run successful.

## 5. Publish the Git commit, tag, and GitHub release

Confirm that authorization remains valid immediately before these outward operations. Reprint and inspect `RELEASE_PATHS`; then stage only that exact approved set and inspect the index:

```sh
printf '%s\n' "${RELEASE_PATHS[@]}"
git add -- "${RELEASE_PATHS[@]}"
git diff --cached --check
git diff --cached --name-only
git diff --cached -- "${RELEASE_PATHS[@]}"
```

If any path outside `RELEASE_PATHS` is staged, do not commit; stop and explain the mismatch. Commit with the exact Conventional Commit subject:

```sh
git commit -m "chore(release): prepare $TAG"
```

Record the commit SHA. Verify `master` still points at that commit, push normally, and then create an annotated tag that points to that exact commit. Verify its remote peeled target after pushing:

```sh
RELEASE_COMMIT=$(git rev-parse HEAD)
git push origin master
git tag -a "$TAG" "$RELEASE_COMMIT" -m "Release $TAG"
TAG_OBJECT=$(git rev-parse "$TAG")
TAG_TARGET=$(git rev-parse "$TAG^{}")
git push origin "$TAG"
REMOTE_TAG_OBJECT=$(git ls-remote --tags origin "refs/tags/$TAG" | cut -f1)
REMOTE_TAG_TARGET=$(git ls-remote --tags origin "refs/tags/$TAG^{}" | cut -f1)
test "$REMOTE_TAG_OBJECT" = "$TAG_OBJECT"
test "$TAG_TARGET" = "$RELEASE_COMMIT"
test "$REMOTE_TAG_TARGET" = "$RELEASE_COMMIT"
```

Treat published annotated tags as immutable by release policy: never force-push, update, delete, move, or recreate a tag. If the tag already exists at any point, stop; corrections require the next version. GitHub does not technically enforce that policy by default; configure a GitHub tag-protection/ruleset separately if hard enforcement is required.

Tags do **not** automatically create GitHub releases. Prepare concise technical release notes from the released changelog section and actual changes in a notes file, then create the release explicitly after the tag push:

```sh
gh release create "$TAG" \
  --target "$RELEASE_COMMIT" \
  --title "$TAG" \
  --notes-file "$RELEASE_NOTES"
```

Create a source-only release with no assets by default. Add assets only when the user explicitly requests them. If a GitHub release already exists, stop and report it rather than editing or recreating it.

## 6. Verify Go module publication

After the tag and GitHub release are public, verify module availability. Use bounded retries for proxy and documentation indexing because propagation can lag. Example: five attempts with 15, 30, 60, 120, and 240 second waits; stop early on success. Do not retag or alter the release while waiting.

First query the public proxy:

```sh
GOPROXY=https://proxy.golang.org go list -m github.com/BumpyClock/hermes@"$TAG"
```

Then verify the three proxy artifacts:

```sh
curl --fail --silent --show-error "https://proxy.golang.org/github.com/%21bumpy%21clock/hermes/@v/$TAG.info"
curl --fail --silent --show-error "https://proxy.golang.org/github.com/%21bumpy%21clock/hermes/@v/$TAG.mod"
curl --fail --silent --show-error "https://proxy.golang.org/github.com/%21bumpy%21clock/hermes/@v/$TAG.zip" -o /tmp/hermes-"$TAG".zip
```

Use an isolated temporary module with fresh module and Go build caches to ensure a consumer download is cache-free and does not depend on the repository checkout. Leave all temporary directories for operating-system cleanup and report their paths in the evidence:

```sh
TMP_MODULE=$(mktemp -d)
VERIFY_MODCACHE=$(mktemp -d)
VERIFY_BUILD_CACHE=$(mktemp -d)
(
  cd "$TMP_MODULE"
  go mod init release-verification
  GOMODCACHE="$VERIFY_MODCACHE" \
    GOCACHE="$VERIFY_BUILD_CACHE" \
    GOPROXY=https://proxy.golang.org \
    GONOSUMDB= \
    GOPRIVATE= \
    GOSUMDB=sum.golang.org \
    go mod download -json github.com/BumpyClock/hermes@"$TAG"
)
printf 'temporary verification module retained at %s\n' "$TMP_MODULE"
printf 'temporary verification module cache retained at %s\n' "$VERIFY_MODCACHE"
printf 'temporary verification Go build cache retained at %s\n' "$VERIFY_BUILD_CACHE"
```

Check checksum-database availability through a consumer download or by confirming the downloaded module's `Sum` and `GoModSum` values. If required, query the lookup endpoint with the escaped module path and version:

```sh
curl --fail --silent --show-error "https://sum.golang.org/lookup/github.com/%21bumpy%21clock/hermes@$TAG"
```

Finally check the package documentation page:

```sh
curl --fail --silent --show-error -I "https://pkg.go.dev/github.com/BumpyClock/hermes@$TAG"
```

Proxy and checksum-database success establishes Go publication success. `pkg.go.dev` can lag indexing: report it as pending rather than failed, continue bounded retries, and never mutate the tag to address indexing delay.

## 7. Final evidence report

Finish by reporting concise, factual evidence:

- numeric `VERSION` and release `TAG`;
- release commit SHA;
- annotated tag object SHA, local and remote peeled tag target SHAs, and evidence that both target the release commit;
- GitHub release URL and whether it is source-only or lists requested assets;
- each local gate, command, and result, including goimports, the exact CLI-version assertion, and `go test -race -coverprofile=coverage.out ./...` with its `coverage.out` result;
- tracked binary artifact inventory, embedded-version consistency result, and any separately approved regeneration or removal;
- CI status for the release commit, including the run URL or the reason no result is available;
- Go proxy command/result and verified `.info`, `.mod`, and `.zip` artifacts;
- cache-free consumer download result and all retained temporary verification paths;
- checksum result (`Sum` and `GoModSum`, or sum.golang.org lookup result);
- pkg.go.dev status: published or pending indexing;
- final `git status --short` and confirmation that local `master` equals `origin/master`.

If any step is blocked, state the exact command, observed output, and safe next action. Do not claim release completion until the commit, annotated tag treated as immutable by release policy with verified remote target, GitHub release, and Go proxy/checksum verification all succeed.

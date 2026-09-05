#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 1 || ! $1 =~ ^v1\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  printf 'Usage: bash scripts/verify-release.sh v1.X.Y\nRequires authorization for public Go proxy, checksum, and pkg.go.dev queries.\n' >&2
  exit 2
fi
tag=$1
module=github.com/BumpyClock/hermes
for command in go curl python3; do
  command -v "$command" >/dev/null || { printf '%s is required\n' "$command" >&2; exit 1; }
done

evidence=$(mktemp -d "${TMPDIR:-/tmp}/hermes-release.XXXXXX")
printf 'Verification evidence retained at %s\n' "$evidence"
mkdir "$evidence/module" "$evidence/modcache" "$evidence/buildcache" "$evidence/bin"
export GOWORK=off GOFLAGS= GOENV=off GO111MODULE=on
export GOMODCACHE="$evidence/modcache" GOCACHE="$evidence/buildcache" GOBIN="$evidence/bin"
export GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org GONOSUMDB= GOPRIVATE= GONOPROXY=
unset GOOS GOARCH GOAMD64 GOARM GOARM64 GO386 GOMIPS GOMIPS64 GOPPC64 GORISCV64 GOWASM
cd "$evidence/module"

proxy="https://proxy.golang.org/github.com/%21bumpy%21clock/hermes/@v/$tag"
for extension in info mod zip; do
  curl --fail --silent --show-error --connect-timeout 15 --max-time 120 \
    "$proxy.$extension" -o "$evidence/$tag.$extension"
done
python3 - "$evidence/$tag.info" "$tag" <<'PY'
import json
import sys

with open(sys.argv[1]) as source:
    info = json.load(source)
if info.get("Version") != sys.argv[2]:
    sys.exit("Proxy version does not match the release tag")
PY

go mod init release-verification
go mod download -json "$module@$tag" > "$evidence/download.json"
python3 - "$evidence/download.json" "$module" "$tag" <<'PY'
import json
import sys

with open(sys.argv[1]) as source:
    download = json.load(source)
if (download.get("Error") or download.get("Path") != sys.argv[2]
        or download.get("Version") != sys.argv[3]
        or not download.get("Sum", "").startswith("h1:")
        or not download.get("GoModSum", "").startswith("h1:")):
    sys.exit("Module download lacks the expected version or authenticated checksums")
print("Module checksum: " + download["Sum"])
print("go.mod checksum: " + download["GoModSum"])
PY
go mod edit "-require=$module@$tag"
cat > main.go <<'GO'
package main

import "github.com/BumpyClock/hermes"

func main() {
	var client *hermes.Client = hermes.New()
	if client == nil {
		panic("hermes.New returned nil")
	}
}
GO
go mod tidy
selected=$(go list -m -f '{{.Version}}' "$module")
if [[ $selected != "$tag" ]]; then
  printf 'Consumer resolved a different module version than %s\n' "$tag" >&2
  exit 1
fi
go build -o "$evidence/bin/consumer" .
go install "$module/cmd/hermes@$tag"
"$evidence/bin/hermes" version > "$evidence/cli-version.txt"
IFS= read -r version < "$evidence/cli-version.txt"
if [[ $version != "Hermes $tag" ]]; then
  printf 'CLI version mismatch: expected Hermes %s\n' "$tag" >&2
  exit 1
fi
printf 'Go proxy, checksum, consumer compile, and tagged CLI checks passed for %s\n' "$tag"
if curl --fail --silent --show-error --head --connect-timeout 15 --max-time 30 \
  "https://pkg.go.dev/$module@$tag" > "$evidence/pkg-go-dev.headers" 2> "$evidence/pkg-go-dev.stderr"; then
  printf 'pkg.go.dev: available\n'
else
  printf 'pkg.go.dev: pending or unavailable (see retained diagnostics)\n'
fi

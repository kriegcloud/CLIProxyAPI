#!/usr/bin/env zsh
# Build the DankStation vendor binary of CLIProxyAPI and (optionally) restart the
# user service. See DANKSTATION.md.
set -euo pipefail
cd "${0:A:h}"
export MISE_QUIET=1
commit="$(git rev-parse --short HEAD)"
base="$(git describe --tags --abbrev=0 2>/dev/null || echo dev)"
version="${base#v}-dankstation.${DANKSTATION_BUILD_SERIAL:-1}"
mise exec -- gofmt -l internal/registry/ | { ! grep . ; }
mise exec -- go test ./internal/registry/
mise exec -- go build -trimpath \
  -ldflags "-s -w -X main.Version=${version} -X main.Commit=${commit} -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/cli-proxy-api ./cmd/server
./bin/cli-proxy-api -h 2>/dev/null | head -1   # banner line; no --version flag exists
if [[ "${1:-}" == "--restart" ]]; then
  systemctl --user restart cli-proxy-api
  sleep 2
  systemctl --user is-active cli-proxy-api
fi

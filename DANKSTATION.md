# DankStation vendor build of CLIProxyAPI

Why this exists: the ChatGPT Codex backend serves `gpt-daybreak-blue-latest`
("Daybreak Blue") to this account, but CLIProxyAPI routes only models present in
its catalog (`internal/registry/models/models.json`, refreshed every 3 h from
`router-for-me/models`). Config-side `oauth-model-alias` cannot add an unknown
upstream id (aliases only rename catalog entries), so the model is injected as a
Codex builtin in `internal/registry/model_definitions.go` (`WithCodexBuiltins`),
which survives remote catalog refreshes. It is appended only when the remote
catalog lacks it, so an upstream definition wins once one is published.

Branch: `dankstation/daybreak`, rebased onto tag `v7.2.155` on 2026-09-08 (was
`v7.2.149`, the pacman `cli-proxy-api-bin` version at the time). Go 1.26.8 is
pinned via `mise.toml`. `v7.2.155` carries `gpt-6-astra` in the stock catalog, so
the default `claudex` model needs no vendor patch; only Daybreak Blue does.
Reference upstream checkout: `~/YeeBois/dev/CLIProxyAPI`.

## Build / install

```sh
./dankstation-build.sh --restart   # gofmt check, registry tests, build bin/, restart service
```

The user service keeps the pacman unit but overrides `ExecStart` through
`~/.config/systemd/user/cli-proxy-api.service.d/override.conf` to point at
`bin/cli-proxy-api` in this checkout. Delete that drop-in and
`systemctl --user daemon-reload && systemctl --user restart cli-proxy-api` to
fall back to `/usr/bin/cli-proxy-api`.

## Verify

```sh
TOK="$(cat ~/.cli-proxy-api/client-token)"
curl -s -H "Authorization: Bearer $TOK" http://127.0.0.1:8317/v1/models | grep -o 'gpt-\(6-astra\|daybreak\)[^"]*'
curl -s -X POST http://127.0.0.1:8317/v1/messages -H "x-api-key: $TOK" -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model":"gpt-daybreak-blue-latest(medium)","max_tokens":20,"messages":[{"role":"user","content":"Reply with exactly: OK"}]}'
```

Known quirk (backend, not proxy): responses to `gpt-daybreak-blue-latest`
carry `"model":"gpt-5.6-sol"`. The same echo appears when calling
`https://chatgpt.com/backend-api/codex/responses` directly with the OAuth token,
while a bogus id returns HTTP 400, so the id is accepted and the echo is the
backend's naming of the served deployment.

## Upgrading

```sh
git fetch origin --tags
git rebase v<new-tag>          # or merge; the patch is one hunk in model_definitions.go
DANKSTATION_BUILD_SERIAL=1 ./dankstation-build.sh --restart
```

If the upstream catalog gains `gpt-daybreak-blue-latest`, the builtin becomes a
no-op and the patch can be dropped along with the systemd drop-in.

To refresh the live catalog from the backend (shows what the account can use):

```sh
mise exec -- go run ./cmd/fetch_codex_models --auths-dir ~/.cli-proxy-api --output /tmp/codex-live.json
```

## Codex credential gotcha (2026-09-08)

The admitted Codex OAuth file (`~/.cli-proxy-api/codex-*-pro.json`) logged
`refresh_token_invalidated` while `~/.codex/auth.json` was fresh: the ChatGPT
refresh token rotates, so a credential copied from the Codex CLI is invalidated
the next time the CLI refreshes. Re-admit with a separate login instead of
copying tokens:

```sh
./bin/cli-proxy-api --config ~/.cli-proxy-api/config.yaml --codex-login
```

Symptom when it finally expires: `503 auth_unavailable: no auth available
(providers=codex, …)` on every `gpt-*` model while `codex exec` works.

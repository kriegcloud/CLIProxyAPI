# DankStation vendor build of CLIProxyAPI

Why this exists: the ChatGPT Codex backend serves `gpt-daybreak-blue-latest`
("Daybreak Blue") to this account, but CLIProxyAPI routes only models present in
its catalog (`internal/registry/models/models.json`, refreshed every 3 h from
`router-for-me/models`). Config-side `oauth-model-alias` cannot add an unknown
upstream id (aliases only rename catalog entries), so the model is injected as a
Codex builtin in `internal/registry/model_definitions.go` (`WithCodexBuiltins`),
which survives remote catalog refreshes. It is appended only when the remote
catalog lacks it, so an upstream definition wins once one is published.

Branch: `dankstation/daybreak`, rebased onto tag `v7.3.12` on 2026-09-22 (previously `v7.2.155` on 2026-09-08; before that
`v7.2.149`, the pacman `cli-proxy-api-bin` version at the time). Go 1.26.8 is
pinned via `mise.toml`. `v7.2.155` carries `gpt-6-astra` in the stock catalog, so
the default `claudex` model needs no vendor patch; only Daybreak Blue does.
Reference upstream checkout: `~/YeeBois/dev/CLIProxyAPI`.

## Workstation update, 2026-10-06

The vendor branch now merges upstream `v8.0.16` (source merge `f84dfd0f`),
preserving the Daybreak builtin. The installed binary reports
`8.0.16-dankstation.1`. Registry, Meta authentication, configuration, and SDK
login tests passed before the service restart. The separate reference checkout
at `~/YeeBois/dev/CLIProxyAPI` is fast-forwarded to the same upstream tag.

Meta/Muse OAuth is admitted from the existing Muse Code device session. A
credential exchange confirmed an active Everyday Usage subscription with no
payment required before admission. The proxy exposes Muse Spark 1.3 alongside
the existing providers; all 52 previously exposed model IDs remained present.
OpenAI chat completions and Anthropic messages both returned successful Muse
responses. The `claudem` zsh function uses the existing `_claude_proxy` helper,
selects `muse-spark-1.3(medium)`, and remains outside automatic fallback routing.
Reload the shell configuration or open a new terminal to use it. Native `muse`
continues to use its existing direct Meta login.

The mode-0600 proxy auth file is under the configured auth directory. Never
print it or copy it into this repository. Reauthentication, if needed, uses
`bin/cli-proxy-api --config ~/.cli-proxy-api/config.yaml --meta-login`.

Rollback material is under `~/.local/state/cliproxyapi/upgrades/20261006/`:
the previous binary, original configuration, and shell configuration. The
Git branch `backup/pre-v8-20261006` retains the previous vendor source. To
roll back the service, atomically replace `bin/cli-proxy-api` with the saved
binary and restart the user service. The original configuration is unchanged.
Remove the added Meta auth file and `claudem` function to undo Muse admission.

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

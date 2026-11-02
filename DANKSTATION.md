# DankStation vendor build of CLIProxyAPI

Why this exists: the ChatGPT Codex backend serves `gpt-daybreak-blue-latest`
("Daybreak Blue") to this account, but CLIProxyAPI routes only models present in
its catalog (`internal/registry/models/models.json`, refreshed every 3 h from
`router-for-me/models`). Config-side `oauth-model-alias` cannot add an unknown
upstream id (aliases only rename catalog entries), so the model is injected as a
Codex builtin in `internal/registry/model_definitions.go` (`WithCodexBuiltins`),
which survives remote catalog refreshes. It is appended only when the remote
catalog lacks it, so an upstream definition wins once one is published.

Fork: `kriegcloud/CLIProxyAPI`, branch `dankstation/main` (upstream v8.0.51 plus
`dankstation:` commits as of 2026-11-01; history before the fork is on
`archive/dankstation/daybreak`). Go is pinned via `mise.toml`. The stock catalog
carries `gpt-6-astra`, `gpt-6.1-sol` and `grok-4.7`, so only Daybreak Blue needs
the registry patch; the other commits are runtime fixes.
Reference upstream checkout: `~/YeeBois/dev/CLIProxyAPI`.

## Routing roles, 2026-10-14

The proxy serves 62 ids. Operator ruling 2026-10-14 (beep-effect
`goals/agent-pool-doctrine/history/2026-10-14-routing-refresh.md`) binds
`claudeg` and web research to `grok-4.7` (x_search injection verified on it),
`claudem` to `muse-wonder-1.0(medium)` (opt-in, never a fallback), the Haiku slot
of every proxy wrapper to `gpt-6-luna`, and Graft deep runs to `claude-opus-5-5`,
which the registry now lists. Daybreak stays routable only when a task names it:

<!-- beep-models:begin deprecated-routable -->
| Role | Surface | Model | Effort |
| --- | --- | --- | --- |
| deprecated.routable | proxy-workflow | `gpt-daybreak-blue-latest` | — |

superseded: gpt-5.6-sol, gpt-6-sol, grok-4.6, gpt-5.6-luna, claude-opus-5, claude-opus-5-thinking-high, muse-spark-1.3, grok-4.5, gpt-5.4, gpt-5.4-mini
<!-- beep-models:end deprecated-routable -->

## Latest workstation update, 2026-11-01

The vendor changes now live in the fork **`kriegcloud/CLIProxyAPI`**, branch
`dankstation/main`: five `dankstation:` commits on upstream `7cad1cb7` (v8.0.51).
This checkout is the only one: `origin` = the fork, `upstream` = router-for-me,
local `main` mirrors `upstream/main`. Older vendor branches (including
`dankstation/daybreak`) are preserved on the fork as `archive/*`. Before the fork
the overlay lived uncommitted in per-update worktrees; those snapshots are in
`~/ai/tmp/update-20261101/retired-cliproxy-worktrees/`. Registry, watcher, claude
auth, management, executor and OpenAI handler tests passed. The binary reports `8.0.51-dankstation.1` and is
built in place at `CLIProxyAPI/bin/cli-proxy-api` (the path the systemd drop-in runs).
After restart all 209 model ids were still listed, and `gpt-6.1-sol(medium)`,
`claude-opus-5-5` and `grok-4.7` each answered a 20-token smoke.

Rollback: `~/ai/tmp/update-20261101/cli-proxy-api.before` (8.0.48-dankstation.4).
The startup refresh 401 (`invalid_refresh_token`) came from an obsolete duplicate,
`codex-078cae47-...-pro.json` (same account as the live `-promax.json`, last
refreshed 2026-09-30 before the plan change). It was moved to
`~/.cli-proxy-api-retired/20261101/`; after restart, 0 refresh errors and 11 auth files.

## Earlier workstation update, 2026-10-06 evening

The vendor branch is rebased onto `v8.0.17` at `2af04608`, preserving its four
local commits. The rebuilt user-service binary reports `8.0.17-dankstation.1`.
Registry, executor, SDK and translator tests passed. After service restart,
all 57 prior model IDs remain available, including Daybreak and Muse. This
update validated model discovery; it did not repeat paid inference.

Rollback: branch `backup/update-20261006-late` and binary
`~/ai/tmp/update-20261006/cli-proxy-api.before`. The pacman package was also
updated to 8.0.17; the user service continues to run this vendor binary.

## Earlier workstation update, 2026-10-06

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
git fetch upstream --tags && git fetch origin
git branch -f main upstream/main
git switch dankstation/main
git rebase upstream/main            # or a release tag; resolve conflicts in the dankstation: commits
mkdir -p ~/ai/tmp/update-$(date +%Y%m%d) && cp -p bin/cli-proxy-api ~/ai/tmp/update-$(date +%Y%m%d)/cli-proxy-api.before
DANKSTATION_BUILD_SERIAL=1 ./dankstation-build.sh          # gofmt, registry tests, build bin/ (overwrites the live binary file)
# targeted tests: go test ./internal/registry/... ./internal/watcher/... ./internal/auth/claude/... \
#   ./internal/api/handlers/management/... ./internal/runtime/executor/... ./sdk/api/handlers/openai/...
systemctl --user restart cli-proxy-api   # then run the Verify block
git push --force-with-lease origin dankstation/main
```

Keep each vendor change as its own `dankstation:` commit so rebases stay reviewable.

If the upstream catalog gains `gpt-daybreak-blue-latest`, the builtin becomes a
no-op and the patch can be dropped along with the systemd drop-in.

To refresh the live catalog from the backend (shows what the account can use):

```sh
mise exec -- go run ./cmd/fetch_codex_models --auths-dir ~/.cli-proxy-api --output /tmp/codex-live.json
```

## Re-measuring the Claude header pin

`claude-header-defaults` in `~/.cli-proxy-api/config.yaml` must look like a real
Claude Code client or Fable fails Anthropic's version gate. When Claude Code
updates, capture what it actually sends: run a throwaway local HTTP listener that
logs request headers (drop any auth/key/token headers) and answers 400, point a
scrubbed `env -i ... ANTHROPIC_BASE_URL=http://127.0.0.1:<port> claude -p hi`
at it, then copy `user-agent`, `x-stainless-package-version` and
`x-stainless-runtime-version` into the config. `-p` mode reports
`(external, sdk-cli)`; keep the interactive `(external, cli)` suffix. The proxy
hot-reloads the config. 2026-11-01: 2.1.316 / 0.128.0 / v26.12.1 (was 2.1.280).

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

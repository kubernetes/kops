# AGENTS.md

## Repo Shape
- This is a Go workspace (`go.work`) with three modules: root `github.com/linode/linodego/v2`, `./k8s`, and `./test`.
- Root package files implement the public API client; `k8s/` is a separate helper module for LKE Kubernetes client behavior; `test/` is a separate module for unit and integration tests and replaces both local modules.
- API resource files follow a flat root pattern (`instances.go`, `volumes.go`, etc.) and usually pair public types with `Client` methods that call helpers in `request_helpers.go`.

## Commands
- Full CI-like local check: `make test` runs build, lint, unit tests, and fixture-backed integration tests; it can be slow because `test-int` uses a 5h timeout.
- Faster focused default: run `go test ./...` at the repo root for root-module unit coverage only, then run focused tests in `test/` or `k8s/` as needed.
- Unit tests excluding integration playback: `make test-unit`; pass focused args as `make TEST_ARGS="-run TestName" test-unit`.
- Integration fixture playback: `make test-int`; focused playback: `make TEST_ARGS="-run TestListVolumes" test-int`.
- K8s module verification: `cd k8s && go test ./...` or use root `make build`/`make vet`, which enter `k8s/` explicitly.
- Tidy all modules after dependency changes: `make tidy`.

## Lint And Formatting
- `make lint` uses Docker by default: `docker run ... golangci/golangci-lint:latest`; set `SKIP_DOCKER=1` to use local `golangci-lint`.
- `golangci-lint fmt` is the preferred formatter, if available. Fallback to `gofumpt` if `golangci-lint` is unavailable. Fallback to `go fmt` if `gofumpt` is unavailable.
- `make build` already depends on `vet` and `lint`; use `SKIP_LINT=1 make test` only when intentionally matching CI's test job behavior.
- CI runs `make tidy`, then installs `gofumpt` at commit `86bffd62437a3c437c0b84d5d5ab244824e762fc` and runs `gofumpt -l -w .`, then fails on any diff.

## Tests And Fixtures
- `test/unit` uses embedded JSON fixtures from `test/unit/fixtures/*.json` and `ClientBaseCase` helpers in `test/unit/base.go`; add/update JSON fixtures there for unit tests.
- `test/integration` uses go-vcr YAML fixtures under `test/integration/fixtures`; `make test-int` runs them in replay mode with a fake token and `LINODE_API_VERSION=v4beta`.
- To record integration fixtures, set a real `LINODE_TOKEN` and run `make fixtures`; focus recording with `make TEST_ARGS="-run TestName" fixtures`.
- Fixture recording creates real Linode resources and `TestMain` creates a Cloud Firewall by default; set `ENABLE_CLOUD_FW=false` only if you intentionally want to skip that setup.
- Fixture sanitization is partial: auth headers, dates, some keys, and IPv6 are scrubbed, but `README.md` warns that sensitive account details such as `fixtures/*Account*.yaml` are not fully sanitized. Inspect recorded fixture diffs before committing.
- Smoke tests are live record-mode tests: `make test-smoke` requires `LINODE_TOKEN`.

## Environment
- `.env` is loaded by the root Makefile if present; `.gitignore` excludes it.
- Common env vars: `LINODE_TOKEN`, `LINODE_DEBUG`, `LINODE_URL`, `LINODE_API_VERSION`, `LINODE_CA`, `LINODE_CONFIG`, and `LINODE_PROFILE`.
- `NewClient` reads `LINODE_URL`, `LINODE_API_VERSION`, `LINODE_CA`, and `LINODE_DEBUG`; `NewClientFromEnv` prefers `LINODE_TOKEN` over config-file profiles.

## Conventions And Gotchas
- Optional fields in create or update options structs must use `json:",omitzero"`.
- Optional fields in create or update options structs must be pointer types so explicit zero values can be serialized when needed.
- List APIs mutate the supplied `*ListOptions` with `Page`, `Pages`, and `Results`; do not reuse one `ListOptions` across list calls.
- Use `formatAPIPath` for endpoint paths with user-provided string path segments so path escaping matches the client helpers.
- CI enforces PR titles like `TPT-1234: Description` unless labels exempt the PR (`dependencies`, `hotfix`, `community-contribution`, `ignore-for-release`).
- The `e2e_scripts` directory is a git submodule used by CI report/upload and cleanup jobs; clone/update submodules before reproducing those workflows.

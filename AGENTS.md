<!--
SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
SPDX-License-Identifier: CC0-1.0
-->

# AGENTS.md

Instructions for AI agents working in this repository.

## What this is

Go library and CLI to discover and control Daikin air conditioners with
BRP072C wifi adapters on the local network. Zero external dependencies;
standard library only.

## Layout

| Path | Purpose |
|------|---------|
| `client.go` | HTTPS client for one adapter (register, status, control) |
| `discover.go` | UDP broadcast discovery on port 30050 |
| `config.go` | JSON config store (`~/.config/daikin/config.json`): client UUID and registered devices |
| `modes.go` | Name/code translation for mode, fan rate and swing |
| `parse.go` | Parses the adapter's `key=value,key=value` response format |
| `cmd/daikin/` | CLI: `discover`, `register`, `status`, `set` |

Library code is package `daikin` at the repo root; each file has a matching
`_test.go` next to it.

## Build and test

```
go build ./...        # build; CLI binary via: go build -o daikin ./cmd/daikin
go test ./...         # unit tests (hermetic: httptest servers, temp dirs)
go test -race ./...   # run before committing
go vet ./...
gofmt -l .            # must print nothing
golangci-lint run
```

There is no Makefile; use the Go toolchain directly.

## Protocol constraints (do not "fix" these)

- The adapter matches the `X-Daikin-uuid` header name **case-sensitively**
  and returns HTTP 403 for Go's canonical `X-Daikin-Uuid` form. The header
  is written into the `http.Header` map directly, bypassing `Header.Set`.
  `TestUUIDHeaderCasing` guards this; staticcheck's SA1008 warning about it
  is a false positive here.
- The adapters serve HTTPS with a self-signed certificate, so the client
  sets `InsecureSkipVerify`. This is intentional.
- `set_control_info` requires the complete setting set on every call, so
  writes always read the current state first and merge changes into it.
- The adapter's embedded server drops connections under load; the client
  retries a failed request once at the transport level.
- Temperature and humidity stay `string`, not float: the adapter reports
  non-numeric values such as `M` and `--` in dry and fan modes.

## Conventions

- Every source file starts with a 2-line `ABOUTME:` comment describing it.
- Errors are returned, wrapped with context via `fmt.Errorf("...: %w", err)`.
- Commit messages: imperative mood, one logical change per commit, no AI
  attribution or tool references.
- Branch: `main`, committed to directly (early-stage project).
- Tests use real I/O against local test servers (`httptest.NewTLSServer`),
  not mocks. Keep it that way.

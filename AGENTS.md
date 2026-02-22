# Agents

## Project Overview

`house` is a Go CLI tool for managing home infrastructure. It uses [Cobra](https://github.com/spf13/cobra) for command structure.

## Repository Layout

- `cmd/house/main.go` — Entry point. Wires the root Cobra command and registers subcommands.
- `cmd/adblock/` — The `adblock` command group and its subcommands (e.g. `disable`).
- `internal/pihole/` — Pi-hole v6 API client. Handles authentication, API calls, and logout.
- `docs/` — User-facing documentation.
- `Makefile` — Build, test, lint, and clean targets.

## Building and Testing

```sh
make build   # produces bin/house
make test    # runs all unit tests
make lint    # runs golangci-lint
```

## Code Conventions

- Standard Go project layout: `cmd/` for entry points, `internal/` for private packages.
- HTTP clients use `net/http` with 10-second timeouts.
- Tests use `net/http/httptest` with TLS test servers — no real network calls.
- Errors are wrapped with `fmt.Errorf("context: %w", err)` for clear chains.
- Concurrent operations use goroutines with `sync.WaitGroup` and a results channel.

## Configuration

Pi-hole passwords are provided via environment variables (`HOUSE_PIHOLE1_PASSWORD`, `HOUSE_PIHOLE2_PASSWORD`). See `docs/getting-started.md` for details.

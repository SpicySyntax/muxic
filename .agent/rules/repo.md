---
trigger: always_on
---

---
description: General project rules and context
globs: "**/*.go"
alwaysApply: true
---
# Project Rules

## Environment
- **Language**: Go 1.23+
- **OS**: Windows (use PowerShell syntax for commands)

## Build & Run
- Use `go run . <command> [args]` to run the application during development.
  - Example: `go run . record my_track`
- Use `go build -o muxic.exe` to build the binary.
- Use `go test ./...` to run tests.
- Clean build artifacts with `Remove-Item muxic.exe`.

## Codebase Context
- **Entry Point**: `main.go`
- **Config**: `config.go` (handles `muxic_config.json`)
- **Recorded Tracks**: Stored in `tracks/` directory as `.wav` files.
- **Documentation**: See `USAGE.md` for detailed instructions.

## Coding Standards
- Follow standard Go formatting (`gofmt`).
- Check errors explicitly.
- Use standard library `os`, `encoding/json` etc.

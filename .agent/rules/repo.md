---
trigger: always_on
---

---
description: General project rules and context
globs: "**/*.zig"
alwaysApply: true
---
# Project Rules

## Environment
- **Language**: Zig `0.16.0-dev.1484+d0ba6642b`
- **OS**: Windows (use PowerShell syntax for commands)

## Build & Run
- Use `zig build run -- <command> [args]` to run the CLI application.
  - Example: `zig build run -- record my_track`
- Use `zig build test` to run the test suite.
- Clean build artifacts with `Remove-Item -Recurse -Force .zig-cache, zig-out`.

## Codebase Context
- **Entry Point**: `src/main.zig`
- **Config**: `src/config.zig` (handles `muxic_config.json`)
- **Recorded Tracks**: Stored in `tracks/` directory as `.wav` files.
- **Documentation**: See `USAGE.md` for detailed build/run instructions.

## Coding Standards
- Follow standard Zig naming conventions.
- When fixing build errors, valid specific Zig version features (e.g., standard library changes in 0.16.0 dev).
- Prefer `std.fs` and `std.io` usage appropriate for the version.
# ADR-0003: Issue-Driven Ground-Up Delivery & Modular Package Contracts

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 3 of 10
* **Related Files:** [`internal/model/types.go`](../../internal/model/types.go), [`internal/stages/registry.go`](../../internal/stages/registry.go), [`internal/app/app.go`](../../internal/app/app.go)

---

## Context

To avoid building a monolithic cleanup script where scanning, UI formatting, and OS mutations are tightly coupled, we decided that the engineering work must be divided into **well-defined, sequential GitHub Issues** starting from the ground up:

1. **Foundation (`internal/model`, `scripts/build.ps1`)** — File layout, folder structure, conventions, and core domain contracts.
2. **Core Engine (`internal/fsutil`, `internal/sys`)** — Concurrent scanner, Win32 drive telemetry, admin check, and command runner.
3. **Virtualization Stages (`internal/stages/wsl.go`, `docker.go`, `vhdx_helper.go`)** — Stage 1 & Stage 2.
4. **User-Space Caches & Installers (`internal/stages/devcache.go`, `installers.go`)** — Stage 3 & Stage 4.
5. **Toolchains, System Audit & Registry (`internal/stages/toolchains.go`, `system.go`, `registry.go`)** — Stage 5 & Stage 6.
6. **TUI & CLI Orchestration (`internal/ui`, `internal/cli`, `internal/app`)** — Interactive selector, ANSI tables, and flags.
7. **End-to-End Verification (`cmd/space-reclaim`, `main.go`)** — Build pipeline and live `--scan` / `--dry-run` validation.
8. **Documentation & Shell Integration (`README.md`, `$PROFILE`)** — User documentation and global aliases.

---

## Decision

1. **Strict Package Boundaries under `internal/`:**
   * `internal/model`: Pure data structures, interfaces (`Stage`, `TargetItem`, `ActionFunc`), risk levels, and byte/ANSI formatters. Depends on no other internal package.
   * `internal/fsutil` & `internal/sys`: Low-level filesystem and OS process primitives.
   * `internal/stages`: Independent stage implementations that satisfy `model.Stage`.
   * `internal/ui`: Terminal rendering and interactive user input parsing.
   * `internal/app`: Lifecycle orchestrator (`Scan -> Select -> Execute -> Summary`).
   * `internal/cli`: Flag parsing and exit code handling.
2. **Unified `Stage` & `TargetItem` Contract (`internal/model/types.go`):**
   Every stage implements `Scan(ctx context.Context, opts ScanOptions) (*StageScanResult, error)` returning both read-only `DiagnosticInfo` entries and actionable `TargetItem` closures (`ActionFunc`).

```go
type ActionFunc func(ctx context.Context, dryRun bool) (int64, string, error)

type Stage interface {
	ID() string
	Title() string
	Description() string
	Scan(ctx context.Context, opts ScanOptions) (*StageScanResult, error)
}
```

---

## Consequences

* **Positive:** Adding a new stage (e.g., a future Stage 7 for Scoop/Chocolatey caches or Cargo/Gradle caches) requires only creating one file in `internal/stages/` and registering it in `AllStages()` in [`internal/stages/registry.go`](../../internal/stages/registry.go).
* **Positive:** Every commit in the Git history maps cleanly to a single GitHub Issue and passes `go test ./...`.

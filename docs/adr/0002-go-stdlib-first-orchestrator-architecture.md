# ADR-0002: Native Go 1.26.5 Stdlib-First Orchestrator Architecture (Option A)

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 2 of 10
* **Related Files:** [`go.mod`](../../go.mod), [`cmd/space-reclaim/main.go`](../../cmd/space-reclaim/main.go), [`main.go`](../../main.go)

---

## Context

When designing `space-reclaim`, we evaluated three candidate architectures to orchestrate multi-stage disk scanning, interactive item selection, and system command execution on Windows 11:

1. **Option A: Native Go 1.26.5 CLI/TUI (`space-reclaim.exe`)** — A compiled single-binary orchestrator modeled after [`E:\aserron\tools\video\re-compressor`](file:///E:/aserron/tools/video/re-compressor) (`vhscomp.exe`), invoking focused PowerShell/WSL commands only where Windows requires native cmdlets (`Optimize-VHD`, `Clear-RecycleBin`, `DISM`).
2. **Option B: Python 3.12+ (`uv`) with Rich/Questionary** — Modeled after [`E:\aserron\tools\tts\script-previewer`](file:///E:/aserron/tools/tts/script-previewer).
3. **Option C: Pure PowerShell Script Suite (`.ps1`)** — A master PowerShell menu invoking per-stage `.ps1` scripts.

Key requirements included sub-second concurrent directory scanning across tens of thousands of files (`AppData`, `node_modules`, caches), immunity to the legacy 260-character `MAX_PATH` limit, zero runtime interpreter dependencies, and instant startup from any PowerShell terminal.

---

## Decision

We selected **Option A: Native Go 1.26.5 with a Stdlib-First ("Zero-Bloat") design**:

* **`go.mod` uses only the Go 1.26.5 standard library** (`os`, `path/filepath`, `sync`, `context`, `syscall`, `flag`, `bufio`, `encoding/json`, `unicode/utf16`) with zero third-party module dependencies.
* **Single Stripped Binary:** Compiled via `go build -ldflags="-s -w" -o space-reclaim.exe ./cmd/space-reclaim` (~1.6 MB standalone executable).
* **Hybrid Native + Shell Execution:** Filesystem traversal, Win32 drive telemetry (`GetDiskFreeSpaceExW`), and file deletion run natively in Go goroutines, while Windows-specific subsystem operations (`wsl.exe`, `Optimize-VHD`, `diskpart`, `Dism.exe`, `Invoke-DockerCleanup.ps1`) are delegated through an injected runner interface.

---

## Consequences

* **Positive:** **Speed & Concurrency:** Scanning 6 stages across `C:\` and `E:\` executes concurrently via goroutines without PowerShell pipeline overhead or Python virtual environment activation latency.
* **Positive:** **Zero Supply-Chain / Dependency Drift:** Having zero external dependencies in `go.mod` guarantees instant compilation and long-term maintainability.
* **Positive:** **Self-Contained Distribution:** `space-reclaim.exe` runs anywhere on the machine even when cleaning Python (`uv`, `pip`) or Node (`nvm`, `npm`, `pnpm`) toolchains themselves.
* **Trade-off:** Interactive terminal tables and selection menus are implemented directly with ANSI escape codes and `bufio.Reader` rather than heavy TUI frameworks.

# ADR-0009: Zero-Dependency Interactive ANSI TUI Selector & Non-Interactive CLI Flags

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 9 of 10
* **Related Files:** [`internal/ui/table.go`](../../internal/ui/table.go), [`internal/ui/selector.go`](../../internal/ui/selector.go), [`internal/cli/cli.go`](../../internal/cli/cli.go), [`internal/app/app.go`](../../internal/app/app.go)

---

## Context

Users need two distinct ways to operate `space-reclaim`:
1. **Interactive Exploration & Granular Toggling (Default):** Inspect current drive telemetry (`C:\` and `E:\`), view read-only system diagnostics, see every discovered cleanup item numbered globally (`#01`..`#N`) with its size, risk badge (`[SAFE]`, `[MODERATE]`, `[DESTRUCTIVE]`), and `[ADMIN]` badge, and toggle individual items, ranges, stages, or risk presets before choosing whether to run live (`r`) or simulate in dry-run (`d`).
2. **Non-Interactive CLI Automation:** Run quick read-only audits (`--scan`) or scheduled/scripted cleanups (`--stages devcache --risk safe --yes`) without interactive prompts.

---

## Decision

1. **Global 1-Based Item Indexing (`ui.IndexedItem`):**
   `PrintScanReport` renders each stage's `Diagnostics` and `Items` in order and returns a flat slice of `IndexedItem` references so the user can toggle items across any stage using concise range syntax:
   * `1,3,5-8` — flips the selection state (`[x]` / `[ ]`) of individual items and contiguous ranges.
   * `stage:<id>` (e.g., `stage:wsl`, `stage:docker`, `stage:devcache`) — toggles an entire stage on or off.
   * `s` (`Safe` preset), `m` (`Safe + Moderate` preset), `a` (Select All), `n` (Select None).
   * `d` (Execute selected items in **Dry-Run** mode even if `--dry-run` wasn't passed on the CLI), `r` (Run selected items), `q` (Quit without changes).
2. **Dual-Mode CLI Flags (`internal/cli/cli.go`):**
   * `--scan` / `-s`: Read-only audit report.
   * `--stages <list>`: Filter stages (`wsl,docker,devcache,installers,toolchains,system,all`).
   * `--risk <level>`: Pre-select items up to `safe`, `moderate`, or `destructive`.
   * `--dry-run` / `-n`: Simulate execution without modifying the disk.
   * `--yes` / `-y`: Bypass the interactive selector and execute pre-selected items immediately.
   * `--min-size <MB>`: Threshold for listing standalone installers, SQL dumps, and `node_modules` (default `50` MB).

---

## Consequences

* **Positive:** Works cleanly in any terminal emulator (Windows Terminal, PowerShell console, VS Code integrated terminal, SSH sessions) without raw-mode cursor hijacking or screen buffer corruption, preserving full scrollback history of the audit table.
* **Positive:** Supports instant switching between inspection and dry-run (`d`) right inside the interactive prompt.

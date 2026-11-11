# 02. Interactive TUI Guide

When you run `space-reclaim` (or `sreclaim`) without `--scan` or `--yes`, the application scans all selected stages, prints the color-coded audit report, and opens the **Interactive TUI Selector** implemented in [`internal/ui/selector.go`](../../internal/ui/selector.go) and [`internal/ui/table.go`](../../internal/ui/table.go).

---

## Reading the Audit Report

Each stage section displays two groups of information:

1. **Read-Only Telemetry (`•` bullets):**
   Informational metrics that are never deleted or modified (e.g., WSL2 distro `BasePath` and `ext4.vhdx` size, Docker `docker_data.vhdx` location, active NVM Node version, `C:\pagefile.sys`, `C:\hiberfil.sys`, `Windows.edb`, and `Windows Reserved Storage`).
2. **Selectable Action Items (`#01`..`#N`):**
   Every actionable item shows:
   * **`[x]` / `[ ]`**: Current selection state (`[x]` = selected to run, `[ ]` = skipped).
   * **`#01`**: Global 1-based item index across the entire scan report.
   * **Risk Badge**: `[SAFE]` (Green), `[MODERATE]` (Yellow), or `[DESTRUCTIVE]` (Red).
   * **Estimated Reclaimable Size**: Formatted in `GB`, `MB`, or `KB` (or `Action` for dynamic commands like `DISM` and `docker prune`).
   * **`[ADMIN]`**: Shown if the action requires an elevated Administrator terminal (warned in Red if your current shell is `[STANDARD USER]`).

---

## Interactive Commands Reference

At the `Select action >` prompt, you can enter any of the following commands:

| Command | Action | Examples |
| :--- | :--- | :--- |
| **`1..N` / Ranges** | Toggle individual items or contiguous ranges on/off | `1` · `1,3,7` · `2-6` · `1,4,8-12` |
| **`stage:<id>`** | Toggle all items belonging to a specific stage (if any item in the stage is unselected, selects all; otherwise unselects all) | `stage:wsl` · `stage:docker` · `stage:devcache` · `stage:installers` · `stage:toolchains` · `stage:system` |
| **`s` / `safe`** | Apply the **Safe Only** preset (selects non-empty `[SAFE]` caches, Recycle Bin, and internal WSL cache cleanups) | `s` |
| **`m` / `moderate`** | Apply the **Safe + Moderate** preset (selects all `[SAFE]` and `[MODERATE]` items, including VHDX compaction and `DISM`) | `m` |
| **`a` / `all`** | Select **all** items across all stages (including `[DESTRUCTIVE]`) | `a` |
| **`n` / `none`** | Deselect all items (`[ ]`) | `n` |
| **`l` / `list`** | Re-print the full scan report table with current selection states | `l` |
| **`d` / `dry-run`** | Execute the currently selected items in **Dry-Run Simulation Mode** (prints commands without modifying disk) | `d` |
| **`r` / `run`** | **Execute** the currently selected items and print the Before/After `C:\` and `E:\` summary | `r` |
| **`q` / `quit`** | Exit immediately without making any changes | `q` |

---

## Typical Interactive Workflow

```text
Select action > n                # Clear all selections
Select action > stage:devcache   # Select all developer cache items
Select action > 2                # Also toggle item #02 (e.g., WSL Ubuntu VHDX compaction)
Select action > d                # Preview selected actions in Dry-Run mode first
```

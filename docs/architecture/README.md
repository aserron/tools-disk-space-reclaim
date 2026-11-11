# Architecture & Dataflow Overview

This document explains the internal package architecture, concurrency model, and execution dataflow of **`space-reclaim`**.

---

## 1. High-Level System Architecture

```mermaid
flowchart TD
    subgraph Entry["Entrypoints"]
        MAIN["main.go & cmd/space-reclaim/main.go"]
        CLI["internal/cli/cli.go\nFlag Parser & Help"]
    end

    subgraph Orchestrator["Application Lifecycle"]
        APP["internal/app/app.go\nScan -> Select -> Execute -> Summary"]
        REG["internal/stages/registry.go\nStage Catalog & Filter"]
    end

    subgraph Stages["6 Modular Stages (internal/stages/)"]
        S1["Stage 1: wsl.go\nWSL2 Distros & VHDX"]
        S2["Stage 2: docker.go\nDocker Prune & VHDX"]
        S3["Stage 3: devcache.go\nScoop/npm/pnpm/uv/pip/Temp"]
        S4["Stage 4: installers.go\nOEM Drivers & Downloads"]
        S5["Stage 5: toolchains.go\nNVM/Ruby/Choco/JetBrains"]
        S6["Stage 6: system.go\nWinSxS/Telemetry/Projects"]
        VHDX["vhdx_helper.go\nOptimize-VHD / diskpart"]
    end

    subgraph Core["Core Engine & UI"]
        FS["internal/fsutil/\ndrive.go (Win32) & walk.go (Worker Pool)"]
        SYS["internal/sys/\nrunner.go (OSRunner / DryRunRunner) & admin.go"]
        UI["internal/ui/\ntable.go (ANSI Tables) & selector.go (TUI)"]
        MOD["internal/model/\ntypes.go & format.go"]
    end

    MAIN --> CLI --> APP
    APP --> REG
    REG --> S1 & S2 & S3 & S4 & S5 & S6
    S1 & S2 --> VHDX
    S1 & S2 & S3 & S4 & S5 & S6 --> FS & SYS & MOD
    APP --> UI & FS & SYS
```

---

## 2. Execution Lifecycle Sequence

```mermaid
sequenceDiagram
    actor User
    participant CLI as internal/cli
    participant App as internal/app
    participant Drive as internal/fsutil (Win32)
    participant Stages as internal/stages (1..6)
    participant TUI as internal/ui
    participant Runner as internal/sys.Runner

    User->>CLI: space-reclaim [flags]
    CLI->>App: app.Run(ctx, cfg)
    App->>Drive: QueryDrives("C", "E") (Before Snapshot)
    App->>Stages: Scan(ctx, ScanOptions) across selected stages
    Stages-->>App: []*StageScanResult (Diagnostics + TargetItems)
    App->>TUI: PrintScanReport(scanResults, isAdmin)
    alt --scan flag set
        App-->>User: Exit 0 (Read-Only Report)
    else Interactive Mode (default)
        TUI->>User: Prompt (1..N, stage:id, s, m, d, r, q)
        User-->>TUI: Toggle items & confirm (r or d)
    end
    loop For each Selected TargetItem
        App->>Runner: item.Execute(ctx, effectiveDryRun)
        Runner-->>App: ExecResult (BytesReclaimed, Duration, Err)
    end
    App->>Drive: QueryDrives("C", "E") (After Snapshot)
    App->>TUI: PrintExecutionSummary(Before vs. After)
```

---

## 3. Package Responsibilities

| Package | Files | Responsibility |
| :--- | :--- | :--- |
| `internal/model` | [`types.go`](../../internal/model/types.go), [`format.go`](../../internal/model/format.go) | Defines `Stage`, `TargetItem`, `RiskLevel`, `DriveStats`, `DiagnosticInfo`, `ExecResult`, and ANSI/byte formatters. Zero internal imports. |
| `internal/fsutil` | [`drive.go`](../../internal/fsutil/drive.go), [`walk.go`](../../internal/fsutil/walk.go) | Queries `kernel32.dll` `GetDiskFreeSpaceExW`, performs concurrent directory sizing (`MultiPathSize`) skipping Windows reparse points (`0x0400`), and deletes files/folders safely (`RemovePath`). |
| `internal/sys` | [`runner.go`](../../internal/sys/runner.go), [`admin.go`](../../internal/sys/admin.go) | Abstracts `pwsh`, `wsl.exe`, and binary execution behind `Runner` (`OSRunner` and `DryRunRunner`), decodes `wsl.exe` UTF-16LE output, and checks Administrator elevation. |
| `internal/stages` | [`registry.go`](../../internal/stages/registry.go), [`wsl.go`](../../internal/stages/wsl.go), [`docker.go`](../../internal/stages/docker.go), [`devcache.go`](../../internal/stages/devcache.go), [`installers.go`](../../internal/stages/installers.go), [`toolchains.go`](../../internal/stages/toolchains.go), [`system.go`](../../internal/stages/system.go), [`vhdx_helper.go`](../../internal/stages/vhdx_helper.go) | Implements the 6 modular stages and shared VHDX compaction helper. |
| `internal/ui` | [`table.go`](../../internal/ui/table.go), [`selector.go`](../../internal/ui/selector.go) | Renders ANSI progress bars, stage diagnostic tables, selectable item tables, Before/After drive comparisons, and the interactive command loop. |
| `internal/app` | [`app.go`](../../internal/app/app.go) | Coordinates drive snapshots, stage filtering, interactive/automated selection, admin checks, and execution reporting. |
| `internal/cli` | [`cli.go`](../../internal/cli/cli.go) | Parses CLI flags (`--scan`, `--stages`, `--risk`, `--dry-run`, `--yes`, `--min-size`, `--projects-dir`) and handles `SIGINT`/`SIGTERM` cancellation. |

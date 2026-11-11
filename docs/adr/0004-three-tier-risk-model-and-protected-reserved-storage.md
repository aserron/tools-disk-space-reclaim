# ADR-0004: Three-Tier Risk Classification & Protected Windows Reserved Storage Policy

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 4 of 10
* **Related Files:** [`internal/model/types.go`](../../internal/model/types.go), [`internal/stages/system.go`](../../internal/stages/system.go), [`internal/ui/selector.go`](../../internal/ui/selector.go)

---

## Context

Disk reclamation utilities operate across a wide spectrum of safety:
* Deleting `AppData\Local\npm-cache` or `uv\cache` is completely safe and regenerable.
* Compacting `ext4.vhdx` or `docker_data.vhdx` requires shutting down WSL2 (`wsl --shutdown`) and running elevated Hyper-V `Optimize-VHD`.
* Deleting `.sql` database dumps, removing `node_modules` across `C:\projects`, pruning Docker volumes (`-IncludeVolumes`), or moving WSL distributions across drives can disrupt active work if triggered accidentally.
* Furthermore, during our initial system audit, we explicitly decided **never to disable Windows Reserved Storage** (`DISM /Set-ReservedStorageState /State:Disabled`), keeping OS update stability intact.

---

## Decision

1. **Mandatory `RiskLevel` Classification on Every `TargetItem`:**
   * **`RiskSafe` (`[SAFE]` — Green):** Regenerable package manager caches (`npm`, `pnpm`, `uv`, `hatch`, `yarn`, `pip`), updater staging (`SquirrelTemp`, `Ollama\updates_v2`), temp directories, Recycle Bin, OEM unpack folders (`C:\SWSetup`, `C:\Drivers`), internal WSL caches, and safe Docker prunes. Pre-selected by default when non-empty.
   * **`RiskModerate` (`[MODERATE]` — Yellow):** Operations that temporarily stop services, require Administrator privileges, or remove standalone installers/secondary toolchains (`Optimize-VHD` compaction on WSL2/Docker VHDX, secondary NVM Node.js versions, duplicate Ruby installations, `DISM /StartComponentCleanup`, and `.exe`/`.msi` installers in `Downloads`).
   * **`RiskDestructive` (`[DESTRUCTIVE]` — Red):** Irreversible or high-impact operations (relocating WSL2 distros via `wsl --manage --move`, Docker deep prune with `-IncludeVolumes`, deleting `.sql`/`.bak` dumps, and deleting project `node_modules`). **Never selected by default** unless explicitly toggled by the user or forced via `--risk destructive --yes`.
2. **Read-Only Telemetry (`DiagnosticInfo`) vs. Actionable Items (`TargetItem`):**
   System files (`C:\pagefile.sys`, `C:\hiberfil.sys`, `Windows.edb`), active NVM Node versions, and **Windows Reserved Storage** are emitted exclusively as `DiagnosticInfo` structs so the user has full visibility into disk usage without exposing dangerous OS toggles.

---

## Consequences

* **Positive:** Running `space-reclaim` interactively (or with default `--risk safe`) can never accidentally delete project data, Docker volumes, or active toolchains.
* **Positive:** Windows Reserved Storage is permanently documented and visible in Stage 6 (`Protected (Never Disabled)`) without any code path capable of disabling it.

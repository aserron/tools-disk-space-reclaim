# 05. Safety Model & Risk Levels

`space-reclaim` is designed around a defense-in-depth safety model so you can audit and reclaim gigabytes of space with zero risk of silent data loss.

---

## 1. Three-Tier Risk Classification (`model.RiskLevel`)

Defined in [`internal/model/types.go`](../../internal/model/types.go):

| Risk Level | Badge | Pre-Selected by Default? | What Belongs Here |
| :--- | :--- | :--- | :--- |
| **`RiskSafe`** (`0`) | `[SAFE]` (Green) | **Yes** (when size > 0) | Regenerable package manager caches (`npm`, `pnpm`, `uv`, `hatch`, `yarn`, `pip`), updater staging (`SquirrelTemp`, `Ollama\updates_v2`), `%TEMP%`, Recycle Bin, OEM unpack folders (`C:\SWSetup`), internal WSL caches, JetBrains `acp-agents`, Chrome AI model caches, and safe Docker prunes. |
| **`RiskModerate`** (`1`) | `[MODERATE]` (Yellow) | **No** (unless `-m` / `--risk moderate`) | VHDX compaction (`wsl --shutdown` + `Optimize-VHD` on `ext4.vhdx` / `docker_data.vhdx`), `DISM /StartComponentCleanup`, secondary NVM Node versions, legacy Ruby stores (`C:\ProgramData\rvm`), and standalone `.exe`/`.msi` installers in `Downloads`. |
| **`RiskDestructive`** (`2`) | `[DESTRUCTIVE]` (Red) | **Never** (requires explicit opt-in) | Moving WSL2 distributions across drives (`wsl --manage --move`), Docker deep prune with `-IncludeVolumes`, and deleting `.sql`/`.bak` database dumps or `node_modules` folders in `C:\projects`. |

---

## 2. Protected Windows Reserved Storage Policy

> [!CAUTION]
> **Windows Reserved Storage is Strictly Read-Only**
> Per [ADR-0004](../adr/0004-three-tier-risk-model-and-protected-reserved-storage.md), `space-reclaim` **never** runs `DISM /Online /Set-ReservedStorageState /State:Disabled`. Disabling Reserved Storage can cause cumulative Windows 11 feature updates to fail when free space is tight. Instead, Stage 6 reports Reserved Storage strictly as a read-only diagnostic entry.

---

## 3. Reparse Point & Junction Immunity

Per [ADR-0006](../adr/0006-concurrent-reparse-safe-filesystem-scanner.md), [`fsutil.IsReparsePoint`](../../internal/fsutil/walk.go) inspects `syscall.Win32FileAttributeData.FileAttributes & 0x0400` (`FILE_ATTRIBUTE_REPARSE_POINT`) and `os.ModeSymlink`. Directory walks never follow NTFS junctions or symbolic links (such as Scoop `current` junctions, NVM active Node symlinks, or `pnpm` virtual store links).

---

## 4. Administrator Privilege Guards

Actions that require an elevated Windows Administrator token (such as Hyper-V `Optimize-VHD`, `diskpart`, and `Dism.exe /StartComponentCleanup`) are tagged with `RequiresAdmin: true`.
* In the audit table, they display `[ADMIN]` in Green if running elevated or `[NEEDS ADMIN]` in Red if running as a standard user.
* During execution ([`internal/app/app.go`](../../internal/app/app.go)), if `RequiresAdmin` is true and the process is not elevated (and not in `--dry-run`), the item is safely skipped with an explanatory message rather than failing mid-run.

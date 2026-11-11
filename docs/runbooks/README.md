# Operational Runbooks

Step-by-step operational procedures for common disk reclamation, virtual disk compaction, and package manager migration scenarios on Windows 11.

---

## Runbook Index

| Runbook | When to Use |
| :--- | :--- |
| **[1. Emergency Low Disk Space on `C:\`](#1-emergency-low-disk-space-on-c)** | When `C:\` drops below 20 GB free and you need immediate, zero-risk space recovery. |
| **[2. WSL2 & Docker VHDX Physical Compaction](#2-wsl2--docker-vhdx-physical-compaction)** | When deleting files inside Ubuntu or pruning Docker containers does not shrink `ext4.vhdx` or `docker_data.vhdx` on Windows. |
| **[3. Migrating Chocolatey to Scoop & Reclaiming `C:\ProgramData\chocolatey`](#3-migrating-chocolatey-to-scoop--reclaiming-cprogramdatachocolatey)** | When migrating CLI utilities from `C:\ProgramData\chocolatey` to Scoop (`~\scoop`) and eliminating Smart App Control shim blocks. |
| **[4. Auditing Old Project `node_modules` & Database Dumps](#4-auditing-old-project-node_modules--database-dumps)** | When `C:\projects` accumulates gigabytes of stale dependencies or `.sql` backups. |

---

## 1. Emergency Low Disk Space on `C:\`

1. **Run a fast read-only audit first:**
   ```powershell
   space-reclaim --scan
   ```
2. **Execute all `[SAFE]` developer caches and temp cleanups immediately:**
   ```powershell
   space-reclaim --stages wsl,devcache,installers --risk safe --yes
   ```
   This cleans `~\scoop\cache`, `npm-cache`, `pnpm-cache` (+ `pnpm store prune`), `uv\cache`, `hatch\cache`, `Yarn\Cache`, `pip\cache`, `SquirrelTemp`, `Ollama\updates_v2`, `%TEMP%`, `C:\tmp`, `C:\OneDriveTemp`, OEM driver folders (`C:\SWSetup`, `C:\Drivers`), internal WSL2 caches, and the Recycle Bin.

---

## 2. WSL2 & Docker VHDX Physical Compaction

Hyper-V `.vhdx` files expand automatically as data is written inside Linux or Docker, but **never shrink automatically** when files are deleted unless `fstrim` and `Optimize-VHD` are executed.

1. **Open an elevated (Administrator) PowerShell 7 terminal.**
2. **Launch `space-reclaim` filtered to stages 1 and 2:**
   ```powershell
   space-reclaim --stages wsl,docker
   ```
3. **In the interactive prompt:**
   * Press `m` to select all `Safe + Moderate` items (which selects internal WSL cache cleanup, Docker safe prune, `ext4.vhdx` compaction, and `docker_data.vhdx` compaction).
   * Press `r` to execute.
4. **What happens under the hood ([`vhdx_helper.go`](../../internal/stages/vhdx_helper.go)):**
   * Runs `fstrim -v /` inside WSL2 as `root`.
   * Runs `wsl --shutdown` and waits 2 seconds for file locks to release.
   * Runs `Optimize-VHD -Path <vhdx> -Mode Full` (or falls back to `diskpart` `compact vdisk`).
   * Reports the exact physical GB saved on `C:\` and `E:\`.

---

## 3. Migrating Chocolatey to Scoop & Reclaiming `C:\ProgramData\chocolatey`

See the full package matrix in **[Wiki 06: Migrating Chocolatey Apps to Scoop](../wiki/06-chocolatey-to-scoop-migration.md)**.

1. **Install the CLI tools in Scoop (standard non-admin terminal):**
   ```powershell
   scoop bucket add extras
   scoop install 7zip act eza fd ffmpeg fzf gh jq make oh-my-posh ruff shfmt uv zoxide
   ```
2. **Verify Scoop shims are active:**
   ```powershell
   which jq fd rg fzf bat eza zoxide ffmpeg gh uv ruff
   ```
3. **Remove `C:\ProgramData\chocolatey` (`1.46 GB`) via `space-reclaim`:**
   * Open an **Administrator** PowerShell terminal and run:
     ```powershell
     space-reclaim --stages toolchains
     ```
   * Toggle the **`Legacy Chocolatey Store (C:\ProgramData\chocolatey — after Scoop migration)`** item (`choco-programdata`) and press `r`.

---

## 4. Auditing Old Project `node_modules` & Database Dumps

1. **Scan `C:\projects` with a custom minimum size threshold (e.g., `100 MB`):**
   ```powershell
   space-reclaim --stages system --min-size 100
   ```
2. **Selectively toggle only the `node_modules` or `.sql` dumps you want to remove** by entering their item numbers (e.g., `4,6-8`), preview with `d` (Dry-Run), and execute with `r`.

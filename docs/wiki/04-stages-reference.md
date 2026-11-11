# 04. Stages Reference (`wsl`..`system`)

`space-reclaim` organizes disk auditing and cleanup into **6 modular stages** registered in [`internal/stages/registry.go`](../../internal/stages/registry.go).

---

## Stage 1: `wsl` — WSL2 Distributions
**Source:** [`internal/stages/wsl.go`](../../internal/stages/wsl.go) & [`internal/stages/vhdx_helper.go`](../../internal/stages/vhdx_helper.go)

Enumerates all WSL2 distributions registered under `HKCU:\Software\Microsoft\Windows\CurrentVersion\Lxss` (excluding `docker-desktop*`, which is handled in Stage 2).

* **Diagnostics:** Distro name, `BasePath`, and current `ext4.vhdx` physical size on disk.
* **Actionable Items (per distro):**
  1. **`wsl-clean-<distro>` (`[SAFE]`)**: Probes and cleans internal Linux caches as `root`:
     * `/tmp/.trash`
     * `/var/cache/apt/archives` (`apt-get clean -y`)
     * `/var/cuda-repo-*/*.deb`
     * `journalctl --vacuum-time=3d`
     * `~/.cache/{ms-playwright,uv,JetBrains,google-chrome,Homebrew,node-chromium,node,node-gyp,pnpm,typescript,prisma}`
     * Old builds in `~/.vscode-remote-containers/bin` (retaining the newest build)
     * Runs `fstrim -v /`
  2. **`wsl-compact-<distro>` (`[MODERATE]`, `[ADMIN]`)**: Runs `fstrim -v /`, shuts down WSL (`wsl --shutdown`), and compacts `ext4.vhdx` via Hyper-V `Optimize-VHD -Mode Full` (or `diskpart` fallback).
  3. **`wsl-move-<distro>` (`[DESTRUCTIVE]`)**: Shown only if `BasePath` is still on `C:\`. Moves the distribution to `E:\Docker\<Distro>` via `wsl --manage <Distro> --move`.

---

## Stage 2: `docker` — Docker Engine & VHDX Compaction
**Source:** [`internal/stages/docker.go`](../../internal/stages/docker.go)

Locates `docker_data.vhdx` across `E:\Docker\DockerDesktopWSL\...` or `%LOCALAPPDATA%\Docker\wsl\...` and integrates [`E:\aserron\tools\bcht\Invoke-DockerCleanup.ps1`](file:///E:/aserron/tools/bcht/Invoke-DockerCleanup.ps1).

* **Diagnostics:** Location and size of `docker_data.vhdx`.
* **Actionable Items:**
  1. **`docker-prune-safe` (`[SAFE]`)**: Runs `Invoke-DockerCleanup.ps1 -Prune -Force` (or `docker system prune -f` + `docker builder prune -f`) if the Docker daemon is running.
  2. **`docker-prune-deep` (`[DESTRUCTIVE]`)**: Runs `Invoke-DockerCleanup.ps1 -Full -IncludeVolumes -Force` (removes all unused images, BuildKit caches, and unused volumes).
  3. **`docker-compact-vhdx` (`[MODERATE]`, `[ADMIN]`)**: Shuts down WSL2 and compacts `docker_data.vhdx` via `Optimize-VHD` / `diskpart`.

---

## Stage 3: `devcache` — Developer Caches, Updater Staging & Temp
**Source:** [`internal/stages/devcache.go`](../../internal/stages/devcache.go)

Concurrently scans 11 Windows developer cache and temporary locations (`MultiPathSize` worker pool):

| Item ID | Target Path / Command | Risk |
| :--- | :--- | :--- |
| `cache-scoop` | `%USERPROFILE%\scoop\cache` | `[SAFE]` |
| `cache-npm` | `%LOCALAPPDATA%\npm-cache` | `[SAFE]` |
| `cache-pnpm` | `%LOCALAPPDATA%\pnpm-cache` + `pnpm store prune` | `[SAFE]` |
| `cache-uv` | `%LOCALAPPDATA%\uv\cache` | `[SAFE]` |
| `cache-hatch` | `%LOCALAPPDATA%\hatch\cache` | `[SAFE]` |
| `cache-yarn` | `%LOCALAPPDATA%\Yarn\Cache` | `[SAFE]` |
| `cache-pip` | `%LOCALAPPDATA%\pip\cache` | `[SAFE]` |
| `updater-squirrel` | `%LOCALAPPDATA%\SquirrelTemp` | `[SAFE]` |
| `updater-ollama` | `%LOCALAPPDATA%\Ollama\updates_v2` | `[SAFE]` |
| `temp-user` | `%LOCALAPPDATA%\Temp` | `[SAFE]` |
| `temp-ctmp` | `C:\tmp` | `[SAFE]` |
| `temp-onedrive` | `C:\OneDriveTemp` | `[SAFE]` |
| `temp-recycle-bin` | `Clear-RecycleBin -DriveLetter C,E -Force` | `[SAFE]` |

---

## Stage 4: `installers` — OEM Driver Unpack Folders & Downloads Installers
**Source:** [`internal/stages/installers.go`](../../internal/stages/installers.go)

* **OEM Driver Extraction Folders (`[SAFE]`):**
  * `oem-swsetup`: `C:\SWSetup` (HP SoftPaq unpacked drivers)
  * `oem-cdrivers`: `C:\Drivers`
  * `oem-hp-downloads`: `%USERPROFILE%\Downloads\HP Downloads`
  * `oem-amd`: `C:\AMD`
  * `oem-nvidia`: `C:\NVIDIA`
* **Standalone Installers in `Downloads` (`[MODERATE]`):**
  * Scans `%USERPROFILE%\Downloads` (up to depth 2) for `.exe`, `.msi`, `.iso`, and `.msix` files `>= --min-size` (default `50 MB`), sorted largest-first with modification dates.

---

## Stage 5: `toolchains` — IDEs, Language Toolchains & Browser Bloat
**Source:** [`internal/stages/toolchains.go`](../../internal/stages/toolchains.go)

* **Diagnostics:**
  * Active NVM Node.js version (`node -v`) in `C:\nvm4w\nvm`.
  * Installed Ruby versions in `C:\Ruby-on-Windows`.
  * JetBrains IDE installations and caches (`>= 50 MB`) across `C:\Program Files Dev\JetBrains`, `%LOCALAPPDATA%\JetBrains`, and `%APPDATA%\JetBrains`.
* **Actionable Items:**
  * `nvm-node-<ver>` (`[MODERATE]`): Secondary/inactive Node.js versions in `C:\nvm4w\nvm`.
  * `ruby-rvm-progdata` (`[MODERATE]`, `[ADMIN]`): Legacy RVM Ruby store in `C:\ProgramData\rvm`.
  * `ruby-c-ruby27` (`[MODERATE]`): Standalone `C:\Ruby27-x64`.
  * `choco-programdata` (`[MODERATE]`, `[ADMIN]`): Legacy Chocolatey store in `C:\ProgramData\chocolatey` (`1.46 GB`, selectable after migrating CLI tools to Scoop).
  * `jetbrains-acp-agents` (`[SAFE]`): `%LOCALAPPDATA%\JetBrains\acp-agents`.
  * `chrome-optguide-model` (`[SAFE]`): Chrome On-Device AI model (`User Data\OptGuideOnDeviceModel`).
  * `chrome-snapshots` (`[SAFE]`): Chrome upgrade snapshots (`User Data\Snapshots`).

---

## Stage 6: `system` — Windows System Audit & Project Bloat
**Source:** [`internal/stages/system.go`](../../internal/stages/system.go)

* **Read-Only OS Diagnostics:**
  * `C:\pagefile.sys` (Virtual memory paging file)
  * `C:\hiberfil.sys` (Hibernation file)
  * `Windows.edb` (Windows Search Indexer database)
  * **Windows Reserved Storage (`Protected (Never Disabled)`)**
* **Actionable Items:**
  * `sys-dism-winsxs` (`[MODERATE]`, `[ADMIN]`): Runs `Dism.exe /Online /Cleanup-Image /StartComponentCleanup`.
  * `proj-dump-<N>` (`[DESTRUCTIVE]`): Large `.sql`, `.dump`, and `.bak` files in `C:\projects` (`>= --min-size`).
  * `proj-nm-<N>` (`[DESTRUCTIVE]`): Top-level `node_modules` directories in `C:\projects` (`>= --min-size`), sorted largest-first.

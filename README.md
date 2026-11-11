# space-reclaim

**`space-reclaim`** is a native Go 1.26.5 CLI/TUI utility for Windows 11 that audits disk usage across `C:\` and `E:\` and orchestrates modular space reclamation across **WSL2 distributions**, **Docker Desktop VHDX**, **developer package caches**, **OEM driver unpack folders**, **duplicate IDEs/toolchains**, and **Windows OS internals**.

---

## Features & Stages

| Stage ID | Stage Title | Key Targets & Actions |
| :--- | :--- | :--- |
| `wsl` | **Stage 1: WSL2 Distributions** | Cleans internal Linux caches (`~/.cache` for Playwright, uv, JetBrains, Chrome, Homebrew, pnpm, node-gyp, Prisma), `/tmp/.trash`, local CUDA `.deb` archives, `apt-get clean`, `journalctl --vacuum-time=3d`, old `.vscode-remote-containers` server builds, runs `fstrim -v /` + `Optimize-VHD` compaction on `ext4.vhdx`, and supports relocating distros off `C:\` to `E:\Docker\<Distro>`. |
| `docker` | **Stage 2: Docker Engine & VHDX** | Integrates `E:\aserron\tools\bcht\Invoke-DockerCleanup.ps1` (Safe Prune or Deep Prune with volumes) and compacts `docker_data.vhdx` via Hyper-V `Optimize-VHD` / `diskpart`. |
| `devcache` | **Stage 3: Dev Caches & Temp** | Scans and cleans `npm-cache`, `pnpm-cache` (+ `pnpm store prune`), `uv`, `hatch`, `Yarn`, `pip`, `SquirrelTemp`, `Ollama\updates_v2`, `%TEMP%`, `C:\tmp`, `C:\OneDriveTemp`, and empties the Windows Recycle Bin (`C:\`, `E:\`). |
| `installers` | **Stage 4: OEM Drivers & Installers** | Removes OEM driver extraction directories (`C:\SWSetup`, `C:\Drivers`, `Downloads\HP Downloads`, `C:\AMD`, `C:\NVIDIA`) and lists large standalone installers (`.exe`, `.msi`, `.iso`, `.msix`) in `Downloads`. |
| `toolchains` | **Stage 5: IDEs & Toolchains** | Detects secondary NVM Node.js versions (`C:\nvm4w\nvm`), duplicate Ruby installations outside `C:\Ruby-on-Windows`, JetBrains `acp-agents` & IDE caches, and Chrome On-Device AI models (`OptGuideOnDeviceModel`, `Snapshots`). |
| `system` | **Stage 6: Windows System & Projects** | Read-only telemetry for `pagefile.sys`, `hiberfil.sys`, Windows Search Index (`Windows.edb`), and **Windows Reserved Storage (protected Read-Only)**. Offers `DISM /StartComponentCleanup` (`WinSxS`) and scans `C:\projects` for large `.sql`/`.bak` dumps and `node_modules` trees. |

---

## Safety Model

Every discovered item is tagged with an explicit `RiskLevel`:
- **`[SAFE]`**: Regenerable package caches, temporary folders, `apt-get clean`, `fstrim`, and safe Docker prunes.
- **`[MODERATE]`**: VHDX virtual disk compaction (`wsl --shutdown` + `Optimize-VHD`), secondary NVM Node removal, `DISM /StartComponentCleanup`, and standalone installer deletion in `Downloads`.
- **`[DESTRUCTIVE]`** *(Never selected by default)*: Moving WSL2 distributions across drives, pruning Docker volumes (`-IncludeVolumes`), or deleting project `.sql` dumps and `node_modules` directories.
- **Protected OS Policy**: Windows Reserved Storage (`DISM /Set-ReservedStorageState`) is **never** disabled or modified.

---

## Project Structure

```text
E:\aserron\tools\disk\space-reclaim\
├── .gitignore
├── go.mod                          # module spacereclaim (go 1.26.5)
├── main.go                         # Root entrypoint wrapper
├── README.md
├── scripts\
│   └── build.ps1                   # Runs unit tests and builds space-reclaim.exe
├── cmd\
│   └── space-reclaim\
│       └── main.go                 # Binary entrypoint
└── internal\
    ├── app\
    │   └── app.go                  # Scan -> Select -> Execute -> Summary orchestrator
    ├── cli\
    │   └── cli.go                  # CLI flags & help documentation
    ├── model\
    │   ├── types.go                # Core domain contracts (Stage, TargetItem, RiskLevel, DriveStats)
    │   └── format.go               # Human-readable byte formatting & ANSI progress bars
    ├── fsutil\
    │   ├── drive.go                # Win32 GetDiskFreeSpaceExW drive telemetry
    │   └── walk.go                 # Concurrent long-path directory scanner (skips junctions/symlinks)
    ├── sys\
    │   ├── admin.go                # Administrator elevation detection
    │   └── runner.go               # Command runner (pwsh, wsl, UTF-16LE decoding, DryRunRunner)
    ├── ui\
    │   ├── table.go                # ANSI tables & Before/After drive space comparison
    │   └── selector.go             # Interactive terminal selector (toggle items, ranges, stages, presets)
    └── stages\
        ├── registry.go             # Central registry of all 6 stages
        ├── vhdx_helper.go          # Shared WSL shutdown + Optimize-VHD / diskpart compaction
        ├── wsl.go                  # Stage 1: WSL2
        ├── docker.go               # Stage 2: Docker
        ├── devcache.go             # Stage 3: Dev Caches & Temp
        ├── installers.go           # Stage 4: OEM Drivers & Installers
        ├── toolchains.go           # Stage 5: IDEs & Toolchains
        └── system.go               # Stage 6: System Audit & Project Bloat
```

---

## Building & Testing

```powershell
Set-Location "E:\aserron\tools\disk\space-reclaim"
.\scripts\build.ps1
```

---

## Usage Examples

```powershell
# Launch full scan + interactive TUI selector (default)
space-reclaim

# Read-only audit report of all stages and drives (makes zero changes)
space-reclaim --scan

# Inspect only WSL2 and Docker stages interactively
space-reclaim --stages wsl,docker

# Dry-run simulation of all Safe + Moderate items across all stages
space-reclaim --stages all --risk moderate --dry-run --yes

# Run only Safe developer cache and temp cleanup non-interactively
space-reclaim --stages devcache --risk safe --yes
```

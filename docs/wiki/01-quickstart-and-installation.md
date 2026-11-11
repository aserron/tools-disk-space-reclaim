# 01. Quickstart & Installation

**`space-reclaim`** is a single-binary Go 1.26.5 CLI/TUI utility located at [`E:\aserron\tools\disk\space-reclaim`](file:///E:/aserron/tools/disk/space-reclaim).

---

## Prerequisites

* **OS:** Windows 11 (x64)
* **Shell:** PowerShell 7 (`pwsh`)
* **Compiler (for building from source):** Go `1.26.5+`
* **Optional Subsystems:**
  * WSL2 (`wsl.exe`) for Stage 1 (`wsl`)
  * Docker Desktop & [`E:\aserron\tools\bcht\Invoke-DockerCleanup.ps1`](file:///E:/aserron/tools/bcht/Invoke-DockerCleanup.ps1) for Stage 2 (`docker`)
  * Elevated Administrator terminal for Hyper-V `Optimize-VHD` / `diskpart` VHDX compaction and `DISM /StartComponentCleanup`

---

## Building from Source

Run [`scripts\build.ps1`](../../scripts/build.ps1) from PowerShell. This script configures a workspace-scoped `$env:GOTMPDIR`, executes the full unit test suite (`go test ./... -v`), and compiles a stripped binary (`-ldflags="-s -w"`):

```powershell
Set-Location "E:\aserron\tools\disk\space-reclaim"
.\scripts\build.ps1
```

Output binary:
* [`E:\aserron\tools\disk\space-reclaim\space-reclaim.exe`](../../space-reclaim.exe) (~1.6 MB)

---

## Global PowerShell Alias (`$PROFILE`)

The following function and aliases are registered in [`C:\Users\andre\Documents\PowerShell\Microsoft.PowerShell_profile.ps1`](file:///C:/Users/andre/Documents/PowerShell/Microsoft.PowerShell_profile.ps1#L351-L356) (placed above the `# DO NOT MODIFY -- coreutils` block):

```powershell
# Disk Space Reclaim & Audit CLI (E:\aserron\tools\disk\space-reclaim)
function Invoke-SpaceReclaim {
    & "E:\aserron\tools\disk\space-reclaim\space-reclaim.exe" @args
}
Set-Alias -Name space-reclaim -Value Invoke-SpaceReclaim
Set-Alias -Name sreclaim -Value Invoke-SpaceReclaim
```

---

## Your First Run

### 1. Read-Only Audit (`--scan`)
Inspect current `C:\` and `E:\` usage and all 6 stages without making any changes:

```powershell
space-reclaim --scan
```

### 2. Interactive TUI Mode (Default)
Launch the interactive selector to toggle stages or individual items before running:

```powershell
space-reclaim
# or shorthand:
sreclaim
```

### 3. Dry-Run Simulation (`--dry-run`)
Preview the exact commands and byte savings without deleting or modifying any file:

```powershell
space-reclaim --stages all --risk moderate --dry-run --yes
```

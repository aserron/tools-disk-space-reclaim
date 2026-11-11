# 06. Migrating Chocolatey Apps to Scoop (`C:\ProgramData\chocolatey` -> `~\scoop`)

This guide documents the full audit of your current **Chocolatey (`C:\ProgramData\chocolatey`)** and **Scoop (`C:\Users\andre\scoop`)** installations, why migrating CLI tools from Chocolatey to Scoop is recommended on your Windows 11 setup, and the exact commands to perform the migration and reclaim **~1.48 GB** on `C:\`.

---

## 1. Why Migrate from Chocolatey to Scoop?

1. **Windows 11 Smart App Control Blocks Chocolatey Shims:**
   Chocolatey generates wrapper executables ("shims") in `C:\ProgramData\chocolatey\bin\*.exe` (for example, `C:\ProgramData\chocolatey\bin\jq.exe`). On Windows 11 with Application Control / Smart App Control enabled, unsigned shim binaries generated in `C:\ProgramData\chocolatey\bin` frequently fail with:
   ```text
   An Application Control policy has blocked this file: C:\ProgramData\chocolatey\bin\jq.exe
   ```
   **Scoop** installs portable binaries inside your user profile (`C:\Users\andre\scoop\apps\<app>\current`) and uses clean shims in `C:\Users\andre\scoop\shims`, avoiding `C:\ProgramData` shim blocks and requiring **zero Administrator prompts** for daily upgrades (`scoop update *`).
2. **Reclaim `1.48 GB` on `C:\ProgramData\chocolatey`:**
   Your `C:\ProgramData\chocolatey\lib` directory currently consumes **1,463.57 MB (~1.46 GB)** on `C:\`, including several tools that are **already duplicated** in your profile (`uv`, `pnpm`, `gh`, `python`, `msys2`, `mambaforge`/`miniforge3`).

---

## 2. Current Audit of `C:\ProgramData\chocolatey` (`1.46 GB`)

### Already Installed in Scoop (`~\scoop\apps`)
The following packages are already cleanly managed by Scoop:
* `bat`, `flow`, `glow`, `lsd`, `ripgrep`, `scoop`

### Installed in Chocolatey (`C:\ProgramData\chocolatey\lib`) — Categorized Action Plan

#### Category A: Core CLI Tools to Reinstall in Scoop (`scoop install ...`)
These are lightweight daily CLI tools currently living in `C:\ProgramData\chocolatey\lib` (`~495 MB` total) that map 1-to-1 to the Scoop `main` and `extras` buckets:

| Chocolatey Package | Size in `choco\lib` | Scoop Equivalent (`main` / `extras`) | Notes |
| :--- | :--- | :--- | :--- |
| `ffmpeg` | `327.10 MB` | `scoop install ffmpeg` | Used by `re-compressor` (`vhscomp`) and `shadowing-tts` (`ffplay`) |
| `act-cli` | `27.58 MB` | `scoop install act` | Local GitHub Actions runner |
| `rsync` | `22.34 MB` | `scoop install cwrsync` *(or `rsync` via MSYS2/Git)* | File synchronization |
| `7zip.portable` / `7zip.commandline` | `16.26 MB` | `scoop install 7zip` | Archive utility |
| `fd` | `9.55 MB` | `scoop install fd` | Fast file finder |
| `fzf` | `6.02 MB` | `scoop install fzf` | Fuzzy finder |
| `shfmt` | `3.18 MB` | `scoop install shfmt` | Shell formatter |
| `eza` | `2.65 MB` | `scoop install eza` | Modern `ls` replacement |
| `zoxide` | `1.73 MB` | `scoop install zoxide` | Smarter `cd` command |
| `make` | `1.48 MB` | `scoop install make` | GNU Make |
| `jq` | `0.98 MB` | `scoop install jq` | Currently blocked in `choco\bin\jq.exe` by App Control! |
| `tree` | `0.11 MB` | *(Built into Windows `cmd /c tree` or `eza --tree`)* | Optional |
| `oh-my-posh` | `0.01 MB` | `scoop install oh-my-posh` | Prompt theme engine |

#### Category B: Already Duplicated Elsewhere on Your System (Safe to Remove from Choco Without Reinstalling)
These packages in `C:\ProgramData\chocolatey\lib` (`~808 MB` total) are redundant because you already have primary installations managed via `uv`, `nvm`, standalone installers, or Scoop:

| Chocolatey Package | Size in `choco\lib` | Why It Is Redundant |
| :--- | :--- | :--- |
| `mise` | `157.50 MB` | Only keep if you actively use `mise`; otherwise you already manage Node via `nvm`, Python via `uv`, and Ruby via `rbenv` (`scoop install mise` if wanted). |
| `mambaforge` + `miniforge3` | `269.15 MB` | Redundant Conda/Mamba installers cached in `choco\lib` (`135.75 MB` + `133.40 MB`). |
| `msys2` | `102.03 MB` | Only keep the active MSYS2 root (`C:\tools\msys64` or `C:\msys64`); the `102 MB` archive inside `choco\lib\msys2` is just the Chocolatey installer payload! |
| `ruff` | `91.13 MB` | Better installed via `uv tool install ruff` or `scoop install ruff`. |
| `uv` | `58.20 MB` | You already have `uv` (`0.13.7`) installed; or manage via `scoop install uv`. |
| `python314` + `python313` + `python` + `python3` | `108.30 MB` | Cached MSI payloads inside `choco\lib`; Python is managed via `uv` / system Python. |
| `pnpm` | `42.51 MB` | Per your global convention, `pnpm` is managed globally inside your active NVM Node version (`corepack` / `npm i -g pnpm`). |
| `gh` | `34.86 MB` | Installed in `C:\Program Files\GitHub CLI\gh.exe` (or `scoop install gh`). |
| `InkScape` | `158.95 MB` | Full GUI app (`C:\Program Files\Inkscape`); the `158.95 MB` inside `choco\lib\InkScape` is just a leftover `.msi` installer copy! |

#### Category C: Windows System Runtime Metadata Packages (`< 1 MB` total)
* `vcredist140`, `vcredist2015`, `dotnetfx`, `KB*`, `visualstudio2022buildtools`, `visualstudio2022-workload-vctools`, `visualstudio-installer`, `llvm`, `golang`, `mingw`.
* These `.nupkg` folders in `choco\lib` are tiny metadata wrappers (`~10 KB` each); the actual Visual C++ Redistributables and VS2022 Build Tools are installed in Windows (`C:\Program Files (x86)\Microsoft Visual Studio`) and **remain untouched** if you remove `C:\ProgramData\chocolatey`.

---

## 3. Step-by-Step Migration Commands

### Step 1: Install the CLI Tools in Scoop (No Admin Required)
Run in any PowerShell terminal:

```powershell
# Ensure extras bucket is available if needed
scoop bucket add extras

# Reinstall core CLI utilities in Scoop
scoop install 7zip act eza fd ffmpeg fzf gh jq make oh-my-posh ruff shfmt uv zoxide
```

### Step 2: Verify Scoop Shims Take Precedence
```powershell
which jq fd rg fzf bat eza zoxide ffmpeg gh uv ruff
jq --version
```
All commands will now resolve from `C:\Users\andre\scoop\shims\*.exe` without any Smart App Control blocks.

### Step 3: Reclaim `1.46 GB` from `C:\ProgramData\chocolatey`

#### Option 1 — Keep Chocolatey Only for Windows System Packages, Purge Leftover Installer Payloads & Migrated CLI Tools (Recommended Safe Path)
From an **Administrator** PowerShell terminal, uninstall the migrated CLI packages from Chocolatey and delete the cached `.msi`/`.exe`/`.7z` installer payloads inside `C:\ProgramData\chocolatey\lib`:

```powershell
# 1. Remove migrated CLI tools from Chocolatey bin/lib
$migrated = @(
    "jq", "fd", "fzf", "eza", "zoxide", "shfmt", "make", "act-cli",
    "ffmpeg", "7zip.portable", "7zip.commandline", "ruff", "uv", "pnpm",
    "mambaforge", "miniforge3"
)
foreach ($pkg in $migrated) {
    Remove-Item -Path "C:\ProgramData\chocolatey\lib\$pkg" -Recurse -Force -ErrorAction SilentlyContinue
}

# 2. Remove orphaned shims from C:\ProgramData\chocolatey\bin
$shims = @(
    "jq.exe", "fd.exe", "fzf.exe", "eza.exe", "zoxide.exe", "shfmt.exe",
    "make.exe", "act.exe", "ffmpeg.exe", "ffplay.exe", "ffprobe.exe",
    "7z.exe", "7za.exe", "7zFM.exe", "7zG.exe", "ruff.exe", "uv.exe",
    "uvw.exe", "uvx.exe", "pnpm.exe"
)
foreach ($shim in $shims) {
    Remove-Item -Path "C:\ProgramData\chocolatey\bin\$shim" -Force -ErrorAction SilentlyContinue
}

# 3. Delete leftover cached installer binaries (.msi, .exe, .7z, .zip) inside remaining choco\lib folders (InkScape, msys2, python313, python314, gh)
Get-ChildItem "C:\ProgramData\chocolatey\lib" -Include *.msi,*.exe,*.7z,*.zip,*.tar.* -Recurse -File | Remove-Item -Force
```

#### Option 2 — Full Chocolatey Retirement (Removes `C:\ProgramData\chocolatey` Entirely — Reclaims Full `1.48 GB`)
If you no longer wish to use `choco` at all (since GUI apps like Inkscape, GitHub CLI, Python, and VS Build Tools are already installed in `C:\Program Files` via Windows Installer and do not need `C:\ProgramData\chocolatey` to run):

```powershell
# Run in an Administrator PowerShell terminal:
Remove-Item -Path "C:\ProgramData\chocolatey" -Recurse -Force

# Optional: clean Chocolatey environment variables from User/Machine scope
[Environment]::SetEnvironmentVariable("ChocolateyInstall", $null, "User")
[Environment]::SetEnvironmentVariable("ChocolateyInstall", $null, "Machine")
[Environment]::SetEnvironmentVariable("ChocolateyLastPathUpdate", $null, "User")
```

### Step 4: Clean Scoop Download Cache Anytime
Whenever you want to free downloaded archive files in `C:\Users\andre\scoop\cache`:
```powershell
scoop cache rm *
scoop cleanup *
```

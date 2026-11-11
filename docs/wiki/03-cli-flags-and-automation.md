# 03. CLI Flags & Automation

`space-reclaim` supports both interactive terminal sessions and non-interactive CLI execution via [`internal/cli/cli.go`](../../internal/cli/cli.go).

---

## Complete CLI Flags Table

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--scan` | `-s` | `false` | Scan and display the full disk audit report without modifying anything, then exit. |
| `--interactive` | `-i` | `true` | Launch the interactive TUI selector after scanning (ignored if `--scan` or `--yes` is set). |
| `--list-stages` | — | `false` | Print the catalog of all 6 available stages and their IDs, then exit. |
| `--stages <list>` | — | `"all"` | Comma-separated list of stage IDs to scan: `wsl`, `docker`, `devcache`, `installers`, `toolchains`, `system`, or `all`. |
| `--risk <level>` | — | `"safe"` | Maximum risk level to pre-select when `--risk` or `--yes` is passed: `safe`, `moderate`, or `destructive`. |
| `--dry-run` | `-n` | `false` | Simulate execution and print the exact commands/paths without deleting or modifying files. |
| `--yes` | `-y` | `false` | Execute pre-selected items immediately without opening the interactive selector prompt. |
| `--min-size <MB>` | — | `50` | Minimum size in MB for listing standalone installers in `Downloads`, `.sql`/`.bak` dumps, or `node_modules` directories. |
| `--projects-dir <path>` | — | `C:\projects` | Root directory scanned in Stage 6 (`system`) for database dumps and `node_modules` folders. |
| `--help` | `-h` | — | Display usage instructions and flag defaults. |

---

## Common CLI Recipes

### 1. Fast Read-Only Audit of Specific Stages
```powershell
# Audit only WSL2 and Docker virtual disks
space-reclaim --scan --stages wsl,docker

# Audit only developer caches and OEM installers
space-reclaim -s --stages devcache,installers
```

### 2. Automated Weekly Safe Cache Cleanup
Clean all `[SAFE]` developer caches (`npm`, `pnpm`, `uv`, `hatch`, `yarn`, `pip`, `SquirrelTemp`, `Temp`, Recycle Bin) and internal WSL caches without prompts:
```powershell
space-reclaim --stages wsl,devcache --risk safe --yes
```

### 3. Scan for Large Installers & Project Bloat (`>= 250 MB`)
Raise the size threshold to `250 MB` when auditing `Downloads` and `C:\projects`:
```powershell
space-reclaim --scan --stages installers,system --min-size 250
```

### 4. Full VHDX Compaction Run (Run as Administrator)
From an elevated PowerShell 7 prompt, run WSL2 and Docker VHDX compaction:
```powershell
space-reclaim --stages wsl,docker --risk moderate
```

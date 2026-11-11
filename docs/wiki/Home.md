# `space-reclaim` Wiki & Help Center

Welcome to the **`space-reclaim`** documentation wiki. Use this index to navigate user guides, CLI references, stage catalogs, and Windows toolchain migration guides.

---

## Wiki Pages

| Page | Description |
| :--- | :--- |
| **[01. Quickstart & Installation](01-quickstart-and-installation.md)** | Building `space-reclaim.exe`, configuring `$PROFILE` aliases (`space-reclaim`, `sreclaim`), and running your first scan. |
| **[02. Interactive TUI Guide](02-interactive-tui-guide.md)** | Mastering the interactive selector: toggling item numbers/ranges (`1,3,5-8`), toggling stages (`stage:wsl`), presets (`s`, `m`), and dry-run (`d`). |
| **[03. CLI Flags & Automation](03-cli-flags-and-automation.md)** | Complete CLI flag reference (`--scan`, `--stages`, `--risk`, `--dry-run`, `--yes`, `--min-size`, `--projects-dir`) and automated task examples. |
| **[04. Stages Reference (`wsl`..`system`)](04-stages-reference.md)** | Detailed breakdown of all 6 stages, what each target item cleans, and read-only diagnostics. |
| **[05. Safety Model & Risk Levels](05-safety-model-and-risk-levels.md)** | How `[SAFE]`, `[MODERATE]`, and `[DESTRUCTIVE]` work, Administrator requirements, and the Protected Windows Reserved Storage policy. |
| **[06. Migrating Chocolatey Apps to Scoop](06-chocolatey-to-scoop-migration.md)** | Audit of `C:\ProgramData\chocolatey` (`1.46 GB`), why Smart App Control blocks `C:\ProgramData\chocolatey\bin\*.exe` shims, and how to migrate CLI apps cleanly to Scoop (`~\scoop`). |

---

## Related Documentation Folders

* **[Architecture Decision Records (`docs/adr/`)](../adr/README.md)** — All 10 architectural decisions in chronological order.
* **[Architecture & Dataflow (`docs/architecture/`)](../architecture/README.md)** — System architecture diagrams, package dependency graphs, and lifecycle flows.
* **[Operational Runbooks (`docs/runbooks/`)](../runbooks/README.md)** — Step-by-step runbooks for WSL2/Docker VHDX compaction, low disk recovery, and Chocolatey-to-Scoop migration.
* **[Development & Extending (`docs/development/`)](../development/README.md)** — How to add a new stage, write unit tests with `DryRunRunner`, and build with `scripts\build.ps1`.

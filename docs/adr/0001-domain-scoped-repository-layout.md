# ADR-0001: Domain-Scoped Repository Location & Naming (`E:\aserron\tools\disk\space-reclaim`)

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 1 of 10
* **Related Artifacts:** [`README.md`](../../README.md), [`Microsoft.PowerShell_profile.ps1`](file:///C:/Users/andre/Documents/PowerShell/Microsoft.PowerShell_profile.ps1)

---

## Context

After conducting an in-depth disk space investigation on `C:\` and `E:\` that reclaimed **+117.57 GB** across WSL2 virtual disks, Docker BuildKit/VHDX storage, developer package caches, OEM driver extraction folders, and duplicate toolchains, we needed a permanent, reusable utility to codify all audit and cleanup steps.

Personal developer utilities in this environment are organized on `E:\aserron\tools` grouped by functional domain:
* `E:\aserron\tools\video\re-compressor` (remote: `aserron/tools-video-re-compressor`)
* `E:\aserron\tools\video\linkedin_voice_coach`
* `E:\aserron\tools\tts\script-previewer` (remote: `aserron/tools-tts-scrprv`)
* `E:\aserron\tools\bcht`

We needed to establish the canonical local path, GitHub repository name, and global PowerShell aliases before scaffolding the project.

---

## Decision

1. **Local Workspace Directory:** Place the project under the `disk` domain at **`E:\aserron\tools\disk\space-reclaim`**.
2. **Remote GitHub Repository:** Use the hyphenated domain-prefixed repository name **`git@github.com:aserron/tools-disk-space-reclaim.git`**.
3. **Binary & Shell Aliases:** Compile the standalone executable as **`space-reclaim.exe`** at the repository root and expose global PowerShell commands **`space-reclaim`** and **`sreclaim`** (via `Invoke-SpaceReclaim` in `$PROFILE`).

---

## Consequences

* **Positive:** Follows the exact `<root>\tools\<domain>\<tool-name>` taxonomy used across `E:\aserron\tools`, making discovery, scripting, and GitHub repository mapping predictable.
* **Positive:** Placing the repository on `E:\` keeps build artifacts and temporary test binaries off `C:\`, aligning with our disk reclamation goals.
* **Neutral:** Global invocation depends on the `$PROFILE` wrapper function `Invoke-SpaceReclaim` pointing to `E:\aserron\tools\disk\space-reclaim\space-reclaim.exe`.

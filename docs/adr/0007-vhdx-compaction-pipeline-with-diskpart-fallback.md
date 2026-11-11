# ADR-0007: Two-Phase VHDX Compaction Pipeline (`fstrim` + `Optimize-VHD` with `diskpart` Fallback)

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 7 of 10
* **Related Files:** [`internal/stages/vhdx_helper.go`](../../internal/stages/vhdx_helper.go), [`internal/stages/wsl.go`](../../internal/stages/wsl.go), [`internal/stages/docker.go`](../../internal/stages/docker.go)

---

## Context

WSL2 distributions (`ext4.vhdx`) and Docker Desktop (`docker_data.vhdx`) use dynamically expanding Hyper-V virtual disks. When tens of gigabytes of files or container layers are deleted inside Linux, the `.vhdx` file on the Windows host (`C:\` or `E:\`) **does not shrink automatically**.

Furthermore, compacting a `.vhdx` on Windows 11 has two failure modes:
1. **Locked VHDX Handle:** If the WSL2 utility VM is still running, `Optimize-VHD` or `diskpart` fails with a file-in-use lock error.
2. **Missing Hyper-V PowerShell Module:** On machines where the Hyper-V Management Tools PowerShell module (`Optimize-VHD`) is not installed, compaction fails unless a fallback to `diskpart` (`compact vdisk`) is provided.

---

## Decision

We centralized virtual disk compaction in [`CompactVHDX`](../../internal/stages/vhdx_helper.go) using a strict two-phase protocol:

1. **Phase 1 — Guest Free-Block Zeroing (`fstrim -v /`):**
   For WSL2 distributions, `space-reclaim` first runs `fstrim -v /` as `root` inside the distribution so the Linux `ext4` filesystem marks deleted blocks as unused (`TRIM` discard) before shutting down the VM.
2. **Phase 2 — Host VM Shutdown & Compaction (`Optimize-VHD` -> `diskpart` Fallback):**
   * Runs `wsl.exe --shutdown` and waits 2 seconds to guarantee all VHDX file locks are released.
   * Checks `if (Get-Command Optimize-VHD -ErrorAction SilentlyContinue)` and runs `Optimize-VHD -Path $vhdx -Mode Full`.
   * If `Optimize-VHD` is unavailable, automatically generates an ASCII script (`select vdisk` -> `attach vdisk readonly` -> `compact vdisk` -> `detach vdisk`) and executes `diskpart /s` with guaranteed `try / finally` temp file cleanup.
   * Compares `os.Stat(vhdxPath).Size()` before and after compaction to report the exact physical bytes reclaimed on the host drive.

---

## Consequences

* **Positive:** Works reliably across both Windows 11 Pro/Enterprise with Hyper-V cmdlets (`Optimize-VHD`) and systems relying on built-in `diskpart`.
* **Positive:** Shared between Stage 1 (`wsl.go`) and Stage 2 (`docker.go`), eliminating duplicate VHDX handling code.
* **Requirement:** Marked with `RequiresAdmin: true` and `RiskModerate` because `Optimize-VHD`/`diskpart` require an elevated Administrator token and temporarily stop WSL2.

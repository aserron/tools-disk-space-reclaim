# ADR-0005: Injected `sys.Runner` Abstraction, Dry-Run Guarantee & WSL UTF-16LE Decoding

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 5 of 10
* **Related Files:** [`internal/sys/runner.go`](../../internal/sys/runner.go), [`internal/sys/runner_test.go`](../../internal/sys/runner_test.go), [`internal/sys/admin.go`](../../internal/sys/admin.go)

---

## Context

`space-reclaim` interacts with multiple Windows and Linux subsystems:
* `pwsh -NoProfile -NonInteractive -Command ...` for registry queries, `Optimize-VHD`, `Clear-RecycleBin`, and `Invoke-DockerCleanup.ps1`.
* `wsl.exe -d <distro> -u root -- bash -c ...` for Linux-side cache probes, cleanup, and `fstrim -v /`.
* `wsl.exe --list`, `wsl.exe --shutdown`, and `wsl.exe --manage` on the Windows host.
* `Dism.exe` and `docker` CLI commands.

Two technical hazards exist when invoking these commands from Go on Windows:
1. **Accidental Execution in Tests or Dry-Run:** If stages call `os/exec.Command` directly, unit tests or dry-run bugs could execute real system mutations.
2. **`wsl.exe` UTF-16LE Output Encoding:** Host management commands in `wsl.exe` (such as `wsl --list --verbose` or `wsl --manage`) emit UTF-16 Little Endian text with NUL (`\x00`) bytes interleaved between ASCII characters, which breaks standard UTF-8 string parsing and terminal output.

---

## Decision

1. **Inject `sys.Runner` into All Stages:**
   No stage imports `os/exec`. All external process execution goes through [`sys.Runner`](../../internal/sys/runner.go):
   * `OSRunner`: Production implementation executing real commands with `context.Context` cancellation support.
   * `DryRunRunner`: Test and simulation implementation that records every command into `Recorded []string` and returns `"[DRY-RUN] ..."` without spawning OS processes.
2. **Automatic UTF-16LE Detection & Decoding (`DecodeUTF16IfNeeded`):**
   `RunWSLHost` inspects the raw byte slice returned by `wsl.exe`. When interleaved NUL bytes characteristic of UTF-16LE are detected, it decodes the `uint16` Little-Endian stream (stripping any `0xFEFF` BOM) into clean UTF-8.
3. **Non-Interactive Root WSL Execution:**
   `RunWSL(ctx, distro, "root", bashCmd)` passes `-u root` explicitly to `wsl.exe` so cleanup commands (`apt-get clean`, `journalctl --vacuum-time=3d`, `fstrim -v /`) never hang waiting for an interactive `sudo` password prompt.

---

## Consequences

* **Positive:** Unit tests run deterministically in milliseconds using `DryRunRunner` with zero risk to the host machine.
* **Positive:** `wsl.exe` output renders cleanly in logs and terminal tables without garbled NUL characters.
* **Positive:** WSL2 probes and cleanups never hang on interactive password prompts.

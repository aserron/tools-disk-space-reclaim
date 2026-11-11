# ADR-0008: Workspace-Scoped `GOTMPDIR` on `E:\` to Bypass Windows Smart App Control Blocks

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 8 of 10
* **Related Files:** [`scripts/build.ps1`](../../scripts/build.ps1)

---

## Context

While implementing Stage 5 and Stage 6 during Issue #5 (`internal/stages/system.go` and `internal/stages/registry_test.go`), running `go test ./... -v` failed with:

```text
fork/exec C:\Users\andre\AppData\Local\Temp\go-build1425099145\b136\stages.test.exe: An Application Control policy has blocked this file.
```

By default, `go test` compiles temporary `.test.exe` binaries into `%LOCALAPPDATA%\Temp` on `C:\`. Windows 11 Application Control / Smart App Control policies block unsigned temporary executables spawned from `%TEMP%` when those binaries contain system administration command literals (`Dism.exe`, `diskpart`, `Optimize-VHD`), whereas executables compiled and run on the dedicated developer volume (`E:\aserron\...`) are trusted.

---

## Decision

In [`scripts/build.ps1`](../../scripts/build.ps1), we configure Go's temporary build directory (`$env:GOTMPDIR`) to a workspace-local `.gotmp` directory inside `E:\aserron\tools\disk\space-reclaim\.gotmp` wrapped in a `try / finally` block:

```powershell
$GoTmp = Join-Path $ProjectRoot ".gotmp"
New-Item -ItemType Directory -Force -Path $GoTmp | Out-Null
$env:GOTMPDIR = $GoTmp

try {
    go test ./... -v
    go build -ldflags="-s -w" -o $Output ./cmd/space-reclaim
} finally {
    Remove-Item -Path $GoTmp -Recurse -Force -ErrorAction SilentlyContinue
}
```

---

## Consequences

* **Positive:** `.\scripts\build.ps1` runs `go test ./... -v` and `go build` deterministically with zero Windows Application Control false-positive blocks.
* **Positive:** Keeps all temporary Go compilation artifacts off `C:\Users\andre\AppData\Local\Temp`, preventing `C:\` temp bloat during development and cleaning `.gotmp` automatically on exit.

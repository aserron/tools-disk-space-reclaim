# ADR-0006: Concurrent Long-Path Scanner Skipping Reparse Points & Win32 Drive Telemetry

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 6 of 10
* **Related Files:** [`internal/fsutil/walk.go`](../../internal/fsutil/walk.go), [`internal/fsutil/drive.go`](../../internal/fsutil/drive.go), [`internal/fsutil/walk_test.go`](../../internal/fsutil/walk_test.go)

---

## Context

Scanning developer directories on Windows 11 (`AppData\Local\pnpm-cache`, `uv\cache`, `JetBrains`, `C:\nvm4w\nvm`, and `C:\projects\*\node_modules`) presents three classic Windows filesystem pitfalls:

1. **Reparse Points, Junctions & Symlinks:** Windows `AppData`, Scoop (`~/scoop/apps/<app>/current`), NVM (`C:\Program Files\nodejs`), and `pnpm` virtual stores make heavy use of NTFS directory junctions (`FILE_ATTRIBUTE_REPARSE_POINT = 0x0400`) and symlinks. Following them during size calculation causes double-counting, infinite directory loops, or accidental deletion of linked target trees.
2. **Legacy 260-Character `MAX_PATH` Limit:** Deeply nested `node_modules` and cache paths frequently exceed 260 characters, causing legacy COM/PowerShell scanners (`Scripting.FileSystemObject`) to silently misreport sizes.
3. **Sequential Scan Bottlenecks:** Scanning 20+ cache and toolchain directories sequentially takes several seconds on cold disk caches.

---

## Decision

1. **Explicit Win32 Reparse Point Detection (`IsReparsePoint` in `internal/fsutil/walk.go`):**
   Before descending into any subdirectory during `filepath.WalkDir`, `IsReparsePoint` checks both Go's `os.ModeSymlink | os.ModeIrregular` bits and the underlying Win32 `syscall.Win32FileAttributeData.FileAttributes & 0x0400` (`FILE_ATTRIBUTE_REPARSE_POINT`). Matching directories immediately return `filepath.SkipDir`.
2. **Bounded Concurrent Worker Pool (`MultiPathSize`):**
   `MultiPathSize` evaluates multiple directory trees concurrently using a semaphore of 8 goroutines (`sem := make(chan struct{}, 8)`), saturating NVMe I/O without overwhelming the OS scheduler.
3. **Direct Win32 `GetDiskFreeSpaceExW` Telemetry (`internal/fsutil/drive.go`):**
   Queries `kernel32.dll` `GetDiskFreeSpaceExW` directly via `syscall.NewLazyDLL("kernel32.dll")` for `C:\` and `E:\` before and after execution, providing exact byte-level Before/After drive deltas.

---

## Consequences

* **Positive:** Full 6-stage system scans complete in ~1–2 seconds and never double-count junctions or follow symlinks outside the target directory.
* **Positive:** Locked files in `%TEMP%` are gracefully skipped during both `PathSize` and best-effort `RemovePath` (`keepRootDir = true`), reporting the exact net bytes freed (`before - after`).

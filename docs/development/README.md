# Development & Extending Guide

This guide covers how to build, test, and extend **`space-reclaim`** with new cleanup targets or stages.

---

## 1. Building & Running Unit Tests

Always use [`scripts\build.ps1`](../../scripts/build.ps1) to run tests and compile the binary. Per [ADR-0008](../adr/0008-gotmpdir-workspace-isolation-for-app-control.md), `build.ps1` sets `$env:GOTMPDIR` to `E:\aserron\tools\disk\space-reclaim\.gotmp` so Windows 11 Smart App Control never blocks temporary `.test.exe` binaries in `%TEMP%`.

```powershell
Set-Location "E:\aserron\tools\disk\space-reclaim"
.\scripts\build.ps1
```

---

## 2. How to Add a New Cache Target to Stage 3 (`devcache`)

To add a new developer cache directory (for example, `Cargo`, `Gradle`, `Maven`, or `Scoop`), append a `cacheTargetDef` entry to the `defs` slice in [`internal/stages/devcache.go`](../../internal/stages/devcache.go):

```go
{
	id:          "cache-scoop",
	name:        "Scoop Download Cache (~\\scoop\\cache)",
	description: "Clears downloaded Scoop installer archives in ~\\scoop\\cache",
	path:        filepath.Join(home, "scoop", "cache"),
	keepRoot:    true,
},
```

`DevCacheStage.Scan` will automatically include its path in the concurrent `fsutil.MultiPathSize` worker pool and expose it as a selectable `[SAFE]` item when non-empty.

---

## 3. How to Create a New Stage (`model.Stage`)

1. **Create `internal/stages/<name>.go`** implementing the [`model.Stage`](../../internal/model/types.go) interface:
   ```go
   type MyStage struct {
       runner sys.Runner
   }

   func NewMyStage(runner sys.Runner) *MyStage {
       return &MyStage{runner: runner}
   }

   func (s *MyStage) ID() string          { return "mystage" }
   func (s *MyStage) Title() string       { return "Stage 7: My Custom Stage" }
   func (s *MyStage) Description() string { return "Description of what this stage scans and cleans" }

   func (s *MyStage) Scan(ctx context.Context, opts model.ScanOptions) (*model.StageScanResult, error) {
       // Populate Diagnostics and Items
   }
   ```
2. **Register the stage in [`internal/stages/registry.go`](../../internal/stages/registry.go)** inside `AllStages(runner sys.Runner)`.
3. **Update [`internal/stages/registry_test.go`](../../internal/stages/registry_test.go)** to assert the updated stage count and test stage selection via `sys.NewDryRunRunner()`.

---

## 4. Coding Conventions

* **Never call `os/exec.Command` directly in `internal/stages/`** — always call `s.runner.Run`, `s.runner.RunPS`, `s.runner.RunWSL`, or `s.runner.RunWSLHost` ([ADR-0005](../adr/0005-injected-command-runner-and-wsl-utf16le-decoding.md)).
* **Always honor `dryRun bool` inside `TargetItem.Execute`** — return the estimated byte savings and a `"[DRY-RUN] Would ..."` message without mutating disk state.
* **Default Selection Safety** — `[DESTRUCTIVE]` items must always initialize with `Selected: false` ([ADR-0004](../adr/0004-three-tier-risk-model-and-protected-reserved-storage.md)).
* **Line Endings** — Use Windows CRLF (`\r\n`) across repository files.

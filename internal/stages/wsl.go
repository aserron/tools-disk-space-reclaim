package stages

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"spacereclaim/internal/model"
	"spacereclaim/internal/sys"
)

// WSLDistro represents a registered WSL2 distribution discovered from the Windows Registry.
type WSLDistro struct {
	Name     string `json:"Name"`
	BasePath string `json:"BasePath"`
	VHDXPath string `json:"VHDXPath"`
}

// WSLStage audits and cleans WSL2 distributions (internal Linux caches, fstrim + VHDX compaction, and drive relocation).
type WSLStage struct {
	runner sys.Runner
}

func NewWSLStage(runner sys.Runner) *WSLStage {
	return &WSLStage{runner: runner}
}

func (s *WSLStage) ID() string    { return "wsl" }
func (s *WSLStage) Title() string { return "Stage 1: WSL2 Distributions (Internal Caches, VHDX Compaction & Relocation)" }
func (s *WSLStage) Description() string {
	return "Cleans Linux caches/trash/apt/journals inside WSL2 distros, runs fstrim + Optimize-VHD, and optionally moves distros off C:\\"
}

func (s *WSLStage) discoverDistros(ctx context.Context) ([]WSLDistro, error) {
	script := `
$lxss = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Lxss'
if (-not (Test-Path $lxss)) { Write-Output '[]'; exit 0 }
$list = @()
Get-ChildItem $lxss | ForEach-Object {
    $p = Get-ItemProperty $_.PSPath
    $base = $p.BasePath
    if ($base -and $base.StartsWith('\\?\')) { $base = $base.Substring(4) }
    $vhdx = if ($base) { Join-Path $base 'ext4.vhdx' } else { '' }
    $list += [PSCustomObject]@{
        Name     = [string]$p.DistributionName
        BasePath = [string]$base
        VHDXPath = [string]$vhdx
    }
}
$list | ConvertTo-Json -Compress
`
	out, err := s.runner.RunPS(ctx, script)
	if err != nil || out == "" || strings.HasPrefix(out, "[DRY-RUN]") {
		return nil, err
	}

	var distros []WSLDistro
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		var single WSLDistro
		if err := json.Unmarshal([]byte(out), &single); err == nil && single.Name != "" {
			distros = append(distros, single)
		}
		return distros, nil
	}
	_ = json.Unmarshal([]byte(out), &distros)
	return distros, nil
}

// probeInternalCacheBytes estimates reclaimable cache/trash/apt/deb bytes inside a running or stopped WSL distro.
func (s *WSLStage) probeInternalCacheBytes(ctx context.Context, distroName string) int64 {
	probeCmd := `
total=0
for p in /tmp/.trash /var/cache/apt/archives /var/cuda-repo-*; do
  if [ -e "$p" ]; then
    sz=$(du -sb "$p" 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
    total=$((total + sz))
  fi
done
for hm in /home/*; do
  for c in ms-playwright uv JetBrains google-chrome Homebrew node-chromium node node-gyp pnpm typescript prisma; do
    if [ -e "$hm/.cache/$c" ]; then
      sz=$(du -sb "$hm/.cache/$c" 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
      total=$((total + sz))
    fi
  done
done
echo "$total"
`
	probeCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	out, err := s.runner.RunWSL(probeCtx, distroName, "root", probeCmd)
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	last := strings.TrimSpace(lines[len(lines)-1])
	val, _ := strconv.ParseInt(last, 10, 64)
	return val
}

func (s *WSLStage) Scan(ctx context.Context, _ model.ScanOptions) (*model.StageScanResult, error) {
	start := time.Now()
	res := &model.StageScanResult{
		StageID:    s.ID(),
		StageTitle: s.Title(),
		Category:   "Virtualization & Linux Subsystems",
	}

	distros, err := s.discoverDistros(ctx)
	if err != nil {
		res.ScanWarning = fmt.Sprintf("Failed to query WSL registry: %v", err)
		res.ScanDuration = time.Since(start)
		return res, nil
	}

	for _, d := range distros {
		// Skip internal docker-desktop distros here (handled in Stage 2: Docker)
		lower := strings.ToLower(d.Name)
		if strings.HasPrefix(lower, "docker-desktop") {
			continue
		}

		var vhdxSize int64
		if d.VHDXPath != "" {
			if fi, err := os.Stat(d.VHDXPath); err == nil {
				vhdxSize = fi.Size()
			}
		}

		res.Diagnostics = append(res.Diagnostics, model.DiagnosticInfo{
			Name:   fmt.Sprintf("WSL2 Distro: %s", d.Name),
			Status: d.BasePath,
			Size:   vhdxSize,
			Notes:  fmt.Sprintf("VHDX: %s", d.VHDXPath),
		})

		distroName := d.Name
		vhdxPath := d.VHDXPath
		basePath := d.BasePath

		// 1. Internal WSL2 Cache, Trash, Apt, Journal & Old VS Code Remote Containers Cleanup
		internalSize := s.probeInternalCacheBytes(ctx, distroName)
		res.Items = append(res.Items, &model.TargetItem{
			ID:          fmt.Sprintf("wsl-clean-%s", lower),
			StageID:     s.ID(),
			Name:        fmt.Sprintf("WSL2 [%s]: Clean internal caches, /tmp/.trash, apt, journals & old VS Code servers", distroName),
			Description: "Cleans ~/.cache (playwright, uv, JetBrains, chrome, Homebrew, pnpm, node-gyp, prisma), /tmp/.trash, local CUDA .debs, apt-get clean, journalctl (3d), and old .vscode-remote-containers",
			Path:        fmt.Sprintf(`\\wsl.localhost\%s`, distroName),
			SizeBytes:   internalSize,
			Risk:        model.RiskSafe,
			Selected:    true,
			Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
				cleanBash := `
set -e
rm -rf /tmp/.trash/* /tmp/.trash/.* 2>/dev/null || true
rm -f /var/cuda-repo-*/*.deb 2>/dev/null || true
apt-get clean -y 2>/dev/null || true
journalctl --vacuum-time=3d 2>/dev/null || true
for hm in /home/*; do
  if [ -d "$hm/.cache" ]; then
    for c in ms-playwright uv JetBrains google-chrome Homebrew node-chromium node node-gyp pnpm typescript prisma; do
      rm -rf "$hm/.cache/$c" 2>/dev/null || true
    done
  fi
  if [ -d "$hm/.vscode-remote-containers/bin" ]; then
    cd "$hm/.vscode-remote-containers/bin" && ls -1dt * 2>/dev/null | tail -n +2 | xargs -r rm -rf
  fi
done
fstrim -v / 2>/dev/null || true
`
				if dryRun {
					return internalSize, fmt.Sprintf("[DRY-RUN] Would clean internal caches, /tmp/.trash, apt, journals and run fstrim inside WSL distro %s", distroName), nil
				}
				out, err := s.runner.RunWSL(execCtx, distroName, "root", cleanBash)
				if err != nil {
					return 0, out, err
				}
				return internalSize, fmt.Sprintf("Cleaned internal caches & ran fstrim in %s (%s)", distroName, strings.TrimSpace(out)), nil
			},
		})

		// 2. VHDX Compaction (fstrim + wsl --shutdown + Optimize-VHD)
		if vhdxPath != "" && vhdxSize > 0 {
			res.Items = append(res.Items, &model.TargetItem{
				ID:            fmt.Sprintf("wsl-compact-%s", lower),
				StageID:       s.ID(),
				Name:          fmt.Sprintf("WSL2 [%s]: Trim & Compact Virtual Disk (ext4.vhdx)", distroName),
				Description:   "Runs fstrim -v /, shuts down WSL2, and compacts ext4.vhdx via Hyper-V Optimize-VHD / diskpart",
				Path:          vhdxPath,
				SizeBytes:     vhdxSize,
				Risk:          model.RiskModerate,
				RequiresAdmin: true,
				Selected:      false,
				Notes:         "Current VHDX size on disk (requires Admin for Optimize-VHD)",
				Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
					if !dryRun {
						_, _ = s.runner.RunWSL(execCtx, distroName, "root", "fstrim -v / 2>/dev/null || true")
					}
					return CompactVHDX(execCtx, s.runner, vhdxPath, dryRun)
				},
			})
		}

		// 3. Relocate Distro from C:\ to E:\Docker\<Distro> if currently on C:\
		if strings.HasPrefix(strings.ToUpper(basePath), `C:\`) {
			targetDir := filepath.Join(`E:\Docker`, distroName)
			res.Items = append(res.Items, &model.TargetItem{
				ID:          fmt.Sprintf("wsl-move-%s", lower),
				StageID:     s.ID(),
				Name:        fmt.Sprintf("WSL2 [%s]: Move distribution off C:\\ to %s", distroName, targetDir),
				Description: fmt.Sprintf("Relocates %s from %s to %s using 'wsl --manage %s --move'", distroName, basePath, targetDir, distroName),
				Path:        basePath,
				SizeBytes:   vhdxSize,
				Risk:        model.RiskDestructive,
				Selected:    false,
				Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
					if dryRun {
						return vhdxSize, fmt.Sprintf("[DRY-RUN] Would run 'wsl --shutdown' and 'wsl --manage %s --move %s'", distroName, targetDir), nil
					}
					if err := os.MkdirAll(targetDir, 0o755); err != nil {
						return 0, "", err
					}
					_, _ = s.runner.RunWSLHost(execCtx, "--shutdown")
					out, err := s.runner.RunWSLHost(execCtx, "--manage", distroName, "--move", targetDir)
					if err != nil {
						return 0, out, err
					}
					return vhdxSize, fmt.Sprintf("Moved WSL distro %s to %s freeing %s on C:\\", distroName, targetDir, model.FormatBytes(vhdxSize)), nil
				},
			})
		}
	}

	for _, it := range res.Items {
		if it.Risk == model.RiskSafe {
			res.TotalSize += it.SizeBytes
		}
	}
	res.ScanDuration = time.Since(start)
	return res, nil
}

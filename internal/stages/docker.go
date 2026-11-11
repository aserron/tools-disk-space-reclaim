package stages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"spacereclaim/internal/model"
	"spacereclaim/internal/sys"
)

const defaultDockerCleanupScript = `E:\aserron\tools\bcht\Invoke-DockerCleanup.ps1`

// DockerStage integrates Invoke-DockerCleanup.ps1 and compacts Docker Desktop's docker_data.vhdx.
type DockerStage struct {
	runner sys.Runner
}

func NewDockerStage(runner sys.Runner) *DockerStage {
	return &DockerStage{runner: runner}
}

func (s *DockerStage) ID() string    { return "docker" }
func (s *DockerStage) Title() string { return "Stage 2: Docker Engine Prune & docker_data.vhdx Compaction" }
func (s *DockerStage) Description() string {
	return "Integrates Invoke-DockerCleanup.ps1 (safe prune or deep volume prune) and compacts docker_data.vhdx"
}

func findDockerVHDX() string {
	candidates := []string{
		`E:\Docker\DockerDesktopWSL\DockerDesktopWSL\disk\docker_data.vhdx`,
		`E:\Docker\DockerDesktopWSL\disk\docker_data.vhdx`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Docker\wsl\disk\docker_data.vhdx`),
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Docker\wsl\data\ext4.vhdx`),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func (s *DockerStage) Scan(_ context.Context, _ model.ScanOptions) (*model.StageScanResult, error) {
	start := time.Now()
	res := &model.StageScanResult{
		StageID:    s.ID(),
		StageTitle: s.Title(),
		Category:   "Containers & Virtual Disks",
	}

	vhdxPath := findDockerVHDX()
	var vhdxSize int64
	if vhdxPath != "" {
		if fi, err := os.Stat(vhdxPath); err == nil {
			vhdxSize = fi.Size()
		}
		res.Diagnostics = append(res.Diagnostics, model.DiagnosticInfo{
			Name:   "Docker Desktop Virtual Disk (docker_data.vhdx)",
			Status: "Found",
			Size:   vhdxSize,
			Notes:  vhdxPath,
		})
	}

	_, scriptErr := os.Stat(defaultDockerCleanupScript)
	hasScript := scriptErr == nil

	// 1. Safe Docker Prune (Stopped containers, dangling images, unused networks, BuildKit cache)
	res.Items = append(res.Items, &model.TargetItem{
		ID:          "docker-prune-safe",
		StageID:     s.ID(),
		Name:        "Docker Safe Prune (Stopped containers, dangling images, networks & BuildKit cache)",
		Description: "Runs Invoke-DockerCleanup.ps1 -Prune -Force and docker builder prune -f",
		Path:        defaultDockerCleanupScript,
		SizeBytes:   0,
		Risk:        model.RiskSafe,
		Selected:    false,
		Notes:       "Requires Docker Engine running; reclaims internal space before VHDX compaction",
		Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
			if dryRun {
				if hasScript {
					return 0, fmt.Sprintf("[DRY-RUN] Would run: pwsh -File %q -Prune -Force", defaultDockerCleanupScript), nil
				}
				return 0, "[DRY-RUN] Would run: docker system prune -f && docker builder prune -f", nil
			}
			// Check if Docker daemon is reachable before pruning
			if _, pingErr := s.runner.Run(execCtx, "docker", "info"); pingErr != nil {
				return 0, "Docker daemon is not running; skipped live prune (start Docker Desktop to prune containers/images)", nil
			}
			if hasScript {
				cmd := fmt.Sprintf(`& %q -Prune -Force`, defaultDockerCleanupScript)
				out, err := s.runner.RunPS(execCtx, cmd)
				return 0, out, err
			}
			out1, err1 := s.runner.Run(execCtx, "docker", "system", "prune", "-f")
			out2, _ := s.runner.Run(execCtx, "docker", "builder", "prune", "-f")
			return 0, out1 + "\n" + out2, err1
		},
	})

	// 2. Deep Docker Prune (-Full -IncludeVolumes)
	res.Items = append(res.Items, &model.TargetItem{
		ID:          "docker-prune-deep",
		StageID:     s.ID(),
		Name:        "Docker Deep Prune (All unused images, all BuildKit cache & unused volumes)",
		Description: "Runs Invoke-DockerCleanup.ps1 -Full -IncludeVolumes -Force (WARNING: removes unused volumes)",
		Path:        defaultDockerCleanupScript,
		SizeBytes:   0,
		Risk:        model.RiskDestructive,
		Selected:    false,
		Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
			if dryRun {
				if hasScript {
					return 0, fmt.Sprintf("[DRY-RUN] Would run: pwsh -File %q -Full -IncludeVolumes -Force", defaultDockerCleanupScript), nil
				}
				return 0, "[DRY-RUN] Would run: docker system prune -a --volumes -f", nil
			}
			if _, pingErr := s.runner.Run(execCtx, "docker", "info"); pingErr != nil {
				return 0, "Docker daemon is not running; skipped deep prune", nil
			}
			if hasScript {
				cmd := fmt.Sprintf(`& %q -Full -IncludeVolumes -Force`, defaultDockerCleanupScript)
				out, err := s.runner.RunPS(execCtx, cmd)
				return 0, out, err
			}
			out, err := s.runner.Run(execCtx, "docker", "system", "prune", "-a", "--volumes", "-f")
			return 0, out, err
		},
	})

	// 3. Compact docker_data.vhdx
	if vhdxPath != "" && vhdxSize > 0 {
		res.Items = append(res.Items, &model.TargetItem{
			ID:            "docker-compact-vhdx",
			StageID:       s.ID(),
			Name:          "Compact Docker Virtual Disk (docker_data.vhdx)",
			Description:   "Shuts down WSL2/Docker utility VM and compacts docker_data.vhdx via Optimize-VHD / diskpart",
			Path:          vhdxPath,
			SizeBytes:     vhdxSize,
			Risk:          model.RiskModerate,
			RequiresAdmin: true,
			Selected:      false,
			Notes:         "Current VHDX file size on disk",
			Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
				return CompactVHDX(execCtx, s.runner, vhdxPath, dryRun)
			},
		})
	}

	res.ScanDuration = time.Since(start)
	return res, nil
}

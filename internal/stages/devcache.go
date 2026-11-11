package stages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"spacereclaim/internal/fsutil"
	"spacereclaim/internal/model"
	"spacereclaim/internal/sys"
)

type cacheTargetDef struct {
	id          string
	name        string
	description string
	path        string
	keepRoot    bool
	extraCmd    func(ctx context.Context, r sys.Runner)
}

// DevCacheStage scans and cleans Windows developer package manager caches, updater staging folders, and temp directories.
type DevCacheStage struct {
	runner sys.Runner
}

func NewDevCacheStage(runner sys.Runner) *DevCacheStage {
	return &DevCacheStage{runner: runner}
}

func (s *DevCacheStage) ID() string    { return "devcache" }
func (s *DevCacheStage) Title() string { return "Stage 3: Developer Package Caches, Updater Staging & Temp Folders" }
func (s *DevCacheStage) Description() string {
	return "Cleans npm, pnpm, uv, hatch, yarn, pip, SquirrelTemp, Ollama staged updates, User Temp, C:\\tmp, C:\\OneDriveTemp & Recycle Bin"
}

func (s *DevCacheStage) Scan(ctx context.Context, _ model.ScanOptions) (*model.StageScanResult, error) {
	start := time.Now()
	res := &model.StageScanResult{
		StageID:    s.ID(),
		StageTitle: s.Title(),
		Category:   "Developer Caches & Temporary Storage",
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	home, _ := os.UserHomeDir()
	if localAppData == "" {
		localAppData = filepath.Join(home, "AppData", "Local")
	}

	defs := []cacheTargetDef{
		{
			id:          "cache-scoop",
			name:        "Scoop Download Cache (~\\scoop\\cache)",
			description: "Clears downloaded Scoop installer archives in ~\\scoop\\cache",
			path:        filepath.Join(home, "scoop", "cache"),
			keepRoot:    true,
		},
		{
			id:          "cache-npm",
			name:        "npm Cache (AppData\\Local\\npm-cache)",
			description: "Clears Node Package Manager tarball and metadata cache",
			path:        filepath.Join(localAppData, "npm-cache"),
			keepRoot:    true,
		},
		{
			id:          "cache-pnpm",
			name:        "pnpm Cache & Unreferenced Store (AppData\\Local\\pnpm-cache)",
			description: "Removes pnpm metadata/dlx cache and runs 'pnpm store prune'",
			path:        filepath.Join(localAppData, "pnpm-cache"),
			keepRoot:    true,
			extraCmd: func(c context.Context, r sys.Runner) {
				_, _ = r.Run(c, "pnpm", "store", "prune")
			},
		},
		{
			id:          "cache-uv",
			name:        "uv Python Package Cache (AppData\\Local\\uv\\cache)",
			description: "Clears Astral uv wheel and source build cache",
			path:        filepath.Join(localAppData, "uv", "cache"),
			keepRoot:    true,
		},
		{
			id:          "cache-hatch",
			name:        "Hatch Python Environment & Wheel Cache (AppData\\Local\\hatch\\cache)",
			description: "Clears Hatch build/artifact cache",
			path:        filepath.Join(localAppData, "hatch", "cache"),
			keepRoot:    true,
		},
		{
			id:          "cache-yarn",
			name:        "Yarn Package Cache (AppData\\Local\\Yarn\\Cache)",
			description: "Clears Yarn package archives cache",
			path:        filepath.Join(localAppData, "Yarn", "Cache"),
			keepRoot:    true,
		},
		{
			id:          "cache-pip",
			name:        "pip HTTP & Wheel Cache (AppData\\Local\\pip\\cache)",
			description: "Clears Python pip wheel cache",
			path:        filepath.Join(localAppData, "pip", "cache"),
			keepRoot:    true,
		},
		{
			id:          "updater-squirrel",
			name:        "Squirrel Installer Staging (AppData\\Local\\SquirrelTemp)",
			description: "Removes orphaned Squirrel updater packages (Discord/Slack/Teams/etc.)",
			path:        filepath.Join(localAppData, "SquirrelTemp"),
			keepRoot:    false,
		},
		{
			id:          "updater-ollama",
			name:        "Ollama Staged Updates (AppData\\Local\\Ollama\\updates_v2)",
			description: "Removes already-applied or stale Ollama installer update payloads",
			path:        filepath.Join(localAppData, "Ollama", "updates_v2"),
			keepRoot:    true,
		},
		{
			id:          "temp-user",
			name:        "Windows User Temp Directory (AppData\\Local\\Temp)",
			description: "Removes unlocked temporary files and folders in %TEMP%",
			path:        filepath.Join(localAppData, "Temp"),
			keepRoot:    true,
		},
		{
			id:          "temp-ctmp",
			name:        "Root Temporary Directory (C:\\tmp)",
			description: "Cleans temporary artifacts stored in C:\\tmp",
			path:        `C:\tmp`,
			keepRoot:    true,
		},
		{
			id:          "temp-onedrive",
			name:        "OneDrive Temporary Staging (C:\\OneDriveTemp)",
			description: "Cleans temporary sync/chunk staging in C:\\OneDriveTemp",
			path:        `C:\OneDriveTemp`,
			keepRoot:    true,
		},
	}

	paths := make([]string, 0, len(defs))
	for _, d := range defs {
		paths = append(paths, d.path)
	}
	sizes := fsutil.MultiPathSize(ctx, paths)

	for _, d := range defs {
		def := d
		sz := sizes[def.path]
		_, statErr := os.Stat(def.path)
		exists := statErr == nil

		if !exists && sz == 0 {
			continue
		}

		res.Items = append(res.Items, &model.TargetItem{
			ID:          def.id,
			StageID:     s.ID(),
			Name:        def.name,
			Description: def.description,
			Path:        def.path,
			SizeBytes:   sz,
			Risk:        model.RiskSafe,
			Selected:    sz > 0,
			Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
				if dryRun {
					return sz, fmt.Sprintf("[DRY-RUN] Would clean %s (%s)", def.path, model.FormatBytes(sz)), nil
				}
				reclaimed, err := fsutil.RemovePath(def.path, def.keepRoot)
				if def.extraCmd != nil {
					def.extraCmd(execCtx, s.runner)
				}
				return reclaimed, fmt.Sprintf("Cleaned %s (reclaimed %s)", def.path, model.FormatBytes(reclaimed)), err
			},
		})
		res.TotalSize += sz
	}

	// Add Recycle Bin cleanup item
	res.Items = append(res.Items, &model.TargetItem{
		ID:          "temp-recycle-bin",
		StageID:     s.ID(),
		Name:        "Windows Recycle Bin (C:\\ & E:\\)",
		Description: "Empties the Windows Recycle Bin across local drives",
		Path:        `C:\$Recycle.Bin`,
		SizeBytes:   0,
		Risk:        model.RiskSafe,
		Selected:    true,
		Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
			script := `Clear-RecycleBin -DriveLetter C,E -Force -ErrorAction SilentlyContinue`
			if dryRun {
				return 0, "[DRY-RUN] Would run: Clear-RecycleBin -DriveLetter C,E -Force", nil
			}
			_, err := s.runner.RunPS(execCtx, script)
			return 0, "Emptied Windows Recycle Bin (C:\\, E:\\)", err
		},
	})

	res.ScanDuration = time.Since(start)
	return res, nil
}

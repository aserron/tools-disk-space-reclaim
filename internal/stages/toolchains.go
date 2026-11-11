package stages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"spacereclaim/internal/fsutil"
	"spacereclaim/internal/model"
	"spacereclaim/internal/sys"
)

// ToolchainsStage audits and cleans duplicate/old IDE installations, secondary NVM Node versions,
// duplicate Ruby installations, and browser on-device AI model bloat.
type ToolchainsStage struct {
	runner sys.Runner
}

func NewToolchainsStage(runner sys.Runner) *ToolchainsStage {
	return &ToolchainsStage{runner: runner}
}

func (s *ToolchainsStage) ID() string    { return "toolchains" }
func (s *ToolchainsStage) Title() string { return "Stage 5: IDEs, Language Toolchains (NVM/Ruby) & Browser Bloat" }
func (s *ToolchainsStage) Description() string {
	return "Audits JetBrains IDEs/caches, secondary NVM Node versions, duplicate Ruby installs, and Chrome On-Device AI models"
}

func (s *ToolchainsStage) Scan(ctx context.Context, _ model.ScanOptions) (*model.StageScanResult, error) {
	start := time.Now()
	res := &model.StageScanResult{
		StageID:    s.ID(),
		StageTitle: s.Title(),
		Category:   "IDEs, Toolchains & Browsers",
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	roamingAppData := os.Getenv("APPDATA")

	// 1. NVM for Windows (C:\nvm4w\nvm)
	nvmRoot := `C:\nvm4w\nvm`
	activeNodeOut, _ := s.runner.Run(ctx, "node", "-v")
	activeNode := strings.TrimSpace(activeNodeOut)

	var nvmVers []string
	var nvmPaths []string
	if entries, err := os.ReadDir(nvmRoot); err == nil {
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "v") {
				nvmVers = append(nvmVers, e.Name())
				nvmPaths = append(nvmPaths, filepath.Join(nvmRoot, e.Name()))
			}
		}
	}
	nvmSizes := fsutil.MultiPathSize(ctx, nvmPaths)

	for i, ver := range nvmVers {
		verPath := nvmPaths[i]
		sz := nvmSizes[verPath]

		if strings.EqualFold(ver, activeNode) {
			res.Diagnostics = append(res.Diagnostics, model.DiagnosticInfo{
				Name:   fmt.Sprintf("NVM Node.js %s (ACTIVE)", ver),
				Status: "Active Default",
				Size:   sz,
				Notes:  verPath,
			})
			continue
		}

		verCopy := ver
		pathCopy := verPath
		res.Items = append(res.Items, &model.TargetItem{
			ID:          fmt.Sprintf("nvm-node-%s", strings.ToLower(verCopy)),
			StageID:     s.ID(),
			Name:        fmt.Sprintf("NVM Secondary Node.js %s (Active is %s)", verCopy, activeNode),
			Description: fmt.Sprintf("Removes secondary Node.js installation at %s", pathCopy),
			Path:        pathCopy,
			SizeBytes:   sz,
			Risk:        model.RiskModerate,
			Selected:    false,
			Execute: func(_ context.Context, dryRun bool) (int64, string, error) {
				if dryRun {
					return sz, fmt.Sprintf("[DRY-RUN] Would remove Node.js %s at %s (%s)", verCopy, pathCopy, model.FormatBytes(sz)), nil
				}
				reclaimed, err := fsutil.RemovePath(pathCopy, false)
				return reclaimed, fmt.Sprintf("Removed Node.js %s (%s)", verCopy, model.FormatBytes(reclaimed)), err
			},
		})
		res.TotalSize += sz
	}

	// 2. Duplicate / Legacy Ruby installations outside active rbenv
	legacyRubyDirs := []struct {
		id   string
		name string
		path string
	}{
		{"ruby-rvm-progdata", "Legacy RVM Ruby Store (C:\\ProgramData\\rvm)", `C:\ProgramData\rvm`},
		{"ruby-c-ruby27", "Standalone RubyInstaller (C:\\Ruby27-x64)", `C:\Ruby27-x64`},
	}
	for _, lr := range legacyRubyDirs {
		item := lr
		sz := fsutil.PathSize(ctx, item.path)
		if sz == 0 {
			continue
		}
		res.Items = append(res.Items, &model.TargetItem{
			ID:          item.id,
			StageID:     s.ID(),
			Name:        item.name,
			Description: "Removes duplicate Ruby installation outside C:\\Ruby-on-Windows",
			Path:        item.path,
			SizeBytes:   sz,
			Risk:        model.RiskModerate,
			Selected:    false,
			Execute: func(_ context.Context, dryRun bool) (int64, string, error) {
				if dryRun {
					return sz, fmt.Sprintf("[DRY-RUN] Would remove %s (%s)", item.path, model.FormatBytes(sz)), nil
				}
				reclaimed, err := fsutil.RemovePath(item.path, false)
				return reclaimed, fmt.Sprintf("Removed %s (%s)", item.path, model.FormatBytes(reclaimed)), err
			},
		})
		res.TotalSize += sz
	}

	// Report active C:\Ruby-on-Windows versions & JetBrains folders concurrently
	var diagPaths []string
	type diagMeta struct {
		name string
		path string
		kind string
	}
	var diagMetas []diagMeta

	if entries, err := os.ReadDir(`C:\Ruby-on-Windows`); err == nil {
		for _, e := range entries {
			if e.IsDir() && e.Name() != "rbenv" {
				p := filepath.Join(`C:\Ruby-on-Windows`, e.Name())
				diagPaths = append(diagPaths, p)
				diagMetas = append(diagMetas, diagMeta{name: fmt.Sprintf("Ruby-on-Windows: %s", e.Name()), path: p, kind: "Installed"})
			}
		}
	}

	for _, root := range []string{
		`C:\Program Files Dev\JetBrains`,
		filepath.Join(localAppData, "JetBrains"),
		filepath.Join(roamingAppData, "JetBrains"),
	} {
		if entries, err := os.ReadDir(root); err == nil {
			for _, e := range entries {
				if e.IsDir() && e.Name() != "acp-agents" {
					p := filepath.Join(root, e.Name())
					diagPaths = append(diagPaths, p)
					diagMetas = append(diagMetas, diagMeta{name: fmt.Sprintf("JetBrains Folder: %s", e.Name()), path: p, kind: "Present"})
				}
			}
		}
	}

	diagSizes := fsutil.MultiPathSize(ctx, diagPaths)
	for _, dm := range diagMetas {
		sz := diagSizes[dm.path]
		if sz >= 50*model.MB {
			res.Diagnostics = append(res.Diagnostics, model.DiagnosticInfo{
				Name:   dm.name,
				Status: dm.kind,
				Size:   sz,
				Notes:  dm.path,
			})
		}
	}

	// 3. JetBrains acp-agents & Chrome On-Device AI Models
	if localAppData != "" {
		acpPath := filepath.Join(localAppData, "JetBrains", "acp-agents")
		chromeUserData := filepath.Join(localAppData, "Google", "Chrome", "User Data")
		optGuidePath := filepath.Join(chromeUserData, "OptGuideOnDeviceModel")
		snapshotsPath := filepath.Join(chromeUserData, "Snapshots")

		extraSizes := fsutil.MultiPathSize(ctx, []string{acpPath, optGuidePath, snapshotsPath})

		if sz := extraSizes[acpPath]; sz > 0 {
			res.Items = append(res.Items, &model.TargetItem{
				ID:          "jetbrains-acp-agents",
				StageID:     s.ID(),
				Name:        "JetBrains AI/ACP Agents Cache (AppData\\Local\\JetBrains\\acp-agents)",
				Description: "Clears downloaded JetBrains ACP agent binaries and caches",
				Path:        acpPath,
				SizeBytes:   sz,
				Risk:        model.RiskSafe,
				Selected:    false,
				Execute: func(_ context.Context, dryRun bool) (int64, string, error) {
					if dryRun {
						return sz, fmt.Sprintf("[DRY-RUN] Would remove %s (%s)", acpPath, model.FormatBytes(sz)), nil
					}
					reclaimed, err := fsutil.RemovePath(acpPath, false)
					return reclaimed, fmt.Sprintf("Removed %s (%s)", acpPath, model.FormatBytes(reclaimed)), err
				},
			})
			res.TotalSize += sz
		}

		browserTargets := []struct {
			id   string
			name string
			path string
			risk model.RiskLevel
		}{
			{
				id:   "chrome-optguide-model",
				name: "Chrome On-Device AI Model (User Data\\OptGuideOnDeviceModel)",
				path: optGuidePath,
				risk: model.RiskSafe,
			},
			{
				id:   "chrome-snapshots",
				name: "Chrome Upgrade Snapshots (User Data\\Snapshots)",
				path: snapshotsPath,
				risk: model.RiskSafe,
			},
		}
		for _, bt := range browserTargets {
			t := bt
			sz := extraSizes[t.path]
			if sz == 0 {
				continue
			}
			res.Items = append(res.Items, &model.TargetItem{
				ID:          t.id,
				StageID:     s.ID(),
				Name:        t.name,
				Description: "Removes regenerable Chrome browser model/snapshot data",
				Path:        t.path,
				SizeBytes:   sz,
				Risk:        t.risk,
				Selected:    false,
				Execute: func(_ context.Context, dryRun bool) (int64, string, error) {
					if dryRun {
						return sz, fmt.Sprintf("[DRY-RUN] Would remove %s (%s)", t.path, model.FormatBytes(sz)), nil
					}
					reclaimed, err := fsutil.RemovePath(t.path, true)
					return reclaimed, fmt.Sprintf("Cleaned %s (%s)", t.path, model.FormatBytes(reclaimed)), err
				},
			})
			res.TotalSize += sz
		}
	}

	res.ScanDuration = time.Since(start)
	return res, nil
}

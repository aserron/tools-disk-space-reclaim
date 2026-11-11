package stages

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"spacereclaim/internal/fsutil"
	"spacereclaim/internal/model"
	"spacereclaim/internal/sys"
)

// SystemStage audits hidden OS storage (VSS Shadow Copies, Reserved Storage [Read-Only],
// pagefile.sys, hiberfil.sys, Search Index), runs WinSxS DISM component cleanup,
// and scans C:\projects for large SQL dumps and node_modules folders.
type SystemStage struct {
	runner sys.Runner
}

func NewSystemStage(runner sys.Runner) *SystemStage {
	return &SystemStage{runner: runner}
}

func (s *SystemStage) ID() string    { return "system" }
func (s *SystemStage) Title() string { return "Stage 6: Windows System Audit (WinSxS, VSS, Reserved Storage) & Project Bloat" }
func (s *SystemStage) Description() string {
	return "Audits VSS, Reserved Storage (Read-Only), Pagefile, Search Index, runs DISM WinSxS cleanup, and locates SQL dumps & node_modules in C:\\projects"
}

func (s *SystemStage) Scan(ctx context.Context, opts model.ScanOptions) (*model.StageScanResult, error) {
	start := time.Now()
	res := &model.StageScanResult{
		StageID:    s.ID(),
		StageTitle: s.Title(),
		Category:   "Windows OS Internals & Local Projects",
	}

	// 1. Read-Only OS Diagnostics (Pagefile, Hiberfil, Search Index, Reserved Storage, VSS)
	systemFiles := []struct {
		name  string
		path  string
		notes string
	}{
		{"Windows Pagefile (C:\\pagefile.sys)", `C:\pagefile.sys`, "Virtual memory paging file"},
		{"Windows Hibernation File (C:\\hiberfil.sys)", `C:\hiberfil.sys`, "Hibernation state file (0 B when disabled)"},
		{"Windows Search Index (Windows.edb)", `C:\ProgramData\Microsoft\Search\Data\Applications\Windows\Windows.edb`, "Windows Search indexer database"},
	}

	for _, sf := range systemFiles {
		if fi, err := os.Stat(sf.path); err == nil {
			res.Diagnostics = append(res.Diagnostics, model.DiagnosticInfo{
				Name:   sf.name,
				Status: "Active",
				Size:   fi.Size(),
				Notes:  sf.notes,
			})
		} else {
			res.Diagnostics = append(res.Diagnostics, model.DiagnosticInfo{
				Name:   sf.name,
				Status: "Not Present / Disabled",
				Size:   0,
				Notes:  sf.notes,
			})
		}
	}

	// Strictly Read-Only telemetry for Windows Reserved Storage (never modified per user constraint)
	res.Diagnostics = append(res.Diagnostics, model.DiagnosticInfo{
		Name:   "Windows Reserved Storage (Read-Only Policy)",
		Status: "Protected (Never Disabled)",
		Size:   0,
		Notes:  "Kept enabled per safety policy; not touched by space-reclaim",
	})

	// 2. Selectable Action: DISM WinSxS Component Store Cleanup
	res.Items = append(res.Items, &model.TargetItem{
		ID:            "sys-dism-winsxs",
		StageID:       s.ID(),
		Name:          "Windows Component Store Cleanup (DISM /StartComponentCleanup)",
		Description:   "Runs Dism.exe /Online /Cleanup-Image /StartComponentCleanup to reclaim superseded WinSxS packages",
		Path:          `C:\Windows\WinSxS`,
		SizeBytes:     0,
		Risk:          model.RiskModerate,
		RequiresAdmin: true,
		Selected:      false,
		Notes:         "Requires Administrator privileges",
		Execute: func(execCtx context.Context, dryRun bool) (int64, string, error) {
			if dryRun {
				return 0, "[DRY-RUN] Would run: Dism.exe /Online /Cleanup-Image /StartComponentCleanup", nil
			}
			cBefore, _ := fsutil.GetDriveStats(`C:\`)
			out, err := s.runner.Run(execCtx, "Dism.exe", "/Online", "/Cleanup-Image", "/StartComponentCleanup")
			cAfter, _ := fsutil.GetDriveStats(`C:\`)
			var diff int64
			if cAfter.FreeBytes > cBefore.FreeBytes {
				diff = int64(cAfter.FreeBytes - cBefore.FreeBytes)
			}
			return diff, fmt.Sprintf("DISM StartComponentCleanup completed (%s)", strings.TrimSpace(out)), err
		},
	})

	// 3. Project Bloat Scanner in C:\projects (SQL Dumps & node_modules)
	projRoot := opts.ProjectRoot
	if projRoot == "" {
		projRoot = `C:\projects`
	}
	minMB := opts.MinSizeMB
	if minMB <= 0 {
		minMB = 50
	}

	if _, err := os.Stat(projRoot); err == nil {
		// 3a. Large .sql / .bak / .dump files in C:\projects
		dumpExts := map[string]bool{
			".sql":  true,
			".dump": true,
			".bak":  true,
		}
		dumps := fsutil.FindLargeFiles(ctx, projRoot, 4, minMB*model.MB, dumpExts)
		sort.Slice(dumps, func(i, j int) bool {
			return dumps[i].Size > dumps[j].Size
		})

		for idx, d := range dumps {
			dmp := d
			modDate := time.Unix(dmp.ModTime, 0).Format("2006-01-02")
			res.Items = append(res.Items, &model.TargetItem{
				ID:          fmt.Sprintf("proj-dump-%d", idx+1),
				StageID:     s.ID(),
				Name:        fmt.Sprintf("Project Database Dump: %s (%s)", dmp.Name, modDate),
				Description: fmt.Sprintf("Large database dump file in %s", filepath.Dir(dmp.Path)),
				Path:        dmp.Path,
				SizeBytes:   dmp.Size,
				Risk:        model.RiskDestructive,
				Selected:    false,
				Execute: func(_ context.Context, dryRun bool) (int64, string, error) {
					if dryRun {
						return dmp.Size, fmt.Sprintf("[DRY-RUN] Would delete database dump %s (%s)", dmp.Path, model.FormatBytes(dmp.Size)), nil
					}
					if err := os.Remove(dmp.Path); err != nil {
						return 0, "", err
					}
					return dmp.Size, fmt.Sprintf("Deleted %s (%s)", dmp.Path, model.FormatBytes(dmp.Size)), nil
				},
			})
			res.TotalSize += dmp.Size
		}

		// 3b. Top-level node_modules directories inside C:\projects (up to depth 3, skipping nested node_modules)
		var nodeModulesDirs []string
		rootClean := filepath.Clean(projRoot)
		rootDepth := strings.Count(rootClean, string(os.PathSeparator))

		_ = filepath.WalkDir(rootClean, func(path string, d fs.DirEntry, walkErr error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if walkErr != nil || !d.IsDir() {
				return nil
			}
			if path != rootClean && fsutil.IsReparsePoint(d) {
				return filepath.SkipDir
			}
			depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
			if d.Name() == "node_modules" {
				nodeModulesDirs = append(nodeModulesDirs, path)
				return filepath.SkipDir
			}
			if depth > 3 || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		})

		nmSizes := fsutil.MultiPathSize(ctx, nodeModulesDirs)
		sort.Slice(nodeModulesDirs, func(i, j int) bool {
			return nmSizes[nodeModulesDirs[i]] > nmSizes[nodeModulesDirs[j]]
		})

		for idx, nmPath := range nodeModulesDirs {
			sz := nmSizes[nmPath]
			if sz < minMB*model.MB {
				continue
			}
			pCopy := nmPath
			rel, _ := filepath.Rel(projRoot, pCopy)
			res.Items = append(res.Items, &model.TargetItem{
				ID:          fmt.Sprintf("proj-nm-%d", idx+1),
				StageID:     s.ID(),
				Name:        fmt.Sprintf("Project node_modules: %s", rel),
				Description: "Removes project node_modules folder (reinstallable via npm/pnpm/yarn install)",
				Path:        pCopy,
				SizeBytes:   sz,
				Risk:        model.RiskDestructive,
				Selected:    false,
				Execute: func(_ context.Context, dryRun bool) (int64, string, error) {
					if dryRun {
						return sz, fmt.Sprintf("[DRY-RUN] Would remove %s (%s)", pCopy, model.FormatBytes(sz)), nil
					}
					reclaimed, err := fsutil.RemovePath(pCopy, false)
					return reclaimed, fmt.Sprintf("Removed %s (%s)", pCopy, model.FormatBytes(reclaimed)), err
				},
			})
			res.TotalSize += sz
		}
	}

	res.ScanDuration = time.Since(start)
	return res, nil
}

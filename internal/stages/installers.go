package stages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"spacereclaim/internal/fsutil"
	"spacereclaim/internal/model"
	"spacereclaim/internal/sys"
)

// InstallersStage detects OEM driver unpack directories (C:\SWSetup, C:\Drivers, HP Downloads)
// and large standalone installer binaries in Downloads.
type InstallersStage struct {
	runner sys.Runner
}

func NewInstallersStage(runner sys.Runner) *InstallersStage {
	return &InstallersStage{runner: runner}
}

func (s *InstallersStage) ID() string    { return "installers" }
func (s *InstallersStage) Title() string { return "Stage 4: OEM Driver Unpack Folders & Stale Downloads Installers" }
func (s *InstallersStage) Description() string {
	return "Detects C:\\SWSetup, C:\\Drivers, HP Downloads unpack folders, and large .exe/.msi/.iso installers in Downloads"
}

func (s *InstallersStage) Scan(ctx context.Context, opts model.ScanOptions) (*model.StageScanResult, error) {
	start := time.Now()
	res := &model.StageScanResult{
		StageID:    s.ID(),
		StageTitle: s.Title(),
		Category:   "OEM Unpack Folders & Installers",
	}

	home, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(home, "Downloads")

	oemFolders := []struct {
		id   string
		name string
		path string
		desc string
	}{
		{
			id:   "oem-swsetup",
			name: "HP Software Setup Unpack Folder (C:\\SWSetup)",
			path: `C:\SWSetup`,
			desc: "Extracted HP SoftPaq driver installers left behind after installation",
		},
		{
			id:   "oem-cdrivers",
			name: "OEM Extracted Drivers Folder (C:\\Drivers)",
			path: `C:\Drivers`,
			desc: "Extracted OEM hardware driver packages in C:\\Drivers",
		},
		{
			id:   "oem-hp-downloads",
			name: "HP Support Assistant Downloads (Downloads\\HP Downloads)",
			path: filepath.Join(downloadsDir, "HP Downloads"),
			desc: "Downloaded HP driver/BIOS installer payloads",
		},
		{
			id:   "oem-amd",
			name: "AMD Driver Unpack Folder (C:\\AMD)",
			path: `C:\AMD`,
			desc: "Extracted AMD GPU/chipset installer staging folder",
		},
		{
			id:   "oem-nvidia",
			name: "NVIDIA Driver Unpack Folder (C:\\NVIDIA)",
			path: `C:\NVIDIA`,
			desc: "Extracted NVIDIA driver installer staging folder",
		},
	}

	paths := make([]string, 0, len(oemFolders))
	for _, f := range oemFolders {
		paths = append(paths, f.path)
	}
	sizes := fsutil.MultiPathSize(ctx, paths)

	for _, f := range oemFolders {
		folder := f
		sz := sizes[folder.path]
		if _, err := os.Stat(folder.path); err != nil && sz == 0 {
			continue
		}
		res.Items = append(res.Items, &model.TargetItem{
			ID:          folder.id,
			StageID:     s.ID(),
			Name:        folder.name,
			Description: folder.desc,
			Path:        folder.path,
			SizeBytes:   sz,
			Risk:        model.RiskSafe,
			Selected:    sz > 0,
			Execute: func(_ context.Context, dryRun bool) (int64, string, error) {
				if dryRun {
					return sz, fmt.Sprintf("[DRY-RUN] Would remove OEM unpack folder %s (%s)", folder.path, model.FormatBytes(sz)), nil
				}
				reclaimed, err := fsutil.RemovePath(folder.path, false)
				return reclaimed, fmt.Sprintf("Removed %s (reclaimed %s)", folder.path, model.FormatBytes(reclaimed)), err
			},
		})
		res.TotalSize += sz
	}

	// Scan Downloads for large installer binaries (*.exe, *.msi, *.iso, *.msix)
	minMB := opts.MinSizeMB
	if minMB <= 0 {
		minMB = 50
	}
	minBytes := minMB * model.MB
	installerExts := map[string]bool{
		".exe":  true,
		".msi":  true,
		".iso":  true,
		".msix": true,
	}

	candidates := fsutil.FindLargeFiles(ctx, downloadsDir, 2, minBytes, installerExts)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Size > candidates[j].Size
	})

	for idx, c := range candidates {
		cand := c
		modDate := time.Unix(cand.ModTime, 0).Format("2006-01-02")
		slug := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			if r >= 'A' && r <= 'Z' {
				return r + ('a' - 'A')
			}
			return '-'
		}, cand.Name)

		res.Items = append(res.Items, &model.TargetItem{
			ID:          fmt.Sprintf("installer-%d-%s", idx+1, slug),
			StageID:     s.ID(),
			Name:        fmt.Sprintf("Installer: %s (modified %s)", cand.Name, modDate),
			Description: fmt.Sprintf("Standalone installer binary in Downloads (>= %d MB)", minMB),
			Path:        cand.Path,
			SizeBytes:   cand.Size,
			Risk:        model.RiskModerate,
			Selected:    false,
			Execute: func(_ context.Context, dryRun bool) (int64, string, error) {
				if dryRun {
					return cand.Size, fmt.Sprintf("[DRY-RUN] Would delete installer %s (%s)", cand.Path, model.FormatBytes(cand.Size)), nil
				}
				if err := os.Remove(cand.Path); err != nil {
					return 0, "", err
				}
				return cand.Size, fmt.Sprintf("Deleted installer %s (%s)", cand.Path, model.FormatBytes(cand.Size)), nil
			},
		})
		res.TotalSize += cand.Size
	}

	res.ScanDuration = time.Since(start)
	return res, nil
}

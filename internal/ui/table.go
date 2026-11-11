package ui

import (
	"fmt"
	"strings"

	"spacereclaim/internal/model"
)

// PrintBanner prints the application header and Admin status badge.
func PrintBanner(isAdmin bool, dryRun bool) {
	adminBadge := model.ANSIYellow + "[STANDARD USER]" + model.ANSIReset
	if isAdmin {
		adminBadge = model.ANSIGreen + "[ADMINISTRATOR]" + model.ANSIReset
	}
	dryBadge := ""
	if dryRun {
		dryBadge = " " + model.ANSICyan + "[DRY-RUN MODE]" + model.ANSIReset
	}

	fmt.Printf("\n%s%s══════════════════════════════════════════════════════════════════════════════%s\n", model.ANSIBold, model.ANSICyan, model.ANSIReset)
	fmt.Printf("  %sSPACE-RECLAIM%s — Windows 11, WSL2 & Docker Disk Space Orchestrator %s%s\n",
		model.ANSIBold, model.ANSIReset, adminBadge, dryBadge)
	fmt.Printf("%s%s══════════════════════════════════════════════════════════════════════════════%s\n", model.ANSIBold, model.ANSICyan, model.ANSIReset)
}

// PrintDriveStats renders a summary table of local drives (e.g., C:\ and E:\).
func PrintDriveStats(title string, drives []model.DriveStats) {
	if len(drives) == 0 {
		return
	}
	fmt.Printf("\n%s==> %s%s\n", model.ANSIBold, title, model.ANSIReset)
	for _, d := range drives {
		bar := model.ProgressBar(d.UsedPercent, 24)
		fmt.Printf("  Drive %-3s  %s  Free: %s%-10s%s  Used: %-10s / %-10s\n",
			d.Drive,
			bar,
			model.ANSIBold+model.ANSIGreen,
			model.FormatUintBytes(d.FreeBytes),
			model.ANSIReset,
			model.FormatUintBytes(d.UsedBytes),
			model.FormatUintBytes(d.TotalBytes),
		)
	}
}

// IndexedItem pairs a global 1-based index with a TargetItem for interactive selection.
type IndexedItem struct {
	Index int
	Item  *model.TargetItem
}

// PrintScanReport renders all stage diagnostics and selectable items, returning the flattened IndexedItem slice.
func PrintScanReport(results []*model.StageScanResult, isAdmin bool) []IndexedItem {
	var indexed []IndexedItem
	nextIdx := 1

	for _, st := range results {
		fmt.Printf("\n%s%s┌── %s [%s]%s\n", model.ANSIBold, model.ANSICyan, st.StageTitle, st.StageID, model.ANSIReset)
		if st.ScanWarning != "" {
			fmt.Printf("│   %s! Warning: %s%s\n", model.ANSIYellow, st.ScanWarning, model.ANSIReset)
		}

		if len(st.Diagnostics) > 0 {
			fmt.Printf("│   %sTelemetry & Read-Only Audit:%s\n", model.ANSIDim, model.ANSIReset)
			for _, dg := range st.Diagnostics {
				szStr := ""
				if dg.Size > 0 {
					szStr = fmt.Sprintf(" (%s)", model.FormatBytes(dg.Size))
				}
				fmt.Printf("│     • %-48s %s%s%s\n", dg.Name, dg.Status, szStr, func() string {
					if dg.Notes != "" {
						return " — " + model.ANSIDim + dg.Notes + model.ANSIReset
					}
					return ""
				}())
			}
		}

		if len(st.Items) == 0 {
			fmt.Printf("└── %s(No actionable bloat items detected in this stage)%s\n", model.ANSIDim, model.ANSIReset)
			continue
		}

		fmt.Printf("│   %sSelectable Cleanup & Optimization Actions:%s\n", model.ANSIBold, model.ANSIReset)
		for _, it := range st.Items {
			idx := nextIdx
			nextIdx++
			indexed = append(indexed, IndexedItem{Index: idx, Item: it})

			check := "[ ]"
			if it.Selected {
				check = model.ANSIGreen + "[x]" + model.ANSIReset
			}

			sizeLabel := model.FormatBytes(it.SizeBytes)
			if it.SizeBytes == 0 {
				sizeLabel = "Dynamic"
			}

			adminTag := ""
			if it.RequiresAdmin {
				if isAdmin {
					adminTag = " " + model.ANSICyan + "[ADMIN-OK]" + model.ANSIReset
				} else {
					adminTag = " " + model.ANSIRed + "[NEEDS-ADMIN]" + model.ANSIReset
				}
			}

			fmt.Printf("│   %s %s#%02d%s %-13s %s%-10s%s %s%s\n",
				check,
				model.ANSIBold, idx, model.ANSIReset,
				model.ColorRisk(it.Risk),
				model.ANSIBold, sizeLabel, model.ANSIReset,
				it.Name,
				adminTag,
			)
			if it.Path != "" {
				fmt.Printf("│              %sPath: %s%s\n", model.ANSIDim, it.Path, model.ANSIReset)
			}
		}
		fmt.Printf("└── %sStage scan completed in %s%s\n", model.ANSIDim, st.ScanDuration.Round(1e6), model.ANSIReset)
	}

	PrintSelectionSummary(indexed)
	return indexed
}

// PrintSelectionSummary shows how many items are currently selected and their combined estimated size.
func PrintSelectionSummary(indexed []IndexedItem) {
	var selectedCount int
	var selectedBytes int64
	var totalBytes int64

	for _, entry := range indexed {
		if entry.Item.Risk == model.RiskSafe || entry.Item.Selected {
			// Avoid adding entire VHDX container sizes to raw cache totals unless selected
		}
		totalBytes += entry.Item.SizeBytes
		if entry.Item.Selected {
			selectedCount++
			selectedBytes += entry.Item.SizeBytes
		}
	}

	fmt.Printf("\n%s──────────────────────────────────────────────────────────────────────────────%s\n", model.ANSIDim, model.ANSIReset)
	fmt.Printf("  %sSelected Actions:%s %d / %d items   |   %sEstimated Target Footprint:%s %s%s%s\n",
		model.ANSIBold, model.ANSIReset,
		selectedCount, len(indexed),
		model.ANSIBold, model.ANSIReset,
		model.ANSIGreen+model.ANSIBold, model.FormatBytes(selectedBytes), model.ANSIReset,
	)
	fmt.Printf("%s──────────────────────────────────────────────────────────────────────────────%s\n", model.ANSIDim, model.ANSIReset)
}

// PrintExecutionSummary prints the per-item execution results and Before/After drive delta.
func PrintExecutionSummary(results []model.ExecResult, beforeDrives, afterDrives []model.DriveStats, dryRun bool) {
	fmt.Printf("\n%s%s══════════════════════════════════════════════════════════════════════════════%s\n", model.ANSIBold, model.ANSICyan, model.ANSIReset)
	if dryRun {
		fmt.Printf("  %sDRY-RUN EXECUTION SUMMARY (No files were modified)%s\n", model.ANSIBold, model.ANSIReset)
	} else {
		fmt.Printf("  %sEXECUTION SUMMARY%s\n", model.ANSIBold, model.ANSIReset)
	}
	fmt.Printf("%s%s══════════════════════════════════════════════════════════════════════════════%s\n", model.ANSIBold, model.ANSICyan, model.ANSIReset)

	var totalReclaimed int64
	for _, r := range results {
		status := model.ANSIGreen + "[OK]" + model.ANSIReset
		if r.Skipped {
			status = model.ANSIYellow + "[SKIPPED]" + model.ANSIReset
		} else if !r.Success {
			status = model.ANSIRed + "[ERROR]" + model.ANSIReset
		} else if r.DryRun {
			status = model.ANSICyan + "[DRY-RUN]" + model.ANSIReset
		}

		fmt.Printf("  %s %-55s (%s)\n", status, r.ItemName, r.Duration.Round(1e6))
		if r.Message != "" {
			for _, line := range strings.Split(strings.TrimSpace(r.Message), "\n") {
				if strings.TrimSpace(line) != "" {
					fmt.Printf("       %s%s%s\n", model.ANSIDim, line, model.ANSIReset)
				}
			}
		}
		if r.Err != nil {
			fmt.Printf("       %sError: %v%s\n", model.ANSIRed, r.Err, model.ANSIReset)
		}
		totalReclaimed += r.BytesReclaimed
	}

	if !dryRun && len(beforeDrives) > 0 && len(afterDrives) > 0 {
		fmt.Printf("\n%s==> Drive Space Comparison (Before vs. After)%s\n", model.ANSIBold, model.ANSIReset)
		afterMap := make(map[string]model.DriveStats)
		for _, d := range afterDrives {
			afterMap[d.Drive] = d
		}
		for _, b := range beforeDrives {
			if a, ok := afterMap[b.Drive]; ok {
				delta := int64(a.FreeBytes) - int64(b.FreeBytes)
				deltaStr := model.FormatBytes(delta)
				if delta >= 0 {
					deltaStr = "+" + deltaStr
				}
				fmt.Printf("  Drive %-3s  Free Before: %-10s -> Free After: %s%-10s%s  (Delta: %s%s%s)\n",
					b.Drive,
					model.FormatUintBytes(b.FreeBytes),
					model.ANSIGreen+model.ANSIBold, model.FormatUintBytes(a.FreeBytes), model.ANSIReset,
					model.ANSICyan+model.ANSIBold, deltaStr, model.ANSIReset,
				)
			}
		}
	} else if totalReclaimed > 0 {
		fmt.Printf("\n  %sEstimated Space Targeted:%s %s\n", model.ANSIBold, model.ANSIReset, model.FormatBytes(totalReclaimed))
	}
	fmt.Println()
}

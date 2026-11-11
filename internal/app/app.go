package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"spacereclaim/internal/fsutil"
	"spacereclaim/internal/model"
	"spacereclaim/internal/stages"
	"spacereclaim/internal/sys"
	"spacereclaim/internal/ui"
)

// Config holds runtime options resolved from CLI flags.
type Config struct {
	ScanOnly    bool
	Interactive bool
	StagesSpec  string
	MaxRisk     model.RiskLevel
	RiskSet     bool
	DryRun      bool
	AutoYes     bool
	MinSizeMB   int64
	ProjectRoot string
}

// PrintStagesCatalog prints all available stages and their descriptions.
func PrintStagesCatalog() {
	runner := sys.NewDryRunRunner()
	all := stages.AllStages(runner)
	fmt.Printf("\n%sAvailable Space-Reclaim Stages:%s\n\n", model.ANSIBold, model.ANSIReset)
	for _, st := range all {
		fmt.Printf("  %s%-12s%s %s\n", model.ANSICyan+model.ANSIBold, st.ID(), model.ANSIReset, st.Title())
		fmt.Printf("               %s%s%s\n\n", model.ANSIDim, st.Description(), model.ANSIReset)
	}
}

// Run executes the full Scan -> Select -> Execute -> Summary lifecycle.
func Run(ctx context.Context, cfg Config) error {
	isAdmin := sys.IsAdmin()
	ui.PrintBanner(isAdmin, cfg.DryRun)

	beforeDrives := fsutil.QueryDrives("C", "E")
	ui.PrintDriveStats("Current Disk Telemetry", beforeDrives)

	var runner sys.Runner = sys.NewOSRunner()
	selectedStages, err := stages.SelectStages(runner, cfg.StagesSpec)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s==> Scanning %d stage(s) for reclaimable space...%s\n", model.ANSIBold, len(selectedStages), model.ANSIReset)
	scanOpts := model.ScanOptions{
		MinSizeMB:   cfg.MinSizeMB,
		ProjectRoot: cfg.ProjectRoot,
	}

	var scanResults []*model.StageScanResult
	for _, st := range selectedStages {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		res, err := st.Scan(ctx, scanOpts)
		if err != nil {
			return fmt.Errorf("stage %s scan failed: %w", st.ID(), err)
		}
		// Apply MaxRisk pre-selection when user explicitly passed --risk or --yes
		if cfg.RiskSet || cfg.AutoYes {
			for _, item := range res.Items {
				item.Selected = item.Risk <= cfg.MaxRisk
			}
		}
		scanResults = append(scanResults, res)
	}

	indexed := ui.PrintScanReport(scanResults, isAdmin)

	if cfg.ScanOnly || len(indexed) == 0 {
		return nil
	}

	effectiveDryRun := cfg.DryRun
	if !cfg.AutoYes {
		outcome := ui.RunInteractiveSelector(os.Stdin, scanResults, indexed, isAdmin, cfg.DryRun)
		if !outcome.Proceed {
			return nil
		}
		effectiveDryRun = outcome.DryRun
	}

	// Execute all selected items in stage order
	var execResults []model.ExecResult
	fmt.Printf("\n%s==> Executing selected actions...%s\n", model.ANSIBold, model.ANSIReset)

	for _, entry := range indexed {
		it := entry.Item
		if !it.Selected {
			continue
		}
		if ctx.Err() != nil {
			break
		}

		if it.RequiresAdmin && !isAdmin && !effectiveDryRun {
			execResults = append(execResults, model.ExecResult{
				ItemID:   it.ID,
				ItemName: it.Name,
				StageID:  it.StageID,
				Skipped:  true,
				Message:  "Skipped: requires Administrator privileges (re-run space-reclaim from an elevated terminal)",
			})
			continue
		}

		fmt.Printf("  • Running %s#%02d%s %s...\n", model.ANSICyan, entry.Index, model.ANSIReset, it.Name)
		itemStart := time.Now()
		reclaimed, msg, execErr := it.Execute(ctx, effectiveDryRun)
		dur := time.Since(itemStart)

		execResults = append(execResults, model.ExecResult{
			ItemID:         it.ID,
			ItemName:       it.Name,
			StageID:        it.StageID,
			Success:        execErr == nil,
			DryRun:         effectiveDryRun,
			BytesReclaimed: reclaimed,
			Duration:       dur,
			Message:        msg,
			Err:            execErr,
		})
	}

	afterDrives := fsutil.QueryDrives("C", "E")
	ui.PrintExecutionSummary(execResults, beforeDrives, afterDrives, effectiveDryRun)
	return nil
}

package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"spacereclaim/internal/app"
	"spacereclaim/internal/model"
)

// Execute parses CLI flags and runs the space-reclaim application.
func Execute() int {
	flags := flag.NewFlagSet("space-reclaim", flag.ContinueOnError)
	flags.Usage = PrintUsage

	var (
		scanOnly    bool
		interactive bool
		listStages  bool
		stagesSpec  string
		riskStr     string
		dryRun      bool
		autoYes     bool
		minSizeMB   int64
		projectRoot string
	)

	flags.BoolVar(&scanOnly, "scan", false, "Scan and display the disk audit report without modifying anything")
	flags.BoolVar(&scanOnly, "s", false, "Shorthand for --scan")
	flags.BoolVar(&interactive, "interactive", true, "Launch interactive selector to choose which stages/items to execute (default)")
	flags.BoolVar(&interactive, "i", true, "Shorthand for --interactive")
	flags.BoolVar(&listStages, "list-stages", false, "List all available cleanup stages and exit")
	flags.StringVar(&stagesSpec, "stages", "all", "Comma-separated stages to include: wsl,docker,devcache,installers,toolchains,system,all")
	flags.StringVar(&riskStr, "risk", "safe", "Maximum risk level to pre-select: safe, moderate, destructive")
	flags.BoolVar(&dryRun, "dry-run", false, "Simulate execution and print commands without deleting or modifying files")
	flags.BoolVar(&dryRun, "n", false, "Shorthand for --dry-run")
	flags.BoolVar(&autoYes, "yes", false, "Execute pre-selected items immediately without interactive prompt")
	flags.BoolVar(&autoYes, "y", false, "Shorthand for --yes")
	flags.Int64Var(&minSizeMB, "min-size", 50, "Minimum size in MB for listing standalone installers, SQL dumps, or node_modules")
	flags.StringVar(&projectRoot, "projects-dir", `C:\projects`, "Root directory for scanning project database dumps and node_modules")

	if err := flags.Parse(os.Args[1:]); err != nil {
		return 1
	}

	if listStages {
		app.PrintStagesCatalog()
		return 0
	}

	riskSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "risk" {
			riskSet = true
		}
	})

	maxRisk, ok := model.ParseRiskLevel(riskStr)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid --risk %q (choose safe, moderate, or destructive)\n", riskStr)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := app.Config{
		ScanOnly:    scanOnly,
		Interactive: interactive,
		StagesSpec:  stagesSpec,
		MaxRisk:     maxRisk,
		RiskSet:     riskSet,
		DryRun:      dryRun,
		AutoYes:     autoYes,
		MinSizeMB:   minSizeMB,
		ProjectRoot: projectRoot,
	}

	if err := app.Run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

// PrintUsage prints the CLI help and usage examples.
func PrintUsage() {
	fmt.Println(`space-reclaim — Modular Windows 11, WSL2 & Docker Disk Space Audit & Reclamation CLI/TUI

Usage:
  space-reclaim [flags]

Stages:
  wsl          Stage 1: WSL2 Distros (internal Linux caches, fstrim + ext4.vhdx compaction, move off C:\)
  docker       Stage 2: Docker Engine Prune (Invoke-DockerCleanup.ps1) & docker_data.vhdx compaction
  devcache     Stage 3: Developer Caches (npm, pnpm, uv, hatch, yarn, pip) & Temp/Updaters (Squirrel, Ollama, RecycleBin)
  installers   Stage 4: OEM Driver Unpack Folders (C:\SWSetup, C:\Drivers, HP Downloads) & Downloads Installers
  toolchains   Stage 5: Duplicate IDEs & Toolchains (JetBrains, secondary NVM Node versions, Ruby, Chrome AI models)
  system       Stage 6: Windows System Audit (WinSxS DISM cleanup, VSS, Reserved Storage [Read-Only], SQL dumps & node_modules)

Examples:
  # Scan all stages and launch the interactive TUI selector (default)
  space-reclaim

  # Read-only audit report of all 6 stages and drive telemetry (zero modifications)
  space-reclaim --scan

  # Inspect only WSL2 and Docker stages interactively
  space-reclaim --stages wsl,docker

  # Simulate running all Safe and Moderate items across all stages (Dry-Run)
  space-reclaim --stages all --risk moderate --dry-run --yes

  # Run safe developer cache and temp cleanup non-interactively
  space-reclaim --stages devcache --risk safe --yes

Flags:`)
	flag.PrintDefaults()
}

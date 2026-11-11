package model

import (
	"context"
	"strings"
	"time"
)

// RiskLevel classifies the safety impact of a cleanup or optimization item.
type RiskLevel int

const (
	// RiskSafe represents easily regenerable caches, temp files, and safe prunes.
	RiskSafe RiskLevel = iota
	// RiskModerate represents VHDX compaction (stops WSL temporarily), secondary toolchain removal,
	// WinSxS component cleanup, or old installer deletion.
	RiskModerate
	// RiskDestructive represents irreversible operations such as WSL distro relocation/unregistration,
	// Docker volume pruning, or deleting SQL dumps / node_modules in project trees.
	RiskDestructive
)

// String returns the lowercase identifier for a RiskLevel.
func (r RiskLevel) String() string {
	switch r {
	case RiskSafe:
		return "safe"
	case RiskModerate:
		return "moderate"
	case RiskDestructive:
		return "destructive"
	default:
		return "unknown"
	}
}

// Label returns a human-readable badge label for terminal output.
func (r RiskLevel) Label() string {
	switch r {
	case RiskSafe:
		return "SAFE"
	case RiskModerate:
		return "MODERATE"
	case RiskDestructive:
		return "DESTRUCTIVE"
	default:
		return "UNKNOWN"
	}
}

// ParseRiskLevel converts a CLI flag string into a RiskLevel.
func ParseRiskLevel(s string) (RiskLevel, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "safe", "low":
		return RiskSafe, true
	case "moderate", "medium", "mod":
		return RiskModerate, true
	case "destructive", "high", "all", "full":
		return RiskDestructive, true
	default:
		return RiskSafe, false
	}
}

// DriveStats captures capacity and free/used space for a Windows logical drive.
type DriveStats struct {
	Drive       string
	TotalBytes  uint64
	FreeBytes   uint64
	UsedBytes   uint64
	UsedPercent float64
}

// DiagnosticInfo holds read-only system audit telemetry (e.g., VSS, Reserved Storage, Pagefile).
type DiagnosticInfo struct {
	Name   string
	Status string
	Size   int64
	Notes  string
}

// ExecResult reports the outcome of running a single TargetItem's cleanup action.
type ExecResult struct {
	ItemID         string
	ItemName       string
	StageID        string
	Success        bool
	Skipped        bool
	DryRun         bool
	BytesReclaimed int64
	Duration       time.Duration
	Message        string
	Err            error
}

// ActionFunc executes the cleanup or optimization for a TargetItem.
// It returns the estimated or measured reclaimed bytes, a status message, and any error.
type ActionFunc func(ctx context.Context, dryRun bool) (int64, string, error)

// TargetItem represents a single selectable action inside a Stage.
type TargetItem struct {
	ID            string
	StageID       string
	Name          string
	Description   string
	Path          string
	SizeBytes     int64
	Risk          RiskLevel
	RequiresAdmin bool
	Selected      bool
	ReadOnlyInfo  bool
	Notes         string
	Execute       ActionFunc
}

// ScanOptions configures parameters passed to Stages during discovery.
type ScanOptions struct {
	MinSizeMB   int64
	ProjectRoot string
}

// StageScanResult holds all discovered TargetItems and read-only diagnostics for a Stage.
type StageScanResult struct {
	StageID       string
	StageTitle    string
	Category      string
	Diagnostics   []DiagnosticInfo
	Items         []*TargetItem
	TotalSize     int64
	ScanDuration  time.Duration
	ScanWarning   string
}

// Stage defines the contract for a modular disk-space audit and reclamation stage.
type Stage interface {
	// ID returns the short CLI identifier (e.g., "wsl", "docker", "devcache").
	ID() string
	// Title returns the human-friendly name of the stage.
	Title() string
	// Description explains what the stage inspects and cleans.
	Description() string
	// Scan inspects the system and returns candidate cleanup items and diagnostics.
	Scan(ctx context.Context, opts ScanOptions) (*StageScanResult, error)
}

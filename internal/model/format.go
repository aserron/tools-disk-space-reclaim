package model

import (
	"fmt"
	"strings"
)

const (
	Byte = 1
	KB   = 1024 * Byte
	MB   = 1024 * KB
	GB   = 1024 * MB
	TB   = 1024 * GB
)

// ANSI color codes for modern Windows Terminal / PowerShell 7 output.
const (
	ANSIReset   = "\033[0m"
	ANSIBold    = "\033[1m"
	ANSIDim     = "\033[2m"
	ANSIRed     = "\033[31m"
	ANSIGreen   = "\033[32m"
	ANSIYellow  = "\033[33m"
	ANSIBlue    = "\033[34m"
	ANSIMagenta = "\033[35m"
	ANSICyan    = "\033[36m"
	ANSIWhite   = "\033[37m"
)

// FormatBytes converts a byte count into a human-readable string (e.g., "14.36 GB", "654.2 MB").
func FormatBytes(bytes int64) string {
	if bytes < 0 {
		return "-" + FormatBytes(-bytes)
	}
	if bytes == 0 {
		return "0 B"
	}
	b := float64(bytes)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", b/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", b/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", b/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", b/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatUintBytes formats an unsigned byte count into a human-readable string.
func FormatUintBytes(bytes uint64) string {
	return FormatBytes(int64(bytes))
}

// ColorRisk returns the colored badge for a RiskLevel.
func ColorRisk(r RiskLevel) string {
	switch r {
	case RiskSafe:
		return ANSIGreen + "[SAFE]" + ANSIReset
	case RiskModerate:
		return ANSIYellow + "[MODERATE]" + ANSIReset
	case RiskDestructive:
		return ANSIRed + "[DESTRUCTIVE]" + ANSIReset
	default:
		return "[UNKNOWN]"
	}
}

// ProgressBar renders an ASCII/Unicode usage bar of the given width.
func ProgressBar(percent float64, width int) string {
	if width <= 0 {
		width = 20
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int((percent / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	color := ANSIGreen
	if percent >= 85.0 {
		color = ANSIRed
	} else if percent >= 70.0 {
		color = ANSIYellow
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("%s%s%s %.1f%%", color, bar, ANSIReset, percent)
}

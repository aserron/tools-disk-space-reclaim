package model

import (
	"strings"
	"testing"
)

func TestParseRiskLevel(t *testing.T) {
	tests := []struct {
		input string
		want  RiskLevel
		ok    bool
	}{
		{"safe", RiskSafe, true},
		{"SAFE", RiskSafe, true},
		{"moderate", RiskModerate, true},
		{"medium", RiskModerate, true},
		{"destructive", RiskDestructive, true},
		{"all", RiskDestructive, true},
		{"invalid", RiskSafe, false},
	}

	for _, tc := range tests {
		got, ok := ParseRiskLevel(tc.input)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ParseRiskLevel(%q) = (%v, %v), want (%v, %v)", tc.input, got, ok, tc.want, tc.ok)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	if got := FormatBytes(0); got != "0 B" {
		t.Errorf("FormatBytes(0) = %q, want 0 B", got)
	}
	if got := FormatBytes(512 * MB); got != "512.0 MB" {
		t.Errorf("FormatBytes(512MB) = %q, want 512.0 MB", got)
	}
	if got := FormatBytes(10 * GB); got != "10.00 GB" {
		t.Errorf("FormatBytes(10GB) = %q, want 10.00 GB", got)
	}
}

func TestProgressBar(t *testing.T) {
	bar := ProgressBar(50.0, 10)
	if !strings.Contains(bar, "50.0%") {
		t.Errorf("ProgressBar(50.0, 10) = %q, expected percentage label", bar)
	}
}

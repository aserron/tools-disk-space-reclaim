package sys

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os/exec"
	"strings"
	"unicode/utf16"
)

// Runner abstracts external command execution so stages can be tested and run safely in DryRun mode.
type Runner interface {
	// Run executes a binary with arguments and returns combined/stdout output.
	Run(ctx context.Context, name string, args ...string) (string, error)
	// RunPS executes a PowerShell (pwsh) snippet non-interactively.
	RunPS(ctx context.Context, script string) (string, error)
	// RunWSL executes a bash command inside a specific WSL2 distribution as the given user (e.g., "root").
	RunWSL(ctx context.Context, distro, user, bashCmd string) (string, error)
	// RunWSLHost executes wsl.exe host management commands (e.g., --list, --shutdown, --manage)
	// and automatically decodes UTF-16LE output if present.
	RunWSLHost(ctx context.Context, args ...string) (string, error)
}

// OSRunner executes real commands on the host Windows OS.
type OSRunner struct{}

// NewOSRunner returns a production Runner.
func NewOSRunner() *OSRunner {
	return &OSRunner{}
}

func (r *OSRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func (r *OSRunner) RunPS(ctx context.Context, script string) (string, error) {
	return r.Run(ctx, "pwsh", "-NoProfile", "-NonInteractive", "-Command", script)
}

func (r *OSRunner) RunWSL(ctx context.Context, distro, user, bashCmd string) (string, error) {
	args := []string{"-d", distro}
	if user != "" {
		args = append(args, "-u", user)
	}
	args = append(args, "--", "bash", "-c", bashCmd)
	return r.Run(ctx, "wsl.exe", args...)
}

func (r *OSRunner) RunWSLHost(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "wsl.exe", args...)
	raw, err := cmd.CombinedOutput()
	decoded := DecodeUTF16IfNeeded(raw)
	return strings.TrimSpace(decoded), err
}

// DecodeUTF16IfNeeded converts UTF-16LE bytes (emitted by wsl.exe) into a standard UTF-8 string.
func DecodeUTF16IfNeeded(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// Detect NUL bytes characteristic of UTF-16LE ASCII text from wsl.exe
	nulCount := 0
	for i := 1; i < len(b) && i < 64; i += 2 {
		if b[i] == 0 {
			nulCount++
		}
	}
	if nulCount >= 2 && len(b)%2 == 0 {
		u16 := make([]uint16, len(b)/2)
		for i := 0; i < len(u16); i++ {
			u16[i] = binary.LittleEndian.Uint16(b[i*2 : i*2+2])
		}
		// Strip BOM if present
		if len(u16) > 0 && u16[0] == 0xFEFF {
			u16 = u16[1:]
		}
		runes := utf16.Decode(u16)
		return strings.ReplaceAll(string(runes), "\x00", "")
	}
	return strings.ReplaceAll(string(b), "\x00", "")
}

// DryRunRunner records commands instead of executing destructive mutations.
type DryRunRunner struct {
	Recorded []string
}

func NewDryRunRunner() *DryRunRunner {
	return &DryRunRunner{}
}

func (d *DryRunRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	entry := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	d.Recorded = append(d.Recorded, entry)
	return "[DRY-RUN] " + entry, nil
}

func (d *DryRunRunner) RunPS(_ context.Context, script string) (string, error) {
	entry := fmt.Sprintf("pwsh -NoProfile -NonInteractive -Command %q", script)
	d.Recorded = append(d.Recorded, entry)
	return "[DRY-RUN] " + entry, nil
}

func (d *DryRunRunner) RunWSL(_ context.Context, distro, user, bashCmd string) (string, error) {
	entry := fmt.Sprintf("wsl.exe -d %s -u %s -- bash -c %q", distro, user, bashCmd)
	d.Recorded = append(d.Recorded, entry)
	return "[DRY-RUN] " + entry, nil
}

func (d *DryRunRunner) RunWSLHost(_ context.Context, args ...string) (string, error) {
	entry := fmt.Sprintf("wsl.exe %s", strings.Join(args, " "))
	d.Recorded = append(d.Recorded, entry)
	return "[DRY-RUN] " + entry, nil
}

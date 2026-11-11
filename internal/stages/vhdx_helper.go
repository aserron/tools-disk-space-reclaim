package stages

import (
	"context"
	"fmt"
	"os"
	"time"

	"spacereclaim/internal/model"
	"spacereclaim/internal/sys"
)

// CompactVHDX shuts down WSL2 to release VM file handles and compacts the target .vhdx file
// using Hyper-V Optimize-VHD (with automatic fallback to diskpart compact vdisk).
func CompactVHDX(ctx context.Context, runner sys.Runner, vhdxPath string, dryRun bool) (int64, string, error) {
	fiBefore, err := os.Stat(vhdxPath)
	if err != nil {
		return 0, "", fmt.Errorf("cannot stat VHDX %s: %w", vhdxPath, err)
	}
	beforeBytes := fiBefore.Size()

	if dryRun {
		return 0, fmt.Sprintf("[DRY-RUN] Would run 'wsl --shutdown' and 'Optimize-VHD -Path %q -Mode Full' (current VHDX size: %s)",
			vhdxPath, model.FormatBytes(beforeBytes)), nil
	}

	// Ensure WSL utility VM releases file locks on the VHDX
	_, _ = runner.RunWSLHost(ctx, "--shutdown")
	time.Sleep(2 * time.Second)

	psScript := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$vhdx = %q
if (Get-Command Optimize-VHD -ErrorAction SilentlyContinue) {
    Optimize-VHD -Path $vhdx -Mode Full
} else {
    $dp = "select vdisk file=""$vhdx""`+"\r\n"+`attach vdisk readonly`+"\r\n"+`compact vdisk`+"\r\n"+`detach vdisk`+"\r\n"+`exit"
    $tmp = [System.IO.Path]::GetTempFileName()
    Set-Content -Path $tmp -Value $dp -Encoding ASCII
    try {
        diskpart /s $tmp | Out-Null
    } finally {
        Remove-Item $tmp -Force -ErrorAction SilentlyContinue
    }
}
`, vhdxPath)

	out, err := runner.RunPS(ctx, psScript)
	if err != nil {
		return 0, out, fmt.Errorf("VHDX compaction failed for %s: %w (%s)", vhdxPath, err, out)
	}

	fiAfter, statErr := os.Stat(vhdxPath)
	if statErr != nil {
		return 0, "Compacted VHDX", nil
	}
	afterBytes := fiAfter.Size()
	reclaimed := beforeBytes - afterBytes
	if reclaimed < 0 {
		reclaimed = 0
	}

	msg := fmt.Sprintf("Shrank %s from %s -> %s (saved %s)",
		vhdxPath, model.FormatBytes(beforeBytes), model.FormatBytes(afterBytes), model.FormatBytes(reclaimed))
	return reclaimed, msg, nil
}

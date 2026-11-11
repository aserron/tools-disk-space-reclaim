package fsutil

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"spacereclaim/internal/model"
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceExW = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// GetDriveStats queries Windows GetDiskFreeSpaceExW for a logical drive (e.g. "C:\\" or "E:\\").
func GetDriveStats(drive string) (model.DriveStats, error) {
	d := strings.TrimSpace(drive)
	if len(d) == 1 {
		d += `:\`
	} else if len(d) == 2 && d[1] == ':' {
		d += `\`
	}

	pathPtr, err := syscall.UTF16PtrFromString(d)
	if err != nil {
		return model.DriveStats{}, err
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	ret, _, callErr := getDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if ret == 0 {
		return model.DriveStats{}, fmt.Errorf("GetDiskFreeSpaceExW(%s) failed: %w", d, callErr)
	}

	used := totalBytes - totalFreeBytes
	var pct float64
	if totalBytes > 0 {
		pct = (float64(used) / float64(totalBytes)) * 100.0
	}

	return model.DriveStats{
		Drive:       strings.ToUpper(d[:2]),
		TotalBytes:  totalBytes,
		FreeBytes:   totalFreeBytes,
		UsedBytes:   used,
		UsedPercent: pct,
	}, nil
}

// QueryDrives returns DriveStats for the given list of drives, skipping drives that are not mounted.
func QueryDrives(drives ...string) []model.DriveStats {
	var out []model.DriveStats
	for _, d := range drives {
		if st, err := GetDriveStats(d); err == nil && st.TotalBytes > 0 {
			out = append(out, st)
		}
	}
	return out
}

package fsutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPathSizeAndRemovePath(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "cache")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	data := make([]byte, 4096)
	fileA := filepath.Join(sub, "a.bin")
	fileB := filepath.Join(sub, "b.bin")
	if err := os.WriteFile(fileA, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fileB, data, 0o644); err != nil {
		t.Fatal(err)
	}

	sz := PathSize(context.Background(), sub)
	if sz != 8192 {
		t.Fatalf("PathSize = %d, want 8192", sz)
	}

	reclaimed, err := RemovePath(sub, true)
	if err != nil {
		t.Fatalf("RemovePath error: %v", err)
	}
	if reclaimed != 8192 {
		t.Fatalf("reclaimed = %d, want 8192", reclaimed)
	}
	if _, err := os.Stat(sub); err != nil {
		t.Fatalf("expected root directory %s to be preserved when keepRootDir=true", sub)
	}
}

func TestGetDriveStats(t *testing.T) {
	st, err := GetDriveStats(`C:\`)
	if err != nil {
		t.Fatalf("GetDriveStats(C:\\) failed: %v", err)
	}
	if st.TotalBytes == 0 || st.FreeBytes == 0 {
		t.Fatalf("unexpected zero drive stats: %+v", st)
	}
}

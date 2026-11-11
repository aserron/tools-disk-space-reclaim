package fsutil

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

const fileAttributeReparsePoint = 0x0400

// IsReparsePoint checks whether a DirEntry or FileInfo is a symlink or Windows junction/reparse point.
func IsReparsePoint(d fs.DirEntry) bool {
	if d.Type()&os.ModeSymlink != 0 || d.Type()&os.ModeIrregular != 0 {
		return true
	}
	info, err := d.Info()
	if err != nil {
		return false
	}
	if sys, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		if sys.FileAttributes&fileAttributeReparsePoint != 0 {
			return true
		}
	}
	return false
}

// PathSize returns the total size in bytes of a file or directory tree.
// It safely skips inaccessible subdirectories and Windows reparse points (junctions/symlinks).
func PathSize(ctx context.Context, root string) int64 {
	info, err := os.Lstat(root)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}

	var total int64
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if walkErr != nil {
			return nil // Ignore inaccessible paths
		}
		if path != root && IsReparsePoint(d) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			if fi, err := d.Info(); err == nil {
				total += fi.Size()
			}
		}
		return nil
	})
	return total
}

// MultiPathSize computes the sum of PathSize across multiple paths concurrently.
func MultiPathSize(ctx context.Context, paths []string) map[string]int64 {
	results := make(map[string]int64, len(paths))
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 8)
	for _, p := range paths {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			sz := PathSize(ctx, p)
			mu.Lock()
			results[p] = sz
			mu.Unlock()
		}()
	}
	wg.Wait()
	return results
}

// RemoveContents deletes all files and subdirectories inside dirPath without deleting dirPath itself,
// or deletes dirPath entirely if removeRoot is true. Best-effort for locked files (e.g., Temp).
func RemovePath(dirOrFile string, keepRootDir bool) (int64, error) {
	before := PathSize(context.Background(), dirOrFile)
	if before == 0 {
		if !keepRootDir {
			_ = os.RemoveAll(dirOrFile)
		}
		return 0, nil
	}

	info, err := os.Lstat(dirOrFile)
	if err != nil {
		return 0, nil
	}
	if !info.IsDir() || !keepRootDir {
		err := os.RemoveAll(dirOrFile)
		after := PathSize(context.Background(), dirOrFile)
		reclaimed := before - after
		if reclaimed < 0 {
			reclaimed = 0
		}
		return reclaimed, err
	}

	entries, err := os.ReadDir(dirOrFile)
	if err != nil {
		return 0, err
	}
	var lastErr error
	for _, e := range entries {
		child := filepath.Join(dirOrFile, e.Name())
		if err := os.RemoveAll(child); err != nil {
			lastErr = err
		}
	}
	after := PathSize(context.Background(), dirOrFile)
	reclaimed := before - after
	if reclaimed < 0 {
		reclaimed = 0
	}
	return reclaimed, lastErr
}

// FindLargeFiles searches a directory (non-recursively or up to maxDepth) for files matching extensions and >= minBytes.
type FileCandidate struct {
	Path    string
	Name    string
	Size    int64
	ModTime int64
}

func FindLargeFiles(ctx context.Context, root string, maxDepth int, minBytes int64, exts map[string]bool) []FileCandidate {
	var out []FileCandidate
	rootClean := filepath.Clean(root)
	rootDepth := strings.Count(rootClean, string(os.PathSeparator))

	_ = filepath.WalkDir(rootClean, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			return nil
		}
		if path != rootClean && IsReparsePoint(d) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
		if d.IsDir() {
			if maxDepth > 0 && depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if len(exts) > 0 && !exts[ext] {
			return nil
		}
		fi, err := d.Info()
		if err != nil || fi.Size() < minBytes {
			return nil
		}
		out = append(out, FileCandidate{
			Path:    path,
			Name:    d.Name(),
			Size:    fi.Size(),
			ModTime: fi.ModTime().Unix(),
		})
		return nil
	})
	return out
}

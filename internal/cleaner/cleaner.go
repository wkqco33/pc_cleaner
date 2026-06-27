// Package cleaner handles the actual deletion of cache items.
package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/seomini/pc_cleaner/internal/scanner"
	"github.com/seomini/pc_cleaner/internal/ui"
)

// Result holds the outcome of cleaning a single item.
type Result struct {
	Item    scanner.CacheItem
	Freed   int64
	Success bool
	Error   error
}

// Clean removes all provided scan results (dry-run skips actual deletion).
func Clean(results []scanner.ScanResult, dryRun bool) []Result {
	var cleaned []Result

	for _, r := range results {
		if !r.Exists || r.Error != nil {
			continue
		}

		res := Result{Item: r.Item}

		if dryRun {
			fmt.Printf("  %s [dry-run] %s\n", ui.Gray("→"), r.Item.Name)
			res.Success = true
			res.Freed = r.Size // dry-run은 스캔 추정치를 그대로 표시
			cleaned = append(cleaned, res)
			continue
		}

		switch r.Item.Type {
		case scanner.TypeDir:
			res.Freed, res.Error = cleanDir(r.Item.Path)
			res.Success = res.Error == nil
		case scanner.TypeGlob:
			res.Freed, res.Error = cleanGlob(r.Item.Path)
			res.Success = res.Error == nil
		case scanner.TypeCommand:
			res.Success, res.Error = runCommand(r.Item.Command)
		}

		if res.Success {
			fmt.Printf("  %s %s\n", ui.Green("✓"), r.Item.Name)
		} else {
			fmt.Printf("  %s %s: %v\n", ui.Red("✗"), r.Item.Name, res.Error)
		}
		cleaned = append(cleaned, res)
	}

	return cleaned
}

// cleanDir removes directory contents but keeps the directory itself.
// It returns the number of bytes actually freed (successful removals only).
func cleanDir(path string) (int64, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, fmt.Errorf("읽기 실패: %w", err)
	}

	var freed int64
	var lastErr error
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		size := pathSize(fullPath)
		if err := os.RemoveAll(fullPath); err != nil {
			lastErr = err
			continue
		}
		freed += size
	}

	return freed, lastErr
}

// cleanGlob deletes the contents of every directory (or each file) matching the
// wildcard pattern, returning the total bytes actually freed.
func cleanGlob(pattern string) (int64, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return 0, fmt.Errorf("패턴 오류: %w", err)
	}

	var freed int64
	var lastErr error
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			continue
		}
		if info.IsDir() {
			f, e := cleanDir(m)
			freed += f
			if e != nil {
				lastErr = e
			}
			continue
		}
		size := info.Size()
		if e := os.Remove(m); e != nil {
			lastErr = e
			continue
		}
		freed += size
	}

	return freed, lastErr
}

// pathSize returns the total size of a file or directory tree.
func pathSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

// runCommand executes an external command.
func runCommand(args []string) (bool, error) {
	if len(args) == 0 {
		return false, fmt.Errorf("명령어 없음")
	}

	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%w\n%s", err, string(out))
	}
	return true, nil
}

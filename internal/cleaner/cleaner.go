// Package cleaner handles the actual deletion of cache items.
package cleaner

import (
	"fmt"
	"os"
	"os/exec"

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

		res := Result{Item: r.Item, Freed: r.Size}

		if dryRun {
			fmt.Printf("  %s [dry-run] %s\n", ui.Gray("→"), r.Item.Name)
			res.Success = true
			cleaned = append(cleaned, res)
			continue
		}

		switch r.Item.Type {
		case scanner.TypeDir:
			res.Success, res.Error = cleanDir(r.Item.Path)
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
func cleanDir(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("읽기 실패: %w", err)
	}

	var lastErr error
	for _, entry := range entries {
		fullPath := path + string(os.PathSeparator) + entry.Name()
		if err := os.RemoveAll(fullPath); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return false, lastErr
	}
	return true, nil
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

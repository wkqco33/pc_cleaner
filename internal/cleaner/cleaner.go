// Package cleaner handles the actual deletion of cache items.
package cleaner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wkqco33/pc_cleaner/internal/scanner"
	"github.com/wkqco33/wcli/rich"
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
	return CleanTo(os.Stdout, results, dryRun)
}

// CleanTo removes scan results and writes progress to out.
func CleanTo(out io.Writer, results []scanner.ScanResult, dryRun bool) []Result {
	if out == nil {
		out = io.Discard
	}
	var cleaned []Result

	for _, r := range results {
		if !r.Exists || r.Error != nil {
			continue
		}

		res := Result{Item: r.Item}

		if dryRun {
			rich.Fprintln(out, "  [dim]→[/dim] [dry-run] %s", r.Item.Name)
			res.Success = true
			res.Freed = r.Size // dry-run은 스캔 추정치를 그대로 표시
			cleaned = append(cleaned, res)
			continue
		}

		if err := validateTarget(r.Item); err != nil {
			res.Error = err
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
			rich.Fprintln(out, "  [green]✓[/green] %s", r.Item.Name)
		} else {
			rich.Fprintln(out, "  [red]✗[/red] %s: %v", r.Item.Name, res.Error)
		}
		cleaned = append(cleaned, res)
	}

	return cleaned
}

func validateTarget(item scanner.CacheItem) error {
	if item.Type != scanner.TypeDir && item.Type != scanner.TypeGlob {
		return nil
	}
	if strings.TrimSpace(item.Path) == "" {
		return fmt.Errorf("삭제 경로가 비어 있습니다")
	}
	cleaned := filepath.Clean(item.Path)
	if cleaned == string(filepath.Separator) || cleaned == "." {
		return fmt.Errorf("위험한 삭제 경로를 거부했습니다: %q", item.Path)
	}
	if home, err := os.UserHomeDir(); err == nil && cleaned == filepath.Clean(home) {
		return fmt.Errorf("홈 디렉터리 전체 삭제를 거부했습니다")
	}
	return nil
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

func sameArgs(got []string, want ...string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
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

	var cmd *exec.Cmd
	switch filepath.Base(args[0]) {
	case "docker":
		if !sameArgs(args, "docker", "system", "prune", "-f") {
			return false, fmt.Errorf("허용되지 않은 docker 인자")
		}
	case "journalctl":
		if !sameArgs(args, "journalctl", "--vacuum-size=100M") {
			return false, fmt.Errorf("허용되지 않은 journalctl 인자")
		}
	case "whoami", "echo":
		// 테스트 및 안전한 진단 명령.
	default:
		return false, fmt.Errorf("허용되지 않은 정리 명령: %q", args[0])
	}
	switch filepath.Base(args[0]) {
	case "docker":
		cmd = exec.Command("docker", args[1:]...)
	case "journalctl":
		cmd = exec.Command("journalctl", args[1:]...)
	case "whoami":
		cmd = exec.Command("whoami", args[1:]...)
	case "echo":
		cmd = exec.Command("echo", args[1:]...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%w\n%s", err, string(out))
	}
	return true, nil
}

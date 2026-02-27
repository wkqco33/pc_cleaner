package cleaner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seomini/pc_cleaner/internal/cleaner"
	"github.com/seomini/pc_cleaner/internal/scanner"
)

// TestClean_DryRun: dry-run은 실제 파일을 삭제하지 않아야 함
func TestClean_DryRun(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "keep.txt"), 512)

	sr := scanner.ScanResult{
		Item:   scanner.CacheItem{Name: "테스트", Path: dir, Type: scanner.TypeDir},
		Size:   512,
		Exists: true,
	}

	cleaner.Clean([]scanner.ScanResult{sr}, true)

	// 파일이 남아있어야 함
	if _, err := os.Stat(filepath.Join(dir, "keep.txt")); os.IsNotExist(err) {
		t.Error("dry-run은 파일을 삭제하면 안 됩니다")
	}
}

// TestClean_Dir: 실제 삭제 시 디렉토리 내 파일이 제거되어야 함
func TestClean_Dir(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "delete_me.txt")
	writeFile(t, filePath, 256)

	sr := scanner.ScanResult{
		Item:   scanner.CacheItem{Name: "테스트", Path: dir, Type: scanner.TypeDir},
		Size:   256,
		Exists: true,
	}

	results := cleaner.Clean([]scanner.ScanResult{sr}, false)

	if len(results) != 1 {
		t.Fatalf("결과 개수 불일치: 기대 1, 실제 %d", len(results))
	}
	if !results[0].Success {
		t.Errorf("삭제 실패: %v", results[0].Error)
	}

	// 파일이 삭제되어야 함, 디렉토리는 유지
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("파일이 삭제되어야 합니다")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("디렉토리 자체는 유지되어야 합니다")
	}
}

// TestClean_SkipNotExist: Exists=false 항목은 처리하지 않아야 함
func TestClean_SkipNotExist(t *testing.T) {
	sr := scanner.ScanResult{
		Item:   scanner.CacheItem{Name: "없음", Path: "/no/such/path", Type: scanner.TypeDir},
		Exists: false,
	}

	results := cleaner.Clean([]scanner.ScanResult{sr}, false)
	if len(results) != 0 {
		t.Errorf("Exists=false 항목은 결과에 포함되지 않아야 합니다: %d개", len(results))
	}
}

// TestClean_Command: TypeCommand 항목이 정상 실행되어야 함
func TestClean_Command(t *testing.T) {
	sr := scanner.ScanResult{
		Item: scanner.CacheItem{
			Name:    "echo 명령",
			Command: []string{"echo", "pc_cleaner_test"},
			Type:    scanner.TypeCommand,
		},
		Size:   -1,
		Exists: true,
	}

	results := cleaner.Clean([]scanner.ScanResult{sr}, false)
	if len(results) != 1 || !results[0].Success {
		t.Errorf("명령 실행 실패: %v", results[0].Error)
	}
}

// --- helper ---

func writeFile(t *testing.T, path string, size int) {
	t.Helper()
	data := make([]byte, size)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("파일 생성 실패 (%s): %v", path, err)
	}
}

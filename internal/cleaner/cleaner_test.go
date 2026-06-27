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

// TestClean_FreedActualBytes: Freed는 스캔 추정치가 아니라 실제 삭제된 바이트여야 함
func TestClean_FreedActualBytes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.bin"), 700)
	writeFile(t, filepath.Join(dir, "b.bin"), 300)

	sr := scanner.ScanResult{
		Item:   scanner.CacheItem{Name: "테스트", Path: dir, Type: scanner.TypeDir},
		Size:   9999, // 의도적으로 부정확한 스캔값
		Exists: true,
	}

	results := cleaner.Clean([]scanner.ScanResult{sr}, false)
	if results[0].Freed != 1000 {
		t.Errorf("Freed는 실제 삭제량(1000)이어야 합니다: 실제 %d", results[0].Freed)
	}
}

// TestClean_Glob: TypeGlob은 매칭된 디렉토리들의 내용물을 삭제하고 실제 용량을 보고해야 함
func TestClean_Glob(t *testing.T) {
	base := t.TempDir()
	for _, p := range []string{"p1", "p2"} {
		dir := filepath.Join(base, p, "Cache")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "data"), 1024)
	}

	sr := scanner.ScanResult{
		Item:   scanner.CacheItem{Name: "glob", Path: filepath.Join(base, "*", "Cache"), Type: scanner.TypeGlob},
		Size:   2048,
		Exists: true,
	}

	results := cleaner.Clean([]scanner.ScanResult{sr}, false)
	if !results[0].Success {
		t.Errorf("glob 삭제 실패: %v", results[0].Error)
	}
	if results[0].Freed != 2048 {
		t.Errorf("Freed 불일치: 기대 2048, 실제 %d", results[0].Freed)
	}
	// 매칭 디렉토리는 유지되고 내용물만 비워져야 함
	if _, err := os.Stat(filepath.Join(base, "p1", "Cache", "data")); !os.IsNotExist(err) {
		t.Error("캐시 파일이 삭제되어야 합니다")
	}
	if _, err := os.Stat(filepath.Join(base, "p1", "Cache")); os.IsNotExist(err) {
		t.Error("캐시 디렉토리 자체는 유지되어야 합니다")
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

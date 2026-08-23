package scanner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wkqco33/pc_cleaner/internal/scanner"
)

// TestDirSize: 임시 디렉토리를 만들어 용량 계산이 정확한지 검증
func TestScanItem_Dir(t *testing.T) {
	// 준비: 임시 디렉토리 + 파일 생성
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), 1024) // 1 KB
	writeFile(t, filepath.Join(dir, "b.txt"), 2048) // 2 KB
	subDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(subDir, "c.txt"), 512) // 0.5 KB

	item := scanner.CacheItem{
		Name: "테스트 디렉토리",
		Path: dir,
		Type: scanner.TypeDir,
	}

	results := scanner.Scan([]scanner.CacheItem{item})
	if len(results) != 1 {
		t.Fatalf("결과 개수 불일치: 기대 1, 실제 %d", len(results))
	}

	r := results[0]
	if !r.Exists {
		t.Error("경로가 존재해야 합니다")
	}
	if r.Error != nil {
		t.Errorf("예상치 못한 오류: %v", r.Error)
	}

	want := int64(1024 + 2048 + 512)
	if r.Size != want {
		t.Errorf("용량 불일치: 기대 %d, 실제 %d", want, r.Size)
	}
}

// TestScanItem_NotExist: 존재하지 않는 경로는 Exists=false, 에러 없어야 함
func TestScanItem_NotExist(t *testing.T) {
	item := scanner.CacheItem{
		Name: "없는 경로",
		Path: "/tmp/pc_cleaner_no_such_path_xyz",
		Type: scanner.TypeDir,
	}

	results := scanner.Scan([]scanner.CacheItem{item})
	r := results[0]

	if r.Exists {
		t.Error("존재하지 않는 경로는 Exists=false 여야 합니다")
	}
	if r.Error != nil {
		t.Errorf("NotExist는 Error가 nil 이어야 합니다: %v", r.Error)
	}
}

// TestScanItem_Command: TypeCommand 항목은 Exists=true, Size=-1
func TestScanItem_Command(t *testing.T) {
	item := scanner.CacheItem{
		Name:    "명령형 항목",
		Command: []string{"echo", "test"},
		Type:    scanner.TypeCommand,
	}

	results := scanner.Scan([]scanner.CacheItem{item})
	r := results[0]

	if !r.Exists {
		t.Error("TypeCommand는 Exists=true 여야 합니다")
	}
	if r.Size != -1 {
		t.Errorf("TypeCommand Size는 -1 이어야 합니다: 실제 %d", r.Size)
	}
}

// TestScan_Parallel: 여러 항목을 동시에 스캔해도 결과 개수가 맞아야 함
func TestScan_Parallel(t *testing.T) {
	n := 10
	items := make([]scanner.CacheItem, n)
	for i := 0; i < n; i++ {
		items[i] = scanner.CacheItem{
			Name: "없는 경로",
			Path: "/tmp/pc_cleaner_no_such_path_parallel",
			Type: scanner.TypeDir,
		}
	}

	results := scanner.Scan(items)
	if len(results) != n {
		t.Errorf("결과 개수 불일치: 기대 %d, 실제 %d", n, len(results))
	}
}

// TestGetItems: OS별 항목 목록이 비어있지 않아야 함
func TestGetItems_NotEmpty(t *testing.T) {
	items := scanner.GetItems()
	if len(items) == 0 {
		t.Error("GetItems는 최소 1개 이상의 항목을 반환해야 합니다")
	}
}

// TestGetItems_NoNestedDirs: 반환된 디렉토리 항목 중 다른 항목의
// 하위 경로인 것이 없어야 함 (중복 집계/삭제 방지)
func TestGetItems_NoNestedDirs(t *testing.T) {
	items := scanner.GetItems()
	for _, a := range items {
		if a.Type != scanner.TypeDir || a.Path == "" {
			continue
		}
		for _, b := range items {
			if b.Type != scanner.TypeDir || b.Path == "" || a.Path == b.Path {
				continue
			}
			prefix := filepath.Clean(b.Path) + string(filepath.Separator)
			if strings.HasPrefix(filepath.Clean(a.Path), prefix) {
				t.Errorf("중첩 경로가 dedup되지 않음: %q ⊂ %q", a.Path, b.Path)
			}
		}
	}
}

// TestScanItem_Glob: 와일드카드 경로가 매칭된 디렉토리들의 용량을 합산해야 함
func TestScanItem_Glob(t *testing.T) {
	base := t.TempDir()
	// base/p1/Cache, base/p2/Cache 두 프로필 캐시 생성
	for _, p := range []string{"p1", "p2"} {
		dir := filepath.Join(base, p, "Cache")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "data"), 1000)
	}

	item := scanner.CacheItem{
		Name: "glob",
		Path: filepath.Join(base, "*", "Cache"),
		Type: scanner.TypeGlob,
	}
	r := scanner.Scan([]scanner.CacheItem{item})[0]

	if !r.Exists {
		t.Error("매칭이 있으면 Exists=true 여야 합니다")
	}
	if r.Size != 2000 {
		t.Errorf("glob 용량 합산 불일치: 기대 2000, 실제 %d", r.Size)
	}
}

// TestScanItem_GlobNoMatch: 매칭이 없으면 Exists=false
func TestScanItem_GlobNoMatch(t *testing.T) {
	item := scanner.CacheItem{
		Name: "glob",
		Path: filepath.Join(t.TempDir(), "*", "Cache"),
		Type: scanner.TypeGlob,
	}
	r := scanner.Scan([]scanner.CacheItem{item})[0]
	if r.Exists {
		t.Error("매칭이 없으면 Exists=false 여야 합니다")
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

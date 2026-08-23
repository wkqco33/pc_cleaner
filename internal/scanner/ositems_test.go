package scanner

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// homeDirForTest returns the current user's home directory.
func homeDirForTest() (string, error) {
	return os.UserHomeDir()
}

// isAbsDrivePath reports whether path is a Windows-style absolute drive path
// (e.g. C:\Windows\Temp), which legitimately lives outside the home dir.
var absDrivePattern = regexp.MustCompile(`^[A-Za-z]:[\\/]`)

func isAbsDrivePath(path string) bool {
	return absDrivePattern.MatchString(filepath.Clean(path))
}

// 각 OS별 항목 목록 함수는 build tag 없이 모든 플랫폼에서 컴파일되므로
// 실행 환경과 무관하게 불변식(invariant)을 검증할 수 있다.

// verifyItemList applies common invariants to any OS item list.
func verifyItemList(t *testing.T, name string, items []CacheItem) {
	t.Helper()
	if len(items) == 0 {
		t.Errorf("%s: 항목이 비어있으면 안 됩니다", name)
	}
	for i, it := range items {
		label := it.Name
		if label == "" {
			t.Errorf("%s[%d]: Name이 비어있으면 안 됩니다", name, i)
		}
		if it.Category == "" {
			t.Errorf("%s[%d](%q): Category가 비어있으면 안 됩니다", name, i, label)
		}
		switch it.Type {
		case TypeDir, TypeGlob:
			if it.Path == "" {
				t.Errorf("%s[%d](%q): Dir/Glob은 Path가 필요합니다", name, i, label)
			}
		case TypeCommand:
			if len(it.Command) == 0 {
				t.Errorf("%s[%d](%q): Command 항목은 명령이 필요합니다", name, i, label)
			}
		default:
			t.Errorf("%s[%d](%q): 알 수 없는 Type %d", name, i, label, it.Type)
		}
	}
}

func TestCommonItems_Invariants(t *testing.T) {
	verifyItemList(t, "common", commonItems())
}

func TestDarwinItems_Invariants(t *testing.T) {
	verifyItemList(t, "darwin", darwinItems())
}

func TestLinuxItems_Invariants(t *testing.T) {
	verifyItemList(t, "linux", linuxItems())
}

func TestWindowsItems_Invariants(t *testing.T) {
	verifyItemList(t, "windows", windowsItems())
}

// 각 OS 목록은 홈 디렉토리를 기준으로 한 하위 경로만 사용해야 한다
// (절대 고정 경로가 어긋나면 안 됨).
func TestOSItems_PathsDerivedFromHome(t *testing.T) {
	home, err := homeDirForTest()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		items []CacheItem
	}{
		{"common", commonItems()},
		{"darwin", darwinItems()},
		{"linux", linuxItems()},
		{"windows", windowsItems()},
	} {
		for _, it := range tc.items {
			if it.Type == TypeCommand {
				continue
			}
			clean := filepath.Clean(it.Path)
			if !strings.HasPrefix(clean, home) && !isAbsDrivePath(clean) {
				t.Errorf("%s(%q): 경로 %q 가 홈(%q) 하위가 아닙니다", tc.name, it.Name, it.Path, home)
			}
		}
	}
}

// 잘 알려진 핵심 항목이 각 OS 목록에 존재하는지 확인.
func TestDarwinItems_HasExpected(t *testing.T) {
	items := darwinItems()
	assertHasItem(t, items, "Trash")
	assertHasItem(t, items, "Xcode DerivedData")
}

func TestLinuxItems_HasExpected(t *testing.T) {
	items := linuxItems()
	assertHasItem(t, items, "Thumbnail 캐시")
}

func TestWindowsItems_HasExpected(t *testing.T) {
	items := windowsItems()
	// 관리자 권한 필요 항목이 명시되어야 함
	admin := map[string]bool{}
	for _, it := range items {
		if it.RequiresAdmin {
			admin[it.Name] = true
		}
	}
	for _, want := range []string{"Windows TEMP", "Windows Prefetch", "Windows Update 캐시"} {
		if !admin[want] {
			t.Errorf("%q 항목은 RequiresAdmin=true 여야 합니다", want)
		}
	}
	// Windows만의 시스템 경로가 존재
	assertHasItem(t, items, "Edge 캐시")
}

func assertHasItem(t *testing.T, items []CacheItem, want string) {
	t.Helper()
	for _, it := range items {
		if it.Name == want {
			return
		}
	}
	t.Errorf("기대 항목 %q 이 목록에 없습니다", want)
}

// Package scanner provides cache path discovery and size calculation.
package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ItemType classifies a cache item's cleanup method.
type ItemType int

const (
	TypeDir     ItemType = iota // 디렉토리 내용물 삭제
	TypeCommand                 // 외부 명령 실행
	TypeGlob                    // 와일드카드 경로(여러 프로필 등) 일괄 처리
)

// CacheItem represents a single cleanable cache target.
type CacheItem struct {
	Name          string
	Path          string   // TypeDir/TypeGlob일 때 사용 (TypeGlob은 와일드카드 포함)
	Command       []string // TypeCommand일 때 사용
	Type          ItemType
	Category      string
	RequiresAdmin bool // 관리자/root 권한이 있어야 삭제 가능한 경로
}

// ScanResult holds scan output for a single CacheItem.
type ScanResult struct {
	Item       CacheItem
	Size       int64
	Exists     bool
	Error      error
	NeedsAdmin bool // 권한 부족으로 정리에서 제외해야 함
}

// GetItems returns all cache items applicable to the current OS.
func GetItems() []CacheItem {
	items := commonItems()
	switch runtime.GOOS {
	case "darwin":
		items = append(items, darwinItems()...)
	case "windows":
		items = append(items, windowsItems()...)
	case "linux":
		items = append(items, linuxItems()...)
	}
	return dedupNested(items)
}

// dedupNested drops directory items whose path is nested under another
// listed directory item (e.g. ~/Library/Caches/org.swift.swiftpm under
// ~/Library/Caches) to avoid double-counting and double-deletion.
func dedupNested(items []CacheItem) []CacheItem {
	out := make([]CacheItem, 0, len(items))
	for i, item := range items {
		if item.Type == TypeDir && item.Path != "" {
			nested := false
			for j, other := range items {
				if i == j || other.Type != TypeDir || other.Path == "" {
					continue
				}
				if isAncestor(other.Path, item.Path) {
					nested = true
					break
				}
			}
			if nested {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}

// isAncestor reports whether parent is a strict ancestor directory of child.
func isAncestor(parent, child string) bool {
	p := filepath.Clean(parent)
	c := filepath.Clean(child)
	if p == c {
		return false
	}
	return strings.HasPrefix(c, p+string(os.PathSeparator))
}

// Scan calculates sizes for all items concurrently.
func Scan(items []CacheItem) []ScanResult {
	results := make([]ScanResult, len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(idx int, ci CacheItem) {
			defer wg.Done()
			results[idx] = scanItem(ci)
		}(i, item)
	}
	wg.Wait()
	return results
}

func scanItem(item CacheItem) ScanResult {
	res := ScanResult{Item: item}

	if item.RequiresAdmin && !isElevated() {
		res.NeedsAdmin = true
	}

	switch item.Type {
	case TypeCommand:
		// 명령형 항목은 존재 여부만 표시, 용량은 알 수 없음 (-1)
		res.Exists = true
		res.Size = -1
		return res

	case TypeGlob:
		matches, _ := filepath.Glob(item.Path)
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil {
				continue
			}
			res.Exists = true
			if info.IsDir() {
				res.Size += dirSize(m)
			} else {
				res.Size += info.Size()
			}
		}
		return res

	default: // TypeDir
		info, err := os.Stat(item.Path)
		if err != nil {
			if os.IsNotExist(err) {
				res.Exists = false
				return res
			}
			res.Error = err
			return res
		}
		res.Exists = true
		if info.IsDir() {
			res.Size = dirSize(item.Path)
		} else {
			res.Size = info.Size()
		}
		return res
	}
}

func dirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // 접근 불가 항목은 skip
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

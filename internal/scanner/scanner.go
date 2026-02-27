// Package scanner provides cache path discovery and size calculation.
package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// ItemType classifies a cache item's cleanup method.
type ItemType int

const (
	TypeDir     ItemType = iota // 디렉토리 전체 삭제
	TypeCommand                 // 외부 명령 실행
)

// CacheItem represents a single cleanable cache target.
type CacheItem struct {
	Name     string
	Path     string   // TypeDir일 때 사용
	Command  []string // TypeCommand일 때 사용
	Type     ItemType
	Category string
}

// ScanResult holds scan output for a single CacheItem.
type ScanResult struct {
	Item   CacheItem
	Size   int64
	Exists bool
	Error  error
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
	return items
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

	if item.Type == TypeCommand {
		// 명령형 항목은 존재 여부만 표시, 용량은 알 수 없음 (-1)
		res.Exists = true
		res.Size = -1
		return res
	}

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

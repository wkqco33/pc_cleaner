package scanner

import (
	"os"
	"path/filepath"
)

func darwinItems() []CacheItem {
	home, _ := os.UserHomeDir()

	return []CacheItem{
		{
			Name:     "시스템 사용자 캐시",
			Path:     filepath.Join(home, "Library", "Caches"),
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "앱 로그",
			Path:     filepath.Join(home, "Library", "Logs"),
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "Xcode DerivedData",
			Path:     filepath.Join(home, "Library", "Developer", "Xcode", "DerivedData"),
			Type:     TypeDir,
			Category: "개발 도구",
		},
		{
			Name:     "iOS 시뮬레이터 캐시",
			Path:     filepath.Join(home, "Library", "Developer", "CoreSimulator", "Caches"),
			Type:     TypeDir,
			Category: "개발 도구",
		},
		{
			Name:     "Swift Package 캐시",
			Path:     filepath.Join(home, "Library", "Caches", "org.swift.swiftpm"),
			Type:     TypeDir,
			Category: "개발 도구",
		},
		{
			Name:     "Trash",
			Path:     filepath.Join(home, ".Trash"),
			Type:     TypeDir,
			Category: "시스템",
		},
	}
}

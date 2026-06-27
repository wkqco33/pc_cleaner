package scanner

import (
	"os"
	"path/filepath"
)

func linuxItems() []CacheItem {
	home, _ := os.UserHomeDir()
	cache := filepath.Join(home, ".cache")

	return []CacheItem{
		{
			Name:     "Thumbnail 캐시",
			Path:     filepath.Join(cache, "thumbnails"),
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "Journald 로그 정리",
			Command:  []string{"journalctl", "--vacuum-size=100M"},
			Type:     TypeCommand,
			Category: "시스템",
		},
		// --- 브라우저 (프로필별 캐시) ---
		{
			Name:     "Chrome 캐시",
			Path:     filepath.Join(cache, "google-chrome", "*", "Cache"),
			Type:     TypeGlob,
			Category: "브라우저",
		},
		{
			Name:     "Chromium 캐시",
			Path:     filepath.Join(cache, "chromium", "*", "Cache"),
			Type:     TypeGlob,
			Category: "브라우저",
		},
		{
			Name:     "Firefox 캐시",
			Path:     filepath.Join(cache, "mozilla", "firefox", "*", "cache2"),
			Type:     TypeGlob,
			Category: "브라우저",
		},
	}
}

package scanner

import (
	"os"
	"path/filepath"
)

func linuxItems() []CacheItem {
	home, _ := os.UserHomeDir()

	return []CacheItem{
		{
			Name:     "Thumbnail 캐시",
			Path:     filepath.Join(home, ".cache", "thumbnails"),
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "사용자 캐시 (~/.cache)",
			Path:     filepath.Join(home, ".cache"),
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "Journald 로그 정리",
			Command:  []string{"journalctl", "--vacuum-size=100M"},
			Type:     TypeCommand,
			Category: "시스템",
		},
		{
			Name:     "/tmp 임시 파일",
			Path:     "/tmp",
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "/var/tmp 임시 파일",
			Path:     "/var/tmp",
			Type:     TypeDir,
			Category: "시스템",
		},
	}
}

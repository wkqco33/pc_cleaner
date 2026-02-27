package scanner

import (
	"os"
	"path/filepath"
)

func windowsItems() []CacheItem {
	home, _ := os.UserHomeDir()
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(home, "AppData", "Local")
	}
	temp := os.Getenv("TEMP")
	if temp == "" {
		temp = filepath.Join(localAppData, "Temp")
	}

	return []CacheItem{
		{
			Name:     "사용자 TEMP",
			Path:     temp,
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "Windows TEMP",
			Path:     `C:\Windows\Temp`,
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "Windows Prefetch",
			Path:     `C:\Windows\Prefetch`,
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "IE/Edge 캐시",
			Path:     filepath.Join(localAppData, "Microsoft", "Windows", "INetCache"),
			Type:     TypeDir,
			Category: "브라우저",
		},
		{
			Name:     "Windows Thumbnail 캐시",
			Path:     filepath.Join(localAppData, "Microsoft", "Windows", "Explorer"),
			Type:     TypeDir,
			Category: "시스템",
		},
		{
			Name:     "Windows Update 캐시",
			Path:     `C:\Windows\SoftwareDistribution\Download`,
			Type:     TypeDir,
			Category: "시스템",
		},
	}
}

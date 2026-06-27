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
			Name:          "Windows TEMP",
			Path:          `C:\Windows\Temp`,
			Type:          TypeDir,
			Category:      "시스템",
			RequiresAdmin: true,
		},
		{
			Name:          "Windows Prefetch",
			Path:          `C:\Windows\Prefetch`,
			Type:          TypeDir,
			Category:      "시스템",
			RequiresAdmin: true,
		},
		{
			Name:          "Windows Update 캐시",
			Path:          `C:\Windows\SoftwareDistribution\Download`,
			Type:          TypeDir,
			Category:      "시스템",
			RequiresAdmin: true,
		},
		{
			Name:     "IE/레거시 WinINet 캐시",
			Path:     filepath.Join(localAppData, "Microsoft", "Windows", "INetCache"),
			Type:     TypeDir,
			Category: "브라우저",
		},
		{
			// Explorer 폴더 전체가 아닌 썸네일/아이콘 캐시 DB만 대상.
			// 실행 중 explorer.exe가 잠근 파일은 부분 실패로 처리됨.
			Name:     "Windows Thumbnail/Icon 캐시",
			Path:     filepath.Join(localAppData, "Microsoft", "Windows", "Explorer", "*cache_*.db"),
			Type:     TypeGlob,
			Category: "시스템",
		},
		// --- 브라우저 (프로필별 캐시) ---
		{
			Name:     "Edge 캐시",
			Path:     filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "*", "Cache"),
			Type:     TypeGlob,
			Category: "브라우저",
		},
		{
			Name:     "Chrome 캐시",
			Path:     filepath.Join(localAppData, "Google", "Chrome", "User Data", "*", "Cache"),
			Type:     TypeGlob,
			Category: "브라우저",
		},
		{
			Name:     "Firefox 캐시",
			Path:     filepath.Join(localAppData, "Mozilla", "Firefox", "Profiles", "*", "cache2"),
			Type:     TypeGlob,
			Category: "브라우저",
		},
	}
}

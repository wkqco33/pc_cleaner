package scanner

import (
	"os"
	"path/filepath"
)

// commonItems returns developer tool cache paths applicable to all OSes.
func commonItems() []CacheItem {
	home, _ := os.UserHomeDir()

	return []CacheItem{
		// --- 빌드 도구 ---
		{
			Name:     "Gradle 캐시",
			Path:     filepath.Join(home, ".gradle", "caches"),
			Type:     TypeDir,
			Category: "빌드 도구",
		},
		{
			Name:     "Maven 로컬 저장소",
			Path:     filepath.Join(home, ".m2", "repository"),
			Type:     TypeDir,
			Category: "빌드 도구",
		},
		// --- Python ---
		{
			Name:     "pip 캐시",
			Path:     filepath.Join(home, ".cache", "pip"),
			Type:     TypeDir,
			Category: "Python",
		},
		{
			Name:     "uv 캐시",
			Path:     filepath.Join(home, ".cache", "uv"),
			Type:     TypeDir,
			Category: "Python",
		},
		// --- Node.js ---
		{
			Name:     "npm 캐시",
			Path:     filepath.Join(home, ".npm", "_cacache"),
			Type:     TypeDir,
			Category: "Node.js",
		},
		{
			Name:     "yarn 캐시",
			Path:     filepath.Join(home, ".yarn", "cache"),
			Type:     TypeDir,
			Category: "Node.js",
		},
		// --- Rust ---
		{
			Name:     "Cargo 레지스트리 캐시",
			Path:     filepath.Join(home, ".cargo", "registry", "cache"),
			Type:     TypeDir,
			Category: "Rust",
		},
		{
			Name:     "Cargo git 캐시",
			Path:     filepath.Join(home, ".cargo", "git", "db"),
			Type:     TypeDir,
			Category: "Rust",
		},
		// --- Go ---
		{
			Name:     "Go 모듈 캐시",
			Path:     filepath.Join(home, "go", "pkg", "mod", "cache"),
			Type:     TypeDir,
			Category: "Go",
		},
		// --- Docker ---
		{
			Name:     "Docker 미사용 리소스",
			Command:  []string{"docker", "system", "prune", "-f"},
			Type:     TypeCommand,
			Category: "Docker",
		},
	}
}

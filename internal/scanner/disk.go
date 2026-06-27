package scanner

import (
	"os"
)

// DiskInfo holds disk space information in bytes.
type DiskInfo struct {
	Total     uint64
	Free      uint64
	Available uint64
}

// GetDiskUsage retrieves disk space information for the partition containing the given path.
// If the path is empty, it queries the partition containing the user's home directory.
func GetDiskUsage(path string) (DiskInfo, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = home
		} else {
			path = "."
		}
	}
	return getPlatformDiskUsage(path)
}

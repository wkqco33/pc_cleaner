//go:build !windows

package scanner

import (
	"syscall"
)

func getPlatformDiskUsage(path string) (DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskInfo{}, err
	}

	return DiskInfo{
		Total:     uint64(stat.Blocks) * uint64(stat.Bsize),
		Free:      uint64(stat.Bfree) * uint64(stat.Bsize),
		Available: uint64(stat.Bavail) * uint64(stat.Bsize),
	}, nil
}

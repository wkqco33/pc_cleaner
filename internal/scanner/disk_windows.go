//go:build windows

package scanner

import (
	"syscall"
	"unsafe"
)

var (
	modkernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetDiskFreeSpaceExW = modkernel32.NewProc("GetDiskFreeSpaceExW")
)

func getPlatformDiskUsage(path string) (DiskInfo, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return DiskInfo{}, err
	}

	var availBytes, totalBytes, freeBytes uint64
	r1, _, err := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&availBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&freeBytes)),
	)
	if r1 == 0 {
		return DiskInfo{}, err
	}

	return DiskInfo{
		Total:     totalBytes,
		Free:      freeBytes,
		Available: availBytes,
	}, nil
}

package scanner

import "os"

// isElevated reports whether the process runs with Administrator rights.
// Opening a raw physical drive handle requires elevation, so a successful
// open is a reliable, dependency-free elevation probe on Windows.
func isElevated() bool {
	f, err := os.Open(`\\.\PHYSICALDRIVE0`)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

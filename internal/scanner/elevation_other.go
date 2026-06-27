//go:build !windows

package scanner

// isElevated is only meaningful on Windows; on other OSes no cache item is
// marked RequiresAdmin, so this is never gated on.
func isElevated() bool { return true }

// Package ui provides lightweight display helpers for the CLI.
// Terminal color/rendering is handled by the wcli/rich package; this package
// only keeps the byte-size formatter that rich does not provide.
package ui

import "fmt"

// FormatBytes converts byte count to human-readable string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

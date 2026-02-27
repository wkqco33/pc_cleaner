// Package ui provides terminal output utilities with color support.
package ui

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorGray   = "\033[90m"
)

var colorEnabled bool

func init() {
	// Windows는 기본적으로 ANSI 미지원, TERM 확인
	if runtime.GOOS == "windows" {
		colorEnabled = os.Getenv("TERM") != "" || os.Getenv("WT_SESSION") != ""
	} else {
		colorEnabled = true
	}
}

func colorize(code, text string) string {
	if !colorEnabled {
		return text
	}
	return code + text + colorReset
}

func Red(s string) string    { return colorize(colorRed, s) }
func Green(s string) string  { return colorize(colorGreen, s) }
func Yellow(s string) string { return colorize(colorYellow, s) }
func Cyan(s string) string   { return colorize(colorCyan, s) }
func Bold(s string) string   { return colorize(colorBold, s) }
func Gray(s string) string   { return colorize(colorGray, s) }

func PrintHeader(title string) {
	line := strings.Repeat("─", 60)
	fmt.Println()
	fmt.Println(Cyan(line))
	fmt.Printf("  %s\n", Bold(title))
	fmt.Println(Cyan(line))
}

func PrintInfo(msg string) {
	fmt.Printf("  %s %s\n", Cyan("ℹ"), msg)
}

func PrintOK(msg string) {
	fmt.Printf("  %s %s\n", Green("✓"), msg)
}

func PrintWarn(msg string) {
	fmt.Printf("  %s %s\n", Yellow("⚠"), msg)
}

func PrintError(msg string) {
	fmt.Printf("  %s %s\n", Red("✗"), msg)
}

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

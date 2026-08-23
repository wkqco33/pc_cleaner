package scanner_test

import (
	"os"
	"testing"

	"github.com/wkqco33/pc_cleaner/internal/scanner"
)

func TestGetDiskUsage(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("User home dir not available")
	}

	info, err := scanner.GetDiskUsage(home)
	if err != nil {
		t.Fatalf("GetDiskUsage failed: %v", err)
	}

	if info.Total == 0 {
		t.Error("Expected Total disk space to be > 0")
	}

	if info.Available > info.Total {
		t.Errorf("Expected Available disk space (%d) to be <= Total (%d)", info.Available, info.Total)
	}
}

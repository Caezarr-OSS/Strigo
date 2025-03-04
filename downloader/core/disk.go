package core

import (
	"fmt"
	"strigo/logging"

	"golang.org/x/sys/unix"
)

// GetAvailableDiskSpace returns available disk space in bytes for a given path
func GetAvailableDiskSpace(path string) (uint64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("failed to check disk space: %w", err)
	}

	// Calculate available space in bytes
	return stat.Bavail * uint64(stat.Bsize), nil
}

// CheckDiskSpace verifies if there is enough available disk space
func CheckDiskSpace(requiredBytes int64, path string) error {
	available, err := GetAvailableDiskSpace(path)
	if err != nil {
		return err
	}

	required := uint64(requiredBytes * 2) // 2x for extraction buffer

	if available < required {
		return fmt.Errorf("need %d bytes, only %d bytes available", required, available)
	}

	logging.LogDebug("💾 Disk space check passed: %d bytes available, %d bytes required", available, required)
	return nil
}

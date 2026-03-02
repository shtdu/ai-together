// Copyright (c) 2025 Code Together
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


//go:build darwin || linux || windows
// +build darwin linux windows

package utils

import (
	"fmt"
	"syscall"
)

const minDiskSpace = 5 * 1024 * 1024 // 5MB minimum free space required

// CheckDiskSpace verifies sufficient disk space before write operations.
// Returns nil if space check fails (don't block on unsupported platforms),
// but logs a warning. Only returns an error if space is definitely insufficient.
func CheckDiskSpace(dbPath string) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dbPath, &stat); err != nil {
		// Can't check disk space on this platform/filesystem
		// Don't fail - proceed with caution
		return nil
	}

	// Calculate available space in bytes
	available := stat.Bavail * uint64(stat.Bsize)

	if available < minDiskSpace {
		return fmt.Errorf("insufficient disk space: %d bytes available "+
			"(minimum required: %d bytes).\n"+
			"Actions to fix:\n"+
			"  1. Free up disk space on this filesystem\n"+
			"  2. Move database to another location using -db-path flag\n"+
			"  3. Run cleanup: claude-hook-collector clean --older-than 30d --compact",
			available, minDiskSpace)
	}

	return nil
}

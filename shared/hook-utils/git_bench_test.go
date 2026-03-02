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


package hookutils

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkGetGitMeta benchmarks the complete GetGitMeta function
// which fetches remote URLs, branch name, and commit hash.
func BenchmarkGetGitMeta(b *testing.B) {
	// Use the current repository as test subject
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	// Find the git repo root by going up directories
	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetGitMeta(testDir)
		if err != nil {
			b.Fatalf("GetGitMeta failed: %v", err)
		}
	}
}

// BenchmarkGetGitMetaFast benchmarks the parallel implementation.
// Expected: ~3x faster than BenchmarkGetGitMeta (target: <20ms).
func BenchmarkGetGitMetaFast(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetGitMetaFast(testDir)
		if err != nil {
			b.Fatalf("GetGitMetaFast failed: %v", err)
		}
	}
}

// BenchmarkGetGitMetaParallel benchmarks GetGitMeta with parallel execution
// to simulate concurrent access patterns.
func BenchmarkGetGitMetaParallel(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := GetGitMeta(testDir)
			if err != nil {
				b.Fatalf("GetGitMeta failed: %v", err)
			}
		}
	})
}

// BenchmarkGetGitMetaFastParallel benchmarks the parallel implementation
// with concurrent access to ensure thread safety.
func BenchmarkGetGitMetaFastParallel(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := GetGitMetaFast(testDir)
			if err != nil {
				b.Fatalf("GetGitMetaFast failed: %v", err)
			}
		}
	})
}

// BenchmarkIsGitRepo benchmarks the git repository check.
func BenchmarkIsGitRepo(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsGitRepo(testDir)
	}
}

// BenchmarkGetBranch benchmarks the branch name fetching.
func BenchmarkGetBranch(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetBranch(testDir)
		if err != nil {
			b.Fatalf("GetBranch failed: %v", err)
		}
	}
}

// BenchmarkGetCommitHash benchmarks the short commit hash fetching.
func BenchmarkGetCommitHash(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCommitHash(testDir, false)
		if err != nil {
			b.Fatalf("GetCommitHash failed: %v", err)
		}
	}
}

// BenchmarkGetCommitHashFull benchmarks the full 40-character commit hash fetching.
func BenchmarkGetCommitHashFull(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCommitHash(testDir, true)
		if err != nil {
			b.Fatalf("GetCommitHash failed: %v", err)
		}
	}
}

// BenchmarkGetUniqueRemoteURLs benchmarks the remote URL fetching with deduplication.
func BenchmarkGetUniqueRemoteURLs(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetUniqueRemoteURLs(testDir)
		if err != nil {
			b.Fatalf("GetUniqueRemoteURLs failed: %v", err)
		}
	}
}

// BenchmarkGetGitRemotes benchmarks the raw remote fetching (includes duplicates).
func BenchmarkGetGitRemotes(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetGitRemotes(testDir)
		if err != nil {
			b.Fatalf("GetGitRemotes failed: %v", err)
		}
	}
}

// BenchmarkParseGitRemoteOutput benchmarks the parsing function only.
// This tests the string parsing overhead without subprocess execution.
func BenchmarkParseGitRemoteOutput(b *testing.B) {
	const sampleGitRemoteOutput = `origin	git@github.com:username/repository.git (fetch)
origin	git@github.com:username/repository.git (push)
upstream	git@github.com:other/repo.git (fetch)
upstream	git@github.com:other/repo.git (push)`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseGitRemoteOutput(sampleGitRemoteOutput)
	}
}

// BenchmarkGetGitMetaIndividualOperations benchmarks each operation separately
// to identify performance bottlenecks in the GetGitMeta workflow.
func BenchmarkGetGitMetaIndividualOperations(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	b.Run("IsGitRepo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsGitRepo(testDir)
		}
	})

	b.Run("GetUniqueRemoteURLs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GetUniqueRemoteURLs(testDir)
		}
	})

	b.Run("GetBranch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GetBranch(testDir)
		}
	})

	b.Run("GetCommitHash_Short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GetCommitHash(testDir, false)
		}
	})

	b.Run("GetCommitHash_Full", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GetCommitHash(testDir, true)
		}
	})
}

// findGitRepoRoot traverses up directories to find the git repository root.
// Returns the current directory if no git root is found.
func findGitRepoRoot(startDir string) string {
	dir := startDir
	for {
		// Check if .git exists in this directory
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, return startDir
			return startDir
		}
		dir = parent
	}
}

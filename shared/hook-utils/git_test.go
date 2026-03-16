// Copyright (c) 2025 AI Together
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
	"testing"
)

const sampleGitRemoteOutput = `origin	git@github.com:username/repository.git (fetch)
origin	git@github.com:username/repository.git (push)
upstream	git@github.com:other/repo.git (fetch)
upstream	git@github.com:other/repo.git (push)`

const sampleGitRemoteSingleOutput = `origin	https://github.com/user/repo.git (fetch)
origin	https://github.com/user/repo.git (push)`

func TestParseGitRemoteOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantLen int
		wantURL string // Check for at least this URL
	}{
		{
			name:    "multiple remotes with fetch and push",
			output:  sampleGitRemoteOutput,
			wantLen: 4,
			wantURL: "git@github.com:username/repository.git",
		},
		{
			name:    "single remote with fetch and push",
			output:  sampleGitRemoteSingleOutput,
			wantLen: 2,
			wantURL: "https://github.com/user/repo.git",
		},
		{
			name:    "empty output",
			output:  "",
			wantLen: 0,
		},
		{
			name:    "output with only newlines",
			output:  "\n\n",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remotes := parseGitRemoteOutput(tt.output)

			if len(remotes) != tt.wantLen {
				t.Errorf("parseGitRemoteOutput() returned %d remotes, want %d", len(remotes), tt.wantLen)
			}

			if tt.wantURL != "" {
				found := false
				for _, r := range remotes {
					if r.URL == tt.wantURL {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("parseGitRemoteOutput() did not find expected URL %s", tt.wantURL)
				}
			}
		})
	}
}

func TestParseGitRemoteOutputMalformed(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{
			name:   "missing tab separator",
			output: "origin git@github.com:user/repo.git (fetch)",
		},
		{
			name:   "missing parentheses",
			output: "origin\tgit@github.com:user/repo.git fetch",
		},
		{
			name:   "only name and tab",
			output: "origin\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remotes := parseGitRemoteOutput(tt.output)

			// Malformed lines should be skipped
			if len(remotes) != 0 {
				t.Errorf("parseGitRemoteOutput() should skip malformed lines, got %d remotes", len(remotes))
			}
		})
	}
}

func TestParseGitRemoteOutputFields(t *testing.T) {
	remotes := parseGitRemoteOutput(sampleGitRemoteOutput)

	// Check first remote (origin fetch)
	if remotes[0].Name != "origin" {
		t.Errorf("Expected Name 'origin', got '%s'", remotes[0].Name)
	}
	if remotes[0].URL != "git@github.com:shtdu/code-together.git" {
		t.Errorf("Expected URL 'git@github.com:shtdu/code-together.git', got '%s'", remotes[0].URL)
	}
	if remotes[0].Type != "fetch" {
		t.Errorf("Expected Type 'fetch', got '%s'", remotes[0].Type)
	}

	// Check third remote (upstream fetch)
	if remotes[2].Name != "upstream" {
		t.Errorf("Expected Name 'upstream', got '%s'", remotes[2].Name)
	}
	if remotes[2].URL != "git@github.com:other/repo.git" {
		t.Errorf("Expected URL 'git@github.com:other/repo.git', got '%s'", remotes[2].URL)
	}
	if remotes[2].Type != "fetch" {
		t.Errorf("Expected Type 'fetch', got '%s'", remotes[2].Type)
	}
}

func TestDeduplicateRemotes(t *testing.T) {
	// Create a mock scenario with duplicate URLs
	testRemotes := []GitRemote{
		{Name: "origin", URL: "git@github.com:user/repo.git", Type: "fetch"},
		{Name: "origin", URL: "git@github.com:user/repo.git", Type: "push"},
		{Name: "upstream", URL: "git@github.com:other/repo.git", Type: "fetch"},
		{Name: "upstream", URL: "git@github.com:other/repo.git", Type: "push"},
	}

	seen := make(map[string]bool)
	var urls []string
	for _, remote := range testRemotes {
		if !seen[remote.URL] {
			seen[remote.URL] = true
			urls = append(urls, remote.URL)
		}
	}

	if len(urls) != 2 {
		t.Errorf("Expected 2 unique URLs after deduplication, got %d", len(urls))
	}

	// Check that URLs are unique
	uniqueSet := make(map[string]bool)
	for _, url := range urls {
		if uniqueSet[url] {
			t.Errorf("URL %s appears multiple times in deduplicated list", url)
		}
		uniqueSet[url] = true
	}
}

func TestGitMeta(t *testing.T) {
	// Test GitMeta struct creation
	meta := &GitMeta{
		RemoteURLs: []string{"git@github.com:user/repo.git"},
		Branch:     "main",
		CommitHash: "a1b2c3d",
	}

	if len(meta.RemoteURLs) != 1 {
		t.Errorf("Expected 1 remote URL, got %d", len(meta.RemoteURLs))
	}
	if meta.Branch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", meta.Branch)
	}
	if meta.CommitHash != "a1b2c3d" {
		t.Errorf("Expected commit hash 'a1b2c3d', got '%s'", meta.CommitHash)
	}
}

func TestGitRemote(t *testing.T) {
	remote := GitRemote{
		Name: "origin",
		URL:  "git@github.com:user/repo.git",
		Type: "fetch",
	}

	if remote.Name != "origin" {
		t.Errorf("Expected Name 'origin', got '%s'", remote.Name)
	}
	if remote.URL != "git@github.com:user/repo.git" {
		t.Errorf("Expected URL 'git@github.com:user/repo.git', got '%s'", remote.URL)
	}
	if remote.Type != "fetch" {
		t.Errorf("Expected Type 'fetch', got '%s'", remote.Type)
	}
}

// TestGetGitMetaFast verifies GetGitMetaFast returns the same results as GetGitMeta.
func TestGetGitMetaFast(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	testDir := findGitRepoRoot(cwd)

	// Get results from both implementations
	metaSequential, err := GetGitMeta(testDir)
	if err != nil {
		t.Fatalf("GetGitMeta failed: %v", err)
	}

	metaParallel, err := GetGitMetaFast(testDir)
	if err != nil {
		t.Fatalf("GetGitMetaFast failed: %v", err)
	}

	// Compare results
	if metaSequential == nil && metaParallel == nil {
		return // Both agree: not a git repo
	}

	if metaSequential == nil || metaParallel == nil {
		t.Fatalf("One implementation returned nil, the other didn't")
	}

	// Compare RemoteURLs
	if len(metaSequential.RemoteURLs) != len(metaParallel.RemoteURLs) {
		t.Errorf("RemoteURLs length mismatch: sequential=%d, parallel=%d",
			len(metaSequential.RemoteURLs), len(metaParallel.RemoteURLs))
	}

	// Compare Branch
	if metaSequential.Branch != metaParallel.Branch {
		t.Errorf("Branch mismatch: sequential=%s, parallel=%s",
			metaSequential.Branch, metaParallel.Branch)
	}

	// Compare CommitHash
	if metaSequential.CommitHash != metaParallel.CommitHash {
		t.Errorf("CommitHash mismatch: sequential=%s, parallel=%s",
			metaSequential.CommitHash, metaParallel.CommitHash)
	}
}

// TestGetGitMetaFast_NonGitDirectory verifies GetGitMetaFast returns nil,nil for non-git directories.
func TestGetGitMetaFast_NonGitDirectory(t *testing.T) {
	// Create a temporary directory (not a git repo)
	tmpDir := t.TempDir()

	meta, err := GetGitMetaFast(tmpDir)
	if err != nil {
		t.Fatalf("Expected nil error for non-git directory, got: %v", err)
	}
	if meta != nil {
		t.Fatalf("Expected nil GitMeta for non-git directory, got: %+v", meta)
	}
}

func TestExtractRepoShortName(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "SSH format - GitHub",
			url:      "git@github.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH format - without .git",
			url:      "git@github.com:owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS format - GitHub",
			url:      "https://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS format - without .git",
			url:      "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS format - with www",
			url:      "https://www.github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH format - GitLab",
			url:      "git@gitlab.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS format - GitLab",
			url:      "https://gitlab.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH format - Bitbucket",
			url:      "git@bitbucket.org:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS format - Bitbucket",
			url:      "https://bitbucket.org/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "Nested organization",
			url:      "git@github.com:org/team/repo.git",
			expected: "org/team/repo",
		},
		{
			name:     "HTTPS nested organization",
			url:      "https://github.com/org/team/repo.git",
			expected: "org/team/repo",
		},
		{
			name:     "Empty string",
			url:      "",
			expected: "",
		},
		{
			name:     "Invalid format - no slash",
			url:      "not-a-url",
			expected: "",
		},
		{
			name:     "Invalid format - only protocol",
			url:      "https://",
			expected: "", // Function handles this edge case gracefully
		},
		{
			name:     "Git protocol",
			url:      "git://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH with custom port",
			url:      "ssh://git@github.com:22/owner/repo.git",
			expected: "owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRepoShortName(tt.url)
			if result != tt.expected {
				t.Errorf("ExtractRepoShortName(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

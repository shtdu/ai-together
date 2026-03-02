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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// ErrNotGitRepo is returned when the directory is not a git repository.
// Callers can check for this error using errors.Is(err, ErrNotGitRepo).
var ErrNotGitRepo = errors.New("not a git repository")

// GitMeta contains all git repository metadata
type GitMeta struct {
	RemoteURLs []string // Unique remote URLs (deduplicated)
	Branch     string   // Current branch name (e.g., "main", "feature/xyz")
	CommitHash string   // Short commit hash (e.g., "a1b2c3d")
}

// GitRemote represents a single git remote entry (used internally)
type GitRemote struct {
	Name string // e.g., "origin"
	URL  string // e.g., "git@github.com:shtdu/code-together.git"
	Type string // "fetch" or "push"
}

// GetGitMeta returns all git metadata for the given directory.
// Returns ErrNotGitRepo if the directory is not inside a git repository.
// Returns an error if git commands fail. Partial results may be returned with an error.
func GetGitMeta(cwd string) (*GitMeta, error) {
	if !IsGitRepo(cwd) {
		return nil, ErrNotGitRepo
	}

	// Collect all errors
	var errs []error

	// Get each piece of metadata, capturing errors
	urls, err := GetUniqueRemoteURLs(cwd)
	if err != nil {
		errs = append(errs, fmt.Errorf("get remote URLs: %w", err))
	}

	branch, err := GetBranch(cwd)
	if err != nil {
		errs = append(errs, fmt.Errorf("get branch: %w", err))
	}

	hash, err := GetCommitHash(cwd, false)
	if err != nil {
		errs = append(errs, fmt.Errorf("get commit hash: %w", err))
	}

	// Return partial results if we have at least some data
	meta := &GitMeta{
		RemoteURLs: urls,
		Branch:     branch,
		CommitHash: hash,
	}

	// Return error if any operations failed, but still return partial metadata
	if len(errs) > 0 {
		return meta, errors.Join(errs...)
	}

	return meta, nil
}

// GetGitMetaFast returns all git metadata for the given directory using parallel execution.
// This is ~3x faster than GetGitMeta by running independent git commands concurrently.
// Returns ErrNotGitRepo if the directory is not inside a git repository.
// Returns an error if git commands fail. Partial results may be returned with an error.
// Verbose logging is enabled when VERBOSE environment variable is set.
func GetGitMetaFast(cwd string) (*GitMeta, error) {
	verbose := os.Getenv("VERBOSE") == "true"

	// Check if git is available first to avoid unnecessary goroutine spawning
	if !IsGitAvailable() {
		if verbose {
			fmt.Fprintf(os.Stderr, "git command not found in PATH\n")
		}
		return nil, ErrNotGitRepo
	}

	// Shared result container with mutex for thread safety
	var result struct {
		meta      GitMeta
		errors    []error
		mu        sync.Mutex
		success   bool // Track if any operation succeeded
	}

	// Use WaitGroup to wait for all goroutines
	var wg sync.WaitGroup
	wg.Add(3) // Three operations to run in parallel

	// Goroutine 1: Get remote URLs
	go func() {
		defer wg.Done()
		urls, err := GetUniqueRemoteURLs(cwd)
		if err != nil {
			result.mu.Lock()
			result.errors = append(result.errors, fmt.Errorf("get remote URLs: %w", err))
			result.mu.Unlock()
			if verbose {
				fmt.Fprintf(os.Stderr, "failed to get git remote URLs: %v\n", err)
			}
			return
		}
		result.mu.Lock()
		result.meta.RemoteURLs = urls
		result.success = true
		result.mu.Unlock()
	}()

	// Goroutine 2: Get branch name
	go func() {
		defer wg.Done()
		branch, err := GetBranch(cwd)
		if err != nil {
			result.mu.Lock()
			result.errors = append(result.errors, fmt.Errorf("get branch: %w", err))
			result.mu.Unlock()
			if verbose {
				fmt.Fprintf(os.Stderr, "failed to get git branch: %v\n", err)
			}
			return
		}
		result.mu.Lock()
		result.meta.Branch = branch
		result.success = true
		result.mu.Unlock()
	}()

	// Goroutine 3: Get commit hash
	go func() {
		defer wg.Done()
		hash, err := GetCommitHash(cwd, false)
		if err != nil {
			result.mu.Lock()
			result.errors = append(result.errors, fmt.Errorf("get commit hash: %w", err))
			result.mu.Unlock()
			if verbose {
				fmt.Fprintf(os.Stderr, "failed to get git commit hash: %v\n", err)
			}
			return
		}
		result.mu.Lock()
		result.meta.CommitHash = hash
		result.success = true
		result.mu.Unlock()
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// If no operations succeeded, we're not in a git repo
	if !result.success {
		return nil, ErrNotGitRepo
	}

	// Return combined error if any operations failed, but still return partial metadata
	if len(result.errors) > 0 {
		return &result.meta, errors.Join(result.errors...)
	}

	return &result.meta, nil
}

// IsGitRepo checks if the directory is inside a git repository.
// Returns false if git is not installed or if the directory is not a git repo.
func IsGitRepo(cwd string) bool {
	// First check if git command is available
	if !IsGitAvailable() {
		return false
	}

	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// IsGitAvailable checks if the git binary is available in the system PATH.
func IsGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

// GetUniqueRemoteURLs returns deduplicated list of remote URLs.
func GetUniqueRemoteURLs(cwd string) ([]string, error) {
	remotes, err := GetGitRemotes(cwd)
	if err != nil {
		return nil, err
	}

	// Deduplicate URLs
	seen := make(map[string]bool)
	var urls []string
	for _, remote := range remotes {
		if !seen[remote.URL] {
			seen[remote.URL] = true
			urls = append(urls, remote.URL)
		}
	}
	return urls, nil
}

// GetGitRemotes runs "git -C <cwd> remote -v" and parses the output.
func GetGitRemotes(cwd string) ([]GitRemote, error) {
	cmd := exec.Command("git", "-C", cwd, "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseGitRemoteOutput(string(output)), nil
}

// parseGitRemoteOutput parses the output of "git remote -v".
// Expected format: "origin\tgit@github.com:user/repo.git (fetch)"
func parseGitRemoteOutput(output string) []GitRemote {
	var remotes []GitRemote
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "name\turl (type)"
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		name := parts[0]
		urlPart := parts[1]

		// Extract URL and type from "url (type)"
		// URL may contain spaces, so we find the last " ("
		openParen := strings.LastIndex(urlPart, " (")
		if openParen == -1 {
			continue // Skip malformed lines
		}

		url := urlPart[:openParen]
		typePart := strings.TrimPrefix(urlPart[openParen:], " (")
		remoteType := strings.TrimSuffix(typePart, ")")

		remotes = append(remotes, GitRemote{
			Name: name,
			URL:  url,
			Type: remoteType,
		})
	}

	return remotes
}

// GetBranch returns the current branch name.
func GetBranch(cwd string) (string, error) {
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCommitHash returns the commit hash. Pass full=true for 40-char hash.
func GetCommitHash(cwd string, full bool) (string, error) {
	args := []string{"-C", cwd, "rev-parse"}
	if full {
		args = append(args, "HEAD")
	} else {
		args = append(args, "--short", "HEAD")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ExtractRepoShortName extracts the "owner/repo" short name from various git URL formats.
// Supports SSH URLs (git@github.com:owner/repo.git) and HTTPS URLs (https://github.com/owner/repo.git).
// Returns "owner/repo" or empty string if the URL cannot be parsed.
func ExtractRepoShortName(remoteURL string) string {
	if remoteURL == "" {
		return ""
	}

	// Remove trailing .git if present
	url := strings.TrimSuffix(remoteURL, ".git")

	// Handle SSH format: git@github.com:owner/repo
	if strings.HasPrefix(url, "git@") {
		// Remove "git@" prefix
		url = strings.TrimPrefix(url, "git@")
		// Split by ":" to get the path part
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			url = parts[1]
		}
	} else if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Handle HTTPS format: https://github.com/owner/repo
		// Parse URL to extract path after hostname
		parts := strings.SplitN(url, "://", 2)
		if len(parts) == 2 {
			// Split by "/" and skip the hostname (first part)
			pathParts := strings.SplitN(parts[1], "/", 2)
			if len(pathParts) == 2 && pathParts[1] != "" {
				url = pathParts[1]
			} else {
				url = ""
			}
		}
	} else {
		// Handle other formats (e.g., git://github.com/owner/repo)
		// Split by "/" and take last 2 parts
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			url = parts[len(parts)-2] + "/" + parts[len(parts)-1]
		}
	}

	// Validate that we have something resembling "owner/repo"
	if strings.Contains(url, "/") {
		return url
	}

	return ""
}

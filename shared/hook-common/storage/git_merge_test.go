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


package storage

import (
	"testing"

	hookutils "github.com/code-together/shared/hook-utils"
	"github.com/tidwall/gjson"
)

func TestMergeGitMeta(t *testing.T) {
	eventJSON := []byte(`{
        "session_id": "test-123",
        "hook_event_name": "UserPromptSubmit",
        "cwd": "/path/to/workspace"
    }`)

	gitMeta := &hookutils.GitMeta{
		RemoteURLs: []string{"git@github.com:user/repo.git"},
		Branch:     "main",
		CommitHash: "a1b2c3d",
	}

	merged, err := mergeGitMeta(eventJSON, gitMeta)
	if err != nil {
		t.Fatalf("mergeGitMeta() error = %v", err)
	}

	// Verify git field was added
	result := gjson.GetBytes(merged, "git.remote_urls")
	if !result.Exists() {
		t.Error("git.remote_urls not found in merged event")
	}
	if len(result.Array()) != 1 {
		t.Errorf("expected 1 remote URL, got %d", len(result.Array()))
	}

	branch := gjson.GetBytes(merged, "git.branch").String()
	if branch != "main" {
		t.Errorf("expected branch 'main', got '%s'", branch)
	}

	commit := gjson.GetBytes(merged, "git.commit_hash").String()
	if commit != "a1b2c3d" {
		t.Errorf("expected commit 'a1b2c3d', got '%s'", commit)
	}

	// Verify original fields preserved
	session := gjson.GetBytes(merged, "session_id").String()
	if session != "test-123" {
		t.Errorf("original session_id lost, got '%s'", session)
	}

	eventName := gjson.GetBytes(merged, "hook_event_name").String()
	if eventName != "UserPromptSubmit" {
		t.Errorf("original hook_event_name lost, got '%s'", eventName)
	}
}

func TestMergeGitMeta_MultipleRemotes(t *testing.T) {
	eventJSON := []byte(`{
        "session_id": "test-456",
        "hook_event_name": "TestEvent"
    }`)

	gitMeta := &hookutils.GitMeta{
		RemoteURLs: []string{
			"git@github.com:user/repo.git",
			"https://gitlab.com/user/repo.git",
		},
		Branch:     "feature/test",
		CommitHash: "xyz123",
	}

	merged, err := mergeGitMeta(eventJSON, gitMeta)
	if err != nil {
		t.Fatalf("mergeGitMeta() error = %v", err)
	}

	result := gjson.GetBytes(merged, "git.remote_urls")
	urls := result.Array()
	if len(urls) != 2 {
		t.Errorf("expected 2 remote URLs, got %d", len(urls))
	}
}

func TestMergeGitMeta_EmptyRemotes(t *testing.T) {
	eventJSON := []byte(`{
        "session_id": "test-789",
        "hook_event_name": "TestEvent"
    }`)

	gitMeta := &hookutils.GitMeta{
		RemoteURLs: []string{},
		Branch:     "main",
		CommitHash: "abc123",
	}

	merged, err := mergeGitMeta(eventJSON, gitMeta)
	if err != nil {
		t.Fatalf("mergeGitMeta() error = %v", err)
	}

	// Should have empty array for remote_urls
	result := gjson.GetBytes(merged, "git.remote_urls")
	if !result.Exists() {
		t.Error("git.remote_urls not found in merged event")
	}
}

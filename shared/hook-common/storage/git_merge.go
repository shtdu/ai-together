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


package storage

import (
	hookutils "github.com/code-together/shared/hook-utils"
	"github.com/tidwall/sjson"
)

// mergeGitMeta merges git metadata into event JSON.
// Adds a "git" field with remote_urls, branch, and commit_hash.
func mergeGitMeta(eventJSON []byte, gitMeta *hookutils.GitMeta) ([]byte, error) {
	var err error

	// Add nested git object with remote_urls array
	eventJSON, err = sjson.SetBytes(eventJSON, "git.remote_urls", gitMeta.RemoteURLs)
	if err != nil {
		return nil, err
	}

	eventJSON, err = sjson.SetBytes(eventJSON, "git.branch", gitMeta.Branch)
	if err != nil {
		return nil, err
	}

	eventJSON, err = sjson.SetBytes(eventJSON, "git.commit_hash", gitMeta.CommitHash)
	if err != nil {
		return nil, err
	}

	return eventJSON, nil
}

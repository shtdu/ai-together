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


package events

// Runs when user submits a prompt, before Claude processes it
type UserPromptSubmit struct {
	// Current working directory
	Cwd string `json:"cwd"`
	// Permission mode (e.g., 'default', 'acceptEdits')
	PermissionMode string `json:"permission_mode"`
	// User prompt submitted
	Prompt string `json:"prompt"`
	// Unique identifier for session
	SessionId string `json:"session_id"`
	// Path to the transcript file
	TranscriptPath string `json:"transcript_path"`
}

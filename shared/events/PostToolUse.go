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

// Runs after tool calls complete
type PostToolUse struct {
	// Current working directory
	Cwd string `json:"cwd"`
	// Permission mode (e.g., 'default', 'acceptEdits')
	PermissionMode string `json:"permission_mode"`
	// Unique identifier for session
	SessionId string `json:"session_id"`
	// Input parameters for the tool
	ToolInput map[string]interface{} `json:"tool_input"`
	// Name of the tool being called
	ToolName string `json:"tool_name"`
	ToolResponse map[string]interface{} `json:"tool_response"`
	// Unique identifier for this tool use
	ToolUseId string `json:"tool_use_id"`
	// Path to the transcript file
	TranscriptPath string `json:"transcript_path"`
}

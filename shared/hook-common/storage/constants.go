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

// Bucket names used in the BoltDB database
const (
	SessionsBucket = "sessions"
)

// GetBucketName returns the appropriate bucket name based on tool name.
// If toolName is empty, returns the default "sessions" bucket.
// Otherwise returns "sessions.{toolName}" for tool-specific storage.
func GetBucketName(toolName string) string {
	if toolName == "" {
		return SessionsBucket
	}
	return SessionsBucket + "." + toolName
}

// GetToolName extracts the tool name from a bucket name.
// For example, "sessions.claude-code" returns "claude-code".
// "sessions" returns "" (default tool).
func GetToolName(bucketName string) string {
	if bucketName == "" || bucketName == SessionsBucket {
		return ""
	}
	// Remove "sessions." prefix if present
	if len(bucketName) > len(SessionsBucket)+1 &&
		bucketName[:len(SessionsBucket)+1] == SessionsBucket+"." {
		return bucketName[len(SessionsBucket)+1:]
	}
	return bucketName
}

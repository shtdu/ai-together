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
	"testing"
)

func TestExtractCWD(t *testing.T) {
	tests := []struct {
		name     string
		jsonData []byte
		want     string
	}{
		{
			name:     "valid JSON with cwd field",
			jsonData: []byte(`{"cwd": "/Users/jian/workspaces/github/code-together", "other": "data"}`),
			want:     "/Users/jian/workspaces/github/code-together",
		},
		{
			name:     "valid JSON with empty cwd",
			jsonData: []byte(`{"cwd": ""}`),
			want:     "",
		},
		{
			name:     "valid JSON without cwd field",
			jsonData: []byte(`{"other": "data", "foo": "bar"}`),
			want:     "",
		},
		{
			name:     "invalid JSON",
			jsonData: []byte(`{invalid json}`),
			want:     "",
		},
		{
			name:     "empty input",
			jsonData: []byte{},
			want:     "",
		},
		{
			name:     "null input",
			jsonData: []byte(`null`),
			want:     "",
		},
		{
			name:     "cwd as non-string",
			jsonData: []byte(`{"cwd": 123}`),
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractCWD(tt.jsonData); got != tt.want {
				t.Errorf("ExtractCWD() = %v, want %v", got, tt.want)
			}
		})
	}
}

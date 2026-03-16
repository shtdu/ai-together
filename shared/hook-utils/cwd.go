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

import "encoding/json"

// ExtractCWD extracts the "cwd" field from raw JSON event data.
// Returns empty string if cwd is not present or JSON is invalid.
func ExtractCWD(jsonData []byte) string {
	var data struct {
		Cwd string `json:"cwd"`
	}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return ""
	}
	return data.Cwd
}

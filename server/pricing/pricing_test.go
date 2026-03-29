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

package pricing

import "testing"

func TestGetPrice_KnownModel(t *testing.T) {
	p := GetPrice("claude-3-5-sonnet-20241022")
	if p.Input != 3.0 || p.Output != 15.0 {
		t.Errorf("expected input=3.0 output=15.0, got input=%.1f output=%.1f", p.Input, p.Output)
	}
}

func TestGetPrice_UnknownModel(t *testing.T) {
	p := GetPrice("some-unknown-model")
	if p != DefaultPrice {
		t.Errorf("expected DefaultPrice for unknown model, got %+v", p)
	}
}

func TestGetPrice_EmptyModel(t *testing.T) {
	p := GetPrice("")
	if p != DefaultPrice {
		t.Errorf("expected DefaultPrice for empty model, got %+v", p)
	}
}

func TestEstimateCost(t *testing.T) {
	tests := []struct {
		model        string
		input        int64
		output       int64
		wantCostMin  float64
		wantCostMax  float64
	}{
		{"claude-3-5-sonnet-20241022", 1_000_000, 1_000_000, 18.0, 18.0},   // $3 + $15
		{"gpt-4o-mini", 1_000_000, 1_000_000, 0.75, 0.75},                  // $0.15 + $0.60
		{"claude-3-5-sonnet-20241022", 0, 0, 0.0, 0.0},                     // zero tokens
		{"claude-3-5-sonnet-20241022", 1000, 500, 0.0105, 0.0105},           // small values
		{"unknown-model", 1_000_000, 1_000_000, 18.0, 18.0},                // fallback
	}

	for _, tt := range tests {
		got := EstimateCost(tt.model, tt.input, tt.output)
		tolerance := 0.001
		if got < tt.wantCostMin-tolerance || got > tt.wantCostMax+tolerance {
			t.Errorf("EstimateCost(%s, %d, %d) = %f, want [%f, %f]",
				tt.model, tt.input, tt.output, got, tt.wantCostMin, tt.wantCostMax)
		}
	}
}

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

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// UserFixture represents a test user from the fixtures file.
type UserFixture struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// ProviderFixture represents a test provider from the fixtures file.
type ProviderFixture struct {
	Name            string   `json:"name"`
	Kind            string   `json:"kind"`
	APIKey          string   `json:"api_key"`
	APIURL          string   `json:"api_url"`
	Enabled         bool     `json:"enabled"`
	Level           int      `json:"level"`
	SupportedModels []string `json:"supported_models"`
}

// Fixtures holds all test fixture data.
type Fixtures struct {
	Users     map[string]UserFixture     `json:"users"`
	Providers map[string]ProviderFixture `json:"providers"`
}

// LoadFixtures loads all fixture data from JSON files.
func LoadFixtures() (*Fixtures, error) {
	fixtures := &Fixtures{
		Users:     make(map[string]UserFixture),
		Providers: make(map[string]ProviderFixture),
	}

	// Load users fixture
	usersFile := filepath.Join(TestDataDir, "fixtures", "users.json")
	if data, err := os.ReadFile(usersFile); err == nil {
		if err := json.Unmarshal(data, &fixtures.Users); err != nil {
			return nil, err
		}
	}

	// Load providers fixture
	providersFile := filepath.Join(TestDataDir, "fixtures", "providers.json")
	if data, err := os.ReadFile(providersFile); err == nil {
		if err := json.Unmarshal(data, &fixtures.Providers); err != nil {
			return nil, err
		}
	}

	return fixtures, nil
}

// GetFixtureData is a convenience function to load fixtures.
// It returns cached fixtures or loads them if not yet loaded.
var fixtureCache *Fixtures

func GetFixtureData() (*Fixtures, error) {
	if fixtureCache != nil {
		return fixtureCache, nil
	}

	fixtures, err := LoadFixtures()
	if err != nil {
		return nil, err
	}

	fixtureCache = fixtures
	return fixtures, nil
}

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


// Package integration provides a type-safe client for the AI Together member client API.
// It is generated from the OpenAPI specification at docs/client_api/server_api.yaml.
//
// This spec is optimized for member client integration and includes:
// - Authentication endpoints (login, register, refresh, verify)
// - Full provider management (CRUD for managers)
// - Read-only team information
// - Usage tracking and analytics
// - User profile management
//
// To regenerate the client after API changes:
//   make shared-api-gen       # From repository root
//   cd shared/integration && go generate  # From shared/integration/ directory
//
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -package integration -generate types,client -o generated.go ../../docs/client_api/server_api.yaml
package integration

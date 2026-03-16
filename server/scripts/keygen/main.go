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

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	license "github.com/vitalvas/go-license/license"
)

// LicenseData is the custom data embedded in go-license's Data field
type LicenseData struct {
	LicenseType       string `json:"license_type"`
	MaxSeats          int    `json:"max_seats"`
	MaxTeams          int    `json:"max_teams"`
	DataRetentionDays int    `json:"data_retention_days"`
}

func main() {
	// Flags
	genKeys := flag.Bool("genkeys", false, "Generate new Ed25519 key pair")
	privKeyFile := flag.String("privkey", "private.key", "Private key file")
	customer := flag.String("customer", "", "Customer name")
	licenseType := flag.String("type", "opensource", "License type (opensource, commercial)")
	maxSeats := flag.Int("maxseats", -1, "Max seats (soft limit, -1 for unlimited)")
	maxTeams := flag.Int("maxteams", 1, "Max teams (1 for opensource, -1 for unlimited)")
	dataRetention := flag.Int("retention", 7, "Data retention days (7 for opensource, 90 for commercial)")
	days := flag.Int("days", 365, "Validity in days")
	flag.Parse()

	if *genKeys {
		generateKeyPair()
		return
	}

	// Auto-generate license ID as UUID
	licenseID := uuid.New().String()

	// Load private key
	privKeyBytes, err := os.ReadFile(*privKeyFile)
	if err != nil {
		log.Fatal("Failed to read private key:", err)
	}
	privateKey := ed25519.PrivateKey(privKeyBytes)

	// Create license data
	licenseData := LicenseData{
		LicenseType:       *licenseType,
		MaxSeats:          *maxSeats,
		MaxTeams:          *maxTeams,
		DataRetentionDays: *dataRetention,
	}
	dataBytes, _ := json.Marshal(licenseData)

	issuedAt := time.Now().Unix()
	expiredAt := issuedAt + int64(*days)*24*60*60
	if issuedAt >= expiredAt {
		issuedAt = expiredAt - 24*60*60 // 1 day before expired
	}
	// fmt.Fprintf(os.Stderr, "Issued at: %d, Expired at: %d\n", issuedAt, expiredAt)
	// Create license
	lic := &license.License{
		ID:        licenseID,
		Customer:  *customer,
		Type:      *licenseType,
		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
		Data:      dataBytes,
	}

	// Print license ID for reference
	fmt.Fprintf(os.Stderr, "License ID: %s\n", licenseID)

	// Encode license
	encoded, err := lic.Encode(privateKey)
	if err != nil {
		log.Fatal("Failed to encode license:", err)
	}

	fmt.Println(string(encoded))
}

func generateKeyPair() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("private.key", priv, 0600)
	if err != nil {
		log.Fatal("Failed to write private key:", err)
	}

	err = os.WriteFile("public.key", pub, 0644)
	if err != nil {
		log.Fatal("Failed to write public key:", err)
	}

	fmt.Println("Keys generated successfully:")
	fmt.Println("Private key saved to: private.key")
	fmt.Println("Public key saved to: public.key")
	fmt.Println()
	fmt.Println("Public key (for LICENSE_PUBLIC_KEY env):")
	fmt.Println(base64.StdEncoding.EncodeToString(pub))
}

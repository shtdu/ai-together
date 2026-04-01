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

package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidatePasswordStrength_ValidPasswords tests passwords that should pass validation
func TestValidatePasswordStrength_ValidPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "minimum valid password with all requirements",
			password: "Abc123!@",
		},
		{
			name:     "password with common special characters",
			password: "Password123!",
		},
		{
			name:     "password with multiple special characters",
			password: "Secure#Pass$2025",
		},
		{
			name:     "password with underscore",
			password: "My_Pass123",
		},
		{
			name:     "password with hyphen",
			password: "My-Pass-123",
		},
		{
			name:     "longer password with all character types",
			password: "VerySecurePassword123!@#",
		},
		{
			name:     "password with bracket special characters",
			password: "Test[123]Pass",
		},
		{
			name:     "password with curly braces",
			password: "Test{123}Pass",
		},
		{
			name:     "password with pipe character",
			password: "Test|123|Pass",
		},
		{
			name:     "password with comma and period",
			password: "Test,123.Pass",
		},
		{
			name:     "password with question mark",
			password: "Test?123?Pass",
		},
		{
			name:     "password with slash",
			password: "Test/123/Pass",
		},
		{
			name:     "password with equals sign",
			password: "Test=123=Pass",
		},
		{
			name:     "password with plus sign",
			password: "Test+123+Pass",
		},
		{
			name:     "password with percent sign",
			password: "Test%123%Pass",
		},
		{
			name:     "password with caret",
			password: "Test^123^Pass",
		},
		{
			name:     "password with ampersand",
			password: "Test&123&Pass",
		},
		{
			name:     "password with asterisk",
			password: "Test*123*Pass",
		},
		{
			name:     "password with single quote",
			password: "Test'123'Pass",
		},
		{
			name:     "password with double quote",
			password: `Test"123"Pass`,
		},
		{
			name:     "password with colon",
			password: "Test:123:Pass",
		},
		{
			name:     "password with semicolon",
			password: "Test;123;Pass",
		},
		{
			name:     "password with at sign",
			password: "Test@123@Pass",
		},
		{
			name:     "password with hash",
			password: "Test#123#Pass",
		},
		{
			name:     "password with tilde (not in special char regex, should fail)",
			password: "Test~123~Pass",
		},
		{
			name:     "password with backtick (not in special char regex, should fail)",
			password: "Test`123`Pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			// Tilde and backtick are not in the special character regex
			if tt.name == "password with tilde (not in special char regex, should fail)" ||
				tt.name == "password with backtick (not in special char regex, should fail)" {
				assert.Error(t, err, "Password should be invalid: %s", tt.password)
				assert.Contains(t, err.Error(), "special character")
			} else {
				assert.NoError(t, err, "Password should be valid: %s", tt.password)
			}
		})
	}
}

// TestValidatePasswordStrength_TooShort tests passwords that are too short
func TestValidatePasswordStrength_TooShort(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "single character",
			password: "A",
		},
		{
			name:     "two characters",
			password: "A1",
		},
		{
			name:     "three characters",
			password: "A1a",
		},
		{
			name:     "four characters",
			password: "A1a!",
		},
		{
			name:     "five characters",
			password: "A1a!B",
		},
		{
			name:     "six characters",
			password: "A1a!B2",
		},
		{
			name:     "seven characters",
			password: "A1a!B2b",
		},
		{
			name:     "seven characters with all types except length",
			password: "A1a!B2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			require.Error(t, err, "Password should be invalid due to length")
			assert.Contains(t, err.Error(), "at least 8 characters",
				"Error message should mention minimum length requirement")
		})
	}
}

// TestValidatePasswordStrength_NoUppercase tests passwords without uppercase letters
func TestValidatePasswordStrength_NoUppercase(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "lowercase and digits only",
			password: "password123!",
		},
		{
			name:     "lowercase, digits, special but no uppercase",
			password: "test123!@#",
		},
		{
			name:     "all lowercase and special",
			password: "password!@#",
		},
		{
			name:     "numbers and special only",
			password: "12345678!@#",
		},
		{
			name:     "special characters only",
			password: "!@#$%^&*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			require.Error(t, err, "Password should be invalid without uppercase letter")
			assert.Contains(t, err.Error(), "uppercase",
				"Error message should mention uppercase requirement")
		})
	}
}

// TestValidatePasswordStrength_NoLowercase tests passwords without lowercase letters
func TestValidatePasswordStrength_NoLowercase(t *testing.T) {
	tests := []struct {
		name                  string
		password              string
		expectedErrorContains string
	}{
		{
			name:                  "uppercase and digits only",
			password:              "PASSWORD123!",
			expectedErrorContains: "lowercase",
		},
		{
			name:                  "uppercase, digits, special but no lowercase",
			password:              "TEST123!@#",
			expectedErrorContains: "lowercase",
		},
		{
			name:                  "all uppercase and special",
			password:              "PASSWORD!@#",
			expectedErrorContains: "lowercase",
		},
		{
			name:                  "numbers and special only - fails uppercase first",
			password:              "12345678!@#",
			expectedErrorContains: "uppercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			require.Error(t, err, "Password should be invalid")
			assert.Contains(t, err.Error(), tt.expectedErrorContains,
				"Error message should mention the expected requirement")
		})
	}
}

// TestValidatePasswordStrength_NoDigit tests passwords without digits
func TestValidatePasswordStrength_NoDigit(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "letters and special only",
			password: "Password!@#",
		},
		{
			name:     "uppercase, lowercase, special but no digit",
			password: "TestPass!@#",
		},
		{
			name:     "all letters and special",
			password: "MyPassword!@#",
		},
		{
			name:     "only letters",
			password: "PasswordNo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			require.Error(t, err, "Password should be invalid without digit")
			assert.Contains(t, err.Error(), "digit",
				"Error message should mention digit requirement")
		})
	}
}

// TestValidatePasswordStrength_NoSpecial tests passwords without special characters
func TestValidatePasswordStrength_NoSpecial(t *testing.T) {
	tests := []struct {
		name                  string
		password              string
		expectedErrorContains string
	}{
		{
			name:                  "letters and digits only",
			password:              "Password123",
			expectedErrorContains: "special character",
		},
		{
			name:                  "uppercase, lowercase, digit but no special",
			password:              "TestPass123",
			expectedErrorContains: "special character",
		},
		{
			name:                  "all alphanumeric",
			password:              "MyPassword2025",
			expectedErrorContains: "special character",
		},
		{
			name:                  "only letters - fails digit first",
			password:              "PasswordNo",
			expectedErrorContains: "digit",
		},
		{
			name:                  "only digits - fails uppercase first",
			password:              "12345678",
			expectedErrorContains: "uppercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			require.Error(t, err, "Password should be invalid")
			assert.Contains(t, err.Error(), tt.expectedErrorContains,
				"Error message should mention the expected requirement")
		})
	}
}

// TestValidatePasswordStrength_MultipleFailures tests passwords that fail multiple requirements
func TestValidatePasswordStrength_MultipleFailures(t *testing.T) {
	tests := []struct {
		name             string
		password         string
		expectedContains []string
	}{
		{
			name:     "only lowercase letters - fails length, uppercase, digit, special",
			password: "pass",
			expectedContains: []string{
				"at least 8 characters",
				"uppercase",
				"digit",
				"special character",
			},
		},
		{
			name:     "only uppercase letters - fails length, lowercase, digit, special",
			password: "PASS",
			expectedContains: []string{
				"at least 8 characters",
				"lowercase",
				"digit",
				"special character",
			},
		},
		{
			name:     "only digits - fails length, uppercase, lowercase, special",
			password: "1234",
			expectedContains: []string{
				"at least 8 characters",
				"uppercase",
				"lowercase",
				"special character",
			},
		},
		{
			name:     "only special characters - fails length, uppercase, lowercase, digit",
			password: "!@#",
			expectedContains: []string{
				"at least 8 characters",
				"uppercase",
				"lowercase",
				"digit",
			},
		},
		{
			name:     "lowercase and uppercase but no digit or special - fails digit, special",
			password: "PasswordNo",
			expectedContains: []string{
				"digit",
				"special character",
			},
		},
		{
			name:     "lowercase and digit but no uppercase or special - fails uppercase, special",
			password: "password123",
			expectedContains: []string{
				"uppercase",
				"special character",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			require.Error(t, err, "Password should be invalid")

			// Check that the error message contains at least one of the expected failure reasons
			// (since the function returns on first failure, we only check for the first one)
			found := false
			for _, expected := range tt.expectedContains {
				if len(tt.password) < 8 && expected == "at least 8 characters" {
					// Length check comes first, so this should be in the error
					if assert.Contains(t, err.Error(), expected) {
						found = true
						break
					}
				} else if len(tt.password) >= 8 {
					// If length is OK, check for other requirements
					if assert.Contains(t, err.Error(), expected) {
						found = true
						break
					}
				}
			}
			assert.True(t, found, "Error should mention one of the missing requirements")
		})
	}
}

// TestValidatePasswordStrength_BoundaryCases tests edge cases and boundary conditions
func TestValidatePasswordStrength_BoundaryCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "exactly 8 characters with all requirements",
			password: "A1b2C3d!",
			wantErr:  false,
		},
		{
			name:     "7 characters fails length check",
			password: "A1b2C3d",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "password with only first character uppercase",
			password: "Aaaaaaaa1!",
			wantErr:  false,
		},
		{
			name:     "password with only last character uppercase",
			password: "aaaaaaa1!A",
			wantErr:  false,
		},
		{
			name:     "password with only first character lowercase",
			password: "AAAAAAAA1!a",
			wantErr:  false,
		},
		{
			name:     "password with only last character lowercase",
			password: "AAAAAAAA1!a",
			wantErr:  false,
		},
		{
			name:     "password with only first character digit",
			password: "1AAAAAAa!",
			wantErr:  false,
		},
		{
			name:     "password with only last character digit",
			password: "AAAAAAa!1",
			wantErr:  false,
		},
		{
			name:     "password with only first character special",
			password: "!AAAAAAa1",
			wantErr:  false,
		},
		{
			name:     "password with only last character special",
			password: "AAAAAAa1!",
			wantErr:  false,
		},
		{
			name:     "very long password",
			password: "ThisIsAVeryLongPassword123!@#ThatShouldStillBeValid",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidatePasswordStrength_CommonPasswordPatterns tests some common weak password patterns
func TestValidatePasswordStrength_CommonPasswordPatterns(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		reason   string
	}{
		{
			name:     "password with only repeated characters (no lowercase)",
			password: "AAAAAAAA1!",
			wantErr:  true, // No lowercase letter
			reason:   "no lowercase letter",
		},
		{
			name:     "password with repeated characters but all types",
			password: "AaAaAa1!",
			wantErr:  false, // Has all required types even if repeated
			reason:   "repeated characters are technically valid by current rules",
		},
		{
			name:     "sequential letters with numbers and special",
			password: "Abcdef123!",
			wantErr:  false, // Technically valid by our rules
			reason:   "sequential patterns are technically valid by current rules",
		},
		{
			name:     "keyboard pattern with numbers and special",
			password: "Qwer1234!",
			wantErr:  false, // Technically valid by our rules
			reason:   "keyboard patterns are technically valid by current rules",
		},
		{
			name:     "common word with numbers and special",
			password: "Password123!",
			wantErr:  false, // Technically valid by our rules
			reason:   "dictionary words are technically valid by current rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, tt.reason)
			}
		})
	}
}

// TestValidatePasswordStrength_Whitespace tests passwords with whitespace
func TestValidatePasswordStrength_Whitespace(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		note     string
	}{
		{
			name:     "password with leading space",
			password: " Abc123!@",
			wantErr:  false, // Has all required types, space is just extra char
			note:     "space is an extra character but password has all required types",
		},
		{
			name:     "password with trailing space",
			password: "Abc123!@ ",
			wantErr:  false, // Has all required types
			note:     "space is an extra character but password has all required types",
		},
		{
			name:     "password with spaces but otherwise valid",
			password: "Abc 123!@",
			wantErr:  false, // Has all required types
			note:     "spaces are extra characters but password has all required types",
		},
		{
			name:     "password with only spaces as special chars",
			password: "Abc 123 ",
			wantErr:  true, // No special character from the allowed set
			note:     "space is not in the special character list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				assert.Error(t, err, tt.note)
			} else {
				assert.NoError(t, err, tt.note)
			}
		})
	}
}

// TestValidatePasswordStrength_Unicode tests passwords with Unicode characters
func TestValidatePasswordStrength_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		note     string
	}{
		{
			name:     "password with emoji plus all required types",
			password: "Abc123!😀",
			wantErr:  false, // Has all required ASCII types, emoji is extra
			note:     "emoji is extra but password already has all required character types",
		},
		{
			name:     "password with accented characters",
			password: "Abc123!é",
			wantErr:  false, // Go regex [a-z] matches accented lowercase letters
			note:     "Go regex is Unicode-aware, accented chars count as lowercase",
		},
		{
			name:     "password with Chinese characters plus all required types",
			password: "Abc123!中",
			wantErr:  false, // Has all required ASCII types
			note:     "Chinese chars are extra but password has all required types",
		},
		{
			name:     "password with only Chinese characters and digits",
			password: "中中中123!",
			wantErr:  true, // No ASCII uppercase or lowercase letters
			note:     "no ASCII letters, Chinese chars don't match [A-Z] or [a-z]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				assert.Error(t, err, tt.note)
			} else {
				assert.NoError(t, err, tt.note)
			}
		})
	}
}

// BenchmarkValidatePasswordStrength benchmarks the password validation function
func BenchmarkValidatePasswordStrength(b *testing.B) {
	passwords := []string{
		"Abc123!@",
		"VerySecurePassword123!@#",
		"Password123!",
	}

	for _, pwd := range passwords {
		b.Run(pwd, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ValidatePasswordStrength(pwd)
			}
		})
	}
}

// BenchmarkValidatePasswordStrength_Invalid benchmarks validation with invalid passwords
func BenchmarkValidatePasswordStrength_Invalid(b *testing.B) {
	invalidPasswords := []string{
		"",           // empty
		"short",      // too short
		"noupper1!",  // no uppercase
		"NOLOWER1!",  // no lowercase
		"NoDigit!!",  // no digit
		"NoSpecial1", // no special
	}

	for _, pwd := range invalidPasswords {
		b.Run(pwd, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ValidatePasswordStrength(pwd)
			}
		})
	}
}

package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple text", "hello world"},
		{"api key", "sk-1234567890abcdefghijklmnopqrstuvwxyz"},
		{"password with special chars", "P@ssw0rd!#$%^&*()"},
		{"unicode text", "你好世界🌍"},
		{"long text", strings.Repeat("a", 1000)},
		{"json-like", `{"key":"value","secret":"token123"}`},
		{"url", "https://api.example.com/v1/endpoint?key=secret123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Empty plaintext should return empty encrypted text
			if tt.plaintext == "" {
				if encrypted != "" {
					t.Errorf("Expected empty encrypted string for empty plaintext, got %q", encrypted)
				}
				return
			}

			// Encrypted text should not equal plaintext
			if encrypted == tt.plaintext {
				t.Errorf("Encrypted text should not equal plaintext")
			}

			// Encrypted text should be base64
			if !strings.ContainsAny(encrypted, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=") {
				t.Errorf("Encrypted text doesn't appear to be base64")
			}

			// Decrypt
			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			// Decrypted should match original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptDeterministic(t *testing.T) {
	plaintext := "test-secret-key"

	// Encrypt the same plaintext twice
	encrypted1, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("First Encrypt() error = %v", err)
	}

	encrypted2, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Second Encrypt() error = %v", err)
	}

	// They should be different due to random salt and nonce
	if encrypted1 == encrypted2 {
		t.Errorf("Expected different encrypted values for same plaintext (randomization check), got same: %v", encrypted1)
	}

	// But both should decrypt to the same plaintext
	decrypted1, _ := Decrypt(encrypted1)
	decrypted2, _ := Decrypt(encrypted2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Errorf("Both encrypted values should decrypt to original plaintext")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"not base64", "not-valid-base64!@#$"},
		{"too short", "YWJj"},                 // "abc" in base64, too short
		{"random base64", "SGVsbG8gV29ybGQh"}, // "Hello World!" in base64
		{"corrupted", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.input)
			if err == nil {
				t.Errorf("Decrypt() should return error for invalid input %q", tt.input)
			}
		})
	}
}

func TestIsEncrypted(t *testing.T) {
	// Encrypt a sample value
	plaintext := "test-api-key-123"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Create a long base64 string that could be a legitimate API key (44+ chars)
	longBase64APIKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty string", "", false},
		{"plain text", "hello", false},
		{"api key format", "sk-1234567890", false},
		{"encrypted value", encrypted, true},
		{"short base64", "YWJj", false},
		{"not base64", "not-base64!", false},
		{"long base64 api key", longBase64APIKey, false}, // Should NOT be detected as encrypted (no version marker)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEncrypted(tt.value)
			if result != tt.expected {
				t.Errorf("IsEncrypted(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestGetMachineID(t *testing.T) {
	// Get machine ID
	machineID, err := GetMachineID()
	if err != nil {
		t.Fatalf("GetMachineID() error = %v", err)
	}

	// Should not be empty
	if machineID == "" {
		t.Error("GetMachineID() returned empty string")
	}

	// Should contain hostname, OS, and arch
	if !strings.Contains(machineID, "-") {
		t.Error("GetMachineID() should contain separators")
	}

	// Should be consistent
	machineID2, err := GetMachineID()
	if err != nil {
		t.Fatalf("Second GetMachineID() error = %v", err)
	}

	if machineID != machineID2 {
		t.Errorf("GetMachineID() should be consistent: %v != %v", machineID, machineID2)
	}
}

func TestMachineIDCandidatesAreStableAndNonEmpty(t *testing.T) {
	ids, err := getMachineIDCandidates()
	if err != nil {
		t.Fatalf("getMachineIDCandidates() error = %v", err)
	}
	if len(ids) == 0 {
		t.Fatalf("expected at least one machine ID candidate")
	}
	for i, id := range ids {
		if strings.TrimSpace(id) == "" {
			t.Fatalf("candidate %d is empty", i)
		}
	}
}

func TestDeriveKey(t *testing.T) {
	machineID := "test-machine-linux-amd64"
	salt := []byte("1234567890123456") // 16 bytes

	// Derive key
	key := DeriveKey(machineID, salt)

	// Should be exactly 32 bytes (AES-256)
	if len(key) != keySize {
		t.Errorf("DeriveKey() returned %d bytes, want %d", len(key), keySize)
	}

	// Should be deterministic with same inputs
	key2 := DeriveKey(machineID, salt)
	if string(key) != string(key2) {
		t.Error("DeriveKey() should be deterministic")
	}

	// Should be different with different salt
	salt2 := []byte("6543210987654321")
	key3 := DeriveKey(machineID, salt2)
	if string(key) == string(key3) {
		t.Error("DeriveKey() should produce different keys with different salts")
	}

	// Should be different with different machine ID
	key4 := DeriveKey("different-machine", salt)
	if string(key) == string(key4) {
		t.Error("DeriveKey() should produce different keys with different machine IDs")
	}
}

func BenchmarkEncrypt(b *testing.B) {
	plaintext := "sk-1234567890abcdefghijklmnopqrstuvwxyz"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encrypt(plaintext)
		if err != nil {
			b.Fatalf("Encrypt() error = %v", err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	plaintext := "sk-1234567890abcdefghijklmnopqrstuvwxyz"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		b.Fatalf("Encrypt() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decrypt(encrypted)
		if err != nil {
			b.Fatalf("Decrypt() error = %v", err)
		}
	}
}

func BenchmarkIsEncrypted(b *testing.B) {
	plaintext := "sk-1234567890abcdefghijklmnopqrstuvwxyz"
	encrypted, _ := Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsEncrypted(encrypted)
	}
}

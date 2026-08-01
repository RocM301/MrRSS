package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Number of iterations for PBKDF2 (OWASP recommends 600,000+ for PBKDF2-HMAC-SHA256 as of 2023)
	pbkdf2Iterations = 600000
	// Key size for AES-256
	keySize = 32
	// Salt size for key derivation
	saltSize = 16
	// Version marker to identify encrypted values (prevents false positives in IsEncrypted)
	versionMarker = "MrRSS-v1:"
)

var (
	// ErrInvalidCiphertext is returned when the ciphertext is too short or invalid
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrDecryptionFailed is returned when decryption fails
	ErrDecryptionFailed = errors.New("decryption failed")

	machineIDsOnce sync.Once
	machineIDs     []string
	machineIDsErr  error
)

// GetMachineID generates a machine-specific identifier for key derivation.
// This ensures that encrypted data is tied to the specific machine.
// On macOS it prefers the stable hardware UUID; hostnames are only used as
// legacy fallbacks because they can change with network conditions.
func GetMachineID() (string, error) {
	ids, err := getMachineIDCandidates()
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", fmt.Errorf("no machine identifiers available")
	}
	return ids[0], nil
}

func getMachineIDCandidates() ([]string, error) {
	machineIDsOnce.Do(func() {
		machineIDs = buildMachineIDCandidates()
		if len(machineIDs) == 0 {
			machineIDsErr = fmt.Errorf("no machine identifiers available")
		}
	})
	return machineIDs, machineIDsErr
}

func buildMachineIDCandidates() []string {
	ids := []string{}

	if runtime.GOOS == "darwin" {
		if uuid := darwinHardwareUUID(); uuid != "" {
			ids = append(ids, fmt.Sprintf("darwin-hardware-%s-%s", runtime.GOARCH, uuid))
		}
	}

	ids = append(ids, legacyMachineIDs()...)
	return dedupeNonEmpty(ids)
}

func legacyMachineIDs() []string {
	// Use hostname as the primary identifier
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}

	// Try to get machine-id from various system locations (Linux/BSD)
	machineUUID := ""
	possiblePaths := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	}
	for _, path := range possiblePaths {
		if data, err := os.ReadFile(path); err == nil {
			machineUUID = string(data)
			break
		}
	}

	hostnames := []string{hostname}
	if runtime.GOOS == "darwin" {
		if localHostName := darwinScutilValue("LocalHostName"); localHostName != "" {
			hostnames = append(hostnames, localHostName, localHostName+".local", localHostName+".lan")
		}
		hostnames = append(hostnames, darwinScutilValue("HostName"), darwinScutilValue("ComputerName"))
	}

	ids := []string{}
	for _, h := range dedupeNonEmpty(hostnames) {
		ids = append(ids, fmt.Sprintf("%s-%s-%s-%s", h, runtime.GOOS, runtime.GOARCH, machineUUID))
	}
	return ids
}

func darwinHardwareUUID() string {
	out, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "IOPlatformUUID") {
			continue
		}
		parts := strings.Split(line, "\"")
		for i, part := range parts {
			if part == "IOPlatformUUID" && i+2 < len(parts) {
				return strings.TrimSpace(parts[i+2])
			}
		}
	}
	return ""
}

func darwinScutilValue(key string) string {
	out, err := exec.Command("/usr/sbin/scutil", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func dedupeNonEmpty(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

// DeriveKey derives a cryptographic key from a machine ID using PBKDF2.
// The salt is stored with the ciphertext to allow key derivation during decryption.
func DeriveKey(machineID string, salt []byte) []byte {
	return pbkdf2.Key([]byte(machineID), salt, pbkdf2Iterations, keySize, sha256.New)
}

// Encrypt encrypts plaintext using AES-256-GCM with a machine-specific key.
// The output format is: [salt(16 bytes)][nonce(12 bytes)][ciphertext+tag]
// Returns base64-encoded result for safe storage in database.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Get the preferred stable machine ID for key derivation.
	machineID, err := GetMachineID()
	if err != nil {
		return "", fmt.Errorf("failed to get machine ID: %w", err)
	}

	// Generate random salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key
	key := DeriveKey(machineID, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Combine salt, nonce, and ciphertext
	result := make([]byte, 0, saltSize+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	// Encode to base64 and prepend version marker for safe storage
	encoded := base64.StdEncoding.EncodeToString(result)
	return versionMarker + encoded, nil
}

// Decrypt decrypts ciphertext that was encrypted with Encrypt.
// The input must be version-prefixed base64-encoded and contain: [salt][nonce][ciphertext+tag]
func Decrypt(ciphertextBase64 string) (string, error) {
	if ciphertextBase64 == "" {
		return "", nil
	}

	// Check and strip version marker
	if !strings.HasPrefix(ciphertextBase64, versionMarker) {
		return "", fmt.Errorf("missing or invalid version marker")
	}
	ciphertextBase64 = strings.TrimPrefix(ciphertextBase64, versionMarker)

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check minimum length: salt + nonce + tag (at least)
	if len(data) < saltSize+12+16 {
		return "", ErrInvalidCiphertext
	}

	// Extract salt
	salt := data[:saltSize]

	// Try the preferred stable machine ID first, then legacy hostname-derived
	// IDs for values encrypted by older versions.
	machineIDs, err := getMachineIDCandidates()
	if err != nil {
		return "", fmt.Errorf("failed to get machine ID: %w", err)
	}

	for _, machineID := range machineIDs {
		plaintext, err := decryptWithMachineID(machineID, salt, data)
		if err == nil {
			return plaintext, nil
		}
	}

	return "", ErrDecryptionFailed
}

func decryptWithMachineID(machineID string, salt, data []byte) (string, error) {
	key := DeriveKey(machineID, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < saltSize+nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// IsEncrypted checks if a value appears to be encrypted by checking for the version marker.
// This definitively identifies encrypted values and prevents false positives.
func IsEncrypted(value string) bool {
	if value == "" {
		return false
	}

	// Check for version marker - this is definitive, not a heuristic
	return strings.HasPrefix(value, versionMarker)
}

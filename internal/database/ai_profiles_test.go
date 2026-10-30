package database

import (
	"testing"

	"MrRSS/internal/models"
)

func TestUpdateAIProfileCanOverwriteUnreadableAPIKey(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	id, err := db.CreateAIProfile(&models.AIProfile{
		Name:     "broken",
		APIKey:   "old-secret",
		Endpoint: "https://api.example.com/v1/chat/completions",
		Model:    "old-model",
	})
	if err != nil {
		t.Fatalf("CreateAIProfile() error = %v", err)
	}

	if _, err := db.Exec(`UPDATE ai_profiles SET api_key = ? WHERE id = ?`, "MrRSS-v1:not-valid", id); err != nil {
		t.Fatalf("corrupt api key: %v", err)
	}

	if _, err := db.GetAIProfile(id); err == nil {
		t.Fatalf("expected corrupted API key to fail decryption")
	}

	exists, err := db.AIProfileExists(id)
	if err != nil {
		t.Fatalf("AIProfileExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("expected profile to exist")
	}

	if err := db.UpdateAIProfile(&models.AIProfile{
		ID:       id,
		Name:     "fixed",
		APIKey:   "new-secret",
		Endpoint: "https://api.example.com/v1/chat/completions",
		Model:    "new-model",
	}); err != nil {
		t.Fatalf("UpdateAIProfile() error = %v", err)
	}

	profile, err := db.GetAIProfile(id)
	if err != nil {
		t.Fatalf("GetAIProfile() after overwrite error = %v", err)
	}
	if profile.APIKey != "new-secret" {
		t.Fatalf("APIKey = %q, want new-secret", profile.APIKey)
	}
}

func TestUpdateAIProfilePreservingKeyDoesNotDecryptExistingKey(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	id, err := db.CreateAIProfile(&models.AIProfile{
		Name:     "profile",
		APIKey:   "old-secret",
		Endpoint: "https://api.example.com/v1/chat/completions",
		Model:    "old-model",
	})
	if err != nil {
		t.Fatalf("CreateAIProfile() error = %v", err)
	}

	if _, err := db.Exec(`UPDATE ai_profiles SET api_key = ? WHERE id = ?`, "MrRSS-v1:not-valid", id); err != nil {
		t.Fatalf("corrupt api key: %v", err)
	}

	if err := db.UpdateAIProfilePreservingKey(&models.AIProfile{
		ID:       id,
		Name:     "renamed",
		Endpoint: "https://api.example.com/v1/chat/completions",
		Model:    "new-model",
	}); err != nil {
		t.Fatalf("UpdateAIProfilePreservingKey() error = %v", err)
	}

	var storedKey string
	if err := db.QueryRow(`SELECT api_key FROM ai_profiles WHERE id = ?`, id).Scan(&storedKey); err != nil {
		t.Fatalf("read stored key: %v", err)
	}
	if storedKey != "MrRSS-v1:not-valid" {
		t.Fatalf("stored key = %q, want original corrupted key preserved", storedKey)
	}
}

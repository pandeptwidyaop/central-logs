package models

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create mcp_tokens table
	_, err = db.Exec(`
		CREATE TABLE mcp_tokens (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			token_prefix TEXT NOT NULL,
			granted_projects TEXT,
			expires_at DATETIME,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_by TEXT NOT NULL,
			last_used_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create mcp_tokens table: %v", err)
	}

	// Create indexes
	_, err = db.Exec("CREATE INDEX idx_mcp_tokens_token_hash ON mcp_tokens(token_hash)")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	return db
}

func TestGenerateMCPToken(t *testing.T) {
	token, hash, prefix, err := GenerateMCPToken()
	if err != nil {
		t.Fatalf("GenerateMCPToken failed: %v", err)
	}

	// Test token format
	if len(token) != 68 { // mcp_ (4) + 64 hex chars
		t.Errorf("Expected token length 68, got %d", len(token))
	}

	if token[:4] != "mcp_" {
		t.Errorf("Expected token to start with 'mcp_', got %s", token[:4])
	}

	// Test hash
	if len(hash) != 64 { // SHA256 hex = 64 chars
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// Test prefix
	expectedPrefix := token[:13] + "..."
	if prefix != expectedPrefix {
		t.Errorf("Expected prefix '%s', got '%s'", expectedPrefix, prefix)
	}

	// Test uniqueness - generate multiple tokens
	token2, _, _, err := GenerateMCPToken()
	if err != nil {
		t.Fatalf("GenerateMCPToken failed on second call: %v", err)
	}

	if token == token2 {
		t.Error("Generated tokens should be unique")
	}
}

func TestHashMCPToken(t *testing.T) {
	token := "mcp_test123456"
	hash1 := HashMCPToken(token)
	hash2 := HashMCPToken(token)

	// Test deterministic hashing
	if hash1 != hash2 {
		t.Error("Hashing same token should produce same hash")
	}

	// Test different tokens produce different hashes
	hash3 := HashMCPToken("mcp_different")
	if hash1 == hash3 {
		t.Error("Different tokens should produce different hashes")
	}

	// Test hash length
	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}
}

func TestMCPTokenRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMCPTokenRepository(db)

	token := &MCPToken{
		Name:            "Test Token",
		GrantedProjects: "*",
		IsActive:        true,
		CreatedBy:       "user123",
	}

	rawToken, err := repo.Create(token)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify raw token returned
	if len(rawToken) != 68 {
		t.Errorf("Expected raw token length 68, got %d", len(rawToken))
	}

	// Verify token has ID
	if token.ID == "" {
		t.Error("Token ID should be set after creation")
	}

	// Verify token has hash
	if token.TokenHash == "" {
		t.Error("Token hash should be set after creation")
	}

	// Verify token has prefix
	if token.TokenPrefix == "" {
		t.Error("Token prefix should be set after creation")
	}

	// Verify token can be retrieved
	retrieved, err := repo.GetByID(token.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved token is nil")
	}

	if retrieved.Name != token.Name {
		t.Errorf("Expected name %s, got %s", token.Name, retrieved.Name)
	}
}

func TestMCPTokenRepository_GetByToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMCPTokenRepository(db)

	token := &MCPToken{
		Name:            "Test Token",
		GrantedProjects: "*",
		IsActive:        true,
		CreatedBy:       "user123",
	}

	rawToken, err := repo.Create(token)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test retrieval by raw token
	retrieved, err := repo.GetByToken(rawToken)
	if err != nil {
		t.Fatalf("GetByToken failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved token is nil")
	}

	if retrieved.ID != token.ID {
		t.Errorf("Expected ID %s, got %s", token.ID, retrieved.ID)
	}

	// Test with invalid token
	invalid, err := repo.GetByToken("mcp_invalid123")
	if err != nil {
		t.Fatalf("GetByToken with invalid token failed: %v", err)
	}

	if invalid != nil {
		t.Error("Expected nil for invalid token")
	}

	// Test with inactive token
	token.IsActive = false
	err = repo.Update(token)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	inactive, err := repo.GetByToken(rawToken)
	if err != nil {
		t.Fatalf("GetByToken with inactive token failed: %v", err)
	}

	if inactive != nil {
		t.Error("Expected nil for inactive token")
	}
}

func TestMCPTokenRepository_GetByToken_Expired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMCPTokenRepository(db)

	// Create token with past expiry
	pastTime := time.Now().Add(-1 * time.Hour)
	token := &MCPToken{
		Name:            "Expired Token",
		GrantedProjects: "*",
		ExpiresAt:       &pastTime,
		IsActive:        true,
		CreatedBy:       "user123",
	}

	rawToken, err := repo.Create(token)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Should return nil for expired token
	retrieved, err := repo.GetByToken(rawToken)
	if err != nil {
		t.Fatalf("GetByToken failed: %v", err)
	}

	if retrieved != nil {
		t.Error("Expected nil for expired token")
	}

	// Create token with future expiry
	futureTime := time.Now().Add(1 * time.Hour)
	token2 := &MCPToken{
		Name:            "Valid Token",
		GrantedProjects: "*",
		ExpiresAt:       &futureTime,
		IsActive:        true,
		CreatedBy:       "user123",
	}

	rawToken2, err := repo.Create(token2)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Should return token with future expiry
	retrieved2, err := repo.GetByToken(rawToken2)
	if err != nil {
		t.Fatalf("GetByToken failed: %v", err)
	}

	if retrieved2 == nil {
		t.Error("Expected token with future expiry to be retrieved")
	}
}

func TestMCPToken_GetGrantedProjectIDs(t *testing.T) {
	tests := []struct {
		name            string
		grantedProjects string
		wantIDs         []string
		wantAllProjects bool
		wantErr         bool
	}{
		{
			name:            "All projects",
			grantedProjects: "*",
			wantIDs:         nil,
			wantAllProjects: true,
			wantErr:         false,
		},
		{
			name:            "Specific projects",
			grantedProjects: `["proj1","proj2","proj3"]`,
			wantIDs:         []string{"proj1", "proj2", "proj3"},
			wantAllProjects: false,
			wantErr:         false,
		},
		{
			name:            "Empty string",
			grantedProjects: "",
			wantIDs:         []string{},
			wantAllProjects: false,
			wantErr:         false,
		},
		{
			name:            "Invalid JSON",
			grantedProjects: "invalid-json",
			wantIDs:         nil,
			wantAllProjects: false,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &MCPToken{
				GrantedProjects: tt.grantedProjects,
			}

			ids, allProjects, err := token.GetGrantedProjectIDs()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetGrantedProjectIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if allProjects != tt.wantAllProjects {
				t.Errorf("GetGrantedProjectIDs() allProjects = %v, want %v", allProjects, tt.wantAllProjects)
			}

			if !tt.wantErr && tt.wantIDs != nil {
				if len(ids) != len(tt.wantIDs) {
					t.Errorf("GetGrantedProjectIDs() ids length = %d, want %d", len(ids), len(tt.wantIDs))
					return
				}
				for i, id := range ids {
					if id != tt.wantIDs[i] {
						t.Errorf("GetGrantedProjectIDs() ids[%d] = %s, want %s", i, id, tt.wantIDs[i])
					}
				}
			}
		})
	}
}

func TestMCPToken_HasAccessToProject(t *testing.T) {
	tests := []struct {
		name            string
		grantedProjects string
		projectID       string
		wantAccess      bool
	}{
		{
			name:            "All projects - has access",
			grantedProjects: "*",
			projectID:       "any-project",
			wantAccess:      true,
		},
		{
			name:            "Specific project - has access",
			grantedProjects: `["proj1","proj2","proj3"]`,
			projectID:       "proj2",
			wantAccess:      true,
		},
		{
			name:            "Specific project - no access",
			grantedProjects: `["proj1","proj2","proj3"]`,
			projectID:       "proj4",
			wantAccess:      false,
		},
		{
			name:            "Empty - no access",
			grantedProjects: "",
			projectID:       "any-project",
			wantAccess:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &MCPToken{
				GrantedProjects: tt.grantedProjects,
			}

			hasAccess, err := token.HasAccessToProject(tt.projectID)
			if err != nil {
				t.Errorf("HasAccessToProject() error = %v", err)
				return
			}

			if hasAccess != tt.wantAccess {
				t.Errorf("HasAccessToProject() = %v, want %v", hasAccess, tt.wantAccess)
			}
		})
	}
}

func TestMCPTokenRepository_UpdateLastUsed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMCPTokenRepository(db)

	token := &MCPToken{
		Name:            "Test Token",
		GrantedProjects: "*",
		IsActive:        true,
		CreatedBy:       "user123",
	}

	_, err := repo.Create(token)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Initially should be nil
	if token.LastUsedAt != nil {
		t.Error("LastUsedAt should be nil initially")
	}

	// Update last used
	time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp difference
	err = repo.UpdateLastUsed(token.ID)
	if err != nil {
		t.Fatalf("UpdateLastUsed failed: %v", err)
	}

	// Retrieve and check
	retrieved, err := repo.GetByID(token.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.LastUsedAt == nil {
		t.Error("LastUsedAt should be set after update")
	}
}

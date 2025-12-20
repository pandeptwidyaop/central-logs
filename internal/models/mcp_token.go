package models

import (
	"central-logs/internal/utils"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MCPToken struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	TokenHash       string     `json:"-"`
	TokenPrefix     string     `json:"token_prefix"`
	GrantedProjects string     `json:"granted_projects"` // JSON array of project IDs or "*"
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	CreatedBy       string     `json:"created_by"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type MCPTokenRepository struct {
	db *sql.DB
}

func NewMCPTokenRepository(db *sql.DB) *MCPTokenRepository {
	return &MCPTokenRepository{db: db}
}

// GenerateMCPToken generates a new MCP token with format: mcp_ + 64 hex characters
func GenerateMCPToken() (token, hash, prefix string, err error) {
	// Generate 32 random bytes (will be 64 hex characters)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}

	// Create the full token with mcp_ prefix
	token = "mcp_" + hex.EncodeToString(bytes)
	prefix = token[:13] + "..." // mcp_ + first 9 hex chars + ...

	// Hash the token for storage
	hashBytes := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(hashBytes[:])

	return token, hash, prefix, nil
}

// HashMCPToken hashes an MCP token using SHA256
func HashMCPToken(token string) string {
	hashBytes := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashBytes[:])
}

// Create creates a new MCP token and returns the raw token (only shown once)
func (r *MCPTokenRepository) Create(mcpToken *MCPToken) (string, error) {
	mcpToken.ID = uuid.New().String()
	mcpToken.CreatedAt = time.Now()
	mcpToken.UpdatedAt = time.Now()

	// Generate token
	rawToken, tokenHash, tokenPrefix, err := GenerateMCPToken()
	if err != nil {
		return "", err
	}
	mcpToken.TokenHash = tokenHash
	mcpToken.TokenPrefix = tokenPrefix

	// Handle nullable fields
	var expiresAt interface{}
	if mcpToken.ExpiresAt != nil {
		expiresAt = mcpToken.ExpiresAt
	}

	_, err = r.db.Exec(`
		INSERT INTO mcp_tokens (id, name, token_hash, token_prefix, granted_projects, expires_at, is_active, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, mcpToken.ID, mcpToken.Name, mcpToken.TokenHash, mcpToken.TokenPrefix, mcpToken.GrantedProjects, expiresAt, mcpToken.IsActive, mcpToken.CreatedBy, mcpToken.CreatedAt, mcpToken.UpdatedAt)

	if err != nil {
		return "", err
	}

	return rawToken, nil // Return the raw token (only time it's available)
}

// GetByID retrieves a token by ID
func (r *MCPTokenRepository) GetByID(id string) (*MCPToken, error) {
	token := &MCPToken{}
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime
	var grantedProjects sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, token_hash, token_prefix, granted_projects, expires_at, is_active, created_by, last_used_at, created_at, updated_at
		FROM mcp_tokens WHERE id = ?
	`, id).Scan(&token.ID, &token.Name, &token.TokenHash, &token.TokenPrefix, &grantedProjects, &expiresAt, &token.IsActive, &token.CreatedBy, &lastUsedAt, &token.CreatedAt, &token.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if grantedProjects.Valid {
		token.GrantedProjects = grantedProjects.String
	}
	if expiresAt.Valid {
		token.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}

	return token, nil
}

// GetByToken retrieves a token by its raw token value (validates and checks expiry)
func (r *MCPTokenRepository) GetByToken(rawToken string) (*MCPToken, error) {
	hashedToken := HashMCPToken(rawToken)

	token := &MCPToken{}
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime
	var grantedProjects sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, token_hash, token_prefix, granted_projects, expires_at, is_active, created_by, last_used_at, created_at, updated_at
		FROM mcp_tokens WHERE token_hash = ? AND is_active = 1
	`, hashedToken).Scan(&token.ID, &token.Name, &token.TokenHash, &token.TokenPrefix, &grantedProjects, &expiresAt, &token.IsActive, &token.CreatedBy, &lastUsedAt, &token.CreatedAt, &token.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if grantedProjects.Valid {
		token.GrantedProjects = grantedProjects.String
	}
	if expiresAt.Valid {
		token.ExpiresAt = &expiresAt.Time
		// Check if token is expired
		if time.Now().After(expiresAt.Time) {
			return nil, nil // Token expired
		}
	}
	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}

	// Additional constant-time verification to prevent timing attacks
	if !utils.SecureCompareHash(token.TokenHash, hashedToken) {
		return nil, nil // Hash mismatch
	}

	return token, nil
}

// GetAll retrieves all tokens
func (r *MCPTokenRepository) GetAll() ([]*MCPToken, error) {
	rows, err := r.db.Query(`
		SELECT id, name, token_hash, token_prefix, granted_projects, expires_at, is_active, created_by, last_used_at, created_at, updated_at
		FROM mcp_tokens ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*MCPToken
	for rows.Next() {
		token := &MCPToken{}
		var expiresAt sql.NullTime
		var lastUsedAt sql.NullTime
		var grantedProjects sql.NullString

		err := rows.Scan(&token.ID, &token.Name, &token.TokenHash, &token.TokenPrefix, &grantedProjects, &expiresAt, &token.IsActive, &token.CreatedBy, &lastUsedAt, &token.CreatedAt, &token.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if grantedProjects.Valid {
			token.GrantedProjects = grantedProjects.String
		}
		if expiresAt.Valid {
			token.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			token.LastUsedAt = &lastUsedAt.Time
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Update updates a token's properties (name, granted_projects, expires_at, is_active)
func (r *MCPTokenRepository) Update(token *MCPToken) error {
	token.UpdatedAt = time.Now()

	var expiresAt interface{}
	if token.ExpiresAt != nil {
		expiresAt = token.ExpiresAt
	}

	_, err := r.db.Exec(`
		UPDATE mcp_tokens
		SET name = ?, granted_projects = ?, expires_at = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`, token.Name, token.GrantedProjects, expiresAt, token.IsActive, token.UpdatedAt, token.ID)

	return err
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *MCPTokenRepository) UpdateLastUsed(id string) error {
	_, err := r.db.Exec(`
		UPDATE mcp_tokens SET last_used_at = ? WHERE id = ?
	`, time.Now(), id)
	return err
}

// Delete deletes a token by ID
func (r *MCPTokenRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM mcp_tokens WHERE id = ?", id)
	return err
}

// GetGrantedProjectIDs parses the granted_projects JSON and returns the array of project IDs
// Returns nil for "*" (all projects), or the array of project IDs
func (token *MCPToken) GetGrantedProjectIDs() ([]string, bool, error) {
	if token.GrantedProjects == "*" {
		return nil, true, nil // true means all projects
	}

	if token.GrantedProjects == "" {
		return []string{}, false, nil
	}

	var projectIDs []string
	if err := json.Unmarshal([]byte(token.GrantedProjects), &projectIDs); err != nil {
		return nil, false, err
	}

	return projectIDs, false, nil
}

// HasAccessToProject checks if the token has access to a specific project
func (token *MCPToken) HasAccessToProject(projectID string) (bool, error) {
	projectIDs, allProjects, err := token.GetGrantedProjectIDs()
	if err != nil {
		return false, err
	}

	if allProjects {
		return true, nil
	}

	for _, id := range projectIDs {
		if id == projectID {
			return true, nil
		}
	}

	return false, nil
}

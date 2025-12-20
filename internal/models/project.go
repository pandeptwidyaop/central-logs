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

type Project struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	IconType        string           `json:"icon_type"`  // "initials", "icon", or "image"
	IconValue       string           `json:"icon_value"` // initials text, icon name, or base64 image
	APIKey          string           `json:"-"`
	APIKeyPrefix    string           `json:"api_key_prefix"`
	IsActive        bool             `json:"is_active"`
	RetentionConfig *RetentionConfig `json:"retention_config,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type RetentionConfig struct {
	MaxAge   string                       `json:"max_age,omitempty"`
	MaxCount int                          `json:"max_count,omitempty"`
	Levels   map[string]LevelRetention    `json:"levels,omitempty"`
}

type LevelRetention struct {
	MaxAge   string `json:"max_age,omitempty"`
	MaxCount int    `json:"max_count,omitempty"`
}

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func GenerateAPIKey() (key, hash, prefix string, err error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}

	// Create the full key with prefix
	key = "cl_" + hex.EncodeToString(bytes)
	prefix = key[:10] + "..."

	// Hash the key for storage
	hashBytes := sha256.Sum256([]byte(key))
	hash = hex.EncodeToString(hashBytes[:])

	return key, hash, prefix, nil
}

func HashAPIKey(key string) string {
	hashBytes := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hashBytes[:])
}

func (r *ProjectRepository) Create(project *Project) (string, error) {
	project.ID = uuid.New().String()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	// Generate API key
	apiKey, apiKeyHash, apiKeyPrefix, err := GenerateAPIKey()
	if err != nil {
		return "", err
	}
	project.APIKey = apiKeyHash
	project.APIKeyPrefix = apiKeyPrefix

	var retentionJSON *string
	if project.RetentionConfig != nil {
		data, err := json.Marshal(project.RetentionConfig)
		if err != nil {
			return "", err
		}
		s := string(data)
		retentionJSON = &s
	}

	_, err = r.db.Exec(`
		INSERT INTO projects (id, name, description, icon_type, icon_value, api_key, api_key_prefix, is_active, retention_config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, project.ID, project.Name, project.Description, project.IconType, project.IconValue, project.APIKey, project.APIKeyPrefix, project.IsActive, retentionJSON, project.CreatedAt, project.UpdatedAt)

	if err != nil {
		return "", err
	}

	return apiKey, nil // Return the raw API key (only time it's available)
}

func (r *ProjectRepository) GetByID(id string) (*Project, error) {
	project := &Project{}
	var retentionJSON sql.NullString
	var description sql.NullString
	var iconType sql.NullString
	var iconValue sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, description, icon_type, icon_value, api_key, api_key_prefix, is_active, retention_config, created_at, updated_at
		FROM projects WHERE id = ?
	`, id).Scan(&project.ID, &project.Name, &description, &iconType, &iconValue, &project.APIKey, &project.APIKeyPrefix, &project.IsActive, &retentionJSON, &project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if description.Valid {
		project.Description = description.String
	}
	if iconType.Valid {
		project.IconType = iconType.String
	}
	if iconValue.Valid {
		project.IconValue = iconValue.String
	}

	if retentionJSON.Valid {
		if err := json.Unmarshal([]byte(retentionJSON.String), &project.RetentionConfig); err != nil {
			return nil, err
		}
	}

	return project, nil
}

func (r *ProjectRepository) GetByAPIKey(apiKey string) (*Project, error) {
	hashedKey := HashAPIKey(apiKey)

	project := &Project{}
	var retentionJSON sql.NullString
	var description sql.NullString
	var iconType sql.NullString
	var iconValue sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, description, icon_type, icon_value, api_key, api_key_prefix, is_active, retention_config, created_at, updated_at
		FROM projects WHERE api_key = ? AND is_active = 1
	`, hashedKey).Scan(&project.ID, &project.Name, &description, &iconType, &iconValue, &project.APIKey, &project.APIKeyPrefix, &project.IsActive, &retentionJSON, &project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if description.Valid {
		project.Description = description.String
	}
	if iconType.Valid {
		project.IconType = iconType.String
	}
	if iconValue.Valid {
		project.IconValue = iconValue.String
	}

	if retentionJSON.Valid {
		if err := json.Unmarshal([]byte(retentionJSON.String), &project.RetentionConfig); err != nil {
			return nil, err
		}
	}

	// Additional constant-time verification to prevent timing attacks
	if !utils.SecureCompareHash(project.APIKey, hashedKey) {
		return nil, nil // Hash mismatch
	}

	return project, nil
}

func (r *ProjectRepository) GetAll() ([]*Project, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, icon_type, icon_value, api_key, api_key_prefix, is_active, retention_config, created_at, updated_at
		FROM projects ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		var retentionJSON sql.NullString
		var description sql.NullString
		var iconType sql.NullString
		var iconValue sql.NullString

		if err := rows.Scan(&project.ID, &project.Name, &description, &iconType, &iconValue, &project.APIKey, &project.APIKeyPrefix, &project.IsActive, &retentionJSON, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}

		if description.Valid {
			project.Description = description.String
		}
		if iconType.Valid {
			project.IconType = iconType.String
		}
		if iconValue.Valid {
			project.IconValue = iconValue.String
		}

		if retentionJSON.Valid {
			if err := json.Unmarshal([]byte(retentionJSON.String), &project.RetentionConfig); err != nil {
				return nil, err
			}
		}

		projects = append(projects, project)
	}
	return projects, nil
}

func (r *ProjectRepository) GetByUserID(userID string) ([]*Project, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.name, p.description, p.icon_type, p.icon_value, p.api_key, p.api_key_prefix, p.is_active, p.retention_config, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN user_projects up ON p.id = up.project_id
		WHERE up.user_id = ?
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		var retentionJSON sql.NullString
		var description sql.NullString
		var iconType sql.NullString
		var iconValue sql.NullString

		if err := rows.Scan(&project.ID, &project.Name, &description, &iconType, &iconValue, &project.APIKey, &project.APIKeyPrefix, &project.IsActive, &retentionJSON, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}

		if description.Valid {
			project.Description = description.String
		}
		if iconType.Valid {
			project.IconType = iconType.String
		}
		if iconValue.Valid {
			project.IconValue = iconValue.String
		}

		if retentionJSON.Valid {
			if err := json.Unmarshal([]byte(retentionJSON.String), &project.RetentionConfig); err != nil {
				return nil, err
			}
		}

		projects = append(projects, project)
	}
	return projects, nil
}

func (r *ProjectRepository) Update(project *Project) error {
	project.UpdatedAt = time.Now()

	var retentionJSON *string
	if project.RetentionConfig != nil {
		data, err := json.Marshal(project.RetentionConfig)
		if err != nil {
			return err
		}
		s := string(data)
		retentionJSON = &s
	}

	_, err := r.db.Exec(`
		UPDATE projects SET name = ?, description = ?, icon_type = ?, icon_value = ?, is_active = ?, retention_config = ?, updated_at = ?
		WHERE id = ?
	`, project.Name, project.Description, project.IconType, project.IconValue, project.IsActive, retentionJSON, project.UpdatedAt, project.ID)
	return err
}

func (r *ProjectRepository) RotateAPIKey(id string) (string, error) {
	apiKey, apiKeyHash, apiKeyPrefix, err := GenerateAPIKey()
	if err != nil {
		return "", err
	}

	_, err = r.db.Exec(`
		UPDATE projects SET api_key = ?, api_key_prefix = ?, updated_at = ?
		WHERE id = ?
	`, apiKeyHash, apiKeyPrefix, time.Now(), id)

	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func (r *ProjectRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	return err
}

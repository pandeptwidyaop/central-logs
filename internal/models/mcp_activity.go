package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MCPActivityLog struct {
	ID            string     `json:"id"`
	TokenID       string     `json:"token_id"`
	ToolName      string     `json:"tool_name"`
	ProjectIDs    string     `json:"project_ids"`    // JSON array of accessed project IDs
	RequestParams string     `json:"request_params"` // JSON of request parameters (sanitized)
	Success       bool       `json:"success"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	DurationMS    int        `json:"duration_ms"`
	CreatedAt     time.Time  `json:"created_at"`
}

type MCPActivityLogRepository struct {
	db *sql.DB
}

func NewMCPActivityLogRepository(db *sql.DB) *MCPActivityLogRepository {
	return &MCPActivityLogRepository{db: db}
}

// Create logs a new MCP activity
func (r *MCPActivityLogRepository) Create(log *MCPActivityLog) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO mcp_activity_logs (id, token_id, tool_name, project_ids, request_params, success, error_message, duration_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, log.ID, log.TokenID, log.ToolName, log.ProjectIDs, log.RequestParams, log.Success, log.ErrorMessage, log.DurationMS, log.CreatedAt)

	return err
}

// GetByTokenID retrieves activity logs for a specific token
func (r *MCPActivityLogRepository) GetByTokenID(tokenID string, limit, offset int) ([]*MCPActivityLog, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM mcp_activity_logs WHERE token_id = ?", tokenID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get logs with pagination
	rows, err := r.db.Query(`
		SELECT id, token_id, tool_name, project_ids, request_params, success, error_message, duration_ms, created_at
		FROM mcp_activity_logs
		WHERE token_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tokenID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*MCPActivityLog
	for rows.Next() {
		log := &MCPActivityLog{}
		var projectIDs sql.NullString
		var requestParams sql.NullString
		var errorMessage sql.NullString

		err := rows.Scan(&log.ID, &log.TokenID, &log.ToolName, &projectIDs, &requestParams, &log.Success, &errorMessage, &log.DurationMS, &log.CreatedAt)
		if err != nil {
			return nil, 0, err
		}

		if projectIDs.Valid {
			log.ProjectIDs = projectIDs.String
		}
		if requestParams.Valid {
			log.RequestParams = requestParams.String
		}
		if errorMessage.Valid {
			log.ErrorMessage = errorMessage.String
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetRecent retrieves the most recent activity logs across all tokens
func (r *MCPActivityLogRepository) GetRecent(limit int) ([]*MCPActivityLog, error) {
	rows, err := r.db.Query(`
		SELECT id, token_id, tool_name, project_ids, request_params, success, error_message, duration_ms, created_at
		FROM mcp_activity_logs
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*MCPActivityLog
	for rows.Next() {
		log := &MCPActivityLog{}
		var projectIDs sql.NullString
		var requestParams sql.NullString
		var errorMessage sql.NullString

		err := rows.Scan(&log.ID, &log.TokenID, &log.ToolName, &projectIDs, &requestParams, &log.Success, &errorMessage, &log.DurationMS, &log.CreatedAt)
		if err != nil {
			return nil, err
		}

		if projectIDs.Valid {
			log.ProjectIDs = projectIDs.String
		}
		if requestParams.Valid {
			log.RequestParams = requestParams.String
		}
		if errorMessage.Valid {
			log.ErrorMessage = errorMessage.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// DeleteOlderThan deletes activity logs older than the specified date
func (r *MCPActivityLogRepository) DeleteOlderThan(before time.Time) (int64, error) {
	result, err := r.db.Exec("DELETE FROM mcp_activity_logs WHERE created_at < ?", before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetProjectIDsFromLog parses the project_ids JSON array from the log
func (log *MCPActivityLog) GetProjectIDsFromLog() ([]string, error) {
	if log.ProjectIDs == "" {
		return []string{}, nil
	}

	var projectIDs []string
	if err := json.Unmarshal([]byte(log.ProjectIDs), &projectIDs); err != nil {
		return nil, err
	}

	return projectIDs, nil
}

// GetRequestParamsMap parses the request_params JSON into a map
func (log *MCPActivityLog) GetRequestParamsMap() (map[string]interface{}, error) {
	if log.RequestParams == "" {
		return map[string]interface{}{}, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(log.RequestParams), &params); err != nil {
		return nil, err
	}

	return params, nil
}

package mcp

import (
	"central-logs/internal/models"
)

// Tool 1: query_logs - Search and filter logs with advanced criteria
type QueryLogsInput struct {
	ProjectIDs []string `json:"project_ids,omitempty"`
	Levels     []string `json:"levels,omitempty"`
	Source     string   `json:"source,omitempty"`
	Search     string   `json:"search,omitempty"`
	StartTime  string   `json:"start_time,omitempty"` // RFC3339 format
	EndTime    string   `json:"end_time,omitempty"`   // RFC3339 format
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
}

type QueryLogsOutput struct {
	Logs   []*models.Log `json:"logs"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// Tool 2: get_log - Retrieve single log by ID
type GetLogInput struct {
	LogID string `json:"log_id"`
}

type GetLogOutput struct {
	Log *models.Log `json:"log"`
}

// Tool 3: list_projects - List accessible projects
type ListProjectsInput struct {
	// No parameters - returns projects based on token's access
}

type ListProjectsOutput struct {
	Projects []*models.Project `json:"projects"`
	Count    int               `json:"count"`
}

// Tool 4: get_project - Get project details with statistics
type GetProjectInput struct {
	ProjectID string `json:"project_id"`
}

type GetProjectOutput struct {
	Project      *models.Project `json:"project"`
	TotalLogs    int             `json:"total_logs"`
	LogsByLevel  map[string]int  `json:"logs_by_level"`
}

// Tool 5: get_stats - System-wide or project-specific statistics
type GetStatsInput struct {
	Scope     string `json:"scope"`      // "overview" or "project"
	ProjectID string `json:"project_id,omitempty"` // Required if scope is "project"
}

type ProjectStatsSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LogCount int    `json:"log_count"`
	IsActive bool   `json:"is_active"`
}

type GetStatsOutput struct {
	// Overview scope fields
	TotalProjects int                    `json:"total_projects,omitempty"`
	TotalLogs     int                    `json:"total_logs,omitempty"`
	LogsToday     int                    `json:"logs_today,omitempty"`
	TotalUsers    int                    `json:"total_users,omitempty"`
	Projects      []ProjectStatsSummary  `json:"projects,omitempty"`

	// Common fields
	LogsByLevel   map[string]int         `json:"logs_by_level"`
	RecentLogs    []*models.Log          `json:"recent_logs,omitempty"`
}

// Tool 6: search_logs - Full-text search wrapper
type SearchLogsInput struct {
	Query      string   `json:"query"`
	ProjectIDs []string `json:"project_ids,omitempty"`
	Levels     []string `json:"levels,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

// Output: QueryLogsOutput

// Tool 7: get_recent_logs - Quick access to recent logs
type GetRecentLogsInput struct {
	ProjectIDs []string `json:"project_ids,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

type GetRecentLogsOutput struct {
	Logs []*models.Log `json:"logs"`
}

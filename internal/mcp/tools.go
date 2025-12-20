package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"central-logs/internal/models"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleGetLog retrieves a single log by ID
func (s *MCPServer) handleGetLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract token from context
	token, ok := ctx.Value("mcp_token").(*models.MCPToken)
	if !ok {
		return mcp.NewToolResultError("Authentication error"), nil
	}

	// Get log_id parameter
	logID, err := request.RequireString("log_id")
	if err != nil {
		s.logToolActivity(token, "get_log", nil, nil, false, fmt.Sprintf("Invalid input: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Retrieve log
	log, err := s.logRepo.GetByID(logID)
	if err != nil {
		s.logToolActivity(token, "get_log", nil, nil, false, fmt.Sprintf("Failed to retrieve log: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve log: %v", err)), nil
	}

	if log == nil {
		s.logToolActivity(token, "get_log", nil, nil, false, "Log not found", startTime)
		return mcp.NewToolResultError("Log not found"), nil
	}

	// Check if token has access to this log's project
	hasAccess, err := token.HasAccessToProject(log.ProjectID)
	if err != nil {
		s.logToolActivity(token, "get_log", nil, nil, false, fmt.Sprintf("Access check failed: %v", err), startTime)
		return mcp.NewToolResultError("Access check failed"), nil
	}

	if !hasAccess {
		s.logToolActivity(token, "get_log", []string{log.ProjectID}, nil, false, "Access denied to this log's project", startTime)
		return mcp.NewToolResultError("Access denied to this log's project"), nil
	}

	// Convert to JSON output
	output := &GetLogOutput{
		Log: log,
	}

	result, err := mcp.NewToolResultJSON(output)
	if err != nil {
		s.logToolActivity(token, "get_log", []string{log.ProjectID}, nil, false, fmt.Sprintf("Failed to serialize result: %v", err), startTime)
		return mcp.NewToolResultError("Failed to serialize result"), nil
	}

	// Log success
	s.logToolActivity(token, "get_log", []string{log.ProjectID}, map[string]interface{}{"log_id": logID}, true, "", startTime)

	return result, nil
}

// handleListProjects lists all projects accessible by the token
func (s *MCPServer) handleListProjects(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract token from context
	token, ok := ctx.Value("mcp_token").(*models.MCPToken)
	if !ok {
		return mcp.NewToolResultError("Authentication error"), nil
	}

	// Get granted project IDs
	grantedProjects, allProjects, err := token.GetGrantedProjectIDs()
	if err != nil {
		s.logToolActivity(token, "list_projects", nil, nil, false, fmt.Sprintf("Failed to get granted projects: %v", err), startTime)
		return mcp.NewToolResultError("Failed to get granted projects"), nil
	}

	var projects []*models.Project

	if allProjects {
		// Get all projects
		projects, err = s.projectRepo.GetAll()
		if err != nil {
			s.logToolActivity(token, "list_projects", nil, nil, false, fmt.Sprintf("Failed to list projects: %v", err), startTime)
			return mcp.NewToolResultError("Failed to list projects"), nil
		}
	} else {
		// Get only granted projects
		for _, projectID := range grantedProjects {
			project, err := s.projectRepo.GetByID(projectID)
			if err != nil {
				continue // Skip projects that can't be retrieved
			}
			if project != nil {
				projects = append(projects, project)
			}
		}
	}

	// Convert to output format
	output := &ListProjectsOutput{
		Projects: projects,
		Count:    len(projects),
	}

	result, err := mcp.NewToolResultJSON(output)
	if err != nil {
		s.logToolActivity(token, "list_projects", nil, nil, false, "Failed to serialize result", startTime)
		return mcp.NewToolResultError("Failed to serialize result"), nil
	}

	// Log success
	s.logToolActivity(token, "list_projects", nil, nil, true, "", startTime)

	return result, nil
}

// handleGetProject gets detailed information about a specific project
func (s *MCPServer) handleGetProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract token from context
	token, ok := ctx.Value("mcp_token").(*models.MCPToken)
	if !ok {
		return mcp.NewToolResultError("Authentication error"), nil
	}

	// Get project_id parameter
	projectID, err := request.RequireString("project_id")
	if err != nil {
		s.logToolActivity(token, "get_project", nil, nil, false, fmt.Sprintf("Invalid input: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Check if token has access to this project
	hasAccess, err := token.HasAccessToProject(projectID)
	if err != nil {
		s.logToolActivity(token, "get_project", nil, nil, false, fmt.Sprintf("Access check failed: %v", err), startTime)
		return mcp.NewToolResultError("Access check failed"), nil
	}

	if !hasAccess {
		s.logToolActivity(token, "get_project", []string{projectID}, nil, false, "Access denied to this project", startTime)
		return mcp.NewToolResultError("Access denied to this project"), nil
	}

	// Get project
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		s.logToolActivity(token, "get_project", []string{projectID}, nil, false, fmt.Sprintf("Failed to get project: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project: %v", err)), nil
	}

	if project == nil {
		s.logToolActivity(token, "get_project", []string{projectID}, nil, false, "Project not found", startTime)
		return mcp.NewToolResultError("Project not found"), nil
	}

	// Get log statistics for this project
	totalLogs, err := s.logRepo.CountByProject(projectID)
	if err != nil {
		totalLogs = 0 // Don't fail if stats can't be retrieved
	}

	// Get logs by level
	logsByLevel := make(map[string]int)
	for _, levelStr := range []string{"debug", "info", "warn", "error"} {
		level := models.LogLevel(levelStr)
		count, err := s.logRepo.CountByProjectAndLevel(projectID, level)
		if err == nil {
			logsByLevel[levelStr] = count
		}
	}

	// Convert to output format
	output := &GetProjectOutput{
		Project:     project,
		TotalLogs:   totalLogs,
		LogsByLevel: logsByLevel,
	}

	result, err := mcp.NewToolResultJSON(output)
	if err != nil {
		s.logToolActivity(token, "get_project", []string{projectID}, nil, false, "Failed to serialize result", startTime)
		return mcp.NewToolResultError("Failed to serialize result"), nil
	}

	// Log success
	args := map[string]interface{}{"project_id": projectID}
	s.logToolActivity(token, "get_project", []string{projectID}, args, true, "", startTime)

	return result, nil
}

// handleGetRecentLogs gets the most recent logs
func (s *MCPServer) handleGetRecentLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract token from context
	token, ok := ctx.Value("mcp_token").(*models.MCPToken)
	if !ok {
		return mcp.NewToolResultError("Authentication error"), nil
	}

	// Get optional parameters
	projectIDs := request.GetStringSlice("project_ids", nil)
	limit := request.GetInt("limit", 50)

	// Enforce max limit
	if limit > 500 {
		limit = 500
	}

	// Validate project access
	allowedProjects, err := ValidateProjectAccess(token, projectIDs)
	if err != nil {
		s.logToolActivity(token, "get_recent_logs", projectIDs, nil, false, fmt.Sprintf("Access denied: %v", err), startTime)
		return mcp.NewToolResultError("Access denied to requested projects"), nil
	}

	// Get recent logs
	logs, err := s.logRepo.GetRecent(allowedProjects, limit)
	if err != nil {
		s.logToolActivity(token, "get_recent_logs", allowedProjects, nil, false, fmt.Sprintf("Failed to retrieve logs: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve logs: %v", err)), nil
	}

	// Convert to output format
	output := &GetRecentLogsOutput{
		Logs: logs,
	}

	result, err := mcp.NewToolResultJSON(output)
	if err != nil {
		s.logToolActivity(token, "get_recent_logs", allowedProjects, nil, false, "Failed to serialize result", startTime)
		return mcp.NewToolResultError("Failed to serialize result"), nil
	}

	// Log success
	args := map[string]interface{}{"project_ids": projectIDs, "limit": limit}
	s.logToolActivity(token, "get_recent_logs", allowedProjects, args, true, "", startTime)

	return result, nil
}

// handleQueryLogs searches and filters logs with advanced criteria
func (s *MCPServer) handleQueryLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract token from context
	token, ok := ctx.Value("mcp_token").(*models.MCPToken)
	if !ok {
		return mcp.NewToolResultError("Authentication error"), nil
	}

	// Parse parameters
	projectIDs := request.GetStringSlice("project_ids", nil)
	levelStrs := request.GetStringSlice("levels", nil)
	source := request.GetString("source", "")
	search := request.GetString("search", "")
	startTimeStr := request.GetString("start_time", "")
	endTimeStr := request.GetString("end_time", "")
	limit := request.GetInt("limit", 100)
	offset := request.GetInt("offset", 0)

	// Enforce max limit
	if limit > 1000 {
		limit = 1000
	}

	// Validate project access
	allowedProjects, err := ValidateProjectAccess(token, projectIDs)
	if err != nil {
		s.logToolActivity(token, "query_logs", projectIDs, nil, false, fmt.Sprintf("Access denied: %v", err), startTime)
		return mcp.NewToolResultError("Access denied to requested projects"), nil
	}

	// Parse time parameters
	var startTime2, endTime2 *time.Time
	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			s.logToolActivity(token, "query_logs", allowedProjects, nil, false, fmt.Sprintf("Invalid start_time: %v", err), startTime)
			return mcp.NewToolResultError(fmt.Sprintf("Invalid start_time format: %v", err)), nil
		}
		startTime2 = &t
	}

	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			s.logToolActivity(token, "query_logs", allowedProjects, nil, false, fmt.Sprintf("Invalid end_time: %v", err), startTime)
			return mcp.NewToolResultError(fmt.Sprintf("Invalid end_time format: %v", err)), nil
		}
		endTime2 = &t
	}

	// Convert level strings to LogLevel type
	var levels []models.LogLevel
	for _, levelStr := range levelStrs {
		levels = append(levels, models.LogLevel(levelStr))
	}

	// Build filter
	filter := &models.LogFilter{
		ProjectIDs: allowedProjects,
		Levels:     levels,
		Source:     source,
		Search:     search,
		StartTime:  startTime2,
		EndTime:    endTime2,
		Limit:      limit,
		Offset:     offset,
	}

	// Query logs
	logs, total, err := s.logRepo.List(filter)
	if err != nil {
		s.logToolActivity(token, "query_logs", allowedProjects, nil, false, fmt.Sprintf("Failed to query logs: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to query logs: %v", err)), nil
	}

	// Convert to output format
	output := &QueryLogsOutput{
		Logs:   logs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	result, err := mcp.NewToolResultJSON(output)
	if err != nil {
		s.logToolActivity(token, "query_logs", allowedProjects, nil, false, "Failed to serialize result", startTime)
		return mcp.NewToolResultError("Failed to serialize result"), nil
	}

	// Log success
	args := map[string]interface{}{
		"project_ids": projectIDs,
		"levels":      levelStrs,
		"source":      source,
		"search":      search,
		"limit":       limit,
		"offset":      offset,
	}
	s.logToolActivity(token, "query_logs", allowedProjects, args, true, "", startTime)

	return result, nil
}

// handleSearchLogs performs full-text search across logs
func (s *MCPServer) handleSearchLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract token from context
	token, ok := ctx.Value("mcp_token").(*models.MCPToken)
	if !ok {
		return mcp.NewToolResultError("Authentication error"), nil
	}

	// Get required query parameter
	query, err := request.RequireString("query")
	if err != nil {
		s.logToolActivity(token, "search_logs", nil, nil, false, fmt.Sprintf("Invalid input: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Parse optional parameters
	projectIDs := request.GetStringSlice("project_ids", nil)
	levelStrs := request.GetStringSlice("levels", nil)
	limit := request.GetInt("limit", 100)

	// Enforce max limit
	if limit > 1000 {
		limit = 1000
	}

	// Validate project access
	allowedProjects, err := ValidateProjectAccess(token, projectIDs)
	if err != nil {
		s.logToolActivity(token, "search_logs", projectIDs, nil, false, fmt.Sprintf("Access denied: %v", err), startTime)
		return mcp.NewToolResultError("Access denied to requested projects"), nil
	}

	// Convert level strings to LogLevel type
	var levels []models.LogLevel
	for _, levelStr := range levelStrs {
		levels = append(levels, models.LogLevel(levelStr))
	}

	// Build filter (search is part of List functionality)
	filter := &models.LogFilter{
		ProjectIDs: allowedProjects,
		Levels:     levels,
		Search:     query,
		Limit:      limit,
		Offset:     0,
	}

	// Search logs
	logs, total, err := s.logRepo.List(filter)
	if err != nil {
		s.logToolActivity(token, "search_logs", allowedProjects, nil, false, fmt.Sprintf("Failed to search logs: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search logs: %v", err)), nil
	}

	// Convert to output format (reuse QueryLogsOutput)
	output := &QueryLogsOutput{
		Logs:   logs,
		Total:  total,
		Limit:  limit,
		Offset: 0,
	}

	result, err := mcp.NewToolResultJSON(output)
	if err != nil {
		s.logToolActivity(token, "search_logs", allowedProjects, nil, false, "Failed to serialize result", startTime)
		return mcp.NewToolResultError("Failed to serialize result"), nil
	}

	// Log success
	args := map[string]interface{}{
		"query":       query,
		"project_ids": projectIDs,
		"levels":      levelStrs,
		"limit":       limit,
	}
	s.logToolActivity(token, "search_logs", allowedProjects, args, true, "", startTime)

	return result, nil
}

// handleGetStats gets system-wide or project-specific statistics
func (s *MCPServer) handleGetStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract token from context
	token, ok := ctx.Value("mcp_token").(*models.MCPToken)
	if !ok {
		return mcp.NewToolResultError("Authentication error"), nil
	}

	// Get required scope parameter
	scope, err := request.RequireString("scope")
	if err != nil {
		s.logToolActivity(token, "get_stats", nil, nil, false, fmt.Sprintf("Invalid input: %v", err), startTime)
		return mcp.NewToolResultError(fmt.Sprintf("Invalid input: %v", err)), nil
	}

	// Validate scope
	if scope != "overview" && scope != "project" {
		s.logToolActivity(token, "get_stats", nil, nil, false, "Invalid scope value", startTime)
		return mcp.NewToolResultError("Scope must be 'overview' or 'project'"), nil
	}

	var output GetStatsOutput

	if scope == "overview" {
		// Get system-wide statistics
		grantedProjects, allProjects, err := token.GetGrantedProjectIDs()
		if err != nil {
			s.logToolActivity(token, "get_stats", nil, nil, false, fmt.Sprintf("Failed to get granted projects: %v", err), startTime)
			return mcp.NewToolResultError("Failed to get granted projects"), nil
		}

		var projects []*models.Project
		if allProjects {
			projects, err = s.projectRepo.GetAll()
			if err != nil {
				s.logToolActivity(token, "get_stats", nil, nil, false, fmt.Sprintf("Failed to get projects: %v", err), startTime)
				return mcp.NewToolResultError("Failed to get projects"), nil
			}
		} else {
			for _, projectID := range grantedProjects {
				project, err := s.projectRepo.GetByID(projectID)
				if err == nil && project != nil {
					projects = append(projects, project)
				}
			}
		}

		// Count total logs across accessible projects
		var projectIDs []string
		for _, p := range projects {
			projectIDs = append(projectIDs, p.ID)
		}

		// Calculate total logs by summing each project
		var totalLogs int
		for _, pid := range projectIDs {
			count, err := s.logRepo.CountByProject(pid)
			if err == nil {
				totalLogs += count
			}
		}

		// Count logs today
		logsToday, _ := s.logRepo.CountToday(projectIDs)

		// Get logs by level aggregated
		logsByLevel := make(map[string]int)
		for _, levelStr := range []string{"debug", "info", "warn", "error"} {
			level := models.LogLevel(levelStr)
			var count int
			for _, pid := range projectIDs {
				c, err := s.logRepo.CountByProjectAndLevel(pid, level)
				if err == nil {
					count += c
				}
			}
			logsByLevel[levelStr] = count
		}

		// Get recent logs
		recentLogs, _ := s.logRepo.GetRecent(projectIDs, 10)

		// Get total users (if token has access to all projects, assume admin)
		var totalUsers int
		if allProjects {
			totalUsers, _ = s.userRepo.Count()
		}

		// Build project summaries
		var projectSummaries []ProjectStatsSummary
		for _, p := range projects {
			logCount, _ := s.logRepo.CountByProject(p.ID)
			projectSummaries = append(projectSummaries, ProjectStatsSummary{
				ID:       p.ID,
				Name:     p.Name,
				LogCount: logCount,
				IsActive: p.IsActive,
			})
		}

		output = GetStatsOutput{
			TotalProjects: len(projects),
			TotalLogs:     totalLogs,
			LogsToday:     logsToday,
			TotalUsers:    totalUsers,
			Projects:      projectSummaries,
			LogsByLevel:   logsByLevel,
			RecentLogs:    recentLogs,
		}

		s.logToolActivity(token, "get_stats", nil, map[string]interface{}{"scope": "overview"}, true, "", startTime)

	} else {
		// Get project-specific statistics
		projectID, err := request.RequireString("project_id")
		if err != nil {
			s.logToolActivity(token, "get_stats", nil, nil, false, "project_id required when scope is 'project'", startTime)
			return mcp.NewToolResultError("project_id required when scope is 'project'"), nil
		}

		// Check access
		hasAccess, err := token.HasAccessToProject(projectID)
		if err != nil || !hasAccess {
			s.logToolActivity(token, "get_stats", []string{projectID}, nil, false, "Access denied to project", startTime)
			return mcp.NewToolResultError("Access denied to project"), nil
		}

		// Get project stats
		totalLogs, _ := s.logRepo.CountByProject(projectID)
		logsToday, _ := s.logRepo.CountToday([]string{projectID})

		logsByLevel := make(map[string]int)
		for _, levelStr := range []string{"debug", "info", "warn", "error"} {
			level := models.LogLevel(levelStr)
			count, err := s.logRepo.CountByProjectAndLevel(projectID, level)
			if err == nil {
				logsByLevel[levelStr] = count
			}
		}

		recentLogs, _ := s.logRepo.GetRecent([]string{projectID}, 10)

		output = GetStatsOutput{
			TotalLogs:   totalLogs,
			LogsToday:   logsToday,
			LogsByLevel: logsByLevel,
			RecentLogs:  recentLogs,
		}

		s.logToolActivity(token, "get_stats", []string{projectID}, map[string]interface{}{"scope": "project", "project_id": projectID}, true, "", startTime)
	}

	result, err := mcp.NewToolResultJSON(output)
	if err != nil {
		s.logToolActivity(token, "get_stats", nil, nil, false, "Failed to serialize result", startTime)
		return mcp.NewToolResultError("Failed to serialize result"), nil
	}

	return result, nil
}

// Helper function to serialize any data to JSON string
func toJSONString(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

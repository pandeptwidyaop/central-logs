package mcp

import (
	"encoding/json"
	"time"

	"central-logs/internal/models"
)

// LogActivity logs an MCP tool call to the activity log
func LogActivity(
	activityRepo *models.MCPActivityLogRepository,
	tokenID string,
	toolName string,
	projectIDs []string,
	requestParams map[string]interface{},
	success bool,
	errorMessage string,
	startTime time.Time,
) error {
	duration := time.Since(startTime).Milliseconds()

	// Serialize project IDs
	projectIDsJSON := SerializeProjectIDs(projectIDs)

	// Sanitize and serialize request params
	paramsJSON := SanitizeRequestParams(requestParams)

	// Create activity log
	log := &models.MCPActivityLog{
		TokenID:       tokenID,
		ToolName:      toolName,
		ProjectIDs:    projectIDsJSON,
		RequestParams: paramsJSON,
		Success:       success,
		ErrorMessage:  errorMessage,
		DurationMS:    int(duration),
	}

	return activityRepo.Create(log)
}

// ConvertParamsToMap converts MCP arguments to a map for logging
func ConvertParamsToMap(args interface{}) map[string]interface{} {
	if args == nil {
		return map[string]interface{}{}
	}

	// Try to convert to map directly
	if m, ok := args.(map[string]interface{}); ok {
		return m
	}

	// Try to marshal and unmarshal
	bytes, err := json.Marshal(args)
	if err != nil {
		return map[string]interface{}{"error": "failed to serialize params"}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return map[string]interface{}{"raw": string(bytes)}
	}

	return result
}

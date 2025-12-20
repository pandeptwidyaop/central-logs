package mcp

import (
	"encoding/json"
	"errors"
	"strings"

	"central-logs/internal/models"
)

var (
	ErrInvalidToken      = errors.New("invalid MCP token")
	ErrExpiredToken      = errors.New("MCP token has expired")
	ErrInactiveToken     = errors.New("MCP token is inactive")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrProjectAccess     = errors.New("access denied to project")
)

// ValidateMCPToken validates an MCP token from the Authorization header
// Returns the token if valid, or an error
func ValidateMCPToken(authHeader string, mcpTokenRepo *models.MCPTokenRepository) (*models.MCPToken, error) {
	if authHeader == "" {
		return nil, ErrUnauthorized
	}

	// Extract token from "Bearer mcp_..."
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, ErrUnauthorized
	}

	rawToken := parts[1]

	// Validate token prefix
	if !strings.HasPrefix(rawToken, "mcp_") {
		return nil, ErrInvalidToken
	}

	// Get token from database (also checks expiry and active status)
	token, err := mcpTokenRepo.GetByToken(rawToken)
	if err != nil {
		return nil, err
	}

	if token == nil {
		// Token not found, expired, or inactive
		return nil, ErrInvalidToken
	}

	return token, nil
}

// GetGrantedProjects parses the granted_projects JSON and returns project IDs
// Returns nil for "*" (all projects), or the array of project IDs
func GetGrantedProjects(token *models.MCPToken) ([]string, bool, error) {
	return token.GetGrantedProjectIDs()
}

// CheckProjectAccess checks if a token has access to a specific project
func CheckProjectAccess(token *models.MCPToken, projectID string) (bool, error) {
	return token.HasAccessToProject(projectID)
}

// IntersectProjectIDs returns the intersection of two project ID slices
// Used to filter requested projects against granted projects
func IntersectProjectIDs(requested []string, granted []string) []string {
	if len(requested) == 0 {
		return granted
	}

	grantedMap := make(map[string]bool)
	for _, id := range granted {
		grantedMap[id] = true
	}

	var result []string
	for _, id := range requested {
		if grantedMap[id] {
			result = append(result, id)
		}
	}

	return result
}

// ValidateProjectAccess validates that requested project IDs are within granted projects
// Returns the filtered list of allowed project IDs
func ValidateProjectAccess(token *models.MCPToken, requestedProjectIDs []string) ([]string, error) {
	grantedProjects, allProjects, err := token.GetGrantedProjectIDs()
	if err != nil {
		return nil, err
	}

	// If token has access to all projects
	if allProjects {
		if len(requestedProjectIDs) > 0 {
			return requestedProjectIDs, nil
		}
		return nil, nil // nil means all projects
	}

	// Token has limited project access
	if len(requestedProjectIDs) > 0 {
		// Filter requested projects against granted projects
		allowedProjects := IntersectProjectIDs(requestedProjectIDs, grantedProjects)
		if len(allowedProjects) == 0 {
			return nil, ErrProjectAccess
		}
		return allowedProjects, nil
	}

	// No specific projects requested, return all granted projects
	return grantedProjects, nil
}

// SerializeProjectIDs converts project IDs to JSON string for logging
func SerializeProjectIDs(projectIDs []string) string {
	if len(projectIDs) == 0 {
		return "[]"
	}

	bytes, err := json.Marshal(projectIDs)
	if err != nil {
		return "[]"
	}

	return string(bytes)
}

// SanitizeRequestParams sanitizes request parameters for logging
// Removes sensitive data and limits size
func SanitizeRequestParams(params map[string]interface{}) string {
	// Create a copy to avoid modifying original
	sanitized := make(map[string]interface{})

	for key, value := range params {
		// Skip potentially sensitive fields
		if key == "password" || key == "token" || key == "secret" {
			sanitized[key] = "[REDACTED]"
			continue
		}

		// Limit string length
		if str, ok := value.(string); ok {
			if len(str) > 100 {
				sanitized[key] = str[:100] + "..."
			} else {
				sanitized[key] = str
			}
			continue
		}

		sanitized[key] = value
	}

	bytes, err := json.Marshal(sanitized)
	if err != nil {
		return "{}"
	}

	// Limit total size
	result := string(bytes)
	if len(result) > 500 {
		result = result[:500] + "..."
	}

	return result
}

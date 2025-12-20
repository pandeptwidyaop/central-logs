package mcp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer wraps the MCP server and dependencies
type MCPServer struct {
	server          *server.MCPServer
	httpServer      *server.StreamableHTTPServer
	mcpTokenRepo    *models.MCPTokenRepository
	mcpActivityRepo *models.MCPActivityLogRepository
	logRepo         *models.LogRepository
	projectRepo     *models.ProjectRepository
	userRepo        *models.UserRepository
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(
	mcpTokenRepo *models.MCPTokenRepository,
	mcpActivityRepo *models.MCPActivityLogRepository,
	logRepo *models.LogRepository,
	projectRepo *models.ProjectRepository,
	userRepo *models.UserRepository,
) *MCPServer {
	mcpServer := &MCPServer{
		mcpTokenRepo:    mcpTokenRepo,
		mcpActivityRepo: mcpActivityRepo,
		logRepo:         logRepo,
		projectRepo:     projectRepo,
		userRepo:        userRepo,
	}

	// Create MCP server with server info
	s := server.NewMCPServer(
		"central-logs",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register all tools
	mcpServer.registerTools(s)

	mcpServer.server = s

	// Create HTTP server with stateless mode (tokens validated per request)
	mcpServer.httpServer = server.NewStreamableHTTPServer(
		s,
		server.WithEndpointPath("/api/mcp/message"),
		server.WithStateLess(true), // Stateless - authenticate each request
	)

	return mcpServer
}

// registerTools registers all MCP tools
func (s *MCPServer) registerTools(srv *server.MCPServer) {
	// Tool 1: query_logs - Search and filter logs
	queryLogsTool := mcp.NewTool("query_logs",
		mcp.WithDescription("Search and filter logs with advanced criteria including project IDs, levels, source, time range, and full-text search"),
		mcp.WithArray("project_ids",
			mcp.WithStringItems(
				mcp.Description("Project ID"),
			),
			mcp.Description("Filter by project IDs (optional)"),
		),
		mcp.WithArray("levels",
			mcp.WithStringItems(
				mcp.Enum("debug", "info", "warn", "error"),
			),
			mcp.Description("Filter by log levels: debug, info, warn, error (optional)"),
		),
		mcp.WithString("source",
			mcp.Description("Filter by log source (optional)"),
		),
		mcp.WithString("search",
			mcp.Description("Full-text search in message and metadata (optional)"),
		),
		mcp.WithString("start_time",
			mcp.Description("Start time in RFC3339 format (optional)"),
		),
		mcp.WithString("end_time",
			mcp.Description("End time in RFC3339 format (optional)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of logs to return (default: 100, max: 1000)"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Offset for pagination (default: 0)"),
		),
	)
	srv.AddTool(queryLogsTool, s.handleQueryLogs)

	// Tool 2: get_log - Retrieve single log by ID
	getLogTool := mcp.NewTool("get_log",
		mcp.WithDescription("Retrieve a single log entry by its ID"),
		mcp.WithString("log_id",
			mcp.Required(),
			mcp.Description("The log ID to retrieve"),
		),
	)
	srv.AddTool(getLogTool, s.handleGetLog)

	// Tool 3: list_projects - List accessible projects
	listProjectsTool := mcp.NewTool("list_projects",
		mcp.WithDescription("List all projects accessible by the MCP token"),
	)
	srv.AddTool(listProjectsTool, s.handleListProjects)

	// Tool 4: get_project - Get project details with statistics
	getProjectTool := mcp.NewTool("get_project",
		mcp.WithDescription("Get detailed information about a specific project including log statistics"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project ID to retrieve"),
		),
	)
	srv.AddTool(getProjectTool, s.handleGetProject)

	// Tool 5: get_stats - System-wide or project-specific statistics
	getStatsTool := mcp.NewTool("get_stats",
		mcp.WithDescription("Get system-wide statistics or project-specific statistics"),
		mcp.WithString("scope",
			mcp.Required(),
			mcp.Description("Scope of statistics: 'overview' for system-wide, 'project' for specific project"),
			mcp.Enum("overview", "project"),
		),
		mcp.WithString("project_id",
			mcp.Description("Required if scope is 'project'"),
		),
	)
	srv.AddTool(getStatsTool, s.handleGetStats)

	// Tool 6: search_logs - Full-text search wrapper
	searchLogsTool := mcp.NewTool("search_logs",
		mcp.WithDescription("Full-text search across log messages and metadata"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query string"),
		),
		mcp.WithArray("project_ids",
			mcp.WithStringItems(
				mcp.Description("Project ID"),
			),
			mcp.Description("Filter by project IDs (optional)"),
		),
		mcp.WithArray("levels",
			mcp.WithStringItems(
				mcp.Enum("debug", "info", "warn", "error"),
			),
			mcp.Description("Filter by log levels (optional)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of results to return (default: 100, max: 1000)"),
		),
	)
	srv.AddTool(searchLogsTool, s.handleSearchLogs)

	// Tool 7: get_recent_logs - Quick access to recent logs
	getRecentLogsTool := mcp.NewTool("get_recent_logs",
		mcp.WithDescription("Get the most recent logs, optionally filtered by projects"),
		mcp.WithArray("project_ids",
			mcp.WithStringItems(
				mcp.Description("Project ID"),
			),
			mcp.Description("Filter by project IDs (optional)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of logs to return (default: 50, max: 500)"),
		),
	)
	srv.AddTool(getRecentLogsTool, s.handleGetRecentLogs)
}

// HandleFiberRequest handles incoming Fiber HTTP requests for MCP
func (s *MCPServer) HandleFiberRequest(c *fiber.Ctx) error {
	// Validate MCP token from Authorization header
	authHeader := c.Get("Authorization")
	token, err := ValidateMCPToken(authHeader, s.mcpTokenRepo)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Invalid or missing MCP token",
		})
	}

	// Update last used timestamp (async)
	go s.mcpTokenRepo.UpdateLastUsed(token.ID)

	// Store token in request context for tool handlers
	ctx := context.WithValue(c.UserContext(), "mcp_token", token)
	c.SetUserContext(ctx)

	// Convert Fiber request to http.Request
	req := c.Context().Request
	httpReq, err := http.NewRequestWithContext(ctx, string(req.Header.Method()), string(req.URI().FullURI()), bytes.NewReader(req.Body()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process request",
		})
	}

	// Copy headers
	req.Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Set(string(key), string(value))
	})

	// Create response writer that captures output
	recorder := httptest.NewRecorder()

	// Handle the request through MCP HTTP server
	s.httpServer.ServeHTTP(recorder, httpReq)

	// Copy response headers
	for key, values := range recorder.Header() {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Set status and return body
	c.Status(recorder.Code)
	return c.Send(recorder.Body.Bytes())
}

// Tool handler implementations are in tools.go

// logToolActivity logs MCP tool activity
func (s *MCPServer) logToolActivity(
	token *models.MCPToken,
	toolName string,
	projectIDs []string,
	args map[string]interface{},
	success bool,
	errorMessage string,
	startTime time.Time,
) {
	// Log asynchronously to avoid blocking
	go func() {
		params := ConvertParamsToMap(args)
		_ = LogActivity(
			s.mcpActivityRepo,
			token.ID,
			toolName,
			projectIDs,
			params,
			success,
			errorMessage,
			startTime,
		)
	}()
}

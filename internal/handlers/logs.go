package handlers

import (
	"context"
	"strconv"
	"time"

	"central-logs/internal/middleware"
	"central-logs/internal/models"
	"central-logs/internal/queue"
	"central-logs/internal/services/notification"
	"central-logs/internal/websocket"

	"github.com/gofiber/fiber/v2"
)

type LogHandler struct {
	logRepo         *models.LogRepository
	channelRepo     *models.ChannelRepository
	userProjectRepo *models.UserProjectRepository
	redisClient     *queue.RedisClient
	pushService     *notification.PushService
	wsHub           *websocket.Hub
}

func NewLogHandler(
	logRepo *models.LogRepository,
	channelRepo *models.ChannelRepository,
	userProjectRepo *models.UserProjectRepository,
	redisClient *queue.RedisClient,
	pushService *notification.PushService,
	wsHub *websocket.Hub,
) *LogHandler {
	return &LogHandler{
		logRepo:         logRepo,
		channelRepo:     channelRepo,
		userProjectRepo: userProjectRepo,
		redisClient:     redisClient,
		pushService:     pushService,
		wsHub:           wsHub,
	}
}

type CreateLogRequest struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Source    string                 `json:"source,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
}

type CreateLogResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// CreateLog handles POST /api/v1/logs (public API with API key)
func (h *LogHandler) CreateLog(c *fiber.Ctx) error {
	project := middleware.GetProject(c)
	if project == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid API key",
		})
	}

	var req CreateLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Message is required",
		})
	}

	// Parse timestamp
	var timestamp time.Time
	if req.Timestamp != "" {
		var err error
		timestamp, err = time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			timestamp = time.Now()
		}
	} else {
		timestamp = time.Now()
	}

	log := &models.Log{
		ProjectID: project.ID,
		Level:     models.ParseLogLevel(req.Level),
		Message:   req.Message,
		Metadata:  req.Metadata,
		Source:    req.Source,
		Timestamp: timestamp,
	}

	if err := h.logRepo.Create(log); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create log",
		})
	}

	// Prepare log data for broadcasting
	logData := map[string]interface{}{
		"id":           log.ID,
		"project_id":   log.ProjectID,
		"project_name": project.Name,
		"level":        log.Level,
		"message":      log.Message,
		"metadata":     log.Metadata,
		"source":       log.Source,
		"created_at":   log.CreatedAt.Format(time.RFC3339),
	}

	// Broadcast to WebSocket clients
	if h.wsHub != nil {
		go h.wsHub.BroadcastLog(logData, project.ID)
	}

	// Publish to Redis for realtime streaming (if Redis is available)
	if h.redisClient != nil {
		go func() {
			ctx := context.Background()
			h.redisClient.PublishLog(ctx, project.ID, logData)
		}()
	}

	// Queue notifications for channels
	go h.queueNotifications(log, project)

	return c.Status(fiber.StatusCreated).JSON(CreateLogResponse{
		ID:     log.ID,
		Status: "received",
	})
}

type BatchLogRequest struct {
	Logs []CreateLogRequest `json:"logs"`
}

type BatchLogResponse struct {
	Received int      `json:"received"`
	IDs      []string `json:"ids"`
}

// CreateBatchLogs handles POST /api/v1/logs/batch
func (h *LogHandler) CreateBatchLogs(c *fiber.Ctx) error {
	project := middleware.GetProject(c)
	if project == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid API key",
		})
	}

	var req BatchLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Logs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No logs provided",
		})
	}

	if len(req.Logs) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Maximum 100 logs per batch",
		})
	}

	logs := make([]*models.Log, 0, len(req.Logs))
	for _, r := range req.Logs {
		if r.Message == "" {
			continue
		}

		var timestamp time.Time
		if r.Timestamp != "" {
			var err error
			timestamp, err = time.Parse(time.RFC3339, r.Timestamp)
			if err != nil {
				timestamp = time.Now()
			}
		} else {
			timestamp = time.Now()
		}

		logs = append(logs, &models.Log{
			ProjectID: project.ID,
			Level:     models.ParseLogLevel(r.Level),
			Message:   r.Message,
			Metadata:  r.Metadata,
			Source:    r.Source,
			Timestamp: timestamp,
		})
	}

	if err := h.logRepo.CreateBatch(logs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create logs",
		})
	}

	// Broadcast to WebSocket, publish to Redis and queue notifications
	go func() {
		ctx := context.Background()
		for _, log := range logs {
			logData := map[string]interface{}{
				"id":           log.ID,
				"project_id":   log.ProjectID,
				"project_name": project.Name,
				"level":        log.Level,
				"message":      log.Message,
				"metadata":     log.Metadata,
				"source":       log.Source,
				"created_at":   log.CreatedAt.Format(time.RFC3339),
			}

			// Broadcast to WebSocket clients
			if h.wsHub != nil {
				h.wsHub.BroadcastLog(logData, project.ID)
			}

			// Publish to Redis
			if h.redisClient != nil {
				h.redisClient.PublishLog(ctx, project.ID, logData)
			}

			h.queueNotifications(log, project)
		}
	}()

	ids := make([]string, len(logs))
	for i, log := range logs {
		ids[i] = log.ID
	}

	return c.Status(fiber.StatusCreated).JSON(BatchLogResponse{
		Received: len(logs),
		IDs:      ids,
	})
}

func (h *LogHandler) queueNotifications(log *models.Log, project *models.Project) {
	// Send push notifications to all devices
	// Service worker will check visibility and skip if page is visible (WebSocket toast handles it)
	if h.pushService != nil {
		go func() {
			if err := h.pushService.SendLogNotification(log, project.Name); err != nil {
				// Log error but don't fail the request
				_ = err
			}
		}()
	}

	// Queue other notifications via Redis (Telegram, Discord, etc.)
	if h.redisClient == nil {
		return // Redis not available, skip other notifications
	}

	channels, err := h.channelRepo.GetActiveByProjectID(project.ID)
	if err != nil {
		return
	}

	ctx := context.Background()
	for _, channel := range channels {
		// Skip PUSH channels as they're handled above
		if channel.Type == models.ChannelTypePush {
			continue
		}
		if channel.ShouldNotify(log.Level) {
			job := &queue.NotificationJob{
				LogID:     log.ID,
				ChannelID: channel.ID,
				ProjectID: project.ID,
				Level:     string(log.Level),
				Message:   log.Message,
				Source:    log.Source,
				Timestamp: log.Timestamp.Format(time.RFC3339),
			}
			h.redisClient.EnqueueNotification(ctx, job)
		}
	}
}

// ListLogs handles GET /api/admin/logs
func (h *LogHandler) ListLogs(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get user's project IDs
	var projectIDs []string
	if user.IsAdmin() {
		// Admin sees all projects, optionally filter by project_id query param
		if projectID := c.Query("project_id"); projectID != "" {
			projectIDs = []string{projectID}
		}
	} else {
		var err error
		projectIDs, err = h.userProjectRepo.GetUserProjectIDs(user.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get projects",
			})
		}
		// Filter by specific project if provided
		if projectID := c.Query("project_id"); projectID != "" {
			found := false
			for _, id := range projectIDs {
				if id == projectID {
					found = true
					break
				}
			}
			if found {
				projectIDs = []string{projectID}
			} else {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Access denied to this project",
				})
			}
		}
	}

	// Parse filters
	filter := &models.LogFilter{
		ProjectIDs: projectIDs,
	}

	if levels := c.Query("levels"); levels != "" {
		for _, l := range splitAndTrim(levels, ",") {
			filter.Levels = append(filter.Levels, models.ParseLogLevel(l))
		}
	}

	if source := c.Query("source"); source != "" {
		filter.Source = source
	}

	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	if start := c.Query("start_time"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartTime = &t
		}
	}

	if end := c.Query("end_time"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndTime = &t
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	logs, total, err := h.logRepo.List(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list logs",
		})
	}

	return c.JSON(fiber.Map{
		"logs":   logs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetLog handles GET /api/admin/logs/:id
func (h *LogHandler) GetLog(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	logID := c.Params("id")
	log, err := h.logRepo.GetByID(logID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get log",
		})
	}

	if log == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Log not found",
		})
	}

	// Check access
	if !user.IsAdmin() {
		hasAccess, _ := h.userProjectRepo.HasAccess(user.ID, log.ProjectID)
		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
	}

	return c.JSON(log)
}

func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := make([]string, 0)
	for _, p := range splitString(s, sep) {
		p = trimString(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

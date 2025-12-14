package handlers

import (
	"central-logs/internal/middleware"
	"central-logs/internal/models"

	"github.com/gofiber/fiber/v2"
)

type StatsHandler struct {
	logRepo         *models.LogRepository
	projectRepo     *models.ProjectRepository
	userProjectRepo *models.UserProjectRepository
	userRepo        *models.UserRepository
}

func NewStatsHandler(
	logRepo *models.LogRepository,
	projectRepo *models.ProjectRepository,
	userProjectRepo *models.UserProjectRepository,
	userRepo *models.UserRepository,
) *StatsHandler {
	return &StatsHandler{
		logRepo:         logRepo,
		projectRepo:     projectRepo,
		userProjectRepo: userProjectRepo,
		userRepo:        userRepo,
	}
}

// GetOverview handles GET /api/admin/stats/overview
func (h *StatsHandler) GetOverview(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get projects
	var projects []*models.Project
	var err error

	if user.IsAdmin() {
		projects, err = h.projectRepo.GetAll()
	} else {
		projects, err = h.projectRepo.GetByUserID(user.ID)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get projects",
		})
	}

	// Get project IDs for filtering
	projectIDs := make([]string, len(projects))
	for i, p := range projects {
		projectIDs[i] = p.ID
	}

	// Calculate stats
	totalLogs := 0
	logsByLevel := make(map[string]int)
	projectStats := make([]fiber.Map, 0)

	for _, project := range projects {
		count, _ := h.logRepo.CountByProject(project.ID)
		totalLogs += count

		stats, _ := h.logRepo.GetProjectStats(project.ID)
		for level, cnt := range stats {
			logsByLevel[level] += cnt
		}

		projectStats = append(projectStats, fiber.Map{
			"id":        project.ID,
			"name":      project.Name,
			"log_count": count,
			"is_active": project.IsActive,
		})
	}

	// Get logs today count
	var logsToday int
	if user.IsAdmin() {
		logsToday, _ = h.logRepo.CountToday(nil)
	} else {
		logsToday, _ = h.logRepo.CountToday(projectIDs)
	}

	// Get recent logs
	var recentLogs []*models.Log
	if user.IsAdmin() {
		recentLogs, _ = h.logRepo.GetRecent(nil, 10)
	} else {
		recentLogs, _ = h.logRepo.GetRecent(projectIDs, 10)
	}

	response := fiber.Map{
		"total_projects": len(projects),
		"total_logs":     totalLogs,
		"logs_today":     logsToday,
		"logs_by_level":  logsByLevel,
		"recent_logs":    recentLogs,
		"projects":       projectStats,
	}

	// Add user count for admins
	if user.IsAdmin() {
		userCount, _ := h.userRepo.Count()
		response["total_users"] = userCount
	}

	return c.JSON(response)
}

// GetProjectStats handles GET /api/admin/stats/projects/:id
func (h *StatsHandler) GetProjectStats(c *fiber.Ctx) error {
	user := middleware.GetUser(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	projectID := c.Params("id")

	// Check access
	if !user.IsAdmin() {
		hasAccess, _ := h.userProjectRepo.HasAccess(user.ID, projectID)
		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
	}

	project, err := h.projectRepo.GetByID(projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get project",
		})
	}

	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	totalLogs, _ := h.logRepo.CountByProject(projectID)
	logsByLevel, _ := h.logRepo.GetProjectStats(projectID)

	return c.JSON(fiber.Map{
		"project":       project,
		"total_logs":    totalLogs,
		"logs_by_level": logsByLevel,
	})
}

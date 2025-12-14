package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	githubRepo    = "pandeptwidyaop/central-logs"
	githubAPIURL  = "https://api.github.com/repos/%s/releases/latest"
	requestTimeout = 10 * time.Second
)

// GitHubRelease represents the GitHub API response for a release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

// UpdateCheckResponse represents the response for update check
type UpdateCheckResponse struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url,omitempty"`
	ReleaseNotes    string `json:"release_notes,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
}

// VersionHandler handles version-related requests
type VersionHandler struct {
	currentVersion string
}

// NewVersionHandler creates a new VersionHandler
func NewVersionHandler(currentVersion string) *VersionHandler {
	return &VersionHandler{
		currentVersion: currentVersion,
	}
}

// CheckUpdate checks for available updates from GitHub releases
func (h *VersionHandler) CheckUpdate(c *fiber.Ctx) error {
	// Fetch latest release from GitHub
	release, err := h.fetchLatestRelease()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to check for updates: %v", err),
		})
	}

	// Normalize versions for comparison (remove 'v' prefix)
	currentVer := normalizeVersion(h.currentVersion)
	latestVer := normalizeVersion(release.TagName)

	// Compare versions
	updateAvailable := compareVersions(currentVer, latestVer) < 0

	response := UpdateCheckResponse{
		CurrentVersion:  h.currentVersion,
		LatestVersion:   release.TagName,
		UpdateAvailable: updateAvailable,
	}

	if updateAvailable {
		response.ReleaseURL = release.HTMLURL
		response.ReleaseNotes = release.Body
		response.PublishedAt = release.PublishedAt.Format(time.RFC3339)
	}

	return c.JSON(response)
}

// fetchLatestRelease fetches the latest release info from GitHub
func (h *VersionHandler) fetchLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: requestTimeout,
	}

	url := fmt.Sprintf(githubAPIURL, githubRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Central-Logs-Update-Checker")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// normalizeVersion removes 'v' prefix and trims whitespace
func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	return version
}

// compareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad shorter version with zeros
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &num1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &num2)
		}

		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}

	return 0
}

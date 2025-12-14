package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type LogLevel string

const (
	LogLevelDebug    LogLevel = "DEBUG"
	LogLevelInfo     LogLevel = "INFO"
	LogLevelWarn     LogLevel = "WARN"
	LogLevelError    LogLevel = "ERROR"
	LogLevelCritical LogLevel = "CRITICAL"
)

func (l LogLevel) Priority() int {
	switch l {
	case LogLevelDebug:
		return 0
	case LogLevelInfo:
		return 1
	case LogLevelWarn:
		return 2
	case LogLevelError:
		return 3
	case LogLevelCritical:
		return 4
	default:
		return -1
	}
}

func ParseLogLevel(s string) LogLevel {
	switch s {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN", "WARNING":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	case "CRITICAL":
		return LogLevelCritical
	default:
		return LogLevelInfo
	}
}

type Log struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"project_id"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Source    string                 `json:"source,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	CreatedAt time.Time              `json:"created_at"`

	// Joined fields
	ProjectName string `json:"project_name,omitempty"`
}

type LogFilter struct {
	ProjectIDs []string
	Levels     []LogLevel
	Source     string
	Search     string
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
	Offset     int
}

type LogRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(log *Log) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()
	if log.Timestamp.IsZero() {
		log.Timestamp = log.CreatedAt
	}

	var metadataJSON *string
	if log.Metadata != nil {
		data, err := json.Marshal(log.Metadata)
		if err != nil {
			return err
		}
		s := string(data)
		metadataJSON = &s
	}

	_, err := r.db.Exec(`
		INSERT INTO logs (id, project_id, level, message, metadata, source, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, log.ID, log.ProjectID, log.Level, log.Message, metadataJSON, log.Source, log.Timestamp, log.CreatedAt)

	return err
}

func (r *LogRepository) CreateBatch(logs []*Log) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO logs (id, project_id, level, message, metadata, source, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, log := range logs {
		log.ID = uuid.New().String()
		log.CreatedAt = time.Now()
		if log.Timestamp.IsZero() {
			log.Timestamp = log.CreatedAt
		}

		var metadataJSON *string
		if log.Metadata != nil {
			data, err := json.Marshal(log.Metadata)
			if err != nil {
				return err
			}
			s := string(data)
			metadataJSON = &s
		}

		if _, err := stmt.Exec(log.ID, log.ProjectID, log.Level, log.Message, metadataJSON, log.Source, log.Timestamp, log.CreatedAt); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *LogRepository) GetByID(id string) (*Log, error) {
	log := &Log{}
	var metadataJSON sql.NullString
	var source sql.NullString

	err := r.db.QueryRow(`
		SELECT l.id, l.project_id, l.level, l.message, l.metadata, l.source, l.timestamp, l.created_at, p.name
		FROM logs l
		INNER JOIN projects p ON l.project_id = p.id
		WHERE l.id = ?
	`, id).Scan(&log.ID, &log.ProjectID, &log.Level, &log.Message, &metadataJSON, &source, &log.Timestamp, &log.CreatedAt, &log.ProjectName)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if source.Valid {
		log.Source = source.String
	}

	if metadataJSON.Valid {
		if err := json.Unmarshal([]byte(metadataJSON.String), &log.Metadata); err != nil {
			return nil, err
		}
	}

	return log, nil
}

func (r *LogRepository) List(filter *LogFilter) ([]*Log, int, error) {
	// Build WHERE clause
	where := "1=1"
	args := []interface{}{}

	if len(filter.ProjectIDs) > 0 {
		placeholders := ""
		for i, id := range filter.ProjectIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		where += " AND l.project_id IN (" + placeholders + ")"
	}

	if len(filter.Levels) > 0 {
		placeholders := ""
		for i, level := range filter.Levels {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, level)
		}
		where += " AND l.level IN (" + placeholders + ")"
	}

	if filter.Source != "" {
		where += " AND l.source = ?"
		args = append(args, filter.Source)
	}

	if filter.Search != "" {
		where += " AND l.message LIKE ?"
		args = append(args, "%"+filter.Search+"%")
	}

	if filter.StartTime != nil {
		where += " AND l.created_at >= ?"
		args = append(args, filter.StartTime)
	}

	if filter.EndTime != nil {
		where += " AND l.created_at <= ?"
		args = append(args, filter.EndTime)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM logs l WHERE " + where
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get logs
	query := `
		SELECT l.id, l.project_id, l.level, l.message, l.metadata, l.source, l.timestamp, l.created_at, p.name
		FROM logs l
		INNER JOIN projects p ON l.project_id = p.id
		WHERE ` + where + `
		ORDER BY l.created_at DESC
		LIMIT ? OFFSET ?
	`

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		log := &Log{}
		var metadataJSON sql.NullString
		var source sql.NullString

		if err := rows.Scan(&log.ID, &log.ProjectID, &log.Level, &log.Message, &metadataJSON, &source, &log.Timestamp, &log.CreatedAt, &log.ProjectName); err != nil {
			return nil, 0, err
		}

		if source.Valid {
			log.Source = source.String
		}

		if metadataJSON.Valid {
			if err := json.Unmarshal([]byte(metadataJSON.String), &log.Metadata); err != nil {
				return nil, 0, err
			}
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

func (r *LogRepository) DeleteOlderThan(projectID string, level LogLevel, before time.Time, batchSize int) (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM logs WHERE id IN (
			SELECT id FROM logs
			WHERE project_id = ? AND level = ? AND created_at < ?
			LIMIT ?
		)
	`, projectID, level, before, batchSize)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *LogRepository) DeleteExcessLogs(projectID string, level LogLevel, maxCount int, batchSize int) (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM logs WHERE id IN (
			SELECT id FROM logs
			WHERE project_id = ? AND level = ?
			ORDER BY created_at ASC
			LIMIT MAX(0, (SELECT COUNT(*) FROM logs WHERE project_id = ? AND level = ?) - ?)
		)
	`, projectID, level, projectID, level, maxCount)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *LogRepository) CountByProjectAndLevel(projectID string, level LogLevel) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM logs WHERE project_id = ? AND level = ?
	`, projectID, level).Scan(&count)
	return count, err
}

func (r *LogRepository) CountByProject(projectID string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM logs WHERE project_id = ?
	`, projectID).Scan(&count)
	return count, err
}

func (r *LogRepository) GetStats() (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT level, COUNT(*) FROM logs GROUP BY level
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		stats[level] = count
	}
	return stats, nil
}

func (r *LogRepository) GetProjectStats(projectID string) (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT level, COUNT(*) FROM logs WHERE project_id = ? GROUP BY level
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		stats[level] = count
	}
	return stats, nil
}

// CountToday returns the count of logs created today
func (r *LogRepository) CountToday(projectIDs []string) (int, error) {
	today := time.Now().Truncate(24 * time.Hour)

	var count int
	var err error

	if len(projectIDs) == 0 {
		err = r.db.QueryRow(`
			SELECT COUNT(*) FROM logs WHERE created_at >= ?
		`, today).Scan(&count)
	} else {
		placeholders := ""
		args := []interface{}{today}
		for i, id := range projectIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		err = r.db.QueryRow(`
			SELECT COUNT(*) FROM logs WHERE created_at >= ? AND project_id IN (`+placeholders+`)
		`, args...).Scan(&count)
	}

	return count, err
}

// GetRecent returns the most recent logs with project names
func (r *LogRepository) GetRecent(projectIDs []string, limit int) ([]*Log, error) {
	if limit <= 0 {
		limit = 10
	}

	var query string
	var args []interface{}

	if len(projectIDs) == 0 {
		query = `
			SELECT l.id, l.project_id, l.level, l.message, l.metadata, l.source, l.timestamp, l.created_at, p.name
			FROM logs l
			INNER JOIN projects p ON l.project_id = p.id
			ORDER BY l.created_at DESC
			LIMIT ?
		`
		args = []interface{}{limit}
	} else {
		placeholders := ""
		for i, id := range projectIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		query = `
			SELECT l.id, l.project_id, l.level, l.message, l.metadata, l.source, l.timestamp, l.created_at, p.name
			FROM logs l
			INNER JOIN projects p ON l.project_id = p.id
			WHERE l.project_id IN (` + placeholders + `)
			ORDER BY l.created_at DESC
			LIMIT ?
		`
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		log := &Log{}
		var metadataJSON sql.NullString
		var source sql.NullString

		if err := rows.Scan(&log.ID, &log.ProjectID, &log.Level, &log.Message, &metadataJSON, &source, &log.Timestamp, &log.CreatedAt, &log.ProjectName); err != nil {
			return nil, err
		}

		if source.Valid {
			log.Source = source.String
		}

		if metadataJSON.Valid {
			if err := json.Unmarshal([]byte(metadataJSON.String), &log.Metadata); err != nil {
				return nil, err
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}

package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PushSubscription struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProjectID *string   `json:"project_id,omitempty"`
	Endpoint  string    `json:"endpoint"`
	P256dh    string    `json:"p256dh"`
	Auth      string    `json:"auth"`
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type PushSubscriptionRepository struct {
	db *sql.DB
}

func NewPushSubscriptionRepository(db *sql.DB) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{db: db}
}

func (r *PushSubscriptionRepository) Create(sub *PushSubscription) error {
	sub.ID = uuid.New().String()
	sub.CreatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO push_subscriptions (id, user_id, project_id, endpoint, p256dh, auth, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, sub.ID, sub.UserID, sub.ProjectID, sub.Endpoint, sub.P256dh, sub.Auth, sub.UserAgent, sub.CreatedAt)

	return err
}

func (r *PushSubscriptionRepository) GetByEndpoint(endpoint string) (*PushSubscription, error) {
	sub := &PushSubscription{}
	var projectID sql.NullString
	var userAgent sql.NullString

	err := r.db.QueryRow(`
		SELECT id, user_id, project_id, endpoint, p256dh, auth, user_agent, created_at
		FROM push_subscriptions WHERE endpoint = ?
	`, endpoint).Scan(&sub.ID, &sub.UserID, &projectID, &sub.Endpoint, &sub.P256dh, &sub.Auth, &userAgent, &sub.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if projectID.Valid {
		sub.ProjectID = &projectID.String
	}
	if userAgent.Valid {
		sub.UserAgent = userAgent.String
	}

	return sub, nil
}

func (r *PushSubscriptionRepository) GetByUserID(userID string) ([]*PushSubscription, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, project_id, endpoint, p256dh, auth, user_agent, created_at
		FROM push_subscriptions WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*PushSubscription
	for rows.Next() {
		sub := &PushSubscription{}
		var projectID sql.NullString
		var userAgent sql.NullString

		if err := rows.Scan(&sub.ID, &sub.UserID, &projectID, &sub.Endpoint, &sub.P256dh, &sub.Auth, &userAgent, &sub.CreatedAt); err != nil {
			return nil, err
		}

		if projectID.Valid {
			sub.ProjectID = &projectID.String
		}
		if userAgent.Valid {
			sub.UserAgent = userAgent.String
		}

		subs = append(subs, sub)
	}
	return subs, nil
}

func (r *PushSubscriptionRepository) GetByProjectID(projectID string) ([]*PushSubscription, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, project_id, endpoint, p256dh, auth, user_agent, created_at
		FROM push_subscriptions WHERE project_id = ? OR project_id IS NULL
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*PushSubscription
	for rows.Next() {
		sub := &PushSubscription{}
		var pid sql.NullString
		var userAgent sql.NullString

		if err := rows.Scan(&sub.ID, &sub.UserID, &pid, &sub.Endpoint, &sub.P256dh, &sub.Auth, &userAgent, &sub.CreatedAt); err != nil {
			return nil, err
		}

		if pid.Valid {
			sub.ProjectID = &pid.String
		}
		if userAgent.Valid {
			sub.UserAgent = userAgent.String
		}

		subs = append(subs, sub)
	}
	return subs, nil
}

func (r *PushSubscriptionRepository) DeleteByEndpoint(endpoint string) error {
	_, err := r.db.Exec(`DELETE FROM push_subscriptions WHERE endpoint = ?`, endpoint)
	return err
}

func (r *PushSubscriptionRepository) DeleteByUserID(userID string) error {
	_, err := r.db.Exec(`DELETE FROM push_subscriptions WHERE user_id = ?`, userID)
	return err
}

// NotificationHistory

type NotificationStatus string

const (
	NotificationStatusPending     NotificationStatus = "PENDING"
	NotificationStatusSent        NotificationStatus = "SENT"
	NotificationStatusFailed      NotificationStatus = "FAILED"
	NotificationStatusRateLimited NotificationStatus = "RATE_LIMITED"
)

type NotificationHistory struct {
	ID           string             `json:"id"`
	LogID        string             `json:"log_id"`
	ChannelID    string             `json:"channel_id"`
	Status       NotificationStatus `json:"status"`
	ErrorMessage *string            `json:"error_message,omitempty"`
	SentAt       *time.Time         `json:"sent_at,omitempty"`
}

type NotificationHistoryRepository struct {
	db *sql.DB
}

func NewNotificationHistoryRepository(db *sql.DB) *NotificationHistoryRepository {
	return &NotificationHistoryRepository{db: db}
}

func (r *NotificationHistoryRepository) Create(nh *NotificationHistory) error {
	nh.ID = uuid.New().String()

	_, err := r.db.Exec(`
		INSERT INTO notification_history (id, log_id, channel_id, status, error_message, sent_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, nh.ID, nh.LogID, nh.ChannelID, nh.Status, nh.ErrorMessage, nh.SentAt)

	return err
}

func (r *NotificationHistoryRepository) UpdateStatus(id string, status NotificationStatus, errorMsg *string) error {
	now := time.Now()
	var sentAt *time.Time
	if status == NotificationStatusSent {
		sentAt = &now
	}

	_, err := r.db.Exec(`
		UPDATE notification_history SET status = ?, error_message = ?, sent_at = ?
		WHERE id = ?
	`, status, errorMsg, sentAt, id)
	return err
}

func (r *NotificationHistoryRepository) DeleteOlderThan(before time.Time, batchSize int) (int64, error) {
	result, err := r.db.Exec(`
		DELETE FROM notification_history WHERE id IN (
			SELECT id FROM notification_history WHERE sent_at < ? LIMIT ?
		)
	`, before, batchSize)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

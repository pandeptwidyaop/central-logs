package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ProjectRole string

const (
	ProjectRoleOwner  ProjectRole = "OWNER"
	ProjectRoleMember ProjectRole = "MEMBER"
	ProjectRoleViewer ProjectRole = "VIEWER"
)

type UserProject struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	ProjectID string      `json:"project_id"`
	Role      ProjectRole `json:"role"`
	CreatedAt time.Time   `json:"created_at"`

	// Joined fields
	User    *User    `json:"user,omitempty"`
	Project *Project `json:"project,omitempty"`
}

type UserProjectRepository struct {
	db *sql.DB
}

func NewUserProjectRepository(db *sql.DB) *UserProjectRepository {
	return &UserProjectRepository{db: db}
}

func (r *UserProjectRepository) Create(up *UserProject) error {
	up.ID = uuid.New().String()
	up.CreatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO user_projects (id, user_id, project_id, role, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, up.ID, up.UserID, up.ProjectID, up.Role, up.CreatedAt)

	return err
}

func (r *UserProjectRepository) GetByUserAndProject(userID, projectID string) (*UserProject, error) {
	up := &UserProject{}
	err := r.db.QueryRow(`
		SELECT id, user_id, project_id, role, created_at
		FROM user_projects WHERE user_id = ? AND project_id = ?
	`, userID, projectID).Scan(&up.ID, &up.UserID, &up.ProjectID, &up.Role, &up.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return up, nil
}

func (r *UserProjectRepository) GetProjectMembers(projectID string) ([]*UserProject, error) {
	rows, err := r.db.Query(`
		SELECT up.id, up.user_id, up.project_id, up.role, up.created_at,
		       u.id, u.email, u.name, u.role, u.is_active, u.created_at, u.updated_at
		FROM user_projects up
		INNER JOIN users u ON up.user_id = u.id
		WHERE up.project_id = ?
		ORDER BY up.created_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*UserProject
	for rows.Next() {
		up := &UserProject{User: &User{}}
		var userPassword string
		if err := rows.Scan(
			&up.ID, &up.UserID, &up.ProjectID, &up.Role, &up.CreatedAt,
			&up.User.ID, &up.User.Email, &up.User.Name, &up.User.Role, &up.User.IsActive, &up.User.CreatedAt, &up.User.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = userPassword // Password not needed
		members = append(members, up)
	}
	return members, nil
}

func (r *UserProjectRepository) GetUserProjects(userID string) ([]*UserProject, error) {
	rows, err := r.db.Query(`
		SELECT up.id, up.user_id, up.project_id, up.role, up.created_at
		FROM user_projects up
		WHERE up.user_id = ?
		ORDER BY up.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*UserProject
	for rows.Next() {
		up := &UserProject{}
		if err := rows.Scan(&up.ID, &up.UserID, &up.ProjectID, &up.Role, &up.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, up)
	}
	return projects, nil
}

func (r *UserProjectRepository) GetUserProjectIDs(userID string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT project_id FROM user_projects WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *UserProjectRepository) UpdateRole(userID, projectID string, role ProjectRole) error {
	_, err := r.db.Exec(`
		UPDATE user_projects SET role = ? WHERE user_id = ? AND project_id = ?
	`, role, userID, projectID)
	return err
}

func (r *UserProjectRepository) Delete(userID, projectID string) error {
	_, err := r.db.Exec(`
		DELETE FROM user_projects WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	return err
}

func (r *UserProjectRepository) HasAccess(userID, projectID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM user_projects WHERE user_id = ? AND project_id = ?
	`, userID, projectID).Scan(&count)
	return count > 0, err
}

func (r *UserProjectRepository) HasRole(userID, projectID string, roles ...ProjectRole) (bool, error) {
	if len(roles) == 0 {
		return false, nil
	}

	// Build query with role placeholders
	query := `SELECT COUNT(*) FROM user_projects WHERE user_id = ? AND project_id = ? AND role IN (`
	args := []interface{}{userID, projectID}
	for i, role := range roles {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, role)
	}
	query += ")"

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count > 0, err
}

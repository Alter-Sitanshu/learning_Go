package database

import (
	"context"
	"database/sql"
)

type Role struct {
	Level int    `json:"id"`
	Name  string `json:"name"`
}

type RoleStore struct {
	db *sql.DB
}

func (r *RoleStore) GetRole(ctx context.Context, role string) (*Role, error) {
	query := `
		SELECT roleid, name
		FROM roles
		WHERE name = $1
	`
	var role_DB Role
	err := r.db.QueryRowContext(ctx, query, role).Scan(
		&role_DB.Level,
		&role_DB.Name,
	)

	if err != nil {
		return nil, err
	}

	return &role_DB, nil
}

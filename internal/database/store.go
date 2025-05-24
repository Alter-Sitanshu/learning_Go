package database

import (
	"context"
	"database/sql"
)

type UserInterface interface {
	Create(context.Context, *User) error
	GetUserByID(context.Context, int64) (*User, error)
}

type PostInterface interface {
	Create(context.Context, *Post) error
	GetPostByID(context.Context, int64) (*Post, error)
}

type Storage interface {
	User() UserInterface
	Post() PostInterface
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) Storage {
	psql := PostgresRepo{
		db: db,
	}

	return &psql
}

func (psql *PostgresRepo) User() UserInterface {
	return &UserStore{db: psql.db}
}

func (psql *PostgresRepo) Post() PostInterface {
	return &PostStore{db: psql.db}
}

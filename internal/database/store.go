package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const QueryTimeOut = time.Minute * 3

var ErrNotFound = errors.New("not found")

type UserInterface interface {
	Create(context.Context, *User) error
	GetUserByID(context.Context, int64) (*User, error)
	Follow(context.Context, int64, int64) error
	Unfollow(context.Context, int64, int64) error
	GetFeed(context.Context, int64, *FilteringQuery) ([]Feed, error)
}

type PostInterface interface {
	Create(context.Context, *Post) error
	GetPostByID(context.Context, int64) (*Post, error)
	DeletePost(context.Context, int64) error
	UpdatePost(context.Context, *Post) error
}

type CommentInterface interface {
	GetComments(context.Context, int64) ([]Comment, error)
	CreateComment(context.Context, *Comment) error
}

type Storage interface {
	User() UserInterface
	Post() PostInterface
	Comment() CommentInterface
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

func (psql *PostgresRepo) Comment() CommentInterface {
	return &CommentStore{db: psql.db}
}

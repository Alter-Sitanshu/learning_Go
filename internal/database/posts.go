package database

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	UserID    int64    `json:"user"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type PostStore struct {
	db *sql.DB
}

func (p *PostStore) GetPostByID(id int) (*Post, error) {
	row := p.db.QueryRow("SELECT * FROM posts WHERE id=$1", id)

	var post Post
	err := row.Scan(&post.ID, &post.Title)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (p *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (title, content, userid, tags)
		VALUES($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`
	err := p.db.QueryRowContext(ctx, query,
		post.Title,
		post.Content,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

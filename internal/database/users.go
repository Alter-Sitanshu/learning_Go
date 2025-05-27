package database

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type User struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Password string `json:"-"`
	Email    string `json:"email,omitempty"`
	Age      int    `json:"age,omitempty"`
	Gender   byte   `json:"gender,omitempty"` // either 0(M) or 1(F)
}

type Feed struct {
	Post         Post `json:"post"`
	CommentCount int  `json:"comment_count"`
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, name, email, age, gender FROM users 
		WHERE id=$1 RETURNING id,name,email,age,gender
	`
	var user User
	err := u.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Age,
		&user.Gender,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserStore) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (name, password, email, age, gender)
		VALUES($1, $2, $3) RETURNING id
	`
	err := u.db.QueryRowContext(ctx, query,
		user.Name,
		user.Password,
		user.Email,
		user.Age,
		user.Gender,
	).Scan(
		&user.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserStore) Follow(ctx context.Context, targetID, userID int64) error {
	query := `
		INSERT INTO followers (userid, follower_id)
		VALUES ($1, $2)
	`
	_, err := u.db.ExecContext(ctx, query, targetID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) Unfollow(ctx context.Context, targetID, userID int64) error {
	query := `
		DELETE FROM followers
		WHERE userid = $1 AND follower_id = $2
	`
	_, err := u.db.ExecContext(ctx, query, targetID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) GetFeed(ctx context.Context, userID int64, fq *FilteringQuery) ([]Feed, error) {
	query := `
		SELECT p.id, p.userid, p.title, p.content, p.tags,
		COUNT(c.id) as comments_count, u.name, p.created_at
		FROM posts p
		LEFT JOIN comments c ON c.postid = p.id
		LEFT JOIN users u ON u.id = p.userid
		LEFT JOIN followers f ON f.userid = p.userid AND f.follower_id = $1
		WHERE f.userid IS NOT NULL AND
		(p.title ILIKE '%' || $2 || '%' OR p.content ILIKE '%' || $2 || '%') AND
		(p.tags @> $3 OR $3 = '{}')
		GROUP BY p.id, u.name
		ORDER BY p.created_at ` + fq.Sort + `
		LIMIT $4 OFFSET $5
	`
	rows, err := u.db.QueryContext(
		ctx, query, userID, fq.Search,
		pq.Array(fq.Tags), fq.Limit, fq.Offset,
	)
	if err != nil {
		return nil, err
	}
	var output []Feed
	for rows.Next() {
		var feed Feed
		rows.Scan(
			&feed.Post.ID,
			&feed.Post.UserID,
			&feed.Post.Title,
			&feed.Post.Content,
			pq.Array(&feed.Post.Tags),
			&feed.CommentCount,
			&feed.Post.User.Name,
			&feed.Post.CreatedAt,
		)
		output = append(output, feed)
	}
	return output, nil

}

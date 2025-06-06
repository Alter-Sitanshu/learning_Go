package database

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	Userid    int64  `json:"userid"`
	Postid    int64  `json:"postid"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (c *CommentStore) GetComments(ctx context.Context, postID int64) ([]Comment, error) {
	query := `
		SELECT c.id, c.content, c.userid, c.postid, users.id, users.name FROM comments c
		JOIN users on users.id = c.userid
		WHERE postid = $1
		ORDER BY created_at DESC
	`
	rows, err := c.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var output []Comment
	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(&c.ID, &c.Content, &c.Userid, &c.Postid, &c.User.ID, &c.User.Name)
		if err != nil {
			return nil, err
		}
		output = append(output, c)
	}
	return output, nil
}

func (c *CommentStore) CreateComment(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (content, userid, postid)
		VALUES($1, $2, $3) RETURNING id
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	err := c.db.QueryRowContext(ctx, query, comment.Content, comment.Userid, comment.Postid).Scan(
		&comment.ID,
	)

	if err != nil {
		return err
	}
	return nil

}

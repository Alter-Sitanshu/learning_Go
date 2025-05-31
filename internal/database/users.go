package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64    `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Password password `json:"-"`
	Email    string   `json:"email,omitempty"`
	Age      int      `json:"age,omitempty"`
	Gender   byte     `json:"gender,omitempty"` // either 0(M) or 1(F)
	Active   bool     `json:"is_active,omitempty"`
	Role     int      `json:"role,omitempty"`
}

type password struct {
	text *string
	hash []byte
}

type UserFromToken struct {
	id     int64
	name   string
	email  string
	active bool
}

func (pass *password) Hash(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	pass.text = &text
	pass.hash = hash

	return nil
}
func (pass *password) CheckPassword(plainPassword string) error {
	return bcrypt.CompareHashAndPassword(pass.hash, []byte(plainPassword))
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
		SELECT id, name, password, email, age, gender, role
		FROM users 
		WHERE id=$1 AND is_active = true
		RETURNING id,name,email,age,gender
	`
	var user User
	err := u.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Password.hash,
		&user.Email,
		&user.Age,
		&user.Gender,
		&user.Role,
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

func (u *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, name, password, email, age, gender, role
		FROM users 
		WHERE email=$1 AND is_active=true
	`
	var user User
	err := u.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Password.hash,
		&user.Email,
		&user.Age,
		&user.Gender,
		&user.Role,
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

func (u *UserStore) create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (name, password, email, age, gender, role)
		VALUES($1, $2, $3, $4, $5) RETURNING id
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	err := tx.QueryRowContext(ctx, query,
		user.Name,
		user.Password.hash,
		user.Email,
		user.Age,
		user.Gender,
		user.Role,
	).Scan(
		&user.ID,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_name_key"`:
			return ErrDupliName
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDupliMail
		}
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

func (u *UserStore) CreateAndInvite(ctx context.Context, user *User,
	token string, expiry time.Duration) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		err := u.create(ctx, tx, user)
		if err != nil {
			return err
		}
		err = createNewToken(ctx, tx, expiry, user.ID, token)
		if err != nil {
			return err
		}

		return nil
	})
}

func createNewToken(ctx context.Context, tx *sql.Tx,
	exp time.Duration, userid int64, token string) error {

	query := `
		INSERT INTO user_tokens (userid, token, expiry)
		VALUES($1, $2, $3)
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userid, token, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}
func (u *UserStore) authorise(ctx context.Context, tx *sql.Tx, token string,
	expiry time.Time) (*UserFromToken, error) {
	query := `
		SELECT u.id, u.name, u.email, u.is_active
		FROM users u
		JOIN user_tokens ut ON ut.userid = u.id
		WHERE ut.token = $1 AND ut.expiry > $2 AND u.is_active = false
	`
	user := &UserFromToken{}
	hashtoken := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(hashtoken[:])
	err := tx.QueryRowContext(ctx, query, hash, expiry).Scan(
		&user.id,
		&user.name,
		&user.email,
		&user.active,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserStore) ActivateUser(ctx context.Context, token string, expiry time.Time) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {

		ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
		defer cancel()

		user, err := u.authorise(ctx, tx, token, expiry)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrTokenExpired
			default:
				return err
			}
		}

		query := `
			UPDATE users
			SET is_active = true
			WHERE id = $1
		`
		_, err = tx.ExecContext(ctx, query, user.id)
		if err != nil {
			return err
		}

		query = `
			DELETE FROM user_tokens
			WHERE userid = $1
		`
		_, err = tx.ExecContext(ctx, query, user.id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStore) DeleteUser(ctx context.Context, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		// The role check will be handled in the middleware
		query := `
			DELETE FROM users
			WHERE id = $1
		`
		_, err := tx.ExecContext(ctx, query, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}

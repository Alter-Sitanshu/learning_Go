package database

import (
	"context"
	"database/sql"
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Password string `json:"-"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Gender   byte   `json:"gender"` // either 0(M) or 1(F)
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) GetUserByID(id int) (*User, error) {
	query := `
		SELECT (id, name, email, age, gender) FROM users 
		WHERE id=$1 RETURNING id,name,email,age,gender
	`
	var user User
	err := u.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Age,
		&user.Gender,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *UserStore) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (name, passw, email, age, gender)
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

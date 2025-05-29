package database

import (
	"context"
	"database/sql"
	"time"
)

func Mount(addr string, MaxConns, MaxIdleConn, MaxIdleTime int) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(time.Duration(MaxIdleTime) * time.Minute)
	db.SetMaxIdleConns(MaxIdleConn)
	db.SetMaxOpenConns(MaxConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

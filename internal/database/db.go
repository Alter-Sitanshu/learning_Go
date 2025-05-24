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

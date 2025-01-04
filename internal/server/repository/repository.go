package repository

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Repository interface {
	Close() error
	Ping() error
}

type repository struct {
	db *sql.DB
}

func NewRepository(dsn string) (Repository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &repository{db}, nil
}

func (r *repository) Close() error {
	return r.db.Close()
}

func (r *repository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return r.db.PingContext(ctx)
}

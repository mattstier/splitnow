package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect(connStr string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		return err
	}
	return Pool.Ping(context.Background())
}

func Close() {
	Pool.Close()
}

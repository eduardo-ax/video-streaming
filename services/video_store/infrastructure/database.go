package infrastructure

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool() *pgxpool.Pool {
	dbPool, err := pgxpool.New(context.Background(), "postgres://db_user:db_password@localhost:5432/video_store_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	return dbPool
}

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(pool *pgxpool.Pool) *Database {
	return &Database{
		pool: pool,
	}
}

func (db *Database) Close() {
	db.pool.Close()
}

func (db *Database) Persist(ctx context.Context, title string, description string) (int, error) {
	var id int
	err := db.pool.QueryRow(ctx, "INSERT INTO videos (title, description) VALUES ($1, $2) RETURNING id", title, description).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

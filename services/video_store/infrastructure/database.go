package infrastructure

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool() *pgxpool.Pool {
	dbURL := os.Getenv("VIDEOS_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Error database host")
	}
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("erro conected database %v", err)
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

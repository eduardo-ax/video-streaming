package infrastructure

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool() *pgxpool.Pool {
	dbURL := os.Getenv("USERS_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Error database host")
	}
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
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

func (db *Database) Persist(ctx context.Context, name string, email string, password string, plan int8) (int, error) {
	var id int
	err := db.pool.QueryRow(ctx, "INSERT INTO users (name,email,password,plan) VALUES ($1, $2, $3, $4) RETURNING id", name, email, password, plan).Scan(&id)

	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	return id, nil
}

package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/jackc/pgx/v5/pgconn"
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

func (db *Database) Persist(ctx context.Context, name string, email string, password string, plan int8) (string, error) {
	var id string
	err := db.pool.QueryRow(ctx, "INSERT INTO users (name,email,password,plan) VALUES ($1, $2, $3, $4) RETURNING id", name, email, password, plan).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return "", fmt.Errorf("email already exist")
			}
		}
		return "", err
	}
	return id, nil
}

func (db *Database) GetUser(ctx context.Context, email string) (*domain.UserAuthData, error) {

	user := &domain.UserAuthData{}
	err := db.pool.QueryRow(ctx, "SELECT id,password FROM users WHERE email = $1", email).Scan(&user.ID, &user.Password)

	if err != nil {
		return user, fmt.Errorf("user doesn't exist")
	}

	return user, nil

}

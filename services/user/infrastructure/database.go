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
	err := db.pool.QueryRow(ctx, "SELECT id,password,plan FROM users WHERE email = $1", email).Scan(&user.ID, &user.Password, &user.Plan)
	if err != nil {
		return user, fmt.Errorf("user doesn't exist")
	}

	return user, nil

}

func (db *Database) CreateSession(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	_, err := db.pool.Exec(ctx, "INSERT INTO sessions (id, email, refresh_token, is_revoked, expires_at) VALUES ($1,$2,$3,$4,$5)", session.ID, session.Email, session.RefreshToken, session.IsRevoked, session.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (db *Database) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	var s domain.Session
	err := db.pool.QueryRow(ctx, `SELECT id, email, refresh_token, is_revoked, expires_at FROM sessions WHERE id = $1`, id).Scan(&s.ID, &s.Email, &s.RefreshToken, &s.IsRevoked, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *Database) RevokeSession(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, "UPDATE sessions SET is_revoked = true WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error revoking session: %w", err)
	}
	return nil
}

func (db *Database) DeleteSession(ctx context.Context, id string) error {
	query, err := db.pool.Exec(ctx, "DELETE FROM sessions WHERE id=$1", id)

	if err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	if query.RowsAffected() == 0 {
		return fmt.Errorf("session doesn't exist")
	}

	return nil
}

package domain

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	Persist(ctx context.Context, name string, email string, password string, plan int8) (string, error)
	GetUser(ctx context.Context, email string) (*UserAuthData, error)
	CreateSession(ctx context.Context, session *Session) (*Session, error)
	GetSession(ctx context.Context, id string) (*Session, error)
	DeleteSession(ctx context.Context, id string) error
	RevokeSession(ctx context.Context, id string) error
}

type TokenInterface interface {
	CreateToken(id string, email string, plan int8, duration time.Duration) (string, *UserClaims, error)
	VerifyToken(tokenStr string) (*UserClaims, error)
}

type UserInterface interface {
	CreateUser(ctx context.Context, name string, email string, plan int8, pass string) error
	UserLogin(ctx context.Context, email string, password string) (*LoginUserRes, error)
	UserLogout(ctx context.Context, id string) error
	RenewAccessToken(ctx context.Context, refreshToken string) (*RenewAccessTokenRes, error)
	RevokeSession(ctx context.Context, id string) error
}

func NewUserManager(db Storage, token TokenInterface) *UserManager {
	return &UserManager{
		db:    db,
		token: token,
	}
}

func (u *UserManager) CreateUser(ctx context.Context, name string, email string, plan int8, password string) error {
	cryptPassword, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = u.db.Persist(ctx, name, email, cryptPassword, plan)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserManager) UserLogin(ctx context.Context, email string, password string) (*LoginUserRes, error) {
	user, err := u.db.GetUser(ctx, email)

	if err != nil {
		return nil, err
	}
	if !CheckPassword(password, user.Password) {
		return nil, fmt.Errorf("password incorrect")
	}

	acessToken, accessClaims, err := u.token.CreateToken(user.ID, email, user.Plan, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	refreshToken, refreshClaim, err := u.token.CreateToken(user.ID, email, user.Plan, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	session, err := u.db.CreateSession(ctx, &Session{
		ID:           refreshClaim.RegisteredClaims.ID,
		Email:        email,
		RefreshToken: refreshToken,
		IsRevoked:    false,
		ExpiresAt:    refreshClaim.RegisteredClaims.ExpiresAt.Time,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating session: %w", err)
	}
	return &LoginUserRes{
		SessionID:             session.ID,
		AccessToken:           acessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaim.RegisteredClaims.ExpiresAt.Time,
		User: UserPayload{
			ID:    user.ID,
			Email: email,
			Plan:  user.Plan,
		},
	}, nil
}

func (u *UserManager) UserLogout(ctx context.Context, id string) error {
	err := u.db.DeleteSession(ctx, id)
	if err != nil {
		return fmt.Errorf("logout error %w", err)
	}
	return nil
}

func (u *UserManager) RenewAccessToken(ctx context.Context, refreshToken string) (*RenewAccessTokenRes, error) {

	refreshClaims, err := u.token.VerifyToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %w", err)
	}

	session, err := u.db.GetSession(ctx, refreshClaims.RegisteredClaims.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting session: %w", err)
	}

	if session.IsRevoked {
		return nil, fmt.Errorf("session revoked: %w", err)
	}

	if session.Email != refreshClaims.Email {
		return nil, fmt.Errorf("invalid session")
	}

	acessToken, accessClaims, err := u.token.CreateToken(refreshClaims.ID, refreshClaims.Email, refreshClaims.Plan, 15*time.Minute)

	if err != nil {
		return nil, fmt.Errorf("error creating token: %w", err)
	}

	return &RenewAccessTokenRes{
		AccessToken:         acessToken,
		AcessTokenExpiresAt: accessClaims.RegisteredClaims.ExpiresAt.Time,
	}, nil
}

func (u *UserManager) RevokeSession(ctx context.Context, id string) error {
	err := u.db.RevokeSession(ctx, id)
	if err != nil {
		return fmt.Errorf("error revoking session %w", err)
	}
	return nil
}

func HashPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(pass, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

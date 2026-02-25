package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/thrgamon/project-template/internal/config"
	"github.com/thrgamon/project-template/internal/db"
	"github.com/thrgamon/project-template/internal/domain"
)

type Service struct {
	queries *db.Queries
	cfg     config.Config
}

func NewService(queries *db.Queries, cfg config.Config) *Service {
	return &Service{queries: queries, cfg: cfg}
}

func (s *Service) Register(ctx context.Context, email, password string) (*domain.AuthResponse, string, error) {
	_, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, "", errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, "", fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, "", fmt.Errorf("creating user: %w", err)
	}

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("creating session: %w", err)
	}

	return &domain.AuthResponse{
		User: domain.UserResponse{ID: user.ID, Email: user.Email},
	}, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*domain.AuthResponse, string, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("creating session: %w", err)
	}

	return &domain.AuthResponse{
		User: domain.UserResponse{ID: user.ID, Email: user.Email},
	}, token, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.queries.DeleteSessionByToken(ctx, token)
}

func (s *Service) ValidateSession(ctx context.Context, token string) (*db.GetSessionByTokenRow, error) {
	session, err := s.queries.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, errors.New("invalid session")
	}
	return &session, nil
}

func (s *Service) DeleteExpiredSessions(ctx context.Context) error {
	return s.queries.DeleteExpiredSessions(ctx)
}

func (s *Service) createSession(ctx context.Context, userID int32) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	_, err = s.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(s.cfg.SessionMaxAge),
	})
	if err != nil {
		return "", fmt.Errorf("storing session: %w", err)
	}

	return token, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

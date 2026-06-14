package service

import (
	"context"
	"net/mail"
	"shortify/server/internal/auth"
	"shortify/server/internal/domain"
	"shortify/server/internal/repository"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
}

func NewAuthService(users *repository.UserRepository, tokens *auth.TokenManager) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !validateEmail(email) || len(password) < 6 {
		return "", domain.ErrInvalidInput
	}

	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		return "", domain.ErrAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	err = s.users.Create(ctx, user)
	if err != nil {
		return "", err
	}

	return s.tokens.Generate(user.ID, email)
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !validateEmail(email) {
		return "", domain.ErrInvalidInput
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrInvalidCreds
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", domain.ErrInvalidCreds
	}

	return s.tokens.Generate(user.ID, email)
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

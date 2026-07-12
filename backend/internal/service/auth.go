package service

import (
	"context"
	"strings"
	"time"

	"ecommerce/internal/domain"
	"ecommerce/internal/repository"
	"ecommerce/internal/security"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (*domain.AuthenticatedUser, string, time.Time, error)

	FindAuthenticatedUserByID(
		ctx context.Context,
		userID string,
	) (*domain.AuthenticatedUser, error)
}

type authService struct {
	repository repository.AuthRepository
	jwtService security.JWTService
}

func NewAuthService(repository repository.AuthRepository, jwtService security.JWTService) AuthService {
	return &authService{
		repository: repository,
		jwtService: jwtService,
	}
}

func (s *authService) Login(
	ctx context.Context,
	email string,
	password string,
) (*domain.AuthenticatedUser, string, time.Time, error) {
	email = strings.TrimSpace(email)

	auth, err := s.repository.FindAuthenticationByEmail(ctx, email)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(auth.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, "", time.Time{}, domain.ErrInvalidCredentials
	}

	if auth.IsDeleted() {
		return nil, "", time.Time{}, domain.ErrInvalidCredentials
	}

	if !auth.Active {
		return nil, "", time.Time{}, domain.ErrUserInactive
	}

	token, expiresAt, err := s.jwtService.GenerateAccessToken(auth.UserID)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	if err := s.repository.UpdateLastLoginAt(ctx, auth.UserID, time.Now().UTC()); err != nil {
		return nil, "", time.Time{}, err
	}

	user := &domain.AuthenticatedUser{
		ID:    auth.UserID,
		Name:  auth.Name,
		Email: auth.Email,
		Role:  auth.Role,
	}

	return user, token, expiresAt, nil
}

func (s *authService) FindAuthenticatedUserByID(
	ctx context.Context,
	userID string,
) (*domain.AuthenticatedUser, error) {
	return s.repository.FindAuthenticatedUserByID(ctx, userID)
}

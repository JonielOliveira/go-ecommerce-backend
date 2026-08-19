package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"ecommerce/internal/domain"
	"ecommerce/internal/logging"
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

// logLoginFailure registra uma tentativa de login malsucedida. Antes de
// encontrar o usuário pelo e-mail só o e-mail está disponível; a partir daí,
// usa-se o user_id — nunca a senha, que não chega perto do logger.
func logLoginFailure(ctx context.Context, email string, userID string, err error) {
	attrs := []any{
		slog.String("operation", "auth.login"),
		slog.String("error", err.Error()),
	}

	if userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	} else {
		attrs = append(attrs, slog.String("email", email))
	}

	logging.FromContext(ctx).Warn("falha no login", attrs...)
}

func (s *authService) Login(
	ctx context.Context,
	email string,
	password string,
) (*domain.AuthenticatedUser, string, time.Time, error) {
	email = strings.TrimSpace(email)

	auth, err := s.repository.FindAuthenticationByEmail(ctx, email)
	if err != nil {
		logLoginFailure(ctx, email, "", err)
		return nil, "", time.Time{}, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(auth.PasswordHash),
		[]byte(password),
	); err != nil {
		logLoginFailure(ctx, email, auth.UserID, domain.ErrInvalidCredentials)
		return nil, "", time.Time{}, domain.ErrInvalidCredentials
	}

	if auth.IsDeleted() {
		logLoginFailure(ctx, email, auth.UserID, domain.ErrInvalidCredentials)
		return nil, "", time.Time{}, domain.ErrInvalidCredentials
	}

	if !auth.Active {
		logLoginFailure(ctx, email, auth.UserID, domain.ErrUserInactive)
		return nil, "", time.Time{}, domain.ErrUserInactive
	}

	token, expiresAt, err := s.jwtService.GenerateAccessToken(auth.UserID)
	if err != nil {
		logging.FromContext(ctx).Error("falha ao gerar token de acesso",
			slog.String("operation", "auth.login"),
			slog.String("user_id", auth.UserID),
			slog.String("error", err.Error()),
		)
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

	logging.FromContext(ctx).Info("login realizado",
		slog.String("operation", "auth.login"),
		slog.String("user_id", user.ID),
	)

	return user, token, expiresAt, nil
}

func (s *authService) FindAuthenticatedUserByID(
	ctx context.Context,
	userID string,
) (*domain.AuthenticatedUser, error) {
	return s.repository.FindAuthenticatedUserByID(ctx, userID)
}

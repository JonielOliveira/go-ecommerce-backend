package repository

import (
	"context"
	"time"

	"ecommerce/internal/domain"
)

type AuthRepository interface {
	FindAuthenticationByEmail(
		ctx context.Context,
		email string,
	) (*domain.UserAuthentication, error)

	FindAuthenticatedUserByID(
		ctx context.Context,
		userID string,
	) (*domain.AuthenticatedUser, error)

	UpdateLastLoginAt(
		ctx context.Context,
		userID string,
		loginAt time.Time,
	) error
}

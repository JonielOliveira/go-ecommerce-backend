package repository

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuthRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAuthRepository(db *pgxpool.Pool) *PostgresAuthRepository {
	return &PostgresAuthRepository{db: db}
}

func (r *PostgresAuthRepository) FindAuthenticationByEmail(
	ctx context.Context,
	email string,
) (*domain.UserAuthentication, error) {
	const query = `
		SELECT
			u.id,
			u.name,
			u.email,
			u.role,
			u.active,
			u.deleted_at,
			c.password_hash,
			c.password_changed_at
		FROM users u
		INNER JOIN user_password_credentials c
			ON c.user_id = u.id
		WHERE u.email = $1
	`

	var (
		auth domain.UserAuthentication
		role string
	)

	err := r.db.QueryRow(ctx, query, email).Scan(
		&auth.UserID,
		&auth.Name,
		&auth.Email,
		&role,
		&auth.Active,
		&auth.DeletedAt,
		&auth.PasswordHash,
		&auth.PasswordChangedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	auth.Role = domain.UserRole(role)

	return &auth, nil
}

func (r *PostgresAuthRepository) FindAuthenticatedUserByID(
	ctx context.Context,
	userID string,
) (*domain.AuthenticatedUser, error) {
	const query = `
		SELECT
			id,
			name,
			email,
			role
		FROM users
		WHERE id = $1
		  AND active = TRUE
		  AND deleted_at IS NULL
	`

	var (
		user domain.AuthenticatedUser
		role string
	)

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&role,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserInactive
	}

	if err != nil {
		return nil, err
	}

	user.Role = domain.UserRole(role)

	return &user, nil
}

func (r *PostgresAuthRepository) UpdateLastLoginAt(
	ctx context.Context,
	userID string,
	loginAt time.Time,
) error {
	const query = `
		UPDATE users
		SET
			last_login_at = $2,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID, loginAt)

	return err
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ecommerce/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolationCode = "23505"

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

type userState struct {
	active    bool
	deletedAt *time.Time
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var (
		id              string
		name            string
		email           string
		avatarURL       *string
		role            string
		active          bool
		emailVerifiedAt *time.Time
		lastLoginAt     *time.Time
		createdAt       time.Time
		updatedAt       time.Time
		deletedAt       *time.Time
	)

	if err := row.Scan(
		&id,
		&name,
		&email,
		&avatarURL,
		&role,
		&active,
		&emailVerifiedAt,
		&lastLoginAt,
		&createdAt,
		&updatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	return domain.RestoreUser(
		id,
		name,
		email,
		avatarURL,
		domain.UserRole(role),
		active,
		emailVerifiedAt,
		lastLoginAt,
		createdAt,
		updatedAt,
		deletedAt,
	)
}

func (r *PostgresUserRepository) findState(id string) (*userState, error) {
	const query = `
		SELECT active, deleted_at
		FROM users
		WHERE id = $1
	`

	var state userState

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(
		&state.active,
		&state.deletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (r *PostgresUserRepository) Create(user *domain.User, passwordHash string) (*domain.User, error) {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const insertUser = `
		INSERT INTO users (
			name, email, avatar_url, role
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id, name, email, avatar_url, role, active,
			email_verified_at, last_login_at, created_at, updated_at, deleted_at
	`

	createdUser, err := scanUser(tx.QueryRow(
		ctx,
		insertUser,
		user.Name(),
		user.Email(),
		user.AvatarURL(),
		string(user.Role()),
	))
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrUserEmailAlreadyExists
		}
		return nil, err
	}

	const insertCredentials = `
		INSERT INTO user_password_credentials (user_id, password_hash)
		VALUES ($1, $2)
	`

	if _, err := tx.Exec(ctx, insertCredentials, createdUser.ID(), passwordHash); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (r *PostgresUserRepository) Update(user *domain.User, passwordHash *string) (*domain.User, error) {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE users
		SET
			name = $1,
			email = $2,
			avatar_url = $3,
			role = $4,
			updated_at = NOW()
		WHERE id = $5
		  AND deleted_at IS NULL
		RETURNING
			id, name, email, avatar_url, role, active,
			email_verified_at, last_login_at, created_at, updated_at, deleted_at
	`

	updatedUser, err := scanUser(tx.QueryRow(
		ctx,
		query,
		user.Name(),
		user.Email(),
		user.AvatarURL(),
		string(user.Role()),
		user.ID(),
	))

	if errors.Is(err, pgx.ErrNoRows) {
		state, stateErr := r.findState(user.ID())
		if stateErr != nil {
			return nil, stateErr
		}

		if state.deletedAt != nil {
			return nil, domain.ErrUserAlreadyDeleted
		}

		return nil, err
	}

	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrUserEmailAlreadyExists
		}
		return nil, err
	}

	if passwordHash != nil {
		const updateCredentials = `
			UPDATE user_password_credentials
			SET password_hash = $1, password_changed_at = NOW()
			WHERE user_id = $2
		`

		if _, err := tx.Exec(ctx, updateCredentials, *passwordHash, updatedUser.ID()); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (r *PostgresUserRepository) FindByID(id string) (*domain.User, error) {
	const query = `
		SELECT
			id, name, email, avatar_url, role, active,
			email_verified_at, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
	`

	user, err := scanUser(r.db.QueryRow(context.Background(), query, id))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func buildUserFilters(filter UserSearchFilter) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	addCondition := func(condition string, value any) {
		args = append(args, value)

		conditions = append(
			conditions,
			fmt.Sprintf(condition, len(args)),
		)
	}

	switch filter.DeletionFilter {
	case DeletionFilterDeleted:
		conditions = append(conditions, "deleted_at IS NOT NULL")

	case DeletionFilterAll:
		// Não adiciona condição.

	default:
		conditions = append(conditions, "deleted_at IS NULL")
	}

	if strings.TrimSpace(filter.Name) != "" {
		addCondition(
			"name ILIKE $%d",
			"%"+strings.TrimSpace(filter.Name)+"%",
		)
	}

	if strings.TrimSpace(filter.Email) != "" {
		addCondition(
			"email ILIKE $%d",
			"%"+strings.TrimSpace(filter.Email)+"%",
		)
	}

	if strings.TrimSpace(filter.Role) != "" {
		addCondition(
			"role = $%d",
			strings.TrimSpace(filter.Role),
		)
	}

	if filter.Active != nil {
		addCondition(
			"active = $%d",
			*filter.Active,
		)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *PostgresUserRepository) Search(filter UserSearchFilter) (*UserSearchResult, error) {
	whereClause, args := buildUserFilters(filter)

	countQuery := `
		SELECT COUNT(*)
		FROM users
	` + whereClause

	var total int64

	if err := r.db.QueryRow(
		context.Background(),
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, err
	}

	selectQuery := `
		SELECT
			id, name, email, avatar_url, role, active,
			email_verified_at, last_login_at, created_at, updated_at, deleted_at
		FROM users
	` + whereClause + fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d
		OFFSET $%d
	`, len(args)+1, len(args)+2)

	selectArgs := append(
		args,
		filter.Limit,
		filter.Offset,
	)

	rows, err := r.db.Query(
		context.Background(),
		selectQuery,
		selectArgs...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*domain.User, 0)

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &UserSearchResult{
		Users: users,
		Total: total,
	}, nil
}

func (r *PostgresUserRepository) DeleteByID(id string) error {
	const query = `
		UPDATE users
		SET
			active = FALSE,
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL
		RETURNING id
	`

	var userID string

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&userID)

	if err == nil {
		return nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	state, err := r.findState(id)
	if err != nil {
		return err
	}

	if state.deletedAt != nil {
		return domain.ErrUserAlreadyDeleted
	}

	return nil
}

func (r *PostgresUserRepository) RestoreByID(id string) error {
	const query = `
		UPDATE users
		SET
			active = FALSE,
			deleted_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		AND deleted_at IS NOT NULL
		RETURNING id
	`

	var userID string

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&userID)

	if err == nil {
		return nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	state, err := r.findState(id)
	if err != nil {
		return err
	}

	if state.deletedAt == nil {
		return domain.ErrUserNotDeleted
	}

	return nil
}

func (r *PostgresUserRepository) ActivateByID(id string) error {
	const query = `
		UPDATE users
		SET
			active = TRUE,
			updated_at = NOW()
		WHERE id = $1
		  AND active = FALSE
		  AND deleted_at IS NULL
		RETURNING id
	`

	var userID string

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&userID)

	if err == nil {
		return nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	state, err := r.findState(id)
	if err != nil {
		return err
	}

	if state.deletedAt != nil {
		return domain.ErrUserAlreadyDeleted
	}

	if state.active {
		return domain.ErrUserAlreadyActive
	}

	return nil
}

func (r *PostgresUserRepository) DeactivateByID(id string) error {
	const query = `
		UPDATE users
		SET
			active = FALSE,
			updated_at = NOW()
		WHERE id = $1
		  AND active = TRUE
		  AND deleted_at IS NULL
		RETURNING id
	`

	var userID string

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&userID)

	if err == nil {
		return nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	state, err := r.findState(id)
	if err != nil {
		return err
	}

	if state.deletedAt != nil {
		return domain.ErrUserAlreadyDeleted
	}

	if !state.active {
		return domain.ErrUserAlreadyInactive
	}

	return nil
}

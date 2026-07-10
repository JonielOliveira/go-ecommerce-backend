package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ecommerce/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	db *pgxpool.Pool
}

type productState struct {
	active    bool
	deletedAt *time.Time
}

func NewPostgresProductRepository(db *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) findState(id string) (*productState, error) {
	const query = `
		SELECT active, deleted_at
		FROM products
		WHERE id = $1
	`

	var state productState

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(
		&state.active,
		&state.deletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}

	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (r *PostgresProductRepository) Create(product *domain.Product) (*domain.Product, error) {
	const query = `
		INSERT INTO products (
			name, description, price, stock, category_id, image_url
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id, name, description, price, stock, category_id, image_url,
			active, created_at, updated_at, deleted_at
	`

	var (
		id          string
		name        string
		description string
		price       float64
		stock       int
		categoryID  *string
		imageURL    *string
		active      bool
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   *time.Time
	)

	err := r.db.QueryRow(
		context.Background(),
		query,
		product.Name(),
		product.Description(),
		product.Price(),
		product.Stock(),
		product.CategoryID(),
		product.ImageURL(),
	).Scan(
		&id,
		&name,
		&description,
		&price,
		&stock,
		&categoryID,
		&imageURL,
		&active,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	return domain.RestoreProduct(
		id,
		name,
		description,
		price,
		stock,
		categoryID,
		imageURL,
		active,
		createdAt,
		updatedAt,
		deletedAt,
	)
}

func (r *PostgresProductRepository) Update(product *domain.Product) (*domain.Product, error) {
	const query = `
		UPDATE products
		SET
			name = $1,
			description = $2,
			price = $3,
			stock = $4,
			category_id = $5,
			image_url = $6,
			updated_at = NOW()
		WHERE id = $7
		  AND deleted_at IS NULL
		RETURNING
			id,
			name,
			description,
			price,
			stock,
			category_id,
			image_url,
			active,
			created_at,
			updated_at,
			deleted_at
	`

	var (
		id          string
		name        string
		description string
		price       float64
		stock       int
		categoryID  *string
		imageURL    *string
		active      bool
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   *time.Time
	)

	err := r.db.QueryRow(
		context.Background(),
		query,
		product.Name(),
		product.Description(),
		product.Price(),
		product.Stock(),
		product.CategoryID(),
		product.ImageURL(),
		product.ID(),
	).Scan(
		&id,
		&name,
		&description,
		&price,
		&stock,
		&categoryID,
		&imageURL,
		&active,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		state, stateErr := r.findState(product.ID())
		if stateErr != nil {
			return nil, stateErr
		}

		if state.deletedAt != nil {
			return nil, domain.ErrProductAlreadyDeleted
		}

		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return domain.RestoreProduct(
		id,
		name,
		description,
		price,
		stock,
		categoryID,
		imageURL,
		active,
		createdAt,
		updatedAt,
		deletedAt,
	)
}

func (r *PostgresProductRepository) FindByID(id string) (*domain.Product, error) {
	const query = `
		SELECT
			id,
			name,
			description,
			price,
			stock,
			category_id,
			image_url,
			active,
			created_at,
			updated_at,
			deleted_at
		FROM products
		WHERE id = $1
	`

	var (
		productID   string
		name        string
		description string
		price       float64
		stock       int
		categoryID  *string
		imageURL    *string
		active      bool
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   *time.Time
	)

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(
		&productID,
		&name,
		&description,
		&price,
		&stock,
		&categoryID,
		&imageURL,
		&active,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.RestoreProduct(
		productID,
		name,
		description,
		price,
		stock,
		categoryID,
		imageURL,
		active,
		createdAt,
		updatedAt,
		deletedAt,
	)
}

func buildProductFilters(filter ProductSearchFilter) (string, []any) {
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

	if filter.CategoryID != nil {
		addCondition(
			"category_id = $%d",
			*filter.CategoryID,
		)
	}

	if filter.Active != nil {
		addCondition(
			"active = $%d",
			*filter.Active,
		)
	}

	if filter.MinPrice != nil {
		addCondition(
			"price >= $%d",
			*filter.MinPrice,
		)
	}

	if filter.MaxPrice != nil {
		addCondition(
			"price <= $%d",
			*filter.MaxPrice,
		)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *PostgresProductRepository) Search(filter ProductSearchFilter) (*ProductSearchResult, error) {
	whereClause, args := buildProductFilters(filter)

	countQuery := `
		SELECT COUNT(*)
		FROM products
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
			id,
			name,
			description,
			price,
			stock,
			category_id,
			image_url,
			active,
			created_at,
			updated_at,
			deleted_at
		FROM products
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

	products := make([]*domain.Product, 0)

	for rows.Next() {
		var (
			id          string
			name        string
			description string
			price       float64
			stock       int
			categoryID  *string
			imageURL    *string
			active      bool
			createdAt   time.Time
			updatedAt   time.Time
			deletedAt   *time.Time
		)

		if err := rows.Scan(
			&id,
			&name,
			&description,
			&price,
			&stock,
			&categoryID,
			&imageURL,
			&active,
			&createdAt,
			&updatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}

		product, err := domain.RestoreProduct(
			id,
			name,
			description,
			price,
			stock,
			categoryID,
			imageURL,
			active,
			createdAt,
			updatedAt,
			deletedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ProductSearchResult{
		Products: products,
		Total:    total,
	}, nil
}

func (r *PostgresProductRepository) DeleteByID(id string) error {
	const query = `
		UPDATE products
		SET
			active = FALSE,
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL
		RETURNING id
	`

	var productID string

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&productID)

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
		return domain.ErrProductAlreadyDeleted
	}

	return nil
}

func (r *PostgresProductRepository) RestoreByID(id string) error {
	const query = `
		UPDATE products
		SET
			active = FALSE,
			deleted_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		AND deleted_at IS NOT NULL
		RETURNING id
	`

	var productID string

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&productID)

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
		return domain.ErrProductNotDeleted
	}

	return nil
}

func (r *PostgresProductRepository) ActivateByID(id string) error {
	const query = `
		UPDATE products
		SET
			active = TRUE,
			updated_at = NOW()
		WHERE id = $1
		  AND active = FALSE
		  AND deleted_at IS NULL
		RETURNING id
	`

	var productID string

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&productID)

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
		return domain.ErrProductAlreadyDeleted
	}

	if state.active {
		return domain.ErrProductAlreadyActive
	}

	return nil
}

func (r *PostgresProductRepository) DeactivateByID(id string) error {
	const query = `
		UPDATE products
		SET
			active = FALSE,
			updated_at = NOW()
		WHERE id = $1
		  AND active = TRUE
		  AND deleted_at IS NULL
		RETURNING id
	`

	var productID string

	err := r.db.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&productID)

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
		return domain.ErrProductAlreadyDeleted
	}

	if !state.active {
		return domain.ErrProductAlreadyInactive
	}

	return nil
}

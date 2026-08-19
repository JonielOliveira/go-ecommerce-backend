package repository

import (
	"context"

	"ecommerce/internal/domain"
)

type DeletionFilter int

const (
	DeletionFilterNotDeleted DeletionFilter = iota
	DeletionFilterDeleted
	DeletionFilterAll
)

type ProductSearchFilter struct {
	Name           string
	CategoryID     *string
	Active         *bool
	DeletionFilter DeletionFilter
	MinPrice       *float64
	MaxPrice       *float64
	Limit          int
	Offset         int
}

type ProductSearchResult struct {
	Products []*domain.Product
	Total    int64
}

type ProductRepository interface {
	// Save(product *domain.Product) (*domain.Product, error)
	// FindAll() ([]*domain.Product, error)
	Create(ctx context.Context, product *domain.Product) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) (*domain.Product, error)

	FindByID(ctx context.Context, id string) (*domain.Product, error)
	Search(ctx context.Context, filter ProductSearchFilter) (*ProductSearchResult, error)

	DeleteByID(ctx context.Context, id string) error
	RestoreByID(ctx context.Context, id string) error
	ActivateByID(ctx context.Context, id string) error
	DeactivateByID(ctx context.Context, id string) error
}

package repository

import "ecommerce/internal/domain"

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
	Create(product *domain.Product) (*domain.Product, error)
	Update(product *domain.Product) (*domain.Product, error)

	FindByID(id string) (*domain.Product, error)
	Search(filter ProductSearchFilter) (*ProductSearchResult, error)

	DeleteByID(id string) error
	RestoreByID(id string) error
	ActivateByID(id string) error
	DeactivateByID(id string) error
}

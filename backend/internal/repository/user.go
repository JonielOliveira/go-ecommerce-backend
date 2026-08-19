package repository

import (
	"context"

	"ecommerce/internal/domain"
)

type UserSearchFilter struct {
	Name           string
	Email          string
	Role           string
	Active         *bool
	DeletionFilter DeletionFilter
	Limit          int
	Offset         int
}

type UserSearchResult struct {
	Users []*domain.User
	Total int64
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User, passwordHash string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User, passwordHash *string) (*domain.User, error)

	FindByID(ctx context.Context, id string) (*domain.User, error)
	Search(ctx context.Context, filter UserSearchFilter) (*UserSearchResult, error)

	DeleteByID(ctx context.Context, id string) error
	RestoreByID(ctx context.Context, id string) error
	ActivateByID(ctx context.Context, id string) error
	DeactivateByID(ctx context.Context, id string) error
}

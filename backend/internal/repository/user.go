package repository

import "ecommerce/internal/domain"

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
	Create(user *domain.User, passwordHash string) (*domain.User, error)
	Update(user *domain.User, passwordHash *string) (*domain.User, error)

	FindByID(id string) (*domain.User, error)
	Search(filter UserSearchFilter) (*UserSearchResult, error)

	DeleteByID(id string) error
	RestoreByID(id string) error
	ActivateByID(id string) error
	DeactivateByID(id string) error
}

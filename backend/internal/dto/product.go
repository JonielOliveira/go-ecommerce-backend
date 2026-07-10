package dto

import "time"

type ProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"required,gte=0"`
	CategoryID  *string `json:"categoryId"`
	ImageURL    *string `json:"imageUrl"`
}

type ProductUpdateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"gt=0"`
	Stock       int     `json:"stock" binding:"gte=0"`
	CategoryID  *string `json:"categoryId"`
	ImageURL    *string `json:"imageUrl"`
}

type ProductResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Stock       int        `json:"stock"`
	CategoryID  *string    `json:"categoryId"`
	ImageURL    *string    `json:"imageUrl"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
}

type DeletionState string

const (
	DeletionStateNotDeleted DeletionState = "not_deleted"
	DeletionStateDeleted    DeletionState = "deleted"
	DeletionStateAll        DeletionState = "all"
)

type ProductSearchRequest struct {
	Name          string        `form:"name"`
	CategoryID    *string       `form:"categoryId"`
	Active        *bool         `form:"active"`
	DeletionState DeletionState `form:"deletionState"`
	MinPrice      *float64      `form:"minPrice"`
	MaxPrice      *float64      `form:"maxPrice"`
	Page          int           `form:"page"`
	PageSize      int           `form:"pageSize"`
}

type ProductPageResponse struct {
	Items      []ProductResponse `json:"items"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalItems int64             `json:"totalItems"`
	TotalPages int               `json:"totalPages"`
}

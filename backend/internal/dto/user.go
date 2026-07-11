package dto

import "time"

type UserRequest struct {
	Name      string  `json:"name" binding:"required"`
	Email     string  `json:"email" binding:"required,email"`
	Password  string  `json:"password" binding:"required,min=8"`
	Role      string  `json:"role"`
	AvatarURL *string `json:"avatarUrl"`
}

type UserUpdateRequest struct {
	Name      string  `json:"name"`
	Email     string  `json:"email" binding:"omitempty,email"`
	Password  string  `json:"password" binding:"omitempty,min=8"`
	Role      string  `json:"role"`
	AvatarURL *string `json:"avatarUrl"`
}

type UserResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	AvatarURL       *string    `json:"avatarUrl"`
	Role            string     `json:"role"`
	Active          bool       `json:"active"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
	LastLoginAt     *time.Time `json:"lastLoginAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
}

type UserSearchRequest struct {
	Name          string        `form:"name"`
	Email         string        `form:"email"`
	Role          string        `form:"role"`
	Active        *bool         `form:"active"`
	DeletionState DeletionState `form:"deletionState"`
	Page          int           `form:"page"`
	PageSize      int           `form:"pageSize"`
}

type UserPageResponse struct {
	Items      []UserResponse `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	TotalItems int64          `json:"totalItems"`
	TotalPages int            `json:"totalPages"`
}

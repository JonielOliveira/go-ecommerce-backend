package dto

import "time"

// CreateUserRequest é o DTO da criação administrativa de usuário
// (POST /users, restrito a admins). "Role" é opcional e um ponteiro para
// distinguir "campo não enviado" (usa o padrão "customer") de "campo
// enviado" — ver service.UserService.Create.
type CreateUserRequest struct {
	Name      string  `json:"name" binding:"required,max=255"`
	Email     string  `json:"email" binding:"required,email,max=255"`
	Password  string  `json:"password" binding:"required,min=8,max=128"`
	Role      *string `json:"role"`
	AvatarURL *string `json:"avatarUrl"`
}

// UserUpdateRequest é usado apenas em rotas administrativas, por isso pode
// expor "role" (permite promover/rebaixar usuários).
type UserUpdateRequest struct {
	Name      string  `json:"name"`
	Email     string  `json:"email" binding:"omitempty,email"`
	Password  string  `json:"password" binding:"omitempty,min=8,max=128"`
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

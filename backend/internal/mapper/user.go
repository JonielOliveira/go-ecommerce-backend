package mapper

import (
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

func NewUserResponse(u *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:              u.ID(),
		Name:            u.Name(),
		Email:           u.Email(),
		AvatarURL:       u.AvatarURL(),
		Role:            string(u.Role()),
		Active:          u.IsActive(),
		EmailVerifiedAt: u.EmailVerifiedAt(),
		LastLoginAt:     u.LastLoginAt(),
		CreatedAt:       u.CreatedAt(),
		UpdatedAt:       u.UpdatedAt(),
		DeletedAt:       u.DeletedAt(),
	}
}

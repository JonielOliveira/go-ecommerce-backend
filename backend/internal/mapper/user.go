package mapper

import (
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

func NewUser(req dto.UserRequest) (*domain.User, error) {
	role := domain.UserRole(req.Role)
	if role == "" {
		role = domain.RoleCustomer
	}

	return domain.NewUser(req.Name, req.Email, role, req.AvatarURL)
}

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

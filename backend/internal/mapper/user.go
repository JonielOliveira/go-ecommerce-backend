package mapper

import (
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

// NewUser constrói um usuário a partir do cadastro público. O papel é
// sempre "customer": UserRequest não expõe "role" para impedir que o
// cliente se autopromova a admin.
func NewUser(req dto.UserRequest) (*domain.User, error) {
	return domain.NewUser(req.Name, req.Email, domain.RoleCustomer, req.AvatarURL)
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

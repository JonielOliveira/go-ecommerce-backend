package dto

// RegisterRequest é o DTO do autocadastro público (POST /auth/register).
// Não expõe "role" de propósito: essa rota sempre cria um usuário
// "customer" — ver service.UserService.Register.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,max=255"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type AuthUserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginResponse struct {
	User AuthUserResponse `json:"user"`
}

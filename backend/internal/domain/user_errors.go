package domain

import "errors"

var (
	// User
	ErrUserNotFound = errors.New("usuário não encontrado")

	// Validation
	ErrInvalidUserName     = errors.New("nome do usuário inválido")
	ErrInvalidUserEmail    = errors.New("email do usuário inválido")
	ErrInvalidUserRole     = errors.New("papel do usuário inválido")
	ErrInvalidUserPassword = errors.New("senha do usuário inválida")

	// Uniqueness
	ErrUserEmailAlreadyExists = errors.New("email já está em uso")

	// Soft Delete
	ErrUserAlreadyDeleted = errors.New("usuário já está removido")
	ErrUserNotDeleted     = errors.New("usuário não está removido")

	// Activation
	ErrUserAlreadyActive   = errors.New("usuário já está ativo")
	ErrUserAlreadyInactive = errors.New("usuário já está inativo")
)

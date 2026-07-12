package domain

import "errors"

var (
	ErrInvalidCredentials     = errors.New("credenciais inválidas")
	ErrUserInactive           = errors.New("usuário inativo")
	ErrUserDeleted            = errors.New("usuário excluído")
	ErrAuthenticationRequired = errors.New("autenticação necessária")
	ErrInvalidToken           = errors.New("token inválido")
	ErrExpiredToken           = errors.New("token expirado")
	ErrForbidden              = errors.New("acesso negado")
)

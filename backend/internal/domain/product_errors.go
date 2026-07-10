package domain

import "errors"

var (
	// Product
	ErrProductNotFound = errors.New("produto não encontrado")

	// Validation
	ErrInvalidProductName        = errors.New("nome do produto inválido")
	ErrInvalidProductDescription = errors.New("descrição do produto inválida")
	ErrInvalidProductPrice       = errors.New("preço do produto inválido")
	ErrInvalidProductStock       = errors.New("estoque do produto inválido")
	ErrInvalidQuantity           = errors.New("quantidade inválida")
	ErrInsufficientStock         = errors.New("estoque insuficiente")

	// Soft Delete
	ErrProductAlreadyDeleted = errors.New("produto já está removido")
	ErrProductNotDeleted     = errors.New("produto não está removido")

	// Activation
	ErrProductAlreadyActive   = errors.New("produto já está ativo")
	ErrProductAlreadyInactive = errors.New("produto já está inativo")
)

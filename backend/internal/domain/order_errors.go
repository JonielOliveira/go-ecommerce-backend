package domain

import "errors"

// ErrInsufficientStock já existe em product_errors.go e é reutilizado aqui
// (mesma condição: estoque insuficiente para atender a quantidade pedida).
var (
	ErrOrderNotFound         = errors.New("pedido não encontrado")
	ErrOrderMustHaveItems    = errors.New("pedido precisa ter ao menos um item")
	ErrInvalidOrderItem      = errors.New("item de pedido inválido")
	ErrInvalidOrderQuantity  = errors.New("quantidade do item deve ser maior que zero")
	ErrInvalidOrderStatus    = errors.New("status de pedido inválido")
	ErrOrderCannotBePaid     = errors.New("pedido não pode ser pago")
	ErrOrderCannotBeCanceled = errors.New("pedido não pode ser cancelado")
	ErrProductUnavailable    = errors.New("produto indisponível")
	ErrOrderOwnerNotFound    = errors.New("proprietário do pedido não encontrado")
	ErrOrderOwnerUnavailable = errors.New("proprietário do pedido indisponível")
	ErrOrderAccessDenied     = errors.New("acesso ao pedido negado")
)

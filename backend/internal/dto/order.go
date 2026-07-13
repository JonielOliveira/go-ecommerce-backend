package dto

import "time"

// CreateOrderItemRequest e CreateOrderRequest não possuem customer_id (nem
// user_id/owner_id): o dono do pedido vem sempre do usuário autenticado —
// ver service.OrderService.Create.
type CreateOrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type OrderItemResponse struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
}

type OrderResponse struct {
	ID          string              `json:"id"`
	CustomerID  string              `json:"customer_id"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	Items       []OrderItemResponse `json:"items"`
	PaidAt      *time.Time          `json:"paid_at"`
	CanceledAt  *time.Time          `json:"canceled_at"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// OrderSearchRequest segue exatamente o mesmo padrão de ProductSearchRequest
// e UserSearchRequest: parâmetros de query "page"/"pageSize", sem tags de
// binding (valores inválidos são normalizados no service, não rejeitados
// aqui — mesmo comportamento já usado por produtos e usuários).
type OrderSearchRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

// OrderPageResponse tem exatamente a mesma forma de ProductPageResponse e
// UserPageResponse (mesmos campos, mesma capitalização de JSON), para manter
// um único formato de resposta paginada em toda a aplicação.
type OrderPageResponse struct {
	Items      []OrderResponse `json:"items"`
	Page       int             `json:"page"`
	PageSize   int             `json:"pageSize"`
	TotalItems int64           `json:"totalItems"`
	TotalPages int             `json:"totalPages"`
}

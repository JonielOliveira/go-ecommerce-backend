package domain

import "time"

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	UnitPrice float64
	CreatedAt time.Time
}

func (i OrderItem) Subtotal() float64 {
	return float64(i.Quantity) * i.UnitPrice
}

// CreateOrderItem representa um item já consolidado (sem product_id
// duplicado) pronto para ser processado pelo OrderRepository.Create.
type CreateOrderItem struct {
	ProductID string
	Quantity  int
}

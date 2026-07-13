package domain

type OrderStatus string

const (
	OrderStatusPending  OrderStatus = "PENDING"
	OrderStatusPaid     OrderStatus = "PAID"
	OrderStatusCanceled OrderStatus = "CANCELED"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusPaid, OrderStatusCanceled:
		return true
	default:
		return false
	}
}

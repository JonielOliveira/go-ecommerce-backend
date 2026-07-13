package domain

import "time"

// Order representa um pedido. CustomerID é o dono do pedido: a coluna do
// banco se chama customer_id, mas o proprietário pode ter papel "customer"
// ou "admin" — ver observação da TASK_3 sobre esse campo.
type Order struct {
	ID          string
	CustomerID  string
	Status      OrderStatus
	TotalAmount float64
	Items       []OrderItem
	PaidAt      *time.Time
	CanceledAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (o *Order) CanPay() bool {
	return o.Status == OrderStatusPending
}

func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusPending
}

func (o *Order) Pay(now time.Time) error {
	if !o.CanPay() {
		return ErrOrderCannotBePaid
	}

	o.Status = OrderStatusPaid
	o.PaidAt = &now
	o.UpdatedAt = now

	return nil
}

func (o *Order) Cancel(now time.Time) error {
	if !o.CanCancel() {
		return ErrOrderCannotBeCanceled
	}

	o.Status = OrderStatusCanceled
	o.CanceledAt = &now
	o.UpdatedAt = now

	return nil
}

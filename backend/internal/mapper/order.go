package mapper

import (
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

func NewOrderResponse(o *domain.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(o.Items))

	for _, item := range o.Items {
		items = append(items, dto.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal:  item.Subtotal(),
		})
	}

	return dto.OrderResponse{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      string(o.Status),
		TotalAmount: o.TotalAmount,
		Items:       items,
		PaidAt:      o.PaidAt,
		CanceledAt:  o.CanceledAt,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

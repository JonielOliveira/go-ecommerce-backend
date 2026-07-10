package mapper

import (
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

func NewProduct(req dto.ProductRequest) (*domain.Product, error) {
	return domain.NewProduct(
		req.Name,
		req.Description,
		req.Price,
		req.Stock,
		req.CategoryID,
		req.ImageURL,
	)
}

func NewProductResponse(p *domain.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:          p.ID(),
		Name:        p.Name(),
		Description: p.Description(),
		Price:       p.Price(),
		Stock:       p.Stock(),
		CategoryID:  p.CategoryID(),
		ImageURL:    p.ImageURL(),
		Active:      p.IsActive(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
		DeletedAt:   p.DeletedAt(),
	}
}

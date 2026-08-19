package service

import (
	"context"
	"log/slog"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/logging"
	"ecommerce/internal/mapper"
	"ecommerce/internal/repository"
)

// ProductService declara os casos de uso de produto consumidos pelo
// Handler. O tipo concreto abaixo satisfaz o contrato implicitamente.
type ProductService interface {
	Create(ctx context.Context, request dto.ProductRequest) (dto.ProductResponse, error)
	Update(ctx context.Context, id string, request dto.ProductUpdateRequest) (dto.ProductResponse, error)
	FindByID(ctx context.Context, id string) (dto.ProductResponse, error)
	Search(ctx context.Context, filter dto.ProductSearchRequest) (dto.ProductPageResponse, error)
	DeleteByID(ctx context.Context, id string) error
	RestoreByID(ctx context.Context, id string) error
	ActivateByID(ctx context.Context, id string) error
	DeactivateByID(ctx context.Context, id string) error
}

type productService struct {
	repository repository.ProductRepository
}

func NewProductService(repository repository.ProductRepository) ProductService {
	return &productService{
		repository: repository,
	}
}

func (s *productService) Create(ctx context.Context, request dto.ProductRequest) (dto.ProductResponse, error) {
	product, err := mapper.NewProduct(request)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	createdProduct, err := s.repository.Create(ctx, product)
	if err != nil {
		logging.FromContext(ctx).Warn("falha ao criar produto",
			slog.String("operation", "product.create"),
			slog.String("error", err.Error()),
		)
		return dto.ProductResponse{}, err
	}

	logging.FromContext(ctx).Info("produto criado",
		slog.String("operation", "product.create"),
		slog.String("product_id", createdProduct.ID()),
	)

	return mapper.NewProductResponse(createdProduct), nil
}

func (s *productService) Update(ctx context.Context, id string, request dto.ProductUpdateRequest) (dto.ProductResponse, error) {
	product, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	if product.IsDeleted() {
		return dto.ProductResponse{}, domain.ErrProductAlreadyDeleted
	}

	if err := product.Update(
		request.Name,
		request.Description,
		request.Price,
		request.Stock,
		request.CategoryID,
		request.ImageURL,
	); err != nil {
		return dto.ProductResponse{}, err
	}

	updatedProduct, err := s.repository.Update(ctx, product)
	if err != nil {
		logging.FromContext(ctx).Warn("falha ao atualizar produto",
			slog.String("operation", "product.update"),
			slog.String("product_id", id),
			slog.String("error", err.Error()),
		)
		return dto.ProductResponse{}, err
	}

	logging.FromContext(ctx).Info("produto atualizado",
		slog.String("operation", "product.update"),
		slog.String("product_id", updatedProduct.ID()),
	)

	return mapper.NewProductResponse(updatedProduct), nil
}

func (s *productService) FindByID(ctx context.Context, id string) (dto.ProductResponse, error) {
	product, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapper.NewProductResponse(product), nil
}

// func (s *productService) FindAll() ([]dto.ProductResponse, error) {
// 	products, err := s.repository.FindAll()
// 	if err != nil {
// 		return nil, err
// 	}

// 	responses := make([]dto.ProductResponse, 0, len(products))

// 	for _, product := range products {
// 		responses = append(responses, mapper.NewProductResponse(product))
// 	}

// 	return responses, nil
// }

func mapDeletionFilter(state dto.DeletionState) repository.DeletionFilter {
	switch state {
	case dto.DeletionStateDeleted:
		return repository.DeletionFilterDeleted

	case dto.DeletionStateAll:
		return repository.DeletionFilterAll

	default:
		return repository.DeletionFilterNotDeleted
	}
}

func (s *productService) Search(ctx context.Context, filter dto.ProductSearchRequest) (dto.ProductPageResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	repositoryFilter := repository.ProductSearchFilter{
		Name:           filter.Name,
		CategoryID:     filter.CategoryID,
		Active:         filter.Active,
		DeletionFilter: mapDeletionFilter(filter.DeletionState),
		MinPrice:       filter.MinPrice,
		MaxPrice:       filter.MaxPrice,
		Limit:          filter.PageSize,
		Offset:         (filter.Page - 1) * filter.PageSize,
	}

	result, err := s.repository.Search(ctx, repositoryFilter)
	if err != nil {
		return dto.ProductPageResponse{}, err
	}

	items := make([]dto.ProductResponse, 0, len(result.Products))

	for _, product := range result.Products {
		items = append(items, mapper.NewProductResponse(product))
	}

	totalPages := int(
		(result.Total + int64(filter.PageSize) - 1) /
			int64(filter.PageSize),
	)

	return dto.ProductPageResponse{
		Items:      items,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalItems: result.Total,
		TotalPages: totalPages,
	}, nil
}

func (s *productService) DeleteByID(ctx context.Context, id string) error {
	if err := s.repository.DeleteByID(ctx, id); err != nil {
		logging.FromContext(ctx).Warn("falha ao excluir produto",
			slog.String("operation", "product.delete"),
			slog.String("product_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	logging.FromContext(ctx).Info("produto excluído",
		slog.String("operation", "product.delete"),
		slog.String("product_id", id),
	)

	return nil
}

func (s *productService) RestoreByID(ctx context.Context, id string) error {
	if err := s.repository.RestoreByID(ctx, id); err != nil {
		logging.FromContext(ctx).Warn("falha ao restaurar produto",
			slog.String("operation", "product.restore"),
			slog.String("product_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	logging.FromContext(ctx).Info("produto restaurado",
		slog.String("operation", "product.restore"),
		slog.String("product_id", id),
	)

	return nil
}

func (s *productService) ActivateByID(ctx context.Context, id string) error {
	if err := s.repository.ActivateByID(ctx, id); err != nil {
		logging.FromContext(ctx).Warn("falha ao ativar produto",
			slog.String("operation", "product.activate"),
			slog.String("product_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	logging.FromContext(ctx).Info("produto ativado",
		slog.String("operation", "product.activate"),
		slog.String("product_id", id),
	)

	return nil
}

func (s *productService) DeactivateByID(ctx context.Context, id string) error {
	if err := s.repository.DeactivateByID(ctx, id); err != nil {
		logging.FromContext(ctx).Warn("falha ao desativar produto",
			slog.String("operation", "product.deactivate"),
			slog.String("product_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	logging.FromContext(ctx).Info("produto desativado",
		slog.String("operation", "product.deactivate"),
		slog.String("product_id", id),
	)

	return nil
}

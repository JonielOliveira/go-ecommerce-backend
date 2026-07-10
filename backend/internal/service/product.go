package service

import (
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/mapper"
	"ecommerce/internal/repository"
)

type ProductService struct {
	repository repository.ProductRepository
}

func NewProductService(repository repository.ProductRepository) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) Create(request dto.ProductRequest) (dto.ProductResponse, error) {
	product, err := mapper.NewProduct(request)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	createdProduct, err := s.repository.Create(product)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapper.NewProductResponse(createdProduct), nil
}

func (s *ProductService) Update(id string, request dto.ProductUpdateRequest) (dto.ProductResponse, error) {
	product, err := s.repository.FindByID(id)
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

	updatedProduct, err := s.repository.Update(product)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapper.NewProductResponse(updatedProduct), nil
}

func (s *ProductService) FindByID(id string) (dto.ProductResponse, error) {
	product, err := s.repository.FindByID(id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapper.NewProductResponse(product), nil
}

// func (s *ProductService) FindAll() ([]dto.ProductResponse, error) {
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

func (s *ProductService) Search(filter dto.ProductSearchRequest) (dto.ProductPageResponse, error) {
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

	result, err := s.repository.Search(repositoryFilter)
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

func (s *ProductService) DeleteByID(id string) error {
	return s.repository.DeleteByID(id)
}

func (s *ProductService) RestoreByID(id string) error {
	return s.repository.RestoreByID(id)
}

func (s *ProductService) ActivateByID(id string) error {
	return s.repository.ActivateByID(id)
}

func (s *ProductService) DeactivateByID(id string) error {
	return s.repository.DeactivateByID(id)
}

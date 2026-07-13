package service

import (
	"context"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/mapper"
	"ecommerce/internal/repository"
)

// Mesmos valores padrão e máximo já usados por ProductService.Search e
// UserService.Search — a paginação de pedidos não introduz um terceiro
// conjunto de limites.
const (
	defaultOrderPageSize = 20
	maxOrderPageSize     = 100
)

type OrderService interface {
	Create(
		ctx context.Context,
		authenticatedUser domain.AuthenticatedUser,
		request dto.CreateOrderRequest,
	) (dto.OrderResponse, error)

	Search(
		ctx context.Context,
		authenticatedUser domain.AuthenticatedUser,
		page int,
		pageSize int,
	) (dto.OrderPageResponse, error)

	FindByID(
		ctx context.Context,
		authenticatedUser domain.AuthenticatedUser,
		orderID string,
	) (dto.OrderResponse, error)

	PayByID(
		ctx context.Context,
		authenticatedUser domain.AuthenticatedUser,
		orderID string,
	) (dto.OrderResponse, error)

	CancelByID(
		ctx context.Context,
		authenticatedUser domain.AuthenticatedUser,
		orderID string,
	) (dto.OrderResponse, error)
}

type orderService struct {
	repository repository.OrderRepository
}

func NewOrderService(repository repository.OrderRepository) OrderService {
	return &orderService{
		repository: repository,
	}
}

// consolidateOrderItems soma as quantidades de product_id repetidos,
// preservando a ordem de primeira ocorrência para uma resposta
// determinística, e detecta overflow ao somar quantidades.
func consolidateOrderItems(items []dto.CreateOrderItemRequest) ([]domain.CreateOrderItem, error) {
	quantities := make(map[string]int, len(items))
	order := make([]string, 0, len(items))

	for _, item := range items {
		current, exists := quantities[item.ProductID]
		if !exists {
			order = append(order, item.ProductID)
		}

		sum := current + item.Quantity
		if sum < current {
			return nil, domain.ErrInvalidOrderQuantity
		}

		quantities[item.ProductID] = sum
	}

	consolidated := make([]domain.CreateOrderItem, 0, len(order))

	for _, productID := range order {
		consolidated = append(consolidated, domain.CreateOrderItem{
			ProductID: productID,
			Quantity:  quantities[productID],
		})
	}

	return consolidated, nil
}

func (s *orderService) Create(
	ctx context.Context,
	authenticatedUser domain.AuthenticatedUser,
	request dto.CreateOrderRequest,
) (dto.OrderResponse, error) {
	if len(request.Items) == 0 {
		return dto.OrderResponse{}, domain.ErrOrderMustHaveItems
	}

	items, err := consolidateOrderItems(request.Items)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	// O proprietário do pedido vem exclusivamente do usuário autenticado —
	// nunca do corpo da requisição, nem mesmo para admin.
	order, err := s.repository.Create(ctx, authenticatedUser.ID, items)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	return mapper.NewOrderResponse(order), nil
}

func (s *orderService) Search(
	ctx context.Context,
	authenticatedUser domain.AuthenticatedUser,
	page int,
	pageSize int,
) (dto.OrderPageResponse, error) {
	// Mesma normalização silenciosa (sem 400) já usada por
	// ProductService.Search e UserService.Search.
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = defaultOrderPageSize
	}

	if pageSize > maxOrderPageSize {
		pageSize = maxOrderPageSize
	}

	filter := repository.OrderFilter{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	// Customer só vê os próprios pedidos, e só é contado entre os
	// próprios; admin não recebe filtro de propriedade. O cliente não tem
	// como sobrescrever isso via query param, pois o filtro é decidido
	// aqui a partir do papel autenticado.
	if authenticatedUser.Role != domain.RoleAdmin {
		filter.CustomerID = &authenticatedUser.ID
	}

	result, err := s.repository.Search(ctx, filter)
	if err != nil {
		return dto.OrderPageResponse{}, err
	}

	items := make([]dto.OrderResponse, 0, len(result.Orders))
	for i := range result.Orders {
		items = append(items, mapper.NewOrderResponse(&result.Orders[i]))
	}

	totalPages := int(
		(result.Total + int64(pageSize) - 1) /
			int64(pageSize),
	)

	return dto.OrderPageResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: result.Total,
		TotalPages: totalPages,
	}, nil
}

func (s *orderService) FindByID(
	ctx context.Context,
	authenticatedUser domain.AuthenticatedUser,
	orderID string,
) (dto.OrderResponse, error) {
	order, err := s.repository.FindByID(ctx, orderID)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	if authenticatedUser.Role != domain.RoleAdmin && order.CustomerID != authenticatedUser.ID {
		return dto.OrderResponse{}, domain.ErrOrderAccessDenied
	}

	return mapper.NewOrderResponse(order), nil
}

func (s *orderService) PayByID(
	ctx context.Context,
	authenticatedUser domain.AuthenticatedUser,
	orderID string,
) (dto.OrderResponse, error) {
	// A propriedade é exigida atomicamente pelo repository (id + ownerID +
	// status PENDING na mesma instrução SQL) — sem exceção para admin.
	order, err := s.repository.PayByID(ctx, orderID, authenticatedUser.ID)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	return mapper.NewOrderResponse(order), nil
}

func (s *orderService) CancelByID(
	ctx context.Context,
	authenticatedUser domain.AuthenticatedUser,
	orderID string,
) (dto.OrderResponse, error) {
	isAdmin := authenticatedUser.Role == domain.RoleAdmin

	order, err := s.repository.CancelByID(ctx, orderID, authenticatedUser.ID, isAdmin)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	return mapper.NewOrderResponse(order), nil
}

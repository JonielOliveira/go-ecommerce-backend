package service

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/repository"
)

// fakeOrderRepository substitui o PostgreSQL e permite controlar cada
// operação no teste. Os métodos abaixo fazem este tipo satisfazer a
// interface repository.OrderRepository.
type fakeOrderRepository struct {
	createFunc     func(ctx context.Context, ownerID string, items []domain.CreateOrderItem) (*domain.Order, error)
	searchFunc     func(ctx context.Context, filter repository.OrderFilter) (*repository.OrderSearchResult, error)
	findByIDFunc   func(ctx context.Context, id string) (*domain.Order, error)
	payByIDFunc    func(ctx context.Context, id string, ownerID string) (*domain.Order, error)
	cancelByIDFunc func(ctx context.Context, id string, requesterID string, isAdmin bool) (*domain.Order, error)
}

func (fake *fakeOrderRepository) Create(ctx context.Context, ownerID string, items []domain.CreateOrderItem) (*domain.Order, error) {
	return fake.createFunc(ctx, ownerID, items)
}

func (fake *fakeOrderRepository) Search(ctx context.Context, filter repository.OrderFilter) (*repository.OrderSearchResult, error) {
	return fake.searchFunc(ctx, filter)
}

func (fake *fakeOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	return fake.findByIDFunc(ctx, id)
}

func (fake *fakeOrderRepository) PayByID(ctx context.Context, id string, ownerID string) (*domain.Order, error) {
	return fake.payByIDFunc(ctx, id, ownerID)
}

func (fake *fakeOrderRepository) CancelByID(ctx context.Context, id string, requesterID string, isAdmin bool) (*domain.Order, error) {
	return fake.cancelByIDFunc(ctx, id, requesterID, isAdmin)
}

// TestConsolidateOrderItems verifica a regra pura usada por
// OrderService.Create: quantidades de product_id repetidos são somadas,
// preservando a ordem da primeira ocorrência, e overflow é detectado.
func TestConsolidateOrderItems(t *testing.T) {
	testCases := []struct {
		name    string
		items   []dto.CreateOrderItemRequest
		want    []domain.CreateOrderItem
		wantErr error
	}{
		{
			name: "soma quantidades repetidas mantendo ordem da primeira ocorrência",
			items: []dto.CreateOrderItemRequest{
				{ProductID: "prod-b", Quantity: 2},
				{ProductID: "prod-a", Quantity: 1},
				{ProductID: "prod-b", Quantity: 3},
			},
			want: []domain.CreateOrderItem{
				{ProductID: "prod-b", Quantity: 5},
				{ProductID: "prod-a", Quantity: 1},
			},
		},
		{
			name: "detecta overflow ao somar quantidades",
			items: []dto.CreateOrderItemRequest{
				{ProductID: "prod-a", Quantity: math.MaxInt},
				{ProductID: "prod-a", Quantity: 1},
			},
			wantErr: domain.ErrInvalidOrderQuantity,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := consolidateOrderItems(testCase.items)

			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("erro recebido = %v; esperado = %v", err, testCase.wantErr)
			}
			if testCase.wantErr == nil && !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("itens consolidados = %#v; esperado = %#v", got, testCase.want)
			}
		})
	}
}

// TestOrderServiceCreate verifica a orquestração: consolidação de itens,
// rejeição de pedidos sem itens e propagação de erros do Repository — sempre
// usando o id do usuário autenticado como dono do pedido.
func TestOrderServiceCreate(t *testing.T) {
	authenticatedUser := domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer}

	t.Run("consolida itens e delega ao Repository com o dono autenticado", func(t *testing.T) {
		var gotOwnerID string
		var gotItems []domain.CreateOrderItem
		fakeRepository := &fakeOrderRepository{
			createFunc: func(_ context.Context, ownerID string, items []domain.CreateOrderItem) (*domain.Order, error) {
				gotOwnerID = ownerID
				gotItems = items
				return &domain.Order{
					ID:          "order-1",
					CustomerID:  ownerID,
					Status:      domain.OrderStatusPending,
					TotalAmount: 50,
					Items: []domain.OrderItem{
						{ID: "item-1", ProductID: "prod-a", Quantity: 5, UnitPrice: 10},
					},
				}, nil
			},
		}
		orderService := NewOrderService(fakeRepository)

		response, err := orderService.Create(context.Background(), authenticatedUser, dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{
				{ProductID: "prod-a", Quantity: 2},
				{ProductID: "prod-a", Quantity: 3},
			},
		})

		if err != nil {
			t.Fatalf("Create retornou erro inesperado: %v", err)
		}
		if gotOwnerID != authenticatedUser.ID {
			t.Errorf("ownerID enviado ao Repository = %q; esperado = %q", gotOwnerID, authenticatedUser.ID)
		}
		wantItems := []domain.CreateOrderItem{{ProductID: "prod-a", Quantity: 5}}
		if !reflect.DeepEqual(gotItems, wantItems) {
			t.Errorf("itens enviados ao Repository = %#v; esperado = %#v", gotItems, wantItems)
		}
		if response.ID != "order-1" || len(response.Items) != 1 || response.Items[0].Subtotal != 50 {
			t.Errorf("pedido criado = %#v", response)
		}
	})

	t.Run("rejeita pedido sem itens", func(t *testing.T) {
		repositoryCalls := 0
		fakeRepository := &fakeOrderRepository{
			createFunc: func(context.Context, string, []domain.CreateOrderItem) (*domain.Order, error) {
				repositoryCalls++
				return nil, nil
			},
		}
		orderService := NewOrderService(fakeRepository)

		_, err := orderService.Create(context.Background(), authenticatedUser, dto.CreateOrderRequest{})

		if !errors.Is(err, domain.ErrOrderMustHaveItems) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrOrderMustHaveItems)
		}
		if repositoryCalls != 0 {
			t.Errorf("chamadas ao Repository = %d; esperado = 0", repositoryCalls)
		}
	})

	t.Run("detecta overflow sem persistir", func(t *testing.T) {
		repositoryCalls := 0
		fakeRepository := &fakeOrderRepository{
			createFunc: func(context.Context, string, []domain.CreateOrderItem) (*domain.Order, error) {
				repositoryCalls++
				return nil, nil
			},
		}
		orderService := NewOrderService(fakeRepository)

		_, err := orderService.Create(context.Background(), authenticatedUser, dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{
				{ProductID: "prod-a", Quantity: math.MaxInt},
				{ProductID: "prod-a", Quantity: 1},
			},
		})

		if !errors.Is(err, domain.ErrInvalidOrderQuantity) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrInvalidOrderQuantity)
		}
		if repositoryCalls != 0 {
			t.Errorf("chamadas ao Repository = %d; esperado = 0", repositoryCalls)
		}
	})

	t.Run("propaga erro do Repository", func(t *testing.T) {
		fakeRepository := &fakeOrderRepository{
			createFunc: func(context.Context, string, []domain.CreateOrderItem) (*domain.Order, error) {
				return nil, domain.ErrProductUnavailable
			},
		}
		orderService := NewOrderService(fakeRepository)

		_, err := orderService.Create(context.Background(), authenticatedUser, dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: "prod-a", Quantity: 1}},
		})

		if !errors.Is(err, domain.ErrProductUnavailable) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrProductUnavailable)
		}
	})
}

// TestOrderServiceSearch verifica a paginação (defaults e teto de PageSize)
// e a regra de autorização: customer só vê os próprios pedidos, admin não
// recebe filtro de propriedade.
func TestOrderServiceSearch(t *testing.T) {
	testCases := []struct {
		name           string
		authenticated  domain.AuthenticatedUser
		page           int
		pageSize       int
		total          int64
		wantPage       int
		wantPageSize   int
		wantTotalPages int
		wantCustomerID *string
	}{
		{
			name:           "cliente recebe filtro pelo próprio id e paginação padrão",
			authenticated:  domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer},
			total:          5,
			wantPage:       1,
			wantPageSize:   defaultOrderPageSize,
			wantTotalPages: 1,
			wantCustomerID: strPtr("user-1"),
		},
		{
			name:           "admin não recebe filtro de propriedade e pageSize é limitado",
			authenticated:  domain.AuthenticatedUser{ID: "admin-1", Role: domain.RoleAdmin},
			page:           2,
			pageSize:       500,
			total:          150,
			wantPage:       2,
			wantPageSize:   maxOrderPageSize,
			wantTotalPages: 2,
			wantCustomerID: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var gotFilter repository.OrderFilter
			fakeRepository := &fakeOrderRepository{
				searchFunc: func(_ context.Context, filter repository.OrderFilter) (*repository.OrderSearchResult, error) {
					gotFilter = filter
					return &repository.OrderSearchResult{Orders: nil, Total: testCase.total}, nil
				},
			}
			orderService := NewOrderService(fakeRepository)

			response, err := orderService.Search(context.Background(), testCase.authenticated, testCase.page, testCase.pageSize)

			if err != nil {
				t.Fatalf("Search retornou erro inesperado: %v", err)
			}
			if response.Page != testCase.wantPage || response.PageSize != testCase.wantPageSize {
				t.Errorf("paginação = %#v; esperado page=%d pageSize=%d", response, testCase.wantPage, testCase.wantPageSize)
			}
			if response.TotalPages != testCase.wantTotalPages {
				t.Errorf("totalPages = %d; esperado = %d", response.TotalPages, testCase.wantTotalPages)
			}
			if !equalStringPtr(gotFilter.CustomerID, testCase.wantCustomerID) {
				t.Errorf("filtro CustomerID = %v; esperado = %v", derefOrNil(gotFilter.CustomerID), derefOrNil(testCase.wantCustomerID))
			}
		})
	}
}

// TestOrderServiceFindByID verifica a regra de acesso: admin vê qualquer
// pedido, customer só o próprio, e erros do Repository são propagados.
func TestOrderServiceFindByID(t *testing.T) {
	t.Run("admin acessa pedido de qualquer cliente", func(t *testing.T) {
		order := &domain.Order{ID: "order-1", CustomerID: "user-2", Status: domain.OrderStatusPending}
		fakeRepository := &fakeOrderRepository{
			findByIDFunc: func(context.Context, string) (*domain.Order, error) { return order, nil },
		}
		orderService := NewOrderService(fakeRepository)

		response, err := orderService.FindByID(context.Background(), domain.AuthenticatedUser{ID: "admin-1", Role: domain.RoleAdmin}, "order-1")

		if err != nil {
			t.Fatalf("FindByID retornou erro inesperado: %v", err)
		}
		if response.ID != "order-1" {
			t.Errorf("pedido recebido = %#v", response)
		}
	})

	t.Run("cliente acessa o próprio pedido", func(t *testing.T) {
		order := &domain.Order{ID: "order-1", CustomerID: "user-1", Status: domain.OrderStatusPending}
		fakeRepository := &fakeOrderRepository{
			findByIDFunc: func(context.Context, string) (*domain.Order, error) { return order, nil },
		}
		orderService := NewOrderService(fakeRepository)

		_, err := orderService.FindByID(context.Background(), domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer}, "order-1")

		if err != nil {
			t.Fatalf("FindByID retornou erro inesperado: %v", err)
		}
	})

	t.Run("cliente não pode acessar pedido de outro cliente", func(t *testing.T) {
		order := &domain.Order{ID: "order-1", CustomerID: "user-2", Status: domain.OrderStatusPending}
		fakeRepository := &fakeOrderRepository{
			findByIDFunc: func(context.Context, string) (*domain.Order, error) { return order, nil },
		}
		orderService := NewOrderService(fakeRepository)

		_, err := orderService.FindByID(context.Background(), domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer}, "order-1")

		if !errors.Is(err, domain.ErrOrderAccessDenied) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrOrderAccessDenied)
		}
	})

	t.Run("propaga erro quando pedido não existe", func(t *testing.T) {
		fakeRepository := &fakeOrderRepository{
			findByIDFunc: func(context.Context, string) (*domain.Order, error) { return nil, domain.ErrOrderNotFound },
		}
		orderService := NewOrderService(fakeRepository)

		_, err := orderService.FindByID(context.Background(), domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer}, "order-404")

		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrOrderNotFound)
		}
	})
}

// TestOrderServicePayByID verifica que id e ownerID chegam ao Repository
// exatamente como recebidos, sem lógica adicional na Service, e que erros
// são propagados.
func TestOrderServicePayByID(t *testing.T) {
	t.Run("delega id e ownerID ao Repository", func(t *testing.T) {
		var gotID, gotOwnerID string
		fakeRepository := &fakeOrderRepository{
			payByIDFunc: func(_ context.Context, id string, ownerID string) (*domain.Order, error) {
				gotID = id
				gotOwnerID = ownerID
				return &domain.Order{ID: id, CustomerID: ownerID, Status: domain.OrderStatusPaid}, nil
			},
		}
		orderService := NewOrderService(fakeRepository)

		response, err := orderService.PayByID(context.Background(), domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer}, "order-1")

		if err != nil {
			t.Fatalf("PayByID retornou erro inesperado: %v", err)
		}
		if gotID != "order-1" || gotOwnerID != "user-1" {
			t.Errorf("Repository.PayByID recebeu id=%q ownerID=%q", gotID, gotOwnerID)
		}
		if response.Status != string(domain.OrderStatusPaid) {
			t.Errorf("status do pedido = %q; esperado = %q", response.Status, domain.OrderStatusPaid)
		}
	})

	t.Run("propaga erro do Repository", func(t *testing.T) {
		fakeRepository := &fakeOrderRepository{
			payByIDFunc: func(context.Context, string, string) (*domain.Order, error) {
				return nil, domain.ErrOrderCannotBePaid
			},
		}
		orderService := NewOrderService(fakeRepository)

		_, err := orderService.PayByID(context.Background(), domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer}, "order-1")

		if !errors.Is(err, domain.ErrOrderCannotBePaid) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrOrderCannotBePaid)
		}
	})
}

// TestOrderServiceCancelByID verifica que isAdmin é derivado do papel do
// usuário autenticado e repassado corretamente ao Repository.
func TestOrderServiceCancelByID(t *testing.T) {
	testCases := []struct {
		name          string
		authenticated domain.AuthenticatedUser
		wantIsAdmin   bool
	}{
		{
			name:          "cliente cancela com isAdmin=false",
			authenticated: domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer},
			wantIsAdmin:   false,
		},
		{
			name:          "admin cancela com isAdmin=true",
			authenticated: domain.AuthenticatedUser{ID: "admin-1", Role: domain.RoleAdmin},
			wantIsAdmin:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var gotRequesterID string
			var gotIsAdmin bool
			fakeRepository := &fakeOrderRepository{
				cancelByIDFunc: func(_ context.Context, id string, requesterID string, isAdmin bool) (*domain.Order, error) {
					gotRequesterID = requesterID
					gotIsAdmin = isAdmin
					return &domain.Order{ID: id, CustomerID: requesterID, Status: domain.OrderStatusCanceled}, nil
				},
			}
			orderService := NewOrderService(fakeRepository)

			_, err := orderService.CancelByID(context.Background(), testCase.authenticated, "order-1")

			if err != nil {
				t.Fatalf("CancelByID retornou erro inesperado: %v", err)
			}
			if gotRequesterID != testCase.authenticated.ID {
				t.Errorf("requesterID enviado ao Repository = %q; esperado = %q", gotRequesterID, testCase.authenticated.ID)
			}
			if gotIsAdmin != testCase.wantIsAdmin {
				t.Errorf("isAdmin enviado ao Repository = %v; esperado = %v", gotIsAdmin, testCase.wantIsAdmin)
			}
		})
	}

	t.Run("propaga erro do Repository", func(t *testing.T) {
		fakeRepository := &fakeOrderRepository{
			cancelByIDFunc: func(context.Context, string, string, bool) (*domain.Order, error) {
				return nil, domain.ErrOrderCannotBeCanceled
			},
		}
		orderService := NewOrderService(fakeRepository)

		_, err := orderService.CancelByID(context.Background(), domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer}, "order-1")

		if !errors.Is(err, domain.ErrOrderCannotBeCanceled) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrOrderCannotBeCanceled)
		}
	})
}

func strPtr(value string) *string {
	return &value
}

func equalStringPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func derefOrNil(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

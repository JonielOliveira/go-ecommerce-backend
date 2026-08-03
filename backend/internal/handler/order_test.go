package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

// fakeOrderService substitui a Service real diretamente — diferente de
// Product/User, OrderHandler depende da interface service.OrderService, o
// que permite injetar uma Service falsa exatamente como no projeto go-tests.
type fakeOrderService struct {
	createFunc     func(ctx context.Context, user domain.AuthenticatedUser, request dto.CreateOrderRequest) (dto.OrderResponse, error)
	searchFunc     func(ctx context.Context, user domain.AuthenticatedUser, page int, pageSize int) (dto.OrderPageResponse, error)
	findByIDFunc   func(ctx context.Context, user domain.AuthenticatedUser, orderID string) (dto.OrderResponse, error)
	payByIDFunc    func(ctx context.Context, user domain.AuthenticatedUser, orderID string) (dto.OrderResponse, error)
	cancelByIDFunc func(ctx context.Context, user domain.AuthenticatedUser, orderID string) (dto.OrderResponse, error)
}

func (fake *fakeOrderService) Create(ctx context.Context, user domain.AuthenticatedUser, request dto.CreateOrderRequest) (dto.OrderResponse, error) {
	return fake.createFunc(ctx, user, request)
}

func (fake *fakeOrderService) Search(ctx context.Context, user domain.AuthenticatedUser, page int, pageSize int) (dto.OrderPageResponse, error) {
	return fake.searchFunc(ctx, user, page, pageSize)
}

func (fake *fakeOrderService) FindByID(ctx context.Context, user domain.AuthenticatedUser, orderID string) (dto.OrderResponse, error) {
	return fake.findByIDFunc(ctx, user, orderID)
}

func (fake *fakeOrderService) PayByID(ctx context.Context, user domain.AuthenticatedUser, orderID string) (dto.OrderResponse, error) {
	return fake.payByIDFunc(ctx, user, orderID)
}

func (fake *fakeOrderService) CancelByID(ctx context.Context, user domain.AuthenticatedUser, orderID string) (dto.OrderResponse, error) {
	return fake.cancelByIDFunc(ctx, user, orderID)
}

var testCustomer = domain.AuthenticatedUser{ID: "user-1", Name: "Ana Souza", Email: "ana@example.com", Role: domain.RoleCustomer}

const validOrderID = "019535d9-3df7-7001-8000-000000000001"

var validCreateOrderRequest = dto.CreateOrderRequest{
	Items: []dto.CreateOrderItemRequest{{ProductID: "019535d9-3df7-7001-8000-000000000099", Quantity: 2}},
}

// TestOrderHandlerCreate verifica: 401 sem usuário autenticado (o dono do
// pedido vem só do contexto), 400 para corpo inválido, o repasse do usuário
// autenticado à Service, e o mapeamento de erros de domínio para status HTTP.
func TestOrderHandlerCreate(t *testing.T) {
	t.Run("sem usuário autenticado retorna 401", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/orders", validCreateOrderRequest)

		serve(context, orderHandler.Create)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
	})

	t.Run("corpo inválido retorna 400", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodPost, "/api/v1/orders", []byte("{invalid"))
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.Create)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("cria pedido para o usuário autenticado e retorna 201", func(t *testing.T) {
		var gotUser domain.AuthenticatedUser
		fakeService := &fakeOrderService{
			createFunc: func(_ context.Context, user domain.AuthenticatedUser, _ dto.CreateOrderRequest) (dto.OrderResponse, error) {
				gotUser = user
				return dto.OrderResponse{ID: "order-1", CustomerID: user.ID}, nil
			},
		}
		orderHandler := NewOrderHandler(fakeService)

		context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/orders", validCreateOrderRequest)
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.Create)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body)
		}
		if gotUser.ID != testCustomer.ID {
			t.Errorf("usuário enviado à Service = %#v", gotUser)
		}
	})

	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "pedido sem itens retorna 400", serviceErr: domain.ErrOrderMustHaveItems, wantStatus: http.StatusBadRequest},
		{name: "quantidade inválida retorna 400", serviceErr: domain.ErrInvalidOrderQuantity, wantStatus: http.StatusBadRequest},
		{name: "produto não encontrado retorna 404", serviceErr: domain.ErrProductNotFound, wantStatus: http.StatusNotFound},
		{name: "produto indisponível retorna 409", serviceErr: domain.ErrProductUnavailable, wantStatus: http.StatusConflict},
		{name: "estoque insuficiente retorna 409", serviceErr: domain.ErrInsufficientStock, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeOrderService{
				createFunc: func(context.Context, domain.AuthenticatedUser, dto.CreateOrderRequest) (dto.OrderResponse, error) {
					return dto.OrderResponse{}, testCase.serviceErr
				},
			}
			orderHandler := NewOrderHandler(fakeService)

			context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/orders", validCreateOrderRequest)
			setAuthenticatedUser(context, testCustomer)

			serve(context, orderHandler.Create)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestOrderHandlerSearch verifica 401 sem usuário, 400 para parâmetros
// inválidos, o repasse do usuário autenticado e a propagação de erro.
func TestOrderHandlerSearch(t *testing.T) {
	t.Run("sem usuário autenticado retorna 401", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodGet, "/api/v1/orders", nil)

		serve(context, orderHandler.Search)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
	})

	t.Run("parâmetros inválidos retornam 400", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodGet, "/api/v1/orders?page=abc", nil)
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.Search)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("busca pedidos do usuário autenticado e retorna 200", func(t *testing.T) {
		var gotUser domain.AuthenticatedUser
		fakeService := &fakeOrderService{
			searchFunc: func(_ context.Context, user domain.AuthenticatedUser, _ int, _ int) (dto.OrderPageResponse, error) {
				gotUser = user
				return dto.OrderPageResponse{}, nil
			},
		}
		orderHandler := NewOrderHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/orders?page=1&pageSize=10", nil)
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.Search)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusOK, recorder.Body)
		}
		if gotUser.ID != testCustomer.ID {
			t.Errorf("usuário enviado à Service = %#v", gotUser)
		}
	})

	t.Run("erro da Service retorna 500", func(t *testing.T) {
		fakeService := &fakeOrderService{
			searchFunc: func(context.Context, domain.AuthenticatedUser, int, int) (dto.OrderPageResponse, error) {
				return dto.OrderPageResponse{}, errors.New("db indisponível")
			},
		}
		orderHandler := NewOrderHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/orders", nil)
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.Search)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusInternalServerError)
		}
	})
}

// TestOrderHandlerFindByID verifica 401 sem usuário, 400 para id fora do
// formato UUID, e que "não encontrado" e "acesso negado" produzem a mesma
// resposta 404 — evitando vazar se o pedido existe e pertence a outro usuário.
func TestOrderHandlerFindByID(t *testing.T) {
	t.Run("sem usuário autenticado retorna 401", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodGet, "/api/v1/orders/"+validOrderID, nil)
		setIDParam(context, validOrderID)

		serve(context, orderHandler.FindByID)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
	})

	t.Run("id com formato inválido retorna 400", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodGet, "/api/v1/orders/not-a-uuid", nil)
		setIDParam(context, "not-a-uuid")
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.FindByID)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("retorna pedido e 200", func(t *testing.T) {
		fakeService := &fakeOrderService{
			findByIDFunc: func(context.Context, domain.AuthenticatedUser, string) (dto.OrderResponse, error) {
				return dto.OrderResponse{ID: validOrderID}, nil
			},
		}
		orderHandler := NewOrderHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/orders/"+validOrderID, nil)
		setIDParam(context, validOrderID)
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.FindByID)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusOK, recorder.Body)
		}
	})

	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "pedido não encontrado retorna 404", serviceErr: domain.ErrOrderNotFound, wantStatus: http.StatusNotFound},
		{name: "acesso negado também retorna 404", serviceErr: domain.ErrOrderAccessDenied, wantStatus: http.StatusNotFound},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeOrderService{
				findByIDFunc: func(context.Context, domain.AuthenticatedUser, string) (dto.OrderResponse, error) {
					return dto.OrderResponse{}, testCase.serviceErr
				},
			}
			orderHandler := NewOrderHandler(fakeService)

			context, recorder := newTestContext(http.MethodGet, "/api/v1/orders/"+validOrderID, nil)
			setIDParam(context, validOrderID)
			setAuthenticatedUser(context, testCustomer)

			serve(context, orderHandler.FindByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestOrderHandlerPayByID verifica 401, 400 (id inválido), sucesso e o
// mapeamento de erros de domínio, incluindo o conflito de status.
func TestOrderHandlerPayByID(t *testing.T) {
	t.Run("sem usuário autenticado retorna 401", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodPost, "/api/v1/orders/"+validOrderID+"/pay", nil)
		setIDParam(context, validOrderID)

		serve(context, orderHandler.PayByID)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
	})

	t.Run("id com formato inválido retorna 400", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodPost, "/api/v1/orders/not-a-uuid/pay", nil)
		setIDParam(context, "not-a-uuid")
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.PayByID)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
	})

	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "paga pedido e retorna 200", wantStatus: http.StatusOK},
		{name: "pedido não encontrado retorna 404", serviceErr: domain.ErrOrderNotFound, wantStatus: http.StatusNotFound},
		{name: "acesso negado também retorna 404", serviceErr: domain.ErrOrderAccessDenied, wantStatus: http.StatusNotFound},
		{name: "pedido não pode ser pago retorna 409", serviceErr: domain.ErrOrderCannotBePaid, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeOrderService{
				payByIDFunc: func(context.Context, domain.AuthenticatedUser, string) (dto.OrderResponse, error) {
					return dto.OrderResponse{ID: validOrderID}, testCase.serviceErr
				},
			}
			orderHandler := NewOrderHandler(fakeService)

			context, recorder := newTestContext(http.MethodPost, "/api/v1/orders/"+validOrderID+"/pay", nil)
			setIDParam(context, validOrderID)
			setAuthenticatedUser(context, testCustomer)

			serve(context, orderHandler.PayByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestOrderHandlerCancelByID espelha TestOrderHandlerPayByID, trocando o
// conflito específico por ErrOrderCannotBeCanceled.
func TestOrderHandlerCancelByID(t *testing.T) {
	t.Run("sem usuário autenticado retorna 401", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodPost, "/api/v1/orders/"+validOrderID+"/cancel", nil)
		setIDParam(context, validOrderID)

		serve(context, orderHandler.CancelByID)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
	})

	t.Run("id com formato inválido retorna 400", func(t *testing.T) {
		orderHandler := NewOrderHandler(&fakeOrderService{})

		context, recorder := newTestContext(http.MethodPost, "/api/v1/orders/not-a-uuid/cancel", nil)
		setIDParam(context, "not-a-uuid")
		setAuthenticatedUser(context, testCustomer)

		serve(context, orderHandler.CancelByID)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
	})

	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "cancela pedido e retorna 200", wantStatus: http.StatusOK},
		{name: "pedido não encontrado retorna 404", serviceErr: domain.ErrOrderNotFound, wantStatus: http.StatusNotFound},
		{name: "acesso negado também retorna 404", serviceErr: domain.ErrOrderAccessDenied, wantStatus: http.StatusNotFound},
		{name: "pedido não pode ser cancelado retorna 409", serviceErr: domain.ErrOrderCannotBeCanceled, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeOrderService{
				cancelByIDFunc: func(context.Context, domain.AuthenticatedUser, string) (dto.OrderResponse, error) {
					return dto.OrderResponse{ID: validOrderID}, testCase.serviceErr
				},
			}
			orderHandler := NewOrderHandler(fakeService)

			context, recorder := newTestContext(http.MethodPost, "/api/v1/orders/"+validOrderID+"/cancel", nil)
			setIDParam(context, validOrderID)
			setAuthenticatedUser(context, testCustomer)

			serve(context, orderHandler.CancelByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

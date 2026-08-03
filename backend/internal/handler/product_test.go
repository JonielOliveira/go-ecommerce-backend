package handler

import (
	"errors"
	"net/http"
	"testing"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

// fakeProductService substitui a Service real diretamente — ProductService é
// uma interface, então o teste de Handler não precisa mais montar uma
// Service real sobre um Repository falso; a Service falsa isola o Handler
// por completo, testando só binding e mapeamento de status.
type fakeProductService struct {
	createFunc         func(request dto.ProductRequest) (dto.ProductResponse, error)
	updateFunc         func(id string, request dto.ProductUpdateRequest) (dto.ProductResponse, error)
	findByIDFunc       func(id string) (dto.ProductResponse, error)
	searchFunc         func(filter dto.ProductSearchRequest) (dto.ProductPageResponse, error)
	deleteByIDFunc     func(id string) error
	restoreByIDFunc    func(id string) error
	activateByIDFunc   func(id string) error
	deactivateByIDFunc func(id string) error
}

func (fake *fakeProductService) Create(request dto.ProductRequest) (dto.ProductResponse, error) {
	return fake.createFunc(request)
}

func (fake *fakeProductService) Update(id string, request dto.ProductUpdateRequest) (dto.ProductResponse, error) {
	return fake.updateFunc(id, request)
}

func (fake *fakeProductService) FindByID(id string) (dto.ProductResponse, error) {
	return fake.findByIDFunc(id)
}

func (fake *fakeProductService) Search(filter dto.ProductSearchRequest) (dto.ProductPageResponse, error) {
	return fake.searchFunc(filter)
}

func (fake *fakeProductService) DeleteByID(id string) error {
	return fake.deleteByIDFunc(id)
}

func (fake *fakeProductService) RestoreByID(id string) error {
	return fake.restoreByIDFunc(id)
}

func (fake *fakeProductService) ActivateByID(id string) error {
	return fake.activateByIDFunc(id)
}

func (fake *fakeProductService) DeactivateByID(id string) error {
	return fake.deactivateByIDFunc(id)
}

var validProductRequest = dto.ProductRequest{
	Name:        "Teclado Mecânico",
	Description: "RGB, switches azuis",
	Price:       199.90,
	Stock:       10,
}

var validProductUpdateRequest = dto.ProductUpdateRequest{
	Name:        "Novo nome",
	Description: "Nova descrição",
	Price:       250,
	Stock:       3,
}

// TestProductHandlerCreate verifica o binding do corpo da requisição e o
// repasse à Service — sem switch de erro nesse endpoint, qualquer falha da
// Service vira 400.
func TestProductHandlerCreate(t *testing.T) {
	t.Run("cria produto e retorna 201", func(t *testing.T) {
		fakeService := &fakeProductService{
			createFunc: func(dto.ProductRequest) (dto.ProductResponse, error) {
				return dto.ProductResponse{ID: "prod-1"}, nil
			},
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/products", validProductRequest)

		serve(context, productHandler.Create)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body)
		}
		got := decodeJSONBody[dto.ProductResponse](t, recorder)
		if got.ID != "prod-1" {
			t.Errorf("produto criado = %#v", got)
		}
	})

	t.Run("corpo inválido retorna 400 sem chamar a Service", func(t *testing.T) {
		serviceCalls := 0
		fakeService := &fakeProductService{
			createFunc: func(dto.ProductRequest) (dto.ProductResponse, error) {
				serviceCalls++
				return dto.ProductResponse{}, nil
			},
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newTestContext(http.MethodPost, "/api/v1/products", []byte("{invalid"))

		serve(context, productHandler.Create)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
		if serviceCalls != 0 {
			t.Errorf("chamadas à Service = %d; esperado = 0", serviceCalls)
		}
	})

	t.Run("erro da Service retorna 400", func(t *testing.T) {
		fakeService := &fakeProductService{
			createFunc: func(dto.ProductRequest) (dto.ProductResponse, error) {
				return dto.ProductResponse{}, domain.ErrInvalidProductName
			},
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/products", validProductRequest)

		serve(context, productHandler.Create)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusBadRequest, recorder.Body)
		}
	})
}

// TestProductHandlerUpdate verifica o mapeamento completo do switch de erro:
// não encontrado (404), já removido (409), dados inválidos (400, em cada
// variante) e erro inesperado (500) — além do 400 de corpo malformado, que
// nem chega a chamar a Service.
func TestProductHandlerUpdate(t *testing.T) {
	t.Run("corpo inválido retorna 400 sem chamar a Service", func(t *testing.T) {
		serviceCalls := 0
		fakeService := &fakeProductService{
			updateFunc: func(string, dto.ProductUpdateRequest) (dto.ProductResponse, error) {
				serviceCalls++
				return dto.ProductResponse{}, nil
			},
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newTestContext(http.MethodPut, "/api/v1/products/prod-1", []byte("{invalid"))
		setIDParam(context, "prod-1")

		serve(context, productHandler.Update)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
		if serviceCalls != 0 {
			t.Errorf("chamadas à Service = %d; esperado = 0", serviceCalls)
		}
	})

	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "atualiza produto e retorna 200", wantStatus: http.StatusOK},
		{name: "produto não encontrado retorna 404", serviceErr: domain.ErrProductNotFound, wantStatus: http.StatusNotFound},
		{name: "produto removido retorna 409", serviceErr: domain.ErrProductAlreadyDeleted, wantStatus: http.StatusConflict},
		{name: "nome inválido retorna 400", serviceErr: domain.ErrInvalidProductName, wantStatus: http.StatusBadRequest},
		{name: "descrição inválida retorna 400", serviceErr: domain.ErrInvalidProductDescription, wantStatus: http.StatusBadRequest},
		{name: "preço inválido retorna 400", serviceErr: domain.ErrInvalidProductPrice, wantStatus: http.StatusBadRequest},
		{name: "estoque inválido retorna 400", serviceErr: domain.ErrInvalidProductStock, wantStatus: http.StatusBadRequest},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeProductService{
				updateFunc: func(id string, _ dto.ProductUpdateRequest) (dto.ProductResponse, error) {
					return dto.ProductResponse{ID: id}, testCase.serviceErr
				},
			}
			productHandler := NewProductHandler(fakeService)

			context, recorder := newJSONTestContext(t, http.MethodPut, "/api/v1/products/prod-1", validProductUpdateRequest)
			setIDParam(context, "prod-1")

			serve(context, productHandler.Update)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestProductHandlerFindByID verifica a rota pública de leitura: sucesso,
// não encontrado e erro inesperado da Service.
func TestProductHandlerFindByID(t *testing.T) {
	t.Run("retorna produto e 200", func(t *testing.T) {
		fakeService := &fakeProductService{
			findByIDFunc: func(id string) (dto.ProductResponse, error) { return dto.ProductResponse{ID: id}, nil },
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/products/prod-1", nil)
		setIDParam(context, "prod-1")

		serve(context, productHandler.FindByID)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusOK)
		}
		got := decodeJSONBody[dto.ProductResponse](t, recorder)
		if got.ID != "prod-1" {
			t.Errorf("produto recebido = %#v", got)
		}
	})

	t.Run("produto não encontrado retorna 404", func(t *testing.T) {
		fakeService := &fakeProductService{
			findByIDFunc: func(string) (dto.ProductResponse, error) { return dto.ProductResponse{}, domain.ErrProductNotFound },
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/products/prod-404", nil)
		setIDParam(context, "prod-404")

		serve(context, productHandler.FindByID)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusNotFound)
		}
	})

	t.Run("erro inesperado retorna 500", func(t *testing.T) {
		fakeService := &fakeProductService{
			findByIDFunc: func(string) (dto.ProductResponse, error) {
				return dto.ProductResponse{}, errors.New("db indisponível")
			},
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/products/prod-1", nil)
		setIDParam(context, "prod-1")

		serve(context, productHandler.FindByID)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusInternalServerError)
		}
	})
}

// TestProductHandlerSearch verifica o binding dos parâmetros de query e o
// repasse de erros da Service.
func TestProductHandlerSearch(t *testing.T) {
	t.Run("retorna página de produtos e 200", func(t *testing.T) {
		fakeService := &fakeProductService{
			searchFunc: func(dto.ProductSearchRequest) (dto.ProductPageResponse, error) {
				return dto.ProductPageResponse{}, nil
			},
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/products?page=1&pageSize=10", nil)

		serve(context, productHandler.Search)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusOK, recorder.Body)
		}
	})

	t.Run("parâmetros inválidos retornam 400 sem chamar a Service", func(t *testing.T) {
		serviceCalls := 0
		fakeService := &fakeProductService{
			searchFunc: func(dto.ProductSearchRequest) (dto.ProductPageResponse, error) {
				serviceCalls++
				return dto.ProductPageResponse{}, nil
			},
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/products?page=abc", nil)

		serve(context, productHandler.Search)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
		if serviceCalls != 0 {
			t.Errorf("chamadas à Service = %d; esperado = 0", serviceCalls)
		}
	})

	t.Run("erro da Service retorna 500", func(t *testing.T) {
		fakeService := &fakeProductService{
			searchFunc: func(dto.ProductSearchRequest) (dto.ProductPageResponse, error) {
				return dto.ProductPageResponse{}, errors.New("db indisponível")
			},
		}
		productHandler := NewProductHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/products", nil)

		serve(context, productHandler.Search)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusInternalServerError)
		}
	})
}

// TestProductHandlerDeleteByID cobre o mapeamento completo de status desse
// endpoint administrativo: 204, 404, 409 e 500.
func TestProductHandlerDeleteByID(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "sucesso retorna 204", wantStatus: http.StatusNoContent},
		{name: "produto não encontrado retorna 404", serviceErr: domain.ErrProductNotFound, wantStatus: http.StatusNotFound},
		{name: "produto já removido retorna 409", serviceErr: domain.ErrProductAlreadyDeleted, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeProductService{
				deleteByIDFunc: func(string) error { return testCase.serviceErr },
			}
			productHandler := NewProductHandler(fakeService)

			context, recorder := newTestContext(http.MethodDelete, "/api/v1/products/prod-1", nil)
			setIDParam(context, "prod-1")

			serve(context, productHandler.DeleteByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestProductHandlerRestoreByID cobre 204, 404, 409 (não removido) e 500.
func TestProductHandlerRestoreByID(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "sucesso retorna 204", wantStatus: http.StatusNoContent},
		{name: "produto não encontrado retorna 404", serviceErr: domain.ErrProductNotFound, wantStatus: http.StatusNotFound},
		{name: "produto não removido retorna 409", serviceErr: domain.ErrProductNotDeleted, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeProductService{
				restoreByIDFunc: func(string) error { return testCase.serviceErr },
			}
			productHandler := NewProductHandler(fakeService)

			context, recorder := newTestContext(http.MethodPatch, "/api/v1/products/prod-1/restore", nil)
			setIDParam(context, "prod-1")

			serve(context, productHandler.RestoreByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestProductHandlerActivateByID cobre 204, 404, 409 (removido ou já ativo) e 500.
func TestProductHandlerActivateByID(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "sucesso retorna 204", wantStatus: http.StatusNoContent},
		{name: "produto não encontrado retorna 404", serviceErr: domain.ErrProductNotFound, wantStatus: http.StatusNotFound},
		{name: "produto removido retorna 409", serviceErr: domain.ErrProductAlreadyDeleted, wantStatus: http.StatusConflict},
		{name: "produto já ativo retorna 409", serviceErr: domain.ErrProductAlreadyActive, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeProductService{
				activateByIDFunc: func(string) error { return testCase.serviceErr },
			}
			productHandler := NewProductHandler(fakeService)

			context, recorder := newTestContext(http.MethodPatch, "/api/v1/products/prod-1/activate", nil)
			setIDParam(context, "prod-1")

			serve(context, productHandler.ActivateByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestProductHandlerDeactivateByID cobre 204, 404, 409 (removido ou já
// inativo) e 500.
func TestProductHandlerDeactivateByID(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "sucesso retorna 204", wantStatus: http.StatusNoContent},
		{name: "produto não encontrado retorna 404", serviceErr: domain.ErrProductNotFound, wantStatus: http.StatusNotFound},
		{name: "produto removido retorna 409", serviceErr: domain.ErrProductAlreadyDeleted, wantStatus: http.StatusConflict},
		{name: "produto já inativo retorna 409", serviceErr: domain.ErrProductAlreadyInactive, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeProductService{
				deactivateByIDFunc: func(string) error { return testCase.serviceErr },
			}
			productHandler := NewProductHandler(fakeService)

			context, recorder := newTestContext(http.MethodPatch, "/api/v1/products/prod-1/deactivate", nil)
			setIDParam(context, "prod-1")

			serve(context, productHandler.DeactivateByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

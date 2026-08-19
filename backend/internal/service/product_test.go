package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/repository"
)

// fakeProductRepository substitui o PostgreSQL e permite controlar cada
// operação no teste. Os métodos abaixo fazem este tipo satisfazer a
// interface repository.ProductRepository.
type fakeProductRepository struct {
	createFunc         func(*domain.Product) (*domain.Product, error)
	updateFunc         func(*domain.Product) (*domain.Product, error)
	findByIDFunc       func(string) (*domain.Product, error)
	searchFunc         func(repository.ProductSearchFilter) (*repository.ProductSearchResult, error)
	deleteByIDFunc     func(string) error
	restoreByIDFunc    func(string) error
	activateByIDFunc   func(string) error
	deactivateByIDFunc func(string) error
}

func (fake *fakeProductRepository) Create(_ context.Context, product *domain.Product) (*domain.Product, error) {
	return fake.createFunc(product)
}

func (fake *fakeProductRepository) Update(_ context.Context, product *domain.Product) (*domain.Product, error) {
	return fake.updateFunc(product)
}

func (fake *fakeProductRepository) FindByID(_ context.Context, id string) (*domain.Product, error) {
	return fake.findByIDFunc(id)
}

func (fake *fakeProductRepository) Search(_ context.Context, filter repository.ProductSearchFilter) (*repository.ProductSearchResult, error) {
	return fake.searchFunc(filter)
}

func (fake *fakeProductRepository) DeleteByID(_ context.Context, id string) error {
	return fake.deleteByIDFunc(id)
}

func (fake *fakeProductRepository) RestoreByID(_ context.Context, id string) error {
	return fake.restoreByIDFunc(id)
}

func (fake *fakeProductRepository) ActivateByID(_ context.Context, id string) error {
	return fake.activateByIDFunc(id)
}

func (fake *fakeProductRepository) DeactivateByID(_ context.Context, id string) error {
	return fake.deactivateByIDFunc(id)
}

// mustNewProduct monta um produto ativo e não removido, com o id informado,
// pronto para ser devolvido pelos fakes de Repository.
func mustNewProduct(t *testing.T, id string) *domain.Product {
	t.Helper()

	product, err := domain.RestoreProduct(
		id,
		"Teclado Mecânico",
		"RGB, switches azuis",
		199.90,
		10,
		nil,
		nil,
		true,
		time.Now(),
		time.Now(),
		nil,
	)
	if err != nil {
		t.Fatalf("montar produto de teste: %v", err)
	}

	return product
}

// mustDeletedProduct monta um produto já removido (soft delete), usado para
// verificar que a Service recusa atualizações sobre ele.
func mustDeletedProduct(t *testing.T, id string) *domain.Product {
	t.Helper()

	deletedAt := time.Now()
	product, err := domain.RestoreProduct(
		id,
		"Teclado Mecânico",
		"RGB, switches azuis",
		199.90,
		10,
		nil,
		nil,
		true,
		time.Now(),
		time.Now(),
		&deletedAt,
	)
	if err != nil {
		t.Fatalf("montar produto removido de teste: %v", err)
	}

	return product
}

// TestProductServiceCreate verifica a orquestração: a Service só delega ao
// Repository quando o Model aceita os dados recebidos.
func TestProductServiceCreate(t *testing.T) {
	testCases := []struct {
		name                string
		request             dto.ProductRequest
		wantErr             error
		wantRepositoryCalls int
	}{
		{
			name: "cria produto válido",
			request: dto.ProductRequest{
				Name:        "Teclado Mecânico",
				Description: "RGB, switches azuis",
				Price:       199.90,
				Stock:       10,
			},
			wantRepositoryCalls: 1,
		},
		{
			name: "rejeita nome vazio",
			request: dto.ProductRequest{
				Name:        "   ",
				Description: "RGB, switches azuis",
				Price:       199.90,
				Stock:       10,
			},
			wantErr: domain.ErrInvalidProductName,
		},
		{
			name: "rejeita preço inválido",
			request: dto.ProductRequest{
				Name:        "Teclado Mecânico",
				Description: "RGB, switches azuis",
				Price:       -5,
				Stock:       10,
			},
			wantErr: domain.ErrInvalidProductPrice,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repositoryCalls := 0
			fakeRepository := &fakeProductRepository{
				createFunc: func(_ *domain.Product) (*domain.Product, error) {
					repositoryCalls++
					return mustNewProduct(t, "prod-1"), nil
				},
			}
			productService := NewProductService(fakeRepository)

			response, err := productService.Create(context.Background(), testCase.request)

			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("erro recebido = %v; esperado = %v", err, testCase.wantErr)
			}
			if repositoryCalls != testCase.wantRepositoryCalls {
				t.Errorf("chamadas ao Repository = %d; esperado = %d", repositoryCalls, testCase.wantRepositoryCalls)
			}
			if testCase.wantErr == nil && response.ID != "prod-1" {
				t.Errorf("produto criado = %#v; ID esperado = %q", response, "prod-1")
			}
		})
	}
}

// TestProductServiceUpdate verifica as regras aplicadas antes de persistir:
// produto removido é recusado e dados inválidos nunca chegam ao Repository.
func TestProductServiceUpdate(t *testing.T) {
	t.Run("atualiza produto existente", func(t *testing.T) {
		existing := mustNewProduct(t, "prod-1")
		updateCalls := 0
		fakeRepository := &fakeProductRepository{
			findByIDFunc: func(id string) (*domain.Product, error) {
				if id != "prod-1" {
					t.Fatalf("id recebido = %s; esperado = prod-1", id)
				}
				return existing, nil
			},
			updateFunc: func(product *domain.Product) (*domain.Product, error) {
				updateCalls++
				return product, nil
			},
		}
		productService := NewProductService(fakeRepository)

		response, err := productService.Update(context.Background(), "prod-1", dto.ProductUpdateRequest{
			Name:        "Teclado Mecânico V2",
			Description: "RGB, switches vermelhos",
			Price:       249.90,
			Stock:       5,
		})

		if err != nil {
			t.Fatalf("Update retornou erro inesperado: %v", err)
		}
		if updateCalls != 1 {
			t.Errorf("chamadas ao Repository.Update = %d; esperado = 1", updateCalls)
		}
		if response.Name != "Teclado Mecânico V2" || response.Price != 249.90 || response.Stock != 5 {
			t.Errorf("produto atualizado = %#v", response)
		}
	})

	t.Run("recusa atualizar produto removido", func(t *testing.T) {
		deleted := mustDeletedProduct(t, "prod-2")
		updateCalls := 0
		fakeRepository := &fakeProductRepository{
			findByIDFunc: func(string) (*domain.Product, error) { return deleted, nil },
			updateFunc: func(product *domain.Product) (*domain.Product, error) {
				updateCalls++
				return product, nil
			},
		}
		productService := NewProductService(fakeRepository)

		_, err := productService.Update(context.Background(), "prod-2", dto.ProductUpdateRequest{
			Name:        "Novo nome",
			Description: "Nova descrição",
			Price:       10,
			Stock:       1,
		})

		if !errors.Is(err, domain.ErrProductAlreadyDeleted) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrProductAlreadyDeleted)
		}
		if updateCalls != 0 {
			t.Errorf("chamadas ao Repository.Update = %d; esperado = 0", updateCalls)
		}
	})

	t.Run("recusa dados inválidos sem persistir", func(t *testing.T) {
		existing := mustNewProduct(t, "prod-3")
		updateCalls := 0
		fakeRepository := &fakeProductRepository{
			findByIDFunc: func(string) (*domain.Product, error) { return existing, nil },
			updateFunc: func(product *domain.Product) (*domain.Product, error) {
				updateCalls++
				return product, nil
			},
		}
		productService := NewProductService(fakeRepository)

		_, err := productService.Update(context.Background(), "prod-3", dto.ProductUpdateRequest{
			Name:        "Teclado",
			Description: "Descrição",
			Price:       0,
			Stock:       1,
		})

		if !errors.Is(err, domain.ErrInvalidProductPrice) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrInvalidProductPrice)
		}
		if updateCalls != 0 {
			t.Errorf("chamadas ao Repository.Update = %d; esperado = 0", updateCalls)
		}
	})

	t.Run("propaga erro quando produto não existe", func(t *testing.T) {
		fakeRepository := &fakeProductRepository{
			findByIDFunc: func(string) (*domain.Product, error) {
				return nil, domain.ErrProductNotFound
			},
		}
		productService := NewProductService(fakeRepository)

		_, err := productService.Update(context.Background(), "prod-404", dto.ProductUpdateRequest{
			Name:        "Teclado",
			Description: "Descrição",
			Price:       10,
			Stock:       1,
		})

		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrProductNotFound)
		}
	})
}

// TestProductServiceSearch verifica os valores padrão de paginação, o limite
// máximo de PageSize e o cálculo de TotalPages a partir do total devolvido
// pelo Repository.
func TestProductServiceSearch(t *testing.T) {
	testCases := []struct {
		name           string
		request        dto.ProductSearchRequest
		total          int64
		wantPage       int
		wantPageSize   int
		wantTotalPages int
	}{
		{
			name:           "aplica página e tamanho padrão quando ausentes",
			request:        dto.ProductSearchRequest{},
			total:          25,
			wantPage:       1,
			wantPageSize:   20,
			wantTotalPages: 2,
		},
		{
			name:           "limita tamanho de página a 100",
			request:        dto.ProductSearchRequest{Page: 2, PageSize: 500},
			total:          150,
			wantPage:       2,
			wantPageSize:   100,
			wantTotalPages: 2,
		},
		{
			name:           "sem resultados retorna zero páginas",
			request:        dto.ProductSearchRequest{Page: 1, PageSize: 10},
			total:          0,
			wantPage:       1,
			wantPageSize:   10,
			wantTotalPages: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var gotFilter repository.ProductSearchFilter
			fakeRepository := &fakeProductRepository{
				searchFunc: func(filter repository.ProductSearchFilter) (*repository.ProductSearchResult, error) {
					gotFilter = filter
					return &repository.ProductSearchResult{Products: nil, Total: testCase.total}, nil
				},
			}
			productService := NewProductService(fakeRepository)

			response, err := productService.Search(context.Background(), testCase.request)

			if err != nil {
				t.Fatalf("Search retornou erro inesperado: %v", err)
			}
			if response.Page != testCase.wantPage || response.PageSize != testCase.wantPageSize {
				t.Errorf("paginação = %#v; esperado page=%d pageSize=%d", response, testCase.wantPage, testCase.wantPageSize)
			}
			if response.TotalPages != testCase.wantTotalPages {
				t.Errorf("totalPages = %d; esperado = %d", response.TotalPages, testCase.wantTotalPages)
			}
			wantOffset := (testCase.wantPage - 1) * testCase.wantPageSize
			if gotFilter.Limit != testCase.wantPageSize || gotFilter.Offset != wantOffset {
				t.Errorf("filtro enviado ao Repository = %#v; esperado limit=%d offset=%d", gotFilter, testCase.wantPageSize, wantOffset)
			}
		})
	}
}

// TestMapDeletionFilter verifica a tradução do estado de exclusão do DTO para
// o filtro usado pelo Repository, inclusive o valor padrão.
func TestMapDeletionFilter(t *testing.T) {
	testCases := []struct {
		name  string
		state dto.DeletionState
		want  repository.DeletionFilter
	}{
		{name: "estado vazio usa não removidos", state: "", want: repository.DeletionFilterNotDeleted},
		{name: "removidos", state: dto.DeletionStateDeleted, want: repository.DeletionFilterDeleted},
		{name: "todos", state: dto.DeletionStateAll, want: repository.DeletionFilterAll},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := mapDeletionFilter(testCase.state)
			if got != testCase.want {
				t.Errorf("filtro = %v; esperado = %v", got, testCase.want)
			}
		})
	}
}

// TestProductServiceDelegatesOperations organiza em subtests os casos de uso
// sem regra adicional, apenas repasse ao Repository.
func TestProductServiceDelegatesOperations(t *testing.T) {
	t.Run("busca produto por id", func(t *testing.T) {
		want := mustNewProduct(t, "prod-1")
		fakeRepository := &fakeProductRepository{
			findByIDFunc: func(id string) (*domain.Product, error) {
				if id != "prod-1" {
					t.Fatalf("id recebido = %s; esperado = prod-1", id)
				}
				return want, nil
			},
		}

		got, err := NewProductService(fakeRepository).FindByID(context.Background(), "prod-1")

		if err != nil {
			t.Fatalf("FindByID retornou erro inesperado: %v", err)
		}
		if got.ID != want.ID() {
			t.Errorf("produto recebido = %#v", got)
		}
	})

	t.Run("exclui produto", func(t *testing.T) {
		deletedID := ""
		fakeRepository := &fakeProductRepository{
			deleteByIDFunc: func(id string) error {
				deletedID = id
				return nil
			},
		}

		err := NewProductService(fakeRepository).DeleteByID(context.Background(), "prod-2")

		if err != nil {
			t.Fatalf("DeleteByID retornou erro inesperado: %v", err)
		}
		if deletedID != "prod-2" {
			t.Errorf("id excluído = %s; esperado = prod-2", deletedID)
		}
	})

	t.Run("restaura produto", func(t *testing.T) {
		restoredID := ""
		fakeRepository := &fakeProductRepository{
			restoreByIDFunc: func(id string) error {
				restoredID = id
				return nil
			},
		}

		err := NewProductService(fakeRepository).RestoreByID(context.Background(), "prod-3")

		if err != nil {
			t.Fatalf("RestoreByID retornou erro inesperado: %v", err)
		}
		if restoredID != "prod-3" {
			t.Errorf("id restaurado = %s; esperado = prod-3", restoredID)
		}
	})

	t.Run("ativa produto", func(t *testing.T) {
		activatedID := ""
		fakeRepository := &fakeProductRepository{
			activateByIDFunc: func(id string) error {
				activatedID = id
				return nil
			},
		}

		err := NewProductService(fakeRepository).ActivateByID(context.Background(), "prod-4")

		if err != nil {
			t.Fatalf("ActivateByID retornou erro inesperado: %v", err)
		}
		if activatedID != "prod-4" {
			t.Errorf("id ativado = %s; esperado = prod-4", activatedID)
		}
	})

	t.Run("desativa produto", func(t *testing.T) {
		deactivatedID := ""
		fakeRepository := &fakeProductRepository{
			deactivateByIDFunc: func(id string) error {
				deactivatedID = id
				return nil
			},
		}

		err := NewProductService(fakeRepository).DeactivateByID(context.Background(), "prod-5")

		if err != nil {
			t.Fatalf("DeactivateByID retornou erro inesperado: %v", err)
		}
		if deactivatedID != "prod-5" {
			t.Errorf("id desativado = %s; esperado = prod-5", deactivatedID)
		}
	})
}

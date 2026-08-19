//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ecommerce/internal/domain"
)

// mustCreateProduct cria um produto válido diretamente pelo Repository,
// sem passar por Service/Handler — o nível mais barato da suíte de
// integração, isolando só a camada SQL.
func mustCreateProduct(t *testing.T, repo *PostgresProductRepository, name, description string, price float64, stock int, categoryID *string) *domain.Product {
	t.Helper()

	product, err := domain.NewProduct(name, description, price, stock, categoryID, nil)
	if err != nil {
		t.Fatalf("montar produto de teste %q: %v", name, err)
	}

	created, err := repo.Create(context.Background(), product)
	if err != nil {
		t.Fatalf("criar produto de teste %q: %v", name, err)
	}

	return created
}

func boolPtr(value bool) *bool        { return &value }
func floatPtr(value float64) *float64 { return &value }
func sameSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	wantSet := make(map[string]bool, len(want))
	for _, id := range want {
		wantSet[id] = true
	}
	for _, id := range got {
		if !wantSet[id] {
			return false
		}
	}
	return true
}

// TestPostgresProductRepositoryCreateAndFindByID usa o PostgreSQL real para
// verificar em conjunto SQL, schema, Scan e mapeamento da entidade.
func TestPostgresProductRepositoryCreateAndFindByID(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresProductRepository(db)

	product, err := domain.NewProduct("Teclado Mecânico", "RGB, switches azuis", 199.90, 10, nil, nil)
	if err != nil {
		t.Fatalf("montar produto de teste: %v", err)
	}

	created, err := repo.Create(context.Background(), product)
	if err != nil {
		t.Fatalf("Create retornou erro: %v", err)
	}
	if created.ID() == "" {
		t.Fatalf("id do produto criado está vazio")
	}
	if created.Name() != "Teclado Mecânico" || created.Price() != 199.90 || created.Stock() != 10 {
		t.Errorf("produto criado = %#v", created)
	}
	if !created.IsActive() || created.IsDeleted() {
		t.Errorf("produto recém-criado deveria estar ativo e não removido")
	}

	got, err := repo.FindByID(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("FindByID retornou erro: %v", err)
	}
	if got.ID() != created.ID() || got.Name() != created.Name() || got.Price() != created.Price() {
		t.Errorf("produto lido = %#v; esperado = %#v", got, created)
	}

	_, err = repo.FindByID(context.Background(), "00000000-0000-7000-8000-000000000000")
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("FindByID inexistente = %v; esperado = %v", err, domain.ErrProductNotFound)
	}
}

// TestPostgresProductRepositoryUpdate verifica os dois caminhos de erro que
// dependem da query real: produto inexistente (nenhuma linha em nenhum
// estado) e produto removido (a linha existe, mas deleted_at IS NOT NULL
// filtra o UPDATE) — o Repository distingue os dois consultando o estado
// atual só quando o UPDATE não afeta nenhuma linha.
func TestPostgresProductRepositoryUpdate(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresProductRepository(db)

	t.Run("atualiza produto existente", func(t *testing.T) {
		created := mustCreateProduct(t, repo, "Produto Update", "desc", 50, 5, nil)

		if err := created.Update("Produto Atualizado", "nova desc", 75, 3, nil, nil); err != nil {
			t.Fatalf("Update do domínio: %v", err)
		}

		updated, err := repo.Update(context.Background(), created)
		if err != nil {
			t.Fatalf("Repository.Update retornou erro: %v", err)
		}
		if updated.Name() != "Produto Atualizado" || updated.Price() != 75 || updated.Stock() != 3 {
			t.Errorf("produto atualizado = %#v", updated)
		}
	})

	t.Run("produto inexistente retorna ErrProductNotFound", func(t *testing.T) {
		ghost, err := domain.RestoreProduct("00000000-0000-7000-8000-000000000000", "Fantasma", "desc", 10, 1, nil, nil, true, time.Now(), time.Now(), nil)
		if err != nil {
			t.Fatalf("montar produto fantasma: %v", err)
		}

		_, err = repo.Update(context.Background(), ghost)
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Fatalf("Update inexistente = %v; esperado = %v", err, domain.ErrProductNotFound)
		}
	})

	t.Run("produto removido retorna ErrProductAlreadyDeleted", func(t *testing.T) {
		created := mustCreateProduct(t, repo, "Produto Removido Update", "desc", 20, 2, nil)
		if err := repo.DeleteByID(context.Background(), created.ID()); err != nil {
			t.Fatalf("DeleteByID: %v", err)
		}

		if err := created.Update("Novo Nome", "nova desc", 20, 2, nil, nil); err != nil {
			t.Fatalf("Update do domínio: %v", err)
		}

		_, err := repo.Update(context.Background(), created)
		if !errors.Is(err, domain.ErrProductAlreadyDeleted) {
			t.Fatalf("Update de removido = %v; esperado = %v", err, domain.ErrProductAlreadyDeleted)
		}
	})
}

// TestPostgresProductRepositoryDeleteRestoreActivateDeactivate cobre as
// transições de estado e o mapeamento de erro de cada uma contra o banco
// real: cada método só decide qual erro devolver depois de consultar o
// estado atual, e essa consulta é a parte que só existe aqui, não no
// Repository falso usado pelos testes de Service.
func TestPostgresProductRepositoryDeleteRestoreActivateDeactivate(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresProductRepository(db)

	t.Run("delete e restore", func(t *testing.T) {
		product := mustCreateProduct(t, repo, "Produto Delete", "desc", 10, 1, nil)

		if err := repo.DeleteByID(context.Background(), product.ID()); err != nil {
			t.Fatalf("DeleteByID: %v", err)
		}
		if err := repo.DeleteByID(context.Background(), product.ID()); !errors.Is(err, domain.ErrProductAlreadyDeleted) {
			t.Fatalf("segundo delete = %v; esperado = %v", err, domain.ErrProductAlreadyDeleted)
		}

		if err := repo.RestoreByID(context.Background(), product.ID()); err != nil {
			t.Fatalf("RestoreByID: %v", err)
		}
		if err := repo.RestoreByID(context.Background(), product.ID()); !errors.Is(err, domain.ErrProductNotDeleted) {
			t.Fatalf("segunda restauração = %v; esperado = %v", err, domain.ErrProductNotDeleted)
		}
	})

	t.Run("activate e deactivate", func(t *testing.T) {
		product := mustCreateProduct(t, repo, "Produto Activate", "desc", 10, 1, nil)

		if err := repo.ActivateByID(context.Background(), product.ID()); !errors.Is(err, domain.ErrProductAlreadyActive) {
			t.Fatalf("ativar produto já ativo = %v; esperado = %v", err, domain.ErrProductAlreadyActive)
		}

		if err := repo.DeactivateByID(context.Background(), product.ID()); err != nil {
			t.Fatalf("DeactivateByID: %v", err)
		}
		if err := repo.DeactivateByID(context.Background(), product.ID()); !errors.Is(err, domain.ErrProductAlreadyInactive) {
			t.Fatalf("desativar de novo = %v; esperado = %v", err, domain.ErrProductAlreadyInactive)
		}

		if err := repo.ActivateByID(context.Background(), product.ID()); err != nil {
			t.Fatalf("ActivateByID: %v", err)
		}
	})

	t.Run("ativar/desativar um produto removido reporta a remoção", func(t *testing.T) {
		product := mustCreateProduct(t, repo, "Produto Removido", "desc", 10, 1, nil)
		if err := repo.DeleteByID(context.Background(), product.ID()); err != nil {
			t.Fatalf("DeleteByID: %v", err)
		}

		if err := repo.ActivateByID(context.Background(), product.ID()); !errors.Is(err, domain.ErrProductAlreadyDeleted) {
			t.Fatalf("ativar produto removido = %v; esperado = %v", err, domain.ErrProductAlreadyDeleted)
		}
		if err := repo.DeactivateByID(context.Background(), product.ID()); !errors.Is(err, domain.ErrProductAlreadyDeleted) {
			t.Fatalf("desativar produto removido = %v; esperado = %v", err, domain.ErrProductAlreadyDeleted)
		}
	})

	t.Run("qualquer operação num id inexistente retorna ErrProductNotFound", func(t *testing.T) {
		const fakeID = "00000000-0000-7000-8000-000000000000"

		if err := repo.DeleteByID(context.Background(), fakeID); !errors.Is(err, domain.ErrProductNotFound) {
			t.Errorf("DeleteByID inexistente = %v; esperado = %v", err, domain.ErrProductNotFound)
		}
		if err := repo.RestoreByID(context.Background(), fakeID); !errors.Is(err, domain.ErrProductNotFound) {
			t.Errorf("RestoreByID inexistente = %v; esperado = %v", err, domain.ErrProductNotFound)
		}
		if err := repo.ActivateByID(context.Background(), fakeID); !errors.Is(err, domain.ErrProductNotFound) {
			t.Errorf("ActivateByID inexistente = %v; esperado = %v", err, domain.ErrProductNotFound)
		}
		if err := repo.DeactivateByID(context.Background(), fakeID); !errors.Is(err, domain.ErrProductNotFound) {
			t.Errorf("DeactivateByID inexistente = %v; esperado = %v", err, domain.ErrProductNotFound)
		}
	})
}

// TestPostgresProductRepositorySearch é o motivo principal de ter testes de
// integração no nível de Repository: prova que buildProductFilters monta o
// WHERE certo para cada filtro (e combinações deles) contra dados reais —
// algo que não faz sentido "vestir" com Service/Handler/HTTP por cima.
func TestPostgresProductRepositorySearch(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresProductRepository(db)

	const categoryPerifericos = "11111111-1111-7111-8111-111111111111"
	const categoryMonitores = "22222222-2222-7222-8222-222222222222"

	perifericos := categoryPerifericos
	monitores := categoryMonitores

	teclado := mustCreateProduct(t, repo, "Teclado Mecânico RGB", "switches azuis", 199.90, 10, &perifericos)
	mouse := mustCreateProduct(t, repo, "Mouse Gamer RGB", "sensor óptico", 89.90, 20, &perifericos)
	monitor := mustCreateProduct(t, repo, "Monitor 4K", "painel IPS", 1999.90, 5, &monitores)
	antigo := mustCreateProduct(t, repo, "Teclado Mecânico Antigo", "descontinuado", 99.90, 0, &perifericos)

	if err := repo.DeactivateByID(context.Background(), monitor.ID()); err != nil {
		t.Fatalf("desativar monitor: %v", err)
	}
	if err := repo.DeleteByID(context.Background(), antigo.ID()); err != nil {
		t.Fatalf("remover teclado antigo: %v", err)
	}

	testCases := []struct {
		name    string
		filter  ProductSearchFilter
		wantIDs []string
	}{
		{
			name:    "sem filtro extra usa o padrão (só não removidos)",
			filter:  ProductSearchFilter{Limit: 10},
			wantIDs: []string{teclado.ID(), mouse.ID(), monitor.ID()},
		},
		{
			name:    "filtra por nome (ILIKE parcial e case-insensitive)",
			filter:  ProductSearchFilter{Name: "mecânico", Limit: 10},
			wantIDs: []string{teclado.ID()}, // "Teclado Mecânico Antigo" também bate no nome, mas está removido
		},
		{
			name:    "filtra por categoria",
			filter:  ProductSearchFilter{CategoryID: &monitores, Limit: 10},
			wantIDs: []string{monitor.ID()},
		},
		{
			name:    "filtra por ativo=true",
			filter:  ProductSearchFilter{Active: boolPtr(true), Limit: 10},
			wantIDs: []string{teclado.ID(), mouse.ID()},
		},
		{
			name:    "filtra por ativo=false",
			filter:  ProductSearchFilter{Active: boolPtr(false), Limit: 10},
			wantIDs: []string{monitor.ID()},
		},
		{
			name:    "filtra por faixa de preço",
			filter:  ProductSearchFilter{MinPrice: floatPtr(100), MaxPrice: floatPtr(500), Limit: 10},
			wantIDs: []string{teclado.ID()},
		},
		{
			name:    "deletionFilter=deleted retorna só os removidos",
			filter:  ProductSearchFilter{DeletionFilter: DeletionFilterDeleted, Limit: 10},
			wantIDs: []string{antigo.ID()},
		},
		{
			name:    "deletionFilter=all ignora o filtro de remoção",
			filter:  ProductSearchFilter{DeletionFilter: DeletionFilterAll, Limit: 10},
			wantIDs: []string{teclado.ID(), mouse.ID(), monitor.ID(), antigo.ID()},
		},
		{
			name:    "combina categoria e faixa de preço",
			filter:  ProductSearchFilter{CategoryID: &perifericos, MinPrice: floatPtr(50), MaxPrice: floatPtr(150), Limit: 10},
			wantIDs: []string{mouse.ID()},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := repo.Search(context.Background(), testCase.filter)
			if err != nil {
				t.Fatalf("Search retornou erro: %v", err)
			}

			gotIDs := make([]string, 0, len(result.Products))
			for _, product := range result.Products {
				gotIDs = append(gotIDs, product.ID())
			}

			if !sameSet(gotIDs, testCase.wantIDs) {
				t.Errorf("produtos = %v; esperado (sem ordem) = %v", gotIDs, testCase.wantIDs)
			}
			if int(result.Total) != len(testCase.wantIDs) {
				t.Errorf("total = %d; esperado = %d", result.Total, len(testCase.wantIDs))
			}
		})
	}
}

// TestPostgresProductRepositorySearchPagination verifica Limit/Offset
// isoladamente do filtro: páginas não se sobrepõem e cobrem exatamente o
// total de registros.
func TestPostgresProductRepositorySearchPagination(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresProductRepository(db)

	const totalProducts = 5
	for i := range totalProducts {
		mustCreateProduct(t, repo, fmt.Sprintf("Produto Paginação %02d", i), "desc", 10, 1, nil)
	}

	seen := make(map[string]bool, totalProducts)
	const pageSize = 2

	for offset := 0; offset < totalProducts; offset += pageSize {
		page, err := repo.Search(context.Background(), ProductSearchFilter{Limit: pageSize, Offset: offset})
		if err != nil {
			t.Fatalf("Search(offset=%d) retornou erro: %v", offset, err)
		}
		if int(page.Total) != totalProducts {
			t.Fatalf("total na página offset=%d = %d; esperado = %d", offset, page.Total, totalProducts)
		}

		for _, product := range page.Products {
			if seen[product.ID()] {
				t.Errorf("produto %s apareceu em mais de uma página", product.ID())
			}
			seen[product.ID()] = true
		}
	}

	if len(seen) != totalProducts {
		t.Errorf("produtos únicos vistos = %d; esperado = %d", len(seen), totalProducts)
	}
}

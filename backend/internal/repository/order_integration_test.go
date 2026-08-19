//go:build integration

package repository

import (
	"errors"
	"testing"

	"ecommerce/internal/domain"
)

// mustCreateOrderOwner cria um cliente comum para ser dono de pedidos nos
// testes de OrderRepository.
func mustCreateOrderOwner(t *testing.T, userRepo *PostgresUserRepository, name, email string) *domain.User {
	t.Helper()
	return mustCreateUser(t, userRepo, name, email, domain.RoleCustomer)
}

// TestPostgresOrderRepositoryCreate foca no que os testes de integração
// HTTP (backend/integration) não alcançam: a checagem de proprietário
// (validateOwner) nunca roda de verdade por ali, porque o middleware
// Authenticate já rejeita (401) um usuário inativo/removido antes da
// requisição chegar ao Service — só uma chamada direta ao Repository
// consegue exercitar a janela de corrida que o comentário de validateOwner
// descreve. Também cobre um pedido com mais de um produto num único
// Create, caso que os testes HTTP não chegaram a montar.
func TestPostgresOrderRepositoryCreate(t *testing.T) {
	db := openTestDatabase(t)
	userRepo := NewPostgresUserRepository(db)
	productRepo := NewPostgresProductRepository(db)
	orderRepo := NewPostgresOrderRepository(db)

	owner := mustCreateOrderOwner(t, userRepo, "Dono do Pedido", "dono.pedido.repo@example.com")
	productA := mustCreateProduct(t, productRepo, "Produto A Pedido", "desc", 50, 10, nil)
	productB := mustCreateProduct(t, productRepo, "Produto B Pedido", "desc", 30, 5, nil)

	t.Run("proprietário inexistente retorna ErrOrderOwnerNotFound", func(t *testing.T) {
		_, err := orderRepo.Create(t.Context(), "00000000-0000-7000-8000-000000000000", []domain.CreateOrderItem{
			{ProductID: productA.ID(), Quantity: 1},
		})
		if !errors.Is(err, domain.ErrOrderOwnerNotFound) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderOwnerNotFound)
		}
	})

	t.Run("proprietário inativo retorna ErrOrderOwnerUnavailable", func(t *testing.T) {
		inactiveOwner := mustCreateOrderOwner(t, userRepo, "Dono Inativo", "dono.inativo.repo@example.com")
		if err := userRepo.DeactivateByID(t.Context(), inactiveOwner.ID()); err != nil {
			t.Fatalf("desativar proprietário: %v", err)
		}

		_, err := orderRepo.Create(t.Context(), inactiveOwner.ID(), []domain.CreateOrderItem{
			{ProductID: productA.ID(), Quantity: 1},
		})
		if !errors.Is(err, domain.ErrOrderOwnerUnavailable) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderOwnerUnavailable)
		}
	})

	t.Run("proprietário removido retorna ErrOrderOwnerUnavailable", func(t *testing.T) {
		deletedOwner := mustCreateOrderOwner(t, userRepo, "Dono Removido", "dono.removido.repo@example.com")
		if err := userRepo.DeleteByID(t.Context(), deletedOwner.ID()); err != nil {
			t.Fatalf("remover proprietário: %v", err)
		}

		_, err := orderRepo.Create(t.Context(), deletedOwner.ID(), []domain.CreateOrderItem{
			{ProductID: productA.ID(), Quantity: 1},
		})
		if !errors.Is(err, domain.ErrOrderOwnerUnavailable) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderOwnerUnavailable)
		}
	})

	t.Run("cria pedido com múltiplos produtos e decrementa o estoque de cada um", func(t *testing.T) {
		order, err := orderRepo.Create(t.Context(), owner.ID(), []domain.CreateOrderItem{
			{ProductID: productA.ID(), Quantity: 2},
			{ProductID: productB.ID(), Quantity: 3},
		})
		if err != nil {
			t.Fatalf("Create retornou erro: %v", err)
		}
		if order.CustomerID != owner.ID() || order.Status != domain.OrderStatusPending {
			t.Errorf("pedido criado = %#v", order)
		}
		wantTotal := 2*productA.Price() + 3*productB.Price()
		if order.TotalAmount != wantTotal {
			t.Errorf("total = %v; esperado = %v", order.TotalAmount, wantTotal)
		}
		if len(order.Items) != 2 {
			t.Fatalf("itens = %#v; esperados 2", order.Items)
		}

		gotA, err := productRepo.FindByID(t.Context(), productA.ID())
		if err != nil {
			t.Fatalf("FindByID produto A: %v", err)
		}
		if gotA.Stock() != productA.Stock()-2 {
			t.Errorf("estoque do produto A = %d; esperado = %d", gotA.Stock(), productA.Stock()-2)
		}

		gotB, err := productRepo.FindByID(t.Context(), productB.ID())
		if err != nil {
			t.Fatalf("FindByID produto B: %v", err)
		}
		if gotB.Stock() != productB.Stock()-3 {
			t.Errorf("estoque do produto B = %d; esperado = %d", gotB.Stock(), productB.Stock()-3)
		}
	})

	t.Run("produto inexistente retorna ErrProductNotFound", func(t *testing.T) {
		_, err := orderRepo.Create(t.Context(), owner.ID(), []domain.CreateOrderItem{
			{ProductID: "00000000-0000-7000-8000-000000000000", Quantity: 1},
		})
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrProductNotFound)
		}
	})

	t.Run("estoque insuficiente não altera o estoque (rollback)", func(t *testing.T) {
		productC := mustCreateProduct(t, productRepo, "Produto C Pedido", "desc", 40, 2, nil)

		_, err := orderRepo.Create(t.Context(), owner.ID(), []domain.CreateOrderItem{
			{ProductID: productC.ID(), Quantity: 5},
		})
		if !errors.Is(err, domain.ErrInsufficientStock) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrInsufficientStock)
		}

		got, err := productRepo.FindByID(t.Context(), productC.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.Stock() != productC.Stock() {
			t.Errorf("estoque = %d; esperado permanecer = %d (transação revertida)", got.Stock(), productC.Stock())
		}
	})
}

// TestPostgresOrderRepositoryFindByIDAndPayByID verifica a leitura com
// itens e a exigência atômica de PayByID (id + ownerID + status PENDING na
// mesma instrução SQL).
func TestPostgresOrderRepositoryFindByIDAndPayByID(t *testing.T) {
	db := openTestDatabase(t)
	userRepo := NewPostgresUserRepository(db)
	productRepo := NewPostgresProductRepository(db)
	orderRepo := NewPostgresOrderRepository(db)

	owner := mustCreateOrderOwner(t, userRepo, "Dono Pagamento Repo", "dono.pagamento.repo@example.com")
	other := mustCreateOrderOwner(t, userRepo, "Outro Pagamento Repo", "outro.pagamento.repo@example.com")
	product := mustCreateProduct(t, productRepo, "Produto Pagamento Repo", "desc", 60, 10, nil)

	order, err := orderRepo.Create(t.Context(), owner.ID(), []domain.CreateOrderItem{{ProductID: product.ID(), Quantity: 2}})
	if err != nil {
		t.Fatalf("preparar pedido: %v", err)
	}

	t.Run("FindByID traz os itens", func(t *testing.T) {
		got, err := orderRepo.FindByID(t.Context(), order.ID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if len(got.Items) != 1 || got.Items[0].ProductID != product.ID() || got.Items[0].Quantity != 2 {
			t.Errorf("pedido lido = %#v", got)
		}
	})

	t.Run("FindByID inexistente retorna ErrOrderNotFound", func(t *testing.T) {
		_, err := orderRepo.FindByID(t.Context(), "00000000-0000-7000-8000-000000000000")
		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderNotFound)
		}
	})

	t.Run("pagar com outro proprietário retorna ErrOrderAccessDenied", func(t *testing.T) {
		_, err := orderRepo.PayByID(t.Context(), order.ID, other.ID())
		if !errors.Is(err, domain.ErrOrderAccessDenied) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderAccessDenied)
		}
	})

	t.Run("pagar id inexistente retorna ErrOrderNotFound", func(t *testing.T) {
		_, err := orderRepo.PayByID(t.Context(), "00000000-0000-7000-8000-000000000000", owner.ID())
		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderNotFound)
		}
	})

	t.Run("dono paga o pedido", func(t *testing.T) {
		paid, err := orderRepo.PayByID(t.Context(), order.ID, owner.ID())
		if err != nil {
			t.Fatalf("PayByID: %v", err)
		}
		if paid.Status != domain.OrderStatusPaid || paid.PaidAt == nil {
			t.Errorf("pedido pago = %#v", paid)
		}
		if len(paid.Items) != 1 {
			t.Errorf("itens do pedido pago = %#v", paid.Items)
		}
	})

	t.Run("pagar de novo (mesmo dono) retorna ErrOrderCannotBePaid", func(t *testing.T) {
		_, err := orderRepo.PayByID(t.Context(), order.ID, owner.ID())
		if !errors.Is(err, domain.ErrOrderCannotBePaid) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderCannotBePaid)
		}
	})
}

// TestPostgresOrderRepositoryCancelByID verifica a assimetria com PayByID
// (CancelByID recebe isAdmin explicitamente) e a restauração exata do
// estoque ao cancelar.
func TestPostgresOrderRepositoryCancelByID(t *testing.T) {
	db := openTestDatabase(t)
	userRepo := NewPostgresUserRepository(db)
	productRepo := NewPostgresProductRepository(db)
	orderRepo := NewPostgresOrderRepository(db)

	owner := mustCreateOrderOwner(t, userRepo, "Dono Cancelamento Repo", "dono.cancelamento.repo@example.com")
	other := mustCreateOrderOwner(t, userRepo, "Outro Cancelamento Repo", "outro.cancelamento.repo@example.com")
	product := mustCreateProduct(t, productRepo, "Produto Cancelamento Repo", "desc", 40, 10, nil)

	t.Run("outro usuário sem ser admin não pode cancelar", func(t *testing.T) {
		order, err := orderRepo.Create(t.Context(), owner.ID(), []domain.CreateOrderItem{{ProductID: product.ID(), Quantity: 2}})
		if err != nil {
			t.Fatalf("preparar pedido: %v", err)
		}

		_, err = orderRepo.CancelByID(t.Context(), order.ID, other.ID(), false)
		if !errors.Is(err, domain.ErrOrderAccessDenied) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderAccessDenied)
		}

		// Cancela de verdade (como dono) para não deixar estoque preso.
		if _, err := orderRepo.CancelByID(t.Context(), order.ID, owner.ID(), false); err != nil {
			t.Fatalf("cleanup: cancelar pedido: %v", err)
		}
	})

	t.Run("cancelar restaura o estoque exato do item", func(t *testing.T) {
		before, err := productRepo.FindByID(t.Context(), product.ID())
		if err != nil {
			t.Fatalf("FindByID antes: %v", err)
		}

		order, err := orderRepo.Create(t.Context(), owner.ID(), []domain.CreateOrderItem{{ProductID: product.ID(), Quantity: 4}})
		if err != nil {
			t.Fatalf("preparar pedido: %v", err)
		}

		canceled, err := orderRepo.CancelByID(t.Context(), order.ID, owner.ID(), false)
		if err != nil {
			t.Fatalf("CancelByID: %v", err)
		}
		if canceled.Status != domain.OrderStatusCanceled || canceled.CanceledAt == nil {
			t.Errorf("pedido cancelado = %#v", canceled)
		}

		after, err := productRepo.FindByID(t.Context(), product.ID())
		if err != nil {
			t.Fatalf("FindByID depois: %v", err)
		}
		if after.Stock() != before.Stock() {
			t.Errorf("estoque = %d; esperado voltar a = %d", after.Stock(), before.Stock())
		}
	})

	t.Run("admin pode cancelar pedido de outro usuário", func(t *testing.T) {
		order, err := orderRepo.Create(t.Context(), owner.ID(), []domain.CreateOrderItem{{ProductID: product.ID(), Quantity: 1}})
		if err != nil {
			t.Fatalf("preparar pedido: %v", err)
		}

		canceled, err := orderRepo.CancelByID(t.Context(), order.ID, other.ID(), true)
		if err != nil {
			t.Fatalf("admin deveria poder cancelar: %v", err)
		}
		if canceled.Status != domain.OrderStatusCanceled {
			t.Errorf("pedido cancelado = %#v", canceled)
		}
	})

	t.Run("cancelar id inexistente retorna ErrOrderNotFound", func(t *testing.T) {
		_, err := orderRepo.CancelByID(t.Context(), "00000000-0000-7000-8000-000000000000", owner.ID(), false)
		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderNotFound)
		}
	})

	t.Run("cancelar pedido já pago retorna ErrOrderCannotBeCanceled", func(t *testing.T) {
		order, err := orderRepo.Create(t.Context(), owner.ID(), []domain.CreateOrderItem{{ProductID: product.ID(), Quantity: 1}})
		if err != nil {
			t.Fatalf("preparar pedido: %v", err)
		}
		if _, err := orderRepo.PayByID(t.Context(), order.ID, owner.ID()); err != nil {
			t.Fatalf("pagar pedido: %v", err)
		}

		_, err = orderRepo.CancelByID(t.Context(), order.ID, owner.ID(), false)
		if !errors.Is(err, domain.ErrOrderCannotBeCanceled) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrOrderCannotBeCanceled)
		}
	})
}

// TestPostgresOrderRepositorySearch verifica que buildOrderFilters aplica
// (ou não) o filtro por CustomerID corretamente contra dados reais.
func TestPostgresOrderRepositorySearch(t *testing.T) {
	db := openTestDatabase(t)
	userRepo := NewPostgresUserRepository(db)
	productRepo := NewPostgresProductRepository(db)
	orderRepo := NewPostgresOrderRepository(db)

	customerA := mustCreateOrderOwner(t, userRepo, "Cliente A Repo", "cliente.a.repo@example.com")
	customerB := mustCreateOrderOwner(t, userRepo, "Cliente B Repo", "cliente.b.repo@example.com")
	product := mustCreateProduct(t, productRepo, "Produto Busca Repo", "desc", 25, 50, nil)

	for range 2 {
		if _, err := orderRepo.Create(t.Context(), customerA.ID(), []domain.CreateOrderItem{{ProductID: product.ID(), Quantity: 1}}); err != nil {
			t.Fatalf("criar pedido do cliente A: %v", err)
		}
	}
	if _, err := orderRepo.Create(t.Context(), customerB.ID(), []domain.CreateOrderItem{{ProductID: product.ID(), Quantity: 1}}); err != nil {
		t.Fatalf("criar pedido do cliente B: %v", err)
	}

	t.Run("sem filtro retorna todos os pedidos", func(t *testing.T) {
		result, err := orderRepo.Search(t.Context(), OrderFilter{Limit: 10})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if result.Total != 3 {
			t.Errorf("total = %d; esperado = 3", result.Total)
		}
	})

	t.Run("filtro por CustomerID retorna só os pedidos do cliente", func(t *testing.T) {
		customerAID := customerA.ID()

		result, err := orderRepo.Search(t.Context(), OrderFilter{CustomerID: &customerAID, Limit: 10})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if result.Total != 2 || len(result.Orders) != 2 {
			t.Fatalf("resultado = %#v", result)
		}
		for _, order := range result.Orders {
			if order.CustomerID != customerA.ID() {
				t.Errorf("pedido de outro cliente vazou no filtro: %#v", order)
			}
		}
	})
}

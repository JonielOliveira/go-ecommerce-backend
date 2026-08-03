//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"ecommerce/internal/dto"
)

// TestOrderCreateFlow verifica a criação de pedido contra o
// PostgresOrderRepository real: produto inexistente, estoque insuficiente
// (sem alterar o estoque — a transação é revertida antes do commit) e
// produto inativo são recusados; a criação válida decrementa o estoque
// atomicamente e calcula o total corretamente. Nenhum desses caminhos é
// verificável com Repository falso, porque dependem do lock (FOR UPDATE) e
// da transação reais.
func TestOrderCreateFlow(t *testing.T) {
	app := newTestApp(t)

	adminClient := createAndLoginAdmin(t, app, "Admin Pedidos", "admin.pedidos@example.com", "SenhaForte@123")
	customerClient := registerAndLoginCustomer(t, app, "Cliente Pedidos", "cliente.pedidos@example.com", "SenhaForte@123")

	product := createTestProduct(t, adminClient, app.Server.URL, "Produto de Pedido", 50, 10)

	t.Run("visitante anônimo não pode criar pedido", func(t *testing.T) {
		anonymousClient := app.newClient(t)

		result := performRequest(t, anonymousClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 1}},
		})

		if result.status != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusUnauthorized)
		}
	})

	t.Run("produto inexistente retorna 404", func(t *testing.T) {
		result := performRequest(t, customerClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: "00000000-0000-7000-8000-000000000000", Quantity: 1}},
		})

		if result.status != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusNotFound, result.body)
		}
	})

	t.Run("estoque insuficiente retorna 409 sem alterar o estoque", func(t *testing.T) {
		result := performRequest(t, customerClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 999}},
		})
		if result.status != http.StatusConflict {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusConflict, result.body)
		}

		getResult := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/products/"+product.ID, nil)
		got := decodeInto[dto.ProductResponse](t, getResult)
		if got.Stock != product.Stock {
			t.Errorf("estoque = %d; esperado permanecer = %d (transação revertida)", got.Stock, product.Stock)
		}
	})

	t.Run("produto inativo retorna 409", func(t *testing.T) {
		inactive := createTestProduct(t, adminClient, app.Server.URL, "Produto Inativo", 20, 5)
		deactivateResult := performRequest(t, adminClient, http.MethodPatch, app.Server.URL+"/api/v1/products/"+inactive.ID+"/deactivate", nil)
		if deactivateResult.status != http.StatusNoContent {
			t.Fatalf("desativar produto: status = %d; body = %s", deactivateResult.status, deactivateResult.body)
		}

		result := performRequest(t, customerClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: inactive.ID, Quantity: 1}},
		})

		if result.status != http.StatusConflict {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusConflict, result.body)
		}
	})

	var created dto.OrderResponse

	t.Run("cliente cria pedido e o estoque é decrementado atomicamente", func(t *testing.T) {
		result := performRequest(t, customerClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 3}},
		})

		if result.status != http.StatusCreated {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusCreated, result.body)
		}
		created = decodeInto[dto.OrderResponse](t, result)
		if created.Status != "PENDING" || len(created.Items) != 1 || created.Items[0].Quantity != 3 {
			t.Fatalf("pedido criado = %#v", created)
		}
		wantTotal := 3 * product.Price
		if created.TotalAmount != wantTotal {
			t.Errorf("total_amount = %v; esperado = %v", created.TotalAmount, wantTotal)
		}

		getResult := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/products/"+product.ID, nil)
		got := decodeInto[dto.ProductResponse](t, getResult)
		if got.Stock != product.Stock-3 {
			t.Errorf("estoque após pedido = %d; esperado = %d", got.Stock, product.Stock-3)
		}
	})

	t.Run("dono lê o próprio pedido, outro cliente recebe 404", func(t *testing.T) {
		ownResult := performRequest(t, customerClient, http.MethodGet, app.Server.URL+"/api/v1/orders/"+created.ID, nil)
		if ownResult.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", ownResult.status, http.StatusOK, ownResult.body)
		}

		otherClient := registerAndLoginCustomer(t, app, "Outro Cliente", "outro.cliente.pedidos@example.com", "SenhaForte@123")
		otherResult := performRequest(t, otherClient, http.MethodGet, app.Server.URL+"/api/v1/orders/"+created.ID, nil)
		if otherResult.status != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d (acesso negado é reportado como não encontrado, para não vazar a existência do pedido)", otherResult.status, http.StatusNotFound)
		}
	})
}

// TestOrderPayOwnership verifica a regra mais sutil de todo o módulo de
// pedidos: PayByID exige atomicamente id + ownerID + status PENDING na
// mesma instrução SQL, sem exceção para admin — só o dono paga o próprio
// pedido. Isso só é verificável de verdade contra o banco, porque um
// Repository falso nunca teria essa checagem incorporada na query.
func TestOrderPayOwnership(t *testing.T) {
	app := newTestApp(t)

	adminClient := createAndLoginAdmin(t, app, "Admin Pagamento", "admin.pagamento@example.com", "SenhaForte@123")
	ownerClient := registerAndLoginCustomer(t, app, "Dono do Pedido", "dono.pagamento@example.com", "SenhaForte@123")
	otherClient := registerAndLoginCustomer(t, app, "Outro Cliente", "outro.pagamento@example.com", "SenhaForte@123")

	product := createTestProduct(t, adminClient, app.Server.URL, "Produto Pagamento", 100, 5)

	createResult := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
		Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 1}},
	})
	if createResult.status != http.StatusCreated {
		t.Fatalf("criar pedido: status = %d; body = %s", createResult.status, createResult.body)
	}
	order := decodeInto[dto.OrderResponse](t, createResult)

	t.Run("outro cliente não pode pagar o pedido", func(t *testing.T) {
		result := performRequest(t, otherClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/pay", nil)
		if result.status != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusNotFound)
		}
	})

	t.Run("admin não pode pagar pedido de outro usuário — sem exceção", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/pay", nil)
		if result.status != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d (PayByID não abre exceção para admin)", result.status, http.StatusNotFound)
		}
	})

	t.Run("dono paga o próprio pedido", func(t *testing.T) {
		result := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/pay", nil)
		if result.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusOK, result.body)
		}
		paid := decodeInto[dto.OrderResponse](t, result)
		if paid.Status != "PAID" || paid.PaidAt == nil {
			t.Errorf("pedido pago = %#v", paid)
		}
	})

	t.Run("pagar de novo um pedido já pago retorna 409", func(t *testing.T) {
		result := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/pay", nil)
		if result.status != http.StatusConflict {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusConflict)
		}
	})

	t.Run("cancelar um pedido já pago retorna 409", func(t *testing.T) {
		result := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/cancel", nil)
		if result.status != http.StatusConflict {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusConflict)
		}
	})
}

// TestOrderCancelRestoresStockAndAllowsAdmin verifica a assimetria entre
// pagar e cancelar: só o dono paga, mas admin PODE cancelar o pedido de
// qualquer cliente (CancelByID recebe isAdmin explicitamente, diferente de
// PayByID). Verifica também que cancelar devolve a quantidade ao estoque.
func TestOrderCancelRestoresStockAndAllowsAdmin(t *testing.T) {
	app := newTestApp(t)

	adminClient := createAndLoginAdmin(t, app, "Admin Cancelamento", "admin.cancelamento@example.com", "SenhaForte@123")
	ownerClient := registerAndLoginCustomer(t, app, "Dono do Pedido", "dono.cancelamento@example.com", "SenhaForte@123")
	otherClient := registerAndLoginCustomer(t, app, "Outro Cliente", "outro.cancelamento@example.com", "SenhaForte@123")

	product := createTestProduct(t, adminClient, app.Server.URL, "Produto Cancelamento", 80, 10)

	t.Run("outro cliente não pode cancelar o pedido", func(t *testing.T) {
		createResult := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 2}},
		})
		order := decodeInto[dto.OrderResponse](t, createResult)

		result := performRequest(t, otherClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/cancel", nil)
		if result.status != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusNotFound)
		}

		// Cancela de verdade (como dono) para não deixar estoque preso e
		// atrapalhar os próximos subtests.
		cleanup := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/cancel", nil)
		if cleanup.status != http.StatusOK {
			t.Fatalf("cleanup: cancelar pedido = %d; body = %s", cleanup.status, cleanup.body)
		}
	})

	t.Run("dono cancela o próprio pedido e o estoque é restaurado", func(t *testing.T) {
		beforeResult := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/products/"+product.ID, nil)
		before := decodeInto[dto.ProductResponse](t, beforeResult)

		createResult := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 3}},
		})
		order := decodeInto[dto.OrderResponse](t, createResult)

		duringResult := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/products/"+product.ID, nil)
		during := decodeInto[dto.ProductResponse](t, duringResult)
		if during.Stock != before.Stock-3 {
			t.Fatalf("estoque após criar pedido = %d; esperado = %d", during.Stock, before.Stock-3)
		}

		cancelResult := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/cancel", nil)
		if cancelResult.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", cancelResult.status, http.StatusOK, cancelResult.body)
		}
		canceled := decodeInto[dto.OrderResponse](t, cancelResult)
		if canceled.Status != "CANCELED" || canceled.CanceledAt == nil {
			t.Errorf("pedido cancelado = %#v", canceled)
		}

		afterResult := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/products/"+product.ID, nil)
		after := decodeInto[dto.ProductResponse](t, afterResult)
		if after.Stock != before.Stock {
			t.Errorf("estoque após cancelar = %d; esperado voltar a = %d", after.Stock, before.Stock)
		}
	})

	t.Run("admin pode cancelar pedido de outro cliente", func(t *testing.T) {
		createResult := performRequest(t, ownerClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 1}},
		})
		order := decodeInto[dto.OrderResponse](t, createResult)

		result := performRequest(t, adminClient, http.MethodPost, app.Server.URL+"/api/v1/orders/"+order.ID+"/cancel", nil)
		if result.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s (admin deve poder cancelar qualquer pedido)", result.status, http.StatusOK, result.body)
		}
		canceled := decodeInto[dto.OrderResponse](t, result)
		if canceled.Status != "CANCELED" {
			t.Errorf("pedido = %#v", canceled)
		}
	})
}

// TestOrderSearchOwnership verifica a regra de propriedade aplicada pelo
// OrderFilter real: customer só vê e só conta os próprios pedidos; admin vê
// e conta todos.
func TestOrderSearchOwnership(t *testing.T) {
	app := newTestApp(t)

	adminClient := createAndLoginAdmin(t, app, "Admin Busca", "admin.busca@example.com", "SenhaForte@123")
	customerAClient := registerAndLoginCustomer(t, app, "Cliente A", "cliente.a.busca@example.com", "SenhaForte@123")
	customerBClient := registerAndLoginCustomer(t, app, "Cliente B", "cliente.b.busca@example.com", "SenhaForte@123")

	product := createTestProduct(t, adminClient, app.Server.URL, "Produto Busca", 30, 20)

	for range 2 {
		result := performRequest(t, customerAClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
			Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 1}},
		})
		if result.status != http.StatusCreated {
			t.Fatalf("criar pedido do cliente A: status = %d; body = %s", result.status, result.body)
		}
	}

	bResult := performRequest(t, customerBClient, http.MethodPost, app.Server.URL+"/api/v1/orders", dto.CreateOrderRequest{
		Items: []dto.CreateOrderItemRequest{{ProductID: product.ID, Quantity: 1}},
	})
	if bResult.status != http.StatusCreated {
		t.Fatalf("criar pedido do cliente B: status = %d; body = %s", bResult.status, bResult.body)
	}

	t.Run("cliente A só vê e só conta os próprios pedidos", func(t *testing.T) {
		searchResult := performRequest(t, customerAClient, http.MethodGet, app.Server.URL+"/api/v1/orders", nil)
		if searchResult.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", searchResult.status, http.StatusOK, searchResult.body)
		}
		page := decodeInto[dto.OrderPageResponse](t, searchResult)
		if page.TotalItems != 2 || len(page.Items) != 2 {
			t.Fatalf("página do cliente A = %#v", page)
		}
	})

	t.Run("admin vê e conta todos os pedidos", func(t *testing.T) {
		searchResult := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/orders", nil)
		if searchResult.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", searchResult.status, http.StatusOK, searchResult.body)
		}
		page := decodeInto[dto.OrderPageResponse](t, searchResult)
		if page.TotalItems != 3 {
			t.Fatalf("total de pedidos visto pelo admin = %d; esperado = 3", page.TotalItems)
		}
	})
}

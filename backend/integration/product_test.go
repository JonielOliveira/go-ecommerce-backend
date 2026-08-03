//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"ecommerce/internal/dto"
)

// TestProductAdminFlow verifica a gestão de produtos de ponta a ponta:
// leitura é pública, escrita exige o papel "admin" de verdade (via
// middleware.Authenticate + middleware.RequireRole reais, não fakes), e o
// soft delete não esconde o produto de uma busca direta por ID.
func TestProductAdminFlow(t *testing.T) {
	app := newTestApp(t)

	adminClient := createAndLoginAdmin(t, app, "Admin Integração", "admin.integracao@example.com", "SenhaForte@123")

	validProduct := dto.ProductRequest{
		Name:        "Teclado Mecânico",
		Description: "RGB, switches azuis",
		Price:       299.90,
		Stock:       15,
	}

	t.Run("visitante anônimo não pode criar produto", func(t *testing.T) {
		anonymousClient := app.newClient(t)

		result := performRequest(t, anonymousClient, http.MethodPost, app.Server.URL+"/api/v1/products", validProduct)

		if result.status != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusUnauthorized)
		}
	})

	t.Run("cliente autenticado não pode criar produto", func(t *testing.T) {
		customerClient := registerAndLoginCustomer(t, app, "Cliente Integração", "cliente.produtos@example.com", "SenhaForte@123")

		result := performRequest(t, customerClient, http.MethodPost, app.Server.URL+"/api/v1/products", validProduct)

		if result.status != http.StatusForbidden {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusForbidden)
		}
	})

	var created dto.ProductResponse

	t.Run("admin cria produto", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPost, app.Server.URL+"/api/v1/products", validProduct)

		if result.status != http.StatusCreated {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusCreated, result.body)
		}
		created = decodeInto[dto.ProductResponse](t, result)
		if created.ID == "" || created.Name != validProduct.Name || !created.Active {
			t.Errorf("produto criado = %#v", created)
		}
	})

	t.Run("qualquer visitante lê o produto criado", func(t *testing.T) {
		anonymousClient := app.newClient(t)

		result := performRequest(t, anonymousClient, http.MethodGet, app.Server.URL+"/api/v1/products/"+created.ID, nil)

		if result.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusOK, result.body)
		}
		got := decodeInto[dto.ProductResponse](t, result)
		if got.ID != created.ID {
			t.Errorf("produto lido = %#v; esperado id = %q", got, created.ID)
		}
	})

	t.Run("admin atualiza o produto", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPut, app.Server.URL+"/api/v1/products/"+created.ID, dto.ProductUpdateRequest{
			Name:        "Teclado Mecânico V2",
			Description: "RGB, switches vermelhos",
			Price:       349.90,
			Stock:       8,
		})

		if result.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusOK, result.body)
		}
		updated := decodeInto[dto.ProductResponse](t, result)
		if updated.Name != "Teclado Mecânico V2" || updated.Stock != 8 {
			t.Errorf("produto atualizado = %#v", updated)
		}
	})

	t.Run("admin remove o produto e a leitura por ID continua funcionando", func(t *testing.T) {
		deleteResult := performRequest(t, adminClient, http.MethodDelete, app.Server.URL+"/api/v1/products/"+created.ID, nil)
		if deleteResult.status != http.StatusNoContent {
			t.Fatalf("status do delete = %d; esperado = %d; body = %s", deleteResult.status, http.StatusNoContent, deleteResult.body)
		}

		secondDelete := performRequest(t, adminClient, http.MethodDelete, app.Server.URL+"/api/v1/products/"+created.ID, nil)
		if secondDelete.status != http.StatusConflict {
			t.Fatalf("status do segundo delete = %d; esperado = %d (já removido)", secondDelete.status, http.StatusConflict)
		}

		anonymousClient := app.newClient(t)
		getResult := performRequest(t, anonymousClient, http.MethodGet, app.Server.URL+"/api/v1/products/"+created.ID, nil)
		if getResult.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d (FindByID não filtra removidos)", getResult.status, http.StatusOK)
		}
		got := decodeInto[dto.ProductResponse](t, getResult)
		if got.DeletedAt == nil {
			t.Errorf("produto deveria estar marcado como removido; got = %#v", got)
		}
	})

	t.Run("produto inexistente retorna 404", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/products/00000000-0000-7000-8000-000000000000", nil)

		if result.status != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusNotFound)
		}
	})
}

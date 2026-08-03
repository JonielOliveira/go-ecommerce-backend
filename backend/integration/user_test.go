//go:build integration

package integration_test

import (
	"net/http"
	"net/url"
	"testing"

	"ecommerce/internal/dto"
)

// TestUserAdminFlow verifica a gestão administrativa de usuários de ponta a
// ponta. Diferente de Product, todo o grupo /users exige "admin" — não há
// rota pública aqui (RegisterUserRoutes aplica authenticate+requireAdmin a
// todo o grupo, sem exceção para leitura).
func TestUserAdminFlow(t *testing.T) {
	app := newTestApp(t)

	adminClient := createAndLoginAdmin(t, app, "Admin Usuários", "admin.usuarios@example.com", "SenhaForte@123")
	customerClient := registerAndLoginCustomer(t, app, "Cliente Comum", "cliente.usuarios@example.com", "SenhaForte@123")

	t.Run("visitante anônimo não pode listar usuários", func(t *testing.T) {
		anonymousClient := app.newClient(t)

		result := performRequest(t, anonymousClient, http.MethodGet, app.Server.URL+"/api/v1/users", nil)

		if result.status != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusUnauthorized)
		}
	})

	t.Run("cliente autenticado não pode criar usuário", func(t *testing.T) {
		result := performRequest(t, customerClient, http.MethodPost, app.Server.URL+"/api/v1/users", dto.CreateUserRequest{
			Name:     "Usuário Qualquer",
			Email:    "usuario.qualquer@example.com",
			Password: "SenhaForte@123",
		})

		if result.status != http.StatusForbidden {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusForbidden)
		}
	})

	var created dto.UserResponse

	t.Run("admin cria usuário", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPost, app.Server.URL+"/api/v1/users", dto.CreateUserRequest{
			Name:     "Usuário Gerenciado Integração",
			Email:    "usuario.gerenciado@example.com",
			Password: "SenhaForte@123",
		})

		if result.status != http.StatusCreated {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusCreated, result.body)
		}
		created = decodeInto[dto.UserResponse](t, result)
		if created.ID == "" || created.Role != "customer" || !created.Active {
			t.Errorf("usuário criado = %#v", created)
		}
	})

	t.Run("email duplicado retorna 409", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPost, app.Server.URL+"/api/v1/users", dto.CreateUserRequest{
			Name:     "Outro Nome",
			Email:    "usuario.gerenciado@example.com",
			Password: "SenhaForte@123",
		})

		if result.status != http.StatusConflict {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusConflict)
		}
	})

	t.Run("admin busca o usuário criado", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/users/"+created.ID, nil)

		if result.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusOK, result.body)
		}
		got := decodeInto[dto.UserResponse](t, result)
		if got.ID != created.ID {
			t.Errorf("usuário lido = %#v; esperado id = %q", got, created.ID)
		}
	})

	t.Run("usuário inexistente retorna 404", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/users/00000000-0000-7000-8000-000000000000", nil)

		if result.status != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusNotFound)
		}
	})

	t.Run("admin busca usuários filtrando por nome", func(t *testing.T) {
		query := url.Values{}
		query.Set("name", "Gerenciado Integração")

		result := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/users?"+query.Encode(), nil)

		if result.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusOK, result.body)
		}
		page := decodeInto[dto.UserPageResponse](t, result)
		if page.TotalItems != 1 || len(page.Items) != 1 || page.Items[0].ID != created.ID {
			t.Fatalf("página filtrada = %#v", page)
		}
	})

	t.Run("admin atualiza o usuário", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPut, app.Server.URL+"/api/v1/users/"+created.ID, dto.UserUpdateRequest{
			Name:  "Usuário Gerenciado Atualizado",
			Email: "usuario.gerenciado@example.com",
			Role:  "customer",
		})

		if result.status != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusOK, result.body)
		}
		updated := decodeInto[dto.UserResponse](t, result)
		if updated.Name != "Usuário Gerenciado Atualizado" {
			t.Errorf("usuário atualizado = %#v", updated)
		}
	})

	t.Run("admin remove o usuário; remover de novo dá 409", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodDelete, app.Server.URL+"/api/v1/users/"+created.ID, nil)
		if result.status != http.StatusNoContent {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusNoContent, result.body)
		}

		again := performRequest(t, adminClient, http.MethodDelete, app.Server.URL+"/api/v1/users/"+created.ID, nil)
		if again.status != http.StatusConflict {
			t.Fatalf("segundo delete: status = %d; esperado = %d", again.status, http.StatusConflict)
		}
	})

	t.Run("restaurar devolve o usuário, mas inativo — não reativa sozinho", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPatch, app.Server.URL+"/api/v1/users/"+created.ID+"/restore", nil)
		if result.status != http.StatusNoContent {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusNoContent, result.body)
		}

		again := performRequest(t, adminClient, http.MethodPatch, app.Server.URL+"/api/v1/users/"+created.ID+"/restore", nil)
		if again.status != http.StatusConflict {
			t.Fatalf("segunda restauração: status = %d; esperado = %d (não está mais removido)", again.status, http.StatusConflict)
		}

		getResult := performRequest(t, adminClient, http.MethodGet, app.Server.URL+"/api/v1/users/"+created.ID, nil)
		got := decodeInto[dto.UserResponse](t, getResult)
		if got.DeletedAt != nil {
			t.Errorf("usuário deveria estar restaurado (deletedAt nulo); got = %#v", got)
		}
		if got.Active {
			t.Errorf("restaurar não deveria reativar automaticamente; got = %#v", got)
		}
	})

	t.Run("ativa o usuário restaurado; ativar de novo dá 409", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPatch, app.Server.URL+"/api/v1/users/"+created.ID+"/activate", nil)
		if result.status != http.StatusNoContent {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusNoContent, result.body)
		}

		again := performRequest(t, adminClient, http.MethodPatch, app.Server.URL+"/api/v1/users/"+created.ID+"/activate", nil)
		if again.status != http.StatusConflict {
			t.Fatalf("segunda ativação: status = %d; esperado = %d (já ativo)", again.status, http.StatusConflict)
		}
	})

	t.Run("desativa o usuário; desativar de novo dá 409", func(t *testing.T) {
		result := performRequest(t, adminClient, http.MethodPatch, app.Server.URL+"/api/v1/users/"+created.ID+"/deactivate", nil)
		if result.status != http.StatusNoContent {
			t.Fatalf("status = %d; esperado = %d; body = %s", result.status, http.StatusNoContent, result.body)
		}

		again := performRequest(t, adminClient, http.MethodPatch, app.Server.URL+"/api/v1/users/"+created.ID+"/deactivate", nil)
		if again.status != http.StatusConflict {
			t.Fatalf("segunda desativação: status = %d; esperado = %d (já inativo)", again.status, http.StatusConflict)
		}
	})

	t.Run("ativar um usuário removido reporta remoção, não 'já ativo'", func(t *testing.T) {
		deleteResult := performRequest(t, adminClient, http.MethodDelete, app.Server.URL+"/api/v1/users/"+created.ID, nil)
		if deleteResult.status != http.StatusNoContent {
			t.Fatalf("remover antes do teste final: status = %d; body = %s", deleteResult.status, deleteResult.body)
		}

		result := performRequest(t, adminClient, http.MethodPatch, app.Server.URL+"/api/v1/users/"+created.ID+"/activate", nil)
		if result.status != http.StatusConflict {
			t.Fatalf("status = %d; esperado = %d", result.status, http.StatusConflict)
		}
		got := decodeInto[apiErrorResponse](t, result)
		if got.Error != "usuário já está removido" {
			t.Errorf("mensagem = %q; esperada = %q (a checagem de remoção vem antes da de já-ativo)", got.Error, "usuário já está removido")
		}
	})
}

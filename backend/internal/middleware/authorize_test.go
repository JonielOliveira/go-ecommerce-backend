package middleware

import (
	"net/http"
	"testing"

	"ecommerce/internal/domain"
)

// TestRequireRole verifica as três saídas possíveis: falta de autenticação
// (401, tratada como erro distinto de falta de permissão), papel fora da
// lista permitida (403), e papel permitido (segue adiante sem abortar).
func TestRequireRole(t *testing.T) {
	t.Run("sem usuário autenticado retorna 401", func(t *testing.T) {
		context, recorder := newTestContext(http.MethodGet, "/qualquer")

		RequireRole(domain.RoleAdmin)(context)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
		if !context.IsAborted() {
			t.Errorf("contexto deveria estar abortado")
		}
	})

	t.Run("papel fora da lista permitida retorna 403", func(t *testing.T) {
		context, recorder := newTestContext(http.MethodGet, "/qualquer")
		context.Set(AuthenticatedUserContextKey, &domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer})

		RequireRole(domain.RoleAdmin)(context)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusForbidden)
		}
		if !context.IsAborted() {
			t.Errorf("contexto deveria estar abortado")
		}
	})

	t.Run("papel permitido segue adiante sem abortar", func(t *testing.T) {
		context, recorder := newTestContext(http.MethodGet, "/qualquer")
		context.Set(AuthenticatedUserContextKey, &domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleAdmin})

		RequireRole(domain.RoleAdmin)(context)

		if context.IsAborted() {
			t.Errorf("contexto não deveria estar abortado")
		}
		if recorder.Code != http.StatusOK {
			t.Errorf("status = %d; esperado permanecer no default = %d (nada deveria ter sido escrito)", recorder.Code, http.StatusOK)
		}
	})

	t.Run("aceita qualquer um dos papéis permitidos", func(t *testing.T) {
		testCases := []domain.UserRole{domain.RoleCustomer, domain.RoleAdmin}

		for _, role := range testCases {
			t.Run(string(role), func(t *testing.T) {
				context, _ := newTestContext(http.MethodGet, "/qualquer")
				context.Set(AuthenticatedUserContextKey, &domain.AuthenticatedUser{ID: "user-1", Role: role})

				RequireRole(domain.RoleCustomer, domain.RoleAdmin)(context)

				if context.IsAborted() {
					t.Errorf("papel %q deveria ser aceito", role)
				}
			})
		}
	})
}

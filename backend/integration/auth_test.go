//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"ecommerce/internal/dto"
)

// TestAuthRegisterLoginMeLogout percorre o fluxo completo de autenticação
// contra PostgreSQL real: autocadastro (sempre "customer"), /me sem sessão,
// senha errada, login (cookie HttpOnly definido pelo servidor e devolvido
// automaticamente pelo cookie jar do client), /me autenticado e logout.
// Nenhuma camada é substituída — bcrypt, JWT e o middleware Authenticate
// rodam de verdade.
func TestAuthRegisterLoginMeLogout(t *testing.T) {
	app := newTestApp(t)
	client := app.newClient(t)

	const email = "cliente.integracao@example.com"
	const password = "SenhaForte@123"

	registerResult := performRequest(t, client, http.MethodPost, app.Server.URL+"/api/v1/auth/register", dto.RegisterRequest{
		Name:     "Cliente Integração",
		Email:    email,
		Password: password,
	})
	if registerResult.status != http.StatusCreated {
		t.Fatalf("status do registro = %d; esperado = %d; body = %s", registerResult.status, http.StatusCreated, registerResult.body)
	}
	registered := decodeInto[dto.UserResponse](t, registerResult)
	if registered.Role != "customer" {
		t.Errorf("papel do autocadastro = %q; esperado = %q (nunca configurável pelo cliente)", registered.Role, "customer")
	}

	meBeforeLogin := performRequest(t, client, http.MethodGet, app.Server.URL+"/api/v1/auth/me", nil)
	if meBeforeLogin.status != http.StatusUnauthorized {
		t.Fatalf("status de /me sem sessão = %d; esperado = %d", meBeforeLogin.status, http.StatusUnauthorized)
	}

	wrongPassword := performRequest(t, client, http.MethodPost, app.Server.URL+"/api/v1/auth/login", dto.LoginRequest{
		Email:    email,
		Password: "senha-errada-123",
	})
	if wrongPassword.status != http.StatusUnauthorized {
		t.Fatalf("status de login com senha errada = %d; esperado = %d", wrongPassword.status, http.StatusUnauthorized)
	}

	loginResult := performRequest(t, client, http.MethodPost, app.Server.URL+"/api/v1/auth/login", dto.LoginRequest{
		Email:    email,
		Password: password,
	})
	if loginResult.status != http.StatusOK {
		t.Fatalf("status do login = %d; esperado = %d; body = %s", loginResult.status, http.StatusOK, loginResult.body)
	}
	loginResponse := decodeInto[dto.LoginResponse](t, loginResult)
	if loginResponse.User.Email != email || loginResponse.User.Role != "customer" {
		t.Errorf("usuário autenticado = %#v", loginResponse.User)
	}

	meResult := performRequest(t, client, http.MethodGet, app.Server.URL+"/api/v1/auth/me", nil)
	if meResult.status != http.StatusOK {
		t.Fatalf("status de /me autenticado = %d; esperado = %d; body = %s", meResult.status, http.StatusOK, meResult.body)
	}
	me := decodeInto[dto.AuthUserResponse](t, meResult)
	if me.Email != email {
		t.Errorf("/me retornou = %#v; esperado email = %q", me, email)
	}

	logoutResult := performRequest(t, client, http.MethodPost, app.Server.URL+"/api/v1/auth/logout", nil)
	if logoutResult.status != http.StatusNoContent {
		t.Fatalf("status do logout = %d; esperado = %d", logoutResult.status, http.StatusNoContent)
	}

	meAfterLogout := performRequest(t, client, http.MethodGet, app.Server.URL+"/api/v1/auth/me", nil)
	if meAfterLogout.status != http.StatusUnauthorized {
		t.Fatalf("status de /me após logout = %d; esperado = %d", meAfterLogout.status, http.StatusUnauthorized)
	}
}

// TestAuthRegisterDuplicateEmail verifica que a constraint única de e-mail
// do banco (users.email CITEXT UNIQUE) é traduzida para
// domain.ErrUserEmailAlreadyExists e vira 409 na resposta HTTP.
func TestAuthRegisterDuplicateEmail(t *testing.T) {
	app := newTestApp(t)
	client := app.newClient(t)

	request := dto.RegisterRequest{
		Name:     "Primeiro Cadastro",
		Email:    "duplicado.integracao@example.com",
		Password: "SenhaForte@123",
	}

	first := performRequest(t, client, http.MethodPost, app.Server.URL+"/api/v1/auth/register", request)
	if first.status != http.StatusCreated {
		t.Fatalf("status do primeiro registro = %d; esperado = %d; body = %s", first.status, http.StatusCreated, first.body)
	}

	second := performRequest(t, client, http.MethodPost, app.Server.URL+"/api/v1/auth/register", request)
	if second.status != http.StatusConflict {
		t.Fatalf("status do segundo registro = %d; esperado = %d; body = %s", second.status, http.StatusConflict, second.body)
	}
}

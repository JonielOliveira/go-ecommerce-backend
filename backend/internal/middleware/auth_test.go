package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"ecommerce/internal/domain"
)

// fakeJWTService substitui a implementação real de segurança e permite
// controlar a validação do token no teste. Os métodos abaixo fazem este
// tipo satisfazer a interface security.JWTService.
type fakeJWTService struct {
	generateAccessTokenFunc func(userID string) (string, time.Time, error)
	validateAccessTokenFunc func(token string) (string, error)
}

func (fake *fakeJWTService) GenerateAccessToken(userID string) (string, time.Time, error) {
	return fake.generateAccessTokenFunc(userID)
}

func (fake *fakeJWTService) ValidateAccessToken(token string) (string, error) {
	return fake.validateAccessTokenFunc(token)
}

// fakeAuthRepository substitui o PostgreSQL e permite controlar a consulta
// do usuário autenticado no teste. Os métodos abaixo fazem este tipo
// satisfazer a interface repository.AuthRepository.
type fakeAuthRepository struct {
	findAuthenticationByEmailFunc func(ctx context.Context, email string) (*domain.UserAuthentication, error)
	findAuthenticatedUserByIDFunc func(ctx context.Context, userID string) (*domain.AuthenticatedUser, error)
	updateLastLoginAtFunc         func(ctx context.Context, userID string, loginAt time.Time) error
}

func (fake *fakeAuthRepository) FindAuthenticationByEmail(ctx context.Context, email string) (*domain.UserAuthentication, error) {
	return fake.findAuthenticationByEmailFunc(ctx, email)
}

func (fake *fakeAuthRepository) FindAuthenticatedUserByID(ctx context.Context, userID string) (*domain.AuthenticatedUser, error) {
	return fake.findAuthenticatedUserByIDFunc(ctx, userID)
}

func (fake *fakeAuthRepository) UpdateLastLoginAt(ctx context.Context, userID string, loginAt time.Time) error {
	return fake.updateLastLoginAtFunc(ctx, userID, loginAt)
}

const testCookieName = "access_token"

type errorResponse struct {
	Error string `json:"error"`
}

func decodeErrorResponse(t *testing.T, body []byte) errorResponse {
	t.Helper()

	var decoded errorResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decodificar corpo da resposta %q: %v", body, err)
	}

	return decoded
}

// TestGetAuthenticatedUser verifica o contrato de "nunca causa panic":
// contexto vazio ou com um valor de tipo inesperado retornam ok=false, em
// vez de um type assertion sem verificação.
func TestGetAuthenticatedUser(t *testing.T) {
	t.Run("retorna false quando não há usuário no contexto", func(t *testing.T) {
		context, _ := newTestContext(http.MethodGet, "/qualquer")

		user, ok := GetAuthenticatedUser(context)

		if ok || user != nil {
			t.Errorf("resultado = (%v, %v); esperado = (nil, false)", user, ok)
		}
	})

	t.Run("retorna false quando o valor tem tipo inesperado", func(t *testing.T) {
		context, _ := newTestContext(http.MethodGet, "/qualquer")
		context.Set(AuthenticatedUserContextKey, "valor-com-tipo-errado")

		user, ok := GetAuthenticatedUser(context)

		if ok || user != nil {
			t.Errorf("resultado = (%v, %v); esperado = (nil, false)", user, ok)
		}
	})

	t.Run("retorna o usuário quando presente", func(t *testing.T) {
		context, _ := newTestContext(http.MethodGet, "/qualquer")
		want := &domain.AuthenticatedUser{ID: "user-1", Role: domain.RoleCustomer}
		context.Set(AuthenticatedUserContextKey, want)

		got, ok := GetAuthenticatedUser(context)

		if !ok || got != want {
			t.Errorf("resultado = (%v, %v); esperado = (%v, true)", got, ok, want)
		}
	})
}

// TestAuthenticate verifica as três etapas em sequência: presença do
// cookie, validade do JWT, e existência do usuário no Repository — cada uma
// com sua própria mensagem de 401 — e que só a combinação das três injeta o
// usuário no contexto e libera a requisição.
func TestAuthenticate(t *testing.T) {
	t.Run("sem cookie retorna 401 sem validar token", func(t *testing.T) {
		validateCalls := 0
		fakeJWT := &fakeJWTService{
			validateAccessTokenFunc: func(string) (string, error) {
				validateCalls++
				return "", nil
			},
		}
		context, recorder := newTestContext(http.MethodGet, "/qualquer")

		Authenticate(fakeJWT, &fakeAuthRepository{}, testCookieName)(context)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
		if got := decodeErrorResponse(t, recorder.Body.Bytes()); got.Error != "autenticação necessária" {
			t.Errorf("mensagem = %q; esperada = %q", got.Error, "autenticação necessária")
		}
		if validateCalls != 0 {
			t.Errorf("chamadas a ValidateAccessToken = %d; esperado = 0", validateCalls)
		}
	})

	t.Run("cookie vazio retorna 401 sem validar token", func(t *testing.T) {
		validateCalls := 0
		fakeJWT := &fakeJWTService{
			validateAccessTokenFunc: func(string) (string, error) {
				validateCalls++
				return "", nil
			},
		}
		context, recorder := newTestContext(http.MethodGet, "/qualquer")
		context.Request.AddCookie(&http.Cookie{Name: testCookieName, Value: ""})

		Authenticate(fakeJWT, &fakeAuthRepository{}, testCookieName)(context)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
		if validateCalls != 0 {
			t.Errorf("chamadas a ValidateAccessToken = %d; esperado = 0", validateCalls)
		}
	})

	t.Run("token inválido retorna 401 com mensagem específica", func(t *testing.T) {
		fakeJWT := &fakeJWTService{
			validateAccessTokenFunc: func(string) (string, error) { return "", domain.ErrInvalidToken },
		}
		context, recorder := newTestContext(http.MethodGet, "/qualquer")
		context.Request.AddCookie(&http.Cookie{Name: testCookieName, Value: "token-qualquer"})

		Authenticate(fakeJWT, &fakeAuthRepository{}, testCookieName)(context)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
		if got := decodeErrorResponse(t, recorder.Body.Bytes()); got.Error != "token de autenticação inválido" {
			t.Errorf("mensagem = %q; esperada = %q", got.Error, "token de autenticação inválido")
		}
	})

	t.Run("token expirado retorna 401 com mensagem de expiração", func(t *testing.T) {
		fakeJWT := &fakeJWTService{
			validateAccessTokenFunc: func(string) (string, error) { return "", domain.ErrExpiredToken },
		}
		context, recorder := newTestContext(http.MethodGet, "/qualquer")
		context.Request.AddCookie(&http.Cookie{Name: testCookieName, Value: "token-qualquer"})

		Authenticate(fakeJWT, &fakeAuthRepository{}, testCookieName)(context)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
		if got := decodeErrorResponse(t, recorder.Body.Bytes()); got.Error != "token de autenticação expirado" {
			t.Errorf("mensagem = %q; esperada = %q", got.Error, "token de autenticação expirado")
		}
	})

	t.Run("usuário não encontrado no Repository retorna 401", func(t *testing.T) {
		fakeJWT := &fakeJWTService{
			validateAccessTokenFunc: func(string) (string, error) { return "user-1", nil },
		}
		fakeRepository := &fakeAuthRepository{
			findAuthenticatedUserByIDFunc: func(context.Context, string) (*domain.AuthenticatedUser, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		context, recorder := newTestContext(http.MethodGet, "/qualquer")
		context.Request.AddCookie(&http.Cookie{Name: testCookieName, Value: "token-valido"})

		Authenticate(fakeJWT, fakeRepository, testCookieName)(context)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
		if got := decodeErrorResponse(t, recorder.Body.Bytes()); got.Error != "sessão do usuário não é mais válida" {
			t.Errorf("mensagem = %q; esperada = %q", got.Error, "sessão do usuário não é mais válida")
		}
	})

	t.Run("autentica e injeta o usuário no contexto", func(t *testing.T) {
		want := &domain.AuthenticatedUser{ID: "user-1", Name: "Ana Souza", Email: "ana@example.com", Role: domain.RoleCustomer}
		var gotUserID string
		fakeJWT := &fakeJWTService{
			validateAccessTokenFunc: func(token string) (string, error) {
				if token != "token-valido" {
					t.Fatalf("token recebido = %q; esperado = %q", token, "token-valido")
				}
				return "user-1", nil
			},
		}
		fakeRepository := &fakeAuthRepository{
			findAuthenticatedUserByIDFunc: func(_ context.Context, userID string) (*domain.AuthenticatedUser, error) {
				gotUserID = userID
				return want, nil
			},
		}
		context, recorder := newTestContext(http.MethodGet, "/qualquer")
		context.Request.AddCookie(&http.Cookie{Name: testCookieName, Value: "token-valido"})

		Authenticate(fakeJWT, fakeRepository, testCookieName)(context)

		if context.IsAborted() {
			t.Fatalf("contexto não deveria estar abortado; body = %s", recorder.Body)
		}
		if gotUserID != "user-1" {
			t.Errorf("userID enviado ao Repository = %q; esperado = %q", gotUserID, "user-1")
		}
		got, ok := GetAuthenticatedUser(context)
		if !ok || got != want {
			t.Errorf("usuário no contexto = (%v, %v); esperado = (%v, true)", got, ok, want)
		}
	})
}

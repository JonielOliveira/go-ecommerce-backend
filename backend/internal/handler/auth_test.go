package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

// fakeAuthService substitui a Service real diretamente — AuthHandler depende
// da interface service.AuthService, o mesmo padrão usado em OrderHandler.
type fakeAuthService struct {
	loginFunc                     func(ctx context.Context, email string, password string) (*domain.AuthenticatedUser, string, time.Time, error)
	findAuthenticatedUserByIDFunc func(ctx context.Context, userID string) (*domain.AuthenticatedUser, error)
}

func (fake *fakeAuthService) Login(ctx context.Context, email string, password string) (*domain.AuthenticatedUser, string, time.Time, error) {
	return fake.loginFunc(ctx, email, password)
}

func (fake *fakeAuthService) FindAuthenticatedUserByID(ctx context.Context, userID string) (*domain.AuthenticatedUser, error) {
	return fake.findAuthenticatedUserByIDFunc(ctx, userID)
}

var testCookieConfig = CookieConfig{Name: "access_token", SameSite: http.SameSiteLaxMode}

// TestAuthHandlerRegister verifica que o autocadastro repassa direto a
// UserService.Register falsa e o mesmo mapeamento de erro de domínio usado
// por UserHandler.Create para nome/e-mail e e-mail duplicado.
func TestAuthHandlerRegister(t *testing.T) {
	t.Run("corpo inválido retorna 400 sem chamar a Service", func(t *testing.T) {
		serviceCalls := 0
		fakeUser := &fakeUserService{
			registerFunc: func(dto.RegisterRequest) (dto.UserResponse, error) {
				serviceCalls++
				return dto.UserResponse{}, nil
			},
		}
		authHandler := NewAuthHandler(&fakeAuthService{}, fakeUser, testCookieConfig)

		context, recorder := newTestContext(http.MethodPost, "/api/v1/auth/register", []byte("{invalid"))

		serve(context, authHandler.Register)

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
		{name: "registra cliente e retorna 201", wantStatus: http.StatusCreated},
		{name: "email já existe retorna 409", serviceErr: domain.ErrUserEmailAlreadyExists, wantStatus: http.StatusConflict},
		{name: "nome inválido retorna 400", serviceErr: domain.ErrInvalidUserName, wantStatus: http.StatusBadRequest},
		{name: "email inválido retorna 400", serviceErr: domain.ErrInvalidUserEmail, wantStatus: http.StatusBadRequest},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeUser := &fakeUserService{
				registerFunc: func(dto.RegisterRequest) (dto.UserResponse, error) {
					return dto.UserResponse{ID: "user-1"}, testCase.serviceErr
				},
			}
			authHandler := NewAuthHandler(&fakeAuthService{}, fakeUser, testCookieConfig)

			context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/auth/register", dto.RegisterRequest{
				Name: "Ana Souza", Email: "ana@example.com", Password: "segredo123",
			})

			serve(context, authHandler.Register)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestAuthHandlerLogin verifica que o login bem-sucedido define o cookie
// HttpOnly com o token retornado pela Service, e que cada erro de domínio
// vira o status HTTP correto (401 credenciais, 403 inativo, 500 genérico).
func TestAuthHandlerLogin(t *testing.T) {
	t.Run("autentica, define cookie e retorna 200", func(t *testing.T) {
		wantUser := &domain.AuthenticatedUser{ID: "user-1", Name: "Ana Souza", Email: "ana@example.com", Role: domain.RoleCustomer}
		wantExpiresAt := time.Now().Add(15 * time.Minute)
		fakeService := &fakeAuthService{
			loginFunc: func(context.Context, string, string) (*domain.AuthenticatedUser, string, time.Time, error) {
				return wantUser, "signed-token", wantExpiresAt, nil
			},
		}
		authHandler := NewAuthHandler(fakeService, &fakeUserService{}, testCookieConfig)

		context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/auth/login", dto.LoginRequest{
			Email: "ana@example.com", Password: "segredo123",
		})

		serve(context, authHandler.Login)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusOK, recorder.Body)
		}
		got := decodeJSONBody[dto.LoginResponse](t, recorder)
		if got.User.ID != "user-1" || got.User.Role != string(domain.RoleCustomer) {
			t.Errorf("usuário retornado = %#v", got)
		}

		cookies := recorder.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != "access_token" || cookies[0].Value != "signed-token" {
			t.Fatalf("cookies = %#v; esperado access_token=signed-token", cookies)
		}
		if !cookies[0].HttpOnly {
			t.Errorf("cookie de autenticação deveria ser HttpOnly")
		}
	})

	t.Run("corpo inválido retorna 400", func(t *testing.T) {
		authHandler := NewAuthHandler(&fakeAuthService{}, &fakeUserService{}, testCookieConfig)

		context, recorder := newTestContext(http.MethodPost, "/api/v1/auth/login", []byte("{invalid"))

		serve(context, authHandler.Login)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
	})

	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "credenciais inválidas retornam 401", serviceErr: domain.ErrInvalidCredentials, wantStatus: http.StatusUnauthorized},
		{name: "usuário inativo retorna 403", serviceErr: domain.ErrUserInactive, wantStatus: http.StatusForbidden},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeAuthService{
				loginFunc: func(context.Context, string, string) (*domain.AuthenticatedUser, string, time.Time, error) {
					return nil, "", time.Time{}, testCase.serviceErr
				},
			}
			authHandler := NewAuthHandler(fakeService, &fakeUserService{}, testCookieConfig)

			context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/auth/login", dto.LoginRequest{
				Email: "ana@example.com", Password: "segredo123",
			})

			serve(context, authHandler.Login)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestAuthHandlerLogout verifica que a rota pública sempre limpa o cookie
// (MaxAge negativo), independentemente de existir sessão válida.
func TestAuthHandlerLogout(t *testing.T) {
	authHandler := NewAuthHandler(&fakeAuthService{}, &fakeUserService{}, testCookieConfig)

	context, recorder := newTestContext(http.MethodPost, "/api/v1/auth/logout", nil)

	serve(context, authHandler.Logout)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusNoContent)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "access_token" || cookies[0].MaxAge >= 0 {
		t.Fatalf("cookie de logout = %#v; esperado MaxAge negativo para limpar o cookie", cookies)
	}
}

// TestAuthHandlerMe verifica o repasse do usuário já autenticado pelo
// middleware (via contexto) e o 401 quando, por algum motivo, ele não está
// presente.
func TestAuthHandlerMe(t *testing.T) {
	t.Run("retorna usuário autenticado e 200", func(t *testing.T) {
		authHandler := NewAuthHandler(&fakeAuthService{}, &fakeUserService{}, testCookieConfig)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/auth/me", nil)
		setAuthenticatedUser(context, testCustomer)

		serve(context, authHandler.Me)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusOK, recorder.Body)
		}
		got := decodeJSONBody[dto.AuthUserResponse](t, recorder)
		if got.ID != testCustomer.ID {
			t.Errorf("usuário retornado = %#v", got)
		}
	})

	t.Run("sem usuário autenticado retorna 401", func(t *testing.T) {
		authHandler := NewAuthHandler(&fakeAuthService{}, &fakeUserService{}, testCookieConfig)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/auth/me", nil)

		serve(context, authHandler.Me)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusUnauthorized)
		}
	})
}

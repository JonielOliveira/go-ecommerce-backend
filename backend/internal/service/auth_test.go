package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"ecommerce/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

// fakeAuthRepository substitui o PostgreSQL e permite controlar cada
// operação no teste. Os métodos abaixo fazem este tipo satisfazer a
// interface repository.AuthRepository.
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

// fakeJWTService substitui a implementação real de segurança e permite
// controlar cada operação no teste. Os métodos abaixo fazem este tipo
// satisfazer a interface security.JWTService.
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

// mustUserAuthentication monta um registro de autenticação com a senha
// informada já em hash, pronto para ser devolvido pelo fake de Repository.
func mustUserAuthentication(t *testing.T, userID, email, password string, role domain.UserRole, active bool, deletedAt *time.Time) *domain.UserAuthentication {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("gerar hash de senha de teste: %v", err)
	}

	return &domain.UserAuthentication{
		UserID:            userID,
		Name:              "Ana Souza",
		Email:             email,
		Role:              role,
		Active:            active,
		DeletedAt:         deletedAt,
		PasswordHash:      string(hash),
		PasswordChangedAt: time.Now(),
	}
}

// TestAuthServiceLogin verifica a orquestração do login: validação de senha,
// bloqueio de contas removidas/inativas, geração de token e atualização do
// último login — nessa ordem, e sem vazar qual etapa falhou nas credenciais.
func TestAuthServiceLogin(t *testing.T) {
	t.Run("autentica, atualiza último login e gera token", func(t *testing.T) {
		authRecord := mustUserAuthentication(t, "user-1", "ana@example.com", "segredo123", domain.RoleCustomer, true, nil)

		var gotEmail string
		var gotLastLoginUserID string
		var gotLastLoginAt time.Time
		fakeRepository := &fakeAuthRepository{
			findAuthenticationByEmailFunc: func(_ context.Context, email string) (*domain.UserAuthentication, error) {
				gotEmail = email
				return authRecord, nil
			},
			updateLastLoginAtFunc: func(_ context.Context, userID string, loginAt time.Time) error {
				gotLastLoginUserID = userID
				gotLastLoginAt = loginAt
				return nil
			},
		}
		wantToken := "signed-token"
		wantExpiresAt := time.Now().Add(15 * time.Minute)
		fakeJWT := &fakeJWTService{
			generateAccessTokenFunc: func(userID string) (string, time.Time, error) {
				if userID != "user-1" {
					t.Fatalf("userID recebido = %s; esperado = user-1", userID)
				}
				return wantToken, wantExpiresAt, nil
			},
		}
		authService := NewAuthService(fakeRepository, fakeJWT)

		user, token, expiresAt, err := authService.Login(context.Background(), "  ana@example.com  ", "segredo123")

		if err != nil {
			t.Fatalf("Login retornou erro inesperado: %v", err)
		}
		if gotEmail != "ana@example.com" {
			t.Errorf("email enviado ao Repository = %q; esperado sem espaços = %q", gotEmail, "ana@example.com")
		}
		if token != wantToken || !expiresAt.Equal(wantExpiresAt) {
			t.Errorf("token/expiração = %q/%v; esperado = %q/%v", token, expiresAt, wantToken, wantExpiresAt)
		}
		if user.ID != "user-1" || user.Role != domain.RoleCustomer {
			t.Errorf("usuário autenticado = %#v", user)
		}
		if gotLastLoginUserID != "user-1" {
			t.Errorf("último login registrado para userID = %q; esperado = user-1", gotLastLoginUserID)
		}
		if gotLastLoginAt.IsZero() {
			t.Errorf("último login registrado com timestamp zero")
		}
	})

	t.Run("recusa senha incorreta sem gerar token", func(t *testing.T) {
		authRecord := mustUserAuthentication(t, "user-1", "ana@example.com", "segredo123", domain.RoleCustomer, true, nil)
		generateCalls := 0
		updateCalls := 0
		fakeRepository := &fakeAuthRepository{
			findAuthenticationByEmailFunc: func(context.Context, string) (*domain.UserAuthentication, error) { return authRecord, nil },
			updateLastLoginAtFunc: func(context.Context, string, time.Time) error {
				updateCalls++
				return nil
			},
		}
		fakeJWT := &fakeJWTService{
			generateAccessTokenFunc: func(string) (string, time.Time, error) {
				generateCalls++
				return "", time.Time{}, nil
			},
		}
		authService := NewAuthService(fakeRepository, fakeJWT)

		_, _, _, err := authService.Login(context.Background(), "ana@example.com", "senhaErrada")

		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrInvalidCredentials)
		}
		if generateCalls != 0 || updateCalls != 0 {
			t.Errorf("chamadas geradas = token:%d login:%d; esperado 0 para ambas", generateCalls, updateCalls)
		}
	})

	t.Run("recusa usuário removido mesmo com senha correta", func(t *testing.T) {
		deletedAt := time.Now()
		authRecord := mustUserAuthentication(t, "user-1", "ana@example.com", "segredo123", domain.RoleCustomer, true, &deletedAt)
		fakeRepository := &fakeAuthRepository{
			findAuthenticationByEmailFunc: func(context.Context, string) (*domain.UserAuthentication, error) { return authRecord, nil },
		}
		authService := NewAuthService(fakeRepository, &fakeJWTService{})

		_, _, _, err := authService.Login(context.Background(), "ana@example.com", "segredo123")

		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrInvalidCredentials)
		}
	})

	t.Run("recusa usuário inativo", func(t *testing.T) {
		authRecord := mustUserAuthentication(t, "user-1", "ana@example.com", "segredo123", domain.RoleCustomer, false, nil)
		fakeRepository := &fakeAuthRepository{
			findAuthenticationByEmailFunc: func(context.Context, string) (*domain.UserAuthentication, error) { return authRecord, nil },
		}
		authService := NewAuthService(fakeRepository, &fakeJWTService{})

		_, _, _, err := authService.Login(context.Background(), "ana@example.com", "segredo123")

		if !errors.Is(err, domain.ErrUserInactive) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrUserInactive)
		}
	})

	t.Run("propaga erro quando Repository não encontra o email", func(t *testing.T) {
		fakeRepository := &fakeAuthRepository{
			findAuthenticationByEmailFunc: func(context.Context, string) (*domain.UserAuthentication, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		authService := NewAuthService(fakeRepository, &fakeJWTService{})

		_, _, _, err := authService.Login(context.Background(), "ana@example.com", "segredo123")

		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrUserNotFound)
		}
	})

	t.Run("propaga erro ao gerar token sem atualizar último login", func(t *testing.T) {
		authRecord := mustUserAuthentication(t, "user-1", "ana@example.com", "segredo123", domain.RoleCustomer, true, nil)
		updateCalls := 0
		tokenErr := errors.New("falha ao assinar token")
		fakeRepository := &fakeAuthRepository{
			findAuthenticationByEmailFunc: func(context.Context, string) (*domain.UserAuthentication, error) { return authRecord, nil },
			updateLastLoginAtFunc: func(context.Context, string, time.Time) error {
				updateCalls++
				return nil
			},
		}
		fakeJWT := &fakeJWTService{
			generateAccessTokenFunc: func(string) (string, time.Time, error) {
				return "", time.Time{}, tokenErr
			},
		}
		authService := NewAuthService(fakeRepository, fakeJWT)

		_, _, _, err := authService.Login(context.Background(), "ana@example.com", "segredo123")

		if !errors.Is(err, tokenErr) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, tokenErr)
		}
		if updateCalls != 0 {
			t.Errorf("chamadas a UpdateLastLoginAt = %d; esperado = 0", updateCalls)
		}
	})

	t.Run("propaga erro ao atualizar último login", func(t *testing.T) {
		authRecord := mustUserAuthentication(t, "user-1", "ana@example.com", "segredo123", domain.RoleCustomer, true, nil)
		loginErr := errors.New("banco indisponível")
		fakeRepository := &fakeAuthRepository{
			findAuthenticationByEmailFunc: func(context.Context, string) (*domain.UserAuthentication, error) { return authRecord, nil },
			updateLastLoginAtFunc: func(context.Context, string, time.Time) error {
				return loginErr
			},
		}
		fakeJWT := &fakeJWTService{
			generateAccessTokenFunc: func(string) (string, time.Time, error) {
				return "signed-token", time.Now().Add(15 * time.Minute), nil
			},
		}
		authService := NewAuthService(fakeRepository, fakeJWT)

		user, token, _, err := authService.Login(context.Background(), "ana@example.com", "segredo123")

		if !errors.Is(err, loginErr) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, loginErr)
		}
		if user != nil || token != "" {
			t.Errorf("retorno em caso de erro = user:%#v token:%q; esperado ambos zerados", user, token)
		}
	})
}

// TestAuthServiceFindAuthenticatedUserByID verifica o repasse direto ao
// Repository, usado pelo middleware de autenticação a cada requisição.
func TestAuthServiceFindAuthenticatedUserByID(t *testing.T) {
	want := &domain.AuthenticatedUser{ID: "user-1", Name: "Ana Souza", Email: "ana@example.com", Role: domain.RoleCustomer}
	fakeRepository := &fakeAuthRepository{
		findAuthenticatedUserByIDFunc: func(_ context.Context, userID string) (*domain.AuthenticatedUser, error) {
			if userID != "user-1" {
				t.Fatalf("userID recebido = %s; esperado = user-1", userID)
			}
			return want, nil
		},
	}
	authService := NewAuthService(fakeRepository, &fakeJWTService{})

	got, err := authService.FindAuthenticatedUserByID(context.Background(), "user-1")

	if err != nil {
		t.Fatalf("FindAuthenticatedUserByID retornou erro inesperado: %v", err)
	}
	if got != want {
		t.Errorf("usuário recebido = %#v; esperado = %#v", got, want)
	}
}

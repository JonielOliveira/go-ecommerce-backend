package security

import (
	"errors"
	"testing"
	"time"

	"ecommerce/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret   = "test-secret-at-least-32-bytes-long"
	testIssuer   = "ecommerce-api"
	testAudience = "ecommerce-clients"
	testTTL      = 15 * time.Minute
)

var testUserID = "019535d9-3df7-7001-8000-000000000001"

func newTestJWTService() JWTService {
	return NewJWTService(testSecret, testIssuer, testAudience, testTTL)
}

// mustSignToken monta um token assinado com HS256 fora do JWTService, para
// que os testes de ValidateAccessToken consigam produzir combinações de
// claims que o próprio serviço nunca geraria (issuer errado, sem expiração,
// subject fora do formato UUID, ...).
func mustSignToken(t *testing.T, secret string, claims jwt.RegisteredClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("assinar token de teste: %v", err)
	}

	return signed
}

// TestJWTServiceGenerateAndValidateAccessToken verifica o ciclo completo:
// o token gerado por GenerateAccessToken é aceito por ValidateAccessToken e
// devolve o mesmo userID, e a expiração respeita o TTL configurado.
func TestJWTServiceGenerateAndValidateAccessToken(t *testing.T) {
	service := newTestJWTService()
	before := time.Now()

	token, expiresAt, err := service.GenerateAccessToken(testUserID)
	if err != nil {
		t.Fatalf("GenerateAccessToken retornou erro inesperado: %v", err)
	}
	if token == "" {
		t.Fatalf("token gerado está vazio")
	}

	wantExpiresAt := before.Add(testTTL)
	if diff := expiresAt.Sub(wantExpiresAt); diff < -2*time.Second || diff > 2*time.Second {
		t.Errorf("expiresAt = %v; esperado próximo de %v (TTL = %v)", expiresAt, wantExpiresAt, testTTL)
	}

	gotUserID, err := service.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken retornou erro inesperado: %v", err)
	}
	if gotUserID != testUserID {
		t.Errorf("userID recebido = %q; esperado = %q", gotUserID, testUserID)
	}
}

// TestJWTServiceValidateAccessToken verifica que cada forma de token
// inválido — mal formado, assinado com outro segredo, issuer/audience
// divergentes, sem expiração, subject vazio ou fora do formato UUID, e
// algoritmo "none" — é rejeitada com domain.ErrInvalidToken.
func TestJWTServiceValidateAccessToken(t *testing.T) {
	testCases := []struct {
		name  string
		token func(t *testing.T) string
	}{
		{
			name:  "token vazio",
			token: func(*testing.T) string { return "" },
		},
		{
			name:  "token malformado",
			token: func(*testing.T) string { return "isto-nao-e-um-jwt" },
		},
		{
			name: "assinado com segredo diferente",
			token: func(t *testing.T) string {
				now := time.Now()
				return mustSignToken(t, "outro-segredo-completamente-diferente", jwt.RegisteredClaims{
					Subject:   testUserID,
					Issuer:    testIssuer,
					Audience:  jwt.ClaimStrings{testAudience},
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
				})
			},
		},
		{
			name: "issuer diferente",
			token: func(t *testing.T) string {
				now := time.Now()
				return mustSignToken(t, testSecret, jwt.RegisteredClaims{
					Subject:   testUserID,
					Issuer:    "outro-issuer",
					Audience:  jwt.ClaimStrings{testAudience},
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
				})
			},
		},
		{
			name: "audience diferente",
			token: func(t *testing.T) string {
				now := time.Now()
				return mustSignToken(t, testSecret, jwt.RegisteredClaims{
					Subject:   testUserID,
					Issuer:    testIssuer,
					Audience:  jwt.ClaimStrings{"outra-audience"},
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
				})
			},
		},
		{
			name: "sem data de expiração",
			token: func(t *testing.T) string {
				now := time.Now()
				return mustSignToken(t, testSecret, jwt.RegisteredClaims{
					Subject:  testUserID,
					Issuer:   testIssuer,
					Audience: jwt.ClaimStrings{testAudience},
					IssuedAt: jwt.NewNumericDate(now),
				})
			},
		},
		{
			name: "subject vazio",
			token: func(t *testing.T) string {
				now := time.Now()
				return mustSignToken(t, testSecret, jwt.RegisteredClaims{
					Subject:   "",
					Issuer:    testIssuer,
					Audience:  jwt.ClaimStrings{testAudience},
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
				})
			},
		},
		{
			name: "subject não é um UUID",
			token: func(t *testing.T) string {
				now := time.Now()
				return mustSignToken(t, testSecret, jwt.RegisteredClaims{
					Subject:   "not-a-uuid",
					Issuer:    testIssuer,
					Audience:  jwt.ClaimStrings{testAudience},
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
				})
			},
		},
		{
			name: "algoritmo none não é aceito",
			token: func(t *testing.T) string {
				now := time.Now()
				token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
					Subject:   testUserID,
					Issuer:    testIssuer,
					Audience:  jwt.ClaimStrings{testAudience},
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
				})
				signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
				if err != nil {
					t.Fatalf("assinar token sem algoritmo: %v", err)
				}
				return signed
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := newTestJWTService()

			_, err := service.ValidateAccessToken(testCase.token(t))

			if !errors.Is(err, domain.ErrInvalidToken) {
				t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrInvalidToken)
			}
		})
	}
}

// TestJWTServiceValidateAccessTokenExpired verifica que um token expirado
// recebe um erro distinto (domain.ErrExpiredToken) dos demais tokens
// inválidos — o middleware de autenticação usa essa distinção para escolher
// a mensagem de erro.
func TestJWTServiceValidateAccessTokenExpired(t *testing.T) {
	service := newTestJWTService()
	now := time.Now()

	expired := mustSignToken(t, testSecret, jwt.RegisteredClaims{
		Subject:   testUserID,
		Issuer:    testIssuer,
		Audience:  jwt.ClaimStrings{testAudience},
		IssuedAt:  jwt.NewNumericDate(now.Add(-time.Hour)),
		ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute)),
	})

	_, err := service.ValidateAccessToken(expired)

	if !errors.Is(err, domain.ErrExpiredToken) {
		t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrExpiredToken)
	}
}

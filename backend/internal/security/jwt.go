package security

import (
	"errors"
	"regexp"
	"time"

	"ecommerce/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateAccessToken(userID string) (string, time.Time, error)
	ValidateAccessToken(token string) (string, error)
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type jwtService struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewJWTService(secret, issuer, audience string, ttl time.Duration) JWTService {
	return &jwtService{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}
}

func (s *jwtService) GenerateAccessToken(userID string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.ttl)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    s.issuer,
		Audience:  jwt.ClaimStrings{s.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (s *jwtService) ValidateAccessToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, domain.ErrInvalidToken
			}

			return s.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", domain.ErrExpiredToken
		}

		return "", domain.ErrInvalidToken
	}

	if !token.Valid {
		return "", domain.ErrInvalidToken
	}

	subject := claims.Subject
	if subject == "" || !uuidPattern.MatchString(subject) {
		return "", domain.ErrInvalidToken
	}

	return subject, nil
}

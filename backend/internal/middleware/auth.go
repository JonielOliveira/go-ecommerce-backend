package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/domain"
	"ecommerce/internal/repository"
	"ecommerce/internal/security"
)

const AuthenticatedUserContextKey = "authenticated_user"

// GetAuthenticatedUser recupera o usuário autenticado colocado no contexto
// pelo middleware Authenticate. Nunca causa panic: retorna ok=false quando
// o valor não existe ou possui tipo inesperado.
func GetAuthenticatedUser(c *gin.Context) (*domain.AuthenticatedUser, bool) {
	value, exists := c.Get(AuthenticatedUserContextKey)
	if !exists {
		return nil, false
	}

	user, ok := value.(*domain.AuthenticatedUser)
	if !ok {
		return nil, false
	}

	return user, true
}

// Authenticate lê o JWT do cookie de autenticação, valida assinatura,
// emissor, audiência e expiração, e então confirma no banco que o usuário
// ainda existe, está ativo e não foi excluído antes de liberar a requisição.
func Authenticate(
	jwtService security.JWTService,
	authRepository repository.AuthRepository,
	cookieName string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err != nil || strings.TrimSpace(token) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "autenticação necessária",
			})
			return
		}

		userID, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			if errors.Is(err, domain.ErrExpiredToken) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "token de autenticação expirado",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token de autenticação inválido",
			})
			return
		}

		user, err := authRepository.FindAuthenticatedUserByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "sessão do usuário não é mais válida",
			})
			return
		}

		c.Set(AuthenticatedUserContextKey, user)
		c.Next()
	}
}

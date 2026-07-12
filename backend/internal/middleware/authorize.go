package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/domain"
)

// RequireRole autoriza a requisição somente se o usuário autenticado
// possuir um dos papéis permitidos. Deve ser usado após Authenticate: a
// ausência do usuário no contexto é tratada como falha de autenticação
// (401), não como falta de permissão (403).
func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := GetAuthenticatedUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		for _, role := range allowedRoles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "you do not have permission to perform this operation",
		})
	}
}

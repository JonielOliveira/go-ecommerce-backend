package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

// RegisterUserRoutes registra as rotas administrativas de usuário. Todo o
// grupo exige autenticação e papel "admin" — o autocadastro público vive em
// POST /auth/register, não aqui.
func RegisterUserRoutes(
	router *gin.RouterGroup,
	handler *handler.UserHandler,
	authenticate gin.HandlerFunc,
	requireAdmin gin.HandlerFunc,
) {
	users := router.Group("/users")
	users.Use(authenticate, requireAdmin)
	{
		// Create
		users.POST("", handler.Create)

		// Read
		users.GET("", handler.Search)
		users.GET("/:id", handler.FindByID)

		// Update
		users.PUT("/:id", handler.Update)
		users.PATCH("/:id/restore", handler.RestoreByID)
		users.PATCH("/:id/activate", handler.ActivateByID)
		users.PATCH("/:id/deactivate", handler.DeactivateByID)

		// Delete
		users.DELETE("/:id", handler.DeleteByID)
	}
}

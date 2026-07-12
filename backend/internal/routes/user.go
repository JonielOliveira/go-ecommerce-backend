package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

func RegisterUserRoutes(
	router *gin.RouterGroup,
	handler *handler.UserHandler,
	authenticate gin.HandlerFunc,
	requireAdmin gin.HandlerFunc,
) {
	users := router.Group("/users")
	{
		// Create (público - cadastro; sempre cria papel "customer")
		users.POST("", handler.Create)

		// Read (admin)
		users.GET("", authenticate, requireAdmin, handler.Search)
		users.GET("/:id", authenticate, requireAdmin, handler.FindByID)

		// Update (admin)
		users.PUT("/:id", authenticate, requireAdmin, handler.Update)
		users.PATCH("/:id/restore", authenticate, requireAdmin, handler.RestoreByID)
		users.PATCH("/:id/activate", authenticate, requireAdmin, handler.ActivateByID)
		users.PATCH("/:id/deactivate", authenticate, requireAdmin, handler.DeactivateByID)

		// Delete (admin)
		users.DELETE("/:id", authenticate, requireAdmin, handler.DeleteByID)
	}
}

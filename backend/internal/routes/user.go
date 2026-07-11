package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

func RegisterUserRoutes(router *gin.RouterGroup, handler *handler.UserHandler) {
	users := router.Group("/users")
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

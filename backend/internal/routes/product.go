package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

func RegisterProductRoutes(
	router *gin.RouterGroup,
	handler *handler.ProductHandler,
	authenticate gin.HandlerFunc,
	requireAdmin gin.HandlerFunc,
) {
	products := router.Group("/products")
	{
		// Read (público)
		products.GET("", handler.Search)
		products.GET("/:id", handler.FindByID)

		// Create (admin)
		products.POST("", authenticate, requireAdmin, handler.Create)

		// Update (admin)
		products.PUT("/:id", authenticate, requireAdmin, handler.Update)
		products.PATCH("/:id/restore", authenticate, requireAdmin, handler.RestoreByID)
		products.PATCH("/:id/activate", authenticate, requireAdmin, handler.ActivateByID)
		products.PATCH("/:id/deactivate", authenticate, requireAdmin, handler.DeactivateByID)

		// Delete (admin)
		products.DELETE("/:id", authenticate, requireAdmin, handler.DeleteByID)
	}
}

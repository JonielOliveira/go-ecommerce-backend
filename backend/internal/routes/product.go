package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

func RegisterProductRoutes(router *gin.RouterGroup, handler *handler.ProductHandler) {
	products := router.Group("/products")
	{
		// Create
		products.POST("", handler.Create)

		// Read
		products.GET("", handler.Search)
		products.GET("/:id", handler.FindByID)

		// Update
		products.PUT("/:id", handler.Update)
		products.PATCH("/:id/restore", handler.RestoreByID)
		products.PATCH("/:id/activate", handler.ActivateByID)
		products.PATCH("/:id/deactivate", handler.DeactivateByID)

		// Delete
		products.DELETE("/:id", handler.DeleteByID)

	}
}

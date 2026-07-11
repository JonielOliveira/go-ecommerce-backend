package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

type Handlers struct {
	Health  *handler.HealthHandler
	Product *handler.ProductHandler
	User    *handler.UserHandler
	// Order *handler.OrderHandler
}

func Register(router *gin.Engine, handlers Handlers) {
	// Infraestrutura
	RegisterHealthRoutes(router, handlers.Health)

	// API
	v1 := router.Group("/api/v1")
	{
		RegisterProductRoutes(v1, handlers.Product)
		RegisterUserRoutes(v1, handlers.User)
		// RegisterOrderRoutes(v1, handlers.Order)
	}
}

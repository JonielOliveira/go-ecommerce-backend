package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

type Handlers struct {
	Health  *handler.HealthHandler
	Product *handler.ProductHandler
	User    *handler.UserHandler
	Auth    *handler.AuthHandler
	// Order *handler.OrderHandler
}

// Middlewares agrupa os middlewares globais de autenticação e autorização
// usados na composição das rotas privadas/administrativas.
type Middlewares struct {
	Authenticate gin.HandlerFunc
	RequireAdmin gin.HandlerFunc
}

func Register(router *gin.Engine, handlers Handlers, middlewares Middlewares) {
	// Infraestrutura
	RegisterHealthRoutes(router, handlers.Health)

	// API
	v1 := router.Group("/api/v1")
	{
		RegisterAuthRoutes(v1, handlers.Auth, middlewares.Authenticate)
		RegisterProductRoutes(v1, handlers.Product, middlewares.Authenticate, middlewares.RequireAdmin)
		RegisterUserRoutes(v1, handlers.User, middlewares.Authenticate, middlewares.RequireAdmin)
		// RegisterOrderRoutes(v1, handlers.Order)
	}
}

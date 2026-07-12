package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

func RegisterAuthRoutes(router *gin.RouterGroup, handler *handler.AuthHandler, authenticate gin.HandlerFunc) {
	auth := router.Group("/auth")
	{
		// Públicas
		auth.POST("/login", handler.Login)
		auth.POST("/logout", handler.Logout)

		// Privada (qualquer usuário autenticado)
		auth.GET("/me", authenticate, handler.Me)
	}
}

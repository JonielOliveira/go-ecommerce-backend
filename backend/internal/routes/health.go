package routes

import (
	"ecommerce/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterHealthRoutes(router *gin.Engine, handler *handler.HealthHandler) {
	router.GET("/health", handler.Health)
}

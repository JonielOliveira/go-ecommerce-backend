package routes

import (
	"github.com/gin-gonic/gin"

	"ecommerce/internal/handler"
)

// RegisterOrderRoutes registra as rotas de pedido. Todo o grupo exige
// autenticação e papel customer ou admin — a distinção de permissão por
// operação (dono vs qualquer pedido) acontece no service, não aqui.
func RegisterOrderRoutes(
	router *gin.RouterGroup,
	handler *handler.OrderHandler,
	authenticate gin.HandlerFunc,
	requireCustomerOrAdmin gin.HandlerFunc,
) {
	orders := router.Group("/orders")
	orders.Use(authenticate, requireCustomerOrAdmin)
	{
		orders.POST("", handler.Create)
		orders.GET("", handler.Search)
		orders.GET("/:id", handler.FindByID)
		orders.POST("/:id/pay", handler.PayByID)
		orders.POST("/:id/cancel", handler.CancelByID)
	}
}

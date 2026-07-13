package handler

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/middleware"
	"ecommerce/internal/service"
)

var orderIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type OrderHandler struct {
	service service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

// authenticatedUser recupera o usuário autenticado do contexto (já
// carregado pelo middleware Authenticate) e escreve a resposta 401 caso,
// por algum motivo inesperado, ele não esteja presente.
func (h *OrderHandler) authenticatedUser(c *gin.Context) (*domain.AuthenticatedUser, bool) {
	user, ok := middleware.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "autenticação necessária",
		})
	}

	return user, ok
}

// Create godoc
// @Summary Create order
// @Description Creates a new order for the authenticated user. The order owner always comes from the authenticated session — customer and admin cannot send customer_id/user_id/owner_id in the body.
// @Tags orders
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param order body dto.CreateOrderRequest true "Order items"
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	user, ok := h.authenticatedUser(c)
	if !ok {
		return
	}

	var request dto.CreateOrderRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "corpo da requisição inválido",
		})
		return
	}

	response, err := h.service.Create(c.Request.Context(), *user, request)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderMustHaveItems),
			errors.Is(err, domain.ErrInvalidOrderItem),
			errors.Is(err, domain.ErrInvalidOrderQuantity):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductUnavailable),
			errors.Is(err, domain.ErrInsufficientStock):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno do servidor",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, response)
}

// Search godoc
// @Summary List orders
// @Description Lists orders with pagination, using the same page/pageSize format as products and users. Customer sees only their own orders (and is counted only among their own); admin sees and counts all.
// @Tags orders
// @Produce json
// @Security CookieAuth
// @Param page query int false "Page number, starting at 1 (default 1)"
// @Param pageSize query int false "Orders per page (default 20, max 100)"
// @Success 200 {object} dto.OrderPageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/orders [get]
func (h *OrderHandler) Search(c *gin.Context) {
	user, ok := h.authenticatedUser(c)
	if !ok {
		return
	}

	var request dto.OrderSearchRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parâmetros de busca inválidos",
		})
		return
	}

	response, err := h.service.Search(c.Request.Context(), *user, request.Page, request.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "erro interno do servidor",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// FindByID godoc
// @Summary Find order by ID
// @Description Returns an order with its items. Customer can only access their own orders; admin can access any order. Orders belonging to another user are reported as not found.
// @Tags orders
// @Produce json
// @Security CookieAuth
// @Param id path string true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) FindByID(c *gin.Context) {
	user, ok := h.authenticatedUser(c)
	if !ok {
		return
	}

	id := c.Param("id")
	if !orderIDPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id do pedido inválido",
		})
		return
	}

	response, err := h.service.FindByID(c.Request.Context(), *user, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound), errors.Is(err, domain.ErrOrderAccessDenied):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "pedido não encontrado",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno do servidor",
			})
		}

		return
	}

	c.JSON(http.StatusOK, response)
}

// PayByID godoc
// @Summary Pay order
// @Description Moves an order from PENDING to PAID. Only the order owner can pay it, whether customer or admin — admin does not get a bypass to pay other people's orders.
// @Tags orders
// @Produce json
// @Security CookieAuth
// @Param id path string true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/orders/{id}/pay [post]
func (h *OrderHandler) PayByID(c *gin.Context) {
	user, ok := h.authenticatedUser(c)
	if !ok {
		return
	}

	id := c.Param("id")
	if !orderIDPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id do pedido inválido",
		})
		return
	}

	response, err := h.service.PayByID(c.Request.Context(), *user, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound), errors.Is(err, domain.ErrOrderAccessDenied):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "pedido não encontrado",
			})

		case errors.Is(err, domain.ErrOrderCannotBePaid):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno do servidor",
			})
		}

		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelByID godoc
// @Summary Cancel order
// @Description Moves an order from PENDING to CANCELED and returns all quantities to stock. Customer can only cancel their own orders; admin can cancel any order.
// @Tags orders
// @Produce json
// @Security CookieAuth
// @Param id path string true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/orders/{id}/cancel [post]
func (h *OrderHandler) CancelByID(c *gin.Context) {
	user, ok := h.authenticatedUser(c)
	if !ok {
		return
	}

	id := c.Param("id")
	if !orderIDPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id do pedido inválido",
		})
		return
	}

	response, err := h.service.CancelByID(c.Request.Context(), *user, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound), errors.Is(err, domain.ErrOrderAccessDenied):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "pedido não encontrado",
			})

		case errors.Is(err, domain.ErrOrderCannotBeCanceled):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno do servidor",
			})
		}

		return
	}

	c.JSON(http.StatusOK, response)
}

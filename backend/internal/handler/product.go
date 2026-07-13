package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/service"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

// Create godoc
// @Summary Create product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param product body dto.ProductRequest true "Product data"
// @Success 201 {object} dto.ProductResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var request dto.ProductRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "corpo da requisição inválido",
		})
		return
	}

	response, err := h.service.Create(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Update godoc
// @Summary Update product
// @Description Update an existing product. Requires authentication and the "admin" role.
// @Tags products
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Product ID"
// @Param product body dto.ProductUpdateRequest true "Product data"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var request dto.ProductUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "corpo da requisição inválido",
		})
		return
	}

	response, err := h.service.Update(id, request)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductAlreadyDeleted):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrInvalidProductName),
			errors.Is(err, domain.ErrInvalidProductDescription),
			errors.Is(err, domain.ErrInvalidProductPrice),
			errors.Is(err, domain.ErrInvalidProductStock):

			c.JSON(http.StatusBadRequest, gin.H{
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

// FindByID godoc
// @Summary Find product by ID
// @Description Returns a product by its ID. Public route.
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} dto.ProductResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	response, err := h.service.FindByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "erro interno do servidor",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// func (h *ProductHandler) FindAll(c *gin.Context) {
// 	response, err := h.service.FindAll()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "erro interno do servidor",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// Search godoc
// @Summary List products
// @Description Lists products with filters and pagination. Public route.
// @Tags products
// @Produce json
// @Param name query string false "Filter by name (partial match)"
// @Param categoryId query string false "Filter by category ID"
// @Param active query bool false "Filter by active status"
// @Param deletionState query string false "Filter by deletion state" Enums(not_deleted, deleted, all)
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param page query int false "Page number, starting at 1 (default 1)"
// @Param pageSize query int false "Products per page (default 20, max 100)"
// @Success 200 {object} dto.ProductPageResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/products [get]
func (h *ProductHandler) Search(c *gin.Context) {
	var request dto.ProductSearchRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parâmetros de busca inválidos",
		})
		return
	}

	response, err := h.service.Search(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "erro interno do servidor",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteByID godoc
// @Summary Delete product
// @Description Soft-deletes a product. Requires authentication and the "admin" role.
// @Tags products
// @Produce json
// @Security CookieAuth
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteByID(id); err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductAlreadyDeleted):
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

	c.Status(http.StatusNoContent)
}

// RestoreByID godoc
// @Summary Restore product
// @Description Restores a soft-deleted product. Requires authentication and the "admin" role.
// @Tags products
// @Produce json
// @Security CookieAuth
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/products/{id}/restore [patch]
func (h *ProductHandler) RestoreByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.RestoreByID(id); err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductNotDeleted):
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

	c.Status(http.StatusNoContent)
}

// ActivateByID godoc
// @Summary Activate product
// @Description Activates a product. Requires authentication and the "admin" role.
// @Tags products
// @Produce json
// @Security CookieAuth
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/products/{id}/activate [patch]
func (h *ProductHandler) ActivateByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.ActivateByID(id); err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductAlreadyDeleted):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductAlreadyActive):
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

	c.Status(http.StatusNoContent)
}

// DeactivateByID godoc
// @Summary Deactivate product
// @Description Deactivates a product. Requires authentication and the "admin" role.
// @Tags products
// @Produce json
// @Security CookieAuth
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/products/{id}/deactivate [patch]
func (h *ProductHandler) DeactivateByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeactivateByID(id); err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductAlreadyDeleted):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrProductAlreadyInactive):
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

	c.Status(http.StatusNoContent)
}

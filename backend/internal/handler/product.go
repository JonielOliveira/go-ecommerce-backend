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

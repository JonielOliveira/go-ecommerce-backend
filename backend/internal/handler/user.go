package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Create godoc
// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UserRequest true "User data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var request dto.UserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "corpo da requisição inválido",
		})
		return
	}

	response, err := h.service.Create(request)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrInvalidUserName),
			errors.Is(err, domain.ErrInvalidUserEmail),
			errors.Is(err, domain.ErrInvalidUserRole):

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

	c.JSON(http.StatusCreated, response)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var request dto.UserUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "corpo da requisição inválido",
		})
		return
	}

	response, err := h.service.Update(id, request)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrUserAlreadyDeleted):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrUserEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrInvalidUserName),
			errors.Is(err, domain.ErrInvalidUserEmail),
			errors.Is(err, domain.ErrInvalidUserRole):

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

func (h *UserHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	response, err := h.service.FindByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
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

func (h *UserHandler) Search(c *gin.Context) {
	var request dto.UserSearchRequest

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

func (h *UserHandler) DeleteByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteByID(id); err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrUserAlreadyDeleted):
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

func (h *UserHandler) RestoreByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.RestoreByID(id); err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrUserNotDeleted):
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

func (h *UserHandler) ActivateByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.ActivateByID(id); err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrUserAlreadyDeleted):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrUserAlreadyActive):
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

func (h *UserHandler) DeactivateByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeactivateByID(id); err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrUserAlreadyDeleted):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrUserAlreadyInactive):
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

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
// @Description Create a new user (customer or admin). Requires authentication and the "admin" role.
// @Tags users
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var request dto.CreateUserRequest

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

// Update godoc
// @Summary Update user
// @Description Update an existing user, including role. Requires authentication and the "admin" role.
// @Tags users
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "User ID"
// @Param user body dto.UserUpdateRequest true "User data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users/{id} [put]
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

// FindByID godoc
// @Summary Find user by ID
// @Description Returns a user by its ID. Requires authentication and the "admin" role.
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/{id} [get]
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

// Search godoc
// @Summary List users
// @Description Lists users with filters and pagination. Requires authentication and the "admin" role.
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param name query string false "Filter by name (partial match)"
// @Param email query string false "Filter by email (partial match)"
// @Param role query string false "Filter by role" Enums(customer, admin)
// @Param active query bool false "Filter by active status"
// @Param deletionState query string false "Filter by deletion state" Enums(not_deleted, deleted, all)
// @Param page query int false "Page number, starting at 1 (default 1)"
// @Param pageSize query int false "Users per page (default 20, max 100)"
// @Success 200 {object} dto.UserPageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/users [get]
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

// DeleteByID godoc
// @Summary Delete user
// @Description Soft-deletes a user. Requires authentication and the "admin" role.
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users/{id} [delete]
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

// RestoreByID godoc
// @Summary Restore user
// @Description Restores a soft-deleted user. Requires authentication and the "admin" role.
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users/{id}/restore [patch]
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

// ActivateByID godoc
// @Summary Activate user
// @Description Activates a user. Requires authentication and the "admin" role.
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users/{id}/activate [patch]
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

// DeactivateByID godoc
// @Summary Deactivate user
// @Description Deactivates a user. Requires authentication and the "admin" role.
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users/{id}/deactivate [patch]
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

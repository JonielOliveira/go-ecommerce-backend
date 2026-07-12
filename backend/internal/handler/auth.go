package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/middleware"
	"ecommerce/internal/service"
)

type CookieConfig struct {
	Name     string
	Secure   bool
	SameSite http.SameSite
	Domain   string
}

func ParseSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return http.SameSiteStrictMode

	case "none":
		return http.SameSiteNoneMode

	default:
		return http.SameSiteLaxMode
	}
}

type AuthHandler struct {
	authService  service.AuthService
	userService  *service.UserService
	cookieConfig CookieConfig
}

func NewAuthHandler(authService service.AuthService, userService *service.UserService, cookieConfig CookieConfig) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		userService:  userService,
		cookieConfig: cookieConfig,
	}
}

func (h *AuthHandler) setAuthCookie(c *gin.Context, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieConfig.Name,
		Value:    token,
		Path:     "/",
		Domain:   h.cookieConfig.Domain,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cookieConfig.Secure,
		SameSite: h.cookieConfig.SameSite,
	})
}

func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieConfig.Name,
		Value:    "",
		Path:     "/",
		Domain:   h.cookieConfig.Domain,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.cookieConfig.Secure,
		SameSite: h.cookieConfig.SameSite,
	})
}

// Register godoc
// @Summary Register
// @Description Publicly register a new account. Always creates the "customer" role — clients cannot choose a different role.
// @Tags auth
// @Accept json
// @Produce json
// @Param registration body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var request dto.RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	response, err := h.userService.Register(request)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, domain.ErrInvalidUserName),
			errors.Is(err, domain.ErrInvalidUserEmail):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary Login
// @Description Authenticate with email and password, setting an HttpOnly access token cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user, token, expiresAt, err := h.authService.Login(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid email or password",
			})

		case errors.Is(err, domain.ErrUserInactive):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "user account is inactive",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
		}

		return
	}

	h.setAuthCookie(c, token, expiresAt)

	c.JSON(http.StatusOK, dto.LoginResponse{
		User: dto.AuthUserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  string(user.Role),
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearAuthCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AuthUserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  string(user.Role),
	})
}

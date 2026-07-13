package handler

import (
	"ecommerce/internal/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	config *config.Config
}

func NewHealthHandler(config *config.Config) *HealthHandler {
	return &HealthHandler{
		config: config,
	}
}

// Health godoc
// @Summary Health check
// @Description Returns application status, name, version and current timestamp. Public route.
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "UP",
		"application": h.config.AppName,
		"version":     h.config.AppVersion,
		"timestamp":   time.Now().UTC(),
	})
}

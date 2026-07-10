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

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "UP",
		"application": h.config.AppName,
		"version":     h.config.AppVersion,
		"timestamp":   time.Now().UTC(),
	})
}

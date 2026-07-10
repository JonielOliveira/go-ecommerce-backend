package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins: true,
		// AllowOrigins: []string{
		// 	"http://localhost:5173", // Vite/Vue/React
		// 	"http://localhost:3000", // React/Next
		// },
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
			"QUERY",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

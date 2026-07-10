package main

import (
	"log"

	"ecommerce/internal/config"
	"ecommerce/internal/database"
	"ecommerce/internal/handler"
	"ecommerce/internal/middleware"
	"ecommerce/internal/repository"
	"ecommerce/internal/routes"
	"ecommerce/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	productRepository := repository.NewPostgresProductRepository(db)
	productService := service.NewProductService(productRepository)
	productHandler := handler.NewProductHandler(productService)

	healthHandler := handler.NewHealthHandler(cfg)

	router := gin.Default()
	router.Use(middleware.CORS())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.Register(router, routes.Handlers{
		Product: productHandler,
		Health:  healthHandler,
	})

	router.Run(":" + cfg.ServerPort)
}

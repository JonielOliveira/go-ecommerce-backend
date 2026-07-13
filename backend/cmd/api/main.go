package main

import (
	"log"

	"ecommerce/internal/config"
	"ecommerce/internal/database"
	"ecommerce/internal/domain"
	"ecommerce/internal/handler"
	"ecommerce/internal/middleware"
	"ecommerce/internal/repository"
	"ecommerce/internal/routes"
	"ecommerce/internal/security"
	"ecommerce/internal/service"

	_ "ecommerce/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title E-commerce API
// @version 1.0
// @description API for the E-commerce backend service.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
// @description JWT access token set as an HttpOnly cookie by POST /api/v1/auth/login.
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

	userRepository := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	seedDefaultAdmin(userService)

	jwtService := security.NewJWTService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.JWTAccessTokenTTL)

	authRepository := repository.NewPostgresAuthRepository(db)
	authService := service.NewAuthService(authRepository, jwtService)
	authHandler := handler.NewAuthHandler(authService, userService, handler.CookieConfig{
		Name:     cfg.AuthCookieName,
		Secure:   cfg.AuthCookieSecure,
		SameSite: handler.ParseSameSite(cfg.AuthCookieSameSite),
		Domain:   cfg.AuthCookieDomain,
	})

	orderRepository := repository.NewPostgresOrderRepository(db)
	orderService := service.NewOrderService(orderRepository)
	orderHandler := handler.NewOrderHandler(orderService)

	healthHandler := handler.NewHealthHandler(cfg)

	authenticateMiddleware := middleware.Authenticate(jwtService, authRepository, cfg.AuthCookieName)
	requireAdminMiddleware := middleware.RequireRole(domain.RoleAdmin)
	requireCustomerOrAdminMiddleware := middleware.RequireRole(domain.RoleCustomer, domain.RoleAdmin)

	router := gin.Default()
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.Register(router, routes.Handlers{
		Product: productHandler,
		User:    userHandler,
		Auth:    authHandler,
		Order:   orderHandler,
		Health:  healthHandler,
	}, routes.Middlewares{
		Authenticate:           authenticateMiddleware,
		RequireAdmin:           requireAdminMiddleware,
		RequireCustomerOrAdmin: requireCustomerOrAdminMiddleware,
	})

	router.Run(":" + cfg.ServerPort)
}

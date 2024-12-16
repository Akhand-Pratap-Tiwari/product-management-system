package main

import (
	"fmt"
	"log"
	"product-management-system/internal/config"
	"product-management-system/internal/handlers"
	"product-management-system/internal/repository"
	"product-management-system/internal/service"
	"product-management-system/internal/cache"
	"product-management-system/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	appLogger := logger.NewLogger()

	// Database connection
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, 
		cfg.Database.User, cfg.Database.Password, 
		cfg.Database.DBName)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}

	// Run database migrations
	if err := db.AutoMigrate(&models.User{}, &models.Product{}); err != nil {
		appLogger.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize Redis Cache
	redisCache := cache.NewRedisCache(cfg.Redis.Host, cfg.Redis.Port)

	// Initialize Repositories
	productRepo := repository.NewProductRepository(db)

	// Initialize Message Queue
	messageQueue := queue.NewRabbitMQQueue(cfg.RabbitMQ.Host, cfg.RabbitMQ.Port)

	// Initialize Image Processor
	s3Client := // Initialize AWS S3 client
	imageProcessor := service.NewImageProcessor(s3Client, cfg.AWS.S3Bucket, appLogger)

	// Initialize Services
	productService := service.NewProductService(
		productRepo, 
		imageProcessor, 
		messageQueue, 
		appLogger,
	)

	// Initialize Handlers
	productHandler := handlers.NewProductHandler(
		productService, 
		redisCache, 
		appLogger,
	)

	// Setup Gin Router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Product Routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/products", productHandler.CreateProduct)
		v1.GET("/products/:id", productHandler.GetProductByID)
		v1.GET("/products", productHandler.ListProducts)
	}

	// Start the server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	appLogger.Info("Starting server", "address", serverAddr)
	
	if err := router.Run(serverAddr); err != nil {
		appLogger.Fatal("Server failed to start", "error", err)
	}
}
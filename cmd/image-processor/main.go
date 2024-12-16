package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"product-management-system/internal/config"
	"product-management-system/internal/models"
	"product-management-system/internal/repository"
	"product-management-system/internal/service"
	"product-management-system/pkg/logger"
	"product-management-system/pkg/utils"

	"github.com/streadway/amqp"
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

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)

	// Initialize S3 client
	s3Client := // Initialize AWS S3 client

	// Initialize image processor
	imageProcessor := service.NewImageProcessor(
		s3Client, 
		cfg.AWS.S3Bucket, 
		appLogger,
	)

	// Setup RabbitMQ connection
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%d", cfg.RabbitMQ.Host, cfg.RabbitMQ.Port))
	if err != nil {
		appLogger.Fatal("Failed to connect to RabbitMQ", "error", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		appLogger.Fatal("Failed to open a channel", "error", err)
	}
	defer ch.Close()

	// Declare queue
	q, err := ch.QueueDeclare(
		"image_processing_queue", // name
		true,                     // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		appLogger.Fatal("Failed to declare a queue", "error", err)
	}

	// Consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		appLogger.Fatal("Failed to register a consumer", "error", err)
	}

	// Process incoming messages
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			// Parse message to ImageProcessingTask
			task := &models.ImageProcessingTask{}
			if err := json.Unmarshal(d.Body, task); err != nil {
				appLogger.Error("Failed to parse message", "error", err)
				d.Nack(false, false) // Negative acknowledge
				continue
			}

			// Process images
			if err := imageProcessor.ProcessImages(task); err != nil {
				appLogger.Error("Image processing failed", "error", err, "productID", task.ProductID)
				d.Nack(false, true) // Requeue
				continue
			}

			// Update product with processed images
			if err := productRepo.UpdateProductImages(
				context.Background(), 
				task.ProductID, 
				task.CompressedImageURLs,
			); err != nil {
				appLogger.Error("Failed to update product images", "error", err)
				d.Nack(false, true) // Requeue
				continue
			}

			// Acknowledge message
			d.Ack(false)
		}
	}()

	appLogger.Info("Image Processing Service started. Waiting for messages...")
	<-forever
}
package service

import (
	"context"
	"errors"
	"product-management-system/internal/models"
	"product-management-system/internal/queue"
	"product-management-system/internal/repository"
	"product-management-system/pkg/logger"

	"gorm.io/gorm"
)

type ProductService struct {
	productRepo    *repository.ProductRepository
	imageProcessor *ImageProcessor
	messageQueue   *queue.RabbitMQQueue
	logger         *logger.Logger
}

func NewProductService(
	productRepo *repository.ProductRepository,
	imageProcessor *ImageProcessor,
	messageQueue *queue.RabbitMQQueue,
	logger *logger.Logger,
) *ProductService {
	return &ProductService{
		productRepo:    productRepo,
		imageProcessor: imageProcessor,
		messageQueue:   messageQueue,
		logger:         logger,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	// Validate product
	if product.ProductName == "" {
		return errors.New("product name is required")
	}

	// Save product
	if err := s.productRepo.Create(ctx, product); err != nil {
		s.logger.Error("Failed to create product", "error", err)
		return err
	}

	// Enqueue image processing task
	if len(product.ProductImages) > 0 {
		task := &models.ImageProcessingTask{
			ProductID: product.ID,
			ImageURLs: product.ProductImages,
		}
		if err := s.messageQueue.EnqueueImageProcessing(task); err != nil {
			s.logger.Error("Failed to enqueue image processing", "error", err)
			return err
		}
	}

	return nil
}

func (s *ProductService) FindProductByID(ctx context.Context, id uint) (*models.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		s.logger.Error("Failed to find product", "error", err)
		return nil, err
	}
	return product, nil
}

func (s *ProductService) ListProductsByUser(ctx context.Context, userID uint, filters map[string]interface{}) ([]models.Product, error) {
	products, err := s.productRepo.FindByUserID(ctx, userID, filters)
	if err != nil {
		s.logger.Error("Failed to list products by user", "error", err)
		return nil, err
	}
	return products, nil
}

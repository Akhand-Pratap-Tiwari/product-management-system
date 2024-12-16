package repository

import (
	"context"
	"errors"
	"product-management-system/internal/models"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	result := r.db.WithContext(ctx).First(&product, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, result.Error
	}
	return &product, nil
}

func (r *ProductRepository) FindByUserID(ctx context.Context, userID uint, filters map[string]interface{}) ([]models.Product, error) {
	var products []models.Product
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if minPrice, ok := filters["min_price"].(float64); ok {
		query = query.Where("product_price >= ?", minPrice)
	}

	if maxPrice, ok := filters["max_price"].(float64); ok {
		query = query.Where("product_price <= ?", maxPrice)
	}

	if productName, ok := filters["product_name"].(string); ok {
		query = query.Where("product_name ILIKE ?", "%"+productName+"%")
	}

	result := query.Find(&products)
	return products, result.Error
}

func (r *ProductRepository) UpdateProductImages(ctx context.Context, productID uint, compressedImages []string) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ?", productID).
		Update("compressed_product_images", compressedImages).Error
}

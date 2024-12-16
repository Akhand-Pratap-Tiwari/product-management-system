package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	UserID                  uint   `gorm:"not null"`
	ProductName             string `gorm:"not null"`
	ProductDescription      string
	ProductImages           []string `gorm:"type:text[]"`
	CompressedProductImages []string `gorm:"type:text[]"`
	ProductPrice            float64  `gorm:"type:decimal(10,2)"`
	ProcessedAt             time.Time
}

type ImageProcessingTask struct {
	ProductID           uint
	ImageURLs           []string
	ProcessedAt         time.Time
	Status              string
	ErrorMessage        string
	CompressedImageURLs []string
}

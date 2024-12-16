package service

import (
	"product-management-system/internal/models"
	"product-management-system/pkg/logger"

	"github.com/aws/aws-sdk-go/service/s3"
)

type ImageProcessor struct {
	s3Client *s3.S3
	bucket   string
	logger   *logger.Logger
}

func (ip *ImageProcessor) ProcessImages(task *models.ImageProcessingTask) error {
	var compressedImages []string

	for _, imageURL := range task.ImageURLs {
		compressedImage, err := ip.compressAndUploadImage(imageURL)
		if err != nil {
			ip.logger.Error("Image processing failed", "url", imageURL, "error", err)
			task.Status = "failed"
			task.ErrorMessage = err.Error()
			return err
		}
		compressedImages = append(compressedImages, compressedImage)
	}

	// Update product with compressed image URLs
	// This would typically be done through a repository method
	return nil
}

func (ip *ImageProcessor) compressAndUploadImage(imageURL string) (string, error) {
	// Implement image download, compression, and S3 upload
	// Return S3 URL of compressed image
	return "", nil
}

func NewImageProcessor(s3Client *s3.S3, s3Bucket string, appLogger *logger.Logger) *ImageProcessor {

	return &ImageProcessor{

		s3Client: s3Client,

		bucket: s3Bucket,

		logger: appLogger,
	}

}

func (ip *ImageProcessor) ProcessImage(imagePath string) error {

	// Dummy implementation for image processing

	ip.logger.Info("Processing image", "path", imagePath)

	return nil

}

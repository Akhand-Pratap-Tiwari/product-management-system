package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"path/filepath"
	"product-management-system/internal/models"
	"product-management-system/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/nfnt/resize"
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
	// Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Decode the image
	var img image.Image
	var format string
	switch resp.Header.Get("Content-Type") {
	case "image/jpeg":
		img, err = jpeg.Decode(resp.Body)
		format = "jpeg"
	case "image/png":
		img, err = png.Decode(resp.Body)
		format = "png"
	default:
		return "", fmt.Errorf("unsupported image format: %s", resp.Header.Get("Content-Type"))
	}
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize the image
	resizedImg := resize.Resize(800, 0, img, resize.Lanczos3)

	// Compress the image
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 75})
	case "png":
		err = png.Encode(&buf, resizedImg)
	}
	if err != nil {
		return "", fmt.Errorf("failed to compress image: %w", err)
	}

	// Upload to S3
	uploader := s3manager.NewUploaderWithClient(ip.s3Client)
	key := fmt.Sprintf("compressed/%s", filepath.Base(imageURL))
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(ip.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image to S3: %w", err)
	}

	// Return the S3 URL of the compressed image
	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", ip.bucket, key)
	return s3URL, nil
}

func NewImageProcessor(s3Client *s3.S3, s3Bucket string, appLogger *logger.Logger) *ImageProcessor {

	return &ImageProcessor{

		s3Client: s3Client,

		bucket: s3Bucket,

		logger: appLogger,
	}

}

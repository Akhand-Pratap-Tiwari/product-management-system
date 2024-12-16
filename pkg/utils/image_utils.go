package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"

	"github.com/nfnt/resize"
)

// ImageProcessor provides utility functions for image processing
type ImageProcessor struct {
	MaxWidth  uint
	MaxHeight uint
	Quality   int
}

// NewImageProcessor creates a new ImageProcessor with default settings
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		MaxWidth:  800,
		MaxHeight: 600,
		Quality:   75,
	}
}

// DownloadImage downloads an image from a given URL
func (ip *ImageProcessor) DownloadImage(url string) (image.Image, string, error) {
	// Download the image
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// Determine image format
	contentType := resp.Header.Get("Content-Type")

	// Decode the image
	var img image.Image
	switch contentType {
	case "image/jpeg":
		img, err = jpeg.Decode(resp.Body)
	case "image/png":
		img, err = png.Decode(resp.Body)
	default:
		return nil, "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	if err != nil {
		return nil, "", err
	}

	return img, contentType, nil
}

// ResizeImage resizes an image maintaining aspect ratio
func (ip *ImageProcessor) ResizeImage(img image.Image) image.Image {
	// Calculate new dimensions
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// Determine scaling
	var newWidth, newHeight uint
	if width > height {
		newWidth = ip.MaxWidth
		ratio := float64(height) / float64(width)
		newHeight = uint(float64(newWidth) * ratio)
	} else {
		newHeight = ip.MaxHeight
		ratio := float64(width) / float64(height)
		newWidth = uint(float64(newHeight) * ratio)
	}

	// Resize the image
	return resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
}

// CompressImage compresses the image to a byte slice
func (ip *ImageProcessor) CompressImage(img image.Image, format string) ([]byte, error) {
	// Create a buffer to store the compressed image
	buf := new(bytes.Buffer)

	// Compress based on format
	switch format {
	case "image/jpeg":
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: ip.Quality})
		if err != nil {
			return nil, err
		}
	case "image/png":
		err := png.Encode(buf, img)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	return buf.Bytes(), nil
}

// ProcessImage combines download, resize, and compression
func (ip *ImageProcessor) ProcessImage(imageURL string) ([]byte, error) {
	// Download the image
	img, format, err := ip.DownloadImage(imageURL)
	if err != nil {
		return nil, err
	}

	// Resize the image
	resizedImg := ip.ResizeImage(img)

	// Compress the image
	compressedImage, err := ip.CompressImage(resizedImg, format)
	if err != nil {
		return nil, err
	}

	return compressedImage, nil
}

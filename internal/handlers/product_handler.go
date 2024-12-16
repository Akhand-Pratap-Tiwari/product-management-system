package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"product-management-system/internal/cache"
	"product-management-system/internal/models"
	"product-management-system/internal/service"
	"product-management-system/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
	redisCache     *cache.RedisCache
	logger         *logger.Logger
}

func NewProductHandler(
	productService *service.ProductService,
	redisCache *cache.RedisCache,
	logger *logger.Logger,
) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		redisCache:     redisCache,
		logger:         logger,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		h.logger.Error("Invalid product data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.productService.CreateProduct(c.Request.Context(), &product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Try cache first
	cacheKey := fmt.Sprintf("product:%d", productID)
	cachedProduct, err := h.redisCache.Get(c.Request.Context(), cacheKey)
	if err == nil {
		c.JSON(http.StatusOK, cachedProduct)
		return
	}

	// Fetch from database
	product, err := h.productService.FindProductByID(c.Request.Context(), uint(productID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Cache the result
	if err := h.redisCache.Set(c.Request.Context(), cacheKey, product, time.Hour); err != nil {
		h.logger.Warn("Failed to cache product", "error", err)
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	filters := map[string]interface{}{
		"min_price":    c.Query("min_price"),
		"max_price":    c.Query("max_price"),
		"product_name": c.Query("product_name"),
	}

	products, err := h.productService.ListProductsByUser(c.Request.Context(), uint(userID), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

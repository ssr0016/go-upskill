package controllers

import (
	"ambassador/src/database"
	"ambassador/src/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateProductRequest struct {
    Title        string  `json:"title" validate:"required,min=1,max=255"`
    Description string  `json:"description" validate:"required,min=10,max=1000"`
    Image       string  `json:"image" validate:"required,url"`
    Price       float64 `json:"price" validate:"required,min=0,max=100000"`
}

type ProductResponse struct {
    ID          uint    `json:"id"`
    Title        string `json:"title"`
    Description string  `json:"description"`
    Image       string  `json:"image"`
    Price       float64 `json:"price"`
}

type ProductListResponse struct {
    ID          uint    `json:"id"`
    Title       string  `json:"title"`
    Description string  `json:"description"`
    Image       string  `json:"image"`
    Price       float64 `json:"price"`
}

func Products(c *fiber.Ctx) error {
	var products  []ProductListResponse

	// Fetch ALL products with ONLY needed fields
	if err := database.DB.
        Model(&models.Product{}).
        Select("id, title, description, image, price").
        Find(&products).Error; err != nil {
        
        log.Printf("Failed to fetch products: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to fetch products",
        })
    }

 	return c.JSON(products) 
}

func CreateProducts(c *fiber.Ctx) error {
	var data CreateProductRequest

	 if err := c.BodyParser(&data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

	// Normalize input
    data.Title = strings.TrimSpace(data.Title)
    data.Description = strings.TrimSpace(data.Description)
    data.Image = strings.TrimSpace(data.Image)

	 // Validate (Register pattern)
    if data.Title == "" || data.Description == "" || data.Image == "" || data.Price < 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "title, description, image, and price are required",
        })
    }

	if data.Price > 100000 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "price too high (max $100,000)",
        })
    }

	product := models.Product{
        Title:        data.Title,
        Description: data.Description,
        Image:       data.Image,
        Price:       data.Price,
    }

    if err := database.DB.Create(&product).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create"})
    }

    // MANUAL MAPPING - Exact control
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "message": "product created successfully",
        "data": ProductResponse{
            ID:          product.ID,
            Title:       product.Title,
            Description: product.Description,
            Image:       product.Image,
            Price:       product.Price,
        },
    })
}



func GetProduct(c *fiber.Ctx) error {
	// Parse & validate ID (same pattern as Register)
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		 return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid product ID",
        })
	}

	if id <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "product ID must be positive",
        })
    }

	// Fetch product with error handling (same as Products)
	var product ProductResponse
    if err := database.DB.
        Model(&models.Product{}).
        Where("id = ?", id).
        Select("id, title, description, image, price").  
        First(&product).Error; err != nil {
        
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "product not found",
            })
        }
        
        log.Printf("Failed to fetch product %d: %v", id, err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to fetch product",
        })
    }

	 return c.JSON(product)
}

type UpdateProductRequest struct {
    Title       string  `json:"title" validate:"omitempty,min=1,max=255"`
    Description string  `json:"description" validate:"omitempty,min=10,max=1000"`
    Image       string  `json:"image" validate:"omitempty,url"`
    Price       float64 `json:"price" validate:"omitempty,min=0,max=100000"`
}

func UpdateProduct(c *fiber.Ctx) error {
	 // Parse & validate ID (same as GetProduct)
	idStr := c.Params("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid product ID",
        })
    }

	if id <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "product ID must be positive",
        })
    }

	// Parse request body (same as CreateProduct)
    var data UpdateProductRequest
    if err := c.BodyParser(&data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

	// Normalize input (same pattern)
    data.Title = strings.TrimSpace(data.Title)
    data.Description = strings.TrimSpace(data.Description)
    data.Image = strings.TrimSpace(data.Image)


	// Validate at least one field provided
    if data.Title == "" && data.Description == "" && data.Image == "" && data.Price == 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "at least one field must be provided",
        })
    }

	// Fetch existing product first
    var existingProduct models.Product
    if err := database.DB.First(&existingProduct, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "product not found",
            })
        }
        log.Printf("Failed to fetch product %d: %v", id, err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to fetch product",
        })
    }

	// Update only provided fields (partial update)
    updates := map[string]interface{}{}

    if data.Title != "" {
        existingProduct.Title = data.Title  // or Name if your model uses Name
        updates["title"] = data.Title
    }

    if data.Description != "" {
        existingProduct.Description = data.Description
        updates["description"] = data.Description
    }

    if data.Image != "" {
        existingProduct.Image = data.Image
        updates["image"] = data.Image
    }

    if data.Price > 0 {
        existingProduct.Price = data.Price
        updates["price"] = data.Price
    }

	// Save to database
    if err := database.DB.Model(&existingProduct).Updates(updates).Error; err != nil {
        log.Printf("Failed to update product %d: %v", id, err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to update product",
        })
    }

	// Return shaped response (same as GetProduct)
    response := ProductResponse{
        ID:          existingProduct.ID,
        Title:       existingProduct.Title,
        Description: existingProduct.Description,
        Image:       existingProduct.Image,
        Price:       existingProduct.Price,
    }

    return c.JSON(fiber.Map{
        "message": "product updated successfully",
        "data":    response,
    })
}

func DeleteProduct(c *fiber.Ctx) error {
	 // Parse & validate ID (same as GetProduct)
	idStr := c.Params("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid product ID",
        })
    }

	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID must be positive",
		})
	}

	// Fetch existing product first
	var existingProduct models.Product
	if err := database.DB.First(&existingProduct, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		}
		log.Printf("Failed to fetch product %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch product",
		})
	}

	// Delete from database (perfect) - Soft delete with gorm.Model
    result := database.DB.Delete(&existingProduct)
    if result.Error != nil {
        log.Printf("Failed to delete product %d: %v", id, result.Error)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to delete product",
        })
    }

	// Verify delete worked
    if result.RowsAffected == 0 {
        log.Printf("No rows affected when deleting product %d", id)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to delete product",
        })
    }

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
    	"message": "product deleted successfully",
	})
}

func ProductFrontEnd(c *fiber.Ctx) error {
    ctx := c.Context()
    cacheKey := "products_frontend"

    // 1. CHECK CACHE FIRST (8ms)
    if cached, err := database.CacheGet(ctx,cacheKey); err == nil {
       log.Printf("✅ Cache HIT for %s", cacheKey)

         // Parse cached JSON back to products
         var products []ProductListResponse
          if jsonErr := json.Unmarshal([]byte(cached), &products); jsonErr != nil {
            log.Printf("Cache parse error: %v", jsonErr)
            // Continue to DB on parse error
        }else {
             return c.JSON(fiber.Map{
                "data": products,
                "source": "cache", // Debug info
                "cached": true,
            })
        }
    }

    // 2. CACHE MISS → DB query (200ms)
    log.Printf("Cache MISS -> Querying database for %s", cacheKey)
    var products []ProductListResponse
    
    if err := database.DB.
        Model(&models.Product{}).
        Select("id, title, description, image, price").
        Find(&products).Error; err != nil {
        
        log.Printf("Failed to fetch products: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to fetch products",
        })
    }
    
    // 3. CACHE RESULT (JSON → Redis, 10min TTL)
    if jsonData, err := json.Marshal(products); err == nil {
        ttl := 10 * time.Minute
        if err := database.CacheSet(ctx, cacheKey, jsonData, ttl); err == nil {
            log.Printf("Cached %d products for %v", len(products), ttl)
        }
    }
    
    return c.JSON(fiber.Map{
        "data": products,
        "source": "database",
        "cached": false,
    })
}

type ProductBackendResponse struct {
    Cached  bool                   `json:"cached"`
    Data    []ProductListResponse  `json:"data"`
    Query   string                 `json:"query"`
    Source  string                 `json:"source"`
    Count   int                    `json:"count"`
}

func ProductBackend(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // DYNAMIC CACHE KEY based on search
    searchQuery := strings.TrimSpace(c.Query("s", ""))
    normalizedQuery := strings.ToLower(searchQuery)
    cacheKey := fmt.Sprintf("products_backend:s:%s", normalizedQuery)
    
    // 1. CHECK CACHE FIRST
    if cached, err := database.CacheGet(ctx, cacheKey); err == nil {
        log.Printf("Cache HIT for %s", cacheKey)
        var products []ProductListResponse
        if jsonErr := json.Unmarshal([]byte(cached), &products); jsonErr == nil {
            return c.JSON(ProductBackendResponse{
                Cached:  true,
                Data:    products,
                Query:   searchQuery,
                Source:  "cache",
                Count:   len(products),
            })
        }
    }
    
    // 2. CACHE MISS → SMART DB QUERY
    log.Printf("Cache MISS → Searching '%s'", searchQuery)
    var products []ProductListResponse
    
    db := database.DB.Model(&models.Product{}).Select("id, title, description, image, price")
    
    if searchQuery != "" {
        searchTerm := "%" + normalizedQuery + "%"
        db = db.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", 
                     searchTerm, searchTerm)
    }
    
    if err := db.Find(&products).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to fetch products",
        })
    }
    
    // 3. CACHE RESULTS
    if jsonData, err := json.Marshal(products); err == nil {
        ttl := 10 * time.Minute
        database.CacheSet(ctx, cacheKey, jsonData, ttl)
        log.Printf("Cached %d products for '%s'", len(products), searchQuery)
    }
    
    return c.JSON(ProductBackendResponse{
        Cached:  false,
        Data:    products,
        Query:   searchQuery,
        Source:  "database",
        Count:   len(products),
    })
}

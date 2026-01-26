package controllers

import (
	"ambassador/src/database"
	"ambassador/src/middlewares"
	"ambassador/src/models"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LinkResponse struct {
    Links []models.Link `json:"links"`
    Count int64         `json:"count"`
    Page  int           `json:"page"`
    Limit int           `json:"limit"`
}

// GetUserLinks returns all links for a specific user
func Link(c *fiber.Ctx) error {
	// Parse user ID with validation
    userIDStr := c.Params("id")
    if userIDStr == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error":   true,
            "message": "User ID is required",
        })
    }
	
	userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error":   true,
            "message": "Invalid user ID format",
        })
    }

	 // Fetch links with associations preloaded
	 var links []models.Link
	 if err := database.DB.
        WithContext(c.Context()).
        Preload("User").
        Preload("Products").
        Where("user_id = ?", userID).
        Find(&links).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error":   true,
            "message": "Failed to fetch user links",
        })
    }

    for i, link := range links {
        var orders models.Order
        database.DB.Where("code = ? and complete = true", link.Code).First(&orders)

        links[i].Orders = []models.Order{orders}
    }

	return c.JSON(fiber.Map{
        "links": links,
    })
}	

type CreateLinkRequest struct {
    Products []int `json:"products"`
}

func CreateLink(c *fiber.Ctx) error {
    var request CreateLinkRequest

    if err := c.BodyParser(&request); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    if len(request.Products) == 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "At least one product required",
        })
    }

    userID, err := middlewares.GetUserID(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized",
        })
    }

    // Create link
    link := models.Link{
        UserID: userID,
        Code:   generateLinkCode(),
    }

    if err := database.DB.Create(&link).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to create link",
        })
    }

    // Associate products
    if err := associateProducts(database.DB, &link, request.Products); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    // Return preloaded result
    var result models.Link
    if err := database.DB.
        Preload("User").
        Preload("Products").
        First(&result, link.ID).Error; err != nil {
        
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch link",
        })
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "link": result,
    })
}


type LinkStat struct {
    Code    string  `json:"code"`
    Count   int     `json:"count"`
    Revenue float64 `json:"revenue"`
}

func Stats(c *fiber.Ctx) error {
    userID, err := middlewares.GetUserID(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Unauthorized",
        })
    }

    // Get user's links
    var links []models.Link
    if err := database.DB.
        WithContext(c.Context()).
        Where("user_id = ?", userID).
        Find(&links).Error; err != nil {
        
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch links",
        })
    }

    var result []LinkStat

    // Calculate stats per link
    for _, link := range links {
        // Get completed orders for this link
        var orders []models.Order
        if err := database.DB.
            WithContext(c.Context()).
            Preload("OrderItems").
            Where("code = ? AND complete = ?", link.Code, true).
            Find(&orders).Error; err != nil {
            
            continue // Skip this link if error
        }

        // Calculate revenue
        revenue := 0.0
        for _, order := range orders {
            revenue += order.GetTotal()
        }

        result = append(result, LinkStat{
            Code:    link.Code,
            Count:   len(orders),
            Revenue: revenue,
        })
    }

    return c.JSON(result)
}


func generateLinkCode() string {
    b := make([]byte, 7)
    charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}

func associateProducts(db *gorm.DB, link *models.Link, productIDs []int) error {
    if len(productIDs) == 0 {
        return nil
    }

    // Verify products exist
    var products []models.Product
    if err := db.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
        return fmt.Errorf("database error: %w", err)
    }

    if len(products) != len(productIDs) {
        return fmt.Errorf("products not found: expected %d, got %d", 
            len(productIDs), len(products))
    }

    // Save associations
    return db.Model(link).Association("Products").Append(products)
}
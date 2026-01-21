package controllers

import (
	"ambassador/src/database"
	"ambassador/src/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
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
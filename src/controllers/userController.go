package controllers

import (
	"ambassador/src/database"
	"ambassador/src/models"

	"github.com/gofiber/fiber/v2"
)

type AmbassadorResponse struct {
	ID           uint   `json:"id"`
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Email        string `json:"email"`
    IsAmbassador bool   `json:"is_ambassador"`
}

func Ambassadors(c *fiber.Ctx) error {
	var users []models.User
	if err := database.DB.
        Where("is_ambassador = ?", true).
        Find(&users).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to fetch ambassadors",
        })
    }

	 // Shape response - NO PASSWORDS!
	ambassadors := make([]AmbassadorResponse, len(users))
	for i, user := range users {
        ambassadors[i] = AmbassadorResponse{
            ID:           user.ID,
            FirstName:    user.FirstName,
            LastName:     user.LastName,
            Email:        user.Email,
            IsAmbassador: user.IsAmbassador,
        }
    }

	return c.JSON(fiber.Map{
        "data":      ambassadors,
        "count":     len(ambassadors),
        "is_success": true,
    })
}

type Ranking struct {
    ID      uint    `json:"id"`
    Name    string  `json:"name"`
    Email   string  `json:"email"`
    Revenue float64 `json:"revenue"`
}

type RankingResponse struct {
    Name   string  `json:"name"`
    Revenue float64 `json:"revenue"`
}

func Rankings(c *fiber.Ctx) error {
    var users []models.User
    if err := database.DB.
        Where("is_ambassador = ?", true).
        Find(&users).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch ambassadors",
        })
    }

    rankings := make(map[string]float64)
    for _, user := range users {
        revenue := user.CalculateRevenue(database.DB)
        rankings[user.Name()] = revenue
    }

    return c.JSON(fiber.Map{
        "data": rankings,
        "count": len(rankings),
        "is_success": true,
    })
}
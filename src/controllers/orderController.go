package controllers

import (
	"ambassador/src/database"
	"ambassador/src/middlewares"
	"ambassador/src/models"

	"github.com/gofiber/fiber/v2"
)

const (
	ScopeAdmin = "admin"
	ScopeAmbassador = "ambassador"
)

type OrderListResponse struct {
	Orders []models.Order `json:"orders"`
	Count int64 `json:"count"`
}

// Orders retrieves all orders with relationships preloaded
func Orders(c *fiber.Ctx) error {
    scope, _ := middlewares.GetScope(c)
    var orders []models.Order

    query := database.DB.Preload("OrderItems")

    switch scope {
    case ScopeAdmin:
        // Admin sees ALL orders
        query.Find(&orders)
        
    case ScopeAmbassador:
        // Ambassador sees ONLY their orders
        user, _ := middlewares.GetUser(c)
        query = query.Where("ambassador_email = ? OR user_id = ?", user.Email, user.ID)
        query.Find(&orders)
        
    default:
        // Invalid scope - security!
        return c.Status(403).JSON(fiber.Map{"error": "Invalid scope"})
    }

    // Calculate presentation fields
    for i, order := range orders {
        orders[i].Name = order.FullName()
        orders[i].Total = order.GetTotal()
    }

    return c.JSON(orders)
}


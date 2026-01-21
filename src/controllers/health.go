package controllers

import (
	"ambassador/src/database"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Check database
    if err := database.HealthCheck(ctx); err != nil {
        return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
            "status":   "unhealthy",
            "database": "disconnected",
            "error":    err.Error(),
        })
    }

    // Get database stats
    stats := database.Stats()

    return c.JSON(fiber.Map{
        "status":   "healthy",
        "database": "connected",
        "time":     time.Now().UTC(),
        "connections": fiber.Map{
            "open":   stats.OpenConnections,
            "in_use": stats.InUse,
            "idle":   stats.Idle,
        },
    })
}
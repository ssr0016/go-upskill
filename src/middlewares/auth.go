package middlewares

import (
	"ambassador/src/config"
	"ambassador/src/database"
	"ambassador/src/models"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func IsAuthenticated(c *fiber.Ctx) error {
	 // Get token from cookie
    cookie := c.Cookies("jwt")

	 // If not in cookie, try Authorization header
    if cookie == "" {
        auth := c.Get("Authorization")
        if auth != "" && strings.HasPrefix(auth, "Bearer ") {
            cookie = strings.TrimPrefix(auth, "Bearer ")
        }
    }

	if cookie == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "unauthenticated",
        })
    }

	// Parse token
    cfg := config.Get()
    token, err := jwt.ParseWithClaims(cookie, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(cfg.JWTSecret), nil
    })

	if err != nil || !token.Valid {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "unauthenticated",
        })
    }

	// Extract claims
    claims, ok := token.Claims.(*jwt.RegisteredClaims)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "invalid token claims",
        })
    }

	// Store claims in context for later use
    c.Locals("user_id", claims.Subject)
    c.Locals("claims", claims)

    return c.Next()
}

// GetUserID retrieves user ID from context
func GetUserID(c *fiber.Ctx) (uint, error) {
    userIDStr, ok := c.Locals("user_id").(string)
    if !ok {
        return 0, fiber.NewError(fiber.StatusUnauthorized, "user ID not found in context")
    }

    userID, err := strconv.ParseUint(userIDStr, 10, 32)
    if err != nil {
        return 0, fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
    }

    return uint(userID), nil
}

// GetUser retrieves the full user from database using context
func GetUser(c *fiber.Ctx) (*models.User, error) {
    userID, err := GetUserID(c)
    if err != nil {
        return nil, err
    }

    var user models.User
    if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
        return nil, fiber.NewError(fiber.StatusUnauthorized, "user not found")
    }

    return &user, nil
}
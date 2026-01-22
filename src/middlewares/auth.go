package middlewares

import (
	"ambassador/src/config"
	"ambassador/src/database"
	"ambassador/src/models"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// Custom claims with scope
type ClaimsWithScope struct {
    jwt.RegisteredClaims
    Scope string `json:"scope"`
}

func IsAuthenticated(c *fiber.Ctx) error {
    // Get token from cookie or Authorization header
    cookie := c.Cookies("jwt")
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

    cfg := config.Get()
    token, err := jwt.ParseWithClaims(cookie, &ClaimsWithScope{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(cfg.JWTSecret), nil
    })

    if err != nil || !token.Valid {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "error":   "UNAUTHORIZED",
            "message": "Invalid or expired token",
            "code":    "INVALID_TOKEN",
            "status":  401,
        })
    }

    claims, ok := token.Claims.(*ClaimsWithScope)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "error":   "UNAUTHORIZED",
            "message": "Invalid token claims",
            "code":    "INVALID_CLAIMS",
            "status":  401,
        })
    }

    // Store claims in context
    c.Locals("user_id", claims.Subject)
    c.Locals("scope", claims.Scope)
    c.Locals("claims", claims)

    return c.Next()
}

// Scope protection middleware
func RequireScope(requiredScope string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        scope, ok := c.Locals("scope").(string)
        if !ok || scope != requiredScope {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "success": false,
                "error":   "FORBIDDEN",
                "message": fmt.Sprintf("%s scope required", requiredScope),
                "code":    "INSUFFICIENT_SCOPE",
                "status":  403,
            })
        }
        return c.Next()
    }
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

// GetScope retrieves scope from context
func GetScope(c *fiber.Ctx) (string, error) {
    scope, ok := c.Locals("scope").(string)
    if !ok {
        return "", fiber.NewError(fiber.StatusUnauthorized, "scope not found in context")
    }
    return scope, nil
}

func GenerateJWT(c *fiber.Ctx, id uint, scope string) error {
    cfg := config.Get()
    
    claims := ClaimsWithScope{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   strconv.Itoa(int(id)),
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.JWTExpireHours))),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
        Scope: scope,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signedToken, err := token.SignedString([]byte(cfg.JWTSecret))
    if err != nil {
        return fmt.Errorf("JWT signing failed: %w", err)
    }

    cookie := &fiber.Cookie{
        Name:     "jwt",
        Value:    signedToken,
        Expires:  time.Now().Add(time.Hour * time.Duration(cfg.JWTExpireHours)),
        HTTPOnly: true,
        Secure:   cfg.Environment == "production",
        SameSite: "Strict",
        Path:     "/",
    }
    
    c.Cookie(cookie)
    return nil
}
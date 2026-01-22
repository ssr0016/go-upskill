package controllers

import (
	"ambassador/src/database"
	"ambassador/src/middlewares"
	"ambassador/src/models"
	"ambassador/src/utils"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type RegisterRequest struct {
    FirstName       string `json:"first_name" validate:"required,min=1,max=100"`
    LastName        string `json:"last_name" validate:"required,min=1,max=100"`
    Email           string `json:"email" validate:"required,email,max=255"`
    Password        string `json:"password" validate:"required,min=8,max=72"`
    PasswordConfirm string `json:"password_confirm" validate:"required"`
}

type UserResponse struct {
    ID          uint   `json:"id"`
    FirstName   string `json:"first_name"`
    LastName    string `json:"last_name"`
    Email       string `json:"email"`
    IsAmbassador bool  `json:"is_ambassador"`
    Revenue     *float64 `json:"revenue,omitempty" gorm:"-"`
}

func Register(c *fiber.Ctx) error {
	// data from request
	var data RegisterRequest

   err := c.BodyParser(&data)
   if err != nil {
	 return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "invalid request body",
	})
   }

   	// normalize input
	data.FirstName = strings.TrimSpace(data.FirstName)
    data.LastName = strings.TrimSpace(data.LastName)
    data.Email = strings.ToLower(strings.TrimSpace(data.Email))

   // required fields
   if data.FirstName == "" || data.LastName == "" || data.Email == "" || data.Password == "" || data.PasswordConfirm == "" {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "all fields are required",
	})
   }

   // Validate input
   if err := validateRegistration(&data); err != nil {
	   return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error":   err.Error(),
	})
   }

	// Hash password
    hashedPassword, err := utils.HashPassword(data.Password)
    if err != nil {
        log.Printf("Password hashing failed: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Registration failed",
        })
    }

    // Determine if ambassador and exact path prefix match for /api/ambassador/register
    isAmbassador := strings.HasPrefix(c.Path(), "/api/ambassador")

	// Create user
    user := models.User{
        FirstName:    data.FirstName,
        LastName:     data.LastName,
        Email:        data.Email,
        Password:     []byte(hashedPassword), // Fix: convert to []byte
        IsAmbassador: isAmbassador, 
    }

	// INSERT FIRST — DB ENFORCES UNIQUENESS
    if err := database.DB.Create(&user).Error; err != nil {
        if strings.Contains(err.Error(), "Duplicate entry") || 
           strings.Contains(err.Error(), "UNIQUE constraint") {
            // Generic message to prevent user enumeration
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "Registration failed",
            })
        }
        log.Printf("User creation failed: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Registration failed",
        })
    }

	return c.Status(fiber.StatusCreated).JSON(UserResponse{
        ID:           user.ID,
        FirstName:    user.FirstName,
        LastName:     user.LastName,
        Email:        user.Email,
        IsAmbassador: user.IsAmbassador,
    })
}

func validateRegistration(data *RegisterRequest) error {
	 // Required fields
    if data.FirstName == "" || data.LastName == "" || 
       data.Email == "" || data.Password == "" || data.PasswordConfirm == "" {
        return fiber.NewError(fiber.StatusBadRequest, "All fields are required")
    }

	// Email format
    if !emailRegex.MatchString(data.Email) {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid email format")
    }

	// Password rules
    if len(data.Password) < 8 {
        return fiber.NewError(fiber.StatusBadRequest, "Password must be at least 8 characters")
    }
    if len(data.Password) > 72 {
        return fiber.NewError(fiber.StatusBadRequest, "Password too long")
    }

    // Password match
    if data.Password != data.PasswordConfirm {
        return fiber.NewError(fiber.StatusBadRequest, "Passwords do not match")
    }

    return nil
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

func Login(c *fiber.Ctx) error {
	var data LoginRequest

    // Parse request body
    if err := c.BodyParser(&data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

	// Normalize email only (NOT password)
    data.Email = strings.ToLower(strings.TrimSpace(data.Email))

	// Validate required fields
    if data.Email == "" || data.Password == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Email and password are required",
        })
    }

	// Find user by email
    var user models.User
    if err := database.DB.Where("email = ?", data.Email).First(&user).Error; err != nil {
        // FIXED: Single check, no duplicate logic
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid credentials",
        })
    }

    // Check password (timing-safe)
    if err := utils.CheckPassword(user.Password, data.Password); err != nil {
        log.Printf("Failed login attempt for: %s", data.Email)
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid credentials",
        })
    }

    // FIXED: Use user.IsAmbassador for scope (not path)
    scope := "admin"
    if user.IsAmbassador {
        scope = "ambassador"
    }

    // FIXED: Pass c to GenerateJWT
    if err := middlewares.GenerateJWT(c, user.ID, scope); err != nil {
        log.Printf("JWT generation failed: %v", err) // Add log import if missing
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to generate token",
        })
    }

	// Success response (cookie already set)
    response := UserResponse{
        ID:           user.ID,
        FirstName:    user.FirstName,
        LastName:     user.LastName,
        Email:        user.Email,
        IsAmbassador: user.IsAmbassador,
    }

	 return c.JSON(fiber.Map{
        "message": "Login successful",
        "user":    response,
        // Don't return token if using cookie-based auth
    })
}

// User returns the authenticated user
func User(c *fiber.Ctx) error {
    user, err := middlewares.GetUser(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "unauthenticated",
        })
    }

    // Calculate revenue
    user.Revenue = user.CalculateRevenue(database.DB)

    return c.JSON(UserResponse{
        ID:           user.ID,
        FirstName:    user.FirstName,
        LastName:     user.LastName,
        Email:        user.Email,
        IsAmbassador: user.IsAmbassador,
        Revenue:      &user.Revenue, 
    })
}


func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "successfully logged out",
	})
}

type UpdateUserRequest struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
}

func UpdateInfo(c *fiber.Ctx) error {
    // Get authenticated user from context
    user, err := middlewares.GetUser(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "unauthenticated",
        })
    }

     var data UpdateUserRequest

    if err := c.BodyParser(&data); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "invalid request body",
            })
    }

     // Normalize input
    data.FirstName = strings.TrimSpace(data.FirstName)
    data.LastName = strings.TrimSpace(data.LastName)
    data.Email = strings.ToLower(strings.TrimSpace(data.Email))

    // Validate at least one field is provided
    if data.FirstName == "" && data.LastName == "" && data.Email == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "at least one field is required",
        })
    }

    // Update only provided fields
    if data.FirstName != "" {
        user.FirstName = data.FirstName
    }
    if data.LastName != "" {
        user.LastName = data.LastName
    }
    if data.Email != "" {
        // Validate email format
        if !emailRegex.MatchString(data.Email) {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "invalid email format",
            })
        }
        user.Email = data.Email
    }

     // Save to database
    if err := database.DB.Save(&user).Error; err != nil {
        if strings.Contains(err.Error(), "Duplicate entry") || 
           strings.Contains(err.Error(), "UNIQUE constraint") {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "email already in use",
            })
        }
        log.Printf("User update failed: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to update user",
        })
    }

    return c.JSON(UserResponse{
        ID:           user.ID,
        FirstName:    user.FirstName,
        LastName:     user.LastName,
        Email:        user.Email,
        IsAmbassador: user.IsAmbassador,
    })
}

type UpdatePasswordRequest struct {
    CurrentPassword    string `json:"current_password" validate:"required,min=8"`
    NewPassword        string `json:"new_password" validate:"required,min=8,max=72"`
    NewPasswordConfirm string `json:"new_password_confirm" validate:"required"`
}

func UpdatePassword(c *fiber.Ctx) error {
    // Get authenticated user from context (same as UpdateInfo)
    user, err := middlewares.GetUser(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "unauthenticated",
        })
    }

    // Parse request body
    var data UpdatePasswordRequest
    if err := c.BodyParser(&data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    // Validate passwords match
    if data.NewPassword != data.NewPasswordConfirm {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "new passwords do not match",
        })
    }

    // Validate new password length
    if len(data.NewPassword) < 8 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "new password must be at least 8 characters",
        })
    }

    // Verify current password
    if err := utils.CheckPassword(user.Password, data.CurrentPassword); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "current password is incorrect",
        })
    }

     // Hash new password
    newHashedPassword, err := utils.HashPassword(data.NewPassword)
    if err != nil {
        log.Printf("Password hashing failed: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to update password",
        })
    }

    // Update password in database
    user.Password = []byte(newHashedPassword)
    if err := database.DB.Save(&user).Error; err != nil {
        log.Printf("Password update failed: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to update password",
        })
    }

    return c.JSON(fiber.Map{
        "message": "password updated successfully",
    })
}
package controllers

import (
	"ambassador/src/config"
	"ambassador/src/database"
	"ambassador/src/models"
	"ambassador/src/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type RegisterRequest struct {
	FirstName       string `json:"first_name"`
    LastName        string `json:"last_name"`
    Email           string `json:"email"`
    Password        string `json:"password"`
    PasswordConfirm string `json:"password_confirm"`
}

type UserResponse struct {
    ID          uint   `json:"id"`
    FirstName   string `json:"first_name"`
    LastName    string `json:"last_name"`
    Email       string `json:"email"`
    IsAmbassador bool  `json:"is_ambassador"`
}


func Register(c *fiber.Ctx) error {
	// data from request
	var data RegisterRequest

   err := c.BodyParser(&data)
   if err != nil {
	 return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "invalid request body",
        "details": err.Error(),
	})
   }

   // check required fields
   data.FirstName = strings.TrimSpace(data.FirstName)
   data.LastName = strings.TrimSpace(data.LastName)
   data.Email = strings.TrimSpace(strings.ToLower(data.Email))
   data.Password = strings.TrimSpace(data.Password)
   data.PasswordConfirm = strings.TrimSpace(data.PasswordConfirm)
   if data.FirstName == "" || data.LastName == "" || data.Email == "" || data.Password == "" || data.PasswordConfirm == "" {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "all fields are required",
	})
   }

	// password validation
	if data.Password != data.PasswordConfirm {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "passwords do not match",
        })
    }

	// email validation
	var existingUser models.User
	database.DB.Where("email = ?", data.Email).First(&existingUser)
	if existingUser.ID != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email already in use",
		})
	}

	// password length check
	if len(data.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 8 characters",
		})
	}		

	// hash password
	hashedPassword, err := utils.HashPassword(data.Password) 
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "cannot hash password",
			"details": err.Error(),
		})
	}

	// create a user
	   user := models.User{
        FirstName:    data.FirstName,
        LastName:     data.LastName,
        Email:        data.Email,
        Password:     hashedPassword,
        IsAmbassador: false,
    }

	// call database to save user
   if err := database.DB.Create(&user).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error":   "cannot create user",
            "details": err.Error(),
        })
    }

	// payload response
	response := UserResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		IsAmbassador: user.IsAmbassador,
	}

	return c.JSON(response)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var data LoginRequest

	err := c.BodyParser(&data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"details": err.Error(),
		})
	}

	// check required fields
	data.Email = strings.TrimSpace(strings.ToLower(data.Email))
	data.Password = strings.TrimSpace(data.Password)
	if data.Email == "" || data.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email and password are required",
		})
	}

	// find user by email
	var user models.User
	result := database.DB.Where("email = ?", data.Email).First(&user)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid email or password",
		})
	}

	// check password
	err = utils.CheckPassword(user.Password, data.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid email or password",
		})
	}

	// payload response
	response := UserResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		IsAmbassador: user.IsAmbassador,		
	}

	// generate JWT token
	cfg := config.Get()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(int(user.ID)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.JWTExpireHours))),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * time.Duration(cfg.JWTExpireHours)),
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"message": "login successful",
		"token":   token,
		"user":    response,
	})
}

// Get the authenticated user
func User(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	if cookie == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthenticated",
		})
	}

	token, err := jwt.ParseWithClaims(cookie, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		cfg := config.Get()
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthenticated",
		})
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthenticated",
		})
	}	

	var user models.User
	err = database.DB.Where("id = ?", claims.Subject).First(&user).Error
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthenticated",
		})
	}

	return c.JSON(fiber.Map{
		"id":            user.ID,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"email":         user.Email,
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
package main

import (
	"ambassador/src/config"
	"ambassador/src/database"
	"ambassador/src/routes"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main () {
	// Load configuration from .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to the database
	database.Connect()

	// Auto-migrate database models
	database.AutoMigrate()	


	app := fiber.New()

	origins := cfg.CORSOrigins
	if origins == "" {
		origins = "*"
	}

	// CORS middleware
	app.Use(cors.New(cors.Config{
	AllowOrigins:     cfg.CORSOrigins,
	AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
	AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	AllowCredentials: true,
}))

	// Setup routes
	routes.Setup(app)

	// Use the port from .env
	port := cfg.AppPort
	if port == "" {
		port = "8000"
	}

	fmt.Printf("Server running on port %s\n", port)
	err = app.Listen(":" + port)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
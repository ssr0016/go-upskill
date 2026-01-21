package main

import (
	"ambassador/src/config"
	"ambassador/src/database"
	"ambassador/src/routes"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main () {
	// Load configuration from .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

    log.Printf("Starting application in %s mode", cfg.Environment)

	// Connect to database
    if err := database.Connect(cfg); err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }

	// Run migrations
    if err := database.AutoMigrate(); err != nil {
        log.Fatalf("Database migration failed: %v", err)
    }

	// Initialize Fiber app
    app := fiber.New(fiber.Config{
        AppName:               "Ambassador API",
        ErrorHandler:          customErrorHandler,
        ReadTimeout:           10 * time.Second,
        WriteTimeout:          10 * time.Second,
        IdleTimeout:           120 * time.Second,
        DisableStartupMessage: false,
        ServerHeader:          "Ambassador",
        StrictRouting:         true,
        CaseSensitive:         true,
    })

	 // Setup routes
    routes.Setup(app, cfg)

	// Setup graceful shutdown
	setupGracefulShutdown(app)

// Start server
    log.Printf("Server starting on port %s", cfg.AppPort)
    if err := app.Listen(":" + cfg.AppPort); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}

func customErrorHandler(c *fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    message := "Internal server error"

    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
        message = e.Message
    }

    // Don't expose internal errors in production
    cfg := config.Get()
    if cfg.IsProduction() && code == fiber.StatusInternalServerError {
        message = "Internal server error"
    }

    return c.Status(code).JSON(fiber.Map{
        "error": message,
    })
}

func setupGracefulShutdown(app *fiber.App) {
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
        <-sigChan

        log.Println("Shutting down gracefully...")

        // Shutdown Fiber with timeout
        if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
            log.Printf("Error during server shutdown: %v", err)
        }

        // Close database
        if err := database.Close(); err != nil {
            log.Printf("Error closing database: %v", err)
        }

        log.Println("Shutdown complete")
        os.Exit(0)
    }()
}
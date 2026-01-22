package routes

import (
	"ambassador/src/config"
	"ambassador/src/controllers"
	"ambassador/src/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Setup(app *fiber.App, cfg *config.Config) {
    // Global middleware
    setupGlobalMiddleware(app, cfg)

    // Health check
    app.Get("/health", controllers.HealthCheck)

     // API routes group
    api := app.Group("/api")

    // Rate limiting 
    // authLimiter := limiter.New(limiter.Config{
    //     Max:        5,
    //     Expiration: 15 * time.Minute,
    //     KeyGenerator: func(c *fiber.Ctx) string {
    //         return c.IP()
    //     },
    //     LimitReached: func(c *fiber.Ctx) error {
    //         return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
    //             "error": "Too many requests. Please try again later.",
    //         })
    //     },
    //     Storage: nil, // Use in-memory storage (for production, use Redis)
    // })

    // PUBLIC AUTH ROUTES
    adminPublic := api.Group("/admin")
    adminPublic.Post("/register" , controllers.Register)
    adminPublic.Post("/login",  controllers.Login)

    // PROTECTED ADMIN ROUTES 
    adminProtected := api.Group("/admin")
    adminProtected.Use(middlewares.IsAuthenticated, middlewares.RequireScope("admin"))
    adminProtected.Get("/user", controllers.User)
    adminProtected.Post("/logout", controllers.Logout)
    adminProtected.Put("/users/info", controllers.UpdateInfo)
    adminProtected.Put("/users/password", controllers.UpdatePassword)
    // AMBASSADORS
    adminProtected.Get("/ambassadors", controllers.Ambassadors)
    // Products
    adminProtected.Get("/products", controllers.Products)
    adminProtected.Post("/products", controllers.CreateProducts)
    adminProtected.Get("/products/:id", controllers.GetProduct)
    adminProtected.Put("/products/:id", controllers.UpdateProduct)
    adminProtected.Delete("/products/:id", controllers.DeleteProduct)
    // Links
    adminProtected.Get("users/:id/links", controllers.Link)
    // Orders
    adminProtected.Get("/orders", controllers.Orders)

    /** ==================================================================== */

    // PUBLIC AMBASSADOR ROUTES
    ambassador := api.Group("/ambassador")
    ambassador.Post("/register", controllers.Register)
    ambassador.Post("/login", controllers.Login)

    ambassador.Get("/products/frontend", controllers.ProductFrontEnd)
    ambassador.Get("/products/backend", controllers.ProductBackend)

    // PROTECTED AMBASSADOR ROUTES
    ambassadorAuthenticated := ambassador.Use(middlewares.IsAuthenticated,  middlewares.RequireScope("ambassador"))
    ambassadorAuthenticated.Get("/user", controllers.User)
    ambassadorAuthenticated.Post("/logout", controllers.Logout)
    ambassadorAuthenticated.Put("/users/info", controllers.UpdateInfo)
    ambassadorAuthenticated.Put("/users/password", controllers.UpdatePassword)
    // Orders
    ambassadorAuthenticated.Get("/orders", controllers.Orders)
}

// setupGlobalMiddleware configures middleware for all routes
func setupGlobalMiddleware(app *fiber.App, cfg *config.Config) {
    // Recover from panics
    app.Use(recover.New(recover.Config{
        EnableStackTrace: !cfg.IsProduction(),
    }))

    // Request logging
    if !cfg.IsProduction() {
        app.Use(logger.New(logger.Config{
            Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
            TimeFormat: "15:04:05",
            TimeZone:   "Local",
        }))
    } else {
        // Production logging (JSON format)
        app.Use(logger.New(logger.Config{
            Format: `{"time":"${time}","status":${status},"latency":"${latency}","method":"${method}","path":"${path}","ip":"${ip}"}` + "\n",
        }))
    }

    // CORS
    app.Use(cors.New(cors.Config{
        AllowOrigins:     cfg.CORSOrigins,
        AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        AllowCredentials: true,
        MaxAge:           86400, // 24 hours
    }))
}


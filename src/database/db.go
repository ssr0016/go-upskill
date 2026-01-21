package database

import (
	"ambassador/src/config"
	"ambassador/src/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect initializes database connection with retries
func Connect(cfg *config.Config) error {
	 const maxRetries = 5
    const retryDelay = 5 * time.Second

    var err error
    var gormDB *gorm.DB

    // Configure GORM based on environment
    gormConfig := &gorm.Config{
        Logger: getLogger(cfg),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
        PrepareStmt: true, // Cache prepared statements
    }

    // Retry connection logic
    for i := 0; i < maxRetries; i++ {
        gormDB, err = gorm.Open(mysql.Open(cfg.GetDSN()), gormConfig)
        if err == nil {
            break
        }

        log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
        
        if i < maxRetries-1 {
            log.Printf("Retrying in %v...", retryDelay)
            time.Sleep(retryDelay)
        }
    }

    if err != nil {
        return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
    }

	// Get underlying SQL DB for connection pool configuration
    sqlDB, err := gormDB.DB()
    if err != nil {
        return fmt.Errorf("failed to get database instance: %w", err)
    }

    // Configure connection pool
    configureConnectionPool(sqlDB, cfg)

    // Verify connection
    if err := sqlDB.Ping(); err != nil {
        return fmt.Errorf("failed to ping database: %w", err)
    }

    DB = gormDB
    log.Println("Database connected successfully")

    return nil
}


// configureConnectionPool sets up database connection pool settings
func configureConnectionPool(sqlDB *sql.DB, cfg *config.Config) {
    // Maximum number of open connections
    maxOpenConns := 25
    if cfg.IsProduction() {
        maxOpenConns = 100
    }
    sqlDB.SetMaxOpenConns(maxOpenConns)

    // Maximum number of idle connections
    maxIdleConns := 5
    if cfg.IsProduction() {
        maxIdleConns = 10
    }
    sqlDB.SetMaxIdleConns(maxIdleConns)

    // Maximum lifetime of a connection
    sqlDB.SetConnMaxLifetime(time.Hour)

    // Maximum idle time for a connection
    sqlDB.SetConnMaxIdleTime(10 * time.Minute)

    log.Printf("Connection pool configured: MaxOpen=%d, MaxIdle=%d", maxOpenConns, maxIdleConns)
}

// getLogger returns appropriate logger based on environment
func getLogger(cfg *config.Config) logger.Interface {
    logLevel := logger.Info
    
    if cfg.IsProduction() {
        logLevel = logger.Error // Only log errors in production
    } else if cfg.Environment == "test" {
        logLevel = logger.Silent // Silent in tests
    }

    return logger.New(
        log.New(log.Writer(), "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             200 * time.Millisecond, // Log slow queries
            LogLevel:                  logLevel,
            IgnoreRecordNotFoundError: true,
            Colorful:                  !cfg.IsProduction(), // Colors in dev only
        },
    )
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
    if DB == nil {
        return fmt.Errorf("database not initialized")
    }

    log.Println("Running database migrations...")

    // Migrate all models
    if err := DB.AutoMigrate(
        &models.User{},
        &models.Product{},
        &models.Link{},
        &models.Order{},
        &models.OrderItem{},
    ); err != nil {
        return fmt.Errorf("auto migrate failed: %w", err)
    }

    log.Println("Database migrated successfully")
    return nil
}

// Close gracefully closes database connection
func Close() error {
    if DB == nil {
        return nil
    }

    sqlDB, err := DB.DB()
    if err != nil {
        return fmt.Errorf("failed to get database instance: %w", err)
    }

    if err := sqlDB.Close(); err != nil {
        return fmt.Errorf("failed to close database: %w", err)
    }

    log.Println("Database connection closed")
    return nil
}

// HealthCheck verifies database connectivity
func HealthCheck(ctx context.Context) error {
    if DB == nil {
        return fmt.Errorf("database not initialized")
    }

    sqlDB, err := DB.DB()
    if err != nil {
        return fmt.Errorf("failed to get database instance: %w", err)
    }

    // Use context with timeout
    if err := sqlDB.PingContext(ctx); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }

    return nil
}

// GetDB returns the database instance (for dependency injection)
func GetDB() *gorm.DB {
    return DB
}

// Stats returns database connection pool statistics
func Stats() sql.DBStats {
    if DB == nil {
        return sql.DBStats{}
    }

    sqlDB, err := DB.DB()
    if err != nil {
        return sql.DBStats{}
    }

    return sqlDB.Stats()
}
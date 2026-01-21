package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
    // Application
    Environment string
    AppPort     string
    
    // Database
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    
    // JWT
    JWTSecret      string
    JWTExpireHours int
    
    // CORS
    CORSOrigins string
}

var (
    cfg  *Config
    once sync.Once
    mu   sync.RWMutex
)

// Load initializes configuration from environment variables
func Load() (*Config, error) {
    var loadErr error
    
    once.Do(func() {
        // Load .env file in development
        if err := godotenv.Load(); err != nil {
            // Only log in development - production uses system env vars
            if os.Getenv("ENVIRONMENT") != "production" {
                log.Printf("Warning: .env file not found, using system environment variables")
            }
        }

        config := &Config{
            Environment:    getEnv("ENVIRONMENT", "development"),
            AppPort:        getEnv("APP_PORT", "8000"),
            DBHost:         getEnv("DB_HOST", "localhost"),
            DBPort:         getEnv("DB_PORT", "3306"),
            DBUser:         getEnv("DB_USER", "root"),
            DBPassword:     getEnv("DB_PASSWORD", ""),
            DBName:         getEnv("DB_NAME", "ambassador"),
            JWTSecret:      getEnv("JWT_SECRET", ""),
            JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
            CORSOrigins:    getEnv("CORS_ORIGINS", "http://localhost:3000"),
        }

        // Validate configuration
        if err := config.Validate(); err != nil {
            loadErr = fmt.Errorf("config validation failed: %w", err)
            return
        }

        cfg = config
    })

    return cfg, loadErr
}

// Get returns the singleton config instance
func Get() *Config {
    mu.RLock()
    defer mu.RUnlock()
    
    if cfg == nil {
        mu.RUnlock()
        mu.Lock()
        defer mu.Unlock()
        
        if cfg == nil {
            var err error
            cfg, err = Load()
            if err != nil {
                log.Fatalf("Failed to load configuration: %v", err)
            }
        }
        return cfg
    }
    
    return cfg
}

// Validate checks if all required configuration values are set
func (c *Config) Validate() error {
    // Critical validations for production
    if c.IsProduction() {
        if c.JWTSecret == "" || c.JWTSecret == "your-secret-key-change-this" {
            return errors.New("JWT_SECRET must be set in production")
        }
        
        if len(c.JWTSecret) < 32 {
            return errors.New("JWT_SECRET must be at least 32 characters in production")
        }
        
        if c.CORSOrigins == "*" {
            return errors.New("CORS_ORIGINS cannot be wildcard (*) in production")
        }
        
        if c.DBPassword == "" {
            return errors.New("DB_PASSWORD must be set in production")
        }
    }

    // General validations
    if c.DBHost == "" {
        return errors.New("DB_HOST is required")
    }
    
    if c.DBName == "" {
        return errors.New("DB_NAME is required")
    }
    
    if c.JWTExpireHours <= 0 || c.JWTExpireHours > 168 { // Max 7 days
        return errors.New("JWT_EXPIRE_HOURS must be between 1 and 168")
    }

    return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
    return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
    return c.Environment == "development"
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        c.DBUser,
        c.DBPassword,
        c.DBHost,
        c.DBPort,
        c.DBName,
    )
}

// MaskSensitive returns a copy of config with sensitive values masked (for logging)
func (c *Config) MaskSensitive() map[string]string {
    return map[string]string{
        "ENVIRONMENT":      c.Environment,
        "APP_PORT":         c.AppPort,
        "DB_HOST":          c.DBHost,
        "DB_PORT":          c.DBPort,
        "DB_USER":          c.DBUser,
        "DB_PASSWORD":      "****",
        "DB_NAME":          c.DBName,
        "JWT_SECRET":       "****",
        "JWT_EXPIRE_HOURS": strconv.Itoa(c.JWTExpireHours),
        "CORS_ORIGINS":     c.CORSOrigins,
    }
}

// Helper functions

func getEnv(key, defaultVal string) string {
    if value, exists := os.LookupEnv(key); exists && value != "" {
        return value
    }
    return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
    valStr := os.Getenv(key)
    if valStr == "" {
        return defaultVal
    }
    
    val, err := strconv.Atoi(valStr)
    if err != nil {
        log.Printf("Warning: invalid integer value for %s: %s, using default: %d", 
            key, valStr, defaultVal)
        return defaultVal
    }
    
    return val
}

func getEnvBool(key string, defaultVal bool) bool {
    valStr := os.Getenv(key)
    if valStr == "" {
        return defaultVal
    }
    
    val, err := strconv.ParseBool(valStr)
    if err != nil {
        log.Printf("Warning: invalid boolean value for %s: %s, using default: %t", 
            key, valStr, defaultVal)
        return defaultVal
    }
    
    return val
}
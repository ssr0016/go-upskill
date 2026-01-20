package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
	JWTExpireHours int
	AppPort       string
	CORSOrigins   string
}

var cfg *Config

func Load() (*Config, error) {
	// Load .env file (not required, will use system env vars if not present)
	_ = godotenv.Load()

	config := &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "ambassador"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-this"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		AppPort:       getEnv("APP_PORT", "8000"),
		CORSOrigins: getEnv("CORS_ORIGINS", "*"),
	}

	cfg = config
	return config, nil
}

func Get() *Config {
	if cfg == nil {
		Load()
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, err := strconv.Atoi(getEnv(key, "")); err == nil {
		return val
	}
	
	return defaultVal
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

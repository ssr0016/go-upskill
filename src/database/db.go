package database

import (
	"ambassador/src/config"
	"ambassador/src/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error

	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config")
	}

	DB, err = gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}
}

func AutoMigrate() {
	DB.AutoMigrate(models.User{})
}
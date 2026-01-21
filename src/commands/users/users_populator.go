package main

import (
	"ambassador/src/config"
	"ambassador/src/database"
	"ambassador/src/models"
	"ambassador/src/utils"
	"log"

	"github.com/bxcodec/faker/v3"
)

func main() {
    // Load config and connect to database
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    if err := database.Connect(cfg); err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer database.Close()
    
    log.Println("Creating 30 fake ambassador users...")
    
    var createdCount int
    for i := 0; i < 30; i++ {
        ambassador := models.User{
            FirstName:    faker.FirstName(),
            LastName:     faker.LastName(),
            Email:        faker.Email(),
            IsAmbassador: true,
        }

        hashPassword, err := utils.HashPassword("12345678")
        if err != nil {
            log.Printf("Failed to hash password for user %d: %v", i, err)
            continue
        }
        ambassador.Password = hashPassword

        if err := database.DB.Create(&ambassador).Error; err != nil {
            log.Printf("Failed to create user %d: %v", i, err)
            continue
        }
        
        createdCount++
        if createdCount%5 == 0 {
            log.Printf("Created %d/30 users", createdCount)
        }
    }
    
    log.Printf("Successfully created %d fake ambassador users!", createdCount)
}

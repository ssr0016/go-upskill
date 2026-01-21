package main

import (
	"ambassador/src/config"
	"ambassador/src/database"
	"ambassador/src/models"
	"fmt"
	"log"
	"math/rand"

	"github.com/bxcodec/faker/v3"
)


func main() {
    // Load config and connect to database (same as users)
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    if err := database.Connect(cfg); err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer database.Close()
    
    fmt.Println("Creating 50 fake products...")
    
    var createdCount int
    productNames := []string{
        "iPhone 15 Pro", "MacBook Pro M3", "AirPods Pro 2", "iPad Pro", "Apple Watch Ultra",
        "Samsung Galaxy S25", "Google Pixel 9", "Sony WH-1000XM5", "Dell XPS 15", "Nike Air Max",
        "Adidas Ultraboost", "PS5 Slim", "Xbox Series X", "Nintendo Switch OLED", "DJI Mini 4 Pro",
        "GoPro Hero 12", "Kindle Paperwhite", "Dyson V15", "Samsung 4K TV", "LG OLED C4",
    }
    
    for i := 0; i < 50; i++ {
        // Random product details
        product := models.Product{
            Title:       productNames[rand.Intn(len(productNames))],
            Description: faker.Sentence(),
            Image:       fmt.Sprintf("https://picsum.photos/300/400?random=%d", i),
            Price:       9.99 + rand.Float64()*990.01, // $9.99 - $1000
        }

        if err := database.DB.Create(&product).Error; err != nil {
            fmt.Printf("Failed to create product %d: %v\n", i, err)
            continue
        }
        
        createdCount++
        if createdCount%10 == 0 {
            fmt.Printf("Created %d/50 products\n", createdCount)
        }
    }
    
    fmt.Printf("\n Successfully created %d fake products!\n", createdCount)
}

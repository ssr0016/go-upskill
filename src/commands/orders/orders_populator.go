package main

import (
	"ambassador/src/config"
	"ambassador/src/database"
	"ambassador/src/models"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bxcodec/faker/v3"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    if err := database.Connect(cfg); err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer database.Close()

    fmt.Println("Creating 100 fake orders using your 30 ambassadors...")

    var users []models.User
    var products []models.Product
    if err := database.DB.Where("is_ambassador = ?", true).Find(&users).Error; err != nil || len(users) == 0 {
        log.Fatal("No ambassadors found!")
    }
    if err := database.DB.Find(&products).Error; err != nil || len(products) == 0 {
        log.Fatal("No products found!")
    }

    fmt.Printf("Using %d ambassadors + %d products\n", len(users), len(products))

    rand.Seed(time.Now().UnixNano())
    streetNames := []string{"Main St", "Quezon Ave", "EDSA", "Makati Ave", "Ayala Ave"}
    cities := []string{"Makati", "Manila", "Quezon City", "Cebu", "Davao"}
    countries := []string{"Philippines", "USA", "Canada"}

    var createdCount int

    for i := 0; i < 100; i++ {
        // Use your existing ambassadors (IDs 6-35)
        user := users[rand.Intn(len(users))]
        numItems := rand.Intn(3) + 1

        address := fmt.Sprintf("%d %s", rand.Intn(999)+1, streetNames[rand.Intn(len(streetNames))])
        city := cities[rand.Intn(len(cities))]
        country := countries[rand.Intn(len(countries))]
        zip := fmt.Sprintf("%04d", rand.Intn(9999))

        order := models.Order{
            TransactionID:   fmt.Sprintf("txn_%08d", i+1000),
            UserID:          user.ID,  //  the existing ambassadors 6-35
            Code:            fmt.Sprintf("ORDER%03d", i+1),
            AmbassadorEmail: faker.Email(),
            FirstName:       faker.FirstName(),
            LastName:        faker.LastName(),
            Email:           faker.Email(),
            Address:         address,
            City:            city,
            Country:         country,
            Zip:             zip,
            Complete:        rand.Float64() > 0.3, // 70% complete
        }

        // Add 1-3 random products
        order.OrderItems = make([]models.OrderItem, numItems)
        for j := 0; j < numItems; j++ {
            product := products[rand.Intn(len(products))]
            qty := uint(rand.Intn(3) + 1)
            order.OrderItems[j] = models.OrderItem{
                ProductID:         product.ID,
                ProductTitle:      product.Title,
                Price:             product.Price,
                Quantity:          qty,
                AdminRevenue:      product.Price * float64(qty) * 0.7,
                AmbassadorRevenue: product.Price * float64(qty) * 0.3,
            }
        }

        if err := database.DB.Create(&order).Error; err != nil {
            fmt.Printf("Failed order %d: %v\n", i, err)
            continue
        }

        createdCount++
        if createdCount%10 == 0 {
            fmt.Printf("Created %d/100 orders\n", createdCount)
        }
    }

    fmt.Printf("\n Successfully created %d orders for your %d ambassadors!\n", createdCount, len(users))
}

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

    fmt.Println(" Clearing old data...")
    database.DB.Exec("DELETE FROM order_items")
    database.DB.Exec("DELETE FROM orders")

    fmt.Println(" Creating 100 REAL revenue orders for ambassadors...")

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
        // 🔥 FIX 1: Pick RANDOM ambassador (not buyer)
        ambassador := users[rand.Intn(len(users))]
        
        // 🔥 FIX 2: Random BUYER (NOT ambassador)
        buyerID := uint(rand.Intn(1000) + 1)
        
        numItems := rand.Intn(3) + 1
        address := fmt.Sprintf("%d %s", rand.Intn(999)+1, streetNames[rand.Intn(len(streetNames))])
        city := cities[rand.Intn(len(cities))]
        country := countries[rand.Intn(len(countries))]
        zip := fmt.Sprintf("%04d", rand.Intn(9999))

        order := models.Order{
            TransactionID:   fmt.Sprintf("txn_%08d", i+1000),
            UserID:          buyerID,                    //  Different buyer
            Code:            fmt.Sprintf("ORD%07d", i+1), //  8 chars max!
            AmbassadorEmail: ambassador.Email,           //  REAL ambassador email!
            FirstName:       faker.FirstName(),
            LastName:        faker.LastName(),
            Email:           faker.Email(),
            Address:         address,
            City:            city,
            Country:         country,
            Zip:             zip,
            Complete:        true,                        // ✅ Always complete
        }

        // Add 1-3 random products with REAL revenue
        order.OrderItems = make([]models.OrderItem, numItems)
        totalRevenue := 0.0
        for j := 0; j < numItems; j++ {
            product := products[rand.Intn(len(products))]
            qty := uint(rand.Intn(3) + 1)
            itemRevenue := product.Price * float64(qty) * 0.3
            totalRevenue += itemRevenue
            
            order.OrderItems[j] = models.OrderItem{
                ProductID:         product.ID,
                ProductTitle:      product.Title,
                Price:             product.Price,
                Quantity:          qty,
                AdminRevenue:      product.Price * float64(qty) * 0.7,
                AmbassadorRevenue: itemRevenue, // Ambassador gets 30%
            }
        }

        if err := database.DB.Create(&order).Error; err != nil {
            fmt.Printf("❌ Failed order %d: %v\n", i, err)
            continue
        }

        createdCount++
        if createdCount%10 == 0 {
            fmt.Printf("Created %d/100 orders (Jaquan got $%.2f so far)\n", createdCount, totalRevenue)
        }
    }

    fmt.Printf("\nSUCCESS! Created %d REAL revenue orders!\n", createdCount)
    fmt.Println("Test now:")
    fmt.Println("curl -H \"Authorization: Bearer YOUR_TOKEN\" http://localhost:8000/api/ambassador/rankings")
}

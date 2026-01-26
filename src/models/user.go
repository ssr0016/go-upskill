package models

import (
	"gorm.io/gorm"
)

type User struct {
    Model
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Email string        `gorm:"uniqueIndex;size:255" json:"email"`
    Password     []byte `json:"-"` // hides password in JSON responses
    IsAmbassador bool   `json:"-"`
    Revenue      float64 `json:"revenue,omitempty" gorm:"-"`
}

func (u *User) CalculateRevenue(db *gorm.DB) float64 {
    if u.IsAmbassador {
        return u.calculateAmbassadorRevenue(db)
    }
    return u.calculateAdminRevenue(db)
}

func (u *User) Name() string {
    return u.FirstName + " " + u.LastName
}

func (u *User) calculateAdminRevenue(db *gorm.DB) float64 {
    var revenue float64
    var orders []Order
    
    db.Preload("OrderItems").
       Where("complete = ?", true).
       Find(&orders)
       
    for _, order := range orders {
        for _, item := range order.OrderItems {
            revenue += item.AdminRevenue
        }
    }
    return revenue
}

func (u *User) calculateAmbassadorRevenue(db *gorm.DB) float64 {
    var revenue float64
    
    err := db.Raw(`
        SELECT COALESCE(SUM(oi.ambassador_revenue), 0) as total
        FROM orders o
        JOIN order_items oi ON o.id = oi.order_id
        WHERE o.ambassador_email = ? 
          AND o.complete = true 
          AND o.user_id != ?
    `, u.Email, u.ID).Scan(&revenue).Error  // 🔥 Removed oi.ambassador_revenue > 0
    
    if err != nil {
        return 0
    }
    
    return revenue
}


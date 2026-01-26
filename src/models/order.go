package models

type Order struct {
    Model
   TransactionID   string `gorm:"size:50;uniqueIndex" json:"transaction_id" validate:"required,min=10"`
    UserID          uint   `gorm:"index" json:"user_id" validate:"required,gt=0"`
    Code            string `gorm:"size:10;uniqueIndex" json:"code" validate:"required,len=10"`
    AmbassadorEmail string `gorm:"size:100;index" json:"ambassador_email" validate:"omitempty,email"`
    FirstName       string `gorm:"size:50;not null" json:"-" validate:"required,min=2,max=50"`
    LastName        string `gorm:"size:50;not null" json:"-" validate:"required,min=2,max=50"`
    Name            string `gorm:"-" json:"name"`
    Email           string `gorm:"size:100;not null" json:"email" validate:"required,email"`
    Address         string `gorm:"size:255;not null" json:"address" validate:"required,min=5"`
    City            string `gorm:"size:50;not null" json:"city" validate:"required,min=2"`
    Country         string `gorm:"size:50;not null" json:"country" validate:"required,min=2"`
    Zip             string `gorm:"size:20" json:"zip" validate:"omitempty"`
    Complete        bool   `gorm:"default:false" json:"-"`
    Total float64 `json:"total" gorm:"-"`

	// Relationships
	OrderItems []OrderItem `json:"order_items" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
    Model
    OrderID           uint    `gorm:"column:order_id;index;not null" json:"order_id" validate:"required,gt=0"`
    ProductID         uint    `gorm:"column:product_id;index" json:"product_id" validate:"required,gt=0"`
    ProductTitle      string  `gorm:"size:255;not null" json:"product_title" validate:"required,min=1,max=255"`
    Price             float64 `gorm:"type:decimal(10,2);not null" json:"price" validate:"required,gt=0"`
    Quantity          uint    `gorm:"not null;default:1" json:"quantity" validate:"required,gte=1"`
    AdminRevenue      float64 `gorm:"type:decimal(10,2);not null;default:0" json:"admin_revenue" validate:"required,gte=0"`
    AmbassadorRevenue float64 `gorm:"type:decimal(10,2);not null;default:0" json:"ambassador_revenue" validate:"required,gte=0"`

    Order             Order   `gorm:"foreignKey:OrderID" json:"-"`
}

func (order *Order) FullName() string {
    return order.FirstName + " " + order.LastName
}

func (order *Order) GetTotal() float64 {
    var total float64 = 0

    for _, orderItem := range order.OrderItems {
        total += orderItem.Price * float64(orderItem.Quantity)
    }

    return total
}


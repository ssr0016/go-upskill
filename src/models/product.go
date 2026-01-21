package models

type Product struct {
    Model
    Title       string  `gorm:"size:255;not null;index" json:"title"` 
    Description string  `gorm:"type:text"`
    Image       string  `gorm:"size:500"`
    Price       float64 `gorm:"type:decimal(10,2);not null" json:"price"`
}
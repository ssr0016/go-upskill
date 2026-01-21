package models

import "gorm.io/gorm"

type Product struct {
    ID          uint    `gorm:"primaryKey"`
    Title       string  `gorm:"size:255;not null;index" json:"title"` 
    Description string  `gorm:"type:text"`
    Image       string  `gorm:"size:500"`
    Price       float64 `gorm:"type:decimal(10,2);not null" json:"price"`
    gorm.Model
}
package models

import "time"

type Link struct {
    ID        uint       `json:"id" gorm:"primaryKey"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
    Code      string     `json:"code" gorm:"type:varchar(255);uniqueIndex;not null"` // ← ADD type:varchar(255)
    UserID    uint       `json:"user_id"`
    
    // Associations
    User     User      `json:"user" gorm:"foreignKey:UserID"`
    Products []Product `json:"products" gorm:"many2many:link_products;"` // ← Make sure this exists!
    Orders   []Order   `json:"orders,omitempty" gorm:"-"`
}
package models

import (
	"gorm.io/gorm"
)

type User struct {
    ID           uint   `gorm:"primaryKey" json:"id"`
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Email string        `gorm:"uniqueIndex;size:255" json:"email"`
    Password     []byte `json:"-"` // hides password in JSON responses
    IsAmbassador bool   `json:"-"`
    gorm.Model            // optional: adds CreatedAt, UpdatedAt, DeletedAt
}






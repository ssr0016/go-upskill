package models

import (
	"gorm.io/gorm"
)

type User struct {
    ID           uint   `gorm:"primaryKey" json:"id"`
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Email        string `gorm:"unique" json:"email"`
    Password     []byte `json:"-"` // hides password in JSON responses
    IsAmbassador bool   `json:"-"`
    gorm.Model            // optional: adds CreatedAt, UpdatedAt, DeletedAt
}





package models

type User struct {
    Model
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Email string        `gorm:"uniqueIndex;size:255" json:"email"`
    Password     []byte `json:"-"` // hides password in JSON responses
    IsAmbassador bool   `json:"-"`
}






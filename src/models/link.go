package models

type Link struct {
    Model
    Code    string    `gorm:"uniqueIndex;size:6;not null" json:"code"`
    UserID  uint      `json:"user_id" gorm:"index"`
    User    User      `json:"user" gorm:"foreignKey:UserID"`
    Products []Product `json:"products" gorm:"many2many:link_products;"`
}

package models

import (
	"time"
)

type Product struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	CategoryID  uint      `json:"category_id"`
	Category    Category  `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"not null;default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *Product) TableName() string {
	return "products"
}

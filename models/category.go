package models

import (
	"time"
)

type Category struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Products    []Product `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}

func (u *Category) TableName() string {
	return "product_categories"
}

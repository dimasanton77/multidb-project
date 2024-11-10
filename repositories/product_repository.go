package repositories

import (
	"fmt"

	"github.com/dimasanton77/multidb-project/config"
	"github.com/dimasanton77/multidb-project/models"
)

type ProductRepository struct{}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

func (r *ProductRepository) FindAll() ([]models.Product, error) {
	var products []models.Product
	fmt.Println("masuk kesini")
	err := config.DBMerged.Preload("Category").Find(&products).Joins("LEFT JOIN product_categories ON products.category_id = product_categories.id").Where("product_categories.name = ?", "Tes2").Error
	return products, err
}

func (r *ProductRepository) FindByID(id uint) (*models.Product, error) {
	var product models.Product
	err := config.DBMerged.Preload("Category").First(&product, id).Error
	return &product, err
}

func (r *ProductRepository) Create(product *models.Product) error {
	return config.DBMerged.Create(product).Error
}

func (r *ProductRepository) Update(product *models.Product) error {
	return config.DBMerged.Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return config.DBMerged.Delete(&models.Product{}, id).Error
}

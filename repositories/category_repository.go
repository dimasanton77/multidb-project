package repositories

import (
	"github.com/dimasanton77/multidb-project/config"
	"github.com/dimasanton77/multidb-project/models"
)

type CategoryRepository struct{}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{}
}

func (r *CategoryRepository) FindAll() ([]models.Category, error) {
	var categories []models.Category
	err := config.DBMerged.Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) FindByID(id uint) (*models.Category, error) {
	var category models.Category
	err := config.DBMerged.First(&category, id).Error
	return &category, err
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return config.DBMerged.Create(category).Error
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return config.DBMerged.Save(category).Error
}

func (r *CategoryRepository) Delete(id uint) error {
	return config.DBMerged.Delete(&models.Category{}, id).Error
}

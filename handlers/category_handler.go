package handlers

import (
	"net/http"
	"strconv"

	"github.com/dimasanton77/multidb-project/models"
	"github.com/dimasanton77/multidb-project/repositories"
	"github.com/labstack/echo/v4"
)

type CategoryHandler struct {
	repo *repositories.CategoryRepository
}

func NewCategoryHandler(repo *repositories.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

func (h *CategoryHandler) GetAll(c echo.Context) error {
	categories, err := h.repo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
		"data":    categories,
	})
}

func (h *CategoryHandler) GetByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	category, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Category not found",
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
		"data":    category,
	})
}

func (h *CategoryHandler) Create(c echo.Context) error {
	category := new(models.Category)
	if err := c.Bind(category); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	if err := h.repo.Create(category); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "success",
		"data":    category,
	})
}

func (h *CategoryHandler) Update(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	category := new(models.Category)
	if err := c.Bind(category); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	category.ID = uint(id)
	if err := h.repo.Update(category); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
		"data":    category,
	})
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Category deleted successfully",
	})
}

package handlers

import (
	"net/http"
	"strconv"

	"github.com/dimasanton77/multidb-project/models"
	"github.com/dimasanton77/multidb-project/repositories"
	"github.com/labstack/echo/v4"
)

type ProductHandler struct {
	repo *repositories.ProductRepository
}

func NewProductHandler(repo *repositories.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) GetAll(c echo.Context) error {
	products, err := h.repo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
		"data":    products,
	})
}

func (h *ProductHandler) GetByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	product, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Product not found",
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
		"data":    product,
	})
}

func (h *ProductHandler) Create(c echo.Context) error {
	product := new(models.Product)
	if err := c.Bind(product); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	if err := h.repo.Create(product); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	// Fetch the complete product with category after creation
	createdProduct, err := h.repo.FindByID(product.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "success",
		"data":    createdProduct,
	})
}

func (h *ProductHandler) Update(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	product := new(models.Product)
	if err := c.Bind(product); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	product.ID = uint(id)
	if err := h.repo.Update(product); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	// Fetch the updated product with category
	updatedProduct, err := h.repo.FindByID(product.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
		"data":    updatedProduct,
	})
}

func (h *ProductHandler) Delete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

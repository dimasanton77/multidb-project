package routes

import (
	"github.com/dimasanton77/multidb-project/handlers"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, categoryHandler *handlers.CategoryHandler, productHandler *handlers.ProductHandler) {
	// Category routes
	e.GET("/categories", categoryHandler.GetAll)
	e.GET("/categories/:id", categoryHandler.GetByID)
	e.POST("/categories", categoryHandler.Create)
	e.PUT("/categories/:id", categoryHandler.Update)
	e.DELETE("/categories/:id", categoryHandler.Delete)

	// Product routes
	e.GET("/products", productHandler.GetAll)
	e.GET("/products/:id", productHandler.GetByID)
	e.POST("/products", productHandler.Create)
	e.PUT("/products/:id", productHandler.Update)
	e.DELETE("/products/:id", productHandler.Delete)
}

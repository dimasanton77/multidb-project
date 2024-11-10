package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dimasanton77/multidb-project/config"
	"github.com/dimasanton77/multidb-project/handlers"
	"github.com/dimasanton77/multidb-project/repositories"
	"github.com/dimasanton77/multidb-project/routes"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	if err := config.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize repositories
	categoryRepo := repositories.NewCategoryRepository()
	productRepo := repositories.NewProductRepository()

	// Initialize handlers
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	productHandler := handlers.NewProductHandler(productRepo)

	// Setup routes
	routes.SetupRoutes(e, categoryHandler, productHandler)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "9900"
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

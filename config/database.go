package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dimasanton77/multidb-project/pkg/dbmerged"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	MaxConn     int
	MinConn     int
	MaxIdleTime int
	MaxLifeTime int
}

var DBMerged *dbmerged.MergedDB

func loadDBConfig(prefix string) DBConfig {
	maxConn, _ := strconv.Atoi(os.Getenv(prefix + "DB_MAX_CONN"))
	minConn, _ := strconv.Atoi(os.Getenv(prefix + "DB_MIN_CONN"))
	maxIdleTime, _ := strconv.Atoi(os.Getenv(prefix + "DB_MAX_IDLE_TIME"))
	maxLifeTime, _ := strconv.Atoi(os.Getenv(prefix + "DB_MAX_LIFE_TIME"))

	return DBConfig{
		Host:        os.Getenv(prefix + "DB_HOST"),
		Port:        os.Getenv(prefix + "DB_PORT"),
		User:        os.Getenv(prefix + "DB_USER"),
		Password:    os.Getenv(prefix + "DB_PASS"),
		DBName:      os.Getenv(prefix + "DB_NAME"),
		MaxConn:     maxConn,
		MinConn:     minConn,
		MaxIdleTime: maxIdleTime,
		MaxLifeTime: maxLifeTime,
	}
}

// config/database.go

func createConnection(config DBConfig) (*gorm.DB, error) {
	fmt.Printf("Mencoba koneksi ke database: %s\n", config.DBName)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
	)

	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Info, // Log level
				IgnoreRecordNotFoundError: false,       // Log record not found error
				Colorful:                  true,        // Enable color
			},
		),
		PrepareStmt: true,
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxConn)
	sqlDB.SetMaxIdleConns(config.MinConn)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.MaxIdleTime) * time.Second)
	sqlDB.SetConnMaxLifetime(time.Duration(config.MaxLifeTime) * time.Second)

	fmt.Printf("Berhasil terkoneksi ke database: %s\n", config.DBName)
	return db, nil
}

func InitDB() error {
	fmt.Println("Memulai inisialisasi database...")

	productsConfig := loadDBConfig("PRODUCTS_")
	categoriesConfig := loadDBConfig("CATEGORIES_")

	productsDB, err := createConnection(productsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to products database: %w", err)
	}

	categoriesDB, err := createConnection(categoriesConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to categories database: %w", err)
	}

	fmt.Println("Membuat instance MergedDB...")
	DBMerged = dbmerged.NewMergedDB(productsDB)
	DBMerged.AddConnection("categories", categoriesDB)

	fmt.Println("\nConfiguring database mappings:")

	// Map tables dengan informasi database yang jelas
	DBMerged.MapTable("products", "products")
	fmt.Println("✓ Mapped 'products' table to 'products' database")

	DBMerged.MapTable("product_categories", "categories")
	fmt.Println("✓ Mapped 'product_categories' table to 'categories' database")

	// Verifikasi mapping
	fmt.Println("\nVerifying database mappings:")
	fmt.Printf("• products -> %s\n", DBMerged.GetDBNameForTable("products"))
	fmt.Printf("• product_categories -> %s\n", DBMerged.GetDBNameForTable("product_categories"))

	return nil
}

func loadMappings() (map[string]string, error) {
	mappings := make(map[string]string)
	mappingsJSON := os.Getenv("DB_MAPPINGS")

	if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
		return nil, fmt.Errorf("failed to parse DB_MAPPINGS: %w", err)
	}

	return mappings, nil
}

package dbmerged

import (
	"sync"

	"gorm.io/gorm"
)

type MergedDB struct {
	*gorm.DB
	connections map[string]*gorm.DB
	tableMap    map[string]TableMapping
	primary     *gorm.DB
	preloads    []string
	mu          sync.RWMutex
}

type TableMapping struct {
	DBName    string
	TableName string
}

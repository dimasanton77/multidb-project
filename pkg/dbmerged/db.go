package dbmerged

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func NewMergedDB(primaryDB *gorm.DB) *MergedDB {
	if primaryDB == nil {
		panic("primary database connection cannot be nil")
	}

	return &MergedDB{
		DB:          primaryDB,
		connections: make(map[string]*gorm.DB),
		tableMap:    make(map[string]TableMapping),
		primary:     primaryDB,
		preloads:    make([]string, 0),
	}
}

func (m *MergedDB) AddConnection(name string, db *gorm.DB) {
	if db == nil {
		panic(fmt.Sprintf("database connection for %s cannot be nil", name))
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[name] = db
}

func (m *MergedDB) MapTable(tableName, dbName string) {
	if tableName == "" || dbName == "" {
		panic("table name and database name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.tableMap[strings.ToLower(tableName)] = TableMapping{
		DBName:    dbName,
		TableName: tableName,
	}
}

func (m *MergedDB) getDBForTable(tableName string) *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mapping, exists := m.tableMap[strings.ToLower(tableName)]
	if !exists {
		return m.primary
	}

	db, exists := m.connections[mapping.DBName]
	if !exists {
		return m.primary
	}

	return db
}

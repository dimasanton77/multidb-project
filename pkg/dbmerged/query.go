package dbmerged

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

func (m *MergedDB) Model(value interface{}) *MergedDB {
	m.DB = m.DB.Model(value)
	return m
}

// Tambahkan interface untuk TableName
type TableNamer interface {
	TableName() string
}

func getTableName(value interface{}) string {
	fmt.Println("Getting table name for:", value)

	// Get the type
	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	// Create a new instance of the type to check TableName method
	modelInstance := reflect.New(t).Interface()

	// Check if type implements TableName method
	if tabler, ok := modelInstance.(TableNamer); ok {
		tableName := tabler.TableName()
		fmt.Printf("Using custom table name: %s\n", tableName)
		return tableName
	}

	// Fallback to type name
	name := t.Name()
	fmt.Printf("Using type name as table name: %s\n", name)
	return strings.ToLower(name)
}

func (m *MergedDB) First(dest interface{}, conds ...interface{}) *MergedDB {
	tableName := getTableName(dest)
	db := m.getDBForTable(tableName)

	if err := db.First(dest, conds...).Error; err != nil {
		m.Error = err
		return m
	}

	if err := m.processPreloads(dest, m.preloads); err != nil {
		m.Error = err
		return m
	}

	return m
}

// Implementasi method-method GORM lainnya
func (m *MergedDB) Where(query interface{}, args ...interface{}) *MergedDB {
	m.DB = m.DB.Where(query, args...)
	return m
}

func (m *MergedDB) Create(value interface{}) *MergedDB {
	tableName := getTableName(value)
	db := m.getDBForTable(tableName)
	m.Error = db.Create(value).Error
	return m
}

func (m *MergedDB) Save(value interface{}) *MergedDB {
	tableName := getTableName(value)
	db := m.getDBForTable(tableName)
	m.Error = db.Save(value).Error
	return m
}

func (m *MergedDB) Delete(value interface{}, conds ...interface{}) *MergedDB {
	tableName := getTableName(value)
	db := m.getDBForTable(tableName)
	m.Error = db.Delete(value, conds...).Error
	return m
}

//=====================

func (m *MergedDB) Preload(query string, args ...interface{}) *MergedDB {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Printf("Starting Preload for: %s\n", query)

	// Store preload for later processing
	m.preloads = append(m.preloads, query)

	// Get the model type from Statement if available
	var tableName string
	if m.Statement != nil && m.Statement.Model != nil {
		tableName = m.Statement.Table
		fmt.Printf("Table name from Statement: %s\n", tableName)
	}

	// If Statement or Table is empty, we'll determine it later during Find
	if tableName == "" {
		fmt.Println("Table name not available yet, will process during Find operation")
		return m
	}

	fmt.Printf("Checking if %s and %s are in same database\n", tableName, query)
	if m.isSameDatabase(tableName, query) {
		fmt.Printf("Tables are in same database, using normal GORM preload\n")
		m.DB = m.DB.Preload(query, args...)
	} else {
		fmt.Printf("Tables are in different databases, will handle during Find\n")
	}

	return m
}

func (m *MergedDB) Find(dest interface{}, conds ...interface{}) *MergedDB {
	fmt.Println("Starting Find operation...")

	// Get current preloads
	m.mu.RLock()
	currentPreloads := make([]string, len(m.preloads))
	copy(currentPreloads, m.preloads)
	fmt.Printf("Current preloads: %v\n", currentPreloads)
	m.mu.RUnlock()

	// Get source table name
	sourceTableName := getTableName(dest)
	fmt.Printf("Source table name: %s\n", sourceTableName)

	// Get appropriate database for source table
	sourceDB := m.getDBForTable(sourceTableName)
	if sourceDB == nil {
		m.Error = fmt.Errorf("database not found for table: %s", sourceTableName)
		return m
	}

	// Execute main query
	fmt.Printf("Executing main query on table: %s\n", sourceTableName)
	if err := sourceDB.Find(dest).Error; err != nil {
		m.Error = err
		return m
	}

	// Process preloads
	if len(currentPreloads) > 0 {
		if err := m.processPreloads(dest, currentPreloads); err != nil {
			m.Error = err
			return m
		}
	}

	return m
}

// Improved isSameDatabase function with logging
func (m *MergedDB) isSameDatabase(table1, table2 string) bool {
	fmt.Printf("Checking if tables are in same database: %s and %s\n", table1, table2)
	if table1 == "" || table2 == "" {
		fmt.Println("One or both table names are empty")
		return false
	}

	db1 := m.getDBForTable(table1)
	db2 := m.getDBForTable(table2)
	result := db1 == db2
	fmt.Printf("Tables %s and %s are in %s database\n", table1, table2, map[bool]string{true: "same", false: "different"}[result])
	return result
}

// Add these debugging methods
func (m *MergedDB) Debug() *MergedDB {
	m.DB = m.DB.Debug()
	return m
}

func (m *MergedDB) Session(config *gorm.Session) *MergedDB {
	m.DB = m.DB.Session(config)
	return m
}

//_

func (m *MergedDB) getDBName(db *gorm.DB) string {
	// Dapatkan DSN dari koneksi database
	sqlDB, err := db.DB()
	if err != nil {
		return "unknown"
	}

	stats := sqlDB.Stats()
	if stats.InUse > 0 {
		return db.Name() // atau bisa menggunakan custom cara lain untuk mendapatkan nama DB
	}

	return "unknown"
}

func (m *MergedDB) processPreloads(dest interface{}, preloads []string) error {
	value := reflect.ValueOf(dest)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Handle both slice and single item
	var items reflect.Value
	if value.Kind() == reflect.Slice {
		items = value
		fmt.Printf("Processing slice with %d items\n", items.Len())
	} else {
		items = reflect.MakeSlice(reflect.SliceOf(value.Type()), 1, 1)
		items.Index(0).Set(value)
		fmt.Println("Processing single item")
	}

	// Get source table info
	sourceType := items.Type().Elem()
	if sourceType.Kind() == reflect.Ptr {
		sourceType = sourceType.Elem()
	}
	sourceInstance := reflect.New(sourceType).Interface()
	var sourceTableName string
	if tabler, ok := sourceInstance.(TableNamer); ok {
		sourceTableName = tabler.TableName()
	} else {
		sourceTableName = strings.ToLower(sourceType.Name())
	}

	sourceDB := m.getDBForTable(sourceTableName)
	fmt.Printf("Source table '%s' is in database: %s\n", sourceTableName, m.GetDBNameForTable(sourceTableName))

	for _, preload := range preloads {
		fmt.Printf("\n=== Processing preload: %s ===\n", preload)

		// Get target model type and table name
		targetType := getPreloadType(items.Type(), preload)
		targetInstance := reflect.New(targetType).Interface()
		var targetTableName string

		if tabler, ok := targetInstance.(TableNamer); ok {
			targetTableName = tabler.TableName()
		} else {
			targetTableName = strings.ToLower(targetType.Name())
		}

		// Get target database info
		targetDBName := m.GetDBNameForTable(targetTableName)
		targetDB := m.getDBForTable(targetTableName)

		fmt.Printf("Target table '%s' is in database: %s\n", targetTableName, targetDBName)

		if targetDB == nil {
			return fmt.Errorf("database not found for table: %s", targetTableName)
		}

		// Print cross-database operation info
		if sourceDB != targetDB {
			fmt.Printf("Cross-database operation: %s.%s -> %s.%s\n",
				m.GetDBNameForTable(sourceTableName), sourceTableName,
				targetDBName, targetTableName)
		} else {
			fmt.Printf("Same-database operation in: %s\n", targetDBName)
		}

		// Load related data
		fmt.Printf("Executing query on database '%s', table '%s'\n", targetDBName, targetTableName)
		if err := m.loadRelatedData(items, preload, targetDB, targetTableName); err != nil {
			return fmt.Errorf("error loading data from %s.%s: %w", targetDBName, targetTableName, err)
		}
	}

	return nil
}

// Helper function untuk mendapatkan nama database dari table mapping
func (m *MergedDB) GetDBNameForTable(tableName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mapping, exists := m.tableMap[strings.ToLower(tableName)]
	if !exists {
		return "primary" // atau nama default database
	}
	return mapping.DBName
}

func getPreloadType(modelType reflect.Type, preload string) reflect.Type {
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() == reflect.Slice {
		modelType = modelType.Elem()
	}
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	field, _ := modelType.FieldByName(preload)
	fieldType := field.Type
	if fieldType.Kind() == reflect.Slice {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	return fieldType
}

// Helper function untuk mendapatkan instance dan nama tabel
func getInstanceAndTableName(modelType reflect.Type) (interface{}, string) {
	// Get real type (bukan pointer atau slice)
	for modelType.Kind() == reflect.Ptr || modelType.Kind() == reflect.Slice {
		modelType = modelType.Elem()
	}

	// Buat instance baru dari type
	instance := reflect.New(modelType).Interface()

	// Cek apakah implements TableNamer
	var tableName string
	if tabler, ok := instance.(TableNamer); ok {
		tableName = tabler.TableName()
		fmt.Printf("Using custom table name: %s\n", tableName)
	} else {
		tableName = strings.ToLower(modelType.Name())
		fmt.Printf("Using default table name: %s\n", tableName)
	}

	return instance, tableName
}

// processCrossDBPreload yang sudah diperbaiki
func (m *MergedDB) processCrossDBPreload(dest interface{}, preload string) error {
	fmt.Printf("Processing cross-database preload %s\n", preload)

	value := reflect.ValueOf(dest)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Handle both single item and slice
	var items reflect.Value
	if value.Kind() == reflect.Slice {
		items = value
	} else {
		items = reflect.MakeSlice(reflect.SliceOf(value.Type()), 1, 1)
		items.Index(0).Set(value)
	}

	fmt.Printf("Processing %d items for preload %s\n", items.Len(), preload)

	// Get preload model type
	modelType := items.Type().Elem()
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	field, ok := modelType.FieldByName(preload)
	if !ok {
		return fmt.Errorf("field %s not found in model", preload)
	}

	// Get target type and table name
	preloadType := field.Type
	if preloadType.Kind() == reflect.Slice {
		preloadType = preloadType.Elem()
	}
	if preloadType.Kind() == reflect.Ptr {
		preloadType = preloadType.Elem()
	}

	_, targetTableName := getInstanceAndTableName(preloadType)

	// Get target database
	targetDB := m.getDBForTable(targetTableName)
	if targetDB == nil {
		return fmt.Errorf("database not found for table: %s", targetTableName)
	}

	fmt.Printf("Loading related data from table %s\n", targetTableName)

	// Load related data using the appropriate database
	return m.loadRelatedData(items, preload, targetDB, targetTableName)
}

// loadRelatedData yang sudah disesuaikan
func (m *MergedDB) loadRelatedData(items reflect.Value, preload string, targetDB *gorm.DB, targetTableName string) error {
	fmt.Printf("Loading related data for %s from table %s\n", preload, targetTableName)

	// Get foreign key field name
	fkField := preload + "ID"
	fmt.Printf("Using foreign key field: %s\n", fkField)

	// Collect foreign keys
	var ids []interface{}
	idMap := make(map[interface{}][]int)

	for i := 0; i < items.Len(); i++ {
		item := items.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		fk := item.FieldByName(fkField)
		if !fk.IsValid() || fk.IsZero() {
			continue
		}

		id := fk.Interface()
		fmt.Printf("Found foreign key: %v\n", id)
		ids = append(ids, id)
		idMap[id] = append(idMap[id], i)
	}

	if len(ids) == 0 {
		fmt.Println("No foreign keys found to load")
		return nil
	}

	fmt.Printf("Found %d unique foreign keys: %v\n", len(ids), ids)

	// Create slice for related items
	modelType := items.Type().Elem()
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	field, _ := modelType.FieldByName(preload)
	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// Create slice to hold results
	relatedSlice := reflect.MakeSlice(reflect.SliceOf(fieldType), 0, len(ids))
	relatedPtr := reflect.New(relatedSlice.Type())
	relatedPtr.Elem().Set(relatedSlice)

	// Execute query
	fmt.Printf("Executing query on table %s: WHERE id IN %v\n", targetTableName, ids)
	query := targetDB.Table(targetTableName).Where("id IN ?", ids)
	if err := query.Find(relatedPtr.Interface()).Error; err != nil {
		return fmt.Errorf("error loading related data: %w", err)
	}

	// Map results back
	relatedSlice = relatedPtr.Elem()
	fmt.Printf("Found %d related records\n", relatedSlice.Len())

	// Map results back to original items
	for i := 0; i < relatedSlice.Len(); i++ {
		related := relatedSlice.Index(i)
		id := related.FieldByName("ID").Interface()
		fmt.Printf("Mapping related record with ID %v\n", id)

		if indices, ok := idMap[id]; ok {
			for _, idx := range indices {
				item := items.Index(idx)
				if item.Kind() == reflect.Ptr {
					item = item.Elem()
				}

				field := item.FieldByName(preload)
				if field.Kind() == reflect.Ptr {
					newValue := reflect.New(field.Type().Elem())
					newValue.Elem().Set(related)
					field.Set(newValue)
				} else {
					field.Set(related)
				}
				fmt.Printf("Mapped related record to item at index %d\n", idx)
			}
		}
	}

	return nil
}

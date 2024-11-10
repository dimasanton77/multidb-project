package dbmerged

// TableDB menyimpan mapping tabel ke database
var TableDB = map[string]string{
	"products":           "produk",
	"product_categories": "kategori",
}

// GetDBNameByTable untuk mendapatkan nama database berdasarkan nama tabel
func GetDBNameByTable(tableName string) string {
	if dbName, exists := TableDB[tableName]; exists {
		return dbName
	}
	return "produk" // default ke database produk jika tidak ditemukan
}

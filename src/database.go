package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Loads database connection information
func DbStart() {
	db, err := gorm.Open(
		sqlite.Open(AppConfiguration.Connectors.SQLite.DatabasePath),
		&gorm.Config{})
	if err != nil {
		panic(err)
	}
	Database = db
}

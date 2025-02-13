package sql

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrar o schema
	_ = DB.AutoMigrate(&Event{})
}

type Event struct {
	gorm.Model
	Name string
	Data string
}

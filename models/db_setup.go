package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres dialect for gorm
)

var (
	db *gorm.DB
)

// SetupDatabase - Creates the tables in the database
func SetupDatabase(db *gorm.DB) error {
	err := db.AutoMigrate(
		&User{},
		&Location{},
	).Error
	if err != nil {
		return err
	}

	return err
}

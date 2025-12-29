package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes a connection to the PostgreSQL database
// Note: Schema creation is handled by migrations (csd-pilotectl database init/update)
func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	return db, nil
}

// GetDB returns the global database connection
func GetDB() *gorm.DB {
	return DB
}

// AutoMigrate runs auto-migration for the given models
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}
	return DB.AutoMigrate(models...)
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

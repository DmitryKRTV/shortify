package db

import (
	"shortify/server/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Link{},
		&domain.Click{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

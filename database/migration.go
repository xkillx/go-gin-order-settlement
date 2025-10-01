package database

import (
	"github.com/xkillx/go-gin-order-settlement/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entities.Product{},
		&entities.Order{},
		&entities.Transaction{},
		&entities.Settlement{},
		&entities.Job{},
	); err != nil {
		return err
	}

	return nil
}

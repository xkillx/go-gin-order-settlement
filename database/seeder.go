package database

import (
	"github.com/xkillx/go-gin-order-settlement/database/seeders/seeds"
	"gorm.io/gorm"
)

func Seeder(db *gorm.DB) error {
	if err := seeds.ListTransactionSeeder(db); err != nil {
		return err
	}

	return nil
}

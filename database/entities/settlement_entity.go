package entities

import "time"

type Settlement struct {
	MerchantID string    `gorm:"type:text;not null;uniqueIndex:idx_settlement_unique,priority:1" db:"merchant_id" json:"merchant_id"`
	Date       time.Time `gorm:"type:date;not null;uniqueIndex:idx_settlement_unique,priority:2" db:"date" json:"date"`
	GrossCents int64     `gorm:"type:bigint;not null" db:"gross_cents" json:"gross_cents"`
	FeeCents   int64     `gorm:"type:bigint;not null" db:"fee_cents" json:"fee_cents"`
	NetCents   int64     `gorm:"type:bigint;not null" db:"net_cents" json:"net_cents"`
	TxnCount   int64     `gorm:"type:bigint;not null" db:"txn_count" json:"txn_count"`

	Timestamp
}

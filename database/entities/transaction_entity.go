package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" db:"id" json:"id"`
	MerchantID  string    `gorm:"type:text;not null;index" db:"merchant_id" json:"merchant_id"`
	AmountCents int64     `gorm:"type:bigint;not null" db:"amount_cents" json:"amount_cents"`
	FeeCents    int64     `gorm:"type:bigint;not null" db:"fee_cents" json:"fee_cents"`
	Status      string    `gorm:"type:text;not null;index" db:"status" json:"status"`
	PaidAt      time.Time `gorm:"type:timestamp with time zone;not null;index" db:"paid_at" json:"paid_at"`

	Timestamp
}

// BeforeCreate hook to ensure UUID is set for databases without uuid_generate_v4 (e.g., SQLite tests)
func (t *Transaction) BeforeCreate(_ *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	BuyerID   string    `gorm:"type:text;not null" json:"buyer_id"`
	Quantity  int       `gorm:"type:int;not null" json:"quantity"`

	Product Product `gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"product"`

	Timestamp
}

// BeforeCreate hook to ensure UUID is set for databases without uuid_generate_v4 (e.g., SQLite tests)
func (o *Order) BeforeCreate(_ *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID    uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name  string    `gorm:"type:text;not null" json:"name"`
	Stock int       `gorm:"type:int;not null;check:stock >= 0" json:"stock"`

	Orders []Order `gorm:"foreignKey:ProductID" json:"-"`

	Timestamp
}

// BeforeCreate hook to ensure UUID is set for databases without uuid_generate_v4 (e.g., SQLite tests)
func (p *Product) BeforeCreate(_ *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

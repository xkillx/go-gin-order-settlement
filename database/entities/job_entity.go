package entities

import "time"

type Job struct {
	ID              string    `gorm:"type:text;primaryKey" db:"id" json:"id"`
	Status          string    `gorm:"type:text;not null" db:"status" json:"status"`
	FromDate        time.Time `gorm:"type:date;not null" db:"from_date" json:"from_date"`
	ToDate          time.Time `gorm:"type:date;not null" db:"to_date" json:"to_date"`
	Progress        int       `gorm:"type:int;not null;default:0" db:"progress" json:"progress"`
	Processed       int64     `gorm:"type:bigint;not null;default:0" db:"processed" json:"processed"`
	Total           int64     `gorm:"type:bigint;not null;default:0" db:"total" json:"total"`
	ResultPath      string    `gorm:"type:text" db:"result_path" json:"result_path"`
	CancelRequested bool      `gorm:"type:boolean;not null;default:false" db:"cancel_requested" json:"cancel_requested"`

	Timestamp
}

package models

import "time"

type Role struct {
	ID        uint   `gorm:"primary_key;auto_increment"`
	Code      string `gorm:"varchar(15);not null"`
	Name      string `gorm:"varchar(255);not null"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

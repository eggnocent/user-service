package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID          uint      `gorm:"primary_key;autoIncrement"`
	UUID        uuid.UUID `gorm:"type:uuid;not null"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Username    string    `gorm:"type:varchar(255);not null"`
	Password    string    `gorm:"type:varchar(255);not null"`
	PhoneNumber string    `gorm:"type:varchar(255);not null"`
	Email       string    `gorm:"type:varchar(255);not null"`
	RoleID      uint      `gorm:"type:uint:not null"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Role        Role `gorm:"foreignkey:role_id;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

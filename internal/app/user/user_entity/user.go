package user_entity

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primarykey"`
	Username  string    `gorm:"type:varchar(32);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Role      string    `gorm:"type:varchar(32);default:'user'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

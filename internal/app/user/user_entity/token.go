package user_entity

//存储用户的登录状态信息
import (
	"time"
)

type Token struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiredAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	User User `gorm:"foreignKey:UserID"`
}

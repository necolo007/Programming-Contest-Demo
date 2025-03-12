package ai_entity

import "time"

type ChatHistory struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Theme     string    `gorm:"size:50;not null" json:"theme"`
	Model     string    `gorm:"size:50;not null" json:"model"`
	Role      string    `gorm:"size:20;not null" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

package ai_entity

import (
	"time"
)

// 历史记录
type ChatHistory struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"not null;index:idx_user_theme" json:"user_id"`
	Theme     string    `gorm:"size:50;not null;index:idx_user_theme" json:"theme"`
	Model     string    `gorm:"size:50;not null" json:"model"`
	Role      string    `gorm:"size:20;not null" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// 聊天主题
type ChatTheme struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"not null;index:idx_user_id" json:"user_id"`
	Theme       string    `gorm:"size:50;not null;uniqueIndex:idx_user_theme" json:"theme"`
	LastMessage time.Time `json:"last_message"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 本地缓存
type LocalChatCache struct {
	UserID      uint                   `json:"user_id"`
	Theme       string                 `json:"theme"`
	LastUpdated time.Time              `json:"last_updated"`
	Messages    []ChatHistory          `json:"messages"`
	Metadata    map[string]interface{} `json:"metadata"`
}

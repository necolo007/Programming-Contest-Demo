package story_entity

import "time"

type Story struct {
	ID        uint      `gorm:"primarykey"`
	Title     string    `gorm:"type:varchar(255);not null"`
	Content   string    `gorm:"type:text;not null"`
	Category  string    `gorm:"type:varchar(50);not null"` // 分类：法律知识或法律小故事
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

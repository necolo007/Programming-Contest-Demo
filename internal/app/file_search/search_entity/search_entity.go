package search_entity

import (
	"gorm.io/gorm"
	"time"
)

// SearchIndex 搜索索引实体
type SearchIndex struct {
	gorm.Model
	FileID      uint      `gorm:"index"`               // 关联的文件ID
	Content     string    `gorm:"type:text"`           // 文件内容
	Keywords    string    `gorm:"type:text"`           // 提取的关键词
	FileType    string    `gorm:"size:50;index"`       // 文件类型
	CreateDate  time.Time `gorm:"index"`               // 文件创建日期
	Parties     string    `gorm:"type:text"`           // 相关方信息
	Description string    `gorm:"type:text"`           // 文件描述
	IndexStatus string    `gorm:"size:20;default:new"` // 索引状态
}

// SearchLog 搜索日志实体
type SearchLog struct {
	gorm.Model
	UserID    uint      `gorm:"index"`     // 用户ID
	Query     string    `gorm:"type:text"` // 搜索查询
	QueryType string    `gorm:"size:20"`   // 查询类型
	Results   int       `gorm:"default:0"` // 结果数量
	Duration  int64     `gorm:"default:0"` // 查询耗时(ms)
	SearchAt  time.Time `gorm:"index"`     // 搜索时间
}

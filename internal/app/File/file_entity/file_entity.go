package file_entity

import (
	"gorm.io/gorm"
)

// File 文件实体
type File struct {
	gorm.Model
	Filename string `gorm:"column:filename;type:varchar(255);not null" json:"filename"` // 文件名
	Category string `gorm:"column:category;type:varchar(50)" json:"category"`           // 文件分类
	Filepath string `gorm:"column:filepath;type:varchar(255);not null" json:"filepath"` // 文件路径
	Content  string `gorm:"column:content;type:text" json:"content"`                    // 文件内容
	Author   string `gorm:"column:author;type:varchar(50)" json:"author"`               // 作者
	Status   int    `gorm:"column:status;type:int;default:1" json:"status"`             // 状态(1:正常 0:删除)
	UserID   uint   `gorm:"column:user_id;not null" json:"user_id"`                     // 用户ID
	Size     int64  `gorm:"column:size" json:"size"`                                    // 文件大小
	MIMEType string `gorm:"column:mime_type;type:varchar(100)" json:"mime_type"`        // MIME类型
	Hash     string `gorm:"column:hash;type:varchar(64);unique" json:"hash"`            // 文件哈希值
}

func (File) TableName() string {
	return "files"
}

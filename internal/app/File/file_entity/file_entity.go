package file_entity

import (
	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	Filename string `gorm:"not null" json:"filename"`  // 原始文件名
	Filepath string `gorm:"not null" json:"filepath"`  // 存储路径
	UserID   uint   `gorm:"not null" json:"user_id"`   //上传者的id
	Size     int64  `gorm:"not null" json:"size"`      // 文件大小（字节）
	MIMEType string `gorm:"not null" json:"mime_type"` // MIME 类型
	Category string `gorm:"not null" json:"category"`  // pdf / txt / word
	Hash     string `gorm:"unique" json:"hash"`        // SHA256 哈希
}

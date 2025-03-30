package file_entity

import (
	"gorm.io/gorm"
)

// File 文件实体
type File struct {
	gorm.Model
	Filename    string `gorm:"column:filename;type:varchar(255);not null" json:"filename"`                 // 文件名
	Category    string `gorm:"column:category;type:varchar(50)" json:"category"`                           // 文件分类
	Filepath    string `gorm:"column:filepath;type:varchar(255);not null" json:"filepath"`                 // 文件路径
	Status      int    `gorm:"column:status;type:int;default:1" json:"status"`                             // 状态(1:正常 0:删除)
	Public      int    `gorm:"column:public;type:int;default:0" json:"public"`                             // 在ai接口处上传的文件默认是0（私密），在分享文件处上传的接口是1
	FileType    string `gorm:"column:file_type;type:varchar(20)" json:"file_type"`                         // 文件类型
	UserID      uint   `gorm:"column:user_id;not null" json:"user_id"`                                     // 用户ID
	Size        int64  `gorm:"column:size" json:"size"`                                                    // 文件大小
	MIMEType    string `gorm:"column:mime_type;type:varchar(100)" json:"mime_type"`                        // MIME类型
	Hash        string `gorm:"column:hash;type:varchar(64);unique" json:"hash"`                            // 文件哈希值
	AuditStatus string `gorm:"column:audit_status;type:varchar(20);default:'pending'" json:"audit_status"` // 审核状态(approved:通过 pending:待审核 rejected:拒绝)
}

func (File) TableName() string {
	return "files"
}

type RejectRequest struct {
	Reason string `json:"reason" binding:"required"` // 拒绝原因
}

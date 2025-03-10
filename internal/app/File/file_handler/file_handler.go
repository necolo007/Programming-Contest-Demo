package file_handler

import (
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/core/libx"
	"Programming-Demo/internal/app/File/file_dto"
	"Programming-Demo/internal/app/File/file_entity"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var validMIMEs = map[string]string{
	"pdf":  "application/pdf",
	"txt":  "text/plain",
	"word": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}
var validExts = map[string]string{
	"pdf":  ".pdf",
	"txt":  ".txt",
	"word": ".docx", // 只支持 `.docx`，不支持 `.doc`
}

// 文件上传
func UploadFileHandler(c *gin.Context) {
	var req file_dto.UploadFileRequest
	uid := libx.Uid(c)

	// 绑定表单数据
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "表单数据绑定失败", "details": err.Error()})
		return
	}

	// 获取上传的文件信息
	fileHeader := req.File
	if fileHeader == nil {
		c.JSON(400, gin.H{"error": "未上传文件"})
		return
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// 校验扩展名
	expectedExt, extExists := validExts[req.Category]
	if !extExists || ext != expectedExt {
		c.JSON(400, gin.H{"error": fmt.Sprintf("文件扩展名不匹配，应为 %s，实际为 %s", expectedExt, ext)})
		return
	}

	// 获取 MIME 类型
	contentType := fileHeader.Header.Get("Content-Type")
	expectedMIME, mimeExists := validMIMEs[req.Category]
	if !mimeExists || !strings.HasPrefix(contentType, expectedMIME) {
		c.JSON(400, gin.H{"error": fmt.Sprintf("文件 MIME 类型不匹配，应为 %s，实际为 %s", expectedMIME, contentType)})
		return
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "无法打开文件"})
		return
	}
	defer file.Close()

	// 计算 SHA256
	hash, err := computeSHA256(file)
	if err != nil {
		c.JSON(500, gin.H{"error": "无法计算文件哈希"})
		return
	}

	// 重新定位文件
	file.Seek(0, io.SeekStart)

	// 检查数据库是否已有相同哈希的文件
	var existingFile file_entity.File
	if err := dbs.DB.Where("hash = ?", hash).First(&existingFile).Error; err == nil {
		c.JSON(409, gin.H{"error": "文件已存在"})
		return
	}

	// 生成存储路径
	savePath := fmt.Sprintf("uploads/%d_%s%s", time.Now().Unix(), hash[:8], ext)

	// 确保 `uploads/` 目录存在
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		c.JSON(500, gin.H{"error": "无法创建上传目录"})
		return
	}

	// 存储文件
	if err := saveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "文件存储失败"})
		return
	}

	// 在 UploadFileHandler 函数中修改创建文件记录的部分
	newFile := file_entity.File{
		Filename: fileHeader.Filename,
		Filepath: savePath,
		UserID:   uid,
		Size:     fileHeader.Size,
		MIMEType: contentType,
		Category: req.Category,
		Hash:     hash,
		FileType: req.Category, // 添加 FileType
		Status:   1,            // 设置状态为正常
		Author:   "",           // 可以从用户信息中获取
		Content:  "",           // 如果需要存储文件内容，可以在这里读取文件内容
	}

	if err := dbs.DB.Create(&newFile).Error; err != nil {
		c.JSON(500, gin.H{"error": "数据库存储失败"})
		return
	}

	// 返回上传成功的响应
	c.JSON(200, gin.H{
		"message": "文件上传成功",
		"file": gin.H{
			"id":       newFile.ID,
			"filename": newFile.Filename,
			"size":     newFile.Size,
			"category": newFile.Category,
			"hash":     newFile.Hash,
		},
	})
}

// 文件下载
func DownloadFileHandler(c *gin.Context) {
	fileID := c.Param("file_id") // 获取文件ID
	uid := libx.Uid(c)           // 获取当前用户的ID

	// 查找文件
	var file file_entity.File
	if err := dbs.DB.First(&file, fileID).Error; err != nil {
		c.JSON(404, gin.H{"error": "文件不存在"})
		return
	}

	// 检查文件所属的用户是否为当前请求用户
	if file.UserID != uid {
		c.JSON(403, gin.H{"error": "没有访问权限"})
		return
	}

	// 检查文件状态
	if file.Status != 1 {
		c.JSON(404, gin.H{"error": "文件已被删除或禁用"})
		return
	}

	// 获取文件扩展名
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filepath), "."))

	// 根据文件扩展名设置 Content-Type
	switch ext {
	case "pdf":
		c.Header("Content-Type", "application/pdf")
	case "txt":
		c.Header("Content-Type", "text/plain; charset=utf-8")
	case "doc", "docx":
		c.Header("Content-Type", "application/msword")
	case "xls", "xlsx":
		c.Header("Content-Type", "application/vnd.ms-excel")
	case "png":
		c.Header("Content-Type", "image/png")
	case "jpg", "jpeg":
		c.Header("Content-Type", "image/jpeg")
	default:
		c.Header("Content-Type", "application/octet-stream") // 默认的二进制文件类型
	}

	// 设置 Content-Disposition，让浏览器下载文件
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(file.Filepath))

	// 直接返回文件
	c.File(file.Filepath)
}

// 文件删除
func DeleteFileHandler(c *gin.Context) {
	fileID := c.Param("file_id")
	uid := libx.Uid(c)

	// 查找文件
	var file file_entity.File
	if err := dbs.DB.First(&file, fileID).Error; err != nil {
		c.JSON(404, gin.H{"error": "文件不存在"})
		return
	}

	// 检查文件所属的用户是否为当前请求用户
	if file.UserID != uid {
		c.JSON(403, gin.H{"error": "没有删除权限"})
		return
	}

	// 执行软删除，更新状态
	if err := dbs.DB.Model(&file).Updates(map[string]interface{}{
		"status": 0,
	}).Error; err != nil {
		c.JSON(500, gin.H{"error": "删除文件失败"})
		return
	}

	// 删除文件成功
	c.JSON(200, gin.H{"message": "文件删除成功"})
}

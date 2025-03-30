package file_handler

import (
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/core/libx"
	"Programming-Demo/internal/app/File/file_dto"
	"Programming-Demo/internal/app/File/file_entity"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

	// 生成存储路径*********************
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
		Filename:    fileHeader.Filename,
		Filepath:    savePath,
		UserID:      uid,
		Size:        fileHeader.Size,
		MIMEType:    contentType,
		Category:    req.Category,
		Hash:        hash,
		FileType:    req.Category, // 添加 FileType
		Status:      1,            // 设置状态为正常
		Public:      req.Public,   //1和0表示私密性
		AuditStatus: "pending",
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
	fileID := c.Param("id")
	uid := libx.Uid(c)

	// 查找文件
	var file file_entity.File
	if err := dbs.DB.First(&file, fileID).Error; err != nil {
		c.JSON(404, gin.H{"error": "文件不存在"})
		return
	}

	// 检查文件状态
	if file.Status != 1 {
		c.JSON(404, gin.H{"error": "文件已被删除或禁用"})
		return
	}

	// 检查审核状态
	if file.AuditStatus != "approved" {
		c.JSON(403, gin.H{"error": "文件未通过审核，无法下载"})
		return
	}

	// 检查访问权限：如果文件是公开的(Public=1)，任何人都可以下载
	// 如果文件不是公开的(Public=0)，只有文件所有者可以下载
	if file.Public == 0 && file.UserID != uid {
		c.JSON(403, gin.H{"error": "没有访问权限"})
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
	fileID := c.Param("id")
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

// 审核文件
func AuditFile(c *gin.Context) {
	fileID := c.Param("id")
	id, err := strconv.Atoi(fileID)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文件ID",
		})
		return
	}

	// 从表单中获取审核操作类型
	action := c.PostForm("action")
	if action != "approve" && action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的审核操作，必须是 approve 或 reject",
		})
		return
	}

	// 查询文件是否存在
	var file file_entity.File
	if err := dbs.DB.First(&file, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
		})
		return
	}

	// 根据操作类型决定审核状态
	switch action {
	case "approve":
		// 批准文件，直接更新状态
		file.AuditStatus = "approved"

		if err := dbs.DB.Save(&file).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新审核状态失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "文件审核通过",
			"data": gin.H{
				"id":           file.ID,
				"filename":     file.Filename,
				"audit_status": file.AuditStatus,
			},
		})

	case "reject":
		// 拒绝文件，从表单中获取拒绝原因
		reason := c.PostForm("reason")
		if reason == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "缺少拒绝原因",
			})
			return
		}

		// 更新审核状态为拒绝
		file.AuditStatus = "rejected"

		if err := dbs.DB.Save(&file).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新审核状态失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "文件已拒绝",
			"data": gin.H{
				"id":           file.ID,
				"filename":     file.Filename,
				"audit_status": file.AuditStatus,
				"reason":       reason,
			},
		})
	}
}

// 列出待审核的文件
func ListPendingFiles(c *gin.Context) {
	// 从表单中获取分页参数
	pageStr := c.DefaultPostForm("page", "1")
	pageSizeStr := c.DefaultPostForm("page_size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询待审核的文件
	var files []file_entity.File
	var total int64

	query := dbs.DB.Model(&file_entity.File{}).Where("audit_status = ?", "pending")

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取文件数量失败",
			"error":   err.Error(),
		})
		return
	}

	// 获取当前页的数据
	if err := query.Limit(pageSize).Offset(offset).Order("created_at DESC").Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取文件列表失败",
			"error":   err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	// 构建简化的文件信息
	type FileInfo struct {
		ID          uint   `json:"id"`
		Filename    string `json:"filename"`
		Category    string `json:"category"`
		FileType    string `json:"file_type"`
		UserID      uint   `json:"user_id"`
		Public      int    `json:"public"`
		AuditStatus string `json:"audit_status"`
		CreatedAt   string `json:"created_at"`
	}

	fileInfos := make([]FileInfo, 0, len(files))
	for _, file := range files {
		fileInfos = append(fileInfos, FileInfo{
			ID:          file.ID,
			Filename:    file.Filename,
			Category:    file.Category,
			FileType:    file.FileType,
			UserID:      file.UserID,
			Public:      file.Public,
			AuditStatus: file.AuditStatus,
			CreatedAt:   file.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取待审核文件列表成功",
		"data": gin.H{
			"files":       fileInfos,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// 获取待审核的文件
func GetPendingFileHandler(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil || fileID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文件ID",
		})
		return
	}

	// 查找文件
	var file file_entity.File
	if err := dbs.DB.First(&file, fileID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
		})
		return
	}

	// 检查文件状态
	if file.Status != 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件已被删除或禁用",
		})
		return
	}

	// 检查是否为待审核状态
	if file.AuditStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该文件不是待审核状态",
		})
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

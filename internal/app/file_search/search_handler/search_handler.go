package search_handler

import (
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/internal/app/File/file_entity"
	"Programming-Demo/internal/app/file_search/search_dto"
	"Programming-Demo/pkg/utils/ai"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// KeywordSearch 关键词搜索
func KeywordSearch(c *gin.Context) {
	var req search_dto.KeywordSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	var results *search_dto.SearchResponse

	if req.Fuzzy {
		// 使用 AI 进行模糊搜索
		prompt := fmt.Sprintf("请在法律文档中搜索与\"%s\"相关的内容，包括相似概念和相关联的法律术语", req.Keyword)
		aiResp, code := ai.GetAIResp(prompt)
		if code != 200 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "AI 搜索失败",
				"error":   aiResp,
			})
			return
		}

		// 构建 AI 搜索结果
		results = &search_dto.SearchResponse{
			Documents: []search_dto.DocumentResult{
				{
					ID:          0, // AI 搜索结果没有实际文件 ID
					Title:       "AI 搜索结果",
					Type:        "AI_SEARCH",
					CreateTime:  time.Now(),
					UpdateTime:  time.Now(),
					Description: aiResp,
					Relevance:   1.0,
				},
			},
			Total: 1,
		}
	} else {
		// 执行普通关键词搜索
		query := dbs.DB.Model(&file_entity.File{}).Where("Public = ?", 1).
			Where("audit_status = ?", "approved").
			Where("filename LIKE ?", "%"+req.Keyword+"%")

		var files []file_entity.File
		var total int64

		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "搜索失败",
				"error":   err.Error(),
			})
			return
		}

		if err := query.Find(&files).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "搜索失败",
				"error":   err.Error(),
			})
			return
		}

		results = &search_dto.SearchResponse{
			Documents: convertToDocumentResults(files),
			Total:     int(total),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索成功",
		"data":    results,
	})
}

// AdvancedSearch 高级搜索
func AdvancedSearch(c *gin.Context) {
	var req search_dto.AdvancedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	query := dbs.DB.Model(&file_entity.File{}).Where("Public = ?", 1).
		Where("audit_status = ?", "approved")

	// 添加查询条件
	if req.Category != "" {
		query = query.Where("file_type = ?", req.Category)
	}
	if req.Filename != "" {
		query = query.Where("filename LIKE ?", "%"+req.Filename+"%")
	}
	if !req.StartDate.IsZero() {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if !req.EndDate.IsZero() {
		query = query.Where("created_at <= ?", req.EndDate)
	}
	for _, keyword := range req.Keywords {
		if strings.TrimSpace(keyword) == "" {
			continue
		}
		query = query.Where("filename LIKE ?", "%"+keyword+"%")
	}
	var files []file_entity.File
	var total int64

	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索失败",
			"error":   err.Error(),
		})
		return
	}

	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索失败",
			"error":   err.Error(),
		})
		return
	}

	results := &search_dto.SearchResponse{
		Documents: convertToDocumentResults(files),
		Total:     int(total),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索成功",
		"data":    results,
	})
}

// SemanticSearch 语义搜索
func SemanticSearch(c *gin.Context) {
	var req search_dto.SemanticSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 使用 AI 理解查询意图
	prompt := fmt.Sprintf("请分析以下法律查询，提取关键信息并转换为结构化搜索条件：%s", req.Query)
	aiResp, code := ai.GetAIResp(prompt)
	if code != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "AI 分析失败",
			"error":   aiResp,
		})
		return
	}

	// 使用 AI 分析结果构建查询
	query := dbs.DB.Model(&file_entity.File{}).Where("Public = ?", 1).
		Where("audit_status = ?", "approved").
		Where("Filename LIKE ?", "%")

	var files []file_entity.File
	var total int64

	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索失败",
			"error":   err.Error(),
		})
		return
	}

	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索失败",
			"error":   err.Error(),
		})
		return
	}

	results := &search_dto.SearchResponse{
		Documents: convertToDocumentResults(files),
		Total:     int(total),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索成功",
		"data":    results,
	})
}

// 辅助函数：转换文件列表为文档结果
func convertToDocumentResults(files []file_entity.File) []search_dto.DocumentResult {
	results := make([]search_dto.DocumentResult, 0, len(files))
	for _, file := range files {

		results = append(results, search_dto.DocumentResult{
			ID:         file.ID,
			Title:      file.Filename,
			Type:       file.Category,
			CreateTime: file.CreatedAt,
			UpdateTime: file.UpdatedAt,
			FilePath:   file.Filepath,
			Status:     file.Status,
		})
	}
	return results
}

// SearchFileByTypeHandler 按文件类型搜索文件
func SearchFileByTypeHandler(c *gin.Context) {
	fileType := c.Query("type")
	if fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件类型不能为空",
		})
		return
	}

	var files []file_entity.File
	if err := dbs.DB.Model(&file_entity.File{}).Where("Public = ?", 1).
		Where("audit_status = ?", "approved").
		Where("file_type = ?", fileType).Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索文件失败",
			"error":   err.Error(),
		})
		return
	}

	results := &search_dto.SearchResponse{
		Documents: convertToDocumentResults(files),
		Total:     len(files),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索成功",
		"data":    results,
	})
}

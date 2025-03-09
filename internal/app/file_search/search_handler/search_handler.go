package search_handler

import (
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/internal/app/File/file_entity"
	"Programming-Demo/internal/app/file_search/search_dto"
	"Programming-Demo/pkg/utils/ai"
	"fmt"
	"math"
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
		query := dbs.DB.Model(&file_entity.File{}).
			Where("content LIKE ? OR filename LIKE ?",
				"%"+req.Keyword+"%", "%"+req.Keyword+"%")

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

	query := dbs.DB.Model(&file_entity.File{})

	// 添加查询条件
	if req.FileType != "" {
		query = query.Where("type = ?", req.FileType)
	}
	if !req.StartDate.IsZero() {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if !req.EndDate.IsZero() {
		query = query.Where("created_at <= ?", req.EndDate)
	}
	for _, keyword := range req.Keywords {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}
	for _, party := range req.Parties {
		query = query.Where("parties LIKE ?", "%"+party+"%")
	}
	if req.CustomQuery != "" {
		query = query.Where("content LIKE ?", "%"+req.CustomQuery+"%")
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
	query := dbs.DB.Model(&file_entity.File{}).
		Where("content LIKE ?", "%"+aiResp+"%")

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
		relevance := calculateRelevance(file)
		highlights := extractHighlights(file.Content)

		results = append(results, search_dto.DocumentResult{
			ID:          file.ID,
			Title:       file.Filename,
			Type:        file.Category,
			CreateTime:  file.CreatedAt,
			UpdateTime:  file.UpdatedAt,
			FilePath:    file.Filepath,
			Description: file.Content,
			Highlights:  highlights,
			Relevance:   relevance,
			Author:      file.Author,
			Status:      file.Status,
		})
	}
	return results
}

// 计算文档相关度
func calculateRelevance(file file_entity.File) float64 {
	var score float64 = 1.0

	timeDiff := time.Since(file.UpdatedAt)
	timeScore := 1.0 / (1.0 + float64(timeDiff.Hours())/24.0)

	sizeScore := math.Min(float64(len(file.Content))/1000.0, 1.0)

	typeWeight := getFileTypeWeight(file.FileType) // 使用 FileType 字段

	score = (timeScore*0.3 + sizeScore*0.3 + typeWeight*0.4) * 5

	return score
}

// 提取文档高亮片段
func extractHighlights(content string) []string {
	// 提取文档中的关键段落作为高亮显示
	highlights := make([]string, 0)

	// 简单实现：按段落分割，选择前几个非空段落
	paragraphs := strings.Split(content, "\n\n")
	for _, p := range paragraphs {
		if len(p) > 0 && len(highlights) < 3 {
			// 清理段落文本
			p = strings.TrimSpace(p)
			if len(p) > 200 {
				p = p[:200] + "..."
			}
			highlights = append(highlights, p)
		}
	}

	return highlights
}

// 获取文件类型权重
func getFileTypeWeight(fileType string) float64 {
	weights := map[string]float64{
		"doc":  0.8,
		"docx": 0.8,
		"pdf":  0.9,
		"txt":  0.6,
		// 可以添加更多文件类型的权重
	}

	if weight, ok := weights[fileType]; ok {
		return weight
	}
	return 0.5 // 默认权重
}

// SearchFileByKeywordHandler 按关键词搜索文件
func SearchFileByKeywordHandler(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "搜索关键词不能为空",
		})
		return
	}

	var files []file_entity.File
	if err := dbs.DB.Where("filename LIKE ?", "%"+keyword+"%").Find(&files).Error; err != nil {
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
	if err := dbs.DB.Where("file_type = ?", fileType).Find(&files).Error; err != nil {
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

// SearchFileByContentHandler 按文件内容搜索文件
func SearchFileByContentHandler(c *gin.Context) {
	content := c.Query("content")
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "搜索内容不能为空",
		})
		return
	}

	var files []file_entity.File
	if err := dbs.DB.Where("content LIKE ?", "%"+content+"%").Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索文件失败",
			"error":   err.Error(),
		})
		return
	}

	// 过滤掉不支持内容搜索的文件类型
	var supportedFiles []file_entity.File
	for _, file := range files {
		if isSupportedForContentSearch(file.FileType) {
			supportedFiles = append(supportedFiles, file)
		}
	}

	results := &search_dto.SearchResponse{
		Documents: convertToDocumentResults(supportedFiles),
		Total:     len(supportedFiles),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索成功",
		"data":    results,
	})
}

// isSupportedForContentSearch 判断文件类型是否支持内容搜索
func isSupportedForContentSearch(fileType string) bool {
	supportedTypes := []string{"txt", "doc", "docx", "pdf"}
	fileType = strings.ToLower(fileType)
	for _, t := range supportedTypes {
		if fileType == t {
			return true
		}
	}
	return false
}

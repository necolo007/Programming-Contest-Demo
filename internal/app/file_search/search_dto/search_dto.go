package search_dto

import (
	"time"
)

// KeywordSearchRequest 关键词搜索请求
type KeywordSearchRequest struct {
	Fuzzy   bool   `json:"fuzzy"`   // 是否启用模糊搜索
	Keyword string `json:"keyword"` // 关键词
}

// AdvancedSearchRequest 高级搜索请求
type AdvancedSearchRequest struct {
	Category  string    `json:"category"`   // 文件分类
	Filename  string    `json:"filename"`   // 文件名
	Keywords  []string  `json:"keywords"`   // 关键词列表
	StartDate time.Time `json:"start_date"` // 开始日期
	EndDate   time.Time `json:"end_date"`   // 结束日期
}

// SemanticSearchRequest 语义搜索请求
type SemanticSearchRequest struct {
	Query string `json:"query" binding:"required"` // 自然语言查询
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Total     int              `json:"total"`     // 总结果数
	Documents []DocumentResult `json:"documents"` // 文档列表
}

// DocumentResult 搜索结果文档
type DocumentResult struct {
	ID          uint      `json:"id"`          // 文档ID
	Title       string    `json:"title"`       // 标题（对应 Filename）
	Type        string    `json:"type"`        // 类型（对应 Category）
	CreateTime  time.Time `json:"create_time"` // 创建时间
	UpdateTime  time.Time `json:"update_time"` // 更新时间
	FilePath    string    `json:"file_path"`   // 文件路径
	Description string    `json:"description"` // 描述（对应 Content）
	Highlights  []string  `json:"highlights"`  // 高亮片段
	Relevance   float64   `json:"relevance"`   // 相关度
	Author      string    `json:"author"`      // 作者
	Status      int       `json:"status"`      // 状态
}

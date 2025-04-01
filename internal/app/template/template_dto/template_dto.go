package template_dto

import "gorm.io/gorm"

// 定义创建模板的请求结构
type CreateTemplatereq struct {
	Name        string `json:"name" binding:"required"`        // 模板名称
	Type        string `json:"type" binding:"required"`        // 一级分类ID
	Environment string `json:"environment" binding:"required"` // 使用环境
}

// 定义模板列表项结构
type TemplateListItem struct {
	gorm.Model
	Name            string `json:"name"`            // 模板名称
	Type            string `json:"type"`            // 类型
	CategoryName    string `json:"categoryName"`    // 类别名称
	SubCategoryName string `json:"subCategoryName"` // 子类别名称
	DocTypeName     string `json:"docTypeName"`     // 文档类型名称
	Environment     string `json:"environment"`     // 适用环境
}

// 定义模板详情响应结构
type TemplateDetailResponse struct {
	gorm.Model
	Name        string      `json:"name"`        // 模板名称
	Type        string      `json:"type"`        // 类型
	Environment string      `json:"environment"` // 适用环境
	Content     interface{} `json:"content"`     // 内容
}

// 初始化响应结构
type InitializeResponse struct {
	Message string `json:"message"` // 消息
	Count   int    `json:"count"`   // 创建数量
}

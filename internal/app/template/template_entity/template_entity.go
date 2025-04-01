package template_entity

import (
	"encoding/json"
	"gorm.io/gorm"
)

// 法律模板数据库
type LegalTemplate struct {
	gorm.Model
	Name        string          `json:"name"`        // 模板名称
	Type        string          `json:"type"`        // 模板类型(合同/起诉状/协议)
	Environment string          `json:"environment"` // 适用环境
	Content     json.RawMessage `json:"content"`     // 模板内容(JSON格式)
}

// 当事人信息
type Party struct {
	Role        string `json:"role"`                  // 角色(甲方/乙方/原告/被告等)
	Name        string `json:"name"`                  // 名称(公司或个人名称)
	Address     string `json:"address,omitempty"`     // 地址 - 可选
	ContactInfo string `json:"contactInfo,omitempty"` // 联系信息 - 可选
}

// 条款信息 - 修改了 Details 为 Content
type Clause struct {
	Title       string `json:"title"`       // 条款标题
	Description string `json:"description"` // 条款描述
	Content     string `json:"content"`     // 条款详细内容
}

// 证据信息
type Evidence struct {
	Name        string `json:"name"`               // 证据名称
	Description string `json:"description"`        // 证据描述
	FilePath    string `json:"filePath,omitempty"` // 证据文件路径 - 可选
}

// 请求信息
type Request struct {
	Type        string `json:"type"`        // 请求类型
	Description string `json:"description"` // 请求详细内容
}

// 签署信息 - 新增结构体
type Signature struct {
	Method     string `json:"method"`     // 签署方式
	DateFormat string `json:"dateFormat"` // 日期格式
}

// 文档分类目录树
type TemplateCategory struct {
	ID          string                 `json:"id"`          // 分类ID
	Name        string                 `json:"name"`        // 分类名称
	Description string                 `json:"description"` // 分类描述
	Children    []*TemplateSubCategory `json:"children,omitempty"`
}

// 文档子分类
type TemplateSubCategory struct {
	ID          string             `json:"id"`          // 子分类ID
	Name        string             `json:"name"`        // 子分类名称
	Description string             `json:"description"` // 子分类描述
	ParentID    string             `json:"parent_id"`   // 父分类ID
	Children    []*TemplateDocType `json:"children,omitempty"`
}

// 文档类型
type TemplateDocType struct {
	ID          string `json:"id"`          // 文档类型ID
	Name        string `json:"name"`        // 文档类型名称
	Description string `json:"description"` // 文档类型描述
	ParentID    string `json:"parent_id"`   // 父类ID
	TemplateID  uint   `json:"template_id,omitempty"`
}

// 模板分类元数据
type TemplateMetadata struct {
	CategoryID      string `json:"category_id"`      // 一级分类ID
	CategoryName    string `json:"category_name"`    // 一级分类名称
	SubCategoryID   string `json:"subcategory_id"`   // 二级分类ID
	SubCategoryName string `json:"subcategory_name"` // 二级分类名称
	DocTypeID       string `json:"doctype_id"`       // 三级文档类型ID
	DocTypeName     string `json:"doctype_name"`     // 三级文档类型名称
}

// 统一的模板内容结构 - 修改了几个关键字段的类型
type TemplateContent struct {
	Title           string           `json:"title"`                     // 文档标题
	Introduction    string           `json:"introduction"`              // 引言部分
	Metadata        TemplateMetadata `json:"metadata"`                  // 分类信息
	Parties         []Party          `json:"parties"`                   // 当事人列表
	Clauses         []Clause         `json:"clauses,omitempty"`         // 条款列表 - 使用修改后的 Clause
	Requests        []Request        `json:"requests,omitempty"`        // 诉讼请求列表
	Evidence        []Evidence       `json:"evidence,omitempty"`        // 证据列表
	Signature       interface{}      `json:"signature"`                 // 签署部分 - 使用 interface{} 以适应多种格式
	LegalStatement  string           `json:"legalStatement"`            // 法律声明
	AdditionalTerms interface{}      `json:"additionalTerms,omitempty"` // 附加条款 - 使用 interface{} 以适应多种格式
}

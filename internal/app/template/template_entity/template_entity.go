package template_entity

import (
	"encoding/json"
	"gorm.io/gorm"
)

type LegalTemplate struct {
	gorm.Model
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Environment string          `json:"environment"`
	Content     json.RawMessage `json:"content"` // 使用 JSON 存储
}

type ContractTemplateContent struct {
	Title           string   `json:"title"`            // 模板标题（如“租赁合同”）
	Introduction    string   `json:"introduction"`     // 引言部分（可选）
	Parties         []Party  `json:"parties"`          // 双方信息（甲方、乙方等）
	Clauses         []Clause `json:"clauses"`          // 各条款内容
	Signature       string   `json:"signature"`        // 签署条款
	LegalStatement  string   `json:"legal_statement"`  // 法律声明或效力
	AdditionalTerms []string `json:"additional_terms"` // 附加条款（可选）
}

type LawsuitTemplateContent struct {
	Title          string     `json:"title"`           // 模板标题（如“民事起诉状”）
	Introduction   string     `json:"introduction"`    // 引言部分
	Parties        []Party    `json:"parties"`         // 双方信息（原告、被告）
	Requests       []Request  `json:"requests"`        // 诉讼请求
	Evidence       []Evidence `json:"evidence"`        // 证据部分
	Signature      string     `json:"signature"`       // 签署部分
	LegalStatement string     `json:"legal_statement"` // 法律声明或效力
}

type AgreementTemplateContent struct {
	Title          string   `json:"title"`           // 模板标题（如“保密协议”）
	Introduction   string   `json:"introduction"`    // 引言部分（可选）
	Parties        []Party  `json:"parties"`         // 双方信息
	Clauses        []Clause `json:"clauses"`         // 协议条款
	Signature      string   `json:"signature"`       // 签署条款
	LegalStatement string   `json:"legal_statement"` // 法律声明
}

type Party struct {
	Role        string `json:"role"`         // 角色（如甲方、乙方、原告、被告）
	Name        string `json:"name"`         // 名称（如公司名称、个人名称）
	Address     string `json:"address"`      // 地址（可选）
	ContactInfo string `json:"contact_info"` // 联系信息（可选）
}

type Clause struct {
	Title       string `json:"title"`       // 条款标题（如“租金条款”、“诉讼请求”）
	Description string `json:"description"` // 条款描述（可选）
	Details     string `json:"details"`     // 具体条款内容
}

type Evidence struct {
	Name        string `json:"name"`        // 证据名称
	Description string `json:"description"` // 证据描述
	FilePath    string `json:"file_path"`   // 证据文件路径（如果有）
}

type Request struct {
	Type        string `json:"type"`        // 请求类型（如“赔偿请求”、“恢复原状”）
	Description string `json:"description"` // 请求详细内容
}

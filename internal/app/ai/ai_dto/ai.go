package ai_dto

type AnalyzeReq struct {
	Model string `json:"model"`
	Name  string `json:"name"`
}

type LawsBase struct {
	// 合同基本信息
	Subject  string `json:"subject"`   // 合同标的
	Purpose  string `json:"purpose"`   // 合同目的
	Location string `json:"location"`  // 签订地点
	SignDate string `json:"sign_date"` // 签订日期

	// 权利义务
	Rights      string `json:"rights"`      // 权利内容
	Obligations string `json:"obligations"` // 义务内容

	// 履行相关
	StartDate   string `json:"start_date"`  // 开始日期
	EndDate     string `json:"end_date"`    // 结束日期
	Performance string `json:"performance"` // 履行方式

	// 价格和支付
	Price   string `json:"price"`   // 价格/报酬
	Payment string `json:"payment"` // 支付方式和条件

	// 违约和争议
	Breach  string `json:"breach"`  // 违约责任
	Dispute string `json:"dispute"` // 争议解决方式

	// 其他条款
	Confidential string `json:"confidential"` // 保密条款
	Force        string `json:"force"`        // 不可抗力
	Termination  string `json:"termination"`  // 终止条件
	Additional   string `json:"additional"`   // 其他补充条款
}

type GenerateLegalDocReq struct {
	Model      string   `json:"model"`      // ai模型
	DocType    string   `json:"doc_type"`   // 文档类型：合同、协议、声明等
	Title      string   `json:"title"`      // 文档标题
	Parties    []Party  `json:"parties"`    // 相关方信息
	Content    LawsBase `json:"content"`    // 现有的 LawsReq 作为内容
	Additional string   `json:"additional"` // 额外要求
}

type Party struct {
	Type    string `json:"type"`    // 甲方、乙方等
	Name    string `json:"name"`    // 名称
	Details string `json:"details"` // 详细信息（如地址、证件号等）
}

type ComplaintBase struct {
	Court       string   `json:"court"`       // 受理法院
	Plaintiff   Party    `json:"plaintiff"`   // 原告信息
	Defendant   Party    `json:"defendant"`   // 被告信息
	Claims      []string `json:"claims"`      // 诉讼请求
	Facts       string   `json:"facts"`       // 事实与理由
	Evidence    []string `json:"evidence"`    // 证据列表
	LawBasis    []string `json:"law_basis"`   // 法律依据
	Attachments []string `json:"attachments"` // 附件清单
}

type LegalOpinionBase struct {
	Background  string   `json:"background"`  // 案件背景
	Issues      []string `json:"issues"`      // 法律问题
	Analysis    string   `json:"analysis"`    // 法律分析
	Risks       []string `json:"risks"`       // 法律风险
	Suggestions []string `json:"suggestions"` // 法律建议
	References  []string `json:"references"`  // 法律依据
}

type GenerateComplaintReq struct {
	Model   string        `json:"model"`   // AI模型
	Content ComplaintBase `json:"content"` // 起诉状内容
}

type GenerateLegalOpinionReq struct {
	Model   string           `json:"model"`   // AI模型
	Content LegalOpinionBase `json:"content"` // 法律意见书内容
}

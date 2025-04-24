package template_handler

import (
	"Programming-Demo/core/client"
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/internal/app/template/template_dto"
	"Programming-Demo/internal/app/template/template_entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/northes/go-moonshot"
	"net/http"
	"strings"
)

// 预定义的法律文件分类树
var documentCategoryTree = []*template_entity.TemplateCategory{
	{
		ID:          "contract",
		Name:        "合同类",
		Description: "约定双方或多方权利义务关系的法律文件，具有法律约束力",
		Children: []*template_entity.TemplateSubCategory{
			{
				ID:          "commercial",
				Name:        "商业合同",
				Description: "用于规范商业活动中各方权利义务的合同文件",
				ParentID:    "contract",
				Children: []*template_entity.TemplateDocType{
					{ID: "sale", Name: "买卖合同", Description: "规定买卖双方在货物交易中的权利和义务", ParentID: "commercial"},
					{ID: "service", Name: "服务合同", Description: "约定服务提供方与接受方之间的服务内容、标准和报酬等", ParentID: "commercial"},
					{ID: "processing", Name: "加工承揽合同", Description: "规定定作方与承揽方关于加工或定作物的权利义务", ParentID: "commercial"},
					{ID: "transportation", Name: "运输合同", Description: "规定承运人将旅客或货物从起点运输到终点的义务及相关责任", ParentID: "commercial"},
					{ID: "storage", Name: "保管合同", Description: "规定保管人对保管物的保管义务及相关责任", ParentID: "commercial"},
				},
			},
			{
				ID:          "realestate",
				Name:        "房地产合同",
				Description: "与房地产开发、买卖、租赁等相关的合同",
				ParentID:    "contract",
				Children: []*template_entity.TemplateDocType{
					{ID: "purchase", Name: "房屋买卖合同", Description: "规定房屋买卖双方在交易中的权利和义务", ParentID: "realestate"},
					{ID: "lease", Name: "房屋租赁合同", Description: "规定出租人与承租人关于房屋租赁的权利义务", ParentID: "realestate"},
					{ID: "construction", Name: "建设工程合同", Description: "规定发包方与承包方关于工程建设的权利义务", ParentID: "realestate"},
					{ID: "property", Name: "物业服务合同", Description: "规定物业服务企业与业主的服务内容和标准", ParentID: "realestate"},
					{ID: "landtransfer", Name: "土地使用权出让/转让合同", Description: "规定土地使用权转移的条件和双方权利义务", ParentID: "realestate"},
				},
			},
			{
				ID:          "financial",
				Name:        "金融合同",
				Description: "与借贷、担保、投资等金融活动相关的合同",
				ParentID:    "contract",
				Children: []*template_entity.TemplateDocType{
					{ID: "loan", Name: "借款合同", Description: "规定借款人与贷款人关于借款的权利义务", ParentID: "financial"},
					{ID: "guarantee", Name: "担保合同", Description: "规定担保人对债务人债务承担保证责任的合同", ParentID: "financial"},
					{ID: "leasing", Name: "融资租赁合同", Description: "规定出租人与承租人关于融资租赁业务的权利义务", ParentID: "financial"},
					{ID: "insurance", Name: "保险合同", Description: "规定投保人与保险人关于保险的权利义务", ParentID: "financial"},
					{ID: "trust", Name: "信托合同", Description: "规定委托人、受托人与受益人关于信托的权利义务", ParentID: "financial"},
				},
			},
			{
				ID:          "intellectual",
				Name:        "知识产权合同",
				Description: "与专利、商标、著作权等知识产权相关的合同",
				ParentID:    "contract",
				Children: []*template_entity.TemplateDocType{
					{ID: "patent", Name: "专利许可/转让合同", Description: "规定专利权人许可他人实施专利或转让专利权的合同", ParentID: "intellectual"},
					{ID: "trademark", Name: "商标许可/转让合同", Description: "规定商标权人许可他人使用商标或转让商标权的合同", ParentID: "intellectual"},
					{ID: "copyright", Name: "著作权许可/转让合同", Description: "规定著作权人许可他人使用作品或转让著作权的合同", ParentID: "intellectual"},
					{ID: "technology", Name: "技术开发合同", Description: "规定当事人之间就技术开发进行协作的合同", ParentID: "intellectual"},
					{ID: "franchise", Name: "特许经营合同", Description: "规定特许人授权被特许人使用其商标、商号等经营资源的合同", ParentID: "intellectual"},
				},
			},
			{
				ID:          "labor",
				Name:        "劳动用工合同",
				Description: "与劳动关系、雇佣关系相关的合同",
				ParentID:    "contract",
				Children: []*template_entity.TemplateDocType{
					{ID: "employment", Name: "劳动合同", Description: "规定用人单位与劳动者之间权利和义务的协议", ParentID: "labor"},
					{ID: "service_labor", Name: "劳务合同", Description: "当事人之间就提供劳务达成的协议", ParentID: "labor"},
					{ID: "non_compete", Name: "竞业限制协议", Description: "限制劳动者离职后从事竞争性工作的协议", ParentID: "labor"},
					{ID: "confidentiality", Name: "保密协议", Description: "约定对用人单位商业秘密保密义务的协议", ParentID: "labor"},
					{ID: "stock_option", Name: "员工持股/股权激励协议", Description: "向员工提供企业股权或期权的激励协议", ParentID: "labor"},
				},
			},
		},
	},
	{
		ID:          "lawsuit",
		Name:        "起诉状类",
		Description: "向法院或仲裁机构提出请求和主张的法律文书，用于启动诉讼或仲裁程序",
		Children: []*template_entity.TemplateSubCategory{
			{
				ID:          "civil",
				Name:        "民事诉讼文书",
				Description: "在民事诉讼程序中使用的法律文书",
				ParentID:    "lawsuit",
				Children: []*template_entity.TemplateDocType{
					{ID: "civil_complaint", Name: "民事起诉状", Description: "向法院提出民事诉讼请求的书面文书", ParentID: "civil"},
					{ID: "civil_defense", Name: "民事答辩状", Description: "被告对原告诉讼请求提出反驳意见的文书", ParentID: "civil"},
					{ID: "civil_counter", Name: "民事反诉状", Description: "被告对原告提出反诉请求的文书", ParentID: "civil"},
					{ID: "civil_appeal", Name: "民事上诉状", Description: "当事人不服一审判决向上级法院提起上诉的文书", ParentID: "civil"},
					{ID: "civil_review", Name: "再审申请书", Description: "当事人对已生效判决申请再审的文书", ParentID: "civil"},
					{ID: "execution", Name: "执行申请书", Description: "申请法院强制执行生效法律文书的申请文件", ParentID: "civil"},
					{ID: "evidence", Name: "证据清单", Description: "列明提交给法院的各项证据的文件", ParentID: "civil"},
				},
			},
			{
				ID:          "criminal",
				Name:        "刑事诉讼文书",
				Description: "在刑事诉讼程序中使用的法律文书",
				ParentID:    "lawsuit",
				Children: []*template_entity.TemplateDocType{
					{ID: "criminal_private", Name: "刑事自诉状", Description: "自诉人直接向法院提起刑事诉讼的文书", ParentID: "criminal"},
					{ID: "criminal_incidental", Name: "刑事附带民事诉讼起诉状", Description: "在刑事诉讼中一并提出民事赔偿请求的文书", ParentID: "criminal"},
					{ID: "criminal_defense", Name: "刑事辩护意见书", Description: "辩护人为被告人进行辩护的书面意见", ParentID: "criminal"},
					{ID: "criminal_appeal", Name: "刑事上诉状", Description: "不服一审刑事判决向上级法院提起上诉的文书", ParentID: "criminal"},
					{ID: "criminal_petition", Name: "刑事申诉书", Description: "对已生效刑事判决提出异议的申诉文书", ParentID: "criminal"},
				},
			},
			{
				ID:          "administrative",
				Name:        "行政诉讼文书",
				Description: "在行政诉讼程序中使用的法律文书",
				ParentID:    "lawsuit",
				Children: []*template_entity.TemplateDocType{
					{ID: "admin_complaint", Name: "行政起诉状", Description: "公民、法人或其他组织对行政行为不服提起诉讼的文书", ParentID: "administrative"},
					{ID: "admin_defense", Name: "行政答辩状", Description: "行政机关针对原告诉讼请求提出反驳意见的文书", ParentID: "administrative"},
					{ID: "admin_appeal", Name: "行政上诉状", Description: "不服一审行政判决向上级法院提起上诉的文书", ParentID: "administrative"},
					{ID: "admin_review", Name: "行政复议申请书", Description: "向行政机关的上级机关或者法定复议机关提出复议请求的文书", ParentID: "administrative"},
				},
			},
			{
				ID:          "arbitration",
				Name:        "仲裁相关文书",
				Description: "在仲裁程序中使用的法律文书",
				ParentID:    "lawsuit",
				Children: []*template_entity.TemplateDocType{
					{ID: "arb_request", Name: "仲裁申请书", Description: "申请人向仲裁委员会申请仲裁的文书", ParentID: "arbitration"},
					{ID: "arb_defense", Name: "仲裁答辩书", Description: "被申请人针对仲裁请求提出反驳意见的文书", ParentID: "arbitration"},
					{ID: "arb_counter", Name: "仲裁反请求申请书", Description: "被申请人向申请人提出反请求的文书", ParentID: "arbitration"},
					{ID: "arb_dismiss", Name: "撤销/不予执行仲裁裁决申请书", Description: "当事人申请法院撤销或不予执行仲裁裁决的文书", ParentID: "arbitration"},
				},
			},
		},
	},
	{
		ID:          "agreement",
		Name:        "协议类",
		Description: "双方或多方达成的共识文件，表明各方对某事项的意思表示一致",
		Children: []*template_entity.TemplateSubCategory{
			{
				ID:          "settlement",
				Name:        "和解/调解协议",
				Description: "当事人通过协商或第三方调解达成的解决争议的协议",
				ParentID:    "agreement",
				Children: []*template_entity.TemplateDocType{
					{ID: "civil_mediation", Name: "民事纠纷调解协议", Description: "通过调解方式解决民事纠纷的书面协议，具有法律效力", ParentID: "settlement"},
					{ID: "labor_mediation", Name: "劳动争议调解协议", Description: "通过调解方式解决劳动争议的书面协议", ParentID: "settlement"},
					{ID: "pre_court", Name: "诉前/诉中和解协议", Description: "诉讼前或诉讼过程中当事人自行达成的和解协议", ParentID: "settlement"},
					{ID: "execution_settlement", Name: "执行和解协议", Description: "在执行程序中达成的和解协议", ParentID: "settlement"},
				},
			},
			{
				ID:          "business",
				Name:        "商事合作协议",
				Description: "企业间业务合作达成的协议",
				ParentID:    "agreement",
				Children: []*template_entity.TemplateDocType{
					{ID: "joint_venture", Name: "合资/合作协议", Description: "成立合资或合作企业的基础性协议", ParentID: "business"},
					{ID: "stock_transfer", Name: "股权转让协议", Description: "转让公司股权的法律文件", ParentID: "business"},
					{ID: "capital_increase", Name: "增资协议", Description: "公司增加注册资本的协议", ParentID: "business"},
					{ID: "acquisition", Name: "收购/并购协议", Description: "企业收购或合并的法律文件", ParentID: "business"},
					{ID: "strategic", Name: "战略合作框架协议", Description: "确立长期战略合作关系的框架性协议", ParentID: "business"},
					{ID: "shareholder", Name: "股东协议", Description: "公司股东之间关于公司治理等事项的协议", ParentID: "business"},
				},
			},
			{
				ID:          "family",
				Name:        "家事协议",
				Description: "与婚姻、家庭关系相关的协议",
				ParentID:    "agreement",
				Children: []*template_entity.TemplateDocType{
					{ID: "prenup", Name: "婚前/婚内财产协议", Description: "约定婚前或婚内财产权属和处理的协议", ParentID: "family"},
					{ID: "divorce", Name: "离婚协议", Description: "夫妻双方协议离婚的条件和财产子女等安排", ParentID: "family"},
					{ID: "child_support", Name: "子女抚养协议", Description: "关于子女抚养权、抚养费等事项的协议", ParentID: "family"},
					{ID: "inheritance", Name: "继承/遗产分割协议", Description: "继承人之间关于遗产分配的协议", ParentID: "family"},
				},
			},
			{
				ID:          "debt",
				Name:        "债权债务协议",
				Description: "与债权债务关系处理相关的协议",
				ParentID:    "agreement",
				Children: []*template_entity.TemplateDocType{
					{ID: "debt_restructure", Name: "债务重组协议", Description: "对债务条件进行调整的协议", ParentID: "debt"},
					{ID: "debt_transfer", Name: "债权转让协议", Description: "债权人将债权转让给第三人的协议", ParentID: "debt"},
					{ID: "repayment", Name: "还款协议", Description: "债务人与债权人就还款事宜达成的协议", ParentID: "debt"},
					{ID: "debt_waiver", Name: "债务豁免协议", Description: "债权人部分或全部免除债务人债务的协议", ParentID: "debt"},
				},
			},
			{
				ID:          "termination",
				Name:        "终止/解除协议",
				Description: "终止或解除已有法律关系的协议",
				ParentID:    "agreement",
				Children: []*template_entity.TemplateDocType{
					{ID: "contract_termination", Name: "合同终止协议", Description: "终止合同关系的协议", ParentID: "termination"},
					{ID: "labor_termination", Name: "劳动关系解除协议", Description: "终止劳动合同关系的协议", ParentID: "termination"},
					{ID: "coop_termination", Name: "合作终止协议", Description: "终止合作关系的协议", ParentID: "termination"},
					{ID: "settlement_termination", Name: "和解协议终止协议", Description: "终止原和解协议的协议", ParentID: "termination"},
				},
			},
		},
	},
}

// 获取所有模板分类
func GetTemplateCategoriesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": documentCategoryTree,
	})
}

// 一级
func GetTopLevelCategoriesHandler(c *gin.Context) {
	// 创建一个只包含顶级分类的简化版本
	var topCategories []map[string]interface{}

	for _, category := range documentCategoryTree {
		// 只提取顶级分类的基本信息
		topCategory := map[string]interface{}{
			"id":          category.ID,
			"name":        category.Name,
			"description": category.Description,
		}

		topCategories = append(topCategories, topCategory)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": topCategories,
	})
}

// 二级
func GetTemplateSubCategoriesHandler(c *gin.Context) {
	categoryID := c.Param("categoryId")

	for _, category := range documentCategoryTree {
		if category.ID == categoryID {
			c.JSON(http.StatusOK, gin.H{
				"data": category.Children,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
}

// 三级
func GetTemplateDocTypesHandler(c *gin.Context) {
	categoryID := c.Param("categoryId")
	subCategoryID := c.Param("subCategoryId")

	for _, category := range documentCategoryTree {
		if category.ID == categoryID {
			for _, subCategory := range category.Children {
				if subCategory.ID == subCategoryID {
					c.JSON(http.StatusOK, gin.H{
						"data": subCategory.Children,
					})
					return
				}
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "子分类不存在"})
}

// 创建
func CreateTemplateHandler(c *gin.Context) {
	var req template_dto.CreateTemplatereq

	// 解析请求 JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 从查询参数获取子分类和文档类型
	subCategoryID := c.Query("subCategoryId")
	docTypeID := c.Query("docTypeId")

	// 调用服务创建模板
	template, err := CreateTemplate(req, subCategoryID, docTypeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 记录保存到数据库的内容
	fmt.Printf("成功创建模板 %s，保存的内容: %s\n", template.Name, string(template.Content))

	// 返回成功响应
	c.JSON(http.StatusOK, template)
}

// 根据分类获取模板列表
func GetTemplatesByCategoryHandler(c *gin.Context) {
	categoryID := c.Param("categoryId")
	subCategoryID := c.Query("subCategoryId")
	docTypeID := c.Query("docTypeId")

	var templates []template_entity.LegalTemplate
	query := dbs.DB.Where("type = ?", categoryID)

	if err := query.Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 如果需要按子分类或文档类型过滤
	if subCategoryID != "" || docTypeID != "" {
		var filteredTemplates []template_entity.LegalTemplate

		for _, tmpl := range templates {
			var content template_entity.TemplateContent
			if err := json.Unmarshal(tmpl.Content, &content); err != nil {
				continue
			}

			// 检查metadata
			metadata := content.Metadata

			// 过滤子分类
			if subCategoryID != "" && metadata.SubCategoryID != subCategoryID {
				continue
			}

			// 过滤文档类型
			if docTypeID != "" && metadata.DocTypeID != docTypeID {
				continue
			}

			filteredTemplates = append(filteredTemplates, tmpl)
		}

		templates = filteredTemplates
	}

	c.JSON(http.StatusOK, gin.H{
		"data": templates,
	})
}

// 获取模板详情
func GetTemplateDetailHandler(c *gin.Context) {
	templateID := c.Param("id")

	var template template_entity.LegalTemplate
	if err := dbs.DB.First(&template, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}

	var content template_entity.TemplateContent

	if err := json.Unmarshal(template.Content, &content); err != nil {
		// 解析失败，仍然返回原始template
		fmt.Printf("模板内容解析失败: %v, 内容: %s\n", err, string(template.Content))
		c.JSON(http.StatusOK, gin.H{
			"template":   template,
			"markdown":   "",
			"parseError": "无法解析模板内容为Markdown格式: " + err.Error(),
		})
		return
	}

	// 转换为Markdown
	markdownContent := convertTemplateToMarkdown(content, template.Type)

	// 返回完整template和markdown内容
	c.JSON(http.StatusOK, gin.H{
		"template": template,        // 完整的template对象
		"markdown": markdownContent, // markdown格式的内容
	})
}

// 将模板内容转换为Markdown格式
func convertTemplateToMarkdown(content template_entity.TemplateContent, templateType string) string {
	var markdownBuilder strings.Builder

	// 标题
	markdownBuilder.WriteString("# " + content.Title + "\n\n")

	// 引言
	if content.Introduction != "" {
		markdownBuilder.WriteString("## 引言\n\n" + content.Introduction + "\n\n")
	}

	// 当事人信息
	markdownBuilder.WriteString("## 当事人信息\n\n")
	for _, party := range content.Parties {
		markdownBuilder.WriteString("### " + party.Role + "\n\n")
		markdownBuilder.WriteString("- 名称: " + party.Name + "\n")
		markdownBuilder.WriteString("- 地址: " + party.Address + "\n")
		if party.ContactInfo != "" {
			markdownBuilder.WriteString("- 联系方式: " + party.ContactInfo + "\n")
		}
		markdownBuilder.WriteString("\n")
	}

	// 根据模板类型生成不同的内容部分
	switch templateType {
	case "contract", "agreement":
		// 条款
		markdownBuilder.WriteString("## 合同条款\n\n")
		for _, clause := range content.Clauses {
			markdownBuilder.WriteString("### " + clause.Title + "\n\n")
			if clause.Description != "" {
				markdownBuilder.WriteString("**" + clause.Description + "**\n\n")
			}
			markdownBuilder.WriteString(clause.Content + "\n\n")
		}

		// 附加条款
		if content.AdditionalTerms != nil {
			markdownBuilder.WriteString("## 附加条款\n\n")
			// 由于 AdditionalTerms 是 interface{} 类型，需要判断具体类型
			switch terms := content.AdditionalTerms.(type) {
			case []interface{}:
				for i, term := range terms {
					markdownBuilder.WriteString(fmt.Sprintf("%d. %v\n\n", i+1, term))
				}
			case string:
				markdownBuilder.WriteString(terms + "\n\n")
			default:
				// 尝试将其转为 JSON 字符串
				if termsJSON, err := json.Marshal(terms); err == nil {
					markdownBuilder.WriteString(string(termsJSON) + "\n\n")
				}
			}
		}
	case "lawsuit":
		// 诉讼请求
		markdownBuilder.WriteString("## 诉讼请求\n\n")
		for i, request := range content.Requests {
			markdownBuilder.WriteString(fmt.Sprintf("### 请求 %d: %s\n\n", i+1, request.Type))
			markdownBuilder.WriteString(request.Description + "\n\n")
		}

		// 证据
		markdownBuilder.WriteString("## 证据材料\n\n")
		for i, evidence := range content.Evidence {
			markdownBuilder.WriteString(fmt.Sprintf("### 证据 %d: %s\n\n", i+1, evidence.Name))
			markdownBuilder.WriteString(evidence.Description + "\n\n")
		}
	}

	// 法律声明
	if content.LegalStatement != "" {
		markdownBuilder.WriteString("## 法律声明\n\n" + content.LegalStatement + "\n\n")
	}

	// 签署部分
	if content.Signature != nil {
		markdownBuilder.WriteString("## 签署\n\n")
		// 由于 Signature 是 interface{} 类型，需要判断具体类型
		switch sig := content.Signature.(type) {
		case string:
			markdownBuilder.WriteString(sig + "\n\n")
		case template_entity.Signature:
			markdownBuilder.WriteString(fmt.Sprintf("签署方式: %s\n\n", sig.Method))
			markdownBuilder.WriteString(fmt.Sprintf("日期格式: %s\n\n", sig.DateFormat))
		case map[string]interface{}:
			if method, ok := sig["method"].(string); ok {
				markdownBuilder.WriteString(fmt.Sprintf("签署方式: %s\n\n", method))
			}
			if dateFormat, ok := sig["dateFormat"].(string); ok {
				markdownBuilder.WriteString(fmt.Sprintf("日期格式: %s\n\n", dateFormat))
			}
		default:
			// 尝试将其转为 JSON 字符串
			if sigJSON, err := json.Marshal(sig); err == nil {
				markdownBuilder.WriteString(string(sigJSON) + "\n\n")
			}
		}
	}

	// 元数据信息
	markdownBuilder.WriteString("---\n\n")
	markdownBuilder.WriteString(fmt.Sprintf("**文档类型**: %s - %s - %s\n\n",
		content.Metadata.CategoryName,
		content.Metadata.SubCategoryName,
		content.Metadata.DocTypeName))

	return markdownBuilder.String()
}

// 初始化系统模板
func InitializeTemplatesHandler(c *gin.Context) {
	count, err := InitializeTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("成功初始化 %d 个模板", count),
	})
}

// 初始化系统默认法律模板
func InitializeTemplates() (int, error) {
	successCount := 0

	// 定义不同类型模板的环境
	environmentMap := map[string]map[string]string{
		"contract": {
			"commercial":   "商业交易环境",
			"realestate":   "房地产交易环境",
			"financial":    "金融业务环境",
			"intellectual": "知识产权保护环境",
			"labor":        "企业用工环境",
		},
		"lawsuit": {
			"civil":          "民事诉讼环境",
			"criminal":       "刑事诉讼环境",
			"administrative": "行政诉讼环境",
			"arbitration":    "商事仲裁环境",
		},
		"agreement": {
			"settlement":  "争议解决环境",
			"business":    "商业合作环境",
			"family":      "家事调解环境",
			"debt":        "债权债务处理环境",
			"termination": "合同解除环境",
		},
	}

	// 遍历整个分类树
	for _, category := range documentCategoryTree {
		categoryID := category.ID
		for _, subCategory := range category.Children {
			subCategoryID := subCategory.ID

			// 获取特定环境，如果未定义则使用"通用环境"
			environmentSetting := "通用环境"
			if envMap, exists := environmentMap[categoryID]; exists {
				if env, exists := envMap[subCategoryID]; exists {
					environmentSetting = env
				}
			}

			for _, docType := range subCategory.Children {
				// 检查模板是否已存在
				var count int64
				dbs.DB.Model(&template_entity.LegalTemplate{}).
					Where("type = ? AND name LIKE ?", category.ID, fmt.Sprintf("%%%s标准模板%%", docType.Name)).
					Count(&count)

				if count > 0 {
					continue // 跳过已存在的模板
				}

				// 创建基础模板
				templateReq := template_dto.CreateTemplatereq{
					Name:        fmt.Sprintf("%s标准模板", docType.Name),
					Type:        category.ID,
					Environment: environmentSetting,
				}

				// 使用现有函数创建模板
				_, err := CreateTemplate(templateReq, subCategory.ID, docType.ID)
				if err != nil {
					return successCount, fmt.Errorf("初始化模板 %s 失败: %v", docType.Name, err)
				}

				successCount++
				fmt.Printf("成功创建模板: %s\n", docType.Name)
			}
		}
	}

	return successCount, nil
}

// 创建新的法律文档模板
func CreateTemplate(req template_dto.CreateTemplatereq, subCategoryID, docTypeID string) (*template_entity.LegalTemplate, error) {
	// 确保 MoonClient 已初始化
	if client.MoonClient == nil || client.MoonClient.GetClient() == nil {
		return nil, errors.New("Moonshot 客户端未初始化")
	}

	// 验证分类是否存在
	valid := validateTemplateCategory(req.Type, subCategoryID, docTypeID)
	if !valid {
		return nil, errors.New("无效的模板分类")
	}

	// 获取分类名称
	categoryName, subCategoryName, docTypeName := getCategoryNames(req.Type, subCategoryID, docTypeID)

	// 创建元数据
	metadata := template_entity.TemplateMetadata{
		CategoryID:      req.Type,
		CategoryName:    categoryName,
		SubCategoryID:   subCategoryID,
		SubCategoryName: subCategoryName,
		DocTypeID:       docTypeID,
		DocTypeName:     docTypeName,
	}

	// 调用 AI 生成模板内容
	content, err := generateTemplateFromMoonshot(req.Type, req.Name, req.Environment, categoryName, subCategoryName, docTypeName, metadata)
	if err != nil {
		return nil, err
	}

	// 创建新的模板
	newTemplate := template_entity.LegalTemplate{
		Name:        req.Name,
		Type:        req.Type,
		Environment: req.Environment,
		Content:     content,
	}

	// 存入数据库
	if err := dbs.DB.Create(&newTemplate).Error; err != nil {
		return nil, err
	}

	return &newTemplate, nil
}

// 验证模板分类是否有效
func validateTemplateCategory(categoryID, subCategoryID, docTypeID string) bool {
	for _, category := range documentCategoryTree {
		if category.ID == categoryID {
			if subCategoryID == "" {
				return true
			}

			for _, subCategory := range category.Children {
				if subCategory.ID == subCategoryID {
					if docTypeID == "" {
						return true
					}

					for _, docType := range subCategory.Children {
						if docType.ID == docTypeID {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// 获取分类的显示名称
func getCategoryNames(categoryID, subCategoryID, docTypeID string) (string, string, string) {
	var categoryName, subCategoryName, docTypeName string

	for _, category := range documentCategoryTree {
		if category.ID == categoryID {
			categoryName = category.Name

			for _, subCategory := range category.Children {
				if subCategory.ID == subCategoryID {
					subCategoryName = subCategory.Name

					for _, docType := range subCategory.Children {
						if docType.ID == docTypeID {
							docTypeName = docType.Name
							break
						}
					}
					break
				}
			}
			break
		}
	}

	return categoryName, subCategoryName, docTypeName
}

// 转换AI返回的JSON结构为符合TemplateContent的结构
func normalizeTemplateContent(rawContent map[string]interface{}, templateType string, metadata template_entity.TemplateMetadata) map[string]interface{} {
	result := make(map[string]interface{})

	// 复制基本字段
	fields := []string{"title", "introduction", "signature", "legalStatement"}
	for _, field := range fields {
		if val, exists := rawContent[field]; exists {
			result[field] = val
		} else {
			result[field] = ""
		}
	}

	// 处理metadata
	result["metadata"] = metadata

	// 处理parties (确保格式正确)
	if parties, exists := rawContent["parties"]; exists {
		result["parties"] = parties
	} else {
		// 添加默认当事人
		switch templateType {
		case "contract", "agreement":
			result["parties"] = []map[string]string{
				{"role": "甲方", "name": "{甲方名称}", "address": "{甲方地址}", "contactInfo": "{甲方联系方式}"},
				{"role": "乙方", "name": "{乙方名称}", "address": "{乙方地址}", "contactInfo": "{乙方联系方式}"},
			}
		case "lawsuit":
			result["parties"] = []map[string]string{
				{"role": "原告", "name": "{原告名称}", "address": "{原告地址}", "contactInfo": "{原告联系方式}"},
				{"role": "被告", "name": "{被告名称}", "address": "{被告地址}", "contactInfo": "{被告联系方式}"},
			}
		}
	}

	// 处理各类型特有字段
	switch templateType {
	case "contract", "agreement":
		if clauses, exists := rawContent["clauses"]; exists {
			result["clauses"] = clauses
		} else {
			result["clauses"] = []map[string]string{
				{"title": "第一条", "description": "基本条款", "details": "详细内容..."},
			}
		}

		if additionalTerms, exists := rawContent["additionalTerms"]; exists {
			result["additionalTerms"] = additionalTerms
		}

	case "lawsuit":
		if requests, exists := rawContent["requests"]; exists {
			result["requests"] = requests
		} else {
			result["requests"] = []map[string]string{
				{"type": "主要请求", "description": "请求内容..."},
			}
		}

		if evidence, exists := rawContent["evidence"]; exists {
			result["evidence"] = evidence
		} else {
			result["evidence"] = []map[string]string{
				{"name": "证据一", "description": "证据描述...", "filePath": ""},
			}
		}
	}

	return result
}

// 生成模板内容
func generateTemplateFromMoonshot(templateType, name, environment, categoryName, subCategoryName, docTypeName string, metadata template_entity.TemplateMetadata) (json.RawMessage, error) {
	// 获取 Moonshot 客户端
	chatClient := client.MoonClient.GetClient().Chat()

	// 构造 Prompt
	var prompt string
	switch templateType {
	case "contract":
		prompt = fmt.Sprintf(`
你是一个专业的法律 AI 助手，擅长起草合同文书。
请基于以下信息生成一个格式正确的合同：
- **合同名称**: %s
- **合同类型**: %s - %s - %s
- **适用环境**: %s
- **结构要求**：
  - 标题
  - 引言
  - 合同双方（甲方、乙方）
  - 合同主要条款
  - 签署条款
  - 法律声明

请生成一个符合TemplateContent结构的JSON，包含以下字段：
- title: 文档标题
- introduction: 引言部分
- parties: 当事人列表（角色包括"甲方"、"乙方"等，名称使用通用占位符如"{公司名称}"）
- clauses: 条款列表（至少包含5-8个条款，每个条款有标题、描述和详细内容）
- signature: 签署部分（包含签署方式和日期格式）
- legalStatement: 法律声明
- additionalTerms: 附加条款（可选）

请使用标准规范的法律语言，内容要专业、全面且具有可操作性。请确保生成的JSON格式正确，所有字段名称符合驼峰命名法。`, name, categoryName, subCategoryName, docTypeName, environment)

	case "lawsuit":
		prompt = fmt.Sprintf(`
你是一个专业的法律 AI 助手，擅长撰写法律诉讼文书。
请基于以下信息生成一个格式正确的法律诉讼文书：
- **文书名称**: %s
- **文书类型**: %s - %s - %s
- **适用环境**: %s
- **结构要求**：
  - 标题
  - 引言
  - 诉讼双方（原告、被告）
  - 诉讼请求
  - 证据
  - 法律依据
  - 签署信息

请生成一个符合TemplateContent结构的JSON，包含以下字段：
- title: 文书标题
- introduction: 引言部分
- parties: 当事人列表（角色包括"原告"、"被告"等，名称使用通用占位符如"{当事人姓名}"）
- requests: 诉讼请求列表（每个请求包含类型和详细描述）
- evidence: 证据列表（包含证据名称和描述）
- signature: 签署部分
- legalStatement: 法律依据说明

请使用标准规范的法律语言，内容要专业、全面且符合中国诉讼法规定。请确保生成的JSON格式正确，所有字段名称符合驼峰命名法。`, name, categoryName, subCategoryName, docTypeName, environment)

	case "agreement":
		prompt = fmt.Sprintf(`
你是一个专业的法律 AI 助手，擅长起草协议文书。
请基于以下信息生成一个格式正确的协议：
- **协议名称**: %s
- **协议类型**: %s - %s - %s
- **适用环境**: %s
- **结构要求**：
  - 标题
  - 引言
  - 协议双方（甲方、乙方）
  - 协议条款
  - 签署信息
  - 法律声明

请生成一个符合TemplateContent结构的JSON，包含以下字段：
- title: 协议标题
- introduction: 引言部分
- parties: 当事人列表（角色包括"甲方"、"乙方"等，名称使用通用占位符如"{公司名称}"或"{个人姓名}"）
- clauses: 条款列表（各项协议条款，包含标题、描述和详细内容）
- signature: 签署部分
- legalStatement: 法律声明
- additionalTerms: 附加条款（可选）

请使用标准规范的法律语言，内容要专业、全面且具有可操作性。请确保生成的JSON格式正确，所有字段名称符合驼峰命名法。`, name, categoryName, subCategoryName, docTypeName, environment)

	default:
		return nil, fmt.Errorf("未知的文书类型: %s", templateType)
	}

	// 构建初始模板内容结构
	templateContent := template_entity.TemplateContent{
		Metadata: metadata,
	}

	// 根据模板类型预设一些基本结构
	switch templateType {
	case "contract":
		templateContent.Parties = []template_entity.Party{
			{Role: "甲方", Name: "{甲方名称}", Address: "{甲方地址}", ContactInfo: "{甲方联系方式}"},
			{Role: "乙方", Name: "{乙方名称}", Address: "{乙方地址}", ContactInfo: "{乙方联系方式}"},
		}
	case "lawsuit":
		templateContent.Parties = []template_entity.Party{
			{Role: "原告", Name: "{原告名称}", Address: "{原告地址}", ContactInfo: "{原告联系方式}"},
			{Role: "被告", Name: "{被告名称}", Address: "{被告地址}", ContactInfo: "{被告联系方式}"},
		}
	case "agreement":
		templateContent.Parties = []template_entity.Party{
			{Role: "甲方", Name: "{甲方名称}", Address: "{甲方地址}", ContactInfo: "{甲方联系方式}"},
			{Role: "乙方", Name: "{乙方名称}", Address: "{乙方地址}", ContactInfo: "{乙方联系方式}"},
		}
	}

	// 生成请求
	chatReq := moonshot.ChatCompletionsRequest{
		Model: moonshot.ModelMoonshotV1Auto,
		Messages: []*moonshot.ChatCompletionsMessage{
			{Role: moonshot.RoleSystem, Content: "你是一个专业的法律 AI 助手，擅长生成法律文档模板。请严格按照JSON格式输出。"},
			{Role: moonshot.RoleUser, Content: prompt},
		},
		MaxTokens:   4096,
		Temperature: 0.2,
	}

	// 发送请求
	resp, err := chatClient.Completions(context.Background(), &chatReq)
	if err != nil {
		return nil, fmt.Errorf("Moonshot API 调用失败: %v", err)
	}

	// 解析 AI 响应
	if len(resp.Choices) == 0 {
		return nil, errors.New("Moonshot 没有返回任何结果")
	}

	// 解析返回的 JSON
	aiResponse := resp.Choices[0].Message.Content

	// 从 AI 返回的文本中提取 JSON 部分
	var jsonContent map[string]interface{}

	// 尝试多种方式解析JSON
	err = json.Unmarshal([]byte(aiResponse), &jsonContent)
	if err != nil {
		// 尝试查找 JSON 部分
		jsonStart := strings.Index(aiResponse, "{")
		jsonEnd := strings.LastIndex(aiResponse, "}")

		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonText := aiResponse[jsonStart : jsonEnd+1]
			err = json.Unmarshal([]byte(jsonText), &jsonContent)
			if err != nil {
				// 记录原始响应以便调试
				fmt.Printf("AI 原始响应: %s\n", aiResponse)
				fmt.Printf("尝试提取的JSON: %s\n", jsonText)
				return nil, fmt.Errorf("无法解析 AI 返回的 JSON: %v", err)
			}
		} else {
			fmt.Printf("AI 原始响应: %s\n", aiResponse)
			return nil, fmt.Errorf("AI 返回内容不是有效的 JSON 格式")
		}
	}

	// 规范化模板内容
	jsonContent = normalizeTemplateContent(jsonContent, templateType, metadata)

	// 转换为 JSON
	finalJSON, err := json.Marshal(jsonContent)
	if err != nil {
		return nil, fmt.Errorf("转换最终 JSON 失败: %v", err)
	}

	// 验证生成的JSON能否被正确解析为TemplateContent结构
	var testContent template_entity.TemplateContent
	if err := json.Unmarshal(finalJSON, &testContent); err != nil {
		fmt.Printf("警告: 生成的JSON不符合TemplateContent结构: %v\n", err)
		// 记录但不中断，让函数继续返回JSON
	}

	return finalJSON, nil
}

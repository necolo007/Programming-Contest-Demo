package prompt

import (
	"Programming-Demo/internal/app/ai/ai_dto"
	"Programming-Demo/pkg/utils/ai"
	"fmt"
	"strconv"
	"strings"
)

func BuildLegalDocPrompt(req ai_dto.GenerateLegalDocReq) string {
	prompt := "请帮我生成一份专业的法律文件，要求如下：\n"
	prompt += "1. 文件类型：" + req.DocType + "\n"
	prompt += "2. 文件标题：" + req.Title + "\n"

	// 添加相关方信息
	prompt += "3. 相关方信息：\n"
	for _, party := range req.Parties {
		prompt += "   - " + party.Type + "：" + party.Name + "\n"
		prompt += "     详细信息：" + party.Details + "\n"
	}

	// 添加合同主要内容
	prompt += "4. 合同基本信息：\n"
	prompt += "   - 合同标的：" + req.Content.Subject + "\n"
	prompt += "   - 合同目的：" + req.Content.Purpose + "\n"
	prompt += "   - 签订地点：" + req.Content.Location + "\n"
	prompt += "   - 签订日期：" + req.Content.SignDate + "\n"

	prompt += "5. 权利义务：\n"
	prompt += "   - 权利内容：" + req.Content.Rights + "\n"
	prompt += "   - 义务内容：" + req.Content.Obligations + "\n"

	prompt += "6. 履行相关：\n"
	prompt += "   - 开始日期：" + req.Content.StartDate + "\n"
	prompt += "   - 结束日期：" + req.Content.EndDate + "\n"
	prompt += "   - 履行方式：" + req.Content.Performance + "\n"

	prompt += "7. 价格和支付：\n"
	prompt += "   - 价格/报酬：" + req.Content.Price + "\n"
	prompt += "   - 支付方式：" + req.Content.Payment + "\n"

	prompt += "8. 违约和争议解决：\n"
	prompt += "   - 违约责任：" + req.Content.Breach + "\n"
	prompt += "   - 争议解决：" + req.Content.Dispute + "\n"

	prompt += "9. 其他重要条款：\n"
	prompt += "   - 保密条款：" + req.Content.Confidential + "\n"
	prompt += "   - 不可抗力：" + req.Content.Force + "\n"
	prompt += "   - 终止条件：" + req.Content.Termination + "\n"
	prompt += "   - 补充条款：" + req.Content.Additional + "\n"

	if req.Additional != "" {
		prompt += "10. 特殊要求：" + req.Additional + "\n"
	}

	prompt += "\n请按照以下要求生成内容：\n"
	prompt += "1. 使用规范的法律文书格式\n"
	prompt += "2. 确保条款的完整性和专业性\n"
	prompt += "3. 使用清晰的条款编号和层次结构\n"
	prompt += "4. 语言表述准确、严谨\n"

	return prompt
}

// BuildComplaintPrompt 构建起诉状提示词
func BuildComplaintPrompt(req ai_dto.GenerateComplaintReq) string {
	prompt := "请帮我生成一份专业的起诉状，要求如下：\n"

	prompt += "1. 基本信息：\n"
	prompt += "   - 受理法院：" + req.Content.Court + "\n"

	prompt += "2. 当事人信息：\n"
	prompt += "   - 原告信息：\n"
	prompt += "     姓名：" + req.Content.Plaintiff.Name + "\n"
	prompt += "     详细信息：" + req.Content.Plaintiff.Details + "\n"
	prompt += "   - 被告信息：\n"
	prompt += "     姓名：" + req.Content.Defendant.Name + "\n"
	prompt += "     详细信息：" + req.Content.Defendant.Details + "\n"

	prompt += "3. 诉讼请求：\n"
	for i, claim := range req.Content.Claims {
		prompt += "   " + strconv.Itoa(i+'1') + ". " + claim + "\n"
	}

	prompt += "4. 事实与理由：\n" + req.Content.Facts + "\n"

	prompt += "5. 证据列表：\n"
	for i, evidence := range req.Content.Evidence {
		prompt += "   " + strconv.Itoa(i+'1') + ". " + evidence + "\n"
	}

	prompt += "6. 法律依据：\n"
	for i, law := range req.Content.LawBasis {
		prompt += "   " + strconv.Itoa(i+'1') + ". " + law + "\n"
	}

	prompt += "7. 附件清单：\n"
	for i, attachment := range req.Content.Attachments {
		prompt += "   " + strconv.Itoa(i+'1') + ". " + attachment + "\n"
	}

	prompt += "\n请按照以下要求生成起诉状：\n"
	prompt += "1. 使用规范的起诉状格式\n"
	prompt += "2. 确保文书格式符合法院要求\n"
	prompt += "3. 语言表述准确、严谨\n"
	prompt += "4. 包含必要的落款和日期\n"

	return prompt
}

// BuildLegalOpinionPrompt 构建法律意见书提示词
func BuildLegalOpinionPrompt(req ai_dto.GenerateLegalOpinionReq) string {
	prompt := "请帮我生成一份专业的法律意见书，要求如下：\n"

	prompt += "1. 案件背景：\n" + req.Content.Background + "\n"

	prompt += "2. 需要解决的法律问题：\n"
	for i, issue := range req.Content.Issues {
		prompt += "   " + strconv.Itoa(i+'1') + ". " + issue + "\n"
	}

	prompt += "3. 法律分析：\n" + req.Content.Analysis + "\n"

	prompt += "4. 法律风险：\n"
	for i, risk := range req.Content.Risks {
		prompt += "   " + strconv.Itoa(i+'1') + ". " + risk + "\n"
	}

	prompt += "5. 法律建议：\n"
	for i, suggestion := range req.Content.Suggestions {
		prompt += "   " + strconv.Itoa(i+'1') + ". " + suggestion + "\n"
	}

	prompt += "6. 法律依据：\n"
	for i, reference := range req.Content.References {
		prompt += "   " + strconv.Itoa(i+'1') + ". " + reference + "\n"
	}

	prompt += "\n请按照以下要求生成法律意见书：\n"
	prompt += "1. 使用规范的法律意见书格式\n"
	prompt += "2. 分析论述要客观、专业\n"
	prompt += "3. 建议要具体、可操作\n"
	prompt += "4. 引用法律依据要准确\n"

	return prompt
}

// BuildLegalAnalysisPrompt 构建法律文件分析提示词
func BuildLegalAnalysisPrompt(content string) string {
	prompt := "请对以下法律文件进行专业分析，要求如下：\n\n"

	prompt += "1. 文件基本信息提取：\n"
	prompt += "   - 文件类型和性质\n"
	prompt += "   - 文件签署日期和生效时间\n"
	prompt += "   - 涉及的主体方\n"
	prompt += "   - 文件的主要目的\n\n"

	prompt += "2. 关键信息提取：\n"
	prompt += "   - 重要条款内容\n"
	prompt += "   - 关键日期节点\n"
	prompt += "   - 金额和支付条件\n"
	prompt += "   - 权利义务关系\n\n"

	prompt += "3. 法律术语解释：\n"
	prompt += "   - 识别文件中的专业法律术语\n"
	prompt += "   - 提供通俗易懂的解释\n"
	prompt += "   - 说明术语在文件中的具体含义和作用\n\n"

	prompt += "4. 条款分类分析：\n"
	prompt += "   - 主要条款分类（如基本条款、履行条款、违约条款等）\n"
	prompt += "   - 各类条款的主要内容概述\n"
	prompt += "   - 条款之间的关联性分析\n\n"

	prompt += "5. 风险提示：\n"
	prompt += "   - 潜在的法律风险点\n"
	prompt += "   - 条款中的不明确或争议之处\n"
	prompt += "   - 建议重点关注的内容\n\n"

	prompt += "待分析的法律文件内容如下：\n"
	prompt += content + "\n\n"

	prompt += "请按照以下格式输出分析结果：\n"
	prompt += "1. 使用清晰的层级结构\n"
	prompt += "2. 重要内容需要突出显示\n"
	prompt += "3. 专业术语解释要通俗易懂\n"
	prompt += "4. 风险提示要具体明确\n"

	return prompt
}

// BuildRAGPrompt 构建RAG提示
func BuildRAGPrompt(query string) (string, []ai.Document) {
	var sb strings.Builder

	// 添加指令
	sb.WriteString("请根据以下参考信息回答问题。如果参考信息不足以回答问题，请直接说明无法从参考信息中找到答案。\n\n")

	// 检索相关文档
	docs, err := ai.SearchSimilarDocumentsWithParam(query, 30)
	if err != nil {
		return "检索相关文档失败: " + err.Error(), docs
	}

	// 添加参考信息
	sb.WriteString("参考信息:\n")
	for i, ctx := range docs {
		sb.WriteString(fmt.Sprintf("[%d] 相关度:%.2f\n%s\n\n", i+1, ctx.Score, ctx.Content))
	}

	// 添加用户问题
	sb.WriteString("问题: " + query + "\n\n")
	sb.WriteString("请根据以上参考信息回答问题:")

	return sb.String(), docs
}

package prompt

import (
	"Programming-Demo/internal/app/ai/ai_dto"
	"strconv"
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

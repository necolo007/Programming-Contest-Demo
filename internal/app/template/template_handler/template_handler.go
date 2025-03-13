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
)

func CreateTemplateHandler(c *gin.Context) {
	var req template_dto.CreateTemplatereq

	// 解析请求 JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 直接调用 service 层
	template, err := CreateTemplate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, template)
}

func CreateTemplate(req template_dto.CreateTemplatereq) (*template_entity.LegalTemplate, error) {
	// 确保 MoonClient 已初始化
	if client.MoonClient == nil || client.MoonClient.GetClient() == nil {
		return nil, errors.New("Moonshot 客户端未初始化")
	}

	// 调用 AI 生成模板内容
	aiContent, err := generateTemplateFromMoonshot(req.Type, req.Name, req.Environment)
	if err != nil {
		return nil, err
	}

	// 创建新的模板
	newTemplate := template_entity.LegalTemplate{
		Name:        req.Name,
		Type:        req.Type,
		Environment: req.Environment,
		Content:     aiContent, // 直接存 JSON
	}

	// 存入数据库
	if err := dbs.DB.Create(&newTemplate).Error; err != nil {
		return nil, err
	}

	return &newTemplate, nil
}

func generateTemplateFromMoonshot(templateType, name, environment string) (json.RawMessage, error) {
	// 获取 Moonshot 客户端
	chatClient := client.MoonClient.GetClient().Chat()

	// 构造 Prompt
	var prompt string
	switch templateType {
	case "Contract":
		prompt = fmt.Sprintf(`
你是一个专业的法律 AI 助手，擅长起草合同文书。
请基于以下信息生成一个格式正确的合同：
- **合同名称**: %s
- **适用环境**: %s
- **结构要求**：
  - **标题**
  - **引言**
  - **合同双方（甲方、乙方）**
  - **合同主要条款**
  - **签署条款**
  - **法律声明**
请严格按照 JSON 形式输出。`, name, environment)

	case "Lawsuit":
		prompt = fmt.Sprintf(`
你是一个专业的法律 AI 助手，擅长撰写法律诉讼文书。
请基于以下信息生成一个格式正确的法律诉讼文书：
- **文书类型**: %s
- **适用环境**: %s
- **结构要求**：
  - **标题**
  - **引言**
  - **诉讼双方（原告、被告）**
  - **诉讼请求**
  - **证据**
  - **法律依据**
  - **签署信息**
请严格按照 JSON 形式输出。`, name, environment)

	case "Agreement":
		prompt = fmt.Sprintf(`
你是一个专业的法律 AI 助手，擅长起草协议文书。
请基于以下信息生成一个格式正确的协议：
- **协议名称**: %s
- **适用环境**: %s
- **结构要求**：
  - **标题**
  - **引言**
  - **协议双方（甲方、乙方）**
  - **协议条款**
  - **签署信息**
  - **法律声明**
请严格按照 JSON 形式输出。`, name, environment)

	default:
		return nil, fmt.Errorf("未知的文书类型: %s", templateType)
	}

	// 生成请求
	chatReq := moonshot.ChatCompletionsRequest{
		Model: moonshot.ModelMoonshotV1Auto,
		Messages: []*moonshot.ChatCompletionsMessage{
			{Role: moonshot.RoleSystem, Content: "你是一个专业的法律 AI 助手，擅长生成法律文档模板"},
			{Role: moonshot.RoleUser, Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.3,
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

	// 直接返回 JSON
	return json.RawMessage(resp.Choices[0].Message.Content), nil
}

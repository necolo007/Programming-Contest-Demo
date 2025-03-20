package ai_handler

import (
	"Programming-Demo/core/client"
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/core/libx"
	"Programming-Demo/internal/app/File/file_entity"
	"Programming-Demo/internal/app/ai/ai_dto"
	"Programming-Demo/internal/app/ai/ai_entity"
	"Programming-Demo/pkg/utils/ai"
	"Programming-Demo/pkg/utils/deepseek"
	"Programming-Demo/pkg/utils/prompt"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/northes/go-moonshot"
)

func PingMoonshot(c *gin.Context) {
	resp, err := client.MoonClient.GetClient().Chat().CompletionsStream(context.Background(), &moonshot.ChatCompletionsRequest{
		Model: moonshot.ModelMoonshotV18K,
		Messages: []*moonshot.ChatCompletionsMessage{
			{
				Role:    moonshot.RoleUser,
				Content: "你好，请问1+1等于多少？",
			},
		},
		Temperature: 0.3,
		Stream:      true,
	})
	var message string
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "message": "moonshot chat failed"})
		return
	} else {
		for receive := range resp.Receive() {
			msg, err1 := receive.GetMessage()
			if err1 != nil {
				if errors.Is(err1, io.EOF) {
					c.JSON(200, gin.H{"message": message})
				}
				return
			}
			message = message + msg.Content
		}
	}
}

func AnalyzeFile(c *gin.Context) {
	var req ai_dto.AnalyzeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	var existingFile file_entity.File
	if err := dbs.DB.Where("filename = ?", req.Name).First(&existingFile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件不存在"})
		return
	}
	// 检查文件是否存在
	fileInfo, err := os.Stat(existingFile.Filepath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件不存在",
			"error":   err.Error(),
		})
		return
	}

	// 检查文件大小（限制为10MB）
	if fileInfo.Size() > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件大小超过限制",
			"error":   "文件大小不能超过10MB",
		})
		return
	}

	content, err := os.ReadFile(existingFile.Filepath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "文件读取错误"})
		return
	}
	Resp, code := ai.GetAIResp(prompt.BuildLegalAnalysisPrompt(string(content)))
	if code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用ai接口失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "message": Resp})
}

func GenerateLegalDocument(c *gin.Context) {
	var req ai_dto.GenerateLegalDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}
	var Resp string
	var code int
	p := prompt.BuildLegalDocPrompt(req)
	// 选择不同的 AI 模型处理
	switch req.Model {
	case "moonshot":
		Resp, code = ai.GetAIResp(p)
	case "deepseek-chat", "deepseek-reasoner":
		Resp, code = deepseek.ChatWithDeepSeek(p, "POST", req.Model)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "模型错误"})
		return
	}
	if code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用ai接口失败", "error": Resp})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "message": Resp})
}

func GenerateLegalOpinion(c *gin.Context) {
	var req ai_dto.GenerateLegalOpinionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}
	var Resp string
	var code int
	p := prompt.BuildLegalOpinionPrompt(req)
	// 选择不同的 AI 模型处理
	switch req.Model {
	case "moonshot":
		Resp, code = ai.GetAIResp(p)
	case "deepseek-chat", "deepseek-reasoner":
		Resp, code = deepseek.ChatWithDeepSeek(p, "POST", req.Model)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "模型错误"})
		return
	}
	if code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用ai接口失败", "error": Resp})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "message": Resp})
}

func GenerateComplaint(c *gin.Context) {
	var req ai_dto.GenerateComplaintReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}
	var Resp string
	var code int
	p := prompt.BuildComplaintPrompt(req)
	// 选择不同的 AI 模型处理
	switch req.Model {
	case "moonshot":
		Resp, code = ai.GetAIResp(p)
	case "deepseek-chat", "deepseek-reasoner":
		Resp, code = deepseek.ChatWithDeepSeek(p, "POST", req.Model)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "模型错误"})
		return
	}
	if code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用ai接口失败", "error": Resp})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "message": Resp})
}

func ChatWithAi(c *gin.Context) {
	uid := libx.Uid(c)
	var req ai_dto.ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取历史对话记录
	var histories []ai_entity.ChatHistory
	dbs.DB.Where("user_id = ?", uid).
		Order("created_at desc").
		Limit(5).
		Find(&histories)

	// 构建对话上下文
	var messages []string
	for i := len(histories) - 1; i >= 0; i-- {
		messages = append(messages, histories[i].Content)
	}
	messages = append(messages, req.Content)

	// 将当前问题保存到历史记录
	userMessage := ai_entity.ChatHistory{
		UserID:  uid,
		Model:   req.Model,
		Theme:   req.Theme,
		Role:    "user",
		Content: req.Content,
	}
	dbs.DB.Create(&userMessage)

	var Resp string
	var code int

	// 选择不同的 AI 模型处理
	switch req.Model {
	case "moonshot":
		Resp, code = ai.GetAIResp(strings.Join(messages, "\n"))
	case "deepseek-chat", "deepseek-reasoner":
		Resp, code = deepseek.ChatWithDeepSeek(strings.Join(messages, "\n"), "POST", req.Model)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "模型错误"})
		return
	}

	if code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用ai接口失败", "error": Resp})
		return
	}

	// 保存 AI 回复到历史记录
	aiMessage := ai_entity.ChatHistory{
		UserID:  uid,
		Model:   req.Model,
		Theme:   req.Theme,
		Role:    "assistant",
		Content: Resp,
	}
	dbs.DB.Create(&aiMessage)

	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": Resp,
		"history": histories,
	})
}

func GetChatHistory(c *gin.Context) {
	uid := libx.Uid(c)
	theme := c.Query("theme")
	var histories []ai_entity.ChatHistory
	if err := dbs.DB.Where("user_id = ? and theme = ?", uid, theme).
		Order("created_at desc").
		Find(&histories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取历史记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    histories,
	})
}

func GetChatThemes(c *gin.Context) {
	uid := libx.Uid(c)
	var themes []string
	if err := dbs.DB.Model(&ai_entity.ChatHistory{}).
		Where("user_id = ?", uid).
		Select("distinct theme").
		Pluck("theme", &themes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取主题失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    themes,
	})
}

func SearchWithMoonshot(c *gin.Context) {
	// 从请求中获取查询参数，不再设置默认值
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索栏不能为空"})
		return
	}

	// 获取封装好的 Moonshot 客户端实例
	getClient := client.MoonClient.GetClient()

	// 创建请求结构体，设置必要的参数
	req := moonshot.ChatCompletionsRequest{
		Model: moonshot.ModelMoonshotV18K,
		Messages: []*moonshot.ChatCompletionsMessage{
			{
				Role:    moonshot.RoleSystem,
				Content: fmt.Sprintf("Search the web for: %s", query),
			},
		},
		Temperature: 0.3,
		Stream:      false,
		Tools: []*moonshot.ChatCompletionsTool{
			{
				Type: moonshot.ChatCompletionsToolTypeBuiltinFunction,
				Function: &moonshot.ChatCompletionsToolFunction{
					Name: "$web_search",
				},
			},
		},
	}

	// 发送请求并获取响应
	resp, err := getClient.Chat().Completions(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "moonshot chat failed"})
		return
	}

	// 检查是否需要调用工具
	if resp.Choices[0].FinishReason == "tool_calls" {
		// 获取工具调用参数
		toolCalls := resp.Choices[0].Message.ToolCalls
		for _, toolCall := range toolCalls {
			if toolCall.Function.Name == "$web_search" {
				// 将工具调用参数返回给模型
				toolResult := toolCall.Function.Arguments

				// 构造新的请求，包含工具调用结果
				newReq := moonshot.ChatCompletionsRequest{
					Model: moonshot.ModelMoonshotV18K,
					Messages: []*moonshot.ChatCompletionsMessage{
						{
							Role:    moonshot.RoleTool,
							Content: toolResult,
						},
					},
					Temperature: 0.3,
					Stream:      false,
				}

				// 发送新的请求
				newResp, err := getClient.Chat().Completions(context.Background(), &newReq)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "moonshot chat failed"})
					return
				}

				// 返回最终结果
				if newResp.Choices[0].Message.Content != "" {
					c.JSON(http.StatusOK, gin.H{"message": newResp.Choices[0].Message.Content})
					return
				}
			}
		}
	} else {
		// 如果不需要调用工具，直接返回结果
		if resp.Choices[0].Message.Content != "" {
			c.JSON(http.StatusOK, gin.H{"message": resp.Choices[0].Message.Content})
			return
		}
	}

	// 如果没有结果
	c.JSON(http.StatusOK, gin.H{"message": "No results found."})
}

func GenerateLegalOpinionBetter(c *gin.Context) {
	var req ai_dto.GenerateLegalOpinionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}
	p := prompt.BuildLegalOpinionPrompt(req)
	resp, code, docs := ai.GetAIRespMore(p)
	if code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用ai接口失败", "error": resp})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "message": resp, "docs": docs})
}

package ai_handler

import (
	"Programming-Demo/core/client"
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/internal/app/File/file_entity"
	"Programming-Demo/internal/app/ai/ai_dto"
	"Programming-Demo/pkg/utils/ai"
	"Programming-Demo/pkg/utils/deepseek"
	"Programming-Demo/pkg/utils/prompt"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/northes/go-moonshot"
	"io"
	"net/http"
	"os"
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
	Resp, code := ai.GetAIResp("请提取下列法律文件中的关键信息，并分析下述法律合同" + string(content))
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

package ai_handler

import (
	"Programming-Demo/core/client"
	"Programming-Demo/internal/app/ai/ai_dto"
	"Programming-Demo/pkg/utils/ai"
	"Programming-Demo/pkg/utils/deepseek"
	"Programming-Demo/pkg/utils/prompt"
	"context"
	"errors"
	"io"
	"net/http"
	"os"

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
	// 检查文件是否存在
	fileInfo, err := os.Stat(req.Path)
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

	content, err := os.ReadFile(req.Path)
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
	if req.Model == "moonshot" {
		Resp, code = ai.GetAIResp(p)
	} else if req.Model == "deepseek-chat" {
		Resp, code = deepseek.ChatWithDeepSeek(p, "POST", req.Model)
	} else if req.Model == "deepseek-reasoner" {
		Resp, code = deepseek.ChatWithDeepSeek(p, "POST", req.Model)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "模型错误"})
		return
	}
	if code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用ai接口失败", "error": Resp})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "message": Resp})
}

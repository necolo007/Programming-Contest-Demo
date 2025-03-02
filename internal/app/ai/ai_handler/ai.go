package ai_handler

import (
	"Programming-Demo/core/client"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/northes/go-moonshot"
	"io"
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

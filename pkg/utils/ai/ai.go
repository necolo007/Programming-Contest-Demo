package ai

import (
	"Programming-Demo/core/client"
	"context"
	"errors"
	"github.com/northes/go-moonshot"
	"io"
)

func GetAIResp(m string) (string, int) {
	resp, err := client.MoonClient.GetClient().Chat().CompletionsStream(context.Background(), &moonshot.ChatCompletionsRequest{
		Model: moonshot.ModelMoonshotV18K,
		Messages: []*moonshot.ChatCompletionsMessage{
			{
				Role:    moonshot.RoleUser,
				Content: m,
			},
		},
		Temperature: 0.3,
		Stream:      true,
	})
	var message string
	if err != nil {
		return "moonshot chat failed", 500
	} else {
		for receive := range resp.Receive() {
			msg, err1 := receive.GetMessage()
			if err1 != nil {
				if errors.Is(err1, io.EOF) {
					return message, 200
				}
				return err1.Error(), 500
			}
			message = message + msg.Content
		}
	}
	return err.Error(), 500
}

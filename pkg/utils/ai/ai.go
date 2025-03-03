package ai

import (
	"Programming-Demo/core/client"
	"context"
	"errors"
	"github.com/northes/go-moonshot"
	"io"
)

func GetAIResp(m string) string {
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
		return "moonshot chat failed"
	} else {
		for receive := range resp.Receive() {
			msg, err1 := receive.GetMessage()
			if err1 != nil {
				if errors.Is(err1, io.EOF) {
					return message
				}
				return message
			}
			message = message + msg.Content
		}
	}
	return message
}

package deepseek

import (
	"Programming-Demo/config"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type DeepseekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Message 使用结构体构建请求
type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type RequestBody struct {
	Messages         []Message `json:"messages"`
	Model            string    `json:"model"`
	FrequencyPenalty float64   `json:"frequency_penalty"`
	MaxTokens        int       `json:"max_tokens"`
	PresencePenalty  float64   `json:"presence_penalty"`
	ResponseFormat   struct {
		Type string `json:"type"`
	} `json:"response_format"`
	Stream        bool        `json:"stream"`
	StreamOptions interface{} `json:"stream_options"`
	Temperature   float64     `json:"temperature"`
	TopP          float64     `json:"top_p"`
	Tools         interface{} `json:"tools"`
	ToolChoice    string      `json:"tool_choice"`
	Logprobs      bool        `json:"logprobs"`
	TopLogprobs   interface{} `json:"top_logprobs"`
}

const BaseURL = "https://api.deepseek.com/chat/completions"

func ChatWithDeepSeek(content string, method string, model string) (string, int) {

	requestBody := RequestBody{
		Messages: []Message{
			{Content: content, Role: "system"},
			{Content: content, Role: "user"},
		},
		Model:            model,
		FrequencyPenalty: 0,
		MaxTokens:        2048,
		PresencePenalty:  0,
		ResponseFormat: struct {
			Type string `json:"type"`
		}{Type: "text"},
		Stream:      false,
		Temperature: 1,
		TopP:        1,
		ToolChoice:  "none",
		Logprobs:    false,
	}
	// 将结构体转换为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "JSON编码失败: " + err.Error(), 500
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, BaseURL, strings.NewReader(string(jsonData)))

	if err != nil {
		return err.Error(), 500
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.GetConfig().DeepSeekKey)

	res, err := client.Do(req)
	if err != nil {
		return "请求失败: " + err.Error(), 500
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "读取响应失败: " + err.Error(), 500
	}

	// 检查响应状态码
	if res.StatusCode != http.StatusOK {
		return "API 请求失败，状态码: " + res.Status + ", 响应内容: " + string(body), res.StatusCode
	}

	// 尝试解析响应
	var response DeepseekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// 打印原始响应内容以便调试
		return "解析响应失败: " + err.Error() + "\n原始响应: " + string(body), 500
	}

	// 验证响应内容
	if len(response.Choices) == 0 {
		return "响应格式正确但没有内容", 500
	}
	log.Println(response)
	return response.Choices[0].Message.Content, 200
}

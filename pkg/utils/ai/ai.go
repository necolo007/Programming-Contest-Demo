package ai

import (
	"Programming-Demo/core/client"
	"Programming-Demo/core/milvus"
	"Programming-Demo/pkg/aliyun"
	"context"
	"errors"
	"fmt"
	"github.com/northes/go-moonshot"
	"io"
	"time"
)

// Document 定义文档结构体
type Document struct {
	ID      int64
	Content string
	Score   float32
}

// GenerateEmbedding 生成文本向量
func GenerateEmbedding(text string) ([]float64, error) {
	time.Sleep(50 * time.Millisecond)
	vectors := aliyun.NplApi(text)
	return vectors, nil
}

// SearchSimilarDocuments 搜索相似文档
func SearchSimilarDocuments(query string, topK int) ([]Document, error) {
	// 生成查询的向量嵌入
	embedding, err := GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("生成查询向量失败: %v", err)
	}

	// 将float64转换为float32
	vector := make([]float32, len(embedding))
	for i, v := range embedding {
		vector[i] = float32(v)
	}

	// 搜索相似向量，同时获取内容
	ids, scores, contents, err := milvus.SearchVectors(vector, topK)
	if err != nil {
		return nil, fmt.Errorf("搜索相似向量失败: %v", err)
	}

	// 创建文档结果集
	docs := make([]Document, len(ids))
	for i := range ids {
		docs[i] = Document{
			ID:      ids[i],
			Content: contents[i],
			Score:   scores[i],
		}
	}

	return docs, nil
}

// SearchSimilarDocumentsWithParam 搜索相似文档
func SearchSimilarDocumentsWithParam(query string, topK int) ([]Document, error) {
	// 生成查询的向量嵌入
	embedding, err := GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("生成查询向量失败: %v", err)
	}

	// 将float64转换为float32
	vector := make([]float32, len(embedding))
	for i, v := range embedding {
		vector[i] = float32(v)
	}

	// 高级搜索参数
	searchParams := map[string]interface{}{
		"metric_type": "IP", // 使用内积可能对嵌入效果更好
		"params": map[string]interface{}{
			"nprobe": 50, // 大幅增加nprobe值提高召回率
		},
	}

	// 搜索相似向量，��用增强版本的搜索函数
	ids, scores, contents, err := milvus.SearchVectorsWithParams(vector, topK*2, searchParams)
	if err != nil {
		// 回退到基本搜索
		ids, scores, contents, err = milvus.SearchVectors(vector, topK)
		if err != nil {
			return nil, fmt.Errorf("搜索相似向量失败: %v", err)
		}
	}

	// 创建文档结果集
	docs := make([]Document, len(ids))
	for i := range ids {
		docs[i] = Document{
			ID:      ids[i],
			Content: contents[i],
			Score:   scores[i],
		}
	}

	return docs, nil
}

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

package ai

import (
	bochalient "Programming-Demo/core/Bocha_client"
	"Programming-Demo/core/client"
	"Programming-Demo/core/milvus"
	"Programming-Demo/internal/app/ai/ai_entity"
	"Programming-Demo/pkg/aliyun"
	"Programming-Demo/pkg/utils/bocha"
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

func WebBaseSearch(content string) (error, string) {
	// 检查博查客户端是否已初始化
	if bochalient.BochaClient == nil {
		return fmt.Errorf("联网搜索功能未配置，请设置BOCHA_API_KEY环境变量"), ""
	}

	// 使用博查API进行搜索
	searchReq := bocha.SearchRequest{
		Query:     content,
		Freshness: "noLimit", // a.使用默认时间范围
		Summary:   true,      // 获取完整摘要
		Count:     15,        // 获取15条结果
		Page:      1,
	}

	// 使用封装的bochalient.BochaClient获取客户端并执行搜索
	searchResult, err := bochalient.BochaClient.GetClient().Search(searchReq)
	if err != nil {
		return err, ""
	}

	// 解析搜索结果，提取有用信息
	searchInfo, err := bocha.ExtractSearchInfo(searchResult)
	if err != nil {
		searchInfo = "搜索结果解析失败，但仍可能包含有用信息: " + searchResult
		return fmt.Errorf("解析搜索结果出错: %v，将使用原始结果\n", err), searchInfo
	}

	return nil, searchInfo
}

// GenerateWebSearchPrompt 生成增强型法律助手提示
func GenerateWebSearchPrompt(theme string, histories []ai_entity.ChatHistory, currentQuestion string, searchInfo string) string {
	basePrompt := `# AI法律助手增强型提示框架

## 角色定义
你是一个专业的法律助手，拥有以下核心特质：
- 精通中国现行法律体系，能准确引用最新法律法规、司法解释及指导案例
- 具备严谨、专业、客观的法律分析能力和批判性思维
- 能提供基于法律条文和司法实践的准确分析和建议
- 善于将复杂的法律概念转化为易于理解的语言，同时保持法律表述的精确性

## 基本工作原则

### 1. 法律专业性原则
- 所有回复必须基于现行有效的中国法律法规，确保引用的法律为最新版本
- 引用法律条文时必须准确标注：《法律名称》第X条第X款第X项，并附上条文原文
- 区分强制性规范与任意性规范，明确说明法律要求与建议性内容的区别
- 对于存在争议的法律问题，应当呈现不同观点和可能的法律后果
- 明确指出法律规定与实践操作之间可能存在的差异

### 2. 专业边界与责任限制原则
- 明确表明所提供的信息仅为一般性法律参考，不构成正式法律意见
- 复杂或高风险问题应建议用户咨询具有执业资格的专业律师
- 不对特定案件结果做出保证或预测
- 对于需要专业判断的问题（如证据采信、责任划分等），提供法律框架而非确定性结论

### 3. 信息安全与隐私保护原则
- 不提供可能违法或有害的建议，拒绝协助规避法律的请求
- 遵循最小信息收集原则，不主动索取无关的个人敏感信息
- 提醒用户在描述法律问题时注意保护个人身份信息和隐私
- 建议用户在讨论敏感法律事项时采取适当的信息安全措施

## 回复框架与质量标准

### 回复结构
1. **法律问题界定**：准确理解并重述用户咨询的法律问题
2. **法律依据分析**：
   - 相关法律法规条文引用（附条文原文）
   - 司法解释或指导性案例（如适用）
   - 法理学原则或学说（如适用）
3. **法律分析与推理**：
   - 将法律条文应用于具体情境
   - 多角度分析可能的法律后果
   - 明确区分事实问题与法律问题
4. **实用建议与风险提示**：
   - 可行的解决途径及其法律后果
   - 潜在风险和注意事项
   - 必要的程序性指导（如适用）
5. **总结与免责声明**：简明扼要地总结核心观点，并附上适当的免责声明

## 当前主题与情境适配
特定主题: ` + theme

	// 添加对话历史
	basePrompt += "\n\n## 对话历史："
	for _, history := range histories {
		role := "用户"
		if history.Role == "assistant" {
			role = "法律助手"
		}
		basePrompt += fmt.Sprintf("\n%s: %s", role, history.Content)
	}

	// 添加当前问题
	basePrompt += fmt.Sprintf("\n\n## 用户最新问题：\n%s", currentQuestion)

	// 添加当前问题
	basePrompt += fmt.Sprintf("\n\n## 联网搜索结果：\n%s", searchInfo)

	// 添加回复指南
	basePrompt += `

## 回复要求
1. 分析用户问题的核心法律问题
2. 引用相关法律条文（包括条文原文）
3. 提供专业法律分析和推理
4. 给出实用建议和风险提示
5. 使用清晰的结构，确保回答易于理解
6. 涉及复杂问题时，建议咨询专业律师进行具体指导
7. 回复结尾添加简短的免责声明

请基于以上指南，提供专业、准确、有深度的法律回答。`

	return basePrompt
}

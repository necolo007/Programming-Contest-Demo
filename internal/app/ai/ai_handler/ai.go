package ai_handler

import (
	bochalient "Programming-Demo/core/Bocha_client"
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/core/libx"
	"Programming-Demo/internal/app/File/file_entity"
	"Programming-Demo/internal/app/ai/ai_dto"
	"Programming-Demo/internal/app/ai/ai_entity"
	"Programming-Demo/internal/app/ai/ai_service"
	"Programming-Demo/pkg/utils/ai"
	"Programming-Demo/pkg/utils/bocha"
	"Programming-Demo/pkg/utils/deepseek"
	"Programming-Demo/pkg/utils/prompt"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DefaultContextMessageCount = 10             // 默认上下文消息数量
	MaxContextMessageCount     = 50             // 最大上下文消息数量
	CacheExpiration            = 24 * time.Hour // 缓存过期时间
)

// 处理用户与AI的聊天请求
func ChatWithAi(c *gin.Context) {
	// 获取用户ID
	uid := libx.Uid(c)

	// 解析请求
	var req ai_dto.ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 如果主题为空，则生成主题名称
	if req.Theme == "" {
		themeName, err := GenerateThemeName(req.Content, req.Model)
		if err != nil {
			// 如果生成失败，使用默认主题名称
			req.Theme = "法律咨询_" + time.Now().Format("20060102150405")
			fmt.Printf("Failed to generate theme name: %v, using default\n", err)
		} else {
			req.Theme = themeName
		}
	}

	// 初始化Ristretto缓存
	if err := ai_service.InitCache(); err != nil {
		// 记录错误但继续执行，因为缓存功能不是核心功能
		fmt.Printf("Failed to initialize cache: %v\n", err)
	}

	// 尝试从缓存加载历史记录
	cache, err := ai_service.LoadChatCache(uid, req.Theme)
	var histories []ai_entity.ChatHistory

	if err != nil || !ai_service.IsCacheValid(cache, CacheExpiration) {
		// 缓存不可用，从数据库加载
		histories, err = ai_service.GetRecentChatHistoryByTheme(uid, req.Theme, DefaultContextMessageCount)
		if err != nil {
			// 记录错误但继续，如果无法获取历史记录，就使用空记录
			fmt.Printf("Failed to get chat history: %v\n", err)
			histories = []ai_entity.ChatHistory{}
		}
	} else {
		// 使用缓存的消息
		histories = cache.Messages
	}

	// 限制上下文消息数量，避免超出AI模型的上下文限制
	contextCount := DefaultContextMessageCount
	if len(histories) > contextCount {
		histories = histories[len(histories)-contextCount:]
	}

	// 保存当前用户消息到历史记录
	userMessage := ai_entity.ChatHistory{
		UserID:  uid,
		Model:   req.Model,
		Theme:   req.Theme,
		Role:    "user",
		Content: req.Content,
	}

	// 使用事务确保数据一致性
	tx := dbs.DB.Begin()
	if err := tx.Create(&userMessage).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存用户消息失败", "error": err.Error()})
		return
	}

	// 更新主题最后消息时间
	if err := ai_service.UpdateOrCreateTheme(uid, req.Theme); err != nil {
		// 记录错误但继续，因为这不是核心功能
		fmt.Printf("Failed to update theme: %v\n", err)
	}

	var prompt, searchInfo string
	if req.Search == true {
		err, searchInfo = ai.WebBaseSearch(req.Content)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "联网搜索失败", "error": err.Error()})
			return
		}
		log.Println(searchInfo)
		prompt = ai.GenerateWebSearchPrompt(req.Theme, histories, req.Content, searchInfo)
	} else {
		prompt = generateLegalAssistantPrompt(req.Theme, histories, req.Content)
		searchInfo = ""
	}

	var Resp string
	var code int

	// 选择不同的 AI 模型处理
	switch req.Model {
	case "moonshot":
		// 使用结构化提示作为输入
		Resp, code = ai.GetAIResp(prompt)
	case "deepseek-chat", "deepseek-reasoner":
		Resp, code = deepseek.ChatWithDeepSeek(prompt, "POST", req.Model)
	default:
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"message": "模型错误"})
		return
	}

	if code != 200 {
		tx.Rollback()
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

	if err := tx.Create(&aiMessage).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存AI回复失败", "error": err.Error()})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "提交事务失败", "error": err.Error()})
		return
	}

	// 更新历史记录列表，包括新的消息
	histories = append(histories, userMessage, aiMessage)

	// 更新缓存 - 使用 goroutine 异步执行，不阻塞主流程
	go func() {
		if err := ai_service.SaveChatCache(uid, req.Theme, histories); err != nil {
			fmt.Printf("Failed to save chat cache: %v\n", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":       code,
		"searchInfo": searchInfo,
		"theme":      req.Theme,
		"message":    Resp,
	})
}

// 获取聊天历史记录
func GetChatHistory(c *gin.Context) {
	uid := libx.Uid(c)
	theme := c.Query("theme")
	limitStr := c.DefaultQuery("limit", "50")
	var limit int
	fmt.Sscanf(limitStr, "%d", &limit)

	if limit <= 0 || limit > MaxContextMessageCount {
		limit = DefaultContextMessageCount
	}

	// 尝试从缓存加载
	cache, err := ai_service.LoadChatCache(uid, theme)
	var histories []ai_entity.ChatHistory

	if err != nil || !ai_service.IsCacheValid(cache, CacheExpiration) {
		// 从数据库加载
		histories, err = ai_service.GetChatHistoryByTheme(uid, theme, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "获取历史记录失败", "error": err.Error()})
			return
		}

		// 更新缓存 - 使用 goroutine 异步执行
		go func() {
			if err := ai_service.SaveChatCache(uid, theme, histories); err != nil {
				fmt.Printf("Failed to save chat cache: %v\n", err)
			}
		}()
	} else {
		histories = cache.Messages
		// 如果请求的限制与缓存不同，则过滤
		if len(histories) > limit {
			histories = histories[len(histories)-limit:]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    histories,
	})
}

// 获取聊天主题
func GetChatThemes(c *gin.Context) {
	uid := libx.Uid(c)

	// 获取所有主题（包括元数据）
	themes, err := ai_service.GetAllChatThemes(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取主题失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    themes,
	})
}

// 删除聊天主题及其历史记录
func DeleteChatTheme(c *gin.Context) {
	uid := libx.Uid(c)
	themeIDStr := c.Query("id")

	if themeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "主题ID不能为空",
		})
		return
	}

	// 将ID字符串转换为uint
	var themeID uint
	if _, err := fmt.Sscanf(themeIDStr, "%d", &themeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的主题ID格式",
			"error":   err.Error(),
		})
		return
	}

	// 首先查询主题以获取主题名称(用于后续删除缓存)
	var theme ai_entity.ChatTheme
	if err := dbs.DB.Where("id = ? AND user_id = ?", themeID, uid).First(&theme).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "主题不存在或无权限访问",
			"error":   err.Error(),
		})
		return
	}

	// 使用事务删除相关记录
	tx := dbs.DB.Begin()

	// 删除聊天历史记录
	if err := tx.Where("user_id = ? AND theme = ?", uid, theme.Theme).
		Delete(&ai_entity.ChatHistory{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除聊天历史失败",
			"error":   err.Error(),
		})
		return
	}

	// 删除主题记录
	if err := tx.Delete(&ai_entity.ChatTheme{}, themeID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除主题失败",
			"error":   err.Error(),
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
			"error":   err.Error(),
		})
		return
	}

	// 删除缓存 - 异步执行
	go func() {
		if err := ai_service.DeleteChatCache(uid, theme.Theme); err != nil {
			fmt.Printf("Failed to delete chat cache: %v\n", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// 文件分析函数
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

// DeepSeek和博查API实现联网搜索
func AiSearch(c *gin.Context) {
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

	// 检查博查客户端是否已初始化
	if bochalient.BochaClient == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "联网搜索功能未配置，请设置BOCHA_API_KEY环境变量",
		})
		return
	}

	// 如果主题为空，则生成主题名称
	if req.Theme == "" {
		themeName, err := GenerateThemeName(req.Content, req.Model)
		if err != nil {
			// 如果生成失败，使用默认主题名称
			req.Theme = "联网搜索_" + time.Now().Format("20060102150405")
			fmt.Printf("Failed to generate theme name for user %v: %v, using default\n", uid, err)
		} else {
			req.Theme = themeName
		}
	}

	// 初始化Ristretto缓存
	if err := ai_service.InitCache(); err != nil {
		fmt.Printf("Failed to initialize cache: %v\n", err)
	}

	// 尝试从缓存加载历史记录
	cache, err := ai_service.LoadChatCache(uid, req.Theme)
	var histories []ai_entity.ChatHistory

	if err != nil || !ai_service.IsCacheValid(cache, CacheExpiration) {
		// 缓存不可用，从数据库加载
		histories, err = ai_service.GetRecentChatHistoryByTheme(uid, req.Theme, DefaultContextMessageCount)
		if err != nil {
			fmt.Printf("Failed to get chat history: %v\n", err)
			histories = []ai_entity.ChatHistory{}
		}
	} else {
		// 使用缓存的消息
		histories = cache.Messages
	}

	// 限制上下文消息数量
	contextCount := DefaultContextMessageCount
	if len(histories) > contextCount {
		histories = histories[len(histories)-contextCount:]
	}

	// 保存当前用户消息到历史记录
	userMessage := ai_entity.ChatHistory{
		UserID:  uid,
		Model:   req.Model,
		Theme:   req.Theme,
		Role:    "user",
		Content: req.Content,
	}

	// 使用事务确保数据一致性
	tx := dbs.DB.Begin()
	if err := tx.Create(&userMessage).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存用户消息失败", "error": err.Error()})
		return
	}

	// 更新主题最后消息时间
	if err := ai_service.UpdateOrCreateTheme(uid, req.Theme); err != nil {
		fmt.Printf("Failed to update theme: %v\n", err)
	}

	// 使用博查API进行搜索
	searchReq := bocha.SearchRequest{
		Query:     req.Content,
		Freshness: "noLimit", // a.使用默认时间范围
		Summary:   true,      // 获取完整摘要
		Count:     15,        // 获取15条结果
		Page:      1,
	}

	// 使用封装的bochalient.BochaClient获取客户端并执行搜索
	searchResult, err := bochalient.BochaClient.GetClient().Search(searchReq)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索失败",
			"error":   err.Error(),
		})
		return
	}

	// 解析搜索结果，提取有用信息
	searchInfo, err := bocha.ExtractSearchInfo(searchResult)
	if err != nil {
		fmt.Printf("解析搜索结果出错: %v，将使用原始结果\n", err)
		searchInfo = "搜索结果解析失败，但仍可能包含有用信息: " + searchResult
	}

	// 构建带搜索信息的提示词
	prompt := fmt.Sprintf(`现在是%s，你是一位专业的AI助手，以下是用户的问题，以及我为你提供的最新网络搜索结果。
请根据这些搜索结果回答用户问题。注意以下几点：
1. 如果搜索结果提供了足够信息，请直接回答问题，并引用搜索结果中的相关信息
2. 如果搜索结果包含多个来源，请综合各个来源的信息进行回答
3. 如果搜索结果不足以回答问题，请诚实告知用户，并尽可能提供相关信息
4. 回答中应引用信息来源(例如网站名称)，以便用户验证
5. 保持客观、准确，不要添加搜索结果中没有的信息

用户问题：%s

网络搜索结果：
%s

请基于上述搜索结果回答用户问题：`,
		time.Now().Format("2006年01月02日"),
		req.Content,
		searchInfo,
	)

	// 将历史记录转换为DeepSeek可用的格式
	var messagesContent []string
	for _, history := range histories {
		rolePrefix := ""
		if history.Role == "user" {
			rolePrefix = "用户: "
		} else if history.Role == "assistant" {
			rolePrefix = "AI助手: "
		}
		messagesContent = append(messagesContent, rolePrefix+history.Content)
	}

	// 用户消息加上搜索提示词，不显示给前端
	finalPrompt := ""
	if len(messagesContent) > 0 {
		finalPrompt = strings.Join(messagesContent, "\n") + "\n\n" + prompt
	} else {
		finalPrompt = prompt
	}

	// 使用DeepSeek获取回复
	resp, code := deepseek.ChatWithDeepSeek(finalPrompt, "POST", req.Model)

	if code != 200 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用AI接口失败", "error": resp})
		return
	}

	// 保存AI回复到历史记录
	aiMessage := ai_entity.ChatHistory{
		UserID:  uid,
		Model:   req.Model,
		Theme:   req.Theme,
		Role:    "assistant",
		Content: resp,
	}

	if err := tx.Create(&aiMessage).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存AI回复失败", "error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "提交事务失败", "error": err.Error()})
		return
	}

	// 更新历史记录列表
	histories = append(histories, userMessage, aiMessage)

	// 更新缓存
	go func() {
		if err := ai_service.SaveChatCache(uid, req.Theme, histories); err != nil {
			fmt.Printf("Failed to save chat cache: %v\n", err)
		}
	}()

	// 返回响应，包含主题名称
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": resp,
		"theme":   req.Theme, // 添加主题到响应中
	})
}

func GenerateLegalDocBetter(c *gin.Context) {
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
	ps, docs := prompt.BuildRAGPrompt(p)
	// 选择不同的 AI 模型处理
	switch req.Model {
	case "moonshot":
		Resp, code = ai.GetAIResp(ps)
	case "deepseek-chat", "deepseek-reasoner":
		Resp, code = deepseek.ChatWithDeepSeek(ps, "POST", req.Model)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "模型错误"})
		return
	}
	if code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "调用ai接口失败", "error": Resp})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "doc": docs, "message": Resp})
}

// 生成增强型法律助手提示
func generateLegalAssistantPrompt(theme string, histories []ai_entity.ChatHistory, currentQuestion string) string {
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

// 根据用户问题生成主题名称
func GenerateThemeName(question string, model string) (string, error) {
	// 构建主题生成提示
	prompt := `请为以下法律咨询问题生成一个简短的主题名称（不超过15个字），主题名称应该概括问题的核心法律领域和关键事项。
    
问题内容：
"""
` + question + `
"""

请只返回主题名称，不需要任何解释或额外内容。`

	// 调用AI获取主题名称
	var themeName string
	var code int

	switch model {
	case "moonshot":
		themeName, code = ai.GetAIResp(prompt)
	case "deepseek-chat", "deepseek-reasoner":
		themeName, code = deepseek.ChatWithDeepSeek(prompt, "POST", model)
	default:
		return "", fmt.Errorf("不支持的模型类型")
	}

	if code != 200 {
		return "", fmt.Errorf("生成主题名称失败: %s", themeName)
	}

	// 清理主题名称中可能的多余内容（如引号、空格等）
	themeName = strings.TrimSpace(themeName)
	themeName = strings.Trim(themeName, "\"'")

	// 限制主题名称长度
	if len([]rune(themeName)) > 15 {
		themeName = string([]rune(themeName)[:15])
	}

	return themeName, nil
}

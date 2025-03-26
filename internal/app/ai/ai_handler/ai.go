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

// 与AI聊天
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

	// 构建对话上下文
	var messagesContent []string
	for _, history := range histories {
		// 格式化消息以区分角色
		rolePrefix := ""
		if history.Role == "user" {
			rolePrefix = "用户: "
		} else if history.Role == "assistant" {
			rolePrefix = "AI助手: "
		}
		messagesContent = append(messagesContent, rolePrefix+history.Content)
	}

	// 添加当前用户消息
	messagesContent = append(messagesContent, "用户: "+req.Content)

	// 将当前问题保存到历史记录
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

	var Resp string
	var code int

	// 选择不同的 AI 模型处理
	switch req.Model {
	case "moonshot":
		// 使用聊天内容作为输入
		Resp, code = ai.GetAIResp(strings.Join(messagesContent, "\n"))
	case "deepseek-chat", "deepseek-reasoner":
		Resp, code = deepseek.ChatWithDeepSeek(strings.Join(messagesContent, "\n"), "POST", req.Model)
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
		"code":    code,
		"message": Resp,
		"history": histories,
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

// BoTool 博查AI搜索工具定义
type BoTool struct {
	Type     string `json:"type"`
	Function BoDef  `json:"function"`
}

// BoDef 博查AI搜索函数定义
type BoDef struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// BochaSearchRequest 博查搜索请求结构
type BochaSearchRequest struct {
	Query     string `json:"query"`
	Freshness string `json:"freshness"`
	Summary   bool   `json:"summary"`
	Count     int    `json:"count"`
	Page      int    `json:"page"`
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
		Freshness: "noLimit", // 使用默认时间范围
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

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": resp,
		"history": histories,
	})
}

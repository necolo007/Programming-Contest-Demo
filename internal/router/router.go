package router

import (
	"Programming-Demo/core/middleware/web"
	"Programming-Demo/internal/app/File/file_handler"
	"Programming-Demo/internal/app/ai/ai_handler"
	"Programming-Demo/internal/app/file_search/search_handler"
	"Programming-Demo/internal/app/story/story_handler"
	"Programming-Demo/internal/app/template/template_handler"
	"Programming-Demo/internal/app/user/user_handler"

	"github.com/gin-gonic/gin"
)

func GenerateRouters(r *gin.Engine) *gin.Engine {
	// 用户相关路由
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/register", user_handler.Register)
		userGroup.POST("/login", user_handler.Login)
		userGroup.POST("/logout", web.JWTAuthMiddleware(), user_handler.Logout)
	}
	aiGroup := r.Group("/api/ai", web.JWTAuthMiddleware())
	{
		aiGroup.POST("/analyze", ai_handler.AnalyzeFile)
		aiGroup.POST("/contract", ai_handler.GenerateLegalDocument)
		aiGroup.POST("/complain", ai_handler.GenerateComplaint)
		aiGroup.POST("/opinion", ai_handler.GenerateLegalOpinion)
		aiGroup.POST("/chat", ai_handler.ChatWithAi)
		aiGroup.GET("/history", ai_handler.GetChatHistory)
		aiGroup.GET("/theme", ai_handler.GetChatThemes)
		aiGroup.DELETE("/delete", ai_handler.DeleteChatTheme)
		aiGroup.GET("/search", ai_handler.AiSearch)
		aiGroup.POST("/doc/more", ai_handler.GenerateLegalDocBetter)
	}
	// 管理员相关路由
	adminGroup := r.Group("/api/admin", web.JWTAuthMiddleware(), web.AdminAuthMiddleware())
	{
		// 这里可以添加需要管理员权限的路由
		// 管理员可以获取所有用户信息
		adminGroup.GET("/users", user_handler.GetAllUsers)
		adminGroup.POST("/audit/:id", file_handler.AuditFile)
		adminGroup.POST("/audit", file_handler.ListPendingFiles)
		adminGroup.GET("/audit/:id/get", file_handler.GetPendingFileHandler)
	}
	fileGroup := r.Group("/api/file", web.JWTAuthMiddleware())
	{
		fileGroup.POST("/upload", file_handler.UploadFileHandler)
		fileGroup.GET("/download/:id", file_handler.DownloadFileHandler)
		fileGroup.DELETE("/delete/:id", file_handler.DeleteFileHandler)
		// 文件搜索相关路由
		searchGroup := fileGroup.Group("/search")
		{
			searchGroup.POST("/keyword", search_handler.KeywordSearch)       // 关键词搜索（支持精确和模糊搜索）
			searchGroup.POST("/advanced", search_handler.AdvancedSearch)     // 高级搜索
			searchGroup.POST("/semantic", search_handler.SemanticSearch)     // 语义搜索
			searchGroup.GET("/type", search_handler.SearchFileByTypeHandler) // 按文件类型搜索
		}
		templateGroup := r.Group("/api/template", web.JWTAuthMiddleware())
		{
			templateGroup.POST("/upload", web.AdminAuthMiddleware(), template_handler.CreateTemplateHandler)
		}
		storyGroup := r.Group("/api/story", web.JWTAuthMiddleware())
		{
			storyGroup.POST("/create", story_handler.CreateStory)
			storyGroup.GET("/get", story_handler.GetRandomStory)
		}
		return r
	}
}

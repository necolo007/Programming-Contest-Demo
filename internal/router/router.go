package router

import (
	"Programming-Demo/core/middleware/web"
	"Programming-Demo/internal/app/File/file_handler"
	"Programming-Demo/internal/app/ai/ai_handler"
	"Programming-Demo/internal/app/file_search/search_handler"
	"Programming-Demo/internal/app/user/user_handler"

	"github.com/gin-gonic/gin"
)

func GenerateRouters(r *gin.Engine) *gin.Engine {
	r.GET("/ping", ai_handler.PingMoonshot)
	// 用户相关路由
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/register", user_handler.Register)
		userGroup.POST("/login", user_handler.Login)
		userGroup.POST("/logout", web.JWTAuthMiddleware(), user_handler.Logout)
	}
	aiGroup := r.Group("/api/ai")
	{
		aiGroup.POST("/analyze", web.JWTAuthMiddleware(), ai_handler.AnalyzeFile)
		aiGroup.POST("/contract", web.JWTAuthMiddleware(), ai_handler.GenerateLegalDocument)
		aiGroup.POST("/complain", web.JWTAuthMiddleware(), ai_handler.GenerateComplaint)
		aiGroup.POST("/opinion", web.JWTAuthMiddleware(), ai_handler.GenerateLegalOpinion)
		aiGroup.POST("/chat", web.JWTAuthMiddleware(), ai_handler.ChatWithAi)
	}
	// 管理员相关路由
	adminGroup := r.Group("/api/admin", web.JWTAuthMiddleware(), web.AdminAuthMiddleware())
	{
		// 这里可以添加需要管理员权限的路由
		// 管理员可以获取所有用户信息
		adminGroup.GET("/users", user_handler.GetAllUsers)
	}
	fileGroup := r.Group("/api/file", web.JWTAuthMiddleware(), web.AdminAuthMiddleware())
	{
		fileGroup.POST("/upload", web.JWTAuthMiddleware(), file_handler.UploadFileHandler)
		fileGroup.GET("/download/:id", web.JWTAuthMiddleware(), file_handler.DownloadFileHandler)
		fileGroup.DELETE("/delete/:id", web.JWTAuthMiddleware(), file_handler.DeleteFileHandler)
		// 文件搜索相关路由
		searchGroup := fileGroup.Group("/search")
		{
			searchGroup.POST("/keyword", search_handler.KeywordSearch)             // 关键词搜索（支持精确和模糊搜索）
			searchGroup.POST("/advanced", search_handler.AdvancedSearch)           // 高级搜索
			searchGroup.POST("/semantic", search_handler.SemanticSearch)           // 语义搜索
			searchGroup.GET("/type", search_handler.SearchFileByTypeHandler)       // 按文件类型搜索
			searchGroup.GET("/content", search_handler.SearchFileByContentHandler) // 按文件内容搜索
		}
	}
	return r
}

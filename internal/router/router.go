package router

import (
	"Programming-Demo/internal/app/ai/ai_handler"
	"github.com/gin-gonic/gin"
)

func GenerateRouters(r *gin.Engine) *gin.Engine {
	r.GET("/ping", ai_handler.PingMoonshot)
	return r
}

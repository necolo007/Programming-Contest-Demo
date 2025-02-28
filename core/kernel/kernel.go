package kernel

import (
	"context"
	"github.com/gin-gonic/gin"
)

type Engine struct {
	// 不由 Engine 统一管理
	//Mysql      *gorm.DB
	//Cache      *redis.Client

	Gin *gin.Engine

	Ctx    context.Context
	Cancel context.CancelFunc
}

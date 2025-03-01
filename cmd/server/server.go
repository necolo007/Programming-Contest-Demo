package server

import (
	"Programming-Demo/config"
	"Programming-Demo/core/database"
	"Programming-Demo/core/gin"
	"Programming-Demo/core/kernel"
	"Programming-Demo/pkg/ip"
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configYml string
	engine    *kernel.Engine
)

func SetUp() {
	// 初始化全局 ctx
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化资源管理器
	engine = &kernel.Engine{Ctx: ctx, Cancel: cancel}

	config.LoadConfig(configYml)

	database.InitDB()
	engine.Gin = gin.GinInit()
}

// Run 运行 Gin 服务器
func Run() {
	// 创建 HTTP 服务器实例
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.GetConfig().Host, config.GetConfig().Port),
		Handler: engine.Gin,
	}

	// 启动服务器（异步启动，避免阻塞）
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			color.Red("Server Error: %s", err.Error())
		}
	}()

	// 打印服务器启动信息
	color.Green("Server running at:")
	color.Green("-  Local:   http://localhost:%s", config.GetConfig().Port)
	for _, host := range ip.GetLocalHost() {
		color.Green("-  Network: http://%s:%s", host, config.GetConfig().Port)
	}

	// 捕获系统信号，等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 捕获 SIGINT 和 SIGTERM 信号
	<-quit
	color.Blue("Shutting down server...")

	// 创建一个带超时的上下文用于优雅地关闭服务器
	ctx, cancel := context.WithTimeout(engine.Ctx, 5*time.Second)
	defer cancel()

	// 优雅地关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		color.Yellow("Server forced to shutdown: %s", err.Error())
	} else {
		color.Green("Server exited gracefully")
	}
}

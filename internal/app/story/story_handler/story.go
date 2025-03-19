package story_handler

import (
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/core/libx"
	"Programming-Demo/internal/app/story/story_entity"
	"Programming-Demo/internal/app/user/user_entity"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetRandomStory(c *gin.Context) {
	// 检查用户登录状态
	uid := libx.Uid(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
		return
	}

	// 查询用户最后一次接收推送的时间
	var lastPushTime time.Time
	if err := dbs.DB.Model(&user_entity.User{}).Where("id = ?", uid).Select("last_push_time").Scan(&lastPushTime).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户信息失败",
			"error":   err.Error(),
		})
		return
	}

	// 检查时间间隔是否达到2天
	if time.Since(lastPushTime).Hours() < 48 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "未达到推送间隔时间",
		})
		return
	}

	var story []story_entity.Story
	if err := dbs.DB.Find(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取故事列表失败",
			"error":   err.Error(),
		})
		return
	}
	if len(story) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "无可用故事",
		})
		return
	}

	randomIndex := rand.Intn(len(story))
	selectedStory := story[randomIndex]

	// 更新用户最后一次接收推送的时间
	if err := dbs.DB.Model(&user_entity.User{}).Where("id = ?", uid).Update("last_push_time", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户信息失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取随机故事成功",
		"data":    selectedStory,
	})
}

func CreateStory(c *gin.Context) {
	var story story_entity.Story
	if err := c.ShouldBindJSON(&story); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	db := dbs.DB
	if err := db.Create(&story).Error; err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "故事写入成功",
	})
}

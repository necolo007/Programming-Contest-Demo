package story_handler

import (
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/core/libx"
	"Programming-Demo/internal/app/story/story_entity"
	"math/rand"
	"net/http"

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

	var stories []story_entity.Story
	if err := dbs.DB.Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取故事列表失败",
			"error":   err.Error(),
		})
		return
	}
	if len(stories) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "无可用故事",
		})
		return
	}

	randomIndex := rand.Intn(len(stories))
	selectedStory := stories[randomIndex]

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

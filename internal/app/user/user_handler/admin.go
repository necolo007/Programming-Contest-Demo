package user_handler

import (
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/internal/app/user/user_entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetAllUsers 获取所有用户信息（仅管理员可用）
func GetAllUsers(c *gin.Context) {
	var users []user_entity.User
	if err := dbs.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户列表失败",
			"error":   err.Error(),
		})
		return
	}

	// 移除敏感信息
	var safeUsers []gin.H
	for _, user := range users {
		safeUsers = append(safeUsers, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取用户列表成功",
		"data":    safeUsers,
	})
}

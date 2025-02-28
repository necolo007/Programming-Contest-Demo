package web

import (
	"Programming-Demo/core/auth"
	"github.com/gin-gonic/gin"
	"net/http"
)

// JWTAuthMiddleware 是一个Gin中间件，用于验证JWT token
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": http.StatusUnauthorized,
					"msg":  "请求未携带token，无权限访问",
				})
				c.Abort()
				return
			}
		} else if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "无效的token形式",
			})
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "无效或者过期的token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中，以便后续处理使用
		c.Set("uid", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

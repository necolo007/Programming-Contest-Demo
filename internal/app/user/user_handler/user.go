package user_handler

import (
	"Programming-Demo/core/auth"
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/internal/app/user/user_dto"
	"Programming-Demo/internal/app/user/user_entity"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func Register(c *gin.Context) {
	var req user_dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existUser user_entity.User
	if err := dbs.DB.Where("username = ?", req.Username).First(&existUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名已存在",
		})
		return
	}

	// 检查邮箱是否已存在
	if err := dbs.DB.Where("email = ?", req.Email).First(&existUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "邮箱已被使用",
		})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "服务器内部错误",
		})
		return
	}

	// 创建新用户
	user := user_entity.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Role:     "user",
	}

	if err := dbs.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "用户创建失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
	})
}

func Login(c *gin.Context) {
	var req user_dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数格式不正确",
		})
		return
	}

	// 查找用户
	var user user_entity.User
	if err := dbs.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 生成JWT令牌
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "系统错误，请稍后重试",
		})
		return
	}

	// 根据用户角色返回不同的响应
	response := gin.H{
		"code":    200,
		"message": "登录成功",
	}

	if user.Role == "admin" {
		response["data"] = user_dto.AdminLoginResp{
			Token: token,
			Role:  user.Role,
		}
	} else {
		response["data"] = user_dto.LoginResp{
			Token: token,
		}
	}

	c.JSON(http.StatusOK, response)
}

func Logout(c *gin.Context) {
	// 由于使用的是JWT，服务端不需要维护会话状态
	// 客户端只需要删除本地存储的token即可
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "退出登录成功",
	})
}

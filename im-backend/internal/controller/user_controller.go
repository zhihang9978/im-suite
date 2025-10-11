package controller

import (
	"net/http"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

// GetCurrentUser 获取当前用户信息
func (uc *UserController) GetCurrentUser(c *gin.Context) {
	// 从上下文获取用户ID（由AuthMiddleware设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	// 查询用户信息
	var user model.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	// 返回用户信息（不包含密码）
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       user.ID,
			"phone":    user.Phone,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"bio":      user.Bio,
			"language": user.Language,
			"theme":    user.Theme,
			"online":   user.Online,
		},
	})
}

// GetFriends 获取好友列表
func (uc *UserController) GetFriends(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	// 查询好友列表（简化实现，实际应该查询好友关系表）
	// 这里暂时返回空列表
	var friends []model.User
	
	// TODO: 实现真实的好友查询逻辑
	// config.DB.Where("user_id = ?", userID).Find(&friends)
	
	_ = userID // 避免未使用变量警告

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    friends, // 空列表也是成功
	})
}


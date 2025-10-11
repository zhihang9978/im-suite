package controller

import (
	"net/http"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/gin-gonic/gin"
)

type UserSearchController struct{}

func NewUserSearchController() *UserSearchController {
	return &UserSearchController{}
}

// SearchUsers 搜索用户（公开API，用于添加好友）
func (usc *UserSearchController) SearchUsers(c *gin.Context) {
	phone := c.Query("phone")
	username := c.Query("username")
	keyword := c.Query("keyword")

	if phone == "" && username == "" && keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请提供搜索条件（phone、username或keyword）",
		})
		return
	}

	var users []model.User
	query := config.DB.Where("is_active = ?", true)

	// 按手机号精确搜索
	if phone != "" {
		query = query.Where("phone = ?", phone)
	}

	// 按用户名精确搜索
	if username != "" {
		query = query.Where("username = ?", username)
	}

	// 按关键词模糊搜索
	if keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ? OR phone LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 限制返回数量
	query = query.Limit(20)

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "搜索失败",
			"details": err.Error(),
		})
		return
	}

	// 移除敏感信息
	for i := range users {
		users[i].Password = ""
		users[i].Salt = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"count":   len(users),
	})
}

// GetUserByPhone 通过手机号获取用户（需要认证）
func (usc *UserSearchController) GetUserByPhone(c *gin.Context) {
	phone := c.Param("phone")

	var user model.User
	if err := config.DB.Where("phone = ? AND is_active = ?", phone, true).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "用户不存在",
		})
		return
	}

	// 移除敏感信息
	user.Password = ""
	user.Salt = ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}


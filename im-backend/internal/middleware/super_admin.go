package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// SuperAdmin 瓒呯骇绠＄悊鍛樹腑闂翠欢
func SuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 鑾峰彇褰撳墠鐢ㄦ埛ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "鏈璇?,
			})
			c.Abort()
			return
		}

		// 鏌ヨ鐢ㄦ埛淇℃伅
		var user model.User
		err := config.DB.First(&user, userID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "鐢ㄦ埛涓嶅瓨鍦?,
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "鏌ヨ鐢ㄦ埛澶辫触",
				})
			}
			c.Abort()
			return
		}

		// 妫€鏌ユ槸鍚︿负瓒呯骇绠＄悊鍛?		if user.Role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "鏉冮檺涓嶈冻锛岄渶瑕佽秴绾х鐞嗗憳鏉冮檺",
			})
			c.Abort()
			return
		}

		// 璁板綍绠＄悊鍛樻搷浣?		c.Set("admin_id", user.ID)
		c.Set("admin_username", user.Username)

		c.Next()
	}
}

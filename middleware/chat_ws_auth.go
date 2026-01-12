package middleware

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ChatWsAuth is for browser WebSocket: validates login status via Session Cookie.
// Note: Browser WebSocket cannot set custom headers, so we don't reuse UserAuth's New-Api-User validation.
func ChatWsAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		role := session.Get("role")
		id := session.Get("id")
		status := session.Get("status")
		group := session.Get("group")

		if username == nil || role == nil || id == nil || status == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无权进行此操作，未登录",
			})
			c.Abort()
			return
		}

		roleInt, ok := role.(int)
		if !ok || roleInt < common.RoleCommonUser {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "无权进行此操作，权限不足",
			})
			c.Abort()
			return
		}

		statusInt, ok := status.(int)
		if !ok || statusInt == common.UserStatusDisabled {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "用户已被封禁",
			})
			c.Abort()
			return
		}

		c.Set("username", username)
		c.Set("role", role)
		c.Set("id", id)
		c.Set("group", group)
		c.Set("user_group", group)
		c.Set("use_access_token", false)
		c.Next()
	}
}

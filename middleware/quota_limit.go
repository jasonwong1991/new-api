/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

For commercial licensing, please contact support@quantumnous.com
*/

package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/quota_limit"

	"github.com/gin-gonic/gin"
)

// QuotaLimit enforces optional per-user daily / weekly quota ceilings.
// It runs after authentication middleware (TokenAuth / UserAuth) so that
// `id`, `username` and `user_group` (when available) are already in context.
// Exempt lists (users or groups) bypass the check entirely.
func QuotaLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !quota_limit.IsEnabled() {
			c.Next()
			return
		}

		userId := c.GetInt("id")
		if userId <= 0 {
			c.Next()
			return
		}

		dailyLimit := quota_limit.GetDailyLimit()
		weeklyLimit := quota_limit.GetWeeklyLimit()
		// Bail out cheaply before any whitelist / DB work when both ceilings are disabled.
		if dailyLimit <= 0 && weeklyLimit <= 0 {
			c.Next()
			return
		}

		username := c.GetString("username")
		if quota_limit.IsUserWhitelisted(strconv.Itoa(userId), username) {
			c.Next()
			return
		}

		// Group whitelist — check both user_group (from UserAuth) and
		// the resolved using-group (set by distributor) if present.
		if g := c.GetString("user_group"); quota_limit.IsGroupWhitelisted(g) {
			c.Next()
			return
		}
		if v, ok := common.GetContextKey(c, constant.ContextKeyUsingGroup); ok {
			if s, ok := v.(string); ok && quota_limit.IsGroupWhitelisted(s) {
				c.Next()
				return
			}
		}
		// TokenAuth doesn't populate user_group — fall back to a cached lookup
		// (fromDB=true enables Redis/in-memory cache, avoiding a DB hit per request).
		if len(quota_limit.GetWhitelistGroupsRaw()) > 0 {
			if g, err := model.GetUserGroup(userId, true); err == nil && quota_limit.IsGroupWhitelisted(g) {
				c.Next()
				return
			}
		}

		daily, weekly, err := model.GetUserQuotaUsageWindows(userId, dailyLimit > 0, weeklyLimit > 0)
		if err != nil {
			// Fail-open: log and proceed so DB outage does not block traffic.
			common.SysError(fmt.Sprintf("quota_limit: query usage failed for user %d: %s", userId, err.Error()))
			c.Next()
			return
		}

		if dailyLimit > 0 && daily >= dailyLimit {
			abortQuotaLimit(c, i18n.T(c, "quota_limit.daily_reached", map[string]any{
				"Used":  daily,
				"Limit": dailyLimit,
			}))
			return
		}
		if weeklyLimit > 0 && weekly >= weeklyLimit {
			abortQuotaLimit(c, i18n.T(c, "quota_limit.weekly_reached", map[string]any{
				"Used":  weekly,
				"Limit": weeklyLimit,
			}))
			return
		}

		c.Next()
	}
}

func abortQuotaLimit(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error": gin.H{
			"message": common.MessageWithRequestId(message, c.GetString(common.RequestIdKey)),
			"type":    "new_api_error",
			"code":    "quota_limit_reached",
		},
	})
	c.Abort()
}

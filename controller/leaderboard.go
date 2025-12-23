package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LeaderboardEntry struct {
	Rank            int     `json:"rank"`
	DisplayName     string  `json:"display_name"`
	LinuxDOUsername string  `json:"linux_do_username"`
	LinuxDOAvatar   string  `json:"linux_do_avatar"`
	LinuxDOLevel    int     `json:"linux_do_level"`
	RequestCount    int     `json:"request_count"`
	UsedQuota       int     `json:"used_quota"`
	AmountUSD       float64 `json:"amount_usd"`
}

type LeaderboardResponse struct {
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
	MyRank      *LeaderboardEntry  `json:"my_rank,omitempty"`
}

func GetUsageLeaderboard(c *gin.Context) {
	users, err := model.GetUsageLeaderboard(100)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	entries := make([]LeaderboardEntry, 0, len(users))
	for i, user := range users {
		displayName := user.DisplayName
		if displayName == "" {
			displayName = "Anonymous"
		}

		entries = append(entries, LeaderboardEntry{
			Rank:            i + 1,
			DisplayName:     displayName,
			LinuxDOUsername: user.LinuxDOUsername,
			LinuxDOAvatar:   user.LinuxDOAvatar,
			LinuxDOLevel:    user.LinuxDOLevel,
			RequestCount:    user.RequestCount,
			UsedQuota:       user.UsedQuota,
			AmountUSD:       float64(user.UsedQuota) / common.QuotaPerUnit,
		})
	}

	response := LeaderboardResponse{
		Leaderboard: entries,
	}

	// Get current user's rank if logged in
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId != nil {
		rank, userData, err := model.GetUserRank(userId.(int))
		if err == nil && userData != nil {
			displayName := userData.DisplayName
			if displayName == "" {
				displayName = "Anonymous"
			}
			response.MyRank = &LeaderboardEntry{
				Rank:            rank,
				DisplayName:     displayName,
				LinuxDOUsername: userData.LinuxDOUsername,
				LinuxDOAvatar:   userData.LinuxDOAvatar,
				LinuxDOLevel:    userData.LinuxDOLevel,
				RequestCount:    userData.RequestCount,
				UsedQuota:       userData.UsedQuota,
				AmountUSD:       float64(userData.UsedQuota) / common.QuotaPerUnit,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    response,
	})
}

type ModelLeaderboardEntry struct {
	Rank         int     `json:"rank"`
	ModelName    string  `json:"model_name"`
	RequestCount int64   `json:"request_count"`
	TotalTokens  int64   `json:"total_tokens"`
	AmountUSD    float64 `json:"amount_usd"`
}

func GetModelLeaderboard(c *gin.Context) {
	models, err := model.GetModelUsageLeaderboard(100)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	entries := make([]ModelLeaderboardEntry, 0, len(models))
	for i, m := range models {
		entries = append(entries, ModelLeaderboardEntry{
			Rank:         i + 1,
			ModelName:    m.ModelName,
			RequestCount: m.RequestCount,
			TotalTokens:  m.TotalTokens,
			AmountUSD:    float64(m.TotalQuota) / common.QuotaPerUnit,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    entries,
	})
}

type BalanceLeaderboardEntry struct {
	Rank            int     `json:"rank"`
	DisplayName     string  `json:"display_name"`
	LinuxDOUsername string  `json:"linux_do_username"`
	LinuxDOAvatar   string  `json:"linux_do_avatar"`
	LinuxDOLevel    int     `json:"linux_do_level"`
	Quota           int     `json:"quota"`
	AmountUSD       float64 `json:"amount_usd"`
}

type BalanceLeaderboardResponse struct {
	Leaderboard []BalanceLeaderboardEntry `json:"leaderboard"`
	MyRank      *BalanceLeaderboardEntry  `json:"my_rank,omitempty"`
}

func GetBalanceLeaderboard(c *gin.Context) {
	users, err := model.GetBalanceLeaderboard(100)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	entries := make([]BalanceLeaderboardEntry, 0, len(users))
	for i, user := range users {
		displayName := user.DisplayName
		if displayName == "" {
			displayName = "Anonymous"
		}

		entries = append(entries, BalanceLeaderboardEntry{
			Rank:            i + 1,
			DisplayName:     displayName,
			LinuxDOUsername: user.LinuxDOUsername,
			LinuxDOAvatar:   user.LinuxDOAvatar,
			LinuxDOLevel:    user.LinuxDOLevel,
			Quota:           user.Quota,
			AmountUSD:       float64(user.Quota) / common.QuotaPerUnit,
		})
	}

	response := BalanceLeaderboardResponse{
		Leaderboard: entries,
	}

	session := sessions.Default(c)
	userId := session.Get("id")
	if userId != nil {
		rank, userData, err := model.GetUserBalanceRank(userId.(int))
		if err == nil && userData != nil {
			displayName := userData.DisplayName
			if displayName == "" {
				displayName = "Anonymous"
			}
			response.MyRank = &BalanceLeaderboardEntry{
				Rank:            rank,
				DisplayName:     displayName,
				LinuxDOUsername: userData.LinuxDOUsername,
				LinuxDOAvatar:   userData.LinuxDOAvatar,
				LinuxDOLevel:    userData.LinuxDOLevel,
				Quota:           userData.Quota,
				AmountUSD:       float64(userData.Quota) / common.QuotaPerUnit,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    response,
	})
}

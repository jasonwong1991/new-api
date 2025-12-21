package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

type LeaderboardEntry struct {
	Rank         int     `json:"rank"`
	DisplayName  string  `json:"display_name"`
	RequestCount int     `json:"request_count"`
	UsedQuota    int     `json:"used_quota"`
	AmountUSD    float64 `json:"amount_usd"`
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
			Rank:         i + 1,
			DisplayName:  displayName,
			RequestCount: user.RequestCount,
			UsedQuota:    user.UsedQuota,
			AmountUSD:    float64(user.UsedQuota) / common.QuotaPerUnit,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    entries,
	})
}

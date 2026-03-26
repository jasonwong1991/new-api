package controller

import (
	"net/http"
	"time"

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

type CacheInfo struct {
	LastUpdated  int64  `json:"last_updated"`
	RefreshHint  string `json:"refresh_hint"`
}

func getLeaderboardCacheInfo(period string) *CacheInfo {
	var lastUpdated time.Time
	var hint string

	switch period {
	case "24h":
		lastUpdated = model.GetLeaderboardCacheLastUpdated24h()
		hint = "每10分钟更新"
	case "7d", "14d", "30d":
		lastUpdated = model.GetLeaderboardCacheLastUpdated()
		hint = "每日零点更新"
	case "all":
		lastUpdated = model.GetLeaderboardCacheLastUpdated()
		hint = "每日零点更新"
	default:
		return nil
	}

	if lastUpdated.IsZero() {
		return &CacheInfo{
			LastUpdated: 0,
			RefreshHint: hint,
		}
	}
	return &CacheInfo{
		LastUpdated: lastUpdated.Unix(),
		RefreshHint: hint,
	}
}

func GetUsageLeaderboard(c *gin.Context) {
	period := c.DefaultQuery("period", "all")

	var entries []LeaderboardEntry
	var myRankEntry *LeaderboardEntry

	if period == "all" {
		// Try cache first
		cachedUsers, cached := model.GetCachedAllTimeUserLeaderboard()
		if cached {
			for i, user := range cachedUsers {
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
					RequestCount:    int(user.RequestCount),
					UsedQuota:       int(user.UsedQuota),
					AmountUSD:       float64(user.UsedQuota) / common.QuotaPerUnit,
				})
			}
		} else {
			// Fallback to direct query
			users, err := model.GetUsageLeaderboard(100)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}

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
		}

		session := sessions.Default(c)
		userId := session.Get("id")
		if userId != nil {
			rank, userData, err := model.GetUserRank(userId.(int))
			if err == nil && userData != nil {
				displayName := userData.DisplayName
				if displayName == "" {
					displayName = "Anonymous"
				}
				myRankEntry = &LeaderboardEntry{
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
	} else {
		// Cache-only for period-based queries to prevent thundering herd on logs table.
		// If cache is empty (cold start), return empty data; cache populates within 30s of boot.
		cachedUsers, cached := model.GetCachedUserLeaderboardByPeriod(period)
		if cached {
			for i, user := range cachedUsers {
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
					RequestCount:    int(user.RequestCount),
					UsedQuota:       int(user.UsedQuota),
					AmountUSD:       float64(user.UsedQuota) / common.QuotaPerUnit,
				})
			}

			// Calculate user rank from cached data instead of expensive per-request DB query
			session := sessions.Default(c)
			userId := session.Get("id")
			if userId != nil {
				// Lightweight lookup: only need username to match against cached entries
				var currentUser struct {
					Username string `gorm:"column:username"`
				}
				if err := model.DB.Table("users").Select("username").
					Where("id = ?", userId.(int)).First(&currentUser).Error; err == nil {
					for i, entry := range cachedUsers {
						if entry.Username == currentUser.Username {
							displayName := entry.DisplayName
							if displayName == "" {
								displayName = "Anonymous"
							}
							myRankEntry = &LeaderboardEntry{
								Rank:            i + 1,
								DisplayName:     displayName,
								LinuxDOUsername: entry.LinuxDOUsername,
								LinuxDOAvatar:   entry.LinuxDOAvatar,
								LinuxDOLevel:    entry.LinuxDOLevel,
								RequestCount:    int(entry.RequestCount),
								UsedQuota:       int(entry.UsedQuota),
								AmountUSD:       float64(entry.UsedQuota) / common.QuotaPerUnit,
							}
							break
						}
					}
				}
			}
		}
	}

	response := LeaderboardResponse{
		Leaderboard: entries,
		MyRank:      myRankEntry,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "",
		"data":       response,
		"cache_info": getLeaderboardCacheInfo(period),
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
	period := c.DefaultQuery("period", "all")

	var models []model.ModelLeaderboardEntry
	var err error

	if period == "all" {
		// Try cache first
		cachedModels, cached := model.GetCachedAllTimeModelLeaderboard()
		if cached {
			models = cachedModels
		} else {
			models, err = model.GetModelUsageLeaderboard(100)
		}
	} else {
		// Cache-only for period-based queries to prevent thundering herd on logs table.
		// If cache is empty (cold start), return empty data; cache populates within 30s of boot.
		cachedModels, cached := model.GetCachedModelLeaderboardByPeriod(period)
		if cached {
			models = cachedModels
		}
	}

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
		"success":    true,
		"message":    "",
		"data":       entries,
		"cache_info": getLeaderboardCacheInfo(period),
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
		"cache_info": &CacheInfo{
			LastUpdated: time.Now().Unix(),
			RefreshHint: "实时数据",
		},
	})
}

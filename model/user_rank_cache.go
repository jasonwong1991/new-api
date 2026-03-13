package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// UserRankInfo stores cached rank information for a user
type UserRankInfo struct {
	UsageRank   int
	BalanceRank int
}

// UserRankCache stores user rank information with periodic refresh
type UserRankCache struct {
	ranks       map[int]UserRankInfo // userId -> ranks
	lastUpdated time.Time
	mu          sync.RWMutex
}

var userRankCache = &UserRankCache{
	ranks: make(map[int]UserRankInfo),
}

const userRankCacheRefreshInterval = 30 * time.Minute

// InitUserRankCache initializes the user rank cache on startup
func InitUserRankCache() {
	common.SysLog("User rank cache will be initialized in background...")
	go func() {
		time.Sleep(10 * time.Second)
		common.SysLog("Starting background user rank cache initialization...")
		RefreshUserRankCache()
		common.SysLog("User rank cache initialized")
	}()
}

// StartUserRankCacheScheduler starts the periodic refresh scheduler
func StartUserRankCacheScheduler() {
	go func() {
		ticker := time.NewTicker(userRankCacheRefreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			RefreshUserRankCache()
		}
	}()
}

// RefreshUserRankCache refreshes all user rank data
func RefreshUserRankCache() {
	newRanks := make(map[int]UserRankInfo)

	// Get usage leaderboard (top users by used_quota)
	usageUsers, err := GetUsageLeaderboard(1000)
	if err != nil {
		common.SysLog("Failed to refresh usage leaderboard for rank cache: " + err.Error())
	} else {
		for i, user := range usageUsers {
			if user.Id > 0 {
				info := newRanks[user.Id]
				info.UsageRank = i + 1
				newRanks[user.Id] = info
			}
		}
	}

	// Get balance leaderboard (top users by quota)
	balanceUsers, err := GetBalanceLeaderboard(1000)
	if err != nil {
		common.SysLog("Failed to refresh balance leaderboard for rank cache: " + err.Error())
	} else {
		for i, user := range balanceUsers {
			if user.Id > 0 {
				info := newRanks[user.Id]
				info.BalanceRank = i + 1
				newRanks[user.Id] = info
			}
		}
	}

	userRankCache.mu.Lock()
	userRankCache.ranks = newRanks
	userRankCache.lastUpdated = time.Now()
	userRankCache.mu.Unlock()

	common.SysLog(fmt.Sprintf("User rank cache refreshed with %d users", len(newRanks)))
}

// GetUserRankFromCache returns cached rank info for a user
func GetUserRankFromCache(userId int) (usageRank int, balanceRank int) {
	if userId <= 0 {
		return 0, 0
	}

	userRankCache.mu.RLock()
	defer userRankCache.mu.RUnlock()

	if info, exists := userRankCache.ranks[userId]; exists {
		return info.UsageRank, info.BalanceRank
	}
	return 0, 0
}

// GetUserRankCacheLastUpdated returns the last update time
func GetUserRankCacheLastUpdated() time.Time {
	userRankCache.mu.RLock()
	defer userRankCache.mu.RUnlock()
	return userRankCache.lastUpdated
}

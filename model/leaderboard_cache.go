package model

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// Leaderboard cache structures
type LeaderboardCache struct {
	UserLeaderboard  map[string][]UsageLeaderboardEntry // period -> entries
	ModelLeaderboard map[string][]ModelLeaderboardEntry // period -> entries
	LastUpdated      time.Time
	mu               sync.RWMutex
}

var leaderboardCache = &LeaderboardCache{
	UserLeaderboard:  make(map[string][]UsageLeaderboardEntry),
	ModelLeaderboard: make(map[string][]ModelLeaderboardEntry),
}

// Periods that need caching (expensive queries)
var cachedPeriods = []string{"7d", "14d", "30d"}

// InitLeaderboardCache initializes the leaderboard cache on startup (async)
func InitLeaderboardCache() {
	common.SysLog("Leaderboard cache will be initialized in background...")
	go func() {
		// Delay initialization to allow service to start first
		time.Sleep(30 * time.Second)
		common.SysLog("Starting background leaderboard cache initialization...")
		RefreshLeaderboardCache()
		common.SysLog("Leaderboard cache initialized")
	}()
}

// RefreshLeaderboardCache refreshes all cached leaderboard data
func RefreshLeaderboardCache() {
	leaderboardCache.mu.Lock()
	defer leaderboardCache.mu.Unlock()

	for _, period := range cachedPeriods {
		// Refresh user leaderboard
		users, err := getUsageLeaderboardByPeriodDirect(period, 100)
		if err != nil {
			common.SysLog("Failed to refresh user leaderboard for period " + period + ": " + err.Error())
		} else {
			leaderboardCache.UserLeaderboard[period] = users
		}

		// Refresh model leaderboard
		models, err := getModelLeaderboardByPeriodDirect(period, 100)
		if err != nil {
			common.SysLog("Failed to refresh model leaderboard for period " + period + ": " + err.Error())
		} else {
			leaderboardCache.ModelLeaderboard[period] = models
		}
	}

	leaderboardCache.LastUpdated = time.Now()
	common.SysLog("Leaderboard cache refreshed at " + leaderboardCache.LastUpdated.Format("2006-01-02 15:04:05"))
}

// StartLeaderboardCacheScheduler starts the daily refresh scheduler
func StartLeaderboardCacheScheduler() {
	go func() {
		for {
			now := time.Now()
			// Calculate next midnight
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := next.Sub(now)

			common.SysLog("Next leaderboard cache refresh scheduled at " + next.Format("2006-01-02 15:04:05"))
			time.Sleep(duration)

			common.SysLog("Starting scheduled leaderboard cache refresh...")
			RefreshLeaderboardCache()
		}
	}()
}

// GetCachedUserLeaderboardByPeriod returns cached user leaderboard for a period
func GetCachedUserLeaderboardByPeriod(period string) ([]UsageLeaderboardEntry, bool) {
	// Only use cache for expensive periods
	if !isPeriodCached(period) {
		return nil, false
	}

	leaderboardCache.mu.RLock()
	defer leaderboardCache.mu.RUnlock()

	entries, exists := leaderboardCache.UserLeaderboard[period]
	if !exists || len(entries) == 0 {
		return nil, false
	}

	// Return a copy to prevent race conditions
	result := make([]UsageLeaderboardEntry, len(entries))
	copy(result, entries)
	return result, true
}

// GetCachedModelLeaderboardByPeriod returns cached model leaderboard for a period
func GetCachedModelLeaderboardByPeriod(period string) ([]ModelLeaderboardEntry, bool) {
	// Only use cache for expensive periods
	if !isPeriodCached(period) {
		return nil, false
	}

	leaderboardCache.mu.RLock()
	defer leaderboardCache.mu.RUnlock()

	entries, exists := leaderboardCache.ModelLeaderboard[period]
	if !exists || len(entries) == 0 {
		return nil, false
	}

	// Return a copy to prevent race conditions
	result := make([]ModelLeaderboardEntry, len(entries))
	copy(result, entries)
	return result, true
}

// GetLeaderboardCacheLastUpdated returns the last update time
func GetLeaderboardCacheLastUpdated() time.Time {
	leaderboardCache.mu.RLock()
	defer leaderboardCache.mu.RUnlock()
	return leaderboardCache.LastUpdated
}

func isPeriodCached(period string) bool {
	for _, p := range cachedPeriods {
		if p == period {
			return true
		}
	}
	return false
}

// Direct database query functions (bypassing cache)
func getUsageLeaderboardByPeriodDirect(period string, limit int) ([]UsageLeaderboardEntry, error) {
	var entries []UsageLeaderboardEntry
	startTimestamp := getPeriodTimestamp(period)

	query := LOG_DB.Table("logs").
		Select("username, COUNT(*) as request_count, SUM(quota) as used_quota").
		Where("type = ?", LogTypeConsume).
		Where("username != ''")

	if startTimestamp > 0 {
		query = query.Where("created_at >= ?", startTimestamp)
	}

	if len(common.LeaderboardHiddenUsers) > 0 {
		query = query.Where("username NOT IN ?", common.LeaderboardHiddenUsers)
	}

	var logEntries []struct {
		Username     string `gorm:"column:username"`
		RequestCount int64  `gorm:"column:request_count"`
		UsedQuota    int64  `gorm:"column:used_quota"`
	}

	err := query.Group("username").
		Order("used_quota DESC").
		Limit(limit).
		Find(&logEntries).Error
	if err != nil {
		return nil, err
	}

	if len(logEntries) == 0 {
		return entries, nil
	}

	usernames := make([]string, len(logEntries))
	for i, e := range logEntries {
		usernames[i] = e.Username
	}

	var users []struct {
		Username        string `gorm:"column:username"`
		DisplayName     string `gorm:"column:display_name"`
		LinuxDOUsername string `gorm:"column:linux_do_username"`
		LinuxDOAvatar   string `gorm:"column:linux_do_avatar"`
		LinuxDOLevel    int    `gorm:"column:linux_do_level"`
	}
	DB.Table("users").
		Select("username, display_name, linux_do_username, linux_do_avatar, linux_do_level").
		Where("username IN ?", usernames).
		Find(&users)

	userMap := make(map[string]struct {
		DisplayName     string
		LinuxDOUsername string
		LinuxDOAvatar   string
		LinuxDOLevel    int
	})
	for _, u := range users {
		userMap[u.Username] = struct {
			DisplayName     string
			LinuxDOUsername string
			LinuxDOAvatar   string
			LinuxDOLevel    int
		}{u.DisplayName, u.LinuxDOUsername, u.LinuxDOAvatar, u.LinuxDOLevel}
	}

	for _, e := range logEntries {
		userInfo := userMap[e.Username]
		entries = append(entries, UsageLeaderboardEntry{
			Username:        e.Username,
			DisplayName:     userInfo.DisplayName,
			LinuxDOUsername: userInfo.LinuxDOUsername,
			LinuxDOAvatar:   userInfo.LinuxDOAvatar,
			LinuxDOLevel:    userInfo.LinuxDOLevel,
			RequestCount:    e.RequestCount,
			UsedQuota:       e.UsedQuota,
		})
	}

	return entries, nil
}

func getModelLeaderboardByPeriodDirect(period string, limit int) ([]ModelLeaderboardEntry, error) {
	var entries []ModelLeaderboardEntry
	startTimestamp := getPeriodTimestamp(period)

	query := LOG_DB.Table("logs").
		Select("model_name, COUNT(*) as request_count, SUM(prompt_tokens) + SUM(completion_tokens) as total_tokens, SUM(quota) as total_quota").
		Where("type = ?", LogTypeConsume).
		Where("model_name != ''")

	if startTimestamp > 0 {
		query = query.Where("created_at >= ?", startTimestamp)
	}

	err := query.Group("model_name").
		Order("request_count DESC").
		Limit(limit).
		Find(&entries).Error

	return entries, err
}

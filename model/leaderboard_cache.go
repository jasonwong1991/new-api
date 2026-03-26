package model

import (
	"context"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// Leaderboard cache structures
type LeaderboardCache struct {
	UserLeaderboard  map[string][]UsageLeaderboardEntry // period -> entries (24h, 7d, 14d, 30d)
	ModelLeaderboard map[string][]ModelLeaderboardEntry // period -> entries (24h, 7d, 14d, 30d)
	// "all" period uses different data sources (users table / quota_data table)
	AllTimeUserLeaderboard  []UsageLeaderboardEntry
	AllTimeModelLeaderboard []ModelLeaderboardEntry
	LastUpdated             time.Time
	LastUpdated24h          time.Time
	mu                      sync.RWMutex
}

var leaderboardCache = &LeaderboardCache{
	UserLeaderboard:  make(map[string][]UsageLeaderboardEntry),
	ModelLeaderboard: make(map[string][]ModelLeaderboardEntry),
}

// Periods that need caching (expensive queries)
var cachedPeriods = []string{"24h", "7d", "14d", "30d"}

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

	// Refresh "all" period (uses different data sources)
	refreshAllTimeLeaderboard()

	leaderboardCache.LastUpdated = time.Now()
	leaderboardCache.LastUpdated24h = time.Now()
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

// StartLeaderboard24hCacheScheduler starts a more frequent refresh for 24h period (every 10 minutes)
func StartLeaderboard24hCacheScheduler() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			refresh24hLeaderboard()
		}
	}()
}

// refreshAllTimeLeaderboard refreshes the "all" period data (called within lock)
func refreshAllTimeLeaderboard() {
	// User "all" leaderboard: use users table (consistent with GetUsageLeaderboard fallback).
	// The users table stores cumulative used_quota that survives log cleanup.
	rawUsers, err := GetUsageLeaderboard(100)
	if err != nil {
		common.SysLog("Failed to refresh all-time user leaderboard: " + err.Error())
	} else {
		entries := make([]UsageLeaderboardEntry, len(rawUsers))
		for i, u := range rawUsers {
			entries[i] = UsageLeaderboardEntry{
				DisplayName:    u.DisplayName,
				LinuxDOUsername: u.LinuxDOUsername,
				LinuxDOAvatar:  u.LinuxDOAvatar,
				LinuxDOLevel:   u.LinuxDOLevel,
				RequestCount:   int64(u.RequestCount),
				UsedQuota:      int64(u.UsedQuota),
			}
		}
		leaderboardCache.AllTimeUserLeaderboard = entries
	}

	// Model "all" leaderboard: uses quota_data table (persistent aggregated data)
	models, err := GetModelUsageLeaderboard(100)
	if err != nil {
		common.SysLog("Failed to refresh all-time model leaderboard: " + err.Error())
	} else {
		leaderboardCache.AllTimeModelLeaderboard = models
	}
}

// refresh24hLeaderboard refreshes only the 24h period cache
func refresh24hLeaderboard() {
	leaderboardCache.mu.Lock()
	defer leaderboardCache.mu.Unlock()

	users, err := getUsageLeaderboardByPeriodDirect("24h", 100)
	if err != nil {
		common.SysLog("Failed to refresh 24h user leaderboard: " + err.Error())
	} else {
		leaderboardCache.UserLeaderboard["24h"] = users
	}

	models, err := getModelLeaderboardByPeriodDirect("24h", 100)
	if err != nil {
		common.SysLog("Failed to refresh 24h model leaderboard: " + err.Error())
	} else {
		leaderboardCache.ModelLeaderboard["24h"] = models
	}

	leaderboardCache.LastUpdated24h = time.Now()
	common.SysLog("24h leaderboard cache refreshed")
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

// GetLeaderboardCacheLastUpdated24h returns the last update time for 24h cache
func GetLeaderboardCacheLastUpdated24h() time.Time {
	leaderboardCache.mu.RLock()
	defer leaderboardCache.mu.RUnlock()
	return leaderboardCache.LastUpdated24h
}

// GetCachedAllTimeUserLeaderboard returns cached all-time user leaderboard
func GetCachedAllTimeUserLeaderboard() ([]UsageLeaderboardEntry, bool) {
	leaderboardCache.mu.RLock()
	defer leaderboardCache.mu.RUnlock()

	if len(leaderboardCache.AllTimeUserLeaderboard) == 0 {
		return nil, false
	}

	result := make([]UsageLeaderboardEntry, len(leaderboardCache.AllTimeUserLeaderboard))
	copy(result, leaderboardCache.AllTimeUserLeaderboard)
	return result, true
}

// GetCachedAllTimeModelLeaderboard returns cached all-time model leaderboard
func GetCachedAllTimeModelLeaderboard() ([]ModelLeaderboardEntry, bool) {
	leaderboardCache.mu.RLock()
	defer leaderboardCache.mu.RUnlock()

	if len(leaderboardCache.AllTimeModelLeaderboard) == 0 {
		return nil, false
	}

	result := make([]ModelLeaderboardEntry, len(leaderboardCache.AllTimeModelLeaderboard))
	copy(result, leaderboardCache.AllTimeModelLeaderboard)
	return result, true
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := LOG_DB.WithContext(ctx).Table("logs").
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
	DB.WithContext(ctx).Table("users").
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := LOG_DB.WithContext(ctx).Table("logs").
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

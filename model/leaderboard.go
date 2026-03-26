package model

import (
	"context"
	"time"

	"github.com/QuantumNous/new-api/common"
)

type ModelLeaderboardEntry struct {
	ModelName    string `json:"model_name"`
	RequestCount int64  `json:"request_count"`
	TotalTokens  int64  `json:"total_tokens"`
	TotalQuota   int64  `json:"total_quota"`
}

func GetModelUsageLeaderboard(limit int) ([]ModelLeaderboardEntry, error) {
	var entries []ModelLeaderboardEntry
	err := DB.Table("quota_data").
		Select("model_name, SUM(count) as request_count, SUM(token_used) as total_tokens, SUM(quota) as total_quota").
		Where("model_name != ''").
		Group("model_name").
		Order("request_count DESC").
		Limit(limit).
		Find(&entries).Error
	return entries, err
}

type UsageLeaderboardEntry struct {
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	LinuxDOUsername  string `json:"linux_do_username"`
	LinuxDOAvatar   string `json:"linux_do_avatar"`
	LinuxDOLevel    int    `json:"linux_do_level"`
	RequestCount    int64  `json:"request_count"`
	UsedQuota       int64  `json:"used_quota"`
}

func getPeriodTimestamp(period string) int64 {
	now := time.Now().Unix()
	switch period {
	case "24h":
		return now - 24*60*60
	case "7d":
		return now - 7*24*60*60
	case "14d":
		return now - 14*24*60*60
	case "30d":
		return now - 30*24*60*60
	default:
		return 0
	}
}

func GetUsageLeaderboardByPeriod(period string, limit int) ([]UsageLeaderboardEntry, error) {
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

func isLeaderboardHiddenUser(username string) bool {
	for _, u := range common.LeaderboardHiddenUsers {
		if u == username {
			return true
		}
	}
	return false
}

func GetUserRankByPeriod(userId int, period string) (int, *UsageLeaderboardEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var user User
	err := DB.WithContext(ctx).Select("id, username, display_name, linux_do_username, linux_do_avatar, linux_do_level").
		Where("id = ?", userId).
		First(&user).Error
	if err != nil {
		return 0, nil, err
	}

	if isLeaderboardHiddenUser(user.Username) {
		return 0, nil, nil
	}

	startTimestamp := getPeriodTimestamp(period)

	var userStats struct {
		RequestCount int64 `gorm:"column:request_count"`
		UsedQuota    int64 `gorm:"column:used_quota"`
	}
	query := LOG_DB.WithContext(ctx).Table("logs").
		Select("COUNT(*) as request_count, SUM(quota) as used_quota").
		Where("type = ?", LogTypeConsume).
		Where("username = ?", user.Username)
	if startTimestamp > 0 {
		query = query.Where("created_at >= ?", startTimestamp)
	}
	query.Scan(&userStats)

	if userStats.UsedQuota <= 0 {
		return 0, nil, nil
	}

	rankQuery := LOG_DB.WithContext(ctx).Table("logs").
		Select("username, SUM(quota) as total_quota").
		Where("type = ?", LogTypeConsume).
		Where("username != ''")
	if startTimestamp > 0 {
		rankQuery = rankQuery.Where("created_at >= ?", startTimestamp)
	}
	if len(common.LeaderboardHiddenUsers) > 0 {
		rankQuery = rankQuery.Where("username NOT IN ?", common.LeaderboardHiddenUsers)
	}

	var rank int64
	subQuery := rankQuery.Group("username").Having("SUM(quota) > ?", userStats.UsedQuota)
	LOG_DB.WithContext(ctx).Table("(?) as t", subQuery).Count(&rank)

	return int(rank + 1), &UsageLeaderboardEntry{
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		LinuxDOUsername: user.LinuxDOUsername,
		LinuxDOAvatar:   user.LinuxDOAvatar,
		LinuxDOLevel:    user.LinuxDOLevel,
		RequestCount:    userStats.RequestCount,
		UsedQuota:       userStats.UsedQuota,
	}, nil
}

func GetModelLeaderboardByPeriod(period string, limit int) ([]ModelLeaderboardEntry, error) {
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

type LeaderboardUser struct {
	Id              int    `json:"id" gorm:"column:id"`
	DisplayName     string `json:"display_name"`
	LinuxDOUsername string `json:"linux_do_username"`
	LinuxDOAvatar   string `json:"linux_do_avatar"`
	LinuxDOLevel    int    `json:"linux_do_level"`
	RequestCount    int    `json:"request_count"`
	UsedQuota       int    `json:"used_quota"`
}

func GetUsageLeaderboard(limit int) ([]LeaderboardUser, error) {
	var users []LeaderboardUser
	query := DB.Model(&User{}).
		Select("id, display_name, linux_do_username, linux_do_avatar, linux_do_level, request_count, used_quota").
		Where("status = ?", common.UserStatusEnabled).
		Where("role != ?", common.RoleRootUser).
		Where("used_quota > 0")
	if len(common.LeaderboardHiddenUsers) > 0 {
		query = query.Where("username NOT IN ?", common.LeaderboardHiddenUsers)
	}
	err := query.Order("used_quota DESC").
		Limit(limit).
		Find(&users).Error
	return users, err
}

func GetUserRank(userId int) (int, *LeaderboardUser, error) {
	var user User
	err := DB.Select("id, username, display_name, linux_do_username, linux_do_avatar, linux_do_level, request_count, used_quota").
		Where("id = ?", userId).
		First(&user).Error
	if err != nil {
		return 0, nil, err
	}
	if user.UsedQuota <= 0 || isLeaderboardHiddenUser(user.Username) {
		return 0, nil, nil
	}
	query := DB.Model(&User{}).
		Where("status = ?", common.UserStatusEnabled).
		Where("role != ?", common.RoleRootUser).
		Where("used_quota > ?", user.UsedQuota)
	if len(common.LeaderboardHiddenUsers) > 0 {
		query = query.Where("username NOT IN ?", common.LeaderboardHiddenUsers)
	}
	var rank int64
	err = query.Count(&rank).Error
	if err != nil {
		return 0, nil, err
	}
	return int(rank + 1), &LeaderboardUser{
		DisplayName:     user.DisplayName,
		LinuxDOUsername: user.LinuxDOUsername,
		LinuxDOAvatar:   user.LinuxDOAvatar,
		LinuxDOLevel:    user.LinuxDOLevel,
		RequestCount:    user.RequestCount,
		UsedQuota:       user.UsedQuota,
	}, nil
}

type BalanceLeaderboardUser struct {
	Id              int    `json:"id" gorm:"column:id"`
	DisplayName     string `json:"display_name"`
	LinuxDOUsername string `json:"linux_do_username"`
	LinuxDOAvatar   string `json:"linux_do_avatar"`
	LinuxDOLevel    int    `json:"linux_do_level"`
	Quota           int    `json:"quota"`
}

func GetBalanceLeaderboard(limit int) ([]BalanceLeaderboardUser, error) {
	var users []BalanceLeaderboardUser
	query := DB.Model(&User{}).
		Select("id, display_name, linux_do_username, linux_do_avatar, linux_do_level, quota").
		Where("status = ?", common.UserStatusEnabled).
		Where("role != ?", common.RoleRootUser).
		Where("quota > 0")
	if len(common.LeaderboardHiddenUsers) > 0 {
		query = query.Where("username NOT IN ?", common.LeaderboardHiddenUsers)
	}
	err := query.Order("quota DESC").
		Limit(limit).
		Find(&users).Error
	return users, err
}

func GetUserBalanceRank(userId int) (int, *BalanceLeaderboardUser, error) {
	var user User
	err := DB.Select("id, username, display_name, linux_do_username, linux_do_avatar, linux_do_level, quota").
		Where("id = ?", userId).
		First(&user).Error
	if err != nil {
		return 0, nil, err
	}
	if user.Quota <= 0 || isLeaderboardHiddenUser(user.Username) {
		return 0, nil, nil
	}
	query := DB.Model(&User{}).
		Where("status = ?", common.UserStatusEnabled).
		Where("role != ?", common.RoleRootUser).
		Where("quota > ?", user.Quota)
	if len(common.LeaderboardHiddenUsers) > 0 {
		query = query.Where("username NOT IN ?", common.LeaderboardHiddenUsers)
	}
	var rank int64
	err = query.Count(&rank).Error
	if err != nil {
		return 0, nil, err
	}
	return int(rank + 1), &BalanceLeaderboardUser{
		DisplayName:     user.DisplayName,
		LinuxDOUsername: user.LinuxDOUsername,
		LinuxDOAvatar:   user.LinuxDOAvatar,
		LinuxDOLevel:    user.LinuxDOLevel,
		Quota:           user.Quota,
	}, nil
}

type BannedUser struct {
	Id               int    `json:"id"`
	DisplayName      string `json:"display_name"`
	LinuxDOUsername  string `json:"linux_do_username"`
	LinuxDOAvatar    string `json:"linux_do_avatar"`
	BanReason        string `json:"ban_reason"`
	HasPendingAppeal bool   `json:"has_pending_appeal"`
}

func GetBannedUsers() ([]BannedUser, error) {
	var users []struct {
		Id              int    `json:"id"`
		DisplayName     string `json:"display_name"`
		LinuxDOUsername string `json:"linux_do_username"`
		LinuxDOAvatar   string `json:"linux_do_avatar"`
		Remark          string `json:"remark"`
	}
	err := DB.Model(&User{}).
		Select("id, display_name, linux_do_username, linux_do_avatar, remark").
		Where("status = ?", common.UserStatusDisabled).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make([]BannedUser, len(users))
	for i, u := range users {
		result[i] = BannedUser{
			Id:               u.Id,
			DisplayName:      u.DisplayName,
			LinuxDOUsername:  u.LinuxDOUsername,
			LinuxDOAvatar:    u.LinuxDOAvatar,
			BanReason:        u.Remark,
			HasPendingAppeal: HasPendingAppeal(u.Id),
		}
	}
	return result, nil
}

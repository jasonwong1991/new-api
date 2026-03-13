package model

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"

	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
)

type Log struct {
	Id               int    `json:"id" gorm:"index:idx_created_at_id,priority:1"`
	UserId           int    `json:"user_id" gorm:"index"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;index:idx_created_at_id,priority:2;index:idx_created_at_type"`
	Type             int    `json:"type" gorm:"index:idx_created_at_type"`
	Content          string `json:"content"`
	Username         string `json:"username" gorm:"index;index:index_username_model_name,priority:2;default:''"`
	TokenName        string `json:"token_name" gorm:"index;default:''"`
	ModelName        string `json:"model_name" gorm:"index;index:index_username_model_name,priority:1;default:''"`
	Quota            int    `json:"quota" gorm:"default:0"`
	PromptTokens     int    `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int    `json:"completion_tokens" gorm:"default:0"`
	UseTime          int    `json:"use_time" gorm:"default:0"`
	IsStream         bool   `json:"is_stream"`
	ChannelId        int    `json:"channel" gorm:"index"`
	ChannelName      string `json:"channel_name" gorm:"->"`
	TokenId          int    `json:"token_id" gorm:"default:0;index"`
	Group            string `json:"group" gorm:"index"`
	Ip               string `json:"ip" gorm:"index;default:''"`
	Other            string `json:"other"`
}

// don't use iota, avoid change log type value
const (
	LogTypeUnknown = 0
	LogTypeTopup   = 1
	LogTypeConsume = 2
	LogTypeManage  = 3
	LogTypeSystem  = 4
	LogTypeError   = 5
	LogTypeRefund  = 6
)

var (
	logBuffer     []*Log
	logBufferLock sync.Mutex
	logBufferSize = 100 // flush when buffer reaches this size
)

func init() {
	logBuffer = make([]*Log, 0, logBufferSize)
	go logBatchWriter()
}

func logBatchWriter() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		flushLogBuffer()
	}
}

func flushLogBuffer() {
	if LOG_DB == nil {
		return
	}
	logBufferLock.Lock()
	if len(logBuffer) == 0 {
		logBufferLock.Unlock()
		return
	}
	logs := logBuffer
	logBuffer = make([]*Log, 0, logBufferSize)
	logBufferLock.Unlock()

	// Batch insert in chunks of 100
	batchSize := 100
	for i := 0; i < len(logs); i += batchSize {
		end := i + batchSize
		if end > len(logs) {
			end = len(logs)
		}
		if err := LOG_DB.CreateInBatches(logs[i:end], batchSize).Error; err != nil {
			common.SysLog("failed to batch insert logs: " + err.Error())
			// Fallback: try individual inserts for failed batch
			for _, log := range logs[i:end] {
				if err := LOG_DB.Create(log).Error; err != nil {
					common.SysLog("failed to insert log: " + err.Error())
				}
			}
		}
	}
}

// FlushLogBuffer flushes any buffered logs to the database.
// Must be called during graceful shutdown before closing DB connections.
func FlushLogBuffer() {
	flushLogBuffer()
}

func bufferLog(log *Log) {
	logBufferLock.Lock()
	logBuffer = append(logBuffer, log)
	shouldFlush := len(logBuffer) >= logBufferSize
	logBufferLock.Unlock()

	if shouldFlush {
		// Use gopool for async flush to avoid blocking the caller
		gopool.Go(func() {
			flushLogBuffer()
		})
	}
}

func formatUserLogs(logs []*Log) {
	for i := range logs {
		logs[i].ChannelName = ""
		var otherMap map[string]interface{}
		otherMap, _ = common.StrToMap(logs[i].Other)
		if otherMap != nil {
			// delete admin
			delete(otherMap, "admin_info")
		}
		logs[i].Other = common.MapToJsonStr(otherMap)
		logs[i].Id = logs[i].Id % 1024
	}
}

func GetLogByKey(key string) (logs []*Log, err error) {
	if os.Getenv("LOG_SQL_DSN") != "" {
		var tk Token
		if err = DB.Model(&Token{}).Where(logKeyCol+"=?", strings.TrimPrefix(key, "sk-")).First(&tk).Error; err != nil {
			return nil, err
		}
		err = LOG_DB.Model(&Log{}).Where("token_id=?", tk.Id).Find(&logs).Error
	} else {
		err = LOG_DB.Joins("left join tokens on tokens.id = logs.token_id").Where("tokens.key = ?", strings.TrimPrefix(key, "sk-")).Find(&logs).Error
	}
	formatUserLogs(logs)
	return logs, err
}

func RecordLog(userId int, logType int, content string) {
	if logType == LogTypeConsume && !common.LogConsumeEnabled {
		return
	}
	username, _ := GetUsernameById(userId, false)
	log := &Log{
		UserId:    userId,
		Username:  username,
		CreatedAt: common.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	bufferLog(log)
}

// RecordErrorLog writes error logs synchronously (not buffered) to ensure
// critical error information is persisted immediately and not lost on crash.
func RecordErrorLog(c *gin.Context, userId int, channelId int, modelName string, tokenName string, content string, tokenId int, useTimeSeconds int,
	isStream bool, group string, other map[string]interface{}) {
	logger.LogInfo(c, fmt.Sprintf("record error log: userId=%d, channelId=%d, modelName=%s, tokenName=%s, content=%s", userId, channelId, modelName, tokenName, content))
	username := c.GetString("username")
	otherStr := common.MapToJsonStr(other)
	// 判断是否需要记录 IP
	needRecordIp := false
	if settingMap, err := GetUserSetting(userId, false); err == nil {
		if settingMap.RecordIpLog {
			needRecordIp = true
		}
	}
	log := &Log{
		UserId:           userId,
		Username:         username,
		CreatedAt:        common.GetTimestamp(),
		Type:             LogTypeError,
		Content:          content,
		PromptTokens:     0,
		CompletionTokens: 0,
		TokenName:        tokenName,
		ModelName:        modelName,
		Quota:            0,
		ChannelId:        channelId,
		TokenId:          tokenId,
		UseTime:          useTimeSeconds,
		IsStream:         isStream,
		Group:            group,
		Ip: func() string {
			if needRecordIp {
				return c.ClientIP()
			}
			return ""
		}(),
		Other: otherStr,
	}
	err := LOG_DB.Create(log).Error
	if err != nil {
		logger.LogError(c, "failed to record log: "+err.Error())
	}
}

type RecordConsumeLogParams struct {
	ChannelId        int                    `json:"channel_id"`
	PromptTokens     int                    `json:"prompt_tokens"`
	CompletionTokens int                    `json:"completion_tokens"`
	ModelName        string                 `json:"model_name"`
	TokenName        string                 `json:"token_name"`
	Quota            int                    `json:"quota"`
	Content          string                 `json:"content"`
	TokenId          int                    `json:"token_id"`
	UseTimeSeconds   int                    `json:"use_time_seconds"`
	IsStream         bool                   `json:"is_stream"`
	Group            string                 `json:"group"`
	Other            map[string]interface{} `json:"other"`
}

func RecordConsumeLog(c *gin.Context, userId int, params RecordConsumeLogParams) {
	if !common.LogConsumeEnabled {
		return
	}
	logger.LogInfo(c, fmt.Sprintf("record consume log: userId=%d, params=%s", userId, common.GetJsonString(params)))
	username := c.GetString("username")
	otherStr := common.MapToJsonStr(params.Other)
	// 判断是否需要记录 IP
	needRecordIp := false
	if settingMap, err := GetUserSetting(userId, false); err == nil {
		if settingMap.RecordIpLog {
			needRecordIp = true
		}
	}
	log := &Log{
		UserId:           userId,
		Username:         username,
		CreatedAt:        common.GetTimestamp(),
		Type:             LogTypeConsume,
		Content:          params.Content,
		PromptTokens:     params.PromptTokens,
		CompletionTokens: params.CompletionTokens,
		TokenName:        params.TokenName,
		ModelName:        params.ModelName,
		Quota:            params.Quota,
		ChannelId:        params.ChannelId,
		TokenId:          params.TokenId,
		UseTime:          params.UseTimeSeconds,
		IsStream:         params.IsStream,
		Group:            params.Group,
		Ip: func() string {
			if needRecordIp {
				return c.ClientIP()
			}
			return ""
		}(),
		Other: otherStr,
	}
	bufferLog(log)
	if common.DataExportEnabled {
		gopool.Go(func() {
			LogQuotaData(userId, username, params.ModelName, params.Quota, common.GetTimestamp(), params.PromptTokens+params.CompletionTokens)
		})
	}
}

func GetAllLogs(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int, group string) (logs []*Log, total int64, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB
	} else {
		tx = LOG_DB.Where("logs.type = ?", logType)
	}

	if modelName != "" {
		tx = tx.Where("logs.model_name like ?", modelName)
	}
	if username != "" {
		var usernames []string
		DB.Model(&User{}).
			Where("username = ? OR linux_do_username = ? OR display_name = ?", username, username, username).
			Pluck("username", &usernames)
		if len(usernames) > 0 {
			tx = tx.Where("logs.username IN ?", usernames)
		} else {
			tx = tx.Where("logs.username = ?", username)
		}
	}
	if tokenName != "" {
		tx = tx.Where("logs.token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("logs.created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("logs.created_at <= ?", endTimestamp)
	}
	if channel != 0 {
		tx = tx.Where("logs.channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where("logs."+logGroupCol+" = ?", group)
	}
	err = tx.Model(&Log{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = tx.Order("logs.id desc").Limit(num).Offset(startIdx).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	channelIds := types.NewSet[int]()
	for _, log := range logs {
		if log.ChannelId != 0 {
			channelIds.Add(log.ChannelId)
		}
	}

	if channelIds.Len() > 0 {
		var channels []struct {
			Id   int    `gorm:"column:id"`
			Name string `gorm:"column:name"`
		}
		if err = DB.Table("channels").Select("id, name").Where("id IN ?", channelIds.Items()).Find(&channels).Error; err != nil {
			return logs, total, err
		}
		channelMap := make(map[int]string, len(channels))
		for _, channel := range channels {
			channelMap[channel.Id] = channel.Name
		}
		for i := range logs {
			logs[i].ChannelName = channelMap[logs[i].ChannelId]
		}
	}

	return logs, total, err
}

func GetUserLogs(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int, group string) (logs []*Log, total int64, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB.Where("logs.user_id = ?", userId)
	} else {
		tx = LOG_DB.Where("logs.user_id = ? and logs.type = ?", userId, logType)
	}

	if modelName != "" {
		tx = tx.Where("logs.model_name like ?", modelName)
	}
	if tokenName != "" {
		tx = tx.Where("logs.token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("logs.created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("logs.created_at <= ?", endTimestamp)
	}
	if group != "" {
		tx = tx.Where("logs."+logGroupCol+" = ?", group)
	}
	err = tx.Model(&Log{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = tx.Order("logs.id desc").Limit(num).Offset(startIdx).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	formatUserLogs(logs)
	return logs, total, err
}

func SearchAllLogs(keyword string) (logs []*Log, err error) {
	err = LOG_DB.Where("type = ? or content LIKE ?", keyword, keyword+"%").Order("id desc").Limit(common.MaxRecentItems).Find(&logs).Error
	return logs, err
}

func SearchUserLogs(userId int, keyword string) (logs []*Log, err error) {
	err = LOG_DB.Where("user_id = ? and type = ?", userId, keyword).Order("id desc").Limit(common.MaxRecentItems).Find(&logs).Error
	formatUserLogs(logs)
	return logs, err
}

type Stat struct {
	Quota int `json:"quota"`
	Rpm   int `json:"rpm"`
	Tpm   int `json:"tpm"`
}

func SumUsedQuota(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int, group string) (stat Stat) {
	// Resolve username aliases once (used by both fast path and normal path)
	var resolvedUsernames []string
	if username != "" {
		DB.Model(&User{}).
			Where("username = ? OR linux_do_username = ? OR display_name = ?", username, username, username).
			Pluck("username", &resolvedUsernames)
	}

	// Fast path: all-time total without detailed filters.
	// Use users.used_quota (O(users)) instead of SUM on logs table (O(millions of logs)).
	// This preserves the ability to query all-time totals without the performance penalty.
	if startTimestamp == 0 && endTimestamp == 0 && modelName == "" && tokenName == "" && channel == 0 && group == "" && logType == LogTypeConsume {
		if username != "" {
			if len(resolvedUsernames) > 0 {
				DB.Model(&User{}).Where("username IN ?", resolvedUsernames).
					Select("COALESCE(SUM(used_quota), 0)").Scan(&stat.Quota)
			} else {
				DB.Model(&User{}).Where("username = ?", username).
					Select("COALESCE(SUM(used_quota), 0)").Scan(&stat.Quota)
			}
		} else {
			DB.Model(&User{}).Select("COALESCE(SUM(used_quota), 0)").Scan(&stat.Quota)
		}
		// RPM/TPM still from logs (already bounded to 60 seconds, always fast)
		rpmTpmQuery := LOG_DB.Table("logs").Select("count(*) rpm, sum(prompt_tokens) + sum(completion_tokens) tpm")
		rpmTpmQuery = rpmTpmQuery.Where("type = ?", LogTypeConsume)
		rpmTpmQuery = rpmTpmQuery.Where("created_at >= ?", time.Now().Add(-60*time.Second).Unix())
		if username != "" {
			if len(resolvedUsernames) > 0 {
				rpmTpmQuery = rpmTpmQuery.Where("username IN ?", resolvedUsernames)
			} else {
				rpmTpmQuery = rpmTpmQuery.Where("username = ?", username)
			}
		}
		rpmTpmQuery.Scan(&stat)
		return stat
	}

	// Normal path: time-bounded or filtered queries use logs table with index support
	tx := LOG_DB.Table("logs").Select("sum(quota) quota")
	rpmTpmQuery := LOG_DB.Table("logs").Select("count(*) rpm, sum(prompt_tokens) + sum(completion_tokens) tpm")

	if username != "" {
		if len(resolvedUsernames) > 0 {
			tx = tx.Where("username IN ?", resolvedUsernames)
			rpmTpmQuery = rpmTpmQuery.Where("username IN ?", resolvedUsernames)
		} else {
			tx = tx.Where("username = ?", username)
			rpmTpmQuery = rpmTpmQuery.Where("username = ?", username)
		}
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
		rpmTpmQuery = rpmTpmQuery.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name like ?", modelName)
		rpmTpmQuery = rpmTpmQuery.Where("model_name like ?", modelName)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
		rpmTpmQuery = rpmTpmQuery.Where("channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where(logGroupCol+" = ?", group)
		rpmTpmQuery = rpmTpmQuery.Where(logGroupCol+" = ?", group)
	}

	tx = tx.Where("type = ?", LogTypeConsume)
	rpmTpmQuery = rpmTpmQuery.Where("type = ?", LogTypeConsume)

	// 只统计最近60秒的rpm和tpm
	rpmTpmQuery = rpmTpmQuery.Where("created_at >= ?", time.Now().Add(-60*time.Second).Unix())

	// 执行查询
	tx.Scan(&stat)
	rpmTpmQuery.Scan(&stat)

	return stat
}

func SumUsedToken(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string) (token int) {
	tx := LOG_DB.Table("logs").Select("ifnull(sum(prompt_tokens),0) + ifnull(sum(completion_tokens),0)")
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	tx.Where("type = ?", LogTypeConsume).Scan(&token)
	return token
}

func DeleteOldLog(ctx context.Context, targetTimestamp int64, limit int) (int64, error) {
	var total int64 = 0

	for {
		if nil != ctx.Err() {
			return total, ctx.Err()
		}

		result := LOG_DB.Where("created_at < ?", targetTimestamp).Limit(limit).Delete(&Log{})
		if nil != result.Error {
			return total, result.Error
		}

		total += result.RowsAffected

		if result.RowsAffected < int64(limit) {
			break
		}
	}

	return total, nil
}

type ModelLeaderboardEntry struct {
	ModelName    string `json:"model_name"`
	RequestCount int64  `json:"request_count"`
	TotalTokens  int64  `json:"total_tokens"`
	TotalQuota   int64  `json:"total_quota"`
}

func GetModelUsageLeaderboard(limit int) ([]ModelLeaderboardEntry, error) {
	var entries []ModelLeaderboardEntry
	// Use quota_data table (persistent aggregated data) instead of logs table
	// This ensures data survives log cleanup
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
	Username         string `json:"username"`
	DisplayName      string `json:"display_name"`
	LinuxDOUsername  string `json:"linux_do_username"`
	LinuxDOAvatar    string `json:"linux_do_avatar"`
	LinuxDOLevel     int    `json:"linux_do_level"`
	RequestCount     int64  `json:"request_count"`
	UsedQuota        int64  `json:"used_quota"`
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

func GetUserRankByPeriod(userId int, period string) (int, *UsageLeaderboardEntry, error) {
	var user User
	err := DB.Select("id, username, display_name, linux_do_username, linux_do_avatar, linux_do_level").
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
	query := LOG_DB.Table("logs").
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

	rankQuery := LOG_DB.Table("logs").
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
	LOG_DB.Table("(?) as t", subQuery).Count(&rank)

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

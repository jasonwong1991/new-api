/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

For commercial licensing, please contact support@quantumnous.com
*/

package model

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/setting/quota_limit"
)

// SumUserConsumedQuotaSince returns the total consumed quota (logs.Quota sum)
// for a given user since the provided unix timestamp (inclusive).
// Reads from LOG_DB so accuracy is not dependent on DataExportEnabled.
// Relies on the composite index (user_id, type, created_at) defined on Log.
func SumUserConsumedQuotaSince(userId int, sinceUnix int64) (int64, error) {
	var total int64
	err := LOG_DB.Table("logs").
		Select("COALESCE(SUM(quota), 0)").
		Where("user_id = ? AND type = ? AND created_at >= ?", userId, LogTypeConsume, sinceUnix).
		Scan(&total).Error
	return total, err
}

// ---------- quota-limit cache (per-user, short TTL) ----------

type quotaLimitSample struct {
	daily     int64
	weekly    int64
	dayStart  int64
	weekStart int64
	expiresAt time.Time
}

const (
	quotaLimitCacheTTL        = 20 * time.Second
	quotaLimitCacheGCInterval = 5 * time.Minute
	quotaLimitCacheMaxEntries = 50000
)

var (
	quotaLimitCache       = make(map[int]*quotaLimitSample)
	quotaLimitCacheMu     sync.RWMutex
	quotaLimitCacheGCOnce sync.Once
)

// startQuotaLimitCacheGC kicks off a best-effort sweeper that drops expired
// entries so the map does not leak memory for no-longer-active users.
func startQuotaLimitCacheGC() {
	go func() {
		ticker := time.NewTicker(quotaLimitCacheGCInterval)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			quotaLimitCacheMu.Lock()
			for id, s := range quotaLimitCache {
				if !s.expiresAt.After(now) {
					delete(quotaLimitCache, id)
				}
			}
			quotaLimitCacheMu.Unlock()
		}
	}()
}

// GetUserQuotaUsageWindows returns the user's consumed quota for today and/or
// this natural week. Pass `needDaily=false` or `needWeekly=false` to skip that
// query — the returned value for a skipped window is 0. Results are cached for
// a short TTL to avoid hammering the log DB. Week starts Monday (ISO 8601).
func GetUserQuotaUsageWindows(userId int, needDaily, needWeekly bool) (daily int64, weekly int64, err error) {
	quotaLimitCacheGCOnce.Do(startQuotaLimitCacheGC)

	now := time.Now()
	dayStart := quota_limit.StartOfDay(now)
	weekStart := quota_limit.StartOfWeek(now)

	quotaLimitCacheMu.RLock()
	if hit, ok := quotaLimitCache[userId]; ok && hit.expiresAt.After(now) &&
		hit.dayStart == dayStart && hit.weekStart == weekStart {
		daily, weekly = hit.daily, hit.weekly
		quotaLimitCacheMu.RUnlock()
		return daily, weekly, nil
	}
	quotaLimitCacheMu.RUnlock()

	if needWeekly {
		weekly, err = SumUserConsumedQuotaSince(userId, weekStart)
		if err != nil {
			return 0, 0, err
		}
	}
	if needDaily {
		daily, err = SumUserConsumedQuotaSince(userId, dayStart)
		if err != nil {
			return 0, 0, err
		}
	}

	quotaLimitCacheMu.Lock()
	// Cheap flood guard: if the cache has grown past its ceiling (e.g. during a
	// traffic spike), drop it wholesale rather than paying for eviction logic.
	if len(quotaLimitCache) >= quotaLimitCacheMaxEntries {
		quotaLimitCache = make(map[int]*quotaLimitSample)
	}
	quotaLimitCache[userId] = &quotaLimitSample{
		daily:     daily,
		weekly:    weekly,
		dayStart:  dayStart,
		weekStart: weekStart,
		expiresAt: now.Add(quotaLimitCacheTTL),
	}
	quotaLimitCacheMu.Unlock()

	return daily, weekly, nil
}

// AddUserQuotaUsageDelta increments the cached sample for a user so a freshly
// consumed request is reflected immediately, without waiting for the TTL to
// expire. Safe to call even if there is no cached entry (no-op then).
// The next cache miss will re-read from the DB and reconcile any drift.
func AddUserQuotaUsageDelta(userId int, delta int64) {
	if delta <= 0 {
		return
	}
	now := time.Now()
	dayStart := quota_limit.StartOfDay(now)
	weekStart := quota_limit.StartOfWeek(now)

	quotaLimitCacheMu.Lock()
	defer quotaLimitCacheMu.Unlock()
	s, ok := quotaLimitCache[userId]
	if !ok {
		return
	}
	// If the cached sample spans a prior window (day/week rollover) discard it
	// so the next read starts clean.
	if s.dayStart != dayStart || s.weekStart != weekStart {
		delete(quotaLimitCache, userId)
		return
	}
	s.daily += delta
	s.weekly += delta
}

// InvalidateUserQuotaUsageCache drops the cached sample for a user. Call this
// when an admin action forces a recheck (e.g. whitelist change).
func InvalidateUserQuotaUsageCache(userId int) {
	quotaLimitCacheMu.Lock()
	delete(quotaLimitCache, userId)
	quotaLimitCacheMu.Unlock()
}

// InvalidateAllQuotaUsageCache clears every cached sample.
func InvalidateAllQuotaUsageCache() {
	quotaLimitCacheMu.Lock()
	quotaLimitCache = make(map[int]*quotaLimitSample)
	quotaLimitCacheMu.Unlock()
}

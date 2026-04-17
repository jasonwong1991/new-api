/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

For commercial licensing, please contact support@quantumnous.com
*/

package quota_limit

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// Runtime configuration — wired to common.OptionMap via model/option.go.
var (
	mu sync.RWMutex

	enabled         bool
	dailyLimit      int64
	weeklyLimit     int64
	whitelistUsers  map[string]struct{} // user IDs as strings, and/or usernames
	whitelistGroups map[string]struct{}
)

// SetEnabled toggles the quota limit enforcement.
func SetEnabled(v bool) {
	mu.Lock()
	defer mu.Unlock()
	enabled = v
}

// IsEnabled reports whether the quota limit feature is active.
func IsEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// SetDailyLimit sets the per-user daily quota ceiling (0 = unlimited).
func SetDailyLimit(v int64) {
	if v < 0 {
		v = 0
	}
	mu.Lock()
	defer mu.Unlock()
	dailyLimit = v
}

// GetDailyLimit returns the configured daily ceiling.
func GetDailyLimit() int64 {
	mu.RLock()
	defer mu.RUnlock()
	return dailyLimit
}

// SetWeeklyLimit sets the per-user weekly quota ceiling (0 = unlimited).
func SetWeeklyLimit(v int64) {
	if v < 0 {
		v = 0
	}
	mu.Lock()
	defer mu.Unlock()
	weeklyLimit = v
}

// GetWeeklyLimit returns the configured weekly ceiling.
func GetWeeklyLimit() int64 {
	mu.RLock()
	defer mu.RUnlock()
	return weeklyLimit
}

// SetWhitelistUsers parses a comma-separated list of user IDs or usernames.
func SetWhitelistUsers(raw string) {
	set := parseCsvSet(raw)
	mu.Lock()
	defer mu.Unlock()
	whitelistUsers = set
}

// GetWhitelistUsersRaw returns the set as a comma-separated string (stable order not guaranteed).
func GetWhitelistUsersRaw() string {
	mu.RLock()
	defer mu.RUnlock()
	return joinSet(whitelistUsers)
}

// SetWhitelistGroups parses a comma-separated list of group names.
func SetWhitelistGroups(raw string) {
	set := parseCsvSet(raw)
	mu.Lock()
	defer mu.Unlock()
	whitelistGroups = set
}

// GetWhitelistGroupsRaw returns the raw CSV string.
func GetWhitelistGroupsRaw() string {
	mu.RLock()
	defer mu.RUnlock()
	return joinSet(whitelistGroups)
}

// IsUserWhitelisted checks if a user id (as string) or username matches the whitelist.
func IsUserWhitelisted(userIdStr, username string) bool {
	mu.RLock()
	defer mu.RUnlock()
	if len(whitelistUsers) == 0 {
		return false
	}
	if userIdStr != "" {
		if _, ok := whitelistUsers[userIdStr]; ok {
			return true
		}
	}
	if username != "" {
		if _, ok := whitelistUsers[username]; ok {
			return true
		}
	}
	return false
}

// IsGroupWhitelisted checks if a group name is in the whitelist.
func IsGroupWhitelisted(group string) bool {
	if group == "" {
		return false
	}
	mu.RLock()
	defer mu.RUnlock()
	if len(whitelistGroups) == 0 {
		return false
	}
	_, ok := whitelistGroups[group]
	return ok
}

// StartOfDay returns the unix timestamp for 00:00 of today in the server local time.
func StartOfDay(now time.Time) int64 {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location()).Unix()
}

// StartOfWeek returns the unix timestamp for Monday 00:00 of the current week in server local time.
// Monday is the first day of the natural week (ISO 8601).
func StartOfWeek(now time.Time) int64 {
	// Go's Weekday: Sunday=0..Saturday=6. ISO week starts Monday.
	wd := int(now.Weekday())
	if wd == 0 {
		wd = 7
	}
	offsetDays := wd - 1
	y, m, d := now.Date()
	monday := time.Date(y, m, d, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -offsetDays)
	return monday.Unix()
}

// parseCsvSet splits "a, b ,c" into a trimmed set.
func parseCsvSet(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	if raw == "" {
		return out
	}
	for _, part := range strings.Split(raw, ",") {
		s := strings.TrimSpace(part)
		if s == "" {
			continue
		}
		out[s] = struct{}{}
	}
	return out
}

// joinSet returns a deterministic comma-separated form. Sorting is needed so
// the value fed back into the admin UI stays identical across reads — without
// it, the UI's dirty-check would flap because Go randomises map iteration.
func joinSet(set map[string]struct{}) string {
	if len(set) == 0 {
		return ""
	}
	parts := make([]string, 0, len(set))
	for k := range set {
		parts = append(parts, k)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

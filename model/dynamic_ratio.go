package model

import "time"

// GetTokenUsed24h returns total tokens used since today's 00:00 local time
// from the pre-aggregated quota_data table (much faster than querying logs directly).
// Despite the historical name, the window is today-from-midnight, not a rolling 24h.
func GetTokenUsed24h() (int64, error) {
	var total int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	err := DB.Table("quota_data").
		Select("COALESCE(SUM(token_used), 0)").
		Where("created_at >= ?", startOfDay).
		Scan(&total).Error
	return total, err
}

package model

import "time"

// GetTokenUsed24h returns total tokens used in the last 24 hours from quota_data table.
// Uses the pre-aggregated quota_data table which is much faster than querying logs directly.
func GetTokenUsed24h() (int64, error) {
	var total int64
	startTime := time.Now().Unix() - 86400
	err := DB.Table("quota_data").
		Select("COALESCE(SUM(token_used), 0)").
		Where("created_at >= ?", startTime).
		Scan(&total).Error
	return total, err
}

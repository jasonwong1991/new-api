/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

package service

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"golang.org/x/sync/singleflight"
)

// ModelMonitorBucket 单个时间桶的聚合数据
type ModelMonitorBucket struct {
	Ts           int64   `json:"ts"`
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	Status       string  `json:"status"`
}

// ModelMonitorSummary 窗口内的汇总
type ModelMonitorSummary struct {
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	RPM          float64 `json:"rpm"`
}

// ModelMonitorResult 单模型 + 单粒度的完整结果
type ModelMonitorResult struct {
	ModelName   string               `json:"model_name"`
	Granularity string               `json:"granularity"`
	BucketSize  int64                `json:"bucket_size"`
	BucketCount int                  `json:"bucket_count"`
	StartTs     int64                `json:"start_ts"`
	EndTs       int64                `json:"end_ts"`
	Buckets     []ModelMonitorBucket `json:"buckets"`
	Summary     ModelMonitorSummary  `json:"summary"`
	GeneratedAt int64                `json:"generated_at"`
}

// GranularitySpec 返回桶大小（秒）与桶数量
func GranularitySpec(granularity string) (int64, int) {
	switch granularity {
	case "minute":
		return 60, 60
	case "day":
		return 86400, 30
	case "hour":
		fallthrough
	default:
		return 3600, 24
	}
}

func monitorCacheTTL(granularity string) time.Duration {
	switch granularity {
	case "minute":
		return 10 * time.Second
	case "day":
		return 120 * time.Second
	default:
		return 30 * time.Second
	}
}

type monitorCacheEntry struct {
	expireAt time.Time
	result   *ModelMonitorResult
}

var (
	monitorCache sync.Map // key: "{granularity}:{model}" -> *monitorCacheEntry
	monitorSF    singleflight.Group
)

func monitorCacheKey(granularity, modelName string) string {
	return granularity + ":" + modelName
}

func monitorCacheGet(key string) *ModelMonitorResult {
	v, ok := monitorCache.Load(key)
	if !ok {
		return nil
	}
	entry, _ := v.(*monitorCacheEntry)
	if entry == nil || time.Now().After(entry.expireAt) {
		return nil
	}
	return entry.result
}

func monitorCacheSet(key string, result *ModelMonitorResult, ttl time.Duration) {
	monitorCache.Store(key, &monitorCacheEntry{
		expireAt: time.Now().Add(ttl),
		result:   result,
	})
}

// logDBIsMySQL 识别日志库是否 MySQL（LOG_SQL_DSN 存在则用 LogSqlType）
func logDBIsMySQL() bool {
	if os.Getenv("LOG_SQL_DSN") != "" {
		return common.LogSqlType == common.DatabaseTypeMySQL
	}
	return common.UsingMySQL
}

// bucketSQLExpr 构建跨 DB 时间桶表达式
func bucketSQLExpr(bucketSize int64) string {
	if logDBIsMySQL() {
		return fmt.Sprintf("(FLOOR(created_at / %d) * %d)", bucketSize, bucketSize)
	}
	return fmt.Sprintf("((created_at / %d) * %d)", bucketSize, bucketSize)
}

// classifyStatus 桶状态分类
func classifyStatus(req, err int64) string {
	total := req + err
	if total == 0 {
		return "no_data"
	}
	if err == 0 {
		return "up"
	}
	rate := float64(err) / float64(total)
	if rate > 0.2 {
		return "down"
	}
	return "degraded"
}

type monitorBucketRow struct {
	ModelName    string  `gorm:"column:model_name"`
	Bucket       int64   `gorm:"column:bucket"`
	RequestCount int64   `gorm:"column:request_count"`
	ErrorCount   int64   `gorm:"column:error_count"`
	AvgUseTime   float64 `gorm:"column:avg_use_time"`
}

// assembleResult 将分桶行聚合为 ModelMonitorResult（填充空洞桶）
func assembleResult(modelName, granularity string, rows []monitorBucketRow, now int64) *ModelMonitorResult {
	bucketSize, bucketCount := GranularitySpec(granularity)
	currentBucket := (now / bucketSize) * bucketSize
	startTs := currentBucket - int64(bucketCount-1)*bucketSize
	endTs := currentBucket + bucketSize

	rowMap := make(map[int64]monitorBucketRow, len(rows))
	for _, r := range rows {
		rowMap[r.Bucket] = r
	}

	buckets := make([]ModelMonitorBucket, bucketCount)
	var sumReq, sumErr int64
	var sumLatencyWeighted float64
	for i := 0; i < bucketCount; i++ {
		ts := startTs + int64(i)*bucketSize
		b := ModelMonitorBucket{Ts: ts, Status: "no_data"}
		if r, ok := rowMap[ts]; ok {
			b.RequestCount = r.RequestCount
			b.ErrorCount = r.ErrorCount
			b.AvgLatencyMs = r.AvgUseTime * 1000.0
			b.Status = classifyStatus(r.RequestCount, r.ErrorCount)
			sumReq += r.RequestCount
			sumErr += r.ErrorCount
			sumLatencyWeighted += b.AvgLatencyMs * float64(r.RequestCount+r.ErrorCount)
		}
		buckets[i] = b
	}

	successRate := 1.0
	total := sumReq + sumErr
	if total > 0 {
		successRate = float64(sumReq) / float64(total)
	}
	avgLatency := 0.0
	if total > 0 {
		avgLatency = sumLatencyWeighted / float64(total)
	}
	windowSec := int64(bucketCount) * bucketSize
	rpm := 0.0
	if windowSec > 0 {
		rpm = float64(sumReq) / (float64(windowSec) / 60.0)
	}

	return &ModelMonitorResult{
		ModelName:   modelName,
		Granularity: granularity,
		BucketSize:  bucketSize,
		BucketCount: bucketCount,
		StartTs:     startTs,
		EndTs:       endTs,
		Buckets:     buckets,
		Summary: ModelMonitorSummary{
			RequestCount: sumReq,
			ErrorCount:   sumErr,
			SuccessRate:  successRate,
			AvgLatencyMs: avgLatency,
			RPM:          rpm,
		},
		GeneratedAt: now,
	}
}

// queryBucketsBatch 一次 SQL 查多个模型的分桶数据
func queryBucketsBatch(models []string, granularity string) (map[string][]monitorBucketRow, int64, int64, error) {
	bucketSize, bucketCount := GranularitySpec(granularity)
	now := time.Now().Unix()
	currentBucket := (now / bucketSize) * bucketSize
	startTs := currentBucket - int64(bucketCount-1)*bucketSize
	endTs := currentBucket + bucketSize

	result := make(map[string][]monitorBucketRow, len(models))
	for _, m := range models {
		result[m] = nil
	}
	if len(models) == 0 {
		return result, startTs, endTs, nil
	}

	var rows []monitorBucketRow
	err := model.LOG_DB.Table("logs").
		Select("model_name, "+bucketSQLExpr(bucketSize)+" AS bucket, "+
			"SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) AS request_count, "+
			"SUM(CASE WHEN type = 5 THEN 1 ELSE 0 END) AS error_count, "+
			"AVG(use_time) AS avg_use_time").
		Where("model_name IN ? AND created_at >= ? AND created_at < ? AND type IN (?, ?)",
			models, startTs, endTs, 2, 5).
		Group("model_name, bucket").
		Scan(&rows).Error
	if err != nil {
		return nil, startTs, endTs, err
	}

	for _, r := range rows {
		result[r.ModelName] = append(result[r.ModelName], r)
	}
	return result, startTs, endTs, nil
}

// GetModelMonitorResult 获取单模型结果（缓存 + 单飞）
func GetModelMonitorResult(modelName, granularity string) (*ModelMonitorResult, error) {
	if granularity != "minute" && granularity != "hour" && granularity != "day" {
		granularity = "hour"
	}
	key := monitorCacheKey(granularity, modelName)

	if cached := monitorCacheGet(key); cached != nil {
		return cached, nil
	}

	v, err, _ := monitorSF.Do(key, func() (interface{}, error) {
		// 二次检查，避免并发穿透
		if cached := monitorCacheGet(key); cached != nil {
			return cached, nil
		}
		batch, _, _, qErr := queryBucketsBatch([]string{modelName}, granularity)
		if qErr != nil {
			return nil, qErr
		}
		res := assembleResult(modelName, granularity, batch[modelName], time.Now().Unix())
		monitorCacheSet(key, res, monitorCacheTTL(granularity))
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*ModelMonitorResult), nil
}

// GetModelMonitorResultsBatch 批量获取：命中缓存的直接返回，未命中的合并为一次 SQL
func GetModelMonitorResultsBatch(models []string, granularity string) (map[string]*ModelMonitorResult, error) {
	if granularity != "minute" && granularity != "hour" && granularity != "day" {
		granularity = "hour"
	}
	out := make(map[string]*ModelMonitorResult, len(models))
	missing := make([]string, 0)
	seen := make(map[string]bool, len(models))

	for _, m := range models {
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		if cached := monitorCacheGet(monitorCacheKey(granularity, m)); cached != nil {
			out[m] = cached
		} else {
			missing = append(missing, m)
		}
	}
	if len(missing) == 0 {
		return out, nil
	}

	// 合并 key 做单飞：同一批 missing 集合在短时间内只打一次 DB
	// 注意：不同批次可能 miss 集不同，用稳定排序 key
	sortedKey := granularity + "|batch|" + joinSorted(missing)
	v, err, _ := monitorSF.Do(sortedKey, func() (interface{}, error) {
		// 再次过滤已经被别人填进缓存的（竞争窗口）
		stillMissing := make([]string, 0, len(missing))
		for _, m := range missing {
			if cached := monitorCacheGet(monitorCacheKey(granularity, m)); cached != nil {
				out[m] = cached
			} else {
				stillMissing = append(stillMissing, m)
			}
		}
		if len(stillMissing) == 0 {
			return out, nil
		}
		batch, _, _, qErr := queryBucketsBatch(stillMissing, granularity)
		if qErr != nil {
			return nil, qErr
		}
		now := time.Now().Unix()
		ttl := monitorCacheTTL(granularity)
		for _, m := range stillMissing {
			res := assembleResult(m, granularity, batch[m], now)
			monitorCacheSet(monitorCacheKey(granularity, m), res, ttl)
			out[m] = res
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(map[string]*ModelMonitorResult), nil
}

// --- Metrics cache (for /metrics endpoint) ---

// MonitorAggRow holds per-model aggregated data from the logs table.
type MonitorAggRow struct {
	ModelName        string  `gorm:"column:model_name"`
	RequestCount     int64   `gorm:"column:request_count"`
	ErrorCount       int64   `gorm:"column:error_count"`
	AvgUseTime       float64 `gorm:"column:avg_use_time"`
	PromptTokens     int64   `gorm:"column:prompt_tokens"`
	CompletionTokens int64   `gorm:"column:completion_tokens"`
	QuotaSum         int64   `gorm:"column:quota_sum"`
}

type metricsCacheEntry struct {
	expireAt time.Time
	rows     []MonitorAggRow
}

const metricsCacheTTL = 30 * time.Second

var (
	metricsCache sync.Map // key: window string -> *metricsCacheEntry
	metricsSF    singleflight.Group
)

// GetCachedMetrics returns aggregated MonitorAggRow for the given window (1h/24h/7d),
// using a 30-second cache backed by singleflight to prevent stampedes.
func GetCachedMetrics(window string) ([]MonitorAggRow, error) {
	if v, ok := metricsCache.Load(window); ok {
		if e, _ := v.(*metricsCacheEntry); e != nil && time.Now().Before(e.expireAt) {
			return e.rows, nil
		}
	}

	result, err, _ := metricsSF.Do(window, func() (interface{}, error) {
		// Double-check inside singleflight to avoid races.
		if v, ok := metricsCache.Load(window); ok {
			if e, _ := v.(*metricsCacheEntry); e != nil && time.Now().Before(e.expireAt) {
				return e.rows, nil
			}
		}

		var seconds int64
		switch window {
		case "1h":
			seconds = 3600
		case "7d":
			seconds = 7 * 86400
		default:
			seconds = 86400
		}
		cutoff := time.Now().Unix() - seconds

		var rows []MonitorAggRow
		err := model.LOG_DB.Table("logs").
			Select("model_name, " +
				"SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) AS request_count, " +
				"SUM(CASE WHEN type = 5 THEN 1 ELSE 0 END) AS error_count, " +
				"AVG(use_time) AS avg_use_time, " +
				"SUM(prompt_tokens) AS prompt_tokens, " +
				"SUM(completion_tokens) AS completion_tokens, " +
				"SUM(quota) AS quota_sum").
			Where("created_at >= ? AND type IN (?, ?) AND model_name <> ''", cutoff, 2, 5).
			Group("model_name").
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}

		metricsCache.Store(window, &metricsCacheEntry{
			expireAt: time.Now().Add(metricsCacheTTL),
			rows:     rows,
		})
		return rows, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]MonitorAggRow), nil
}

// --- Models list cache (for /models endpoint) ---

type modelsListCacheEntry struct {
	expireAt time.Time
	models   []string
}

const modelsListCacheTTL = 5 * time.Minute

var (
	modelsListCache atomic.Value // stores *modelsListCacheEntry
	modelsListMu    sync.Mutex
)

// GetCachedModelsList returns the list of model names that have log entries in
// the last 7 days, using a 5-minute cache. Only used when there is no explicit
// whitelist configured.
func GetCachedModelsList() ([]string, error) {
	if e, _ := modelsListCache.Load().(*modelsListCacheEntry); e != nil && time.Now().Before(e.expireAt) {
		return e.models, nil
	}

	modelsListMu.Lock()
	defer modelsListMu.Unlock()

	// Double-check after acquiring the lock.
	if e, _ := modelsListCache.Load().(*modelsListCacheEntry); e != nil && time.Now().Before(e.expireAt) {
		return e.models, nil
	}

	cutoff := time.Now().Unix() - 7*86400
	var rows []struct {
		ModelName string `gorm:"column:model_name"`
		Total     int64  `gorm:"column:total"`
	}
	err := model.LOG_DB.Table("logs").
		Select("model_name, COUNT(*) AS total").
		Where("created_at >= ? AND type IN (?, ?) AND model_name <> ''", cutoff, 2, 5).
		Group("model_name").
		Order("total DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	models := make([]string, 0, len(rows))
	for _, r := range rows {
		models = append(models, r.ModelName)
	}

	modelsListCache.Store(&modelsListCacheEntry{
		expireAt: time.Now().Add(modelsListCacheTTL),
		models:   models,
	})
	return models, nil
}

// InvalidateModelMonitorCache 手动失效（目前未使用，预留给配置变更触发）
func InvalidateModelMonitorCache() {
	monitorCache.Range(func(k, _ interface{}) bool {
		monitorCache.Delete(k)
		return true
	})
	metricsCache.Range(func(k, _ interface{}) bool {
		metricsCache.Delete(k)
		return true
	})
	modelsListCache.Store((*modelsListCacheEntry)(nil))
}

// joinSorted 稳定排序后 join，保证相同集合生成相同 key
func joinSorted(list []string) string {
	cp := make([]string, len(list))
	copy(cp, list)
	// 简单插入排序：列表通常不超过几十项
	for i := 1; i < len(cp); i++ {
		for j := i; j > 0 && cp[j-1] > cp[j]; j-- {
			cp[j-1], cp[j] = cp[j], cp[j-1]
		}
	}
	buf := ""
	for i, s := range cp {
		if i > 0 {
			buf += ","
		}
		buf += s
	}
	return buf
}

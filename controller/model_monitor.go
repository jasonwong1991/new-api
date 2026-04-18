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

package controller

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/console_setting"

	"github.com/gin-gonic/gin"
)

type modelMonitorMetric struct {
	ModelName        string  `json:"model_name"`
	RequestCount     int64   `json:"request_count"`
	ErrorCount       int64   `json:"error_count"`
	SuccessRate      float64 `json:"success_rate"`
	AvgLatencyMs     float64 `json:"avg_latency_ms"`
	RPM              float64 `json:"rpm"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	Quota            int64   `json:"quota"`
}

type modelMonitorAggRow struct {
	ModelName        string  `gorm:"column:model_name"`
	RequestCount     int64   `gorm:"column:request_count"`
	ErrorCount       int64   `gorm:"column:error_count"`
	AvgUseTime       float64 `gorm:"column:avg_use_time"`
	PromptTokens     int64   `gorm:"column:prompt_tokens"`
	CompletionTokens int64   `gorm:"column:completion_tokens"`
	QuotaSum         int64   `gorm:"column:quota_sum"`
}

func windowSeconds(window string) int64 {
	switch window {
	case "1h":
		return 3600
	case "7d":
		return 7 * 86400
	case "24h":
		fallthrough
	default:
		return 86400
	}
}

// GetModelMonitorMetrics 聚合日志表返回模型维度监控数据
func GetModelMonitorMetrics(c *gin.Context) {
	cfg := console_setting.GetModelMonitorConfig()

	window := c.Query("window")
	if !(window == "1h" || window == "24h" || window == "7d") {
		window = cfg.DefaultWindow
	}
	if window == "" {
		window = "24h"
	}

	seconds := windowSeconds(window)
	cutoff := time.Now().Unix() - seconds

	var rows []modelMonitorAggRow
	err := model.LOG_DB.Table("logs").
		Select("model_name, "+
			"SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) AS request_count, "+
			"SUM(CASE WHEN type = 5 THEN 1 ELSE 0 END) AS error_count, "+
			"AVG(use_time) AS avg_use_time, "+
			"SUM(prompt_tokens) AS prompt_tokens, "+
			"SUM(completion_tokens) AS completion_tokens, "+
			"SUM(quota) AS quota_sum").
		Where("created_at >= ? AND type IN (?, ?) AND model_name <> ''", cutoff, 2, 5).
		Group("model_name").
		Scan(&rows).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	whitelist := map[string]bool{}
	for _, m := range cfg.Models {
		if m != "" {
			whitelist[m] = true
		}
	}

	minutes := float64(seconds) / 60.0
	metrics := make([]modelMonitorMetric, 0, len(rows))
	for _, r := range rows {
		if len(whitelist) > 0 && !whitelist[r.ModelName] {
			continue
		}
		total := r.RequestCount + r.ErrorCount
		successRate := 1.0
		if total > 0 {
			successRate = float64(r.RequestCount) / float64(total)
		}
		rpm := 0.0
		if minutes > 0 {
			rpm = float64(r.RequestCount) / minutes
		}
		metrics = append(metrics, modelMonitorMetric{
			ModelName:        r.ModelName,
			RequestCount:     r.RequestCount,
			ErrorCount:       r.ErrorCount,
			SuccessRate:      successRate,
			AvgLatencyMs:     r.AvgUseTime * 1000.0,
			RPM:              rpm,
			PromptTokens:     r.PromptTokens,
			CompletionTokens: r.CompletionTokens,
			TotalTokens:      r.PromptTokens + r.CompletionTokens,
			Quota:            r.QuotaSum,
		})
	}

	sortMetrics(metrics, cfg.SortBy, cfg.SortOrder)

	if cfg.Limit > 0 && len(metrics) > cfg.Limit {
		metrics = metrics[:cfg.Limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"window":       window,
			"generated_at": time.Now().Unix(),
			"models":       metrics,
			"sort_by":      cfg.SortBy,
			"sort_order":   cfg.SortOrder,
			"refresh_sec":  cfg.RefreshSec,
		},
	})
}

// GetModelMonitorModels 轻量 endpoint：仅返回受监控的模型列表（供前端快速构建行）
func GetModelMonitorModels(c *gin.Context) {
	cfg := console_setting.GetModelMonitorConfig()

	models := make([]string, 0)
	if len(cfg.Models) > 0 {
		// 显式白名单
		for _, m := range cfg.Models {
			if m != "" {
				models = append(models, m)
			}
		}
	} else {
		// 无白名单：使用最近 7 天内有日志的模型作为默认清单
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
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			return
		}
		for _, r := range rows {
			models = append(models, r.ModelName)
		}
	}

	if cfg.Limit > 0 && len(models) > cfg.Limit {
		models = models[:cfg.Limit]
	}

	// 时间窗口 -> 粒度：管理员选择后所有用户强制按此粒度查看
	windowToGranularity := map[string]string{
		"1h":  "minute",
		"24h": "hour",
		"7d":  "day",
	}
	granularity := windowToGranularity[cfg.DefaultWindow]
	if granularity == "" {
		granularity = "hour"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"models":              models,
			"refresh_sec":         cfg.RefreshSec,
			"default_granularity": granularity,
			"default_window":      cfg.DefaultWindow,
		},
	})
}

// GetModelMonitorBars 按时间粒度返回一个或多个模型的状态桶
// 支持：?model=X  或  ?models=X,Y,Z
func GetModelMonitorBars(c *gin.Context) {
	granularity := c.Query("granularity")
	if granularity != "minute" && granularity != "hour" && granularity != "day" {
		granularity = "hour"
	}

	// 收集请求的模型列表
	reqModels := make([]string, 0)
	if raw := c.Query("models"); raw != "" {
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				reqModels = append(reqModels, p)
			}
		}
	}
	if single := strings.TrimSpace(c.Query("model")); single != "" {
		reqModels = append(reqModels, single)
	}
	if len(reqModels) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "缺少 model/models 参数"})
		return
	}

	// 上限防御：单次最多 100 个模型，超过拒绝
	if len(reqModels) > 100 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "一次查询模型数不能超过 100"})
		return
	}

	results, err := service.GetModelMonitorResultsBatch(reqModels, granularity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 单模型兼容格式（保留旧 ?model= 的 data 形状）
	if len(reqModels) == 1 && c.Query("models") == "" {
		if r, ok := results[reqModels[0]]; ok {
			c.JSON(http.StatusOK, gin.H{"success": true, "data": r})
			return
		}
	}

	bucketSize, bucketCount := service.GranularitySpec(granularity)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"granularity":  granularity,
			"bucket_size":  bucketSize,
			"bucket_count": bucketCount,
			"results":      results,
			"generated_at": time.Now().Unix(),
		},
	})
}

func sortMetrics(list []modelMonitorMetric, sortBy, sortOrder string) {
	asc := sortOrder == "asc"
	less := func(i, j int) bool {
		a, b := list[i], list[j]
		var av, bv float64
		switch sortBy {
		case "success_rate":
			av, bv = a.SuccessRate, b.SuccessRate
		case "avg_latency":
			av, bv = a.AvgLatencyMs, b.AvgLatencyMs
		case "quota":
			av, bv = float64(a.Quota), float64(b.Quota)
		case "total_tokens":
			av, bv = float64(a.TotalTokens), float64(b.TotalTokens)
		default:
			av, bv = float64(a.RequestCount), float64(b.RequestCount)
		}
		if av == bv {
			return a.ModelName < b.ModelName
		}
		if asc {
			return av < bv
		}
		return av > bv
	}
	sort.SliceStable(list, less)
}

// GetModelMonitorConfigAPI 返回当前模型监控配置（已登录用户可见）
func GetModelMonitorConfigAPI(c *gin.Context) {
	cfg := console_setting.GetModelMonitorConfig()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

// UpdateModelMonitorConfigAPI 更新模型监控配置（仅 root）
func UpdateModelMonitorConfigAPI(c *gin.Context) {
	var cfg console_setting.ModelMonitorConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if cfg.Models == nil {
		cfg.Models = []string{}
	}
	if err := console_setting.ValidateModelMonitorConfig(cfg); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	jsonBytes, err := common.Marshal(cfg)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.UpdateOption("console_setting.model_monitor", string(jsonBytes)); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	// 让后端缓存立即失效，避免用户继续看到旧配置下的数据
	service.InvalidateModelMonitorCache()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

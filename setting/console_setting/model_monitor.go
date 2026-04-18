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

package console_setting

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
)

// ModelMonitorConfig 模型监控页面配置
type ModelMonitorConfig struct {
	Enabled       bool     `json:"enabled"`
	Models        []string `json:"models"`
	SortBy        string   `json:"sort_by"`
	SortOrder     string   `json:"sort_order"`
	DefaultWindow string   `json:"default_window"`
	Limit         int      `json:"limit"`
	RefreshSec    int      `json:"refresh_sec"`
}

var defaultModelMonitorConfig = ModelMonitorConfig{
	Enabled:       true,
	Models:        []string{},
	SortBy:        "request_count",
	SortOrder:     "desc",
	DefaultWindow: "24h",
	Limit:         20,
	RefreshSec:    30,
}

var validSortBy = map[string]bool{
	"request_count": true,
	"success_rate":  true,
	"avg_latency":   true,
	"quota":         true,
	"total_tokens":  true,
}

var validSortOrder = map[string]bool{
	"asc":  true,
	"desc": true,
}

var validWindow = map[string]bool{
	"1h":  true,
	"24h": true,
	"7d":  true,
}

// GetModelMonitorConfig 读取当前模型监控配置（解析自 console_setting.ModelMonitor JSON）
func GetModelMonitorConfig() ModelMonitorConfig {
	cs := GetConsoleSetting()
	cfg := defaultModelMonitorConfig
	if cs == nil || cs.ModelMonitor == "" {
		return cfg
	}
	if err := common.UnmarshalJsonStr(cs.ModelMonitor, &cfg); err != nil {
		return defaultModelMonitorConfig
	}
	// 回填缺省值
	if cfg.SortBy == "" {
		cfg.SortBy = defaultModelMonitorConfig.SortBy
	}
	if cfg.SortOrder == "" {
		cfg.SortOrder = defaultModelMonitorConfig.SortOrder
	}
	if cfg.DefaultWindow == "" {
		cfg.DefaultWindow = defaultModelMonitorConfig.DefaultWindow
	}
	if cfg.RefreshSec <= 0 {
		cfg.RefreshSec = defaultModelMonitorConfig.RefreshSec
	}
	if cfg.Models == nil {
		cfg.Models = []string{}
	}
	return cfg
}

// ValidateModelMonitorConfig 校验配置
func ValidateModelMonitorConfig(cfg ModelMonitorConfig) error {
	if !validSortBy[cfg.SortBy] {
		return fmt.Errorf("sort_by 取值不合法")
	}
	if !validSortOrder[cfg.SortOrder] {
		return fmt.Errorf("sort_order 取值不合法")
	}
	if !validWindow[cfg.DefaultWindow] {
		return fmt.Errorf("default_window 取值不合法")
	}
	if cfg.Limit < 0 || cfg.Limit > 200 {
		return fmt.Errorf("limit 范围应在 0-200 之间")
	}
	if cfg.RefreshSec < 5 || cfg.RefreshSec > 3600 {
		return fmt.Errorf("refresh_sec 范围应在 5-3600 之间")
	}
	if len(cfg.Models) > 500 {
		return fmt.Errorf("models 数量不能超过 500")
	}
	return nil
}

// ValidateModelMonitorJSON 验证 JSON 字符串是否满足配置规则
func ValidateModelMonitorJSON(jsonStr string) error {
	if jsonStr == "" {
		return nil
	}
	var cfg ModelMonitorConfig
	if err := common.UnmarshalJsonStr(jsonStr, &cfg); err != nil {
		return fmt.Errorf("model_monitor 配置 JSON 解析失败: %s", err.Error())
	}
	return ValidateModelMonitorConfig(cfg)
}

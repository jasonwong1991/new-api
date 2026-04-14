package setting

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// Configuration constants
const (
	DynamicRatioTokenThreshold int64   = 10_000_000_000 // 10B tokens
	DynamicRatioTokenRange     int64   = 20_000_000_000 // 20B range (10B-30B)
	DynamicRatioRPMThreshold   float64 = 1000.0
	DynamicRatioTokenWeight    float64 = 0.7
	DynamicRatioRPMWeight      float64 = 0.3
	DynamicRatioMax            float64 = 5.0
	DynamicRatioMin            float64 = 1.0
)

var DynamicRatioEnabled = false

type dynamicRatioCache struct {
	mu          sync.RWMutex
	tokens24h   int64
	currentRPM  int64
	ratio       float64
	lastUpdated time.Time
}

var drCache = &dynamicRatioCache{ratio: 1.0}

// Atomic request counter for RPM calculation
var requestCounter int64
var lastMinuteRPM int64

func IncrementRequestCount() {
	atomic.AddInt64(&requestCounter, 1)
}

func GetDynamicRatio() float64 {
	if !DynamicRatioEnabled {
		return 1.0
	}
	drCache.mu.RLock()
	defer drCache.mu.RUnlock()
	return drCache.ratio
}

type DynamicRatioInfo struct {
	Tokens24h    int64   `json:"tokens_24h"`
	CurrentRPM   int64   `json:"current_rpm"`
	DynamicRatio float64 `json:"dynamic_ratio"`
	Enabled      bool    `json:"enabled"`
	UpdatedAt    int64   `json:"updated_at"`
}

func GetDynamicRatioInfo() DynamicRatioInfo {
	drCache.mu.RLock()
	defer drCache.mu.RUnlock()
	return DynamicRatioInfo{
		Tokens24h:    drCache.tokens24h,
		CurrentRPM:   drCache.currentRPM,
		DynamicRatio: drCache.ratio,
		Enabled:      DynamicRatioEnabled,
		UpdatedAt:    drCache.lastUpdated.Unix(),
	}
}

func CalculateDynamicRatio(tokens24h int64, currentRPM int64) float64 {
	if tokens24h <= DynamicRatioTokenThreshold {
		return DynamicRatioMin
	}

	tokenExcess := float64(tokens24h - DynamicRatioTokenThreshold)
	tokenFactor := math.Min(tokenExcess/float64(DynamicRatioTokenRange), 1.0)

	rpmFactor := math.Min(float64(currentRPM)/DynamicRatioRPMThreshold, 1.0)

	combined := tokenFactor*DynamicRatioTokenWeight + rpmFactor*DynamicRatioRPMWeight
	ratio := DynamicRatioMin + combined*(DynamicRatioMax-DynamicRatioMin)
	ratio = math.Min(ratio, DynamicRatioMax)

	// Round to 1 decimal place
	ratio = math.Round(ratio*10) / 10

	return ratio
}

// refreshTokenCountFunc is set by model package to avoid import cycles
var refreshTokenCountFunc func() (int64, error)

func SetRefreshTokenCountFunc(f func() (int64, error)) {
	refreshTokenCountFunc = f
}

func StartDynamicRatioScheduler() {
	common.SysLog("Starting dynamic ratio scheduler...")

	// RPM sampling goroutine - every 60 seconds
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			count := atomic.SwapInt64(&requestCounter, 0)
			atomic.StoreInt64(&lastMinuteRPM, count)
			updateDynamicRatioCache()
		}
	}()

	// Token count refresh - every 5 minutes
	go func() {
		// Initial load after short delay
		time.Sleep(10 * time.Second)
		refreshTokens()
		updateDynamicRatioCache()

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			refreshTokens()
			updateDynamicRatioCache()
		}
	}()
}

func refreshTokens() {
	if refreshTokenCountFunc == nil {
		return
	}
	tokens, err := refreshTokenCountFunc()
	if err != nil {
		common.SysError("failed to query 24h tokens for dynamic ratio: " + err.Error())
		return
	}
	drCache.mu.Lock()
	drCache.tokens24h = tokens
	drCache.mu.Unlock()
}

func updateDynamicRatioCache() {
	rpm := atomic.LoadInt64(&lastMinuteRPM)

	drCache.mu.Lock()
	defer drCache.mu.Unlock()

	drCache.currentRPM = rpm
	drCache.ratio = CalculateDynamicRatio(drCache.tokens24h, rpm)
	drCache.lastUpdated = time.Now()

	if common.DebugEnabled {
		common.SysLog(fmt.Sprintf("dynamic ratio updated: tokens_24h=%d, rpm=%d, ratio=%.1f",
			drCache.tokens24h, rpm, drCache.ratio))
	}
}

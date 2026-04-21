package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
)

func setupModelRateLimitTest(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	originalRedisEnabled := common.RedisEnabled
	originalEnabled := setting.ModelRequestRateLimitEnabled
	originalDuration := setting.ModelRequestRateLimitDurationMinutes
	originalUserCount := setting.ModelRequestRateLimitCount
	originalIPCount := setting.ModelRequestIPRateLimitCount
	originalSuccessCount := setting.ModelRequestRateLimitSuccessCount
	originalWhitelist := make(map[string]struct{}, len(setting.RateLimitExemptWhitelist))
	for key := range setting.RateLimitExemptWhitelist {
		originalWhitelist[key] = struct{}{}
	}

	common.RedisEnabled = false
	inMemoryRateLimiter = common.InMemoryRateLimiter{}
	setting.ModelRequestRateLimitEnabled = true
	setting.ModelRequestRateLimitDurationMinutes = 1
	setting.ModelRequestRateLimitCount = 0
	setting.ModelRequestIPRateLimitCount = 0
	setting.ModelRequestRateLimitSuccessCount = 1000
	setting.RateLimitExemptWhitelist = map[string]struct{}{}

	t.Cleanup(func() {
		common.RedisEnabled = originalRedisEnabled
		inMemoryRateLimiter = common.InMemoryRateLimiter{}
		setting.ModelRequestRateLimitEnabled = originalEnabled
		setting.ModelRequestRateLimitDurationMinutes = originalDuration
		setting.ModelRequestRateLimitCount = originalUserCount
		setting.ModelRequestIPRateLimitCount = originalIPCount
		setting.ModelRequestRateLimitSuccessCount = originalSuccessCount
		setting.RateLimitExemptWhitelist = originalWhitelist
	})
}

func newModelRateLimitTestRouter() *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		userID, _ := strconv.Atoi(c.GetHeader("X-Test-User-Id"))
		c.Set("id", userID)
		c.Set("username", c.GetHeader("X-Test-Username"))
		c.Next()
	})
	router.Use(ModelRequestRateLimit())
	router.POST("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func performModelRateLimitRequest(router *gin.Engine, remoteAddr string, userID int, username string) int {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = remoteAddr
	req.Header.Set("X-Test-User-Id", strconv.Itoa(userID))
	req.Header.Set("X-Test-Username", username)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder.Code
}

func TestModelRequestRateLimitIPLimitComplementsUserLimit(t *testing.T) {
	setupModelRateLimitTest(t)

	setting.ModelRequestIPRateLimitCount = 1
	if err := setting.UpdateRateLimitExemptWhitelistByJSONString("[2]"); err != nil {
		t.Fatalf("failed to configure whitelist: %v", err)
	}

	router := newModelRateLimitTestRouter()
	remoteAddr := "198.51.100.10:1234"

	if code := performModelRateLimitRequest(router, remoteAddr, 1, "normal-user"); code != http.StatusOK {
		t.Fatalf("expected first request to pass, got status %d", code)
	}
	if code := performModelRateLimitRequest(router, remoteAddr, 2, "whitelist-user"); code != http.StatusOK {
		t.Fatalf("expected whitelisted user to bypass IP limit, got status %d", code)
	}
	if code := performModelRateLimitRequest(router, remoteAddr, 3, "another-user"); code != http.StatusTooManyRequests {
		t.Fatalf("expected third request to be blocked by IP limit, got status %d", code)
	}
}

func TestModelRequestRateLimitWhitelistBypassesUserAndIPCounters(t *testing.T) {
	setupModelRateLimitTest(t)

	setting.ModelRequestRateLimitCount = 1
	setting.ModelRequestIPRateLimitCount = 1
	if err := setting.UpdateRateLimitExemptWhitelistByJSONString("[9]"); err != nil {
		t.Fatalf("failed to configure whitelist: %v", err)
	}

	router := newModelRateLimitTestRouter()
	remoteAddr := "203.0.113.8:1234"

	if code := performModelRateLimitRequest(router, remoteAddr, 9, "whitelist-user"); code != http.StatusOK {
		t.Fatalf("expected first whitelisted request to pass, got status %d", code)
	}
	if code := performModelRateLimitRequest(router, remoteAddr, 9, "whitelist-user"); code != http.StatusOK {
		t.Fatalf("expected second whitelisted request to pass, got status %d", code)
	}
	if code := performModelRateLimitRequest(router, remoteAddr, 1, "normal-user"); code != http.StatusOK {
		t.Fatalf("expected non-whitelisted request to remain available after whitelist bypass, got status %d", code)
	}
	if code := performModelRateLimitRequest(router, remoteAddr, 1, "normal-user"); code != http.StatusTooManyRequests {
		t.Fatalf("expected second non-whitelisted request to hit a rate limit, got status %d", code)
	}
}

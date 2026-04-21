package setting

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
)

var ModelRequestRateLimitEnabled = false
var ModelRequestRateLimitDurationMinutes = 1
var ModelRequestRateLimitCount = 0
var ModelRequestIPRateLimitCount = 0
var ModelRequestRateLimitSuccessCount = 1000
var ModelRequestRateLimitGroup = map[string][2]int{}
var ModelRequestRateLimitMutex sync.RWMutex
var RateLimitExemptWhitelist = map[string]struct{}{}
var RateLimitExemptWhitelistMutex sync.RWMutex

func ModelRequestRateLimitGroup2JSONString() string {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	jsonBytes, err := common.Marshal(ModelRequestRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRequestRateLimitGroupByJSONString(jsonStr string) error {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	ModelRequestRateLimitGroup = make(map[string][2]int)
	return common.UnmarshalJsonStr(jsonStr, &ModelRequestRateLimitGroup)
}

func GetGroupRateLimit(group string) (totalCount, successCount int, found bool) {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	if ModelRequestRateLimitGroup == nil {
		return 0, 0, false
	}

	limits, found := ModelRequestRateLimitGroup[group]
	if !found {
		return 0, 0, false
	}
	return limits[0], limits[1], true
}

func CheckModelRequestRateLimitGroup(jsonStr string) error {
	checkModelRequestRateLimitGroup := make(map[string][2]int)
	err := common.UnmarshalJsonStr(jsonStr, &checkModelRequestRateLimitGroup)
	if err != nil {
		return err
	}
	for group, limits := range checkModelRequestRateLimitGroup {
		if limits[0] < 0 || limits[1] < 1 {
			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
		}
		if limits[0] > math.MaxInt32 || limits[1] > math.MaxInt32 {
			return fmt.Errorf("group %s [%d, %d] has max rate limits value 2147483647", group, limits[0], limits[1])
		}
	}

	return nil
}

func UpdateRateLimitExemptWhitelistByJSONString(jsonStr string) error {
	whitelist, err := parseRateLimitExemptWhitelist(jsonStr)
	if err != nil {
		return err
	}

	RateLimitExemptWhitelistMutex.Lock()
	defer RateLimitExemptWhitelistMutex.Unlock()
	RateLimitExemptWhitelist = whitelist
	return nil
}

func CheckRateLimitExemptWhitelist(jsonStr string) error {
	_, err := parseRateLimitExemptWhitelist(jsonStr)
	return err
}

func IsRateLimitExemptUser(userIdStr, username string) bool {
	RateLimitExemptWhitelistMutex.RLock()
	defer RateLimitExemptWhitelistMutex.RUnlock()

	if len(RateLimitExemptWhitelist) == 0 {
		return false
	}
	if userIdStr != "" {
		if _, ok := RateLimitExemptWhitelist[userIdStr]; ok {
			return true
		}
	}
	if username != "" {
		if _, ok := RateLimitExemptWhitelist[username]; ok {
			return true
		}
	}
	return false
}

func parseRateLimitExemptWhitelist(jsonStr string) (map[string]struct{}, error) {
	whitelist := make(map[string]struct{})
	if strings.TrimSpace(jsonStr) == "" {
		return whitelist, nil
	}

	var rawList []any
	if err := common.UnmarshalJsonStr(jsonStr, &rawList); err != nil {
		return nil, err
	}

	for _, item := range rawList {
		switch value := item.(type) {
		case float64:
			if value < 0 || value != math.Trunc(value) {
				return nil, fmt.Errorf("rate limit exempt whitelist only supports non-negative integer user IDs or usernames")
			}
			whitelist[strconv.FormatInt(int64(value), 10)] = struct{}{}
		case string:
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				return nil, fmt.Errorf("rate limit exempt whitelist does not support empty usernames")
			}
			whitelist[trimmed] = struct{}{}
		default:
			return nil, fmt.Errorf("rate limit exempt whitelist only supports non-negative integer user IDs or usernames")
		}
	}

	return whitelist, nil
}

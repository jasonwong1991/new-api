package setting

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"github.com/QuantumNous/new-api/common"
)

var ModelRequestRateLimitEnabled = false
var ModelRequestRateLimitDurationMinutes = 1
var ModelRequestRateLimitCount = 0
var ModelRequestRateLimitSuccessCount = 1000
var ModelRequestRateLimitGroup = map[string][2]int{}
var ModelRequestRateLimitMutex sync.RWMutex

// RateLimitExemptWhitelist stores user IDs that are exempt from all rate limits (highest priority)
var RateLimitExemptWhitelist = map[int]struct{}{}
var RateLimitExemptWhitelistMutex sync.RWMutex

func ModelRequestRateLimitGroup2JSONString() string {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	jsonBytes, err := json.Marshal(ModelRequestRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRequestRateLimitGroupByJSONString(jsonStr string) error {
	ModelRequestRateLimitMutex.Lock()
	defer ModelRequestRateLimitMutex.Unlock()

	ModelRequestRateLimitGroup = make(map[string][2]int)
	return json.Unmarshal([]byte(jsonStr), &ModelRequestRateLimitGroup)
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
	err := json.Unmarshal([]byte(jsonStr), &checkModelRequestRateLimitGroup)
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

// IsUserExemptFromRateLimit checks if a user is in the rate limit exemption whitelist (highest priority)
func IsUserExemptFromRateLimit(userId int) bool {
	RateLimitExemptWhitelistMutex.RLock()
	defer RateLimitExemptWhitelistMutex.RUnlock()
	_, exists := RateLimitExemptWhitelist[userId]
	return exists
}

// RateLimitExemptWhitelist2JSONString converts the whitelist map to JSON array string
func RateLimitExemptWhitelist2JSONString() string {
	RateLimitExemptWhitelistMutex.RLock()
	defer RateLimitExemptWhitelistMutex.RUnlock()

	userIds := make([]int, 0, len(RateLimitExemptWhitelist))
	for userId := range RateLimitExemptWhitelist {
		userIds = append(userIds, userId)
	}
	jsonBytes, err := json.Marshal(userIds)
	if err != nil {
		common.SysLog("error marshalling rate limit exempt whitelist: " + err.Error())
		return "[]"
	}
	return string(jsonBytes)
}

// UpdateRateLimitExemptWhitelistByJSONString updates the whitelist from JSON array string
func UpdateRateLimitExemptWhitelistByJSONString(jsonStr string) error {
	RateLimitExemptWhitelistMutex.Lock()
	defer RateLimitExemptWhitelistMutex.Unlock()

	if jsonStr == "" {
		RateLimitExemptWhitelist = make(map[int]struct{})
		return nil
	}

	var userIds []int
	err := json.Unmarshal([]byte(jsonStr), &userIds)
	if err != nil {
		return err
	}

	newWhitelist := make(map[int]struct{})
	for _, userId := range userIds {
		if userId > 0 {
			newWhitelist[userId] = struct{}{}
		}
	}
	RateLimitExemptWhitelist = newWhitelist
	return nil
}

// CheckRateLimitExemptWhitelist validates the whitelist JSON string
func CheckRateLimitExemptWhitelist(jsonStr string) error {
	if jsonStr == "" {
		return nil
	}
	var userIds []int
	err := json.Unmarshal([]byte(jsonStr), &userIds)
	if err != nil {
		return fmt.Errorf("invalid JSON array format: %v", err)
	}
	for _, userId := range userIds {
		if userId <= 0 {
			return fmt.Errorf("user ID must be positive integer, got: %d", userId)
		}
	}
	return nil
}

package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasActiveUserSubscription_CacheInvalidation(t *testing.T) {
	truncateTables(t)

	sub := &UserSubscription{
		UserId:      101,
		PlanId:      1,
		AmountTotal: 100,
		Status:      "active",
		StartTime:   time.Now().Add(-time.Hour).Unix(),
		EndTime:     time.Now().Add(time.Hour).Unix(),
	}
	require.NoError(t, DB.Create(sub).Error)

	hasSub, err := HasActiveUserSubscription(sub.UserId)
	require.NoError(t, err)
	assert.True(t, hasSub)

	require.NoError(t, DB.Delete(&UserSubscription{}, sub.Id).Error)

	hasSub, err = HasActiveUserSubscription(sub.UserId)
	require.NoError(t, err)
	assert.True(t, hasSub, "should still hit the short-lived cache before invalidation")

	InvalidateSubStatusCache(sub.UserId)

	hasSub, err = HasActiveUserSubscription(sub.UserId)
	require.NoError(t, err)
	assert.False(t, hasSub)
}

func TestHasActiveUserSubscription_IgnoresExpiredSubscriptions(t *testing.T) {
	truncateTables(t)

	sub := &UserSubscription{
		UserId:      102,
		PlanId:      1,
		AmountTotal: 100,
		Status:      "active",
		StartTime:   time.Now().Add(-2 * time.Hour).Unix(),
		EndTime:     time.Now().Add(-time.Minute).Unix(),
	}
	require.NoError(t, DB.Create(sub).Error)

	hasSub, err := HasActiveUserSubscription(sub.UserId)
	require.NoError(t, err)
	assert.False(t, hasSub)
}

func TestCleanupExpiredSubStatusCache(t *testing.T) {
	subStatusCache.Store(201, &subStatusEntry{hasActive: true, expireAt: time.Now().Add(-time.Second)})
	subStatusCache.Store(202, &subStatusEntry{hasActive: true, expireAt: time.Now().Add(time.Minute)})
	t.Cleanup(func() {
		subStatusCache.Delete(201)
		subStatusCache.Delete(202)
	})

	cleanupExpiredSubStatusCache(time.Now())

	_, expiredExists := subStatusCache.Load(201)
	_, liveExists := subStatusCache.Load(202)
	assert.False(t, expiredExists)
	assert.True(t, liveExists)
}

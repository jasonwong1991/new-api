package controller

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	checkInTableReady bool
	checkInTableMu    sync.Mutex
	errAlreadyChecked = errors.New("already checked in today")
)

type CheckInRecord struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"index:idx_checkin_user_date,unique"`
	Date      string    `json:"date" gorm:"type:char(10);index:idx_checkin_user_date,unique"`
	Quota     int       `json:"quota"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (CheckInRecord) TableName() string {
	return "user_checkins"
}

func ensureCheckInTable() {
	if checkInTableReady {
		return
	}
	checkInTableMu.Lock()
	defer checkInTableMu.Unlock()
	if checkInTableReady {
		return
	}
	if model.DB == nil {
		return
	}
	if err := model.DB.AutoMigrate(&CheckInRecord{}); err != nil {
		common.SysLog("failed to migrate user_checkins table: " + err.Error())
		return
	}
	checkInTableReady = true
}

func todayDate() string {
	return time.Now().Format("2006-01-02")
}

func hasCheckedInToday(userID int) (*CheckInRecord, bool, error) {
	ensureCheckInTable()
	if model.DB == nil {
		return nil, false, errors.New("database not initialized")
	}
	var record CheckInRecord
	err := model.DB.Where("user_id = ? AND date = ?", userID, todayDate()).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &record, true, nil
}

func createCheckInRecord(userID, quota int) (*CheckInRecord, error) {
	ensureCheckInTable()
	if model.DB == nil {
		return nil, errors.New("database not initialized")
	}
	record := &CheckInRecord{
		UserID: userID,
		Date:   todayDate(),
		Quota:  quota,
	}
	if err := model.DB.Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func rollbackCheckInRecord(id int) {
	if id == 0 || model.DB == nil {
		return
	}
	if err := model.DB.Delete(&CheckInRecord{}, id).Error; err != nil {
		common.SysLog(fmt.Sprintf("failed to rollback check-in record %d: %v", id, err))
	}
}

func randomQuota(min, max int) int {
	if min >= max {
		return min
	}
	return rand.Intn(max-min+1) + min
}

func GetCheckInStatus(c *gin.Context) {
	cfg := operation_setting.GetCheckInSetting()
	minQuota, maxQuota := operation_setting.GetCheckInQuotaRange()
	resp := gin.H{
		"enabled":    cfg.Enabled,
		"min_quota":  minQuota,
		"max_quota":  maxQuota,
		"checked_in": false,
		"quota":      0,
		"checked_at": 0,
	}
	if !cfg.Enabled {
		common.ApiSuccess(c, resp)
		return
	}
	userID := c.GetInt("id")
	record, checked, err := hasCheckedInToday(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if checked {
		resp["checked_in"] = true
		resp["quota"] = record.Quota
		resp["checked_at"] = record.CreatedAt.Unix()
	}
	common.ApiSuccess(c, resp)
}

func PostCheckIn(c *gin.Context) {
	cfg := operation_setting.GetCheckInSetting()
	if !cfg.Enabled {
		common.ApiErrorMsg(c, "签到功能未开启")
		return
	}
	minQuota, maxQuota := operation_setting.GetCheckInQuotaRange()
	if maxQuota <= 0 {
		common.ApiErrorMsg(c, "管理员尚未配置有效的签到额度")
		return
	}
	userID := c.GetInt("id")
	if _, checked, err := hasCheckedInToday(userID); err != nil {
		common.ApiError(c, err)
		return
	} else if checked {
		common.ApiErrorMsg(c, "今天已经签到过啦")
		return
	}

	quota := randomQuota(minQuota, maxQuota)
	record, err := createCheckInRecord(userID, quota)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.IncreaseUserQuota(userID, record.Quota, true); err != nil {
		rollbackCheckInRecord(record.ID)
		common.ApiError(c, fmt.Errorf("发放签到额度失败: %w", err))
		return
	}
	model.RecordLog(userID, model.LogTypeSystem, fmt.Sprintf("每日签到获得 %s", logger.LogQuota(record.Quota)))
	common.ApiSuccess(c, gin.H{
		"checked_in": true,
		"quota":      record.Quota,
		"checked_at": record.CreatedAt.Unix(),
		"min_quota":  minQuota,
		"max_quota":  maxQuota,
	})
}

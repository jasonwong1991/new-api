package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type InvitationCode struct {
	Id          int    `json:"id" gorm:"primaryKey"`
	Code        string `json:"code" gorm:"type:varchar(32);uniqueIndex"`
	Name        string `json:"name" gorm:"type:varchar(64)"`
	CreatedBy   int    `json:"created_by" gorm:"index"`
	CreatedTime int64  `json:"created_time" gorm:"bigint"`
	ExpiredTime int64  `json:"expired_time" gorm:"bigint;default:-1"` // -1 means never expired
	MaxUses     int    `json:"max_uses" gorm:"default:1"`             // -1 means unlimited
	UsedCount   int    `json:"used_count" gorm:"default:0"`
	Status      int    `json:"status" gorm:"default:1"` // 1: enabled, 2: disabled
}

const (
	InvitationCodeStatusEnabled  = 1
	InvitationCodeStatusDisabled = 2
)

func GetAllInvitationCodes(pageInfo *common.PageInfo) (codes []*InvitationCode, total int64, err error) {
	db := DB.Model(&InvitationCode{})
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&codes).Error
	return codes, total, err
}

func SearchInvitationCodes(keyword string, pageInfo *common.PageInfo) (codes []*InvitationCode, total int64, err error) {
	db := DB.Model(&InvitationCode{}).Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&codes).Error
	return codes, total, err
}

func GetInvitationCodeById(id int) (*InvitationCode, error) {
	var code InvitationCode
	err := DB.First(&code, id).Error
	return &code, err
}

func (code *InvitationCode) Insert() error {
	code.CreatedTime = common.GetTimestamp()
	return DB.Create(code).Error
}

func (code *InvitationCode) Update() error {
	return DB.Model(code).Select("name", "expired_time", "max_uses", "status").Updates(code).Error
}

func DeleteInvitationCode(id int) error {
	return DB.Delete(&InvitationCode{}, id).Error
}

// CheckInvitationCode validates an invitation code without consuming it
func CheckInvitationCode(codeStr string) error {
	var code InvitationCode
	err := DB.Where("code = ?", codeStr).First(&code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("邀请码不存在")
		}
		return err
	}
	if code.Status != InvitationCodeStatusEnabled {
		return errors.New("邀请码已被禁用")
	}
	if code.ExpiredTime != -1 && code.ExpiredTime < time.Now().Unix() {
		return errors.New("邀请码已过期")
	}
	if code.MaxUses != -1 && code.UsedCount >= code.MaxUses {
		return errors.New("邀请码已达到最大使用次数")
	}
	return nil
}

// RedeemInvitationCode validates and consumes an invitation code atomically
func RedeemInvitationCode(codeStr string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var code InvitationCode
		// Lock the row for update to prevent race conditions
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("code = ?", codeStr).First(&code).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("邀请码不存在")
			}
			return err
		}

		if code.Status != InvitationCodeStatusEnabled {
			return errors.New("邀请码已被禁用")
		}
		if code.ExpiredTime != -1 && code.ExpiredTime < time.Now().Unix() {
			return errors.New("邀请码已过期")
		}
		if code.MaxUses != -1 && code.UsedCount >= code.MaxUses {
			return errors.New("邀请码已达到最大使用次数")
		}

		code.UsedCount++
		if err := tx.Save(&code).Error; err != nil {
			return err
		}
		return nil
	})
}

// RevertInvitationCode reverts the usage of an invitation code atomically (for rollback on user creation failure)
func RevertInvitationCode(codeStr string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var code InvitationCode
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("code = ?", codeStr).First(&code).Error; err != nil {
			return err
		}
		if code.UsedCount > 0 {
			code.UsedCount--
			return tx.Save(&code).Error
		}
		return nil
	})
}

// BatchCreateInvitationCodes creates multiple invitation codes at once
func BatchCreateInvitationCodes(name string, count int, maxUses int, expiredTime int64, createdBy int) ([]*InvitationCode, error) {
	codes := make([]*InvitationCode, 0, count)
	now := common.GetTimestamp()

	for i := 0; i < count; i++ {
		code := &InvitationCode{
			Code:        common.GetRandomString(16),
			Name:        name,
			CreatedBy:   createdBy,
			CreatedTime: now,
			ExpiredTime: expiredTime,
			MaxUses:     maxUses,
			UsedCount:   0,
			Status:      InvitationCodeStatusEnabled,
		}
		codes = append(codes, code)
	}

	if err := DB.Create(&codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

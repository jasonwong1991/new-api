package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
)

const (
	AppealStatusPending  = 1
	AppealStatusApproved = 2
	AppealStatusRejected = 3
)

type Appeal struct {
	Id         int    `json:"id" gorm:"primaryKey"`
	UserId     int    `json:"user_id" gorm:"index;not null"`
	Reason     string `json:"reason" gorm:"type:text;not null"`
	Status     int    `json:"status" gorm:"type:int;default:1;index"`
	AdminNote  string `json:"admin_note" gorm:"type:text"`
	ReviewedBy int    `json:"reviewed_by" gorm:"type:int;default:0"`
	ReviewedAt int64  `json:"reviewed_at" gorm:"bigint;default:0"`
	CreatedAt  int64  `json:"created_at" gorm:"bigint;not null;index"`
}

type AppealWithUser struct {
	Appeal
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	LinuxDOUsername string `json:"linux_do_username"`
	LinuxDOAvatar   string `json:"linux_do_avatar"`
	BanReason       string `json:"ban_reason"`
}

func (a *Appeal) Insert() error {
	a.CreatedAt = common.GetTimestamp()
	a.Status = AppealStatusPending
	return DB.Create(a).Error
}

func (a *Appeal) Update() error {
	return DB.Save(a).Error
}

func GetAppealById(id int) (*Appeal, error) {
	var appeal Appeal
	err := DB.First(&appeal, id).Error
	return &appeal, err
}

func GetPendingAppealByUserId(userId int) (*Appeal, error) {
	var appeal Appeal
	err := DB.Where("user_id = ? AND status = ?", userId, AppealStatusPending).First(&appeal).Error
	return &appeal, err
}

func HasPendingAppeal(userId int) bool {
	var count int64
	DB.Model(&Appeal{}).Where("user_id = ? AND status = ?", userId, AppealStatusPending).Count(&count)
	return count > 0
}

func GetAppealsByUserId(userId int) ([]*Appeal, error) {
	var appeals []*Appeal
	err := DB.Where("user_id = ?", userId).Order("created_at DESC").Find(&appeals).Error
	return appeals, err
}

func GetAllAppeals(status int, page, pageSize int) ([]*AppealWithUser, int64, error) {
	var appeals []*AppealWithUser
	var total int64

	query := DB.Table("appeals").
		Select("appeals.*, users.username, users.display_name, users.linux_do_username, users.linux_do_avatar, users.remark as ban_reason").
		Joins("LEFT JOIN users ON appeals.user_id = users.id")

	if status > 0 {
		query = query.Where("appeals.status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("appeals.created_at DESC").Offset(offset).Limit(pageSize).Find(&appeals).Error
	return appeals, total, err
}

func ApproveAppeal(appealId int, adminId int, note string) error {
	appeal, err := GetAppealById(appealId)
	if err != nil {
		return err
	}
	if appeal.Status != AppealStatusPending {
		return errors.New("该申诉已被处理")
	}

	appeal.Status = AppealStatusApproved
	appeal.ReviewedBy = adminId
	appeal.ReviewedAt = common.GetTimestamp()
	appeal.AdminNote = note

	if err := appeal.Update(); err != nil {
		return err
	}

	user, err := GetUserById(appeal.UserId, false)
	if err != nil {
		return err
	}
	user.Status = common.UserStatusEnabled
	user.Remark = ""
	return user.Update(false)
}

func RejectAppeal(appealId int, adminId int, note string) error {
	appeal, err := GetAppealById(appealId)
	if err != nil {
		return err
	}
	if appeal.Status != AppealStatusPending {
		return errors.New("该申诉已被处理")
	}

	appeal.Status = AppealStatusRejected
	appeal.ReviewedBy = adminId
	appeal.ReviewedAt = common.GetTimestamp()
	appeal.AdminNote = note

	return appeal.Update()
}

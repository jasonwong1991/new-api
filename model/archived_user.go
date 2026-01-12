package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type ArchivedUser struct {
	Id                 int    `json:"id" gorm:"primaryKey;autoIncrement"`
	OriginalUserId     int    `json:"original_user_id" gorm:"index"`
	Username           string `json:"username" gorm:"index"`
	Password           string `json:"-"`
	DisplayName        string `json:"display_name"`
	Role               int    `json:"role"`
	Status             int    `json:"status"`
	Email              string `json:"email" gorm:"index"`
	GitHubId           string `json:"github_id"`
	DiscordId          string `json:"discord_id"`
	OidcId             string `json:"oidc_id"`
	WeChatId           string `json:"wechat_id"`
	TelegramId         string `json:"telegram_id"`
	Quota              int    `json:"quota"`
	UsedQuota          int    `json:"used_quota"`
	RequestCount       int    `json:"request_count"`
	Group              string `json:"group"`
	AffCode            string `json:"aff_code"`
	AffCount           int    `json:"aff_count"`
	AffQuota           int    `json:"aff_quota"`
	AffHistoryQuota    int    `json:"aff_history_quota"`
	InviterId          int    `json:"inviter_id"`
	LinuxDOId          string `json:"linux_do_id"`
	LinuxDOUsername    string `json:"linux_do_username"`
	LinuxDOAvatar      string `json:"linux_do_avatar"`
	LinuxDOLevel       int    `json:"linux_do_level"`
	Setting            string `json:"setting"`
	Remark             string `json:"remark"`
	StripeCustomer     string `json:"stripe_customer"`
	CreatedAt          int64  `json:"created_at"`
	InvitationCodeUsed string `json:"invitation_code_used"`
	ArchivedAt         int64  `json:"archived_at" gorm:"bigint;index"`
	ArchivedBy         int    `json:"archived_by" gorm:"index"`
	ArchivedReason     string `json:"archived_reason" gorm:"type:varchar(255)"`
}

func (ArchivedUser) TableName() string {
	return "archived_users"
}

func GetAllArchivedUsers(pageInfo *common.PageInfo) (users []*ArchivedUser, total int64, err error) {
	db := DB.Model(&ArchivedUser{})
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Order("archived_at desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&users).Error
	return users, total, err
}

func SearchArchivedUsers(keyword string, pageInfo *common.PageInfo) (users []*ArchivedUser, total int64, err error) {
	db := DB.Model(&ArchivedUser{}).Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Order("archived_at desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&users).Error
	return users, total, err
}

func GetArchivedUserById(id int) (*ArchivedUser, error) {
	var user ArchivedUser
	err := DB.First(&user, id).Error
	return &user, err
}

func FindArchivedUserByKeyword(keyword string) (*ArchivedUser, error) {
	var user ArchivedUser
	err := DB.Where("username = ? OR display_name = ? OR CAST(original_user_id AS CHAR) = ?", keyword, keyword, keyword).First(&user).Error
	return &user, err
}

func (au *ArchivedUser) Delete() error {
	return DB.Delete(au).Error
}

func GetInactiveUsers(minDays int, startId int, endId int) ([]*User, error) {
	threshold := time.Now().AddDate(0, 0, -minDays).Unix()
	var users []*User
	query := DB.Where("used_quota = 0 AND request_count = 0 AND (invitation_code_used IS NULL OR invitation_code_used = '') AND role = ? AND deleted_at IS NULL AND (created_at IS NULL OR created_at = 0 OR created_at <= ?)",
		common.RoleCommonUser, threshold)
	if startId > 0 {
		query = query.Where("id >= ?", startId)
	}
	if endId > 0 {
		query = query.Where("id <= ?", endId)
	}
	err := query.Find(&users).Error
	return users, err
}

func CountInactiveUsers(minDays int, startId int, endId int) (int64, error) {
	threshold := time.Now().AddDate(0, 0, -minDays).Unix()
	var count int64
	query := DB.Model(&User{}).Where("used_quota = 0 AND request_count = 0 AND (invitation_code_used IS NULL OR invitation_code_used = '') AND role = ? AND deleted_at IS NULL AND (created_at IS NULL OR created_at = 0 OR created_at <= ?)",
		common.RoleCommonUser, threshold)
	if startId > 0 {
		query = query.Where("id >= ?", startId)
	}
	if endId > 0 {
		query = query.Where("id <= ?", endId)
	}
	err := query.Count(&count).Error
	return count, err
}

func ArchiveAndDeleteUser(user *User, archivedBy int, reason string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		archived := &ArchivedUser{
			OriginalUserId:     user.Id,
			Username:           user.Username,
			Password:           user.Password,
			DisplayName:        user.DisplayName,
			Role:               user.Role,
			Status:             user.Status,
			Email:              user.Email,
			GitHubId:           user.GitHubId,
			DiscordId:          user.DiscordId,
			OidcId:             user.OidcId,
			WeChatId:           user.WeChatId,
			TelegramId:         user.TelegramId,
			Quota:              user.Quota,
			UsedQuota:          user.UsedQuota,
			RequestCount:       user.RequestCount,
			Group:              user.Group,
			AffCode:            user.AffCode,
			AffCount:           user.AffCount,
			AffQuota:           user.AffQuota,
			AffHistoryQuota:    user.AffHistoryQuota,
			InviterId:          user.InviterId,
			LinuxDOId:          user.LinuxDOId,
			LinuxDOUsername:    user.LinuxDOUsername,
			LinuxDOAvatar:      user.LinuxDOAvatar,
			LinuxDOLevel:       user.LinuxDOLevel,
			Setting:            user.Setting,
			Remark:             user.Remark,
			StripeCustomer:     user.StripeCustomer,
			CreatedAt:          user.CreatedAt,
			InvitationCodeUsed: user.InvitationCodeUsed,
			ArchivedAt:         common.GetTimestamp(),
			ArchivedBy:         archivedBy,
			ArchivedReason:     reason,
		}
		if err := tx.Create(archived).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.Id).Delete(&Token{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&User{}, "id = ?", user.Id).Error; err != nil {
			return err
		}
		return nil
	})
}

func BatchArchiveInactiveUsers(minDays int, startId int, endId int, archivedBy int, reason string) (int, error) {
	users, err := GetInactiveUsers(minDays, startId, endId)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, user := range users {
		if err := ArchiveAndDeleteUser(user, archivedBy, reason); err != nil {
			common.SysLog("failed to archive user " + user.Username + ": " + err.Error())
			continue
		}
		count++
	}
	return count, nil
}

func RestoreArchivedUser(archivedId int) error {
	archived, err := GetArchivedUserById(archivedId)
	if err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		exist, err := CheckUserExistOrDeleted(archived.Username, archived.Email)
		if err != nil {
			return err
		}
		if exist {
			return errors.New("用户名或邮箱已被占用")
		}
		user := &User{
			Username:           archived.Username,
			Password:           archived.Password,
			DisplayName:        archived.DisplayName,
			Role:               archived.Role,
			Status:             archived.Status,
			Email:              archived.Email,
			GitHubId:           archived.GitHubId,
			DiscordId:          archived.DiscordId,
			OidcId:             archived.OidcId,
			WeChatId:           archived.WeChatId,
			TelegramId:         archived.TelegramId,
			Quota:              archived.Quota,
			UsedQuota:          archived.UsedQuota,
			RequestCount:       archived.RequestCount,
			Group:              archived.Group,
			AffCode:            common.GetRandomString(4),
			AffCount:           archived.AffCount,
			AffQuota:           archived.AffQuota,
			AffHistoryQuota:    archived.AffHistoryQuota,
			InviterId:          archived.InviterId,
			LinuxDOId:          archived.LinuxDOId,
			LinuxDOUsername:    archived.LinuxDOUsername,
			LinuxDOAvatar:      archived.LinuxDOAvatar,
			LinuxDOLevel:       archived.LinuxDOLevel,
			Setting:            archived.Setting,
			Remark:             archived.Remark,
			StripeCustomer:     archived.StripeCustomer,
			CreatedAt:          archived.CreatedAt,
			InvitationCodeUsed: archived.InvitationCodeUsed,
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if err := tx.Delete(&ArchivedUser{}, "id = ?", archivedId).Error; err != nil {
			return err
		}
		return nil
	})
}

// FindArchivedUserByLinuxDOUsername finds archived user by LinuxDO username
func FindArchivedUserByLinuxDOUsername(linuxDOUsername string) (*ArchivedUser, error) {
	if linuxDOUsername == "" {
		return nil, errors.New("LinuxDO username 不能为空")
	}
	var user ArchivedUser
	err := DB.Where("linux_do_username = ?", linuxDOUsername).Order("archived_at DESC").First(&user).Error
	return &user, err
}

// RecoverQuotaToUser transfers quota from archived user to current user
func RecoverQuotaToUser(currentUserId int, archivedId int) (int, error) {
	var recoveredQuota int
	err := DB.Transaction(func(tx *gorm.DB) error {
		var currentUser User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&currentUser, currentUserId).Error; err != nil {
			return errors.New("当前用户不存在")
		}

		var archivedUser ArchivedUser
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&archivedUser, archivedId).Error; err != nil {
			return errors.New("归档用户不存在")
		}

		if currentUser.LinuxDOUsername == "" || currentUser.LinuxDOUsername != archivedUser.LinuxDOUsername {
			return errors.New("LinuxDO 用户名不匹配，无法恢复额度")
		}

		if archivedUser.Quota <= 0 {
			return errors.New("归档用户没有可恢复的额度")
		}

		recoveredQuota = archivedUser.Quota
		if err := tx.Model(&currentUser).Update("quota", recoveredQuota).Error; err != nil {
			return err
		}

		return tx.Delete(&archivedUser).Error
	})
	return recoveredQuota, err
}

package controller

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupInvitationControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)

	oldUsingSQLite := common.UsingSQLite
	oldUsingMySQL := common.UsingMySQL
	oldUsingPostgreSQL := common.UsingPostgreSQL
	oldRedisEnabled := common.RedisEnabled
	oldRegisterEnabled := common.RegisterEnabled
	oldPasswordRegisterEnabled := common.PasswordRegisterEnabled
	oldEmailVerificationEnabled := common.EmailVerificationEnabled
	oldInvitationCodeRequired := common.InvitationCodeRequired
	oldQuotaForNewUser := common.QuotaForNewUser
	oldQuotaForInvitee := common.QuotaForInvitee
	oldQuotaForInviter := common.QuotaForInviter
	oldGenerateDefaultToken := constant.GenerateDefaultToken

	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	common.RegisterEnabled = true
	common.PasswordRegisterEnabled = true
	common.EmailVerificationEnabled = false
	common.InvitationCodeRequired = false
	common.QuotaForNewUser = 0
	common.QuotaForInvitee = 0
	common.QuotaForInviter = 0
	constant.GenerateDefaultToken = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.InvitationCode{}, &model.Log{}))

	t.Cleanup(func() {
		common.UsingSQLite = oldUsingSQLite
		common.UsingMySQL = oldUsingMySQL
		common.UsingPostgreSQL = oldUsingPostgreSQL
		common.RedisEnabled = oldRedisEnabled
		common.RegisterEnabled = oldRegisterEnabled
		common.PasswordRegisterEnabled = oldPasswordRegisterEnabled
		common.EmailVerificationEnabled = oldEmailVerificationEnabled
		common.InvitationCodeRequired = oldInvitationCodeRequired
		common.QuotaForNewUser = oldQuotaForNewUser
		common.QuotaForInvitee = oldQuotaForInvitee
		common.QuotaForInviter = oldQuotaForInviter
		constant.GenerateDefaultToken = oldGenerateDefaultToken

		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func createInvitationCode(t *testing.T, db *gorm.DB, code string, maxUses int, usedCount int) {
	t.Helper()

	invitationCode := &model.InvitationCode{
		Code:        code,
		Name:        "test",
		CreatedBy:   1,
		CreatedTime: common.GetTimestamp(),
		ExpiredTime: -1,
		MaxUses:     maxUses,
		UsedCount:   usedCount,
		Status:      model.InvitationCodeStatusEnabled,
	}
	require.NoError(t, db.Create(invitationCode).Error)
}

func TestRedeemInvitationCodeIfPresent_OptionalCodeStillConsumesUsage(t *testing.T) {
	db := setupInvitationControllerTestDB(t)
	createInvitationCode(t, db, "INVITE123", 2, 0)

	require.NoError(t, redeemInvitationCodeIfPresent("INVITE123"))

	var updated model.InvitationCode
	require.NoError(t, db.Where("code = ?", "INVITE123").First(&updated).Error)
	require.Equal(t, 1, updated.UsedCount)
}

func TestRegisterConsumesInvitationCodeAndTracksUser(t *testing.T) {
	db := setupInvitationControllerTestDB(t)
	common.InvitationCodeRequired = true
	createInvitationCode(t, db, "BOUND123", 1, 0)

	ctx, recorder := newAuthenticatedContext(t, http.MethodPost, "/api/user/register", map[string]any{
		"username":        "invite_user",
		"password":        "password123",
		"invitation_code": "BOUND123",
	}, 0)

	Register(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)

	var createdUser model.User
	require.NoError(t, db.Where("username = ?", "invite_user").First(&createdUser).Error)
	require.Equal(t, "BOUND123", createdUser.InvitationCodeUsed)

	var updated model.InvitationCode
	require.NoError(t, db.Where("code = ?", "BOUND123").First(&updated).Error)
	require.Equal(t, 1, updated.UsedCount)
}

func TestRegisterRejectsReusedBoundInvitationCode(t *testing.T) {
	db := setupInvitationControllerTestDB(t)
	common.InvitationCodeRequired = true
	createInvitationCode(t, db, "ONCEONLY", 1, 1)

	ctx, recorder := newAuthenticatedContext(t, http.MethodPost, "/api/user/register", map[string]any{
		"username":        "second_user",
		"password":        "password123",
		"invitation_code": "ONCEONLY",
	}, 0)

	Register(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)

	var userCount int64
	require.NoError(t, db.Model(&model.User{}).Where("username = ?", "second_user").Count(&userCount).Error)
	require.Zero(t, userCount)

	var updated model.InvitationCode
	require.NoError(t, db.Where("code = ?", "ONCEONLY").First(&updated).Error)
	require.Equal(t, 1, updated.UsedCount)
}

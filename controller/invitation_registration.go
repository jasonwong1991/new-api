package controller

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

func redeemInvitationCodeIfPresent(code string) error {
	if common.InvitationCodeRequired && code == "" {
		return errors.New("管理员开启了邀请注册，请先填写邀请码")
	}
	if code == "" {
		return nil
	}
	if err := model.CheckInvitationCode(code); err != nil {
		return fmt.Errorf("邀请码无效: %s", err.Error())
	}
	if err := model.RedeemInvitationCode(code); err != nil {
		return fmt.Errorf("邀请码核销失败: %s", err.Error())
	}
	return nil
}

func revertInvitationCodeIfPresent(code string) {
	if code == "" {
		return
	}
	if err := model.RevertInvitationCode(code); err != nil {
		common.SysError(fmt.Sprintf("failed to revert invitation code %s: %v", code, err))
	}
}

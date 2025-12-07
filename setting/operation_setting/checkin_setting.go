package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type CheckInSetting struct {
	Enabled  bool `json:"enabled"`
	MinQuota int  `json:"min_quota"`
	MaxQuota int  `json:"max_quota"`
}

var checkInSetting = CheckInSetting{
	Enabled:  false,
	MinQuota: 1000,
	MaxQuota: 5000,
}

func init() {
	config.GlobalConfig.Register("checkin_setting", &checkInSetting)
}

func GetCheckInSetting() *CheckInSetting {
	return &checkInSetting
}

func IsCheckInEnabled() bool {
	return checkInSetting.Enabled
}

func GetCheckInQuotaRange() (int, int) {
	min := checkInSetting.MinQuota
	max := checkInSetting.MaxQuota
	if min <= 0 {
		min = 1
	}
	if max < min {
		max = min
	}
	return min, max
}

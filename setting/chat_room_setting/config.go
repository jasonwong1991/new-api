package chat_room_setting

import "github.com/QuantumNous/new-api/setting/config"

type ChatRoomSetting struct {
	Enabled          bool   `json:"enabled"`
	MessageLimit     int    `json:"message_limit"`
	MaxMessageLength int    `json:"max_message_length"`
	// Image settings
	ImageEnabled       bool     `json:"image_enabled"`
	ImageDir           string   `json:"image_dir"`
	ImageMaxBytes      int64    `json:"image_max_bytes"`
	ImageCacheMaxBytes int64    `json:"image_cache_max_bytes"`
	// Anti-hotlinking settings
	AntiHotlinkEnabled bool     `json:"anti_hotlink_enabled"`
	AllowedReferers    []string `json:"allowed_referers"`
}

var defaultChatRoomSetting = ChatRoomSetting{
	Enabled:            true,
	MessageLimit:       1000,
	MaxMessageLength:   8000,
	ImageEnabled:       true,
	ImageDir:           "data/chat_room_images",
	ImageMaxBytes:      10 << 20,
	ImageCacheMaxBytes: 1 << 30,
	AntiHotlinkEnabled: true,
	AllowedReferers:    []string{},
}

var chatRoomSetting = defaultChatRoomSetting

func init() {
	config.GlobalConfig.Register("chat_room_setting", &chatRoomSetting)
}

func GetChatRoomSetting() *ChatRoomSetting {
	return &chatRoomSetting
}

package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for StringArray")
	}
	return json.Unmarshal(bytes, a)
}

type ChatMessage struct {
	Id              int64       `json:"id" gorm:"primaryKey;autoIncrement;index:idx_chat_room_id,priority:2"`
	Room            string      `json:"room" gorm:"type:varchar(64);not null;default:'global';index:idx_chat_room_id,priority:1"`
	UserId          int         `json:"user_id" gorm:"type:int;not null;index"`
	Username        string      `json:"username" gorm:"type:varchar(64);not null;index"`
	DisplayName     string      `json:"display_name" gorm:"type:varchar(64);not null"`
	Content         string      `json:"content" gorm:"type:text;not null"`
	ImageUrls       StringArray `json:"image_urls" gorm:"type:json"`
	ImageTotalBytes int64       `json:"image_total_bytes" gorm:"type:bigint;default:0"`
	CreatedAt       int64       `json:"created_at" gorm:"bigint;index;autoCreateTime"`
	// Output-only fields (not stored in DB)
	Avatar      string `json:"avatar" gorm:"-:all"`
	Quota       int    `json:"quota" gorm:"-:all"`
	UsedQuota   int    `json:"used_quota" gorm:"-:all"`
	UsageRank   int    `json:"usage_rank" gorm:"-:all"`
	BalanceRank int    `json:"balance_rank" gorm:"-:all"`
}

func (m *ChatMessage) Insert() error {
	hasContent := strings.TrimSpace(m.Content) != ""
	hasImages := len(m.ImageUrls) > 0
	if !hasContent && !hasImages {
		return errors.New("消息内容不能为空")
	}
	return DB.Create(m).Error
}

// ListChatMessages returns messages in chronological order (old -> new).
// beforeId > 0 means only return messages with id < beforeId.
func ListChatMessages(room string, limit int, beforeId int64) ([]ChatMessage, error) {
	room = strings.TrimSpace(room)
	if room == "" {
		room = "global"
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 5000 {
		limit = 5000
	}

	query := DB.Where("room = ?", room)
	if beforeId > 0 {
		query = query.Where("id < ?", beforeId)
	}

	var messages []ChatMessage
	if err := query.Order("id desc").Limit(limit).Find(&messages).Error; err != nil {
		return nil, err
	}

	// Reverse to chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// TrimChatMessages deletes old messages exceeding the limit and returns deleted messages.
func TrimChatMessages(room string, keep int) ([]ChatMessage, error) {
	room = strings.TrimSpace(room)
	if room == "" {
		room = "global"
	}
	if keep <= 0 {
		return nil, nil
	}

	var count int64
	if err := DB.Model(&ChatMessage{}).Where("room = ?", room).Count(&count).Error; err != nil {
		return nil, err
	}

	if count <= int64(keep) {
		return nil, nil
	}

	deleteCount := int(count) - keep
	var toDelete []ChatMessage
	if err := DB.Where("room = ?", room).Order("id asc").Limit(deleteCount).Find(&toDelete).Error; err != nil {
		return nil, err
	}

	if len(toDelete) == 0 {
		return nil, nil
	}

	ids := make([]int64, len(toDelete))
	for i, m := range toDelete {
		ids[i] = m.Id
	}

	if err := DB.Where("id IN ?", ids).Delete(&ChatMessage{}).Error; err != nil {
		return nil, err
	}

	return toDelete, nil
}

// GetImageCacheTotalBytes returns total bytes of all images in chat messages.
func GetImageCacheTotalBytes() (int64, error) {
	var total int64
	err := DB.Model(&ChatMessage{}).Select("COALESCE(SUM(image_total_bytes), 0)").Scan(&total).Error
	return total, err
}

// TrimImageCacheBySize deletes oldest messages with images until total size is under maxBytes.
func TrimImageCacheBySize(maxBytes int64, imageDir string) error {
	for {
		total, err := GetImageCacheTotalBytes()
		if err != nil {
			return err
		}
		if total <= maxBytes {
			return nil
		}

		var oldest ChatMessage
		err = DB.Where("image_total_bytes > 0").Order("id asc").First(&oldest).Error
		if err != nil {
			return nil
		}

		DeleteChatMessageImages([]ChatMessage{oldest}, imageDir)

		if err := DB.Delete(&oldest).Error; err != nil {
			return err
		}
	}
}

// DeleteChatMessageImages deletes image files associated with messages.
func DeleteChatMessageImages(messages []ChatMessage, imageDir string) {
	for _, m := range messages {
		for _, url := range m.ImageUrls {
			if !strings.HasPrefix(url, "/api/chat/images/") {
				continue
			}
			relativePath := strings.TrimPrefix(url, "/api/chat/images/")
			filePath := filepath.Join(imageDir, relativePath)
			cleanPath := filepath.Clean(filePath)
			absImageDir, _ := filepath.Abs(imageDir)
			absCleanPath, _ := filepath.Abs(cleanPath)
			if !strings.HasPrefix(absCleanPath, absImageDir) {
				continue
			}
			_ = os.Remove(cleanPath)
		}
	}
}

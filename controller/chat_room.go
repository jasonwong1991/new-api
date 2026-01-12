package controller

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/chat_room_setting"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type chatRoomConfigDTO struct {
	Enabled          bool   `json:"enabled"`
	MessageLimit     int    `json:"message_limit"`
	MaxMessageLength int    `json:"max_message_length"`
	WsPath           string `json:"ws_path"`
	ImageEnabled     bool   `json:"image_enabled"`
	Announcement     string `json:"announcement"`
}

type chatRoomWsEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type chatRoomInitData struct {
	Room     string              `json:"room"`
	Config   chatRoomConfigDTO   `json:"config"`
	Messages []model.ChatMessage `json:"messages"`
}

type chatRoomMessageData struct {
	Message model.ChatMessage `json:"message"`
}

type chatRoomRestSendRequest struct {
	Content   string   `json:"content"`
	Room      string   `json:"room"`
	ImageUrls []string `json:"image_urls"`
}

type chatRoomWsClientMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type chatRoomWsSendData struct {
	Content   string   `json:"content"`
	Room      string   `json:"room"`
	ImageUrls []string `json:"image_urls"`
}

type chatRoomImageUploadResponse struct {
	Url         string `json:"url"`
	Bytes       int64  `json:"bytes"`
	ContentType string `json:"content_type"`
}

const (
	chatRoomWsWriteWait        = 10 * time.Second
	chatRoomWsPongWait         = 60 * time.Second
	chatRoomWsPingPeriod       = (chatRoomWsPongWait * 9) / 10
	chatRoomWsMaxIncomingBytes = 512 * 1024
	defaultAvatarUrl           = "/avatar.png"
)

var chatRoomUpgrader = websocket.Upgrader{
	Subprotocols: []string{"chat"},
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, r.Host)
	},
}

var allowedImageMimes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func normalizeChatRoomMessageLimit(v int) int {
	if v <= 0 {
		return 1000
	}
	if v > 5000 {
		return 5000
	}
	return v
}

func normalizeChatRoomMaxMessageLength(v int) int {
	if v <= 0 {
		return 8000
	}
	if v > 50000 {
		return 50000
	}
	return v
}

func normalizeChatRoomRoom(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "global"
	}
	if len(v) > 64 {
		return v[:64]
	}
	return v
}

func enrichChatMessage(msg *model.ChatMessage) {
	if msg.UserId <= 0 {
		msg.Avatar = defaultAvatarUrl
		return
	}

	user, err := model.GetUserById(msg.UserId, false)
	if err != nil || user == nil {
		msg.Avatar = defaultAvatarUrl
		return
	}

	if user.LinuxDOAvatar != "" {
		msg.Avatar = user.LinuxDOAvatar
	} else {
		msg.Avatar = defaultAvatarUrl
	}
	msg.Quota = user.Quota
	msg.UsedQuota = user.UsedQuota
}

func enrichChatMessages(messages []model.ChatMessage) {
	for i := range messages {
		enrichChatMessage(&messages[i])
	}
}

func GetChatRoomConfig(c *gin.Context) {
	cfg := chat_room_setting.GetChatRoomSetting()
	common.ApiSuccess(c, chatRoomConfigDTO{
		Enabled:          cfg.Enabled,
		MessageLimit:     normalizeChatRoomMessageLimit(cfg.MessageLimit),
		MaxMessageLength: normalizeChatRoomMaxMessageLength(cfg.MaxMessageLength),
		WsPath:           "/api/chat/ws",
		ImageEnabled:     cfg.ImageEnabled,
		Announcement:     cfg.Announcement,
	})
}

func GetChatRoomMessageCount(c *gin.Context) {
	cfg := chat_room_setting.GetChatRoomSetting()
	if !cfg.Enabled {
		common.ApiSuccess(c, gin.H{"count": 0, "enabled": false})
		return
	}

	room := normalizeChatRoomRoom(c.Query("room"))
	var count int64
	if err := model.DB.Model(&model.ChatMessage{}).Where("room = ?", room).Count(&count).Error; err != nil {
		common.ApiSuccess(c, gin.H{"count": 0, "enabled": true})
		return
	}

	common.ApiSuccess(c, gin.H{"count": count, "enabled": true})
}

func ListChatRoomMessages(c *gin.Context) {
	cfg := chat_room_setting.GetChatRoomSetting()
	if !cfg.Enabled {
		common.ApiErrorMsg(c, "聊天室已关闭")
		return
	}

	room := normalizeChatRoomRoom(c.Query("room"))
	limit := normalizeChatRoomMessageLimit(cfg.MessageLimit)
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n < limit {
				limit = n
			}
		}
	}

	var beforeId int64
	if v := c.Query("before_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			beforeId = n
		}
	}

	messages, err := model.ListChatMessages(room, limit, beforeId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	enrichChatMessages(messages)
	common.ApiSuccess(c, messages)
}

func PostChatRoomMessage(c *gin.Context) {
	cfg := chat_room_setting.GetChatRoomSetting()
	if !cfg.Enabled {
		common.ApiErrorMsg(c, "聊天室已关闭")
		return
	}

	var req chatRoomRestSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "无效的参数")
		return
	}

	content := strings.TrimSpace(req.Content)
	hasContent := content != ""
	hasImages := len(req.ImageUrls) > 0

	if !hasContent && !hasImages {
		common.ApiErrorMsg(c, "消息内容不能为空")
		return
	}

	maxLen := normalizeChatRoomMaxMessageLength(cfg.MaxMessageLength)
	if utf8.RuneCountInString(content) > maxLen {
		common.ApiErrorMsg(c, "消息过长")
		return
	}

	room := normalizeChatRoomRoom(req.Room)
	userId := c.GetInt("id")
	username := c.GetString("username")

	displayName := username
	if userId > 0 {
		if user, err := model.GetUserById(userId, false); err == nil && user != nil && user.DisplayName != "" {
			displayName = user.DisplayName
		}
	}

	msg := &model.ChatMessage{
		Room:        room,
		UserId:      userId,
		Username:    username,
		DisplayName: displayName,
		Content:     content,
		ImageUrls:   req.ImageUrls,
	}
	if err := msg.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}

	trimAndCleanup(room, cfg)
	enrichChatMessage(msg)
	broadcastChatRoomMessage(*msg)
	common.ApiSuccess(c, msg)
}

func UploadChatRoomImage(c *gin.Context) {
	cfg := chat_room_setting.GetChatRoomSetting()
	if !cfg.Enabled {
		common.ApiErrorMsg(c, "聊天室已关闭")
		return
	}
	if !cfg.ImageEnabled {
		common.ApiErrorMsg(c, "图片功能已关闭")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.ImageMaxBytes+1024)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		common.ApiErrorMsg(c, "请选择要上传的图片")
		return
	}
	defer file.Close()

	if header.Size > cfg.ImageMaxBytes {
		common.ApiErrorMsg(c, fmt.Sprintf("图片大小不能超过 %d MB", cfg.ImageMaxBytes/(1<<20)))
		return
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		common.ApiErrorMsg(c, "读取文件失败")
		return
	}
	contentType := http.DetectContentType(buf[:n])

	ext, ok := allowedImageMimes[contentType]
	if !ok {
		common.ApiErrorMsg(c, "不支持的图片格式，仅支持 PNG/JPEG/GIF/WebP")
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		common.ApiErrorMsg(c, "读取文件失败")
		return
	}

	dateDir := time.Now().Format("20060102")
	dirPath := filepath.Join(cfg.ImageDir, dateDir)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		common.ApiErrorMsg(c, "创建目录失败")
		return
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		common.ApiErrorMsg(c, "生成文件名失败")
		return
	}
	fileName := hex.EncodeToString(randBytes) + ext
	filePath := filepath.Join(dirPath, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		common.ApiErrorMsg(c, "创建文件失败")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath)
		common.ApiErrorMsg(c, "保存文件失败")
		return
	}

	imageUrl := fmt.Sprintf("/api/chat/images/%s/%s", dateDir, fileName)

	common.ApiSuccess(c, chatRoomImageUploadResponse{
		Url:         imageUrl,
		Bytes:       written,
		ContentType: contentType,
	})
}

func GetChatRoomImage(c *gin.Context) {
	cfg := chat_room_setting.GetChatRoomSetting()

	if cfg.AntiHotlinkEnabled {
		referer := c.Request.Header.Get("Referer")
		if referer != "" {
			refererUrl, err := url.Parse(referer)
			if err == nil {
				allowed := false
				requestHost := c.Request.Host
				if strings.EqualFold(refererUrl.Host, requestHost) {
					allowed = true
				}
				for _, allowedReferer := range cfg.AllowedReferers {
					if strings.Contains(refererUrl.Host, allowedReferer) {
						allowed = true
						break
					}
				}
				if !allowed {
					c.Status(http.StatusForbidden)
					return
				}
			}
		}
	}

	dateDir := c.Param("date")
	fileName := c.Param("name")

	if dateDir == "" || fileName == "" {
		c.Status(http.StatusNotFound)
		return
	}

	if strings.Contains(dateDir, "..") || strings.Contains(fileName, "..") {
		c.Status(http.StatusForbidden)
		return
	}

	filePath := filepath.Join(cfg.ImageDir, dateDir, fileName)
	cleanPath := filepath.Clean(filePath)
	absImageDir, _ := filepath.Abs(cfg.ImageDir)
	absCleanPath, _ := filepath.Abs(cleanPath)
	if !strings.HasPrefix(absCleanPath, absImageDir) {
		c.Status(http.StatusForbidden)
		return
	}

	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		c.Status(http.StatusNotFound)
		return
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	contentType := "application/octet-stream"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(cleanPath)
}

func ChatRoomWs(c *gin.Context) {
	cfg := chat_room_setting.GetChatRoomSetting()
	if !cfg.Enabled {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "聊天室已关闭",
		})
		return
	}

	room := normalizeChatRoomRoom(c.Query("room"))
	conn, err := chatRoomUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.SetReadLimit(chatRoomWsMaxIncomingBytes)
	_ = conn.SetReadDeadline(time.Now().Add(chatRoomWsPongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(chatRoomWsPongWait))
		return nil
	})

	userId := c.GetInt("id")
	username := c.GetString("username")
	displayName := username
	if userId > 0 {
		if user, err := model.GetUserById(userId, false); err == nil && user != nil && user.DisplayName != "" {
			displayName = user.DisplayName
		}
	}

	client := service.NewChatRoomClient(128)
	hub := service.GetChatRoomHub()
	hub.Register(client)
	defer hub.Unregister(client)

	go chatRoomWritePump(conn, client)

	initLimit := normalizeChatRoomMessageLimit(cfg.MessageLimit)
	initMessages, _ := model.ListChatMessages(room, initLimit, 0)
	enrichChatMessages(initMessages)

	initPayload, _ := json.Marshal(chatRoomWsEvent{
		Type: "init",
		Data: chatRoomInitData{
			Room: room,
			Config: chatRoomConfigDTO{
				Enabled:          cfg.Enabled,
				MessageLimit:     initLimit,
				MaxMessageLength: normalizeChatRoomMaxMessageLength(cfg.MaxMessageLength),
				WsPath:           "/api/chat/ws",
				ImageEnabled:     cfg.ImageEnabled,
				Announcement:     cfg.Announcement,
			},
			Messages: initMessages,
		},
	})
	if initPayload != nil {
		client.Send <- initPayload
	}

	chatRoomReadPump(conn, client, room, userId, username, displayName)
}

func broadcastChatRoomMessage(message model.ChatMessage) {
	payload, err := json.Marshal(chatRoomWsEvent{
		Type: "message",
		Data: chatRoomMessageData{Message: message},
	})
	if err != nil {
		return
	}
	service.GetChatRoomHub().Broadcast(payload)
}

func sendChatRoomWsError(client *service.ChatRoomClient, message string) {
	if client == nil {
		return
	}
	payload, _ := json.Marshal(chatRoomWsEvent{
		Type: "error",
		Data: gin.H{"message": message},
	})
	if payload == nil {
		return
	}
	select {
	case client.Send <- payload:
	default:
	}
}

func chatRoomWritePump(conn *websocket.Conn, client *service.ChatRoomClient) {
	ticker := time.NewTicker(chatRoomWsPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case payload, ok := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(chatRoomWsWriteWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(chatRoomWsWriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func chatRoomReadPump(conn *websocket.Conn, client *service.ChatRoomClient, room string, userId int, username, displayName string) {
	cfg := chat_room_setting.GetChatRoomSetting()
	maxLen := normalizeChatRoomMaxMessageLength(cfg.MaxMessageLength)

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var envelope chatRoomWsClientMessage
		if err := json.Unmarshal(msgBytes, &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "send":
			var data chatRoomWsSendData
			if err := json.Unmarshal(envelope.Data, &data); err != nil {
				continue
			}
			content := strings.TrimSpace(data.Content)
			hasContent := content != ""
			hasImages := len(data.ImageUrls) > 0

			if !hasContent && !hasImages {
				continue
			}
			if utf8.RuneCountInString(content) > maxLen {
				sendChatRoomWsError(client, "消息过长")
				continue
			}

			finalRoom := room
			if data.Room != "" {
				finalRoom = normalizeChatRoomRoom(data.Room)
			}

			chatMsg := &model.ChatMessage{
				Room:        finalRoom,
				UserId:      userId,
				Username:    username,
				DisplayName: displayName,
				Content:     content,
				ImageUrls:   data.ImageUrls,
			}
			if err := chatMsg.Insert(); err != nil {
				continue
			}

			trimAndCleanup(finalRoom, cfg)
			enrichChatMessage(chatMsg)
			broadcastChatRoomMessage(*chatMsg)
		default:
			continue
		}
	}
}

func trimAndCleanup(room string, cfg *chat_room_setting.ChatRoomSetting) {
	limit := normalizeChatRoomMessageLimit(cfg.MessageLimit)
	deleted, err := model.TrimChatMessages(room, limit)
	if err == nil && len(deleted) > 0 {
		model.DeleteChatMessageImages(deleted, cfg.ImageDir)
	}

	if cfg.ImageCacheMaxBytes > 0 {
		_ = model.TrimImageCacheBySize(cfg.ImageCacheMaxBytes, cfg.ImageDir)
	}
}

type SetAnnouncementRequest struct {
	Announcement string `json:"announcement"`
}

func SetChatRoomAnnouncement(c *gin.Context) {
	var req SetAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "无效的请求参数")
		return
	}

	if err := model.UpdateOption("chat_room_setting.announcement", req.Announcement); err != nil {
		common.ApiError(c, err)
		return
	}

	cfg := chat_room_setting.GetChatRoomSetting()
	cfg.Announcement = req.Announcement

	broadcastAnnouncement(req.Announcement)

	common.ApiSuccess(c, gin.H{
		"message":      "公告设置成功",
		"announcement": req.Announcement,
	})
}

func broadcastAnnouncement(announcement string) {
	payload, err := json.Marshal(chatRoomWsEvent{
		Type: "announcement",
		Data: gin.H{"announcement": announcement},
	})
	if err != nil {
		return
	}
	service.GetChatRoomHub().Broadcast(payload)
}

package service

import "sync"

type ChatRoomClient struct {
	Send chan []byte
}

type ChatRoomHub struct {
	register   chan *ChatRoomClient
	unregister chan *ChatRoomClient
	broadcast  chan []byte
	clients    map[*ChatRoomClient]struct{}
}

var chatRoomHubOnce sync.Once
var chatRoomHub *ChatRoomHub

func GetChatRoomHub() *ChatRoomHub {
	chatRoomHubOnce.Do(func() {
		chatRoomHub = &ChatRoomHub{
			register:   make(chan *ChatRoomClient, 64),
			unregister: make(chan *ChatRoomClient, 64),
			broadcast:  make(chan []byte, 256),
			clients:    make(map[*ChatRoomClient]struct{}),
		}
		go chatRoomHub.run()
	})
	return chatRoomHub
}

func NewChatRoomClient(buffer int) *ChatRoomClient {
	if buffer <= 0 {
		buffer = 64
	}
	return &ChatRoomClient{
		Send: make(chan []byte, buffer),
	}
}

func (h *ChatRoomHub) Register(client *ChatRoomClient) {
	h.register <- client
}

func (h *ChatRoomHub) Unregister(client *ChatRoomClient) {
	h.unregister <- client
}

func (h *ChatRoomHub) Broadcast(payload []byte) {
	if payload == nil {
		return
	}
	h.broadcast <- payload
}

func (h *ChatRoomHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
		case payload := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.Send <- payload:
				default:
					delete(h.clients, client)
					close(client.Send)
				}
			}
		}
	}
}

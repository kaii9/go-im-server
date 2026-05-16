package ws

import (
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

var hubLog, _ = zap.NewProduction()

type Hub struct {
	clients    map[int64]*Client
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
}

var DefaultHub = &Hub{
	clients:    make(map[int64]*Client),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			h.broadcastOnlineStatus(client.UserID, true)

		case client := <-h.Unregister:
			h.mu.Lock()
			if c, ok := h.clients[client.UserID]; ok && c == client {
				delete(h.clients, client.UserID)
			}
			h.mu.Unlock()
			close(client.Send)
			h.broadcastOnlineStatus(client.UserID, false)
		}
	}
}

func (h *Hub) IsOnline(uid int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[uid]
	return ok
}

func (h *Hub) SendToUser(uid int64, msg interface{}) {
	h.mu.RLock()
	client, ok := h.clients[uid]
	h.mu.RUnlock()
	if ok {
		data, err := json.Marshal(msg)
		if err != nil {
			hubLog.Error("序列化消息失败", zap.Error(err))
			return
		}
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (h *Hub) broadcastOnlineStatus(uid int64, online bool) {
}

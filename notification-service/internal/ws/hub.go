package ws

import (
	"context"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	redisclient "github.com/richardktran/realtime-quiz/pkg/cache/redis"
)

type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type ClientInfo struct {
	UserID string
	Name   string
}

type Hub struct {
	mu          sync.RWMutex
	sessions    map[string]map[*websocket.Conn]*ClientInfo
	connSession map[*websocket.Conn]string
	hosts       map[string]string
	redis       *redisclient.Client
}

func NewHub(redis *redisclient.Client) *Hub {
	return &Hub{
		sessions:    make(map[string]map[*websocket.Conn]*ClientInfo),
		connSession: make(map[*websocket.Conn]string),
		hosts:       make(map[string]string),
		redis:       redis,
	}
}

func (h *Hub) Join(sessionID string, conn *websocket.Conn, info *ClientInfo, hostID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.sessions[sessionID]; !ok {
		h.sessions[sessionID] = make(map[*websocket.Conn]*ClientInfo)
	}
	h.sessions[sessionID][conn] = info
	h.connSession[conn] = sessionID

	if hostID != "" {
		h.hosts[sessionID] = hostID
	}

	ctx := context.Background()
	if err := h.redis.SAdd(ctx, redisclient.SessionConnsKey(sessionID), info.UserID); err != nil {
		log.Printf("failed to add conn to redis set: %v", err)
	}
}

func (h *Hub) Leave(conn *websocket.Conn) (sessionID, userID string, wasHost bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sessionID, ok := h.connSession[conn]
	if !ok {
		return "", "", false
	}

	info, ok := h.sessions[sessionID][conn]
	if !ok {
		delete(h.connSession, conn)
		return sessionID, "", false
	}

	userID = info.UserID
	wasHost = h.hosts[sessionID] == userID

	delete(h.sessions[sessionID], conn)
	delete(h.connSession, conn)

	if len(h.sessions[sessionID]) == 0 {
		delete(h.sessions, sessionID)
		delete(h.hosts, sessionID)
	}

	ctx := context.Background()
	if err := h.redis.SRem(ctx, redisclient.SessionConnsKey(sessionID), userID); err != nil {
		log.Printf("failed to remove conn from redis set: %v", err)
	}

	return sessionID, userID, wasHost
}

func (h *Hub) GetClient(conn *websocket.Conn) (*ClientInfo, string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessionID, ok := h.connSession[conn]
	if !ok {
		return nil, ""
	}

	info, ok := h.sessions[sessionID][conn]
	if !ok {
		return nil, sessionID
	}

	return info, sessionID
}

func (h *Hub) SetHost(sessionID, hostID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hosts[sessionID] = hostID
}

func (h *Hub) Broadcast(sessionID, eventType string, data interface{}) {
	h.mu.RLock()
	conns := h.sessions[sessionID]
	snapshot := make([]*websocket.Conn, 0, len(conns))
	for conn := range conns {
		snapshot = append(snapshot, conn)
	}
	h.mu.RUnlock()

	msg := Message{
		Type:    eventType,
		Payload: map[string]interface{}{"data": data},
	}

	for _, conn := range snapshot {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("failed to broadcast message: %v", err)
		}
	}
}

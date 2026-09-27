package ws

import (
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"
)

type Hub struct {
	// shopID -> clients
	clients    map[uint]map[*Client]bool
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
}

var DefaultHub *Hub

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.clients[client.ShopID] == nil {
				h.clients[client.ShopID] = make(map[*Client]bool)
			}
			h.clients[client.ShopID][client] = true
			h.mu.Unlock()
			log.Info().Uint("shop_id", client.ShopID).Msg("POS client connected")

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.ShopID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
				}
				if len(clients) == 0 {
					delete(h.clients, client.ShopID)
				}
			}
			h.mu.Unlock()
			log.Info().Uint("shop_id", client.ShopID).Msg("POS client disconnected")
		}
	}
}

// PushToShop 向指定店铺的所有 POS 客户端推送消息
func (h *Hub) PushToShop(shopID uint, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("marshal ws message error")
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[shopID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
	}
}

func (h *Hub) DisconnectShop(shopID uint) {
	h.mu.RLock()
	shopClients, ok := h.clients[shopID]
	if !ok || len(shopClients) == 0 {
		h.mu.RUnlock()
		return
	}
	clients := make([]*Client, 0, len(shopClients))
	for client := range shopClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		_ = client.Conn.Close()
	}
}

func InitHub() {
	DefaultHub = NewHub()
	go DefaultHub.Run()
}

// hub.go
package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	ID       string
	UserData UserData
	Hub      *Hub
	Send     chan []byte
}

type UserData struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	Image   string `json:"image,omitempty"`
	IsOwner bool   `json:"isOwner,omitempty"`
}

type Hub struct {
	Clients map[string]*Client
	Mutex   sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[string]*Client),
	}
}

func (h *Hub) Broadcast(senderID string, message []byte) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	for id, client := range h.Clients {
		if id != senderID {
			select {
			case client.Send <- message:
			default:
				log.Printf("Client %s channel full, skipping", id)
				close(client.Send)
				delete(h.Clients, id)
			}
		}
	}
}

func (h *Hub) AddClient(c *Client) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	h.Clients[c.ID] = c
	h.BroadcastConnectedUsers()
}

func (h *Hub) RemoveClient(id string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	if client, exists := h.Clients[id]; exists {
		close(client.Send)
		delete(h.Clients, id)
		h.BroadcastConnectedUsers()
	}
}

func (h *Hub) BroadcastConnectedUsers() {
	users := make([]UserData, 0, len(h.Clients))
	for _, client := range h.Clients {
		users = append(users, client.UserData)
	}

	message := map[string]interface{}{
		"type":  "users_update",
		"users": users,
	}

	msg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	for _, client := range h.Clients {
		select {
		case client.Send <- msg:
		default:
			log.Printf("Client %s channel full during users update", client.ID)
		}
	}
}

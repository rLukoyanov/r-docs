package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	ID   string
	Send chan []byte
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

func (h *Hub) Run() {

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

	delete(h.Clients, id)
	h.BroadcastConnectedUsers()
}

func (h *Hub) BroadcastConnectedUsers() {
	users := make([]string, 0)
	for id := range h.Clients {
		users = append(users, id)
	}

	msg := []byte("users:" + join(users, ","))
	log.Println(users, msg)
	for _, client := range h.Clients {
		client.Send <- msg
	}
}

func join(arr []string, sep string) string {
	out := ""
	for i, val := range arr {
		out += val
		if i < len(arr)-1 {
			out += sep
		}
	}
	return out
}

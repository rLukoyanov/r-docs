// websockets.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"r-docs/ws"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Message struct {
	Type    string       `json:"type"`
	User    *ws.UserData `json:"user,omitempty"`
	Payload interface{}  `json:"payload"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWebSocket(c echo.Context, hub *ws.Hub) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	query := c.Request().URL.Query()
	userData := ws.UserData{
		ID:    query.Get("userId"),
		Name:  query.Get("userName"),
		Email: query.Get("userEmail"),
		Image: query.Get("userImage"),
	}

	clientID := userData.ID
	if clientID == "" {
		clientID = uuid.New().String()
		userData.ID = clientID
	}

	client := &ws.Client{
		Conn:     conn,
		ID:       clientID,
		UserData: userData,
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	hub.AddClient(client)

	go readPump(client, hub)
	go writePump(client)

	return nil
}

func readPump(client *ws.Client, hub *ws.Hub) {
	defer func() {
		hub.RemoveClient(client.ID)
		client.Conn.Close()
	}()

	for {
		_, msg, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}

		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println(err)
			continue
		}

		outgoing := Message{
			Type:    incoming.Type,
			User:    &client.UserData,
			Payload: incoming.Payload,
		}

		msgBytes, err := json.Marshal(outgoing)
		if err != nil {
			log.Println("Error marshaling message:", err)
			continue
		}

		switch incoming.Type {
		case "text_update", "cursor_update":
			hub.Broadcast(client.ID, msgBytes)
		}
	}
}

func writePump(client *ws.Client) {
	defer client.Conn.Close()

	for msg := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("writePump error:", err)
			break
		}
	}
}

package handlers

import (
	"log"
	"net/http"
	"r-docs/ws"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWebSocket(c echo.Context, hub *ws.Hub) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	clientID := uuid.New().String()
	client := &ws.Client{
		Conn: conn,
		ID:   clientID,
		Send: make(chan []byte, 256),
	}

	log.Println(client)

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
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writePump(client *ws.Client) {
	for msg := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("writePump error:", err)
			break
		}
	}
}

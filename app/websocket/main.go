package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) addClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = true
}

func (h *Hub) removeClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	conn.Close()
}

func (h *Hub) broadcast(messageType int, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Brodcast error: %v", err)
		}
	}
}

func main() {
	hub := NewHub()
	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Upgrade error: %v", err)
			return
		}
		hub.addClient(conn)
		defer hub.removeClient(conn)

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				break
			}
			hub.broadcast(messageType, message)
		}
	})

	router.Run(":8080")
}

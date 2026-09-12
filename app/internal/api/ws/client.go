package ws

import "github.com/gorilla/websocket"

type Client struct {
	UserID   int64
	Username string
	ConvID   int64
	Send     chan []byte
	Conn     *websocket.Conn
	Hub      *Hub
}

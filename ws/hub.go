package ws

import "github.com/gorilla/websocket"

type Hub struct {
	Connection *websocket.Conn
}

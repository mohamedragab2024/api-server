package ws

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	Connections map[*Hub]bool
	Messages    chan []byte
	mu          sync.Mutex
}

func (client *Client) ServeMonitoring(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		client.Messages <- message

	}
}

func (client *Client) ServeOutBound(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	hub := Hub{
		Connection: c,
	}
	client.Connections[&hub] = true
	fmt.Printf("clients cound : %d \n", len(client.Connections))
	for k := range r.Context().Done() {
		logrus.Info(fmt.Sprintf("disconnecting %v", k))
		delete(client.Connections, &hub)
		return
	}

}

func (c *Client) ReadMessage() {
	for m := range c.Messages {

		for cl, v := range c.Connections {
			if v {
				c.mu.Lock()
				err := cl.Connection.WriteMessage(websocket.TextMessage, m)

				if err != nil {
					delete(c.Connections, cl)
					logrus.Error(err)
				}
				c.mu.Unlock()
			}

		}

	}
}

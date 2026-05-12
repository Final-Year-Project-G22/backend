package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait    = 10 * time.Second
	pongWait     = 60 * time.Second
	pingInterval = 30 * time.Second
	maxMsgSize   = 4096
	sendBufSize  = 256
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	accountID uuid.UUID
	send      chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

func NewClient(hub *Hub, conn *websocket.Conn, accountID uuid.UUID) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		accountID: accountID,
		send:      make(chan []byte, sendBufSize),
		done:      make(chan struct{}),
	}
}

func (c *Client) ReadPump() {
	defer c.hub.Unregister(c.accountID)

	c.conn.SetReadLimit(maxMsgSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		threadID, err := uuid.Parse(msg.ThreadID)
		if err != nil {
			continue
		}

		switch msg.Type {
		case MsgTypeSubscribe:
			c.hub.Subscribe(c, threadID)
		case MsgTypeUnsubscribe:
			c.hub.Unsubscribe(c, threadID)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

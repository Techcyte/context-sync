package ws

import (
	"encoding/json"
	"strings"
	"tcs/internal/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	id          string
	application string
	transaction string
	manager     model.Manager
	connection  *websocket.Conn
	send        chan []byte
}

func NewWebsocketClient(manager model.Manager, conn *websocket.Conn, msg []byte) (*WebsocketClient, error) {
	application := ""

	if len(msg) != 0 {
		var message model.Message
		err := json.Unmarshal(msg, &message)
		if err != nil {
			manager.PrintErr(err, "error unmarshalling message %v", string(msg))
			return nil, err
		}

		if message.Info != nil && message.Info.Application != "" {
			application = message.Info.Application
		}
	}

	client := &WebsocketClient{
		id:          uuid.New().String(),
		application: application,
		manager:     manager,
		connection:  conn,
		send:        make(chan []byte, 1024),
	}

	return client, nil
}

func (c *WebsocketClient) SendMessage(msg []byte) {
	c.send <- msg
}

func (c *WebsocketClient) Read() {
	defer func() {
		c.manager.Disconnect() <- c
		c.connection.Close()
	}()

	for {
		_, msg, err := c.connection.ReadMessage()
		if err != nil {
			// I'm not aware of a better way to check this specific error.
			if strings.Contains(err.Error(), "websocket: close") {
				return
			}

			// Print the error if it's something other than a closed websocket.
			c.manager.PrintErr(err, "error reading message")
			return
		}

		c.manager.ReceiveMessage(c, msg)
	}
}

func (c *WebsocketClient) Write() {
	defer func() {
		c.connection.Close()
	}()

	for message := range c.send {
		if len(message) == 0 {
			c.manager.Println("Empty message, skipping")
			continue
		}

		w, err := c.connection.NextWriter(websocket.TextMessage)
		if err != nil {
			c.manager.PrintErr(err, "error getting next writer")
			return
		}
		w.Write(message)

		err = w.Close()
		if err != nil {
			c.manager.PrintErr(err, "error closing writer")
			return
		}
	}
}

func (c *WebsocketClient) Close() {
	close(c.send)
}

func (c WebsocketClient) ID() string {
	return c.id
}

func (c WebsocketClient) Application() string {
	return c.application
}

func (c *WebsocketClient) SetTransaction(transaction string) {
	c.transaction = transaction
}

func (c WebsocketClient) Transaction() string {
	return c.transaction
}

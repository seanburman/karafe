package connection

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type (
	Connection struct {
		websocket *websocket.Conn
		Pool      *Pool
		Key       interface{}
		Messages  chan interface{}
	}
)

func NewConnection(ctx echo.Context) (*Connection, error) {
	// Upgrade connection to websocket
	upgrader := websocket.Upgrader{}
	websocket, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

	if err != nil {
		log.Println("error upgrading connection: ", err.Error())
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	c := &Connection{
		websocket: websocket,
		Key:       uuid.New(),
		Messages:  make(chan interface{}, 16),
	}

	return c, nil
}

func (c *Connection) Close() error {
	if c == nil {
		return fmt.Errorf("attemped to close nil Connection")
	}
	c.Pool.removeConnection(c)
	c.websocket.Close()
	return nil
}

func (c *Connection) Listen() {

	go func(c *Connection) {
		for {
			_, _, err := c.websocket.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
					log.Printf("error: %v", err)
				}
				close(c.Messages)
				break
			}
		}
	}(c)

	for {
		msg, ok := <-c.Messages
		if !ok {
			// Channel has been closed
			fmt.Println("CLOSING")
			c.Close()
			break
		}
		log.Println("Message: ", msg)
		err := c.websocket.WriteJSON(msg)
		if err != nil {
			log.Println(fmt.Errorf("error sending message: %v", err))
			c.Close()
		}
	}
}

func (c *Connection) Publish(msg interface{}) {
	c.Messages <- msg
}

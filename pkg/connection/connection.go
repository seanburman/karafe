package connection

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type (
	Connection struct {
		websocket *websocket.Conn
		Pool      *Pool
		Key       interface{}
		msgs      chan []byte
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
	c := &Connection{websocket: websocket, msgs: make(chan []byte, 16)}

	return c, nil
}

func (c *Connection) Close() error {
	if c == nil {
		log.Fatal("attemped to close nil Connection")
	}
	c.Pool.removeConnection(c)
	c.websocket.Close()
	return nil
}

func (c *Connection) WriteMessage(msg []byte) {
	for {
		select {
		case c.msgs <- msg:
		default:
			c.Close()
		}
		err := c.websocket.WriteJSON(msg)
		if err != nil {
			log.Println(fmt.Errorf("error relaying message"))
			log.Panic(err)
		}
	}
}

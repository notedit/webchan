package webchan

import (
    "log"
    "time"
    "net/http"
    "io/ioutil"
    "github.com/garyburd/go-websocket/websocket"
)

const (
    // Time allowed to write a message to the client.
    writeWait = 10 * time.Second

    // Time allowed to read the next message from the client.
    readWait = 60 * time.Second

    // Send pings to client with this period. Must be less than readWait.
    pingPeriod = (readWait * 9) / 10

    // Maximum message size allowed from client.
    maxMessageSize = 512
)

// connection is an middleman between the websocket connection and the hub.
type connection struct {
    // The websocket connection.
    ws *websocket.Conn

    // Buffered channel of outbound messages.
    send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
    defer func() {
        h.unregister <- c
        c.ws.Close()
    }()
    c.ws.SetReadLimit(maxMessageSize)
    c.ws.SetReadDeadline(time.Now().Add(readWait))

    // should send the cliet a  WELCOME
    h.broadcast <- ""

    for {
        op, r, err := c.ws.NextReader()
        if err != nil {
            break
        }
        switch op {
        case websocket.OpPong:
            c.ws.SetReadDeadline(time.Now().Add(readWait))
        case websocket.OpText:
            message, err := ioutil.ReadAll(r)
            if err != nil {
                break
            }
            h.broadcast <- message
        }
    }
}

// write writes a message with the given opCode and payload.
func (c *connection) write(opCode int, payload []byte) error {
    c.ws.SetWriteDeadline(time.Now().Add(writeWait))
    return c.ws.WriteMessage(opCode, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.ws.Close()
    }()
    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.write(websocket.OpClose, []byte{})
                return
            }
            if err := c.write(websocket.OpText, message); err != nil {
                return
            }
        case <-ticker.C:
            if err := c.write(websocket.OpPing, []byte{}); err != nil {
                return
            }
        }
    }
}

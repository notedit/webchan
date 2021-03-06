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
    // sessionId
    SessionId  string

    // channel 
    ChannelName string

    // The websocket connection.
    ws *websocket.Conn

    // Buffered channel of outbound messages.
    send chan []byte
}

// interface method

func (c *connection) Channel() string {
    return c.ChannelName
}

func (c *connection) UniqId() string {
    return c.SessionId
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
    defer func() {
        h.unregister <- c
        c.ws.Close()
    }()
    c.ws.SetReadLimit(maxMessageSize)
    c.ws.SetReadDeadline(time.Now().Add(readWait))

    welcomeMsg := &WelcomeMsg{}
    welcomeb,err := welcomeMsg.Marshal(c.SessionId)
    if err != nil {
        log.Error("could not create welcome message: %s",err)
        return
    }
    c.send  <- welcomeb

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
            var data []interface{}
            err := json.Unmarshal(message,&data)
            if err != nil {
                break
            }
            c.handleMessage(data)
        }
    }
}

// 解析message 成对应的struct 
// 创建channel 和 connection的对应关系

func (c *connection)handleMessage(message []interface{}) error {
    var typeid int
    if typeid,ok := message[0].(int); !ok {
        return
    }
    
    switch typeid {
    case SUBSCRIBE:
        subscribeMsg := &SubscribeMsg{}
        if err := subscribeMsg.UnmarshalJSON(message); err {
            return
        }
        c.handleSubscribe(subscribeMsg)
    case UNSUBSCRIBE:
        unsubscribeMsg := &UnsubscribeMsg{}
        if err := unsubscribeMsg.UnmarshalJSON(message); err {
            return
        }
        c.handleUnsubscribe(unsubscribeMsg)
    case PUBLISH:
        publishMsg := &PublishMsg{}
        if err := publishMsg.UnmarshalJSON(message); err {
            return
        }
        c.handlePublish(publishMsg)
    }
}

func (c *connection)handleSubscribe(msg *SubscribeMsg ) {
    if HUB.unsubscribeConns[c] {
        HUB.mu.Lock()
        delete(HUB.unsubscirbeConns,c)
        HUB.mu.Unlock()
    }
    if c.Channel() {
        HUB.storeLock.Lock()
        HUB.connStore.DelConn(c)
        HUB.storeLock.Unlock()
    }
    c.ChannelName = msg.Channel
    HUB.storeLock.Lock()
    HUB.connStore.AddConn(c)
    HUB.storeLock.Unlock()
}

func (c *connection)handleUnsubcribe(msg *UnsubscribeMsg){
    if c.Channel() == msg.Channel {
        HUB.storeLock.Lock()
        HUB.connStore.DelConn(c)
        HUB.storeLock.Unlock()
    }
}

func (c *connection)handlePublish(msg *PublishMsg){
    if c.Channel() != msg.Channel {
        return
    }
    eventMsg := &EventMsg{Channel:msg.Channel,Event:msg.Event,Data:msg.Data}
    HUB.event <- eventMsg
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


func ServeRequest(w http.ResponseWriter, r *http.Request){
    if r.Method != "GET" {
        http.Error(w,"Method not allowed",405)
        return
    }
    if r.Header.Get("Origin") != "http://" + r.Host {
        http.Error(w,"Origin not allowed",403)
        return
    }
    ws,err := websocket.Upgrade(w,r.Header,nil,1024,1024)
    if _,ok := err.(websocket.HandshakeError); ok {
        http.Error(w,"Not a websocket handshake",400)
        return
    } else if err != nil {
        log.Println(err)
        return
    }

    tid,err := uuid.NewV4()
    if err != nil {
        log.Error("could not create sessionid")
        return
    }
    id := tid.String()

    c := &connection{SessionId:id,send:make(chan []byte,256),ws:ws}
    h.register <- c
    go c.writePump()
    c.readPump()
}

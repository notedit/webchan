package webchan

type Hub struct {
    mu                  sync.RWMutex
    unsubcribeConns     map[*connection]bool
    storeLock           sync.Mutex
    conns               connStore 
    event               chan *EventMsg
    register            chan *connection
    unregister          chan *connection
}

var HUB = Hub{
        unsubcribeConns: make(map[*connection]bool),
        conns:          newConnStore("map"),
        event:          make(chan *EventMsg,255),
        register:       make(chan *connection),
        unregister:     make(chan *connection)
    }


func (h *Hub) run() {
    for {
        select {
        case c := <-h.register:
            h.handleRegiste(c)
        case c := <-h.unregister:
            h.handleUnregiste(c)
        case e := <-h.event:
            h.handleEvent(e)
        }
    }
}

func (h *Hub) handleRegiste(c *connection) {
    if h.unsubcribeConns[c] {
        return
    }
    h.mu.Lock()
    h.unsubcribeConns[c] = true
    h.mu.Unlock()
}

func (h *Hub) handleUnregiste(c *connection) {
    close(c.send)
    if h.unsubscibeConns[c]{
        h.mu.Lock()
        delete(h.unsubscribeConns,c)
        h.mu.Unlock()
    }
    if c.Channel() {
        h.storeLock.Lock()
        h.connStore.DelConn(c)
        h.storeLock.Unlock()
    }
}

func (h *Hub) handleEvent(e *EventMsg){
    conns := h.connStore.GetConn(e.Channel)
    if conns == nil {
        return
    }
    data,err := e.Marshal()
    for _,c := range conns {
        select {
        case c.send <- data:
        default:
            h.handleUnregiste(c)
        }
    }

}

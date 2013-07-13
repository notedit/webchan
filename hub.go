package webchan

type Hub struct {
    mu                  sync.RWMutex
    unsubcribeConns     map[*connection]bool
    storeLock           sync.Mutex
    conns               connStore 
    event               chan []bype
    register            chan *connection
    unregister          chan *connection
}

var HUB = Hub{
        unsubcribeConns: make(map[*connection]bool),
        conns:          newConnStore("map"),
        event:          make(chan []byte,255),
        register:       make(chan *connection),
        unregister:     make(chan *connection)
    }


func (h *Hub) run() {
    for {
        select {
        case c := <-h.register:
            h.connections[c] = true
        case c := <-h.unregister:
            delete(h.connections, c)
            close(c.send)
        case m := <-h.broadcast:
            for c := range h.connections {
                select {
                case c.send <- m:
                default:
                    close(c.send)
                    delete(h.connections, c)
                }
            }
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

func (h *Hub) handleEvent(){
    // todo
}

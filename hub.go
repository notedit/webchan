package webchan

type Hub struct {
    connections map[*connection]bool
    broadcast   chan []byte
    register    chan *connection
    unregister  chan *connection
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

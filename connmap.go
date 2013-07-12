package webchan

import (
    "errors"
    "github.com/petar/GoLLRB/llrb"
)

type mConn interface {
    // Userid() string   if a user has multiconnection
    Channel() string
    UniqId() string
}

type connStore interface {
    AddConn(conn mConn) error
    GetConn(channel string) []mConn
    DelConn(conn mConn) bool
}

type connListItem struct{
    channel string
    list []mConn
}

func (self *connListItem) key() string {
    return self.channel
}

func (self *connListItem) Less(than llrb.Item) bool {
    selfKey := llrb.String(self.key())
    thanKey := llrb.String(than.(*connListItem).key())
    return selfKey.Less(thanKey)
}

type connMap struct {
    mmap    map[string]*connListItem 
}

func (self *connMap) GetConn(channel string) []mConn {
    if channel == nil {
        return nil
    }
    clif,ok := self.mmap[channel]
    if !ok {
        return nil
    }
    return clif.list
}

func (self *connMap) AddConn(conn mConn) error {
    if conn == nil {
        return nil
    }
    var ok bool
    var clif *connListItem
    clif,ok = self.mmap[conn.Channel()]
    if !ok {
        cl := make([]mConn)
        cl = append(cl,conn)
        clif = &connListItem{channel:conn.Channel(),cl}
        self.mmap[conn.Channel()] = clif
    } else {
        clif.list = append(clif.list,conn)
        self.mmap[conn.Channel()] =  clif
    }
    return nil
}

func (self *connMap) DelConn(conn mConn) bool {
    if conn == nil {
        return false
    }
    var ok bool
    var clif *connListItem
    clif,ok = self.mmap[conn.Channel()]
    if !ok {
        return false
    }
    var i = -1
    var v  mConn
    for i,v := range clif.list {
        if v.UniqueId() == conn.UniqueId() {
            break
        }
    }
    if v.UniqueId() != conn.UniqueId() {
        return false
    }
    if len(clif.list) == 1 {
        delete(self.mmap,conn.Channel())
        return true
    }
    clif.list[i] = clif.list[len(clif.list)-1]
    clif.list = clif.list[:len(clif.list)-1]
    self.mmap[conn.Channel()] = clif
    return true
}

type connTree struct {
    tree    *llrb.LLRB
}

func (self *connTree)GetConn(channel string) []mConn {
    key := &connListItem{channel:channel,list:nil}
    clif := self.tree.Get(key)
    cl,ok := clif.(*connListItem)
    if !ok || cl == nil {
        return nil
    }
    return cl.list
}

func (self *connTree) AddConn(conn mConn) error {
    if conn == nil {
        return nil
    }
    var cl []mConn
    cl = self.GetConn(conn.Channel())
    if cl == nil {
        cl = make([]mConn)
    }
    for _,c := range cl {
        if c.UniqId() == conn.UniqId() {
            return nil
        }
    }
    cl = append(cl,conn)
    key := &connListItem{channel:conn.Channel(),list:cl}
    self.tree.ReplaceOrInsert(key)
    return nil
}

func (self *connTree) DelConn(conn mConn) bool {
    if conn == nil {
        return false
    }
    cl := self.GetConn(conn.Channel())
    if cl == nil {
        return false
    }
    i := -1
    var c mConn
    for i,c = range cl {
        if c.UniqId() == conn.UniqId() {
            break
        }
    }
    if i < 0 {
        return false
    }
    if len(cl) == 1 {
        key := &connListItem{channel:conn.Channel(),list:cl}
        c := self.tree.Delete(key)
        if c == nil {
            return false
        }
        return true
    }
    cl[i] = cl[len(cl)-1]
    cl = cl[:len(cl)-1]
    key := &connListItem{channel:conn.Channel(),list:cl}
    if len(cl) == 0 {
        self.tree.Delete(key)
    } else {
        self.tree.ReplaceOrInsert(key)
    }
    return true
}


func newConnStore(type) connStore {
    if type == "tree" {
        ret := new(connTree)
        ret.tree = llrb.New()
        return ret
    } else {
        ret := new(connMap)
        return ret
    }
}

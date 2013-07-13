package webchan



const SERVER_IDENT = "WEBCHAN"

const (
    WELCOME = 0
    SUBSCRIBE = 5
    UNSUBSCRIBE = 6
    PUBLISH = 7
    EVENT = 8
)

type ProtocolError struct{
    Msg     string
}

func (self *ProtocolError) Error() string {
    return  "webchan:" + self.Msg
}

var ErrInvalidNumArgs = &ProtocolError{ "invalid number of argument" }

type WelcomeMsg struct {
    SessionId           string
    ProtocalVersion     string
    ServerIdent         string
}

func (self *WelcomeMsg)Marshal(sessionId string) ([]byte,error){
    data := [...]interface{}{WELCOME,sessionId,"1",SERVER_IDENT}
    b,err := json.Marshal(data)
    return b,err
}

type SubscribeMsg struct {
    Channel             string
}

func (self *SubscribeMsg)UnmarshalJSON(data []interface{}) error {
    if len(data) != 2 {
        return ErrInvalidNumArgs
    }
    var ok bool
    if self.Channel,ok = data[1].(string); !ok {
        return &ProtocolError{"invalid channel"}
    }
    return nil
}

type UnsubscribeMsg struct{
    Channel             string
}

func (self *UnsubcribeMsg)UnmarshalJSON(data []interface{}) error {
    if len(data) != 2 {
        return ErrInvalidNumArgs
    }
    var ok bool
    if self.Channel,ok = data[1].(string); !ok {
        return &ProtocalError{"invalid channel"}
    }
    return nil
}

type PublishMsg struct {
    Channel             string
    Event               string
    Data                interface{}
}

func (self *PublishMsg)UnmarshalJSON(data []interface{}) error {
    if len(data) != 4 {
        return ErrInvalidNumArgs
    }
    var ok bool
    if self.Channel,ok = data[1].(string); !ok {
        return &ProtocolError{"invalid channel"}
    }
    if self.Event,ok = data[2].(string); !ok {
        return &ProtocolError{"invalid event"}
    }
    self.Data = data[3]
    return nil
}


type EventMsg struct {
    Channel             string
    Event               string
    Data                interface{}
}

func (self *EventMsg)Marshal() ([]byte,error){
    data := []interface{}{EVENT,self.Channel,self.Event,self.Data}
    b,err :=  json.Marshal(data)
    return b,err
}


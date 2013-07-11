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

func (self *WelcomeMsg)NewWelcomeMsg(sessionId string) (string,error){
    data := [...]interface{}{WELCOME,sessionId,"1",SERVER_IDENT}
    b,err := json.Marshal(data)
    return string(b),err
}

type SubscribeMsg struct {
    Channel             string
}

func (self *SubscribeMsg)UnmarshalJSON(jsonData []byte) error {
    var data []interface{}
    err := json.Unmarshal(jsonData,&data)
    if err != nil {
        return err
    }
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

func (self *UnsubcribeMsg)UnmarshalJSON(jsonData []byte) error {
    var data []interface{}
    err := json.Unmarshal(jsonData,&data)
    if err != nil {
        return err
    }
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
    ExcludeMe           bool
    ExcludeList         []string
    EligibleList        []string
}

func (self *PublishMsg)UnmarshalJSON(jsonData []byte) error {
    var data []interface{}
    err := json.Unmarshal(jsonData,&data)
    if err != nil {
        return err
    }
    if len(data) < 3 || len(data) > 5 {
        return ErrInvalidNumArgs
    }
    var ok bool
    if self.Channel,ok = data[1].(string); !ok {
        return &ProtocolError{"invalid channel"}
    }
    self.Event = data[2]
    if len(data) > 3 {
        if self.ExcludeMe,ok = data[3].(bool); !ok {
            var arr []interface{}
            if arr,ok = data[3].([]interface{}); !ok && data[3] != nil {
                return &ProtocolError{"invalid exclude argument"}
            }
            for _,v  := range arr {
                if val,ok := v.(string); !ok {
                    return &ProtocolError{"invalid exclude list"}
                } else {
                    self.ExcludeList = append(self.ExcludeList,val)
                }
            }

            if len(data) == 5 {
                if arr,ok = data[4].([]interface{}); !ok && data[3] != nil {
                    return &ProtocolError{"invalid eligable list"}
                }
                for _,v := range arr {
                    if val,ok := v.(string); !ok {
                        return &ProtocolError{"invalid eligalbe list"}
                    } else {
                        self.EligibleList = append(self.EligibleList,val)
                    }
                }
            }
        }
    }
    return nil
}


type EventMsg struct {
    Channel             string
    Event               interface{}
}

func (self *EventMsg) NewEventMsg(channel string, event interface{}) (string,error){
    data := [...]interface{}{EVENT,channel,event}
    b,err :=  json.Marshal(data)
    return string(b),err
}


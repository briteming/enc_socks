package relay

import "errors"

type AuthMsg struct {
    User string
    Pwd string
}

func NewAuthMsg(user, pwd string) *AuthMsg {
    return &AuthMsg{user, pwd}
}

type AuthMap struct {
    info map[string]* AuthMsg
}

func NewAuthMap() *AuthMap {
    return &AuthMap{info : make(map[string]* AuthMsg)}
}

func(this *AuthMap) Add(msg *AuthMsg) error {
    if msg == nil {
        return errors.New("null auth msg")
    }
    this.info[msg.User] = msg
    return nil
}

type AuthIterInterface interface {
    OnGet(msg *AuthMsg)
}

func(this *AuthMap) IterAll(iterInterface AuthIterInterface) {
    for _, value := range this.info {
        iterInterface.OnGet(value)
    }
}

func(this *AuthMap) Check(msg *AuthMsg) bool {
    if msg == nil {
        return false
    }
    v, ok := this.info[msg.User]
    if ok && v.Pwd == msg.Pwd {
        return true
    }
    return false
}
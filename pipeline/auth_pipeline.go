package pipeline

import (
    "net"
    "time"
    "errors"
    "encoding/binary"
    "fmt"
    "encoding/gob"
    "bytes"
)

type AuthSvrPipeline struct {
    mp *map[string]interface{}
}

type AuthCliPipeline struct {
    mp *map[string]interface{}
}

type UserInfo struct {
    User string
    Pwd string
}

func(this *AuthSvrPipeline) Process(conn net.Conn) (net.Conn, error) {
    timeout := 5 * time.Second
    tmpTimeout, err := GetInt(this.mp, "server.auth_timeout")
    if err == nil {
        timeout = time.Duration(tmpTimeout) * time.Second
    }
    //total + gob(UserInfo)
    buf := make([]byte, 6)
    readIndex := 0
    for {
        conn.SetReadDeadline(time.Now().Add(timeout))
        cnt, err := conn.Read(buf[readIndex:])
        if err != nil {
            return conn, errors.New("read auth info failed, err:" + err.Error())
        }
        readIndex += cnt
        if readIndex < 6 {
            continue
        } else if readIndex == 6 {
            total := binary.BigEndian.Uint16(buf)
            if total > 128 {
                return conn, errors.New("auth info too long, skip")
            }
            tmp := make([]byte, total)
            copy(tmp, buf)
            buf = tmp
        } else if readIndex == len(buf) {
            decoder := gob.NewDecoder(bytes.NewBuffer(buf[2:]))
            ui := UserInfo{}
            err := decoder.Decode(&ui)
            if err != nil {
                return conn, errors.New("decode auth info failed, err:" + err.Error())
            }
            authinfo, err := Get(this.mp, "server.auth_info")
            if err == nil {
                pwd, ok := authinfo.(map[string]interface{})[ui.User]
                if !ok || pwd.(string) != ui.Pwd {
                    return conn, errors.New(fmt.Sprintf("auth check failed, user:%s, pwd:%s", ui.User, ui.Pwd))
                }
            }
            conn.SetReadDeadline(time.Time{})
            break
        }
    }
    return conn, nil
}

func(this *AuthCliPipeline) Process(conn net.Conn) (net.Conn, error) {
    ui := UserInfo{}
    if user, err := GetString(this.mp, "client.user"); err != nil {
        return conn, errors.New("not found user setting, e:" + err.Error())
    } else {
        ui.User = user
    }
    if pwd, err := GetString(this.mp, "client.pwd"); err != nil {
        return conn, errors.New("not found pwd setting, e:" + err.Error())
    } else {
        ui.Pwd = pwd
    }
    buf:=new(bytes.Buffer)
    encoder := gob.NewEncoder(buf)
    err := encoder.Encode(ui)
    if err != nil {
        return conn, errors.New("encode user info failed, e:" + err.Error())
    }
    data := make([]byte, 2 + len(buf.Bytes()))
    binary.BigEndian.PutUint16(data, uint16(len(data)))
    copy(data[2:], buf.Bytes())
    writeIndex := 0
    writeTotal := len(data)
    for ; writeIndex < writeTotal; {
        cnt, err := conn.Write(data[writeIndex:])
        if err != nil {
            return conn, errors.New("write auth info failed, err:" + err.Error())
        }
        writeIndex += cnt
    }
    return conn, nil
}

type AuthFactory struct {
    mp *map[string]interface{}
}

func(this *AuthFactory) GetName() string {
    return "auth"
}

func(this *AuthFactory) GetSvrPipe() Pipeline {
    return &AuthSvrPipeline{ this.mp }
}

func(this *AuthFactory) GetCliPipe() Pipeline {
    return &AuthCliPipeline{this.mp }
}

func(this *AuthFactory) Init(mp *map[string]interface{}) error {
    this.mp = mp
    return nil
}

func init() {
    auth := &AuthFactory{}
    GetHolder().Regist(auth)
}
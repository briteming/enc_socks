package enc_socks

import (
    "time"
    "bytes"
    "fmt"
)

const (
    SERVER_TYPE_LOCAL = 1;
    SERVER_TYPE_REMOTE = 2;
)

type ServerConfig struct {
    LocalAddr  string
    RemoteAddr string
    Timeout    time.Duration
    ServerType int32
}

func(this *ServerConfig) String() string {
    var buffer bytes.Buffer
    buffer.WriteString("ServerConfig:[")
    buffer.WriteString("local:" + this.LocalAddr + ", ")
    buffer.WriteString("remote:" + this.RemoteAddr + ", ")
    buffer.WriteString("timeout:" + fmt.Sprintf("%ds, ", this.Timeout / time.Second))
    buffer.WriteString("type:"  + fmt.Sprintf("%d, ", this.ServerType))
    buffer.WriteString("]")
    return buffer.String()
}
